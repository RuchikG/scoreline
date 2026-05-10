// Package sports defines shared sport identifiers used by app routing and settings.
package sports

// Sport identifies a supported sport.
type Sport string

const (
	None    Sport = ""
	Soccer  Sport = "soccer"
	Cricket Sport = "cricket"
)

// IsValid reports whether s is a selectable sport.
func (s Sport) IsValid() bool {
	return s == Soccer || s == Cricket
}
