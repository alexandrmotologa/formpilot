package receiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type InboundPayload struct {
	Source      string            `json:"source"`
	Sender      string            `json:"sender"`
	Recipient   string            `json:"recipient"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	RawPayload  string            `json:"raw_payload"`
	Attachments []AttachmentInfo  `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// ParseSendGridWebhook parses a SendGrid Inbound Parse HTTP request.
func ParseSendGridWebhook(r *http.Request) (*InboundPayload, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("parse multipart form: %w", err)
	}

	from := r.FormValue("from")
	to := r.FormValue("to")
	subject := r.FormValue("subject")
	text := r.FormValue("text")
	html := r.FormValue("html")

	body := text
	if body == "" && html != "" {
		body = StripHTML(html)
	}

	payload := &InboundPayload{
		Source:     "sendgrid",
		Sender:     from,
		Recipient:  to,
		Subject:    subject,
		Body:       body,
		RawPayload: fmt.Sprintf("from: %s\nto: %s\nsubject: %s\n\n%s", from, to, subject, body),
		Headers:    make(map[string]string),
	}

	// Parse file attachments
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, files := range r.MultipartForm.File {
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}
				content, _ := io.ReadAll(file)
				_ = file.Close()

				att := AttachmentInfo{
					Filename:    fileHeader.Filename,
					ContentType: fileHeader.Header.Get("Content-Type"),
					Size:        len(content),
				}

				if strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf") {
					if pdfText, err := ExtractTextFromPDF(content); err == nil && pdfText != "" {
						att.TextContent = pdfText
						payload.Body += "\n\n--- Attachment: " + fileHeader.Filename + " ---\n" + pdfText
					}
				}

				payload.Attachments = append(payload.Attachments, att)
			}
		}
	}

	return payload, nil
}

// PostmarkInbound represents the Postmark inbound webhook JSON payload.
type PostmarkInbound struct {
	FromName        string `json:"FromName"`
	From            string `json:"From"`
	To              string `json:"To"`
	Subject         string `json:"Subject"`
	TextBody        string `json:"TextBody"`
	HtmlBody        string `json:"HtmlBody"`
	Attachments     []struct {
		Name            string `json:"Name"`
		Content         string `json:"Content"` // base64
		ContentType     string `json:"ContentType"`
		ContentLength   int    `json:"ContentLength"`
	} `json:"Attachments"`
}

// ParsePostmarkWebhook parses a Postmark Inbound JSON payload.
func ParsePostmarkWebhook(bodyBytes []byte) (*InboundPayload, error) {
	var pm PostmarkInbound
	if err := json.Unmarshal(bodyBytes, &pm); err != nil {
		return nil, fmt.Errorf("unmarshal postmark payload: %w", err)
	}

	body := pm.TextBody
	if body == "" && pm.HtmlBody != "" {
		body = StripHTML(pm.HtmlBody)
	}

	payload := &InboundPayload{
		Source:     "postmark",
		Sender:     pm.From,
		Recipient:  pm.To,
		Subject:    pm.Subject,
		Body:       body,
		RawPayload: string(bodyBytes),
		Headers:    make(map[string]string),
	}

	for _, att := range pm.Attachments {
		raw := DecodeBase64(att.Content)
		attInfo := AttachmentInfo{
			Filename:    att.Name,
			ContentType: att.ContentType,
			Size:        att.ContentLength,
		}

		if strings.HasSuffix(strings.ToLower(att.Name), ".pdf") {
			if pdfText, err := ExtractTextFromPDF(raw); err == nil && pdfText != "" {
				attInfo.TextContent = pdfText
				payload.Body += "\n\n--- Attachment: " + att.Name + " ---\n" + pdfText
			}
		}

		payload.Attachments = append(payload.Attachments, attInfo)
	}

	return payload, nil
}

// ParseMailgunWebhook parses a Mailgun Inbound multipart request.
func ParseMailgunWebhook(r *http.Request) (*InboundPayload, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("parse mailgun multipart: %w", err)
	}

	sender := r.FormValue("sender")
	recipient := r.FormValue("recipient")
	subject := r.FormValue("subject")
	bodyPlain := r.FormValue("body-plain")
	bodyHtml := r.FormValue("body-html")

	body := bodyPlain
	if body == "" && bodyHtml != "" {
		body = StripHTML(bodyHtml)
	}

	payload := &InboundPayload{
		Source:     "mailgun",
		Sender:     sender,
		Recipient:  recipient,
		Subject:    subject,
		Body:       body,
		RawPayload: fmt.Sprintf("sender: %s\nrecipient: %s\nsubject: %s\n\n%s", sender, recipient, subject, body),
		Headers:    make(map[string]string),
	}

	return payload, nil
}

// SESNotification represents an AWS SNS wrapping SES inbound notification.
type SESNotification struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

type SESMessage struct {
	Mail struct {
		Source      string   `json:"source"`
		Destination []string `json:"destination"`
		CommonHeaders struct {
			Subject string `json:"subject"`
			From    []string `json:"from"`
		} `json:"commonHeaders"`
	} `json:"mail"`
	Content string `json:"content"`
}

// ParseSESWebhook parses an AWS SES SNS webhook notification.
func ParseSESWebhook(bodyBytes []byte) (*InboundPayload, error) {
	var sns SESNotification
	if err := json.Unmarshal(bodyBytes, &sns); err != nil {
		return nil, fmt.Errorf("unmarshal ses sns payload: %w", err)
	}

	if sns.Type != "Notification" || sns.Message == "" {
		return nil, fmt.Errorf("not an ses notification event")
	}

	var ses SESMessage
	if err := json.Unmarshal([]byte(sns.Message), &ses); err != nil {
		// If message is raw email text
		parsed, err := ParseRawEmail([]byte(sns.Message))
		if err != nil {
			return nil, fmt.Errorf("parse raw ses email: %w", err)
		}
		return &InboundPayload{
			Source:     "ses",
			Sender:     parsed.Sender,
			Recipient:  parsed.Recipient,
			Subject:    parsed.Subject,
			Body:       parsed.Combined,
			RawPayload: string(bodyBytes),
		}, nil
	}

	sender := ses.Mail.Source
	if len(ses.Mail.CommonHeaders.From) > 0 {
		sender = ses.Mail.CommonHeaders.From[0]
	}
	recipient := ""
	if len(ses.Mail.Destination) > 0 {
		recipient = ses.Mail.Destination[0]
	}

	parsed, err := ParseRawEmail([]byte(ses.Content))
	body := ""
	if err == nil {
		body = parsed.Combined
	}

	return &InboundPayload{
		Source:     "ses",
		Sender:     sender,
		Recipient:  recipient,
		Subject:    ses.Mail.CommonHeaders.Subject,
		Body:       body,
		RawPayload: string(bodyBytes),
	}, nil
}

// GenericInboundRequest represents a direct REST or form submission.
type GenericInboundRequest struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Source    string `json:"source"`
}

// ParseGenericWebhook parses standard JSON or Form lead submissions.
func ParseGenericWebhook(r *http.Request) (*InboundPayload, error) {
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		var req GenericInboundRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal json: %w", err)
		}
		source := req.Source
		if source == "" {
			source = "generic_api"
		}
		return &InboundPayload{
			Source:     source,
			Sender:     req.Sender,
			Recipient:  req.Recipient,
			Subject:    req.Subject,
			Body:       req.Body,
			RawPayload: string(bodyBytes),
		}, nil
	}

	// Fallback to Form
	_ = r.ParseForm()
	sender := r.FormValue("sender")
	if sender == "" {
		sender = r.FormValue("from")
	}
	subject := r.FormValue("subject")
	body := r.FormValue("body")
	if body == "" {
		body = r.FormValue("message")
	}
	source := r.FormValue("source")
	if source == "" {
		source = "web_form"
	}

	return &InboundPayload{
		Source:     source,
		Sender:     sender,
		Recipient:  r.FormValue("recipient"),
		Subject:    subject,
		Body:       body,
		RawPayload: fmt.Sprintf("sender: %s\nsubject: %s\nbody: %s", sender, subject, body),
	}, nil
}

// ParseFileToInbound parses a local .eml, .txt, or .pdf into an InboundPayload.
func ParseFileToInbound(filename string, data []byte) (*InboundPayload, error) {
	lower := strings.ToLower(filename)

	if strings.HasSuffix(lower, ".pdf") {
		text, err := ExtractTextFromPDF(data)
		if err != nil {
			return nil, fmt.Errorf("extract pdf: %w", err)
		}
		return &InboundPayload{
			Source:     "rfq_file",
			Sender:     "attachment@inbound.pdf",
			Subject:    "Inbound PDF RFQ: " + filename,
			Body:       text,
			RawPayload: string(data[:min(len(data), 1024)]),
			Attachments: []AttachmentInfo{
				{Filename: filename, ContentType: "application/pdf", Size: len(data), TextContent: text},
			},
		}, nil
	}

	// Try RFC 2822 email parser
	parsed, err := ParseRawEmail(data)
	if err == nil && (parsed.Sender != "" || parsed.Subject != "") {
		return &InboundPayload{
			Source:      "eml_file",
			Sender:      parsed.Sender,
			Recipient:   parsed.Recipient,
			Subject:     parsed.Subject,
			Body:        parsed.Combined,
			RawPayload:  string(data),
			Attachments: parsed.Attachments,
			Headers:     parsed.Headers,
		}, nil
	}

	// Treat as plain text
	return &InboundPayload{
		Source:     "text_file",
		Sender:     "unknown@file.inbound",
		Subject:    "Inbound File: " + filename,
		Body:       string(data),
		RawPayload: string(data),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
