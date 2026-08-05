package excel

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	reportstatus "simpkl-api/internal/platform/report/status"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (Generator) Generate(sheetName string, headers []string, rows [][]any) ([]byte, error) {
	workbook := excelize.NewFile()
	defer workbook.Close()
	defaultSheet := workbook.GetSheetName(0)
	_ = workbook.SetSheetName(defaultSheet, sheetName)

	lastColumn, _ := excelize.ColumnNumberToName(len(headers))
	_ = workbook.MergeCell(sheetName, "A1", lastColumn+"1")
	_ = workbook.MergeCell(sheetName, "A2", lastColumn+"2")
	_ = workbook.MergeCell(sheetName, "A3", lastColumn+"3")
	_ = workbook.SetCellValue(sheetName, "A1", "SIMPKL | SMK CITRA NEGARA")
	_ = workbook.SetCellValue(sheetName, "A2", "Laporan Penempatan Praktik Kerja Lapangan")
	_ = workbook.SetCellValue(sheetName, "A3", fmt.Sprintf("Dibuat %s  |  Total data: %d", formatGeneratedAt(), len(rows)))

	headerRow := 5
	for column, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(column+1, headerRow)
		_ = workbook.SetCellValue(sheetName, cell, header)
	}
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			if columnIndex < len(headers) && strings.EqualFold(headers[columnIndex], "status") {
				value = reportstatus.Label(fmt.Sprint(value))
			}
			cell, _ := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+headerRow+1)
			_ = workbook.SetCellValue(sheetName, cell, value)
		}
	}

	_ = workbook.SetCellStyle(sheetName, "A1", lastColumn+"1", titleStyle(workbook))
	_ = workbook.SetCellStyle(sheetName, "A2", lastColumn+"2", subtitleStyle(workbook))
	_ = workbook.SetCellStyle(sheetName, "A3", lastColumn+"3", metadataStyle(workbook))
	_ = workbook.SetCellStyle(sheetName, fmt.Sprintf("A%d", headerRow), lastColumn+fmt.Sprint(headerRow), headerStyle(workbook))
	if len(rows) > 0 {
		dataStart := headerRow + 1
		dataEnd := headerRow + len(rows)
		_ = workbook.SetCellStyle(sheetName, fmt.Sprintf("A%d", dataStart), lastColumn+fmt.Sprint(dataEnd), bodyStyle(workbook))
		for row := dataStart; row <= dataEnd; row++ {
			if (row-dataStart)%2 == 1 {
				_ = workbook.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), lastColumn+fmt.Sprint(row), alternateRowStyle(workbook))
			}
			for column, header := range headers {
				if strings.EqualFold(header, "status") {
					cell, _ := excelize.CoordinatesToCellName(column+1, row)
					value, _ := workbook.GetCellValue(sheetName, cell)
					_ = workbook.SetCellStyle(sheetName, cell, cell, statusStyle(workbook, value))
				}
			}
			_ = workbook.SetRowHeight(sheetName, row, 24)
		}
		_ = workbook.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastColumn, dataEnd), nil)
	}
	_ = workbook.SetRowHeight(sheetName, 1, 30)
	_ = workbook.SetRowHeight(sheetName, 2, 24)
	_ = workbook.SetRowHeight(sheetName, headerRow, 28)
	_ = workbook.SetPanes(sheetName, &excelize.Panes{Freeze: true, YSplit: headerRow})
	for column := 1; column <= len(headers); column++ {
		name, _ := excelize.ColumnNumberToName(column)
		_ = workbook.SetColWidth(sheetName, name, name, columnWidth(headers[column-1]))
	}
	buffer, err := workbook.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buffer.Bytes()), nil
}

func formatGeneratedAt() string {
	return timeNow().Format("02 January 2006 15:04")
}

var timeNow = func() time.Time { return time.Now() }

func headerStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"16805C"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    []excelize.Border{{Type: "bottom", Color: "0B5B43", Style: 2}},
	})
	return style
}

func titleStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 16}, Fill: excelize.Fill{Type: "pattern", Color: []string{"0B5B43"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	return style
}

func subtitleStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "14532D", Size: 12}, Fill: excelize.Fill{Type: "pattern", Color: []string{"DCFCE7"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center"}})
	return style
}

func metadataStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "64748B", Italic: true, Size: 10}, Alignment: &excelize.Alignment{Vertical: "center"}})
	return style
}

func bodyStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "1E293B", Size: 10}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}, Border: []excelize.Border{{Type: "bottom", Color: "E2E8F0", Style: 1}}})
	return style
}

func alternateRowStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{"F0FDF4"}, Pattern: 1}, Font: &excelize.Font{Color: "1E293B", Size: 10}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}, Border: []excelize.Border{{Type: "bottom", Color: "E2E8F0", Style: 1}}})
	return style
}

func statusStyle(workbook *excelize.File, value string) int {
	fill := "E2E8F0"
	font := "334155"
	switch strings.ToLower(value) {
	case "active", "aktif", "approved", "disetujui", "completed", "selesai", "ready", "siap mulai pkl", "sedang pkl":
		fill, font = "DCFCE7", "166534"
	case "cancelled", "dibatalkan":
		fill, font = "FEE2E2", "991B1B"
	}
	style, _ := workbook.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: font, Size: 10}, Fill: excelize.Fill{Type: "pattern", Color: []string{fill}, Pattern: 1}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}, Border: []excelize.Border{{Type: "bottom", Color: "E2E8F0", Style: 1}}})
	return style
}

func columnWidth(header string) float64 {
	switch strings.ToLower(header) {
	case "nis":
		return 14
	case "nama siswa":
		return 26
	case "kelas", "jurusan":
		return 17
	case "perusahaan":
		return 30
	case "guru pembimbing":
		return 25
	case "mulai", "selesai":
		return 16
	case "status":
		return 17
	default:
		return 22
	}
}
