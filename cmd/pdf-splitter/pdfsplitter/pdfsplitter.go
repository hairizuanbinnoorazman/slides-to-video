package pdfsplitter

import "fmt"

type PdfSplitJob struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	PdfFileName        string `json:"pdf_filename"`
	IdemKeySetRunning  string `json:"running_idem_key"`
	IdemKeyCompleteRec string `json:"complete_rec_idem_key"`
}

func (j *PdfSplitJob) Validate() error {
	allError := ""
	if j.ID == "" {
		allError = allError + fmt.Sprintf("ID cannot be empty")
	}
	if j.PdfFileName == "" {
		allError = allError + fmt.Sprintf("PDF file name cannot be empty")
	}
	if allError != "" {
		return fmt.Errorf(allError)
	}
	return nil
}

type PDFSplitter interface {
	Process(job PdfSplitJob) error
}
