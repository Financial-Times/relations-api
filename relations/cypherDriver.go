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

func NewCypherDriver(driver *cmneo4j.Driver) *cypherDriver {
	return &cypherDriver{driver}
}

func (cd *cypherDriver) checkConnectivity() error {
	return cd.driver.VerifyConnectivity()
}

func (cd *cypherDriver) findContentRelations(contentUUID string) (relations, bool, error) {
	neoCRC := []neoRelatedContent{}
	neoCPContains := []neoRelatedContent{}
	neoCPContainedIn := []neoRelatedContent{}

	queryCRC := &cmneo4j.Query{
		Cypher: `
                MATCH (c:Content{uuid:$contentUUID})<-[:IS_CURATED_FOR]-(cc:Curation)
                MATCH (cc)-[rel:SELECTS]->(t:Content)
                RETURN t.uuid as uuid
                ORDER BY rel.order
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCRC,
	}

	queryCPContains := &cmneo4j.Query{
		Cypher: `
                MATCH (cp:ContentPackage{uuid:$contentUUID})-[:CONTAINS]->(cc:ContentCollection)
                MATCH (cc)-[rel:CONTAINS]->(c:Content)
                RETURN c.uuid as uuid
                ORDER BY rel.order
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCPContains,
	}

	queryCPContainedIn := &cmneo4j.Query{
		Cypher: `
                MATCH (c:Content{uuid:$contentUUID})<-[:CONTAINS]-(cc:ContentCollection)
                MATCH (cc)<-[rel:CONTAINS]-(cp:ContentPackage)
                RETURN cp.uuid as uuid
                ORDER BY rel.order
                `,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &neoCPContainedIn,
	}

	err := cd.driver.Read(queryCRC, queryCPContains, queryCPContainedIn)
	if err != nil && !errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return relations{}, false, fmt.Errorf("Error querying Neo for uuid=%s, err=%v", contentUUID, err)
	}

	var found bool

	if len(neoCRC) != 0 || len(neoCPContains) != 0 || len(neoCPContainedIn) != 0 {
		found = true
	}

	mappedCRC := transformToRelatedContent(neoCRC)
	mappedCPC := transformToRelatedContent(neoCPContains)
	mappedCIC := transformToRelatedContent(neoCPContainedIn)
	relations := relations{mappedCRC, mappedCPC, mappedCIC}

	return relations, found, nil
}

func (cd *cypherDriver) findContentCollectionRelations(contentCollectionUUID string) (ccRelations, bool, error) {
	neoCPContainedIn := []neoRelatedContent{}
	neoCPContains := []neoRelatedContent{}

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
