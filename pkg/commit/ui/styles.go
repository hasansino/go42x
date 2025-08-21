package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	primaryColor   = "#00D9FF" // Cyan
	secondaryColor = "#FF6B9D" // Pink
	accentColor    = "#FFE45E" // Yellow
	successColor   = "#6BCF7F" // Green
	warningColor   = "#FF8C42" // Orange
	errorColor     = "#FF5E5B" // Red
	textColor      = "#FFFFFF" // White
	mutedColor     = "#94A3B8" // Gray
	bgColor        = "#0F172A" // Dark blue
	borderColor    = "#334155" // Medium gray
)

var (
	// Minimalistic title
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)).
			Bold(true)

	// Simple provider label
	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(successColor)).
			Bold(true).
			MarginRight(1)

	// Minimal message styling
	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textColor)).
			MarginLeft(2)

	// Selected message
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(bgColor)).
			Background(lipgloss.Color(primaryColor)).
			Bold(true).
			MarginLeft(2)

	// Simple cursor
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(accentColor)).
			Bold(true)

	// Simple input styling
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(textColor)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor)).
			Padding(0, 1)

	// Minimal help text
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedColor)).
			Italic(true)

	// Simple error styling
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(errorColor)).
			Bold(true)

	// Custom option styling
	customStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(warningColor)).
			Bold(true).
			MarginLeft(2)

	// Selected custom option
	selectedCustomStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(bgColor)).
				Background(lipgloss.Color(warningColor)).
				Bold(true).
				MarginLeft(2)
)
