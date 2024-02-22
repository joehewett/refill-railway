package internal

import (
	"os"
	"path/filepath"

	"github.com/unidoc/unipdf/v3/extractor"
	pdfModel "github.com/unidoc/unipdf/v3/model"
)

type FileType int

const (
	txt FileType = iota
	pdf
)

type FileDetails struct {
	bytes     []byte
	name      string
	extension FileType
}

type File interface {
	Name() string
	Type() FileType
	Load() (string, error)
}

func NewFile(details FileDetails) (File, error) {
	if details.extension == pdf {
		return &PDF{details}, nil
	} else {
		return &TXT{details}, nil
	}
}

type PDF struct {
	FileDetails
}

func (p *PDF) Name() string {
	return p.name
}

func (p *PDF) Type() FileType {
	return pdf
}

func (p *PDF) Load() (string, error) {
	// Convert the PDF bytes to a file so we can load it into unipdf
	f, err := os.Create(filepath.Join(os.TempDir(), p.name))
	if err != nil {
		return "", err
	}
	defer f.Close()

	defer func() {
		os.Remove(f.Name())
	}()

	_, err = f.Write(p.bytes)
	if err != nil {
		return "", err
	}

	pdfReader, err := pdfModel.NewPdfReader(f)
	if err != nil {
		return "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	var text string
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return "", err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}

		pageText, err := ex.ExtractText()
		if err != nil {
			return "", err
		}

		text += pageText
	}

	return text, nil
}

type TXT struct {
	FileDetails
}

func (t *TXT) Name() string {
	return t.name
}

func (t *TXT) Type() FileType {
	return txt
}

func (t *TXT) Load() (string, error) {
	data := string(t.bytes)

	return data, nil
}
