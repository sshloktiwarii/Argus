package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestServer() *Server {
	b := &EventBatcher{
		events:        make([]LLMEvent, 0, 1000),
		maxEvents:     1000,
		flushInterval: 1 * time.Hour,
		ticker:        time.NewTicker(1 * time.Hour),
		done:          make(chan bool),
	}
	return &Server{batcher: b}
}

func TestHandleHealth(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if resp["status"] != "healthy" {
		t.Fatalf("expected status healthy, got %s", resp["status"])
	}
}

func TestHandleEvent_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/event", nil)
	w := httptest.NewRecorder()

	s.handleEvent(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestHandleEvent_Valid(t *testing.T) {
	s := newTestServer()
	event := LLMEvent{
		TraceID:          uuid.New(),
		SpanID:           uuid.New(),
		AgentID:          "sentinel-test",
		Model:            "claude-3-5-sonnet",
		Provider:         "anthropic",
		PromptTokens:     100,
		CompletionTokens: 50,
		CostUSD:          0.0015,
		LatencyMs:        450,
		Status:           "success",
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handleEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if len(s.batcher.events) != 1 {
		t.Fatalf("expected 1 event in batcher, got %d", len(s.batcher.events))
	}
}

func TestHandleEvent_MissingTraceID(t *testing.T) {
	s := newTestServer()
	event := LLMEvent{
		SpanID:  uuid.New(),
		AgentID: "test",
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handleEvent(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing trace_id, got %d", w.Code)
	}
}

func TestHandleEvent_MissingSpanID(t *testing.T) {
	s := newTestServer()
	event := LLMEvent{
		TraceID: uuid.New(),
		AgentID: "test",
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handleEvent(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for missing span_id, got %d", w.Code)
	}
}

func TestHandleBatch_Valid(t *testing.T) {
	s := newTestServer()
	events := []LLMEvent{
		{TraceID: uuid.New(), SpanID: uuid.New(), AgentID: "agent-1"},
		{TraceID: uuid.New(), SpanID: uuid.New(), AgentID: "agent-2"},
	}
	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handleBatch(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	if len(s.batcher.events) != 2 {
		t.Fatalf("expected 2 events in batcher, got %d", len(s.batcher.events))
	}
}

func TestHandleIngest_Normalization(t *testing.T) {
	s := newTestServer()
	payload := IngestPayload{
		Tokens:     100,
		Cost:       0.005,
		StatusCode: 200,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handleIngest(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	if len(s.batcher.events) != 1 {
		t.Fatalf("expected 1 event recorded, got %d", len(s.batcher.events))
	}

	e := s.batcher.events[0]
	if e.AgentID != "default-agent" {
		t.Errorf("expected default-agent, got %s", e.AgentID)
	}
	if e.Model != "unknown-model" {
		t.Errorf("expected unknown-model, got %s", e.Model)
	}
	if e.Provider != "local" {
		t.Errorf("expected local provider, got %s", e.Provider)
	}
	if e.CostUSD != 0.005 {
		t.Errorf("expected cost 0.005, got %f", e.CostUSD)
	}
	if e.Status != "success" {
		t.Errorf("expected success status, got %s", e.Status)
	}
	if e.PromptTokens != 75 || e.CompletionTokens != 25 {
		t.Errorf("expected 75/25 token split, got %d/%d", e.PromptTokens, e.CompletionTokens)
	}
}
