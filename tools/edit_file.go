package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

var EditFIleDefinition = ToolDefinition{
	Name:        "edit_file",
	Function:    EditFIle,
	InputSchema: EditFileInputSchema,
	Description: `Make edits to a text file.
								Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.
								If the file specified with path doesn't exist, it will be created.`,
}

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func EditFIle(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", err
	}

	if (editFileInput.Path == "") || (editFileInput.NewStr == editFileInput.OldStr) {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsExist(err) && editFileInput.OldStr == "" {
			return createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", err
	}

	oldContent := string(content)
	newContent := strings.ReplaceAll(oldContent, editFileInput.OldStr, editFileInput.NewStr)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}
	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", nil
	}

	return "", nil
}

func createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	return fmt.Sprintf("successfully created file %s", filePath), nil
}
