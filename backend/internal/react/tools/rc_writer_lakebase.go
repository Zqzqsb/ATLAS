package tools

import (
	"context"
	"fmt"
)

// LakebaseRCWriter wraps a set of repository methods to satisfy RCWriter.
// It doesn't depend on the full LakebaseService — only the minimal write interface.
type LakebaseRCWriter struct {
	repo rcRepo
}

// rcRepo is the minimal repository interface needed by LakebaseRCWriter.
type rcRepo interface {
	UpdateTableDescription(ctx context.Context, dsID int64, tableName, description, source string, confidence float64) error
	UpdateColumnDescription(ctx context.Context, dsID int64, tableName, columnName, description, source string, confidence float64) error
	UpdateColumnSampleValues(ctx context.Context, dsID int64, tableName, columnName, sampleValues string) error
	UpdateColumnSynonyms(ctx context.Context, dsID int64, tableName, columnName, synonyms string) error
	UpsertTerm(ctx context.Context, dsID int64, term, definition, synonyms, examples, category string) error
}

// NewLakebaseRCWriter creates a RCWriter backed by lakebase repository.
func NewLakebaseRCWriter(repo rcRepo) *LakebaseRCWriter {
	if repo == nil {
		panic("react/tools: repo must not be nil")
	}
	return &LakebaseRCWriter{repo: repo}
}

func (w *LakebaseRCWriter) SetTableDescription(ctx context.Context, dsID int64, tableName, description string) error {
	if err := w.repo.UpdateTableDescription(ctx, dsID, tableName, description, "react_agent", 0.85); err != nil {
		return fmt.Errorf("set table description: %w", err)
	}
	return nil
}

func (w *LakebaseRCWriter) SetColumnDescription(ctx context.Context, dsID int64, tableName, columnName, description string) error {
	if err := w.repo.UpdateColumnDescription(ctx, dsID, tableName, columnName, description, "react_agent", 0.85); err != nil {
		return fmt.Errorf("set column description: %w", err)
	}
	return nil
}

func (w *LakebaseRCWriter) SetColumnSampleValues(ctx context.Context, dsID int64, tableName, columnName, sampleValues string) error {
	if err := w.repo.UpdateColumnSampleValues(ctx, dsID, tableName, columnName, sampleValues); err != nil {
		return fmt.Errorf("set column sample_values: %w", err)
	}
	return nil
}

func (w *LakebaseRCWriter) SetColumnSynonyms(ctx context.Context, dsID int64, tableName, columnName, synonyms string) error {
	if err := w.repo.UpdateColumnSynonyms(ctx, dsID, tableName, columnName, synonyms); err != nil {
		return fmt.Errorf("set column synonyms: %w", err)
	}
	return nil
}

func (w *LakebaseRCWriter) AddBusinessTerm(ctx context.Context, dsID int64, term, definition, synonyms, examples, category string) error {
	if err := w.repo.UpsertTerm(ctx, dsID, term, definition, synonyms, examples, category); err != nil {
		return fmt.Errorf("add business term: %w", err)
	}
	return nil
}
