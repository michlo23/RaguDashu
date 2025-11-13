package services

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFService handles PDF text extraction
type PDFService struct{}

// NewPDFService creates a new PDF service
func NewPDFService() *PDFService {
	return &PDFService{}
}

// ExtractText extracts text from PDF content
func (s *PDFService) ExtractText(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	// Read PDF
	pdfReader, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Extract text from all pages
	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Continue with other pages even if one fails
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	extractedText := textBuilder.String()
	if strings.TrimSpace(extractedText) == "" {
		return "", fmt.Errorf("no text content found in PDF")
	}

	return extractedText, nil
}

// ExtractMetadata extracts PDF metadata
func (s *PDFService) ExtractMetadata(content []byte) (map[string]string, error) {
	reader := bytes.NewReader(content)
	pdfReader, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]string)

	// Get number of pages
	metadata["pages"] = fmt.Sprintf("%d", pdfReader.NumPage())

	// Try to get document info
	if info := pdfReader.Trailer().Key("Info"); !info.IsNull() {
		if title := info.Key("Title"); !title.IsNull() {
			metadata["title"] = title.String()
		}
		if author := info.Key("Author"); !author.IsNull() {
			metadata["author"] = author.String()
		}
		if subject := info.Key("Subject"); !subject.IsNull() {
			metadata["subject"] = subject.String()
		}
		if keywords := info.Key("Keywords"); !keywords.IsNull() {
			metadata["keywords"] = keywords.String()
		}
		if creator := info.Key("Creator"); !creator.IsNull() {
			metadata["creator"] = creator.String()
		}
		if producer := info.Key("Producer"); !producer.IsNull() {
			metadata["producer"] = producer.String()
		}
	}

	return metadata, nil
}

// ExtractTextFromReader extracts text from io.Reader
func (s *PDFService) ExtractTextFromReader(r io.Reader) (string, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return s.ExtractText(content)
}
