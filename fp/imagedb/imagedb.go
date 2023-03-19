package imagedb

type ImageEntry struct {
	Size     int64    `json:"imageSize"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags"`
}

var (
	OrientationTags = []string{"wide", "tall", "square"}
	SpectrumTags    = []string{"blackandwhite", "color"}
)
