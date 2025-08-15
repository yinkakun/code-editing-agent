package main

import (
	"agent/tools"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/fatih/color"
)

type Agent struct {
	client         *anthropic.Client
	tools          []tools.ToolDefinition
	getUserMessage func() (string, bool)
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []tools.ToolDefinition) *Agent {
	return &Agent{
		tools:          tools,
		client:         client,
		getUserMessage: getUserMessage,
	}
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}

	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				InputSchema: tool.InputSchema,
				Description: anthropic.String(tool.Description),
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
	})

	return message, err
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDefinition tools.ToolDefinition
	var found bool

	for _, tool := range a.tools {
		if tool.Name == name {
			found = true
			toolDefinition = tool
			break
		}
	}

	if !found {
		return anthropic.NewToolResultBlock(id, "oops... tool not found", true)
	}

	fmt.Printf("%s: %s(%s)\n", color.GreenString("tool"), name, input)
	response, err := toolDefinition.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}

	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (ctrl+c to quit)")
	readUserInput := true
	for {
		if readUserInput {
			color.Blue("You:")
			userInput, ok := a.getUserMessage()

			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)

		if err != nil {
			return err
		}

		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("%s: %s\n", color.YellowString("Claude"), content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}

		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}

		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func main() {
	client := anthropic.NewClient()
	tools := []tools.ToolDefinition{tools.ReadFileDefinition, tools.ListFileDefinition}

	scanner := bufio.NewScanner(os.Stdin)

	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(&client, getUserMessage, tools)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
