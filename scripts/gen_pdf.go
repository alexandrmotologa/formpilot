package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
)

func main() {
	var streamBuf bytes.Buffer
	zw := zlib.NewWriter(&streamBuf)

	text := "BT\n/F1 14 Tf\n(Enterprise RFQ: Apex Clearing Systems) Tj\n/F1 12 Tf\n(Company: Apex Clearing) Tj\n(Seats: 250 enterprise seats) Tj\n(Budget: $120,000 USD) Tj\n(Timeline: P0 Immediate 30 days) Tj\n(Compliance: SOC2 Type II, Okta SAML SSO) Tj\nET"

	_, _ = zw.Write([]byte(text))
	_ = zw.Close()

	compressed := streamBuf.Bytes()

	pdfContent := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n%s\nendstream\nendobj\n2 0 obj\n<< /Type /Page /Contents 1 0 R >>\nendobj\n3 0 obj\n<< /Type /Pages /Kids [2 0 R] /Count 1 >>\nendobj\n4 0 obj\n<< /Type /Catalog /Pages 3 0 R >>\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000010 00000 n \ntrailer\n<< /Root 4 0 R >>\n%%%%EOF", len(compressed), string(compressed))

	_ = os.WriteFile("examples/sample_rfq.pdf", []byte(pdfContent), 0644)
	fmt.Println("Generated examples/sample_rfq.pdf")
}
