package pdf

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	reportstatus "simpkl-api/internal/platform/report/status"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (Generator) Generate(title string, headers []string, rows [][]any) ([]byte, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("laporan membutuhkan minimal satu kolom")
	}
	columns := []float64{45, 110, 65, 70, 125, 105, 78, 78, 116}
	if len(headers) != len(columns) {
		columns = equalColumnWidths(len(headers), 792)
	}
	const rowsPerPage = 16
	pageCount := int(math.Max(1, math.Ceil(float64(len(rows))/rowsPerPage)))
	pageContents := make([]string, pageCount)
	for page := 0; page < pageCount; page++ {
		start := page * rowsPerPage
		end := start + rowsPerPage
		if end > len(rows) {
			end = len(rows)
		}
		pageContents[page] = renderPage(title, headers, rows[start:end], columns, page+1, pageCount)
	}

	var body bytes.Buffer
	body.WriteString("%PDF-1.4\n")
	objects := make([]int, 0, pageCount*2+4)
	writeObject := func(value string) int {
		objectNumber := len(objects) + 1
		objects = append(objects, body.Len())
		body.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objectNumber, value))
		return objectNumber
	}

	fontObject := writeObject("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	contentObjects := make([]int, 0, pageCount)
	for _, content := range pageContents {
		contentObjects = append(contentObjects, writeObject(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content)))
	}

	pageObjects := make([]int, pageCount)
	pagesObject := len(objects) + pageCount + 1
	for index, contentObject := range contentObjects {
		pageObjects[index] = writeObject(fmt.Sprintf("<< /Type /Page /Parent %d 0 R /MediaBox [0 0 842 595] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>", pagesObject, fontObject, contentObject))
	}
	kids := make([]string, len(pageObjects))
	for index, pageObject := range pageObjects {
		kids[index] = fmt.Sprintf("%d 0 R", pageObject)
	}
	writeObject(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), pageCount))
	writeObject(fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesObject))

	xref := body.Len()
	body.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for _, offset := range objects {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}
	body.WriteString(fmt.Sprintf("trailer << /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, len(objects), xref))
	return body.Bytes(), nil
}

func renderPage(title string, headers []string, rows [][]any, widths []float64, page, pageCount int) string {
	const left = 25.0
	const tableTop = 458.0
	const headerHeight = 32.0
	const rowHeight = 25.0
	var content strings.Builder
	content.WriteString("q 0.043 0.357 0.263 sc 0 510 842 85 re f Q\n")
	text(&content, left, 558, 20, "1 1 1", "SIMPkl | PRACTICAL WORK PLACEMENT")
	text(&content, left, 536, 13, "0.86 0.99 0.91", title)
	text(&content, left, 518, 9, "0.82 0.91 0.86", fmt.Sprintf("Laporan penempatan PKL  |  Dibuat %s  |  Total data: %d", time.Now().Format("02 January 2006 15:04"), len(rows)))

	captionY := tableTop + 12
	text(&content, left, captionY, 9, "0.20 0.28 0.25", fmt.Sprintf("Data penempatan%s", pageLabel(page, pageCount)))
	drawTable(&content, left, tableTop, headers, rows, widths, headerHeight, rowHeight)
	text(&content, left, 28, 8, "0.39 0.45 0.42", "Official SIMPkl practical work placement report")
	textRight(&content, 817, 28, 8, "0.39 0.45 0.42", fmt.Sprintf("Halaman %d dari %d", page, pageCount))
	return content.String()
}

func drawTable(content *strings.Builder, left, top float64, headers []string, rows [][]any, widths []float64, headerHeight, rowHeight float64) {
	x := left
	for index, header := range headers {
		fillRect(content, x, top-headerHeight, widths[index], headerHeight, "0.086 0.502 0.361")
		text(content, x+5, top-20, 8, "1 1 1", header)
		x += widths[index]
	}
	for rowIndex, row := range rows {
		y := top - headerHeight - float64(rowIndex+1)*rowHeight
		x = left
		fill := "1 1 1"
		if rowIndex%2 == 1 {
			fill = "0.941 0.992 0.957"
		}
		for columnIndex := range headers {
			fillRect(content, x, y, widths[columnIndex], rowHeight, fill)
			value := ""
			if columnIndex < len(row) {
				value = fmt.Sprint(row[columnIndex])
			}
			if columnIndex == len(headers)-1 {
				rawStatus := value
				value = reportstatus.Label(rawStatus)
				statusFill, statusText := statusColors(rawStatus)
				fillRect(content, x+4, y+6, widths[columnIndex]-8, rowHeight-12, statusFill)
				text(content, x+8, y+15, 7, statusText, truncate(value, int(widths[columnIndex]/5.1)))
			} else {
				text(content, x+5, y+15, 7.5, "0.12 0.18 0.15", truncate(value, int(widths[columnIndex]/4.8)))
			}
			x += widths[columnIndex]
		}
	}
	lineY := top - headerHeight - float64(len(rows))*rowHeight
	strokeLine(content, left, lineY, left+sum(widths), lineY, "0.75 0.84 0.79")
}

func text(content *strings.Builder, x, y, size float64, color, value string) {
	content.WriteString(fmt.Sprintf("BT /F1 %.1f Tf %s rg 1 0 0 1 %.1f %.1f Tm (%s) Tj ET\n", size, color, x, y, escape(value)))
}

func textRight(content *strings.Builder, x, y, size float64, color, value string) {
	text(content, x-float64(len(value))*size*0.48, y, size, color, value)
}

func fillRect(content *strings.Builder, x, y, width, height float64, color string) {
	content.WriteString(fmt.Sprintf("q %s sc %.1f %.1f %.1f %.1f re f Q\n", color, x, y, width, height))
}

func strokeLine(content *strings.Builder, x1, y1, x2, y2 float64, color string) {
	content.WriteString(fmt.Sprintf("q %s RG 0.5 w %.1f %.1f m %.1f %.1f l S Q\n", color, x1, y1, x2, y2))
}

func statusColors(status string) (string, string) {
	switch strings.ToLower(status) {
	case "active", "approved", "completed":
		return "0.82 0.96 0.87", "0.08 0.38 0.24"
	case "cancelled":
		return "1 0.88 0.88", "0.60 0.12 0.12"
	default:
		return "0.91 0.93 0.96", "0.22 0.28 0.35"
	}
}

func truncate(value string, max int) string {
	if max < 4 || len(value) <= max {
		return value
	}
	return value[:max-3] + "..."
}

func equalColumnWidths(count int, total float64) []float64 {
	width := total / float64(count)
	result := make([]float64, count)
	for index := range result {
		result[index] = width
	}
	return result
}

func sum(values []float64) float64 {
	result := 0.0
	for _, value := range values {
		result += value
	}
	return result
}

func pageLabel(page, pageCount int) string {
	if pageCount <= 1 {
		return ""
	}
	return fmt.Sprintf("  |  Bagian %d", page)
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.ReplaceAll(value, "\r", " ")
}
