package relations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var givenNeoRelatedContent []string = []string{
	"db90a9db-6cb6-4ba0-8648-c0676087aba2",
	"f78c1482-abab-413e-b753-ca3ce3cb84f0",
}

var expectedRelatedContent []relatedContent = []relatedContent{
	{ID: "http://api.ft.com/things/db90a9db-6cb6-4ba0-8648-c0676087aba2", APIURL: "http://api.ft.com/content/db90a9db-6cb6-4ba0-8648-c0676087aba2"},
	{ID: "http://api.ft.com/things/f78c1482-abab-413e-b753-ca3ce3cb84f0", APIURL: "http://api.ft.com/content/f78c1482-abab-413e-b753-ca3ce3cb84f0"},
}

func TestTransformToRelatedContentHappyFlow(t *testing.T) {
	relatedContent := transformToRelatedContent(givenNeoRelatedContent)

	expected, _ := json.Marshal(expectedRelatedContent)
	actual, _ := json.Marshal(relatedContent)
	assert.JSONEq(t, string(expected), string(actual))
}

func TestTransformToRelatedContentNoRelations(t *testing.T) {
	givenNeoRelatedContent := []string{}
	expectedRelatedContent := []string{}

	relatedContent := transformToRelatedContent(givenNeoRelatedContent)

	expected, _ := json.Marshal(expectedRelatedContent)
	actual, _ := json.Marshal(relatedContent)
	assert.JSONEq(t, string(expected), string(actual))
}
