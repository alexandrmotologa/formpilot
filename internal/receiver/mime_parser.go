package receiver

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"
)

type ParsedEmail struct {
	Sender      string            `json:"sender"`
	Recipient   string            `json:"recipient"`
	Subject     string            `json:"subject"`
	Date        string            `json:"date"`
	MessageID   string            `json:"message_id"`
	Headers     map[string]string `json:"headers"`
	TextBody    string            `json:"text_body"`
	HTMLBody    string            `json:"html_body"`
	Combined    string            `json:"combined"`
	Attachments []AttachmentInfo  `json:"attachments"`
}

type AttachmentInfo struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	TextContent string `json:"text_content,omitempty"`
}

// ParseRawEmail decodes a raw RFC 2822 email payload.
func ParseRawEmail(raw []byte) (*ParsedEmail, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		// If not a full RFC message with headers, treat as raw plain text
		return &ParsedEmail{
			TextBody: string(raw),
			Combined: string(raw),
		}, nil
	}

	result := &ParsedEmail{
		Sender:    msg.Header.Get("From"),
		Recipient: msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		Date:      msg.Header.Get("Date"),
		MessageID: msg.Header.Get("Message-ID"),
		Headers:   make(map[string]string),
	}

	for k, v := range msg.Header {
		if len(v) > 0 {
			result.Headers[k] = v[0]
		}
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		body, _ := io.ReadAll(msg.Body)
		result.TextBody = string(body)
		result.Combined = result.TextBody
		return result, nil
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		// Single part message
		body, _ := io.ReadAll(msg.Body)
		if strings.Contains(mediaType, "html") {
			result.HTMLBody = string(body)
			result.TextBody = StripHTML(result.HTMLBody)
		} else {
			result.TextBody = string(body)
		}
		result.Combined = result.TextBody
		return result, nil
	}

	mr := multipart.NewReader(msg.Body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		partType, partParams, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
		disposition, dispParams, _ := mime.ParseMediaType(part.Header.Get("Content-Disposition"))

		filename := dispParams["filename"]
		if filename == "" {
			filename = partParams["name"]
		}

		partBytes, _ := io.ReadAll(part)

		if disposition == "attachment" || filename != "" {
			att := AttachmentInfo{
				Filename:    filename,
				ContentType: partType,
				Size:        len(partBytes),
			}

			// If PDF, attempt text extraction
			if strings.HasSuffix(strings.ToLower(filename), ".pdf") || partType == "application/pdf" {
				if pdfText, err := ExtractTextFromPDF(partBytes); err == nil && pdfText != "" {
					att.TextContent = pdfText
				}
			} else if strings.HasSuffix(strings.ToLower(filename), ".txt") {
				att.TextContent = string(partBytes)
			}

			result.Attachments = append(result.Attachments, att)
		} else if strings.Contains(partType, "text/plain") {
			result.TextBody = string(partBytes)
		} else if strings.Contains(partType, "text/html") {
			result.HTMLBody = string(partBytes)
		}
	}

	if result.TextBody == "" && result.HTMLBody != "" {
		result.TextBody = StripHTML(result.HTMLBody)
	}

	// Build combined text with attachment notes
	var combined strings.Builder
	if result.TextBody != "" {
		combined.WriteString(result.TextBody)
	}
	for _, att := range result.Attachments {
		if att.TextContent != "" {
			combined.WriteString("\n\n--- Attachment: " + att.Filename + " ---\n")
			combined.WriteString(att.TextContent)
		}
	}

	result.Combined = strings.TrimSpace(combined.String())
	return result, nil
}

// StripHTML converts simple HTML tags to readable plaintext.
func StripHTML(html string) string {
	// Replace break and paragraph tags with newlines
	reBreaks := regexp.MustCompile(`(?i)<(?:br|p|/p|div|/div|tr|/tr)\s*/?>`)
	text := reBreaks.ReplaceAllString(html, "\n")

	// Strip remaining HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)
	text = reTags.ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// Collapse repeated empty lines
	reMultipleLines := regexp.MustCompile(`\n{3,}`)
	text = reMultipleLines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// DecodeBase64 decodes standard base64 or returns original string if not valid base64.
func DecodeBase64(data string) []byte {
	clean := strings.TrimSpace(data)
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return []byte(data)
	}
	return decoded
}
