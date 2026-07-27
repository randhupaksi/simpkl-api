package placements

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlacementStatusTransitions(t *testing.T) {
	require.True(t, placementTransitions["draft"]["pending_verification"])
	require.True(t, placementTransitions["approved"]["ready"])
	require.True(t, placementTransitions["active"]["transferred"])
	require.True(t, placementTransitions["active"]["completed"])
	require.False(t, placementTransitions["draft"]["active"])
	require.False(t, placementTransitions["completed"]["active"])
}
