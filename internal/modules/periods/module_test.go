package periods

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeriodStatusTransitions(t *testing.T) {
	require.True(t, allowedTransitions["draft"]["preparation"])
	require.True(t, allowedTransitions["active"]["completed"])
	require.True(t, allowedTransitions["completed"]["archived"])
	require.False(t, allowedTransitions["draft"]["active"])
	require.False(t, allowedTransitions["archived"]["active"])
}
