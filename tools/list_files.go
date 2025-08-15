package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var ListFileDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	Function:    ListFiles,
	InputSchema: ListFilesInputSchema,
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFileInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFileInput)
	if err != nil {
		panic(err)
	}

	dir := "."
	if listFileInput.Path != "" {
		dir = listFileInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			suffix := ""
			if info.IsDir() {
				suffix = "/"
			}
			files = append(files, relPath+suffix)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
