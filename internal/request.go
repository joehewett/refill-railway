package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

var (
	systemPrompt = `
		You are a parser of unstructured text data. Your task is to return JSON in the format specified, 
		where each JSON value is filled in using information provided in the data.
		If the data does not contain the information required, return an empty string for that value.

	`

	extraInstructionsTagPrompt = `
		[Extra Instructions]
	`

	jsonTagPrompt = `
		[JSON Structure]
	`

	dataTagPrompt = `
		[Data to fill JSON structure with]
	`

	filledJSONTagPrompt = `
		[Filled JSON structure]
	`
	openAIKey = os.Getenv("OPENAI_API_KEY")
	model     = openai.GPT3Dot5Turbo16K0613
)

func requestFill(file string, json string, instructions string) (string, error) {
	// Get the actual data from the file
	// data, err := file.Load()
	// if err != nil {
	// 	return "", fmt.Errorf("failed to load file: %w", err)
	// }

	var messages = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	if instructions != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: extraInstructionsTagPrompt,
		})

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: instructions,
		})
	}

	others := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: jsonTagPrompt,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: json,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: dataTagPrompt,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: file,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: filledJSONTagPrompt,
		},
	}
	messages = append(messages, others...)

	completion, err := createChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return completion, nil
}

func createChatCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	if openAIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(openAIKey)

	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	completion := response.Choices[0].Message.Content

	return completion, nil
}
