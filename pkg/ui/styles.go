package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Minimal/subtle color scheme
	subtle      = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
	highlight   = lipgloss.AdaptiveColor{Light: "#ac64d6ff", Dark: "#752d9fff"}
	success     = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73D788"}
	error       = lipgloss.AdaptiveColor{Light: "#FF4545", Dark: "#FF6B6B"}
	warning     = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FFB84D"}
	userColor   = lipgloss.AdaptiveColor{Light: "#5B5B5B", Dark: "#D4D4D4"}
	aiColor     = lipgloss.AdaptiveColor{Light: "#4A90E2", Dark: "#6BA3E8"}
	borderColor = lipgloss.AdaptiveColor{Light: "#D9D9D9", Dark: "#3A3A3A"}

	// Base styles
	BaseStyle = lipgloss.NewStyle()

	// Header style for section titles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			MarginBottom(1)

	// Subtle text for secondary information
	SubtleStyle = lipgloss.NewStyle().
			Foreground(subtle)

	// Success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(success).
			Bold(true)

	// Error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(error).
			Bold(true)

	// Warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(warning)

	// List item style
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// User message in chat
	UserStyle = lipgloss.NewStyle().
			Foreground(userColor).
			Bold(true)

	// AI message in chat
	AIStyle = lipgloss.NewStyle().
		Foreground(aiColor)

	// Chat message box
	MessageBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			MarginBottom(1)

	// Info box style
	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	// Prompt style
	PromptStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)
)

// FormatHeader formats a header with optional subtitle
func FormatHeader(title string, subtitle ...string) string {
	header := HeaderStyle.Render(title)
	if len(subtitle) > 0 && subtitle[0] != "" {
		header += "\n" + SubtleStyle.Render(subtitle[0])
	}
	return header
}

// FormatListItem formats a list item with bullet point
func FormatListItem(item string) string {
	return ListItemStyle.Render("• " + item)
}

// FormatNestedListItem formats a nested list item with indentation
func FormatNestedListItem(item string) string {
	return lipgloss.NewStyle().PaddingLeft(4).Render("• " + item)
}

// FormatSuccess formats a success message
func FormatSuccess(message string) string {
	return SuccessStyle.Render("✓ " + message)
}

// FormatError formats an error message
func FormatError(message string) string {
	return ErrorStyle.Render("✗ " + message)
}

// FormatWarning formats a warning message
func FormatWarning(message string) string {
	return WarningStyle.Render("⚠ " + message)
}

// FormatUserMessage formats a user message in chat
func FormatUserMessage(message string) string {
	content := UserStyle.Render("You: ") + message
	return MessageBoxStyle.Render(content)
}

// FormatAIMessage formats an AI message in chat
func FormatAIMessage(llmName, message string) string {
	content := AIStyle.Render(llmName+": ") + message
	return MessageBoxStyle.Render(content)
}

// FormatInfoBox formats an informational box
func FormatInfoBox(content string) string {
	return InfoBoxStyle.Render(content)
}

// FormatDivider creates a subtle divider line
func FormatDivider() string {
	return SubtleStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
