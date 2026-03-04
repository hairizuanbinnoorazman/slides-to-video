package scriptgenerator

import "context"

type SlideImage struct {
	ImageID string
	Order   int
	Content []byte // raw PNG bytes from blob storage
}

type ScriptGenerator interface {
	// GenerateDescription analyzes all slides to produce a deck description
	GenerateDescription(ctx context.Context, slides []SlideImage) (string, error)
	// GenerateScript generates narration text for a single slide given project context
	GenerateScript(ctx context.Context, description string, slide SlideImage) (string, error)
}
