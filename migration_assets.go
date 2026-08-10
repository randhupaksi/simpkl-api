package migrationassets

import "embed"

// FS keeps the versioned SQL migrations inside the API build. The runtime can
// therefore apply schema changes without requiring a separate SQL import step.
//
//go:embed migrations/*.sql
var FS embed.FS
