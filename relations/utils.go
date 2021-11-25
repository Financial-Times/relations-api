package relations

import "github.com/Financial-Times/neo-model-utils-go/mapper"

func transformToRelatedContent(uuids []string) []relatedContent {
	mappedRelatedContent := []relatedContent{}
	for _, u := range uuids {
		c := relatedContent{
			APIURL: mapper.APIURL(u, []string{"Content"}, "local"),
			ID:     mapper.IDURL(u),
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
