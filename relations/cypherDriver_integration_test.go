//go:build integration
// +build integration

package relations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	"github.com/Financial-Times/content-collection-rw-neo4j/collection"
	"github.com/Financial-Times/content-rw-neo4j/v3/content"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type payloadData struct {
	uuid   string
	path   string
	id     string
	apiURL string
}

type collectionService interface {
	DecodeJSON(dec *json.Decoder) (interface{}, string, error)
	Write(newThing interface{}, transactionID string) error
}

type contentService interface {
	Write(thing interface{}, transId string) error
	DecodeJSON(dec *json.Decoder) (interface{}, string, error)
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
	driver := getNeo4jDriver(t)
	contents := []payloadData{leadContentSP, relatedContent1, relatedContent2, relatedContent3}

	writeContent(t, driver, contents)
	writeContentCollection(t, driver, []payloadData{storyPackage}, "StoryPackage")
	defer cleanDB(t, driver, allData)

	cypherDriver := NewCypherDriver(driver)
	actualRelations, found, err := cypherDriver.findContentRelations(leadContentSP.uuid)
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
	driver := getNeo4jDriver(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}

	writeContent(t, driver, contents)
	writeContentCollection(t, driver, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, driver, allData)

	cypherDriver := NewCypherDriver(driver)
	actualRelations, found, err := cypherDriver.findContentRelations(leadContentCP.uuid)
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
	driver := getNeo4jDriver(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}

	writeContent(t, driver, contents)
	writeContentCollection(t, driver, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, driver, allData)

	cypherDriver := NewCypherDriver(driver)
	actualRelations, found, err := cypherDriver.findContentRelations(relatedContent1.uuid)
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
	driver := getNeo4jDriver(t)
	contents := []payloadData{leadContentCP, relatedContent1, relatedContent2}

	writeContent(t, driver, contents)
	writeContentCollection(t, driver, []payloadData{contentPackage}, "ContentPackage")
	defer cleanDB(t, driver, allData)

	cypherDriver := NewCypherDriver(driver)
	actualRelations, found, err := cypherDriver.findContentCollectionRelations(contentPackage.uuid)
	assert.NoError(t, err, "Unexpected error for content package %s", contentPackage.uuid)
	assert.True(t, found, "Found no relations for content package %s", contentPackage.uuid)

	assert.Equal(t, actualRelations.ContainedIn, expectedResponse.ContainedIn)
	assert.Equal(t, len(expectedResponse.Contains), len(actualRelations.Contains), "Didn't get the same number of content in contains")
	assertListContainsAll(t, actualRelations.Contains, expectedResponse.Contains)
}

func writeContent(t testing.TB, driver *cmneo4j.Driver, data []payloadData) {
	contentRW := content.NewContentService(driver)
	assert.NoError(t, contentRW.Initialise())
	for _, d := range data {
		writeJSONWithService(t, contentRW, d.path)
	}
}

func writeContentCollection(t testing.TB, driver *cmneo4j.Driver, data []payloadData, ccType string) {
	labels := []string{}
	relation := "CONTAINS"
	if ccType == "StoryPackage" {
		labels = []string{"Curation", "StoryPackage"}
		relation = "SELECTS"
	}

	contentCollectionRW := collection.NewContentCollectionService(driver, labels, relation, "")
	assert.NoError(t, contentCollectionRW.Initialise())
	for _, d := range data {
		writeJSONWithContentCollectionService(t, contentCollectionRW, d.path)
	}
}

func writeJSONWithService(t testing.TB, service contentService, pathToJSONFile string) {
	path, err := filepath.Abs(pathToJSONFile)
	require.NoError(t, err)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	require.NoError(t, err)
	err = service.Write(inst, "TEST_TRANSACTION_ID")
	require.NoError(t, err)
}

func writeJSONWithContentCollectionService(t testing.TB, service collectionService, pathToJSONFile string) {
	path, err := filepath.Abs(pathToJSONFile)
	require.NoError(t, err)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	require.NoError(t, err)
	err = service.Write(inst, "test")
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

func getNeo4jDriver(t testing.TB) *cmneo4j.Driver {
	t.Helper()

	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "bolt://localhost:7687"
	}
	log := logger.NewUPPLogger("cm-neo4j-driver-integration-tests", "PANIC")
	driver, err := cmneo4j.NewDefaultDriver(url, log)
	assert.NoError(t, err, "Unexpected error when creating a new db driver")
	return driver
}

func cleanDB(t testing.TB, driver *cmneo4j.Driver, data []payloadData) {
	qs := make([]*cmneo4j.Query, len(data))
	for i, d := range data {
		qs[i] = &cmneo4j.Query{
			Cypher: `MATCH (a:Thing {uuid: $uuid}) DETACH DELETE a`,
			Params: map[string]interface{}{"uuid": d.uuid},
		}
	}
	err := driver.Write(qs...)
	assert.NoError(t, err)
}
