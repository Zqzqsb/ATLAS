package services

import (
	"context"
)

// PassthroughTranslator is a no-op translator that returns original text.
// Previously lived in bridge/context_bridge.go as TranslatorBridge.
// TODO: implement real LLM-based translation when needed.
type PassthroughTranslator struct{}

func NewPassthroughTranslator() *PassthroughTranslator {
	return &PassthroughTranslator{}
}

func (t *PassthroughTranslator) TranslateTexts(_ context.Context, texts []string, _ string) (map[string]string, error) {
	result := make(map[string]string, len(texts))
	for _, text := range texts {
		result[text] = text
	}
	return result, nil
}
