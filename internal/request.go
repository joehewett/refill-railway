package internal

import (
	"context"
	"fmt"

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
	model = openai.GPT3Dot5Turbo
)

func requestFill(file string, json string, instructions string, openAIKey string) (string, error) {
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
			Content: string(file),
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: filledJSONTagPrompt,
		},
	}
	messages = append(messages, others...)

	for _, message := range messages {
		fmt.Printf("%s: %s\n", message.Role, message.Content)
	}

	completion, err := createChatCompletion(messages, openAIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	fmt.Printf("Completion: %s\n", completion)

	return completion, nil
}

func createChatCompletion(messages []openai.ChatCompletionMessage, openAIKey string) (string, error) {
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
