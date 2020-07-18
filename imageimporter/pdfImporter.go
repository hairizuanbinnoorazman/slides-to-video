package imageimporter

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type PDFImporter interface {
	Start(pdfslideimages.PDFSlideImages) error
}

// PDFImporter servers to be the holding struct to handle importing pdf to slide images
type BasicPDFImporter struct {
	Queue queue.Queue
}

func (p BasicPDFImporter) Start(s pdfslideimages.PDFSlideImages) error {
	values := map[string]string{"id": s.ID, "pdfFileName": s.PDFFile}
	jsonValue, _ := json.Marshal(values)

	err := p.Queue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}

	return nil
}
