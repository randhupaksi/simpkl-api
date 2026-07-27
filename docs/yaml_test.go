package docs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLDocumentsAreValid(t *testing.T) {
	for _, path := range []string{"openapi.yaml", "../docker-compose.yml"} {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			require.NoError(t, err)
			var document any
			require.NoError(t, yaml.Unmarshal(content, &document))
		})
	}
}
