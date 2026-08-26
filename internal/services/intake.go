package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"paperviz/internal/external"
	"paperviz/internal/repository"
)

// IntakeResult holds data after successful validation and initial DB insertion.
type IntakeResult struct {
	DocumentID   string
	SourceType   string
	OriginalText string
	PDFBytes     []byte
}

// ValidateAndInsert validates the input (PDF file or pasted text), extracts text,
// generates a unique ID, and inserts the initial document row with status 'processing'.
// This is synchronous (Acceptance Scenario 1 & Failure Scenario 1).
func ValidateAndInsert(db *sql.DB, readingLevel string, hasFile bool, pdfBytes []byte, pastedText string, userID *string) (IntakeResult, string, error) {
	var originalText, sourceType string

	if hasFile {
		sourceType = repository.SourceTypePDF
		var err error
		originalText, err = external.ExtractText(pdfBytes)
		if err != nil {
			if errors.Is(err, external.ErrNoTextLayer) {
				return IntakeResult{}, "no_text_layer", err
			}
			return IntakeResult{}, "invalid_file_type", err
		}
	} else {
		sourceType = repository.SourceTypePastedText
		originalText = pastedText
	}

	id, err := repository.NewID()
	if err != nil {
		return IntakeResult{}, "internal_error", err
	}

	now := time.Now().Unix()
	doc := repository.Document{
		ID:             id,
		CreatedAt:      now,
		LastAccessedAt: now,
		Status:         repository.StatusProcessing,
		SourceType:     sourceType,
		ReadingLevel:   readingLevel,
		Title:          deriveTitle(originalText),
		OriginalText:   originalText,
		UserID:         userID,
	}

	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(doc); err != nil {
		return IntakeResult{}, "internal_error", err
	}

	return IntakeResult{
		DocumentID:   id,
		SourceType:   sourceType,
		OriginalText: originalText,
		PDFBytes:     pdfBytes,
	}, "", nil
}

const backgroundPipelineTimeout = 20 * time.Minute

// RunPipelineAndPersist runs the full asynchronous pipeline and persists the result
// in a single transaction (ARCHITECTURE.md Section 4 Transaction Policy).
func RunPipelineAndPersist(db *sql.DB, gemini *external.GeminiClient, documentID string, input PipelineInput) {
	ctx, cancel := context.WithTimeout(context.Background(), backgroundPipelineTimeout)
	defer cancel()

	startTime := time.Now()

	input.OnStage = func(stage string) {
		docRepo := repository.NewDocumentRepo(db)
		s := stage
		if err := docRepo.UpdateStage(documentID, &s); err != nil {
			slog.Error("update processing stage failed", "document_id", documentID, "stage", stage, "error", err)
		}
	}

	output := RunPipeline(ctx, gemini, input)

	if err := savePipelineResult(db, documentID, output); err != nil {
		slog.Error("save pipeline result failed", "document_id", documentID, "error", err)
	}

	// Record elapsed processing time in ms for usage measurement (best-effort).
	elapsed := int(time.Since(startTime).Milliseconds())
	docRepo := repository.NewDocumentRepo(db)
	_ = docRepo.SetProcessingTime(documentID, elapsed)
}

func savePipelineResult(db *sql.DB, documentID string, output PipelineOutput) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	docRepo := repository.NewDocumentRepo(tx)

	if output.Status == repository.StatusFailed {
		errMsg := output.ErrorMessage
		if err := docRepo.UpdateStatus(documentID, repository.StatusFailed, nil, &errMsg, false); err != nil {
			return err
		}
		return tx.Commit()
	}

	simplified := output.SimplifiedText
	if err := docRepo.UpdateStatus(documentID, output.Status, &simplified, nil, output.ChartExtractionDegraded); err != nil {
		return err
	}

	claimDiffRepo := repository.NewClaimDiffRepo(tx)
	originalClaimsJSON, err := json.Marshal(output.Verify.OriginalClaims)
	if err != nil {
		return fmt.Errorf("marshal original claims: %w", err)
	}
	simplifiedClaimsJSON, err := json.Marshal(output.Verify.SimplifiedClaims)
	if err != nil {
		return fmt.Errorf("marshal simplified claims: %w", err)
	}
	claimDiffID, err := repository.NewID()
	if err != nil {
		return fmt.Errorf("generate claim_diff id: %w", err)
	}
	detail := output.Verify.MismatchDetail
	if err := claimDiffRepo.Insert(repository.ClaimDiff{
		ID:               claimDiffID,
		DocumentID:       documentID,
		OriginalClaims:   string(originalClaimsJSON),
		SimplifiedClaims: string(simplifiedClaimsJSON),
		MismatchDetected: output.Verify.MismatchDetected,
		MismatchDetail:   &detail,
	}); err != nil {
		return err
	}

	chapterRepo := repository.NewChapterRepo(tx)
	chapterIndexToID := make(map[int]string, len(output.Chapters))
	for i, ch := range output.Chapters {
		chapterID, err := repository.NewID()
		if err != nil {
			return fmt.Errorf("generate chapter id: %w", err)
		}
		if err := chapterRepo.Insert(repository.Chapter{
			ID:           chapterID,
			DocumentID:   documentID,
			Title:        ch.Title,
			Summary:      ch.Summary,
			Excerpt:      ch.Excerpt,
			DisplayOrder: i,
		}); err != nil {
			return err
		}
		chapterIndexToID[i] = chapterID
	}

	chartRepo := repository.NewChartRepo(tx)
	evidenceRepo := repository.NewEvidenceRepo(tx)
	for _, c := range output.Charts {
		chartID, err := repository.NewID()
		if err != nil {
			return fmt.Errorf("generate chart id: %w", err)
		}
		var chartDataPtr, annotationPtr *string
		if c.ChartData != "" {
			chartDataPtr = &c.ChartData
		}
		if c.Annotation != "" {
			annotationPtr = &c.Annotation
		}
		pageNum := c.PageNumber
		var chapterIDPtr *string
		if c.ChapterIndex >= 0 {
			if id, ok := chapterIndexToID[c.ChapterIndex]; ok {
				chapterIDPtr = &id
			}
		}
		if err := chartRepo.Insert(repository.Chart{
			ID:           chartID,
			DocumentID:   documentID,
			SourceMethod: c.SourceMethod,
			ChartData:    chartDataPtr,
			ImageBlob:    c.ImageBlob,
			Annotation:   annotationPtr,
			PageNumber:   &pageNum,
			DisplayOrder: c.DisplayOrder,
			ChapterID:    chapterIDPtr,
		}); err != nil {
			return err
		}

		if c.PageNumber > 0 && strings.TrimSpace(c.SourceText) != "" {
			evidenceID, err := repository.NewID()
			if err != nil {
				return fmt.Errorf("generate evidence id: %w", err)
			}
			reference := fmt.Sprintf("Figure on page %d", c.PageNumber)
			if err := evidenceRepo.Insert(repository.Evidence{
				ID:              evidenceID,
				PaperID:         documentID,
				Page:            &pageNum,
				FigureID:        &chartID,
				SourceText:      c.SourceText,
				SourceReference: reference,
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func deriveTitle(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > 200 {
				return line[:200]
			}
			return line
		}
	}
	return "Untitled paper"
}
