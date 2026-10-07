---
id: 1
title: Fix issues board clipping in growing columns
status: closed
priority: high
labels:
    - bug
created: "2026-10-07"
updated: "2026-10-07"
closed: "2026-10-07"
---

## Context
The Kanban board estimates every issue card as four terminal lines, but cards with labels take an additional line. The board also inserts top/bottom scroll indicators outside the card-height estimate. Once a column grows, rendered content can exceed the fixed column height and get clipped.

## Success Criteria
- [x] Column rendering respects the requested height for labeled and unlabeled cards, with or without scroll indicators.
- [x] The selected issue remains visible while navigating long columns.
- [x] Regression tests cover clipping at multiple positions in a growing column.
- [x] `go test ./...` and `go build ./...` pass.

## Implementation
1. Replace fixed approximate card-height windowing with a calculation based on the rendered card heights and available rows.
2. Account for column header/borders and optional top/bottom scroll indicators when selecting the visible issue range.
3. Add tests that assert rendered columns fit their assigned height and preserve cursor visibility.

## Affected Files
- `internal/tui/column.go` — height-aware card windowing and scroll indicator budgeting.
- `internal/tui/board_test.go` — regression coverage for rendering/clipping.

## Verification
- `go test ./...` — passed.
- `go build ./...` — passed.
- `git diff --check` — passed.
- Focused regression asserts fixed requested height, selected-card visibility, and scroll indicators across top/middle/bottom cursor positions.
