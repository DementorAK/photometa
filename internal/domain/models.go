package domain

// ImageFile represents a processed image with its properties.
type ImageFile struct {
	Path     string   `json:"path"`
	Name     string   `json:"name"` // Filename or "<stream>"
	Metadata Metadata `json:"metadata"`
}

// Metadata contains all image metadata organized by category.
type Metadata struct {
	Format   string    `json:"format"`    // jpeg, png, tiff, webp, gif
	FileSize int64     `json:"file_size"` // in bytes
	Tags     []TagInfo `json:"tags,omitempty"`
}

// AllGroups returns all metadata property group names.
func AllGroups() []string {
	return []string{"File", "Shooting", "Photo", "Location", "Equipment", "Author", "Other"}
}

// TagInfo represents a single parsed metadata property.
type TagInfo struct {
	Type  string `json:"type"`
	Group string `json:"group"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}
