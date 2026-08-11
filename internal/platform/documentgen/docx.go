package documentgen

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// DOCX creates a standards-compliant Office Open XML document without an
// external office runtime. The generated file can be edited in Microsoft Word,
// LibreOffice, or any OOXML-compatible editor.
func DOCX(letter Letter) ([]byte, error) {
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	parts := map[string]string{
		"[Content_Types].xml":          contentTypes,
		"_rels/.rels":                  packageRelationships,
		"docProps/core.xml":            coreProperties,
		"docProps/app.xml":             appProperties,
		"word/_rels/document.xml.rels": documentRelationships,
		"word/styles.xml":              stylesXML,
		"word/settings.xml":            settingsXML,
		"word/document.xml":            documentXML(letter),
	}
	for name, content := range parts {
		part, err := archive.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func documentXML(letter Letter) string {
	var body strings.Builder
	body.WriteString(paragraph(letter.Profile.Type, true, 24, "center", 0, 0))
	body.WriteString(paragraph(strings.ToUpper(letter.Profile.Name), true, 32, "center", 0, 0))
	if letter.Profile.Tagline != "" {
		body.WriteString(paragraph(letter.Profile.Tagline, false, 20, "center", 0, 0))
	}
	contact := strings.Join(nonEmpty([]string{letter.Profile.Address, contactLine(letter.Profile)}), " | ")
	body.WriteString(paragraph(contact, false, 18, "center", 0, 0))
	body.WriteString(`<w:p><w:pPr><w:pBdr><w:bottom w:val="double" w:sz="8" w:space="4" w:color="16805C"/></w:pBdr><w:spacing w:after="240"/></w:pPr></w:p>`)
	body.WriteString(infoTable(letter))
	body.WriteString(paragraph("Yth. "+fallback(letter.Recipient, "Pimpinan Perusahaan/Instansi"), false, 22, "left", 0, 0))
	if letter.RecipientAddress != "" {
		body.WriteString(paragraph("di "+letter.RecipientAddress, false, 22, "left", 0, 240))
	}
	for _, block := range strings.Split(strings.ReplaceAll(letter.Body, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			body.WriteString(`<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>`)
			continue
		}
		body.WriteString(paragraph(trimmed, false, 22, "both", 720, 120))
	}
	body.WriteString(signatureTable(letter))

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>` + body.String() + `<w:sectPr>
<w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1417" w:header="720" w:footer="720"/>
<w:cols w:space="720"/><w:docGrid w:linePitch="360"/></w:sectPr></w:body></w:document>`
}

func infoTable(letter Letter) string {
	rows := [][2]string{{"Nomor", letter.Number}, {"Lampiran", "-"}, {"Perihal", letter.Subject}}
	var result strings.Builder
	result.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr>`)
	for _, row := range rows {
		result.WriteString(`<w:tr><w:tc><w:tcPr><w:tcW w:w="1500" w:type="dxa"/></w:tcPr>` + paragraph(row[0], false, 22, "left", 0, 0) + `</w:tc><w:tc><w:tcPr><w:tcW w:w="250" w:type="dxa"/></w:tcPr>` + paragraph(":", false, 22, "left", 0, 0) + `</w:tc><w:tc>` + paragraph(row[1], row[0] == "Perihal", 22, "left", 0, 0) + `</w:tc></w:tr>`)
	}
	result.WriteString(`</w:tbl>`)
	return result.String()
}

func signatureTable(letter Letter) string {
	cityDate := strings.Trim(strings.Join(nonEmpty([]string{letter.Profile.City, letter.Date}), ", "), " ,")
	return `<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr><w:tr><w:tc><w:tcPr><w:tcW w:w="4800" w:type="dxa"/></w:tcPr><w:p/></w:tc><w:tc><w:tcPr><w:tcW w:w="4200" w:type="dxa"/></w:tcPr>` +
		paragraph(cityDate, false, 22, "left", 0, 0) + paragraph(letter.Signatory.Title, false, 22, "left", 0, 0) +
		`<w:p><w:pPr><w:spacing w:after="1050"/></w:pPr></w:p>` + paragraph(letter.Signatory.Name, true, 22, "left", 0, 0) +
		paragraph(employeeLabel(letter.Signatory.EmployeeNumber), false, 20, "left", 0, 0) + `</w:tc></w:tr></w:tbl>`
}

func paragraph(value string, bold bool, size int, align string, firstLine, after int) string {
	weight := ""
	if bold {
		weight = `<w:b/>`
	}
	indent := ""
	if firstLine > 0 {
		indent = fmt.Sprintf(`<w:ind w:firstLine="%d"/>`, firstLine)
	}
	return fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="%s"/><w:spacing w:after="%d" w:line="360" w:lineRule="auto"/>%s</w:pPr><w:r><w:rPr>%s<w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, align, after, indent, weight, size, size, xmlEscape(value))
}

func xmlEscape(value string) string {
	var result bytes.Buffer
	_ = xml.EscapeText(&result, []byte(value))
	return result.String()
}

func contactLine(profile SchoolProfile) string {
	return strings.Join(nonEmpty([]string{profile.Phone, profile.Email, profile.Website}), " | ")
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func fallback(value, alternative string) string {
	if strings.TrimSpace(value) == "" {
		return alternative
	}
	return value
}

func employeeLabel(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "NIP/NIK. " + value
}

const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`
const packageRelationships = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`
const documentRelationships = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/></Relationships>`
const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:eastAsia="Arial" w:cs="Arial"/><w:sz w:val="22"/><w:szCs w:val="22"/><w:lang w:val="id-ID"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="360" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style></w:styles>`
const settingsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:zoom w:percent="100"/><w:defaultTabStop w:val="720"/></w:settings>`
const coreProperties = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/"><dc:title>Dokumen PKL</dc:title><dc:creator>SIMPKL</dc:creator><dc:subject>Praktik Kerja Lapangan</dc:subject></cp:coreProperties>`
const appProperties = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>SIMPKL</Application><Company>SIMPKL</Company></Properties>`
