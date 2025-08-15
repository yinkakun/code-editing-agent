package tools

import (
	"encoding/json"
	"os"
)

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	fileContent, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read file",
	Function:    ReadFile,
	InputSchema: ReadFileInputSchema,
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
}
