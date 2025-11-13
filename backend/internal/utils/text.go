package utils

import (
	"strings"
)

// ChunkText splits text into chunks with overlap
func ChunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	chunks := []string{}
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize
	}

	for i := 0; i < len(words); i += step {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		if end == len(words) {
			break
		}
	}

	return chunks
}

// EstimateTokens estimates token count (rough: 1 token ≈ 4 characters)
func EstimateTokens(text string) int {
	return len(text) / 4
}

// TruncateText truncates text to max tokens
func TruncateText(text string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "..."
}
