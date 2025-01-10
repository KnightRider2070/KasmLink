package dockercli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fatih/color"
)

// PrintBuildLogs reads and formats Docker build logs for readability.
func PrintBuildLogs(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var msg struct {
		Stream string `json:"stream"`
		Error  string `json:"error"`
	}

	successColor := color.New(color.FgGreen).SprintFunc()
	errorColor := color.New(color.FgRed).SprintFunc()

	for decoder.More() {
		if err := decoder.Decode(&msg); err != nil {
			return fmt.Errorf("error decoding build logs: %w", err)
		}

		if msg.Error != "" {
			fmt.Println(errorColor(fmt.Sprintf("Error: %s", msg.Error)))
		} else if msg.Stream != "" {
			fmt.Print(successColor(msg.Stream))
		}
	}

	return nil
}
