package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type RefillRequest struct {
	Keys         []string `json:"keys"`
	Files        []File   `json:"files"`
	Instructions string   `json:"instructions"`
	OpenAIKey    string   `json:"openai_api_key"`
}

func doRefill(request RefillRequest) (string, error) {
	if request.OpenAIKey == "" {
		return "", fmt.Errorf("OpenAI API key is required")
	}

	request.OpenAIKey = getAdminKey(request.OpenAIKey)

	// Convert the array of strings into a JSON object where each string is a key and each value is an empty string
	jsonSkeleton := make(map[string]interface{})
	for _, key := range request.Keys {
		jsonSkeleton[key] = key
	}

	// Convert the JSON object back into a string
	jsonStr, err := json.Marshal(jsonSkeleton)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON skeleton: %w", err)
	}

	startTime := time.Now()
	defer func() {
		fmt.Printf("Total time taken: %s\n", time.Since(startTime))
	}()

	ch := make(chan string)

	for _, file := range request.Files {
		go fill(file, string(jsonStr), request.Instructions, request.OpenAIKey, ch)
	}

	results := []string{}
	for range request.Files {
		results = append(results, <-ch)
	}

	fmt.Printf(results[0])

	endTime := time.Now()
	fmt.Printf("Total time taken: %s\n", endTime.Sub(startTime))

	// Concatenate all the results together
	result := "["
	for _, r := range results {
		result += r + ","
	}
	result = result[:len(result)-1] + "]"

	return result, nil
}

func fill(file File, jsonStr string, instructions string, openAIKey string, ch chan string) {
	startTime := time.Now()

	fmt.Println("Requesting filled data from LM")

	result, err := requestFill(file.Data, jsonStr, instructions, openAIKey)
	if err != nil {
		ch <- fmt.Sprintf("\nFailed to request fill for file: %sz\n", err)
		return
	}

	// Unmarshal the result and add a new filename key to it
	var resultJSON map[string]interface{}
	err = json.Unmarshal([]byte(result), &resultJSON)
	if err != nil {
		ch <- fmt.Sprintf("\nFailed to unmarshal result for file %s: %s\n", file, err)
		return

	}

	// resultJSON["filename"] = "file.Name()"
	bytes, err := json.MarshalIndent(resultJSON, "", "\t")
	if err != nil {
		ch <- fmt.Sprintf("\nFailed to marshal result for file %s: %s\n", file, err)
		return
	}

	ch <- string(bytes)

	fmt.Printf("Time taken for file %s: %s\n", file, time.Since(startTime))
}

// getAdminKey returns the OpenAI API key if the given key is the admin key.
func getAdminKey(key string) string {
	if key == os.Getenv("ADMIN_KEY") {
		return os.Getenv("OPENAI_API_KEY")
	}

	return key
}
