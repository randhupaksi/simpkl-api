package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStudent(t *testing.T) {
	headers := headerMap([]string{"nis", "name", "class_id", "major_id", "cohort"})
	student, errors := parseStudent(
		[]string{"1001", "Randhu", "class-id", "major-id", "2026"},
		headers,
		2,
	)
	require.Empty(t, errors)
	require.Equal(t, "1001", student.NIS)
	require.Equal(t, 2026, student.Cohort)
	require.Equal(t, "unplaced", student.PKLStatus)
}

func TestParseStudentReportsRequiredColumns(t *testing.T) {
	headers := headerMap([]string{"nis", "name", "class_id", "major_id", "cohort"})
	_, errors := parseStudent([]string{"", "", "", "", "invalid"}, headers, 3)
	require.Len(t, errors, 5)
}
