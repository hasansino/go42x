package kwb

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type SearchResult struct {
	Path    string
	Score   float64
	Type    string
	Preview string
}

type searcher struct {
	settings     *Settings
	indexManager *indexManager
}

func newSearcher(settings *Settings, indexManager *indexManager) *searcher {
	return &searcher{
		settings:     settings,
		indexManager: indexManager,
	}
}

func (s *searcher) Search(queryStr string, limit int) ([]SearchResult, error) {
	index, err := s.indexManager.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	if limit <= 0 {
		limit = s.settings.SearchLimit
	}

	// Build query - use query string for flexibility
	bleveQuery := bleve.NewQueryStringQuery(queryStr)

	searchRequest := bleve.NewSearchRequestOptions(bleveQuery, limit, 0, false)
	searchRequest.Fields = []string{"path", "type"}

	// Configure highlighting
	highlight := bleve.NewHighlight()
	if s.settings.HighlightStyle == "html" {
		highlight = bleve.NewHighlightWithStyle("html")
	}
	// Default highlight style works well for ANSI
	searchRequest.Highlight = highlight

	result, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	results := make([]SearchResult, 0, len(result.Hits))
	for _, hit := range result.Hits {
		sr := SearchResult{
			Path:  hit.ID,
			Score: hit.Score,
		}

		if pathField, ok := hit.Fields["path"].(string); ok {
			sr.Path = pathField
		}
		if typeField, ok := hit.Fields["type"].(string); ok {
			sr.Type = typeField
		}

		if len(hit.Fragments) > 0 {
			for _, fragments := range hit.Fragments {
				if len(fragments) > 0 {
					sr.Preview = fragments[0]
					break
				}
			}
		}

		results = append(results, sr)
	}

	return results, nil
}

func (s *searcher) GetFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(content), nil
}

func (s *searcher) ListFiles(fileType string) ([]string, error) {
	index, err := s.indexManager.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	var q query.Query
	if fileType != "" {
		termQuery := bleve.NewTermQuery(fileType)
		termQuery.SetField("type")
		q = termQuery
	} else {
		q = bleve.NewMatchAllQuery()
	}

	searchRequest := bleve.NewSearchRequestOptions(q, 1000, 0, false)
	searchRequest.Fields = []string{"path", "type"}

	result, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	files := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		files = append(files, hit.ID)
	}

	return files, nil
}
