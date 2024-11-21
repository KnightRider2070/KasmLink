package cmd

import (
	"fmt"
	"os"
)

// HandleError handles an error by logging it and exiting the program if it's not nil.
func HandleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1) // Properly exit with an error code.
	}
}
