package assets

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed icons/*.svg
var iconsFS embed.FS

// GetIcon returns the SVG content of the requested icon.
// name should be without .svg extension, e.g. "format_png" or "group_author".
func GetIcon(name string) ([]byte, error) {
	return iconsFS.ReadFile(fmt.Sprintf("icons/%s.svg", name))
}

// ListIcons returns the names of all available icons.
func ListIcons() ([]string, error) {
	entries, err := fs.ReadDir(iconsFS, "icons")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
