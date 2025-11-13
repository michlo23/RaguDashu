package services

import (
	"bytes"
	"testing"
)

func TestPDFService_ExtractText(t *testing.T) {
	service := NewPDFService()

	// Note: You would need a real PDF file for actual testing
	// This is a placeholder test structure
	t.Run("Empty PDF", func(t *testing.T) {
		_, err := service.ExtractText([]byte{})
		if err == nil {
			t.Error("Expected error for empty PDF")
		}
	})

	t.Run("Invalid PDF", func(t *testing.T) {
		_, err := service.ExtractText([]byte("not a pdf"))
		if err == nil {
			t.Error("Expected error for invalid PDF")
		}
	})
}

func TestPDFService_ExtractTextFromReader(t *testing.T) {
	service := NewPDFService()

	t.Run("Empty reader", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})
		_, err := service.ExtractTextFromReader(reader)
		if err == nil {
			t.Error("Expected error for empty reader")
		}
	})
}
