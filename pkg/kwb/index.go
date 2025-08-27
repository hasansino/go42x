package kwb

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

var defaultExcludedDirs = []string{
	".git",
	"vendor",
	"node_modules",
	".idea",
	".vscode",
	"dist",
	"build",
	"bin",
	".go42x",
}

var defaultExtensions = map[string]bool{
	".go":    true,
	".md":    true,
	".yaml":  true,
	".yml":   true,
	".mod":   true,
	".sum":   true,
	".proto": true,
	".sql":   true,
	".json":  true,
	".toml":  true,
	".env":   true,
	".sh":    true,
}

var allowedExtensions = []string{
	"Makefile",
	"Dockerfile",
	".gitignore",
}

type IndexManager struct {
	logger   *slog.Logger
	settings *Settings
	index    bleve.Index
}

func newIndexManager(settings *Settings, logger *slog.Logger) *IndexManager {
	return &IndexManager{
		logger:   logger,
		settings: settings,
	}
}

func (m *IndexManager) BuildIndex(rootPath string) error {
	// Remove old index if exists
	err := os.RemoveAll(m.settings.IndexPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing old index: %w", err)
	}

	// Create directory for index
	indexDir := filepath.Dir(m.settings.IndexPath)
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return fmt.Errorf("creating index directory: %w", err)
	}

	// Create optimized index mapping
	mapping := m.createOptimizedMapping()

	// Use configured index type (scorch is faster and more memory efficient)
	indexType := m.settings.IndexType
	if indexType == "" {
		indexType = "scorch"
	}

	index, err := bleve.NewUsing(m.settings.IndexPath, mapping, indexType, indexType, nil)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	defer index.Close()

	// Walk and index files with batch processing
	fileCount := 0
	batch := index.NewBatch()
	batchSize := 0
	maxBatchSize := m.settings.BatchSize
	if maxBatchSize <= 0 {
		maxBatchSize = 100
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			m.logger.Error("error accessing path", "err", err, "path", path)
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be excluded
			dirName := filepath.Base(path)
			for _, excl := range defaultExcludedDirs {
				if dirName == excl {
					m.logger.Info("skipping excluded directory", slog.String("path", path))
					return filepath.SkipDir
				}
			}
			for _, excl := range m.settings.ExcludeDirs {
				if dirName == excl {
					m.logger.Info("skipping user-excluded directory", slog.String("path", path))
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file should be indexed
		ext := filepath.Ext(path)
		if !m.shouldIndexFile(m.settings.ExtraExtensions, info.Name(), ext) {
			return nil
		}

		// Skip very large files
		if info.Size() > int64(m.settings.MaxFileSize) {
			m.logger.Warn("skipping large file",
				slog.String("path", path),
				slog.Int64("size", info.Size()))
			return nil
		}

		// Read and index file
		content, err := os.ReadFile(path)
		if err != nil {
			m.logger.Warn("failed to read file",
				slog.String("path", path),
				slog.String("error", err.Error()))
			return nil
		}

		doc := Document{
			ID:      path,
			Path:    path,
			Content: string(content),
			Type:    GetFileType(path),
		}

		// Add to batch
		batch.Index(doc.ID, doc)
		batchSize++

		// Process batch when it reaches max size
		if batchSize >= maxBatchSize {
			if err := index.Batch(batch); err != nil {
				m.logger.Error("failed to process batch",
					slog.String("error", err.Error()))
				return fmt.Errorf("batch indexing failed: %w", err)
			}
			m.logger.Info("processed batch", slog.Int("size", batchSize))
			batch = index.NewBatch()
			batchSize = 0
		}

		fileCount++
		m.logger.Debug("queued file for indexing", slog.String("path", path))
		return nil
	})

	// Process remaining documents in batch
	if batchSize > 0 {
		if err := index.Batch(batch); err != nil {
			return fmt.Errorf("final batch indexing failed: %w", err)
		}
		m.logger.Info("processed final batch", slog.Int("size", batchSize))
	}

	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	count, _ := index.DocCount()
	m.logger.Info("indexing complete",
		slog.Uint64("documents", count),
		slog.Int("files_processed", fileCount))

	return nil
}

func (m *IndexManager) OpenIndex() error {
	if m.index != nil {
		return nil // Already open
	}

	index, err := bleve.Open(m.settings.IndexPath)
	if err != nil {
		return fmt.Errorf("opening index: %w", err)
	}

	m.index = index
	return nil
}

func (m *IndexManager) CloseIndex() error {
	if m.index != nil {
		err := m.index.Close()
		m.index = nil
		return err
	}
	return nil
}

func (m *IndexManager) GetIndex() (bleve.Index, error) {
	if m.index == nil {
		if err := m.OpenIndex(); err != nil {
			return nil, err
		}
	}
	return m.index, nil
}

func (m *IndexManager) GetStats() (map[string]interface{}, error) {
	index, err := m.GetIndex()
	if err != nil {
		return nil, err
	}

	count, err := index.DocCount()
	if err != nil {
		return nil, fmt.Errorf("getting doc count: %w", err)
	}

	stats := map[string]interface{}{
		"document_count": count,
		"index_path":     m.settings.IndexPath,
	}

	return stats, nil
}

func (m *IndexManager) createOptimizedMapping() mapping.IndexMapping {
	mapping := bleve.NewIndexMapping()

	// Configure default analyzer for better code search
	mapping.DefaultAnalyzer = "standard"

	// Create document mapping
	docMapping := bleve.NewDocumentMapping()

	// Path field - keyword for exact matches
	pathField := bleve.NewKeywordFieldMapping()
	pathField.Store = true
	pathField.IncludeInAll = true
	docMapping.AddFieldMappingsAt("path", pathField)

	// Type field - keyword for filtering
	typeField := bleve.NewKeywordFieldMapping()
	typeField.Store = true
	typeField.IncludeInAll = false
	docMapping.AddFieldMappingsAt("type", typeField)

	// Content field - text with custom analyzer
	contentField := bleve.NewTextFieldMapping()
	contentField.Store = true // Store content for retrieval
	contentField.IncludeInAll = true
	contentField.IncludeTermVectors = true // For highlighting
	contentField.Analyzer = "standard"
	docMapping.AddFieldMappingsAt("content", contentField)

	// Set as default mapping
	mapping.DefaultMapping = docMapping

	// Configure to not index dynamic fields
	mapping.IndexDynamic = false
	mapping.StoreDynamic = false

	return mapping
}

func (m *IndexManager) shouldIndexFile(extra []string, name string, ext string) bool {
	if ext == "" {
		for _, sf := range allowedExtensions {
			if name == sf {
				return true
			}
		}
		for _, sf := range extra {
			if name == sf {
				return true
			}
		}
		return false
	}
	return defaultExtensions[ext]
}
