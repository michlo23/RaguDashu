package utils

import (
	"strings"
	"testing"
)

func TestChunkText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		chunkSize int
		overlap   int
		wantMin   int
		wantMax   int
	}{
		{
			name:      "Basic chunking",
			text:      "one two three four five six seven eight nine ten",
			chunkSize: 3,
			overlap:   1,
			wantMin:   4,
			wantMax:   5,
		},
		{
			name:      "Empty text",
			text:      "",
			chunkSize: 5,
			overlap:   1,
			wantMin:   0,
			wantMax:   0,
		},
		{
			name:      "Text shorter than chunk size",
			text:      "short text",
			chunkSize: 10,
			overlap:   2,
			wantMin:   1,
			wantMax:   1,
		},
		{
			name:      "No overlap",
			text:      "one two three four five six",
			chunkSize: 2,
			overlap:   0,
			wantMin:   3,
			wantMax:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkText(tt.text, tt.chunkSize, tt.overlap)

			if len(chunks) < tt.wantMin || len(chunks) > tt.wantMax {
				t.Errorf("ChunkText() got %d chunks, want between %d and %d",
					len(chunks), tt.wantMin, tt.wantMax)
			}

			// Verify no chunk is empty
			for i, chunk := range chunks {
				if strings.TrimSpace(chunk) == "" {
					t.Errorf("Chunk %d is empty", i)
				}
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "Empty string",
			text: "",
			want: 0,
		},
		{
			name: "Short text",
			text: "Hello",
			want: 1, // 5 chars / 4 = 1.25 -> 1
		},
		{
			name: "Long text",
			text: strings.Repeat("a", 400),
			want: 100, // 400 / 4 = 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.text)
			if got != tt.want {
				t.Errorf("EstimateTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxTokens int
		wantLen   int
	}{
		{
			name:      "No truncation needed",
			text:      "short",
			maxTokens: 10,
			wantLen:   5,
		},
		{
			name:      "Truncation needed",
			text:      strings.Repeat("a", 1000),
			maxTokens: 10,
			wantLen:   43, // 40 chars + "..."
		},
		{
			name:      "Empty text",
			text:      "",
			maxTokens: 10,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateText(tt.text, tt.maxTokens)
			if len(got) != tt.wantLen {
				t.Errorf("TruncateText() length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
