# Report Backend Contract

The report generator uses sqlc-generated `InspectionReport` data, the project's image records, and explicit reviewer approval.

```go
func (a *App) GenerateInspectionReport(
	reportID int64,
	outputPath string,
) (inspectionreports.Result, error)
```

`Result` returns the generated path, page count, and included approved-damage image count.

## Inclusion Rules

1. A reviewer edits boxes and optional photo notes.
2. `ApproveImageReview` persists that final annotation set and marks the image approved.
3. An approved image with one or more boxes is included in the PDF.
4. An approved image with no boxes is a reviewed no-damage image and is excluded.
5. Editing boxes or notes resets approval.
6. Unapproved images are excluded; the frontend warns about them but does not block generation.

## Output Rules

- The generator makes a high-resolution square center crop with the same framing as the frontend preview.
- It draws the final approved boxes onto temporary image copies without modifying source images.
- Short summaries may place up to two approved-damage photos on page one; longer summaries begin photos on page two.
- Photo notes appear below their numbered image when present.
- The report documents observations and photos only. It does not make coverage, repair-cost, causation, or claim-approval conclusions.
