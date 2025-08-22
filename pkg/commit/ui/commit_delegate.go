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

func (d commitDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d commitDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	commit, ok := item.(CommitItem)
	if !ok {
		return
	}

	var title, desc string
	var titleStyle, descStyle lipgloss.Style

	isSelected := index == m.Index()
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied

	// Determine styles based on state
	if isSelected {
		titleStyle = d.styles.SelectedTitle
		descStyle = d.styles.SelectedDesc
	} else if isFiltered && m.FilterState() != list.FilterApplied {
		titleStyle = d.styles.DimmedTitle
		descStyle = d.styles.DimmedDesc
	} else {
		titleStyle = d.styles.NormalTitle
		descStyle = d.styles.NormalDesc
	}

	// Format title
	if isSelected {
		title = "> " + commit.Title()
	} else {
		title = "  " + commit.Title()
	}

	// Format description - show multi-line for selected items
	if commit.provider == "manual" {
		desc = "    " + commit.Description()
	} else if isSelected && len(commit.lines) > 1 {
		// Show full multi-line message when selected
		var descLines []string
		for i, line := range commit.lines {
			if i == 0 {
				descLines = append(descLines, "    "+line)
			} else {
				descLines = append(descLines, "    "+line)
			}
			if i >= 4 { // Limit to 5 lines in the list
				descLines = append(descLines, "    ...")
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
		desc = "    " + firstLine
		if len(commit.lines) > 1 {
			desc += fmt.Sprintf(" (+%d lines)", len(commit.lines)-1)
		}
	}

	// Render
	_, _ = fmt.Fprintf(w, "%s\n%s", titleStyle.Render(title), descStyle.Render(desc))
}
