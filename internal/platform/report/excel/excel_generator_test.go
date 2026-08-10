package excel

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestGenerateCreatesStyledReadableWorkbook(t *testing.T) {
	data, err := New().Generate(
		"Rekap PKL",
		[]string{"NIS", "Nama Siswa", "Kelas", "Perusahaan", "Status"},
		[][]any{{"2026001", "Aulia Rahma Putri", "XII RPL 1", "PT Teknologi Indonesia", "active"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	workbook, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open generated workbook: %v", err)
	}
	defer workbook.Close()
	if got, _ := workbook.GetCellValue("Rekap PKL", "A1"); got != "SIMPkl | PRACTICAL WORK PLACEMENT" {
		t.Fatalf("unexpected title: %q", got)
	}
	if got, _ := workbook.GetCellValue("Rekap PKL", "B6"); got != "Aulia Rahma Putri" {
		t.Fatalf("unexpected student: %q", got)
	}
	if got, _ := workbook.GetCellValue("Rekap PKL", "E6"); got != "Aktif" {
		t.Fatalf("status was not humanized: %q", got)
	}
	panes, err := workbook.GetPanes("Rekap PKL")
	if err != nil || !panes.Freeze {
		t.Fatal("header pane is not frozen")
	}
	styleID, err := workbook.GetCellStyle("Rekap PKL", "A5")
	if err != nil || styleID == 0 {
		t.Fatal("header style was not applied")
	}
}
