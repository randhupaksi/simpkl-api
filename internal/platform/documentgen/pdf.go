package documentgen

import (
	"bytes"
	"fmt"
	"strings"
)

// PDF renders an A4 portrait official letter using built-in PDF fonts. Keeping
// the renderer self-contained makes generation deterministic on every server.
func PDF(letter Letter) ([]byte, error) {
	content := renderLetterPage(letter)
	var body bytes.Buffer
	body.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, 6)
	writeObject := func(value string) int {
		number := len(offsets) + 1
		offsets = append(offsets, body.Len())
		body.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", number, value))
		return number
	}
	font := writeObject("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	fontBold := writeObject("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>")
	stream := writeObject(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	page := writeObject(fmt.Sprintf("<< /Type /Page /Parent 5 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> /Contents %d 0 R >>", font, fontBold, stream))
	writeObject(fmt.Sprintf("<< /Type /Pages /Kids [%d 0 R] /Count 1 >>", page))
	writeObject("<< /Type /Catalog /Pages 5 0 R >>")
	xref := body.Len()
	body.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(offsets)+1))
	for _, offset := range offsets {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}
	body.WriteString(fmt.Sprintf("trailer << /Size %d /Root 6 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets)+1, xref))
	return body.Bytes(), nil
}

func renderLetterPage(letter Letter) string {
	var out strings.Builder
	centerPDF(&out, 595, 798, 11, true, strings.ToUpper(letter.Profile.Type))
	centerPDF(&out, 595, 779, 16, true, strings.ToUpper(letter.Profile.Name))
	if letter.Profile.Tagline != "" {
		centerPDF(&out, 595, 763, 9, false, letter.Profile.Tagline)
	}
	centerPDF(&out, 595, 749, 8, false, strings.Join(nonEmpty([]string{letter.Profile.Address, contactLine(letter.Profile)}), " | "))
	out.WriteString("q 0.086 0.502 0.361 RG 1.5 w 48 738 m 547 738 l S Q\n")
	out.WriteString("q 0.086 0.502 0.361 RG 0.5 w 48 734 m 547 734 l S Q\n")
	y := 710.0
	pdfText(&out, 52, y, 10, false, "Nomor")
	pdfText(&out, 118, y, 10, false, ": "+letter.Number)
	y -= 16
	pdfText(&out, 52, y, 10, false, "Lampiran")
	pdfText(&out, 118, y, 10, false, ": -")
	y -= 16
	pdfText(&out, 52, y, 10, false, "Perihal")
	pdfText(&out, 118, y, 10, true, ": "+letter.Subject)
	y -= 34
	pdfText(&out, 52, y, 10, false, "Yth. "+fallback(letter.Recipient, "Pimpinan Perusahaan/Instansi"))
	y -= 16
	if letter.RecipientAddress != "" {
		pdfText(&out, 52, y, 10, false, "di "+letter.RecipientAddress)
		y -= 26
	}
	for _, paragraph := range strings.Split(strings.ReplaceAll(letter.Body, "\r\n", "\n"), "\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			y -= 8
			continue
		}
		for _, line := range wrapText(paragraph, 88) {
			pdfText(&out, 68, y, 10, false, line)
			y -= 15
		}
		y -= 3
		if y < 175 {
			break
		}
	}
	signX := 350.0
	pdfText(&out, signX, 155, 10, false, strings.Trim(strings.Join(nonEmpty([]string{letter.Profile.City, letter.Date}), ", "), " ,"))
	pdfText(&out, signX, 139, 10, false, letter.Signatory.Title)
	pdfText(&out, signX, 72, 10, true, letter.Signatory.Name)
	if letter.Signatory.EmployeeNumber != "" {
		pdfText(&out, signX, 57, 9, false, employeeLabel(letter.Signatory.EmployeeNumber))
	}
	pdfText(&out, 48, 28, 7.5, false, "Dokumen resmi dibuat melalui SIMPkl")
	return out.String()
}

func pdfText(out *strings.Builder, x, y, size float64, bold bool, value string) {
	font := "F1"
	if bold {
		font = "F2"
	}
	out.WriteString(fmt.Sprintf("BT /%s %.1f Tf 0.10 0.15 0.13 rg 1 0 0 1 %.1f %.1f Tm (%s) Tj ET\n", font, size, x, y, escapePDF(value)))
}

func centerPDF(out *strings.Builder, width, y, size float64, bold bool, value string) {
	x := (width - float64(len([]rune(value)))*size*0.51) / 2
	if x < 48 {
		x = 48
	}
	pdfText(out, x, y, size, bold, value)
}

func wrapText(value string, limit int) []string {
	words := strings.Fields(value)
	if len(words) == 0 {
		return nil
	}
	lines := []string{words[0]}
	for _, word := range words[1:] {
		index := len(lines) - 1
		if len([]rune(lines[index]+" "+word)) <= limit {
			lines[index] += " " + word
		} else {
			lines = append(lines, word)
		}
	}
	return lines
}

func escapePDF(value string) string {
	value = strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)", "\r", " ", "\n", " ").Replace(value)
	return value
}
