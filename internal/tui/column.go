package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/steviee/git-issues/internal/issue"
)

// Column represents a status column on the board.
type Column struct {
	Status string
	Issues []*issue.Issue
	Cursor int
	Offset int // scroll offset for windowing
}

// NewColumn creates a column for the given status.
func NewColumn(status string) *Column {
	return &Column{
		Status: status,
		Issues: nil,
		Cursor: 0,
		Offset: 0,
	}
}

// Add adds an issue to the column in sorted position.
func (c *Column) Add(iss *issue.Issue) {
	c.Issues = append(c.Issues, iss)
	c.sortIssues()
}

// Remove removes an issue by ID and returns it. Returns nil if not found.
func (c *Column) Remove(id int) *issue.Issue {
	for i, iss := range c.Issues {
		if iss.ID == id {
			removed := c.Issues[i]
			c.Issues = append(c.Issues[:i], c.Issues[i+1:]...)
			// Adjust cursor
			if c.Cursor >= len(c.Issues) && c.Cursor > 0 {
				c.Cursor = len(c.Issues) - 1
			}
			c.clampOffset()
			return removed
		}
	}
	return nil
}

// Selected returns the currently selected issue, or nil if empty.
func (c *Column) Selected() *issue.Issue {
	if len(c.Issues) == 0 {
		return nil
	}
	if c.Cursor < 0 || c.Cursor >= len(c.Issues) {
		return nil
	}
	return c.Issues[c.Cursor]
}

// CursorDown moves the cursor down.
func (c *Column) CursorDown() {
	if c.Cursor < len(c.Issues)-1 {
		c.Cursor++
		c.clampOffset()
	}
}

// CursorUp moves the cursor up.
func (c *Column) CursorUp() {
	if c.Cursor > 0 {
		c.Cursor--
		c.clampOffset()
	}
}

func (c *Column) clampOffset() {
	// Will be called with visible count from Render
}

// Render renders the column with the given dimensions.
func (c *Column) Render(width, height int, active bool) string {
	// Header
	title := fmt.Sprintf(" %s (%d) ", strings.ToUpper(c.Status), len(c.Issues))
	headerStyle := columnHeaderStyle
	if active {
		headerStyle = activeColumnHeaderStyle
	}
	header := headerStyle.Width(width - 2).Render(title)

	// Determine how many rows remain after the header and outer border.
	innerWidth := width - 2 // column border
	availableLines := height - 2 - lipgloss.Height(header)
	if availableLines < 0 {
		availableLines = 0
	}

	// Measure the rendered cards instead of assuming a fixed height. Labels can
	// wrap to additional lines, and the selected-card style can affect rendering.
	cardRenders := make([]string, len(c.Issues))
	cardLines := make([]int, len(c.Issues))
	for i, iss := range c.Issues {
		cardRenders[i] = RenderCard(iss, innerWidth, active && i == c.Cursor)
		cardLines[i] = lipgloss.Height(cardRenders[i])
	}

	var cards []string
	start, end := 0, 0
	if len(c.Issues) > 0 {
		cursor := c.Cursor
		if cursor < 0 {
			cursor = 0
		} else if cursor >= len(c.Issues) {
			cursor = len(c.Issues) - 1
		}

		preferredStart := c.Offset
		if preferredStart < 0 {
			preferredStart = 0
		}
		if preferredStart > cursor {
			preferredStart = cursor
		}

		// Keep the current window start when it still fits the cursor. If not,
		// advance it until the selected card fits, then fill the remaining rows.
		for candidateStart := preferredStart; candidateStart <= cursor; candidateStart++ {
			candidateLines := 0
			for i := candidateStart; i <= cursor; i++ {
				candidateLines += cardLines[i]
			}

			for candidateEnd := cursor + 1; candidateEnd <= len(c.Issues); candidateEnd++ {
				if candidateEnd > cursor+1 {
					candidateLines += cardLines[candidateEnd-1]
				}
				bodyLines := candidateLines
				elements := candidateEnd - candidateStart
				if candidateStart > 0 {
					bodyLines++
					elements++
				}
				if candidateEnd < len(c.Issues) {
					bodyLines++
					elements++
				}
				if elements > 1 {
					bodyLines += elements - 1
				}
				if bodyLines > availableLines {
					break
				}
				start, end = candidateStart, candidateEnd
			}
			if end > start {
				break
			}
		}

		// If even the selected card is taller than the available body, keep it
		// visible rather than dropping the selection altogether.
		if end == start {
			start, end = cursor, cursor+1
		}
		c.Offset = start

		if start > 0 {
			cards = append(cards, fmt.Sprintf(" ↑ %d more", start))
		}
		for i := start; i < end; i++ {
			cards = append(cards, cardRenders[i])
		}
		if end < len(c.Issues) {
			cards = append(cards, fmt.Sprintf(" ↓ %d more", len(c.Issues)-end))
		}
	}

	body := strings.Join(cards, "\n")

	// Pad body to fill the column; the card window above already includes
	// indicator rows and inter-card separators in its height budget.
	bodyLines := lipgloss.Height(body)
	for bodyLines < availableLines {
		body += "\n"
		bodyLines++
	}

	content := header + "\n" + body

	// Lipgloss Height excludes the border, while height is the total column
	// height requested by the board. Reserve the two border rows here.
	contentHeight := height - 2
	if contentHeight < 0 {
		contentHeight = 0
	}
	colStyle := columnStyle.Width(width).Height(contentHeight)
	if active {
		colStyle = activeColumnStyle.Width(width).Height(contentHeight)
	}

	return colStyle.Render(content)
}

func (c *Column) sortIssues() {
	sort.SliceStable(c.Issues, func(i, j int) bool {
		ri := issue.PriorityRank(c.Issues[i].Priority)
		rj := issue.PriorityRank(c.Issues[j].Priority)
		if ri != rj {
			return ri > rj
		}
		return c.Issues[i].ID < c.Issues[j].ID
	})
}
