package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Custom delegate for multi-line commit messages
type commitDelegate struct {
	styles     list.DefaultItemStyles
	height     int
	spacing    int
	showDetail bool
}

func newCommitDelegate() commitDelegate {
	d := commitDelegate{
		styles:     list.NewDefaultItemStyles(),
		height:     3, // Height for each item
		spacing:    1,
		showDetail: true,
	}

	// Customize styles
	d.styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	d.styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	d.styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	d.styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	d.styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("238"))

	d.styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("236"))

	return d
}

func (d commitDelegate) Height() int {
	return d.height + d.spacing
}

func (d commitDelegate) Spacing() int {
	return d.spacing
}

func (d commitDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d commitDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	commit, ok := item.(CommitItem)
	if !ok {
		return
	}

	var content strings.Builder
	isSelected := index == m.Index()
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied

	// Build title without arrow indicator
	title := commit.Title()

	// Build description
	var desc string
	if commit.provider == "manual" {
		desc = commit.Description()
	} else if isSelected && len(commit.lines) > 1 {
		// Show full multi-line message when selected
		var descLines []string
		for i, line := range commit.lines {
			descLines = append(descLines, line)
			if i >= 10 { // Limit to 5 lines in the list
				descLines = append(descLines, "...")
				break
			}
		}
		desc = strings.Join(descLines, "\n")
	} else {
		// Show only first line when not selected
		firstLine := ""
		if len(commit.lines) > 0 {
			firstLine = commit.lines[0]
			if len(firstLine) > 60 {
				firstLine = firstLine[:57] + "..."
			}
		}
		desc = firstLine
		if len(commit.lines) > 1 {
			desc += fmt.Sprintf(" (+%d lines)", len(commit.lines)-1)
		}
	}

	// Apply styles based on selection state
	if isSelected {
		// Highlight with left border only, no background
		selectedStyle := lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("170")). // Purple accent on left
			PaddingLeft(1).
			PaddingRight(2).
			Width(m.Width() - 8) // Account for list padding

		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

		content.WriteString(titleStyle.Render(title))
		if desc != "" {
			content.WriteString("\n")
			content.WriteString(descStyle.Render(desc))
		}

		_, _ = fmt.Fprint(w, selectedStyle.Render(content.String()))
	} else {
		// Normal items without highlight
		var itemStyle lipgloss.Style

		if isFiltered && m.FilterState() != list.FilterApplied {
			// Dimmed items during filtering
			itemStyle = lipgloss.NewStyle().
				PaddingLeft(3). // Extra padding to align with selected items
				PaddingRight(2)

			titleStyle := d.styles.DimmedTitle
			descStyle := d.styles.DimmedDesc

			content.WriteString(titleStyle.Render(title))
			if desc != "" {
				content.WriteString("\n")
				content.WriteString(descStyle.Render(desc))
			}
		} else {
			// Normal items
			itemStyle = lipgloss.NewStyle().
				PaddingLeft(3). // Extra padding to align with selected items
				PaddingRight(2)

			titleStyle := d.styles.NormalTitle
			descStyle := d.styles.NormalDesc

			content.WriteString(titleStyle.Render(title))
			if desc != "" {
				content.WriteString("\n")
				content.WriteString(descStyle.Render(desc))
			}
		}

		_, _ = fmt.Fprint(w, itemStyle.Render(content.String()))
	}
}
