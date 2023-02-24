package relations

import (
	"strings"
)

const thingURL = "http://api.ft.com/things/"

func transformToRelatedContent(uuids []string, publicAPIURL string) []relatedContent {
	mappedRelatedContent := []relatedContent{}
	for _, u := range uuids {
		c := relatedContent{
			APIURL: apiURL(u, publicAPIURL),
			ID:     thingIDURL(u),
		}
		mappedRelatedContent = append(mappedRelatedContent, c)
	}

	return mappedRelatedContent
}

func transformContainedInToCCRelations(containedIn []neoRelatedContent) string {
	var leadArticleUuid string
	if len(containedIn) != 0 {
		leadArticleUuid = containedIn[0].UUID
	}
	return leadArticleUuid
}

func transformContainsToCCRelations(neoRelatedContent []neoRelatedContent) []string {
	var contains []string
	for _, neoContent := range neoRelatedContent {
		contains = append(contains, neoContent.UUID)
	}
	return contains
}

func thingIDURL(uuid string) string {
	return thingURL + uuid
}

func apiURL(uuid, baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/content/" + uuid
}
