package relations

import (
	"strings"
)

func transformToRelatedContent(uuids []string, publicAPIURL string) []relatedContent {
	mappedRelatedContent := []relatedContent{}
	for _, u := range uuids {
		c := relatedContent{
			APIURL: buildAPIURL(u, "content", publicAPIURL),
			ID:     buildAPIURL(u, "things", publicAPIURL),
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

func buildAPIURL(uuid, path, baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/" + path + "/" + uuid
}
