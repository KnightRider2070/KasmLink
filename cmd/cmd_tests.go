package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"kasmlink/pkg/procedures"
	sshmanager "kasmlink/pkg/sshmanager"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
	"os"
	"time"
)

//Notes: requires user to be in docker group sudo usermod -aG docker $user

// Init initializes the root command.
func init() {
	// Define "compose" command
	composeCmd := &cobra.Command{
		Use: "test",
	}

	// Add subcommands for generating Docker Compose files
	composeCmd.AddCommand(createTestEnv())

	// Add "compose" to the root command
	RootCmd.AddCommand(composeCmd)

}

func createTestEnv() *cobra.Command {
	return &cobra.Command{
		Use:  "api",
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			// Create Sample Target User
			tuser := webApi.TargetUser{
				UserID:       "",
				Username:     "testUser",
				FirstName:    "Test",
				LastName:     "User",
				Locked:       false,
				Disabled:     false,
				Organization: "TestFlight",
				Password:     "secure",
			}

			//Create SampleUserConf
			uDetails := userParser.UserDetails{
				TargetUser:             tuser,
				Role:                   "All Users",
				AssignedContainerTag:   "kasmweb/chrome:1.16.1",
				KasmSessionOfContainer: "",
				EnvironmentArgs:        make(map[string]string),
			}

			uCOnfig := userParser.UsersConfig{
				UserDetails: []userParser.UserDetails{uDetails},
			}

			// Create a temporary file
			tempFile, err := os.CreateTemp("", "sample_user_conf_*.yaml")
			if err != nil {
				fmt.Printf("Failed to create temporary file: %v\n", err)
				return
			}
			defer os.Remove(tempFile.Name())

			// Write uDetails to the temporary file
			encoder := yaml.NewEncoder(tempFile)
			if err := encoder.Encode(uCOnfig); err != nil {
				fmt.Printf("Failed to write to temporary file: %v\n", err)
				return
			}
			encoder.Close()

			//Create ssh config
			sshConfig, _ := sshmanager.NewSSHConfig("thor", "stark", "192.168.120.5", 22, "C:\\Users\\Thor\\.ssh\\known_hosts", 10*time.Second)

			//Create KASM API
			kApi := webApi.NewKasmAPI("https://192.168.120.5", "C6QmU5ohTUIE", "91MRn9E7FyBSPJ5HtexWrubIG3SYLkB5", true, 10*time.Second)

			ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)
			err = procedures.CreateTestEnvironment(ctx, tempFile.Name(), sshConfig, kApi)
			if err != nil {
				return
			}
		},
	}
}
