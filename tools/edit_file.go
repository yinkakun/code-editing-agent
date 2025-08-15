package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	if editFileInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if editFileInput.NewStr == editFileInput.OldStr {
		return "", fmt.Errorf("old_str and new_str must be different")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			return createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)

	count := strings.Count(oldContent, editFileInput.OldStr)
	if count == 0 {
		return "", fmt.Errorf("old_str not found in file")
	}
	if count > 1 {
		return "", fmt.Errorf("old_str found %d times - must match exactly once", count)
	}

	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, 1)

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced text in %s", editFileInput.Path), nil
}

func createNewFile(filePath, content string) (string, error) {
	dir := filepath.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	return fmt.Sprintf("Successfully created file %s", filePath), nil
}
