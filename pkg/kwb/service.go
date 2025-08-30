package kwb

import (
	"context"
	"fmt"
	"log/slog"
)

type Service struct {
	logger       *slog.Logger
	settings     *Settings
	indexManager *indexManager
	searcher     *searcher
}

func NewService(settings *Settings, opts ...Option) (*Service, error) {
	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}

	svc := &Service{
		settings: settings,
	}

	for _, opt := range opts {
		opt(svc)
	}

	if svc.logger == nil {
		svc.logger = slog.New(slog.DiscardHandler)
	}

	svc.indexManager = newIndexManager(
		settings,
		svc.logger.With("component", "index_manager"),
	)
	svc.searcher = newSearcher(settings, svc.indexManager)

	return svc, nil
}

func (s *Service) BuildIndex(ctx context.Context, rootPath string) error {
	s.logger.InfoContext(ctx, "Building knowledge base index",
		slog.String("root", rootPath),
		slog.String("index_path", s.settings.IndexPath))

	if err := s.indexManager.BuildIndex(rootPath); err != nil {
		return fmt.Errorf("building index: %w", err)
	}

	stats, err := s.indexManager.GetStats()
	if err != nil {
		return fmt.Errorf("error getting stats: %w", err)
	}

	s.logger.InfoContext(ctx, "Index built successfully",
		"documents_indexed", stats["document_count"].(uint64),
		"index_path", stats["index_path"].(string),
	)

	return nil
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	s.logger.InfoContext(ctx, "Searching knowledge base",
		slog.String("query", query),
		slog.Int("limit", limit))

	results, err := s.searcher.Search(query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}

	s.logger.InfoContext(ctx, "Search complete",
		slog.Int("results", len(results)))

	return results, nil
}

func (s *Service) GetFile(ctx context.Context, path string) (string, error) {
	s.logger.InfoContext(ctx, "Getting file content",
		slog.String("path", path))

	content, err := s.searcher.GetFile(path)
	if err != nil {
		return "", fmt.Errorf("getting file: %w", err)
	}

	return content, nil
}

func (s *Service) ListFiles(ctx context.Context, fileType string) ([]string, error) {
	s.logger.InfoContext(ctx, "Listing files",
		slog.String("type", fileType))

	files, err := s.searcher.ListFiles(fileType)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	s.logger.InfoContext(ctx, "List complete",
		slog.Int("count", len(files)))

	return files, nil
}

func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	s.logger.InfoContext(ctx, "Getting index stats")

	stats, err := s.indexManager.GetStats()
	if err != nil {
		return nil, fmt.Errorf("getting stats: %w", err)
	}

	return stats, nil
}

func (s *Service) Close() error {
	return s.indexManager.CloseIndex()
}
