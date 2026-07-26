package excel

import (
	"bytes"

	"github.com/xuri/excelize/v2"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (Generator) Generate(sheetName string, headers []string, rows [][]any) ([]byte, error) {
	workbook := excelize.NewFile()
	defaultSheet := workbook.GetSheetName(0)
	_ = workbook.SetSheetName(defaultSheet, sheetName)
	for column, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		_ = workbook.SetCellValue(sheetName, cell, header)
	}
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+2)
			_ = workbook.SetCellValue(sheetName, cell, value)
		}
	}
	lastColumn, _ := excelize.ColumnNumberToName(len(headers))
	_ = workbook.SetCellStyle(sheetName, "A1", lastColumn+"1", headerStyle(workbook))
	_ = workbook.SetPanes(sheetName, &excelize.Panes{Freeze: true, YSplit: 1})
	for column := 1; column <= len(headers); column++ {
		name, _ := excelize.ColumnNumberToName(column)
		_ = workbook.SetColWidth(sheetName, name, name, 22)
	}
	buffer, err := workbook.WriteToBuffer()
	_ = workbook.Close()
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buffer.Bytes()), nil
}

func headerStyle(workbook *excelize.File) int {
	style, _ := workbook.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return style
}
