package pdf

import (
	"bytes"
	"fmt"
	"strings"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (Generator) Generate(title string, headers []string, rows [][]any) ([]byte, error) {
	lines := []string{title, strings.Join(headers, " | ")}
	for _, row := range rows {
		values := make([]string, len(row))
		for index, value := range row {
			values[index] = fmt.Sprint(value)
		}
		lines = append(lines, strings.Join(values, " | "))
	}
	content := "BT /F1 9 Tf 40 800 Td "
	for index, line := range lines {
		if index > 0 {
			content += "0 -14 Td "
		}
		content += "(" + escape(line) + ") Tj "
	}
	content += "ET"

	var body bytes.Buffer
	offsets := []int{0}
	write := func(value string) {
		offsets = append(offsets, body.Len())
		body.WriteString(value)
	}
	body.WriteString("%PDF-1.4\n")
	write("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n")
	write("2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n")
	write("3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >> endobj\n")
	write(fmt.Sprintf("4 0 obj << /Length %d >> stream\n%s\nendstream endobj\n", len(content), content))
	write("5 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n")
	xref := body.Len()
	body.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for index := 1; index <= 5; index++ {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[index]))
	}
	body.WriteString(fmt.Sprintf("trailer << /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", xref))
	return body.Bytes(), nil
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	return strings.ReplaceAll(value, ")", "\\)")
}
