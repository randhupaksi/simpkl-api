package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"simpkl-api/internal/modules/students/entity"
)

type ImportRowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ImportResult struct {
	Total    int              `json:"total"`
	Valid    int              `json:"valid"`
	Imported int              `json:"imported"`
	Failed   int              `json:"failed"`
	Errors   []ImportRowError `json:"errors"`
}

type ImportService struct{ db *gorm.DB }

func NewImportService(db *gorm.DB) *ImportService { return &ImportService{db} }

func (s *ImportService) Import(
	ctx context.Context,
	source io.Reader,
	commit bool,
) (*ImportResult, error) {
	workbook, err := excelize.OpenReader(source)
	if err != nil {
		return nil, fmt.Errorf("file Excel tidak valid: %w", err)
	}
	defer workbook.Close()
	sheet := workbook.GetSheetName(0)
	rows, err := workbook.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	result := &ImportResult{}
	if len(rows) < 2 {
		return result, nil
	}
	headers := headerMap(rows[0])
	required := []string{"nis", "name", "class_id", "major_id", "cohort"}
	for _, column := range required {
		if _, exists := headers[column]; !exists {
			return nil, fmt.Errorf("kolom %s wajib tersedia", column)
		}
	}
	var students []entity.Student
	seenNIS := map[string]bool{}
	for index, row := range rows[1:] {
		if emptyRow(row) {
			continue
		}
		result.Total++
		rowNumber := index + 2
		student, rowErrors := parseStudent(row, headers, rowNumber)
		if student.NIS != "" {
			if seenNIS[student.NIS] {
				rowErrors = append(rowErrors, ImportRowError{rowNumber, "nis", "NIS duplikat di dalam file"})
			}
			seenNIS[student.NIS] = true
			var count int64
			_ = s.db.WithContext(ctx).Table("students").Where("nis = ? AND deleted_at IS NULL", student.NIS).Count(&count).Error
			if count > 0 {
				rowErrors = append(rowErrors, ImportRowError{rowNumber, "nis", "NIS sudah terdaftar"})
			}
		}
		if !s.referenceExists(ctx, "classes", student.ClassID) {
			rowErrors = append(rowErrors, ImportRowError{rowNumber, "class_id", "Kelas tidak ditemukan"})
		}
		if !s.referenceExists(ctx, "majors", student.MajorID) {
			rowErrors = append(rowErrors, ImportRowError{rowNumber, "major_id", "Jurusan tidak ditemukan"})
		}
		if len(rowErrors) > 0 {
			result.Errors = append(result.Errors, rowErrors...)
			result.Failed++
			continue
		}
		result.Valid++
		students = append(students, student)
	}
	if commit && len(students) > 0 {
		if err := s.db.WithContext(ctx).CreateInBatches(&students, 100).Error; err != nil {
			return nil, err
		}
		result.Imported = len(students)
	}
	return result, nil
}

func (s *ImportService) referenceExists(ctx context.Context, table, id string) bool {
	if id == "" {
		return false
	}
	var count int64
	_ = s.db.WithContext(ctx).Table(table).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0
}

func parseStudent(row []string, headers map[string]int, rowNumber int) (entity.Student, []ImportRowError) {
	value := func(key string) string {
		index, exists := headers[key]
		if !exists || index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	cohort, cohortErr := strconv.Atoi(value("cohort"))
	student := entity.Student{
		NIS: value("nis"), NISN: value("nisn"), Name: value("name"),
		Nickname: value("nickname"), Gender: strings.ToLower(value("gender")),
		ClassID: value("class_id"), MajorID: value("major_id"), Cohort: cohort,
		Phone: value("phone"), Email: strings.ToLower(value("email")),
		Address: value("address"), ParentName: value("parent_name"),
		ParentPhone: value("parent_phone"), Status: "active", PKLStatus: "unplaced",
	}
	var errors []ImportRowError
	for field, fieldValue := range map[string]string{
		"nis": student.NIS, "name": student.Name, "class_id": student.ClassID, "major_id": student.MajorID,
	} {
		if fieldValue == "" {
			errors = append(errors, ImportRowError{rowNumber, field, "Wajib diisi"})
		}
	}
	if cohortErr != nil || cohort < 2000 || cohort > 2200 {
		errors = append(errors, ImportRowError{rowNumber, "cohort", "Tahun angkatan tidak valid"})
	}
	return student, errors
}

func headerMap(headers []string) map[string]int {
	result := make(map[string]int, len(headers))
	for index, header := range headers {
		result[strings.ToLower(strings.TrimSpace(header))] = index
	}
	return result
}

func emptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
