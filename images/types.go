package images

import "time"

type GalleryImage struct {
	File         *string            `json:"file"`
	Name         *string            `json:"name"`
	Url          *string            `json:"url"`
	Width        *int               `json:"width"`
	Height       *int               `json:"height"`
	OriginalSize *int64             `json:"originalSize"`
	Sizes        *map[string]string `json:"sizes"`
	CreatedAt    time.Time          `json:"createdAt"`
}

type ImageSizes struct {
	Small      *string `json:"256"`
	Medium     *string `json:"512"`
	Large      *string `json:"1024"`
	ExtraLarge *string `json:"2048"`
}

const (
	Small      = "256"
	Medium     = "512"
	Large      = "1024"
	ExtraLarge = "2048"
	Original   = "original"
)

type Message struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
