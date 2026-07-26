package pdf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateProducesPDF(t *testing.T) {
	data, err := New().Generate(
		"Rekap PKL",
		[]string{"NIS", "Nama"},
		[][]any{{"1001", "Randhu"}},
	)
	require.NoError(t, err)
	require.Greater(t, len(data), 100)
	require.Equal(t, "%PDF", string(data[:4]))
}
