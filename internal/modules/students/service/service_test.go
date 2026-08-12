package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStudent(t *testing.T) {
	headers := headerMap([]string{"nis", "name", "class_id", "major_id"})
	student, errors := parseStudent(
		[]string{"1001", "Randhu", "class-id", "major-id"},
		headers,
		2,
	)
	require.Empty(t, errors)
	require.Equal(t, "1001", student.NIS)
	require.Equal(t, "unplaced", student.PKLStatus)
}

func TestParseStudentReportsRequiredColumns(t *testing.T) {
	headers := headerMap([]string{"nis", "name", "class_id", "major_id"})
	_, errors := parseStudent([]string{"", "", "", ""}, headers, 3)
	require.Len(t, errors, 4)
}
