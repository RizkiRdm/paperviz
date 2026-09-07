package services

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

func TestChartFailureCategories(t *testing.T) {
	tests := []struct {
		name     string
		category ChartFailureCategory
		want     string
	}{
		{"extraction error", FailureExtractionError, "EXTRACTION_ERROR"},
		{"dataset error", FailureDatasetError, "DATASET_ERROR"},
		{"chart selection error", FailureChartSelectionError, "CHART_SELECTION_ERROR"},
		{"grounding error", FailureGroundingError, "GROUNDING_ERROR"},
		{"schema error", FailureSchemaError, "SCHEMA_ERROR"},
		{"render error", FailureRenderError, "RENDER_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.want {
				t.Errorf("category = %q, want %q", string(tt.category), tt.want)
			}
		})
	}
}

func TestLogChartFailureFields(t *testing.T) {
	ctx := context.Background()
	var logged bool
	var logMsg string
	var logAttrs []slog.Attr

	handler := &captureHandler{
		before: func(msg string, attrs []slog.Attr) {
			logged = true
			logMsg = msg
			logAttrs = attrs
		},
	}

	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(slog.New(slog.Default().Handler()))

	logChartFailure(ctx, FailureSchemaError, "test_stage", "test_chapter", "test_ds", fmt.Errorf("test error"), true)

	if !logged {
		t.Fatal("logChartFailure did not log")
	}
	if logMsg != "chart failure" {
		t.Errorf("message = %q, want 'chart failure'", logMsg)
	}

	attrs := attrsToMap(logAttrs)
	if attrs["error_category"] != "SCHEMA_ERROR" {
		t.Errorf("error_category = %q, want SCHEMA_ERROR", attrs["error_category"])
	}
	if attrs["stage"] != "test_stage" {
		t.Errorf("stage = %q, want test_stage", attrs["stage"])
	}
	if attrs["chapter"] != "test_chapter" {
		t.Errorf("chapter = %q, want test_chapter", attrs["chapter"])
	}
	if attrs["dataset_id"] != "test_ds" {
		t.Errorf("dataset_id = %q, want test_ds", attrs["dataset_id"])
	}
	if attrs["fallback_used"] != "true" {
		t.Errorf("fallback_used = %q, want true", attrs["fallback_used"])
	}
	if attrs["safe_error_message"] != "test error" {
		t.Errorf("safe_error_message = %q, want 'test error'", attrs["safe_error_message"])
	}
}

func attrsToMap(attrs []slog.Attr) map[string]string {
	m := make(map[string]string)
	for _, a := range attrs {
		m[a.Key] = a.Value.String()
	}
	return m
}

type captureHandler struct {
	before func(msg string, attrs []slog.Attr)
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	if h.before != nil {
		h.before(r.Message, attrs)
	}
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }
