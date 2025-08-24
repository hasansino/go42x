package claudehook

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TranscriptEntry represents a single entry in the transcript
type TranscriptEntry struct {
	Type    string                 `json:"type"`
	Message map[string]interface{} `json:"message"`
}

// TranscriptParser provides utilities for parsing Claude transcripts
type TranscriptParser struct{}

// NewTranscriptParser creates a new transcript parser
func NewTranscriptParser() *TranscriptParser {
	return &TranscriptParser{}
}

// ExtractLastAssistantMessage extracts the last assistant message from a transcript file
func (p *TranscriptParser) ExtractLastAssistantMessage(transcriptPath string) (string, error) {
	if transcriptPath == "" {
		return "", fmt.Errorf("transcript path is empty")
	}

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}

	return p.ExtractLastAssistantMessageFromData(data)
}

// ExtractLastAssistantMessageFromData extracts the last assistant message from transcript data
func (p *TranscriptParser) ExtractLastAssistantMessageFromData(data []byte) (string, error) {
	lines := strings.Split(string(data), "\n")
	lastMessage := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Check if this is an assistant message
		if entry.Type == "assistant" {
			if msg := p.extractMessageContent(&entry); msg != "" {
				lastMessage = msg
			}
		}
	}

	return lastMessage, nil
}

// ExtractAllAssistantMessages extracts all assistant messages from a transcript file
func (p *TranscriptParser) ExtractAllAssistantMessages(transcriptPath string) ([]string, error) {
	if transcriptPath == "" {
		return nil, fmt.Errorf("transcript path is empty")
	}

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var messages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Check if this is an assistant message
		if entry.Type == "assistant" {
			if msg := p.extractMessageContent(&entry); msg != "" {
				messages = append(messages, msg)
			}
		}
	}

	return messages, nil
}

// extractMessageContent extracts text content from a message entry
func (p *TranscriptParser) extractMessageContent(entry *TranscriptEntry) string {
	message, ok := entry.Message["content"].([]interface{})
	if !ok || len(message) == 0 {
		return ""
	}

	var textParts []string
	for _, item := range message {
		if contentMap, ok := item.(map[string]interface{}); ok {
			if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
				if text, ok := contentMap["text"].(string); ok {
					textParts = append(textParts, text)
				}
			}
		}
	}

	return strings.Join(textParts, "\n")
}
