package videogenerator

import (
	"context"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

// VideoGenerator is expected to take a while to complete
// It is assumed to be async in nature
type VideoGenerator interface {
	Start(context.Context, videosegment.VideoSegment) error
}
