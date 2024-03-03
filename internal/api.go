package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dslipak/pdf"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

type APIServer struct {
	engine *gin.Engine
}

const MAX_FILE_SIZE = 1024 * 1024

type apiFunc func(w http.ResponseWriter, r *http.Request) error

// makeHTTPHandler is a helper function that wraps an apiFunc and returns an http.HandlerFunc.
func makeHTTPHandler(fn apiFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := fn(c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{Error: err.Error()})
		}
	}
}

type APIError struct {
	Error string `json:"error"`
}

// WriteJSON is a helper function that writes a JSON response to the http.ResponseWriter.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(v)
}

func NewAPIServer() *APIServer {
	engine := gin.Default()

	// Add CORS middleware
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return &APIServer{
		engine,
	}
}

func (s *APIServer) Run() error {
	s.engine.GET("/", makeHTTPHandler(s.healthCheck))
	s.engine.POST("/refill", makeHTTPHandler(s.handleRefill))

	port := os.Getenv("PORT")
	fmt.Printf("Starting server on port %s\n", port)
	return s.engine.Run()
}

func (s *APIServer) handleRefill(w http.ResponseWriter, r *http.Request) error {
	var refillRequest RefillRequest
	var parsedFiles []File

	fmt.Printf("Body is %#v\n", r.Body)

	r.Body = http.MaxBytesReader(w, r.Body, MAX_FILE_SIZE)
	if err := r.ParseMultipartForm(MAX_FILE_SIZE); err != nil {
		return fmt.Errorf("failed to parse multipart form: %s", err)
	}

	// 32 MB is the default used by FormFile()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return fmt.Errorf("failed to parse multipart form: %s", err)
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]
	fmt.Printf("lengeth of files is %d\n", len(files))

	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file: %s", err)
		}

		defer file.Close()

		buff := make([]byte, 512)
		n, err := file.Read(buff)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file into buffer: %s", err)
		}
		buff = buff[:n]

		filetype := http.DetectContentType(buff)
		// If the file isn't a TXT or a PDF file, return an error
		if filetype != "text/plain" && filetype != "application/pdf" {
			return fmt.Errorf("invalid file type: %s", filetype)
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek file: %s", err)
		}

		// Open the PDF
		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create uploads directory: %s", err)
		}

		fileName := fmt.Sprintf("./%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		f, err := os.Create(fmt.Sprintf("./uploads/%s", fileName))
		if err != nil {
			return fmt.Errorf("failed to create file: %s", err)
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return fmt.Errorf("failed to copy file: %s", err)
		}

		// Open the PDF
		pdfText, err := parsePDF(fmt.Sprintf("./uploads/%s", fileName))
		if err != nil {
			return fmt.Errorf("failed to parse PDF: %s", err)
		}

		newFile := File{
			Name: fileHeader.Filename,
			Data: pdfText,
		}

		parsedFiles = append(parsedFiles, newFile)
	}

	refillRequest = RefillRequest{
		Keys:         r.MultipartForm.Value["keys"],
		Files:        parsedFiles,
		Instructions: r.MultipartForm.Value["instructions"][0],
		OpenAIKey:    r.MultipartForm.Value["openai_api_key"][0],
	}

	result, err := doRefill(refillRequest)
	if err != nil {
		return fmt.Errorf("an error occurred while refilling: %w", err)
	}

	err = WriteJSON(w, http.StatusOK, result)
	if err != nil {
		return fmt.Errorf("an error occurred while writing the response: %w", err)
	}

	return nil
}

func parsePDF(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}

	text, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to get text content: %w", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(text)
	return buf.String(), nil
}

func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, "Get your jiffies out Thomas")
}
