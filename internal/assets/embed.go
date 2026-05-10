// Package assets provides embedded static assets for the scoreline application.
package assets

import (
	_ "embed"
)

// Logo is the scoreline logo PNG image, embedded at compile time.
// Used for desktop notifications on Linux and Windows.
//
//go:embed scoreline-logo.png
var Logo []byte
