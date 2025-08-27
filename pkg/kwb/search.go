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

type Searcher struct {
	settings     *Settings
	indexManager *IndexManager
}

func newSearcher(settings *Settings, indexManager *IndexManager) *Searcher {
	return &Searcher{
		settings:     settings,
		indexManager: indexManager,
	}
}

func (s *Searcher) Search(queryStr string, limit int) ([]SearchResult, error) {
	index, err := s.indexManager.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	if limit <= 0 {
		limit = s.settings.SearchLimit
	}

	bleveQuery := bleve.NewQueryStringQuery(queryStr)
	searchRequest := bleve.NewSearchRequestOptions(bleveQuery, limit, 0, false)
	searchRequest.Fields = []string{"Path", "Type"}
	searchRequest.Highlight = bleve.NewHighlight()

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

		if pathField, ok := hit.Fields["Path"].(string); ok {
			sr.Path = pathField
		}
		if typeField, ok := hit.Fields["Type"].(string); ok {
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

func (s *Searcher) GetFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(content), nil
}

func (s *Searcher) ListFiles(fileType string) ([]string, error) {
	index, err := s.indexManager.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("getting index: %w", err)
	}

	var q query.Query
	if fileType != "" {
		termQuery := bleve.NewTermQuery(fileType)
		termQuery.SetField("Type")
		q = termQuery
	} else {
		q = bleve.NewMatchAllQuery()
	}

	searchRequest := bleve.NewSearchRequestOptions(q, 1000, 0, false)
	searchRequest.Fields = []string{"Path", "Type"}

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
