package receiver

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ExtractTextFromPDF decodes text from a raw PDF byte buffer using pure Go stream extraction.
func ExtractTextFromPDF(data []byte) (string, error) {
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		return "", fmt.Errorf("invalid PDF signature")
	}

	var sb strings.Builder

	// Find all streams in PDF: stream ... endstream
	streamRegex := regexp.MustCompile(`(?s)stream\r?\n(.*?)[\r\n]endstream`)
	matches := streamRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		rawStream := match[1]

		// Try zlib FlateDecode decompress
		decompressed, err := decompressFlate(rawStream)
		streamContent := rawStream
		if err == nil && len(decompressed) > 0 {
			streamContent = decompressed
		}

		// Extract text from text blocks: BT ... ET
		extracted := extractTextFromStream(streamContent)
		if len(extracted) > 0 {
			sb.WriteString(extracted)
			sb.WriteString("\n")
		}
	}

	result := strings.TrimSpace(sb.String())
	return result, nil
}

func decompressFlate(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func extractTextFromStream(data []byte) string {
	content := string(data)
	var sb strings.Builder

	// Match text inside parenthesis for Tj and ' operators: (Some Text) Tj
	tjRegex := regexp.MustCompile(`\((.*?)\)\s*(?:Tj|'|")`)
	tjMatches := tjRegex.FindAllStringSubmatch(content, -1)
	for _, m := range tjMatches {
		if len(m) > 1 {
			sb.WriteString(unescapePDFString(m[1]))
			sb.WriteString(" ")
		}
	}

	// Match array TJ operators: [(Part 1) -120 (Part 2)] TJ
	arrayRegex := regexp.MustCompile(`\[(.*?)\]\s*TJ`)
	arrayMatches := arrayRegex.FindAllStringSubmatch(content, -1)
	for _, am := range arrayMatches {
		if len(am) > 1 {
			inner := am[1]
			itemRegex := regexp.MustCompile(`\((.*?)\)`)
			items := itemRegex.FindAllStringSubmatch(inner, -1)
			for _, item := range items {
				if len(item) > 1 {
					sb.WriteString(unescapePDFString(item[1]))
				}
			}
			sb.WriteString(" ")
		}
	}

	return strings.TrimSpace(sb.String())
}

func unescapePDFString(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\(`, "(")
	s = strings.ReplaceAll(s, `\)`, ")")
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}
