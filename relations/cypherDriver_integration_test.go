package relations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/content-collection-rw-neo4j/collection"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type payloadData struct {
	uuid   string
	path   string
	id     string
	apiURL string
}

var (
	leadContentSP = payloadData{"3fc9fe3e-af8c-4a4a-961a-e5065392bb31", "./fixtures/Content-with-SP-3fc9fe3e-af8c-4a4a-961a-e5065392bb31.json",
		"http://api.ft.com/things/3fc9fe3e-af8c-4a4a-961a-e5065392bb31", "http://api.ft.com/content/3fc9fe3e-af8c-4a4a-961a-e5065392bb31"}
	leadContentCP = payloadData{"3fc9fe3e-af8c-1b1b-961a-e5065392bb31", "./fixtures/Content-with-CP-3fc9fe3e-af8c-1b1b-961a-e5065392bb31.json",
		"http://api.ft.com/things/3fc9fe3e-af8c-1b1b-961a-e5065392bb31", "http://api.ft.com/content/3fc9fe3e-af8c-1b1b-961a-e5065392bb31"}
	relatedContent1 = payloadData{"3fc9fe3e-af8c-1a1a-961a-e5065392bb31", "./fixtures/Content-3fc9fe3e-af8c-1a1a-961a-e5065392bb31.json",
		"http://api.ft.com/things/3fc9fe3e-af8c-1a1a-961a-e5065392bb31", "http://api.ft.com/content/3fc9fe3e-af8c-1a1a-961a-e5065392bb31"}
	relatedContent2 = payloadData{"3fc9fe3e-af8c-2a2a-961a-e5065392bb31", "./fixtures/Content-3fc9fe3e-af8c-2a2a-961a-e5065392bb31.json",
		"http://api.ft.com/things/3fc9fe3e-af8c-2a2a-961a-e5065392bb31", "http://api.ft.com/content/3fc9fe3e-af8c-2a2a-961a-e5065392bb31"}
	relatedContent3 = payloadData{"3fc9fe3e-af8c-3a3a-961a-e5065392bb31", "./fixtures/Content-3fc9fe3e-af8c-3a3a-961a-e5065392bb31.json",
		"http://api.ft.com/things/3fc9fe3e-af8c-3a3a-961a-e5065392bb31", "http://api.ft.com/content/3fc9fe3e-af8c-3a3a-961a-e5065392bb31"}
	storyPackage = payloadData{"63559ba7-b48d-4467-b2b0-ce956f9e9494", "./fixtures/StoryPackage-63559ba7-b48d-4467-b2b0-ce956f9e9494.json",
		"", ""}
	contentPackage = payloadData{"63559ba7-b48d-4467-1b1b-ce956f9e9494", "./fixtures/ContentPackage-63559ba7-b48d-4467-1b1b-ce956f9e9494.json",
		"", ""}
	allData = []payloadData{leadContentSP, leadContentCP, relatedContent1, relatedContent2, relatedContent3, storyPackage, contentPackage}
)

func TestFindContentRelations_StoryPackage_Ok(t *testing.T) {
	if testing.Short() {
		t.Skip("Short flag is set. Skipping integration test")
	}
	expectedResponse := relations{
		CuratedRelatedContents: []relatedContent{
			{relatedContent1.id, relatedContent1.apiURL},
			{relatedContent2.id, relatedContent2.apiURL},
			{relatedContent3.id, relatedContent3.apiURL},
		},
	}
	conn := getDatabaseConnection(t)
	contents := []payloadData{leadContentSP, relatedContent1, relatedContent2, relatedContent3}
	cleanDB(t, conn, allData)

	writeContent(t, conn, contents)
	writeContentCollection(t, conn, []payloadData{storyPackage}, "StoryPackage")
	defer cleanDB(t, conn, allData)

	driver := NewCypherDriver(conn)
	actualRelations, found, err := driver.findContentRelations(leadContentSP.uuid)
	assert.NoError(t, err, "Unexpected error for content %s", leadContentSP.uuid)
	assert.True(t, found, "Found no relations for content %s", leadContentSP.uuid)

	assert.Equal(t, len(expectedResponse.CuratedRelatedContents), len(actualRelations.CuratedRelatedContents), "Didn't get the same number of curated related content")
	assertListContainsAll(t, actualRelations.CuratedRelatedContents, expectedResponse.CuratedRelatedContents)
}

func TestFindContentRelations_ContentPackage_Ok(t *testing.T) {
	if testing.Short() {
		t.Skip("Short flag is set. Skipping integration test")
	}
	expectedResponse := relations{
		Contains: []relatedContent{
			{relatedContent1.id, relatedContent1.apiURL},
			{relatedContent2.id, relatedContent2.apiURL},
		},
	}
	conn := getDatabaseConnection(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}
	cleanDB(t, conn, allData)

	writeContent(t, conn, contents)
	writeContentCollection(t, conn, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, conn, allData)

	driver := NewCypherDriver(conn)
	actualRelations, found, err := driver.findContentRelations(leadContentCP.uuid)
	assert.NoError(t, err, "Unexpected error for content %s", leadContentCP.uuid)
	assert.True(t, found, "Found no relations for content %s", leadContentCP.uuid)

	assert.Equal(t, len(expectedResponse.Contains), len(actualRelations.Contains), "Didn't get the same number of content in contains")
	assertListContainsAll(t, actualRelations.Contains, expectedResponse.Contains)
}

func TestFindContentRelations_Content_In_ContentPackage_Ok(t *testing.T) {
	if testing.Short() {
		t.Skip("Short flag is set. Skipping integration test")
	}
	expectedResponse := relations{
		ContainedIn: []relatedContent{
			{leadContentCP.id, leadContentCP.apiURL},
		},
	}
	conn := getDatabaseConnection(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}
	cleanDB(t, conn, allData)

	writeContent(t, conn, contents)
	writeContentCollection(t, conn, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, conn, allData)

	driver := NewCypherDriver(conn)
	actualRelations, found, err := driver.findContentRelations(relatedContent1.uuid)
	assert.NoError(t, err, "Unexpected error for content %s", relatedContent1.uuid)
	assert.True(t, found, "Found no relations for content %s", relatedContent1.uuid)

	assert.Equal(t, len(expectedResponse.ContainedIn), len(actualRelations.ContainedIn), "Didn't get the same number of containedIn content")
	assertListContainsAll(t, actualRelations.ContainedIn, expectedResponse.ContainedIn)
}

func TestFindContentCollectionRelations_Ok(t *testing.T) {
	if testing.Short() {
		t.Skip("Short flag is set. Skipping integration test")
	}
	expectedResponse := ccRelations{
		ContainedIn: "3fc9fe3e-af8c-1b1b-961a-e5065392bb31",
		Contains:    []string{"3fc9fe3e-af8c-1a1a-961a-e5065392bb31", "3fc9fe3e-af8c-2a2a-961a-e5065392bb31"},
	}
	conn := getDatabaseConnection(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}
	cleanDB(t, conn, allData)

	writeContent(t, conn, contents)
	writeContentCollection(t, conn, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, conn, allData)

	driver := NewCypherDriver(conn)
	actualRelations, found, err := driver.findContentCollectionRelations(contentPackage.uuid)
	assert.NoError(t, err, "Unexpected error for content package %s", contentPackage.uuid)
	assert.True(t, found, "Found no relations for content package %s", contentPackage.uuid)

	assert.Equal(t, actualRelations.ContainedIn, expectedResponse.ContainedIn)
	assert.Equal(t, len(expectedResponse.Contains), len(actualRelations.Contains), "Didn't get the same number of content in contains")
	assertListContainsAll(t, actualRelations.Contains, expectedResponse.Contains)
}

func writeContent(t testing.TB, conn neoutils.NeoConnection, data []payloadData) baseftrwapp.Service {
	contentRW := content.NewCypherContentService(conn)
	assert.NoError(t, contentRW.Initialise())
	for _, d := range data {
		writeJSONWithService(t, contentRW, d.path)
	}
	return contentRW
}

func writeContentCollection(t testing.TB, conn neoutils.NeoConnection, data []payloadData, ccType string) {
	labels := []string{}
	relation := "CONTAINS"
	if ccType == "StoryPackage" {
		labels = []string{"Curation", "StoryPackage"}
		relation = "SELECTS"
	}

	contentCollectionRW := collection.NewContentCollectionService(conn, labels, relation)
	assert.NoError(t, contentCollectionRW.Initialise())
	for _, d := range data {
		writeJSONWithContentCollectionService(t, contentCollectionRW, d.path)
	}
}

func writeJSONWithService(t testing.TB, service baseftrwapp.Service, pathToJSONFile string) {
	path, err := filepath.Abs(pathToJSONFile)
	require.NoError(t, err)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	require.NoError(t, err)
	err = service.Write(inst)
	require.NoError(t, err)
}

func writeJSONWithContentCollectionService(t testing.TB, service baseftrwapp.Service, pathToJSONFile string) {
	path, err := filepath.Abs(pathToJSONFile)
	require.NoError(t, err)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	require.NoError(t, err)
	err = service.Write(inst)
	require.NoError(t, err)
}

func assertListContainsAll(t *testing.T, list interface{}, items ...interface{}) {
	if reflect.TypeOf(items[0]).Kind().String() == "slice" {
		expected := reflect.ValueOf(items[0])
		expectedLength := expected.Len()
		assert.Len(t, list, expectedLength)
		for i := 0; i < expectedLength; i++ {
			assert.Contains(t, list, expected.Index(i).Interface())
		}
	} else {
		assert.Len(t, list, len(items))
		for _, item := range items {
			assert.Contains(t, list, item)
		}
	}
}

func getDatabaseConnection(t testing.TB) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	conn, err := neoutils.Connect(url, conf)
	require.NoError(t, err, "Failed to connect to Neo4j")
	return conn
}

func cleanDB(t testing.TB, conn neoutils.NeoConnection, data []payloadData) {
	qs := make([]*neoism.CypherQuery, len(data))
	for i, d := range data {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)<-[iden:IDENTIFIES]-(i:Identifier)
			DELETE iden, i
			DETACH DELETE a`, d.uuid)}
	}
	err := conn.CypherBatch(qs)
	assert.NoError(t, err)
}
