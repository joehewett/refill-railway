package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

type APIServer struct {
	engine *gin.Engine
}

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

	fmt.Printf("Body is %v\n", r.Body)

	err := json.NewDecoder(r.Body).Decode(&refillRequest)
	if err != nil {
		return fmt.Errorf("failed to decode refill request: %s", err)
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

func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, "Get your jiffies out Thomas")
}
