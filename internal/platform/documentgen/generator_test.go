package documentgen

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleLetter() Letter {
	return Letter{
		Profile:   SchoolProfile{Name: "SMK Nusantara Teknologi", Type: "Sekolah Menengah Kejuruan", Address: "Jl. Pendidikan No. 10, Depok", City: "Depok", Phone: "(021) 7700000", Email: "info@sekolah.sch.id", Tagline: "Terampil, Profesional, dan Berintegritas"},
		Signatory: Signatory{Name: "Drs. Ahmad Fauzi, M.Pd.", Title: "Kepala Sekolah", EmployeeNumber: "19750512 200501 1 001"},
		Number:    "001/PKL/VIII/2026", Date: "10 Agustus 2026", Subject: "Permohonan Praktik Kerja Lapangan",
		Recipient: "Pimpinan PT Teknologi Indonesia", RecipientAddress: "Kota Depok",
		Body: "Dengan hormat,\n\nDalam rangka pelaksanaan program Praktik Kerja Lapangan, kami mengajukan peserta didik berikut:\n\nNama: Aulia Rahma Putri\nNIS: 2026001\nKelas: XII RPL 1\n\nAtas perhatian dan kerja sama yang baik, kami sampaikan terima kasih.",
	}
}

func TestDOCXProducesValidOOXMLPackage(t *testing.T) {
	data, err := DOCX(sampleLetter())
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open generated docx: %v", err)
	}
	required := map[string]bool{"[Content_Types].xml": false, "_rels/.rels": false, "word/document.xml": false, "word/styles.xml": false}
	for _, file := range reader.File {
		if _, ok := required[file.Name]; !ok {
			continue
		}
		required[file.Name] = true
		content, readErr := readZip(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		decoder := xml.NewDecoder(bytes.NewReader(content))
		for {
			if _, decodeErr := decoder.Token(); decodeErr == io.EOF {
				break
			} else if decodeErr != nil {
				t.Fatalf("invalid XML in %s: %v", file.Name, decodeErr)
			}
		}
		if file.Name == "word/document.xml" && !strings.Contains(string(content), "Aulia Rahma Putri") {
			t.Fatal("document data was not rendered")
		}
	}
	for name, found := range required {
		if !found {
			t.Errorf("missing OOXML part %s", name)
		}
	}
	writeQAArtifact(t, "surat-pengantar-pkl.docx", data)
}

func TestPDFProducesReadableA4Document(t *testing.T) {
	data, err := PDF(sampleLetter())
	if err != nil {
		t.Fatal(err)
	}
	value := string(data)
	if !strings.HasPrefix(value, "%PDF-1.4") {
		t.Fatal("missing PDF signature")
	}
	if !strings.Contains(value, "/MediaBox [0 0 595 842]") {
		t.Fatal("PDF is not A4 portrait")
	}
	if !strings.Contains(value, "Aulia Rahma Putri") {
		t.Fatal("letter data was not rendered")
	}
	if !strings.HasSuffix(value, "%%EOF") {
		t.Fatal("missing PDF EOF marker")
	}
	writeQAArtifact(t, "surat-pengantar-pkl.pdf", data)
}

func readZip(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func writeQAArtifact(t *testing.T, name string, data []byte) {
	t.Helper()
	directory := os.Getenv("SIMPKL_ARTIFACT_DIR")
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, name), data, 0o640); err != nil {
		t.Fatal(err)
	}
}
