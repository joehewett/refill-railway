package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

// makeHTTPHandler is a helper function that wraps an apiFunc and returns an http.HandlerFunc.
func makeHTTPHandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
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

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/", makeHTTPHandler(s.healthCheck)).Methods("GET")
	router.HandleFunc("/health", makeHTTPHandler(s.healthCheck)).Methods("GET")
	router.HandleFunc("/refill", makeHTTPHandler(s.handleRefill)).Methods("GET")

	fmt.Printf("Listening on %s\n", s.listenAddr)

	return http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleRefill(w http.ResponseWriter, r *http.Request) error {
	var refillRequest RefillRequest
	err := json.NewDecoder(r.Body).Decode(&refillRequest)
	if err != nil {
		return err
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
	return WriteJSON(w, http.StatusOK, "API is healthy")
}
