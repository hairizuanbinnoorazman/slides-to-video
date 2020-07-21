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
type basicPDFImporter struct {
	queue queue.Queue
}

func NewBasicPDFImporter(q queue.Queue) basicPDFImporter {
	return basicPDFImporter{
		queue: q,
	}
}

func (p basicPDFImporter) Start(s pdfslideimages.PDFSlideImages) error {
	values := map[string]string{
		"id":                    s.ID,
		"project_id":            s.ProjectID,
		"pdf_filename":          s.PDFFile,
		"running_idem_key":      s.SetRunningIdemKey,
		"complete_rec_idem_key": s.CompleteRecIdemKey,
	}
	jsonValue, _ := json.Marshal(values)

	err := p.queue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}

	return nil
}
