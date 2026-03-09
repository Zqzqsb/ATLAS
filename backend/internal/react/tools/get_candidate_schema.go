package tools

import (
	"context"
	"sync/atomic"
	"time"
)

// GetCandidateSchema is a ReAct tool that reads candidate schema from a shared
// memory slot. Used for cold-start acceleration: the ReAct engine starts in
// parallel with vector retrieval. When the LLM calls this tool, it either gets
// the schema (if retrieval is done) or an empty response (if still waiting).
//
// The LLM is instructed to call this tool first; if it returns empty, the LLM
// naturally loops back (Thought → Action → Observation cycle) until data arrives.
type GetCandidateSchema struct {
	slot      *atomic.Pointer[string]
	callCount int
	// firstDataTime records the timestamp when non-empty data was first returned (T1.1).
	// Zero value means data has not been delivered yet.
	firstDataTime time.Time
}

// NewGetCandidateSchema creates a new tool backed by a shared atomic slot.
// The caller writes schema data into the slot when vector retrieval completes.
func NewGetCandidateSchema(slot *atomic.Pointer[string]) *GetCandidateSchema {
	return &GetCandidateSchema{slot: slot}
}

func (t *GetCandidateSchema) Name() string { return "get_candidate_schema" }
func (t *GetCandidateSchema) Description() string {
	return `Retrieve candidate database schema for analysis.
Call this tool FIRST before analyzing tables. If the result is empty (schema not yet available), wait and call again.
Input: ignored (no input needed).
Output: candidate table schemas with columns, types, descriptions, and vector relevance scores. Empty string means data is not ready yet.`
}

func (t *GetCandidateSchema) Call(_ context.Context, _ string) (string, error) {
	t.callCount++
	data := t.slot.Load()
	if data == nil {
		return "", nil
	}
	// Record T1.1: first time we deliver non-empty schema data to the LLM
	if t.firstDataTime.IsZero() {
		t.firstDataTime = time.Now()
	}
	return *data, nil
}

func (t *GetCandidateSchema) CallCount() int { return t.callCount }

// FirstDataTime returns the timestamp when non-empty data was first returned (T1.1).
// Returns zero time if data has never been delivered.
func (t *GetCandidateSchema) FirstDataTime() time.Time { return t.firstDataTime }
