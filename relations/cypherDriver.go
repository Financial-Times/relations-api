package relations

import (
	"errors"
	"fmt"

	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
)

type Driver interface {
	findContentRelations(UUID string) (res relations, found bool, err error)
	findContentCollectionRelations(UUID string) (res ccRelations, found bool, err error)
	checkConnectivity() error
}

type cypherDriver struct {
	driver *cmneo4j.Driver
}

func NewCypherDriver(driver *cmneo4j.Driver) Driver {
	return &cypherDriver{driver}
}

func (cd *cypherDriver) checkConnectivity() error {
	return cd.driver.VerifyConnectivity()
}

func (cd *cypherDriver) findContentRelations(contentUUID string) (relations, bool, error) {
	var neoCRC, neoCPContains, neoCPContainedIn struct {
		UUIDs []string `json:"uuids"`
	}

	// All of the queries use OPTIONAL MATCH because when a query doesn't match
	// anything, the driver is not executing the queries after that one

	queryCRC := &cmneo4j.Query{
		Cypher: `
                OPTIONAL MATCH (c:Content{uuid:$contentUUID})<-[:IS_CURATED_FOR]-(cc:Curation)
                OPTIONAL MATCH (cc)-[rel:SELECTS]->(t:Content)
                WITH t.uuid as uuid
                ORDER BY rel.order
                RETURN COLLECT(uuid) as uuids
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCRC,
	}

	queryCPContains := &cmneo4j.Query{
		Cypher: `
                OPTIONAL MATCH (cp:ContentPackage{uuid:$contentUUID})-[:CONTAINS]->(cc:ContentCollection)
                OPTIONAL MATCH (cc)-[rel:CONTAINS]->(c:Content)
                WITH c.uuid as uuid
                ORDER BY rel.order
                RETURN COLLECT(uuid) as uuids
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCPContains,
	}

	queryCPContainedIn := &cmneo4j.Query{
		Cypher: `
                OPTIONAL MATCH (c:Content{uuid:$contentUUID})<-[:CONTAINS]-(cc:ContentCollection)
                OPTIONAL MATCH (cc)<-[rel:CONTAINS]-(cp:ContentPackage)
                WITH cp.uuid as uuid
                ORDER BY rel.order
                RETURN COLLECT(uuid) as uuids
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCPContainedIn,
	}

	err := cd.driver.Read(queryCRC, queryCPContains, queryCPContainedIn)
	if err != nil && !errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return relations{}, false, fmt.Errorf("Error querying Neo for uuid=%s, err=%v", contentUUID, err)
	}

	found := len(neoCRC.UUIDs) != 0 || len(neoCPContains.UUIDs) != 0 || len(neoCPContainedIn.UUIDs) != 0

	mappedCRC := transformToRelatedContent(neoCRC.UUIDs)
	mappedCPC := transformToRelatedContent(neoCPContains.UUIDs)
	mappedCIC := transformToRelatedContent(neoCPContainedIn.UUIDs)
	relations := relations{mappedCRC, mappedCPC, mappedCIC}

	return relations, found, nil
}

func (cd *cypherDriver) findContentCollectionRelations(contentCollectionUUID string) (ccRelations, bool, error) {
	neoCPContainedIn := []neoRelatedContent{}
	neoCPContains := []neoRelatedContent{}

	// There is no need to use OPTIONAL MATCH here because the first query
	// determines whether or not the method returns found=true/false

	queryCPContainedIn := &cmneo4j.Query{
		Cypher: `
                MATCH (cc:ContentCollection{uuid:$contentCollectionUUID})<-[rel:CONTAINS]-(cp:ContentPackage)
                RETURN cp.uuid as uuid
                `,
		Params: map[string]interface{}{"contentCollectionUUID": contentCollectionUUID},
		Result: &neoCPContainedIn,
	}

	queryCPContains := &cmneo4j.Query{
		Cypher: `
                MATCH (cc:ContentCollection{uuid:$contentCollectionUUID})-[rel:CONTAINS]->(c:Content)
                RETURN c.uuid as uuid
                ORDER BY rel.order
                `,
		Params: map[string]interface{}{"contentCollectionUUID": contentCollectionUUID},
		Result: &neoCPContains,
	}

	err := cd.driver.Read(queryCPContains, queryCPContainedIn)
	if err != nil && !errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return ccRelations{}, false, fmt.Errorf("Error querying Neo for uuid=%s, err=%v", contentCollectionUUID, err)
	}

	found := len(neoCPContainedIn) != 0

	mappedContainedIn := transformContainedInToCCRelations(neoCPContainedIn)
	mappedContains := transformContainsToCCRelations(neoCPContains)
	ccRelations := ccRelations{mappedContainedIn, mappedContains}

	return ccRelations, found, nil
}
