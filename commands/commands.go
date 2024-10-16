package commands

import (
	"fmt"
	"log"
	"os/exec"

	"kasmlink/api"

	"github.com/spf13/cobra"
)

// Base KasmAPI instance that can be used across commands
var apiInstance *api.KasmAPI

// Flags for API credentials
var (
	apiKey    string
	apiSecret string
	baseURL   string = "http://localhost:8080"
)

func init() {
	// Initialize API instance with default values (these can be set via flags)
	apiInstance = api.NewKasmAPI(baseURL, "your-api-key", "your-api-secret")
}

// Execute initializes the CLI commands and handles their execution.
func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "kasmlink",
		Short: "KasmLink CLI for interacting with Kasm API",
		Long:  `KasmLink CLI allows you to interact with Kasm API to manage users, images, sessions, SSH connections, and more.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			apiInstance = api.NewKasmAPI(baseURL, apiKey, apiSecret)
		},
	}

	// Add persistent flags for API credentials
	rootCmd.PersistentFlags().StringVar(&apiKey, "api_key", "", "API key for Kasm API (required)")
	rootCmd.PersistentFlags().StringVar(&apiSecret, "api_secret", "", "API secret for Kasm API (required)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base_url", baseURL, "Base URL for the Kasm API")

	// Add subcommands here
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(listImagesCmd)
	rootCmd.AddCommand(requestSessionCmd)
	rootCmd.AddCommand(getUserCmd)
	rootCmd.AddCommand(deleteUserCmd)
	rootCmd.AddCommand(execCommandCmd)
	rootCmd.AddCommand(destroySessionCmd)
	rootCmd.AddCommand(updateUserCmd)
	rootCmd.AddCommand(sshConnectCmd)
	rootCmd.AddCommand(dockerRunCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// Utility function to check required parameters
func checkRequiredParams(params map[string]string) {
	for key, value := range params {
		if value == "" {
			log.Fatalf("%s is a required field", key)
		}
	}
}

// SSH Connect Utility Function
func initiateSSHConnection(host string, port int, sshUser, keyFile string) error {
	sshCommand := []string{"ssh", fmt.Sprintf("%s@%s", sshUser, host), "-p", fmt.Sprintf("%d", port)}
	if keyFile != "" {
		sshCommand = append(sshCommand, "-i", keyFile)
	}

	execCmd := exec.Command(sshCommand[0], sshCommand[1:]...)
	execCmd.Stdout = log.Writer()
	execCmd.Stderr = log.Writer()

	log.Printf("Initiating SSH connection to %s@%s:%d...", sshUser, host, port)
	return execCmd.Run()
}

// Create User Command
var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user in Kasm",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		firstName, _ := cmd.Flags().GetString("first_name")
		lastName, _ := cmd.Flags().GetString("last_name")
		password, _ := cmd.Flags().GetString("password")

		checkRequiredParams(map[string]string{
			"username": username,
			"password": password,
		})

		request := api.CreateUserRequest{
			APIKey:       apiKey,
			APIKeySecret: apiSecret,
			TargetUser: api.UserInfo{
				Username:  username,
				FirstName: firstName,
				LastName:  lastName,
				Password:  password,
			},
		}

		response, err := apiInstance.CreateUser(request)
		if err != nil {
			log.Fatalf("Error creating user: %v", err)
		}

		fmt.Printf("User created successfully: %+v\n", response)
	},
}

func init() {
	createUserCmd.Flags().StringP("username", "u", "", "Username for the new user (required)")
	createUserCmd.Flags().StringP("first_name", "f", "", "First name of the user")
	createUserCmd.Flags().StringP("last_name", "l", "", "Last name of the user")
	createUserCmd.Flags().StringP("password", "p", "", "Password for the new user (required)")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("password")
}

// List Images Command
var listImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "List all available images in Kasm",
	Run: func(cmd *cobra.Command, args []string) {
		images, err := apiInstance.ListImages()
		if err != nil {
			log.Fatalf("Error fetching images: %v", err)
		}

		for _, img := range images {
			fmt.Printf("Image ID: %s, Name: %s, Available: %v\n", img.ImageID, img.FriendlyName, img.Available)
		}
	},
}

// Request Kasm Session Command
var requestSessionCmd = &cobra.Command{
	Use:   "request-session",
	Short: "Request a new Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		imageID, _ := cmd.Flags().GetString("image_id")

		checkRequiredParams(map[string]string{
			"user_id":  userID,
			"image_id": imageID,
		})

		response, err := apiInstance.RequestKasmSession(userID, imageID)
		if err != nil {
			log.Fatalf("Error requesting session: %v", err)
		}

		fmt.Printf("Session created successfully: %+v\n", response)
	},
}

func init() {
	requestSessionCmd.Flags().StringP("user_id", "u", "", "User ID to create a session for (required)")
	requestSessionCmd.Flags().StringP("image_id", "i", "", "Image ID to use for the session (required)")
	requestSessionCmd.MarkFlagRequired("user_id")
	requestSessionCmd.MarkFlagRequired("image_id")
}

// Get User Command
var getUserCmd = &cobra.Command{
	Use:   "get-user",
	Short: "Retrieve a user's details from Kasm",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
		})

		response, err := apiInstance.GetUser(userID, "")
		if err != nil {
			log.Fatalf("Error retrieving user: %v", err)
		}

		fmt.Printf("User details: %+v\n", response)
	},
}

func init() {
	getUserCmd.Flags().StringP("user_id", "u", "", "User ID to fetch details for (required)")
	getUserCmd.MarkFlagRequired("user_id")
}

// Delete User Command
var deleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete a user in Kasm",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
		})

		if err := apiInstance.DeleteUser(userID, false); err != nil {
			log.Fatalf("Error deleting user: %v", err)
		}

		fmt.Printf("User with ID %s successfully deleted\n", userID)
	},
}

func init() {
	deleteUserCmd.Flags().StringP("user_id", "u", "", "User ID to delete (required)")
	deleteUserCmd.MarkFlagRequired("user_id")
}

// Exec Command in Kasm Session Command
var execCommandCmd = &cobra.Command{
	Use:   "exec-command",
	Short: "Execute a command in a Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		kasmID, _ := cmd.Flags().GetString("kasm_id")
		command, _ := cmd.Flags().GetString("command")

		checkRequiredParams(map[string]string{
			"user_id": userID,
			"kasm_id": kasmID,
			"command": command,
		})

		execConfig := api.ExecConfig{
			Cmd: command,
		}

		response, err := apiInstance.ExecCommand(userID, kasmID, execConfig)
		if err != nil {
			log.Fatalf("Error executing command: %v", err)
		}

		fmt.Printf("Command executed successfully: %+v\n", response)
	},
}

func init() {
	execCommandCmd.Flags().StringP("user_id", "u", "", "User ID associated with the Kasm session (required)")
	execCommandCmd.Flags().StringP("kasm_id", "k", "", "Kasm ID of the session (required)")
	execCommandCmd.Flags().StringP("command", "c", "", "Command to execute in the session (required)")
	execCommandCmd.MarkFlagRequired("user_id")
	execCommandCmd.MarkFlagRequired("kasm_id")
	execCommandCmd.MarkFlagRequired("command")
}

// Destroy Kasm Session Command
var destroySessionCmd = &cobra.Command{
	Use:   "destroy-session",
	Short: "Destroy a Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		kasmID, _ := cmd.Flags().GetString("kasm_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
			"kasm_id": kasmID,
		})

		if err := apiInstance.DestroyKasmSession(userID, kasmID); err != nil {
			log.Fatalf("Error destroying session: %v", err)
		}

		fmt.Printf("Session with ID %s successfully destroyed\n", kasmID)
	},
}

func init() {
	destroySessionCmd.Flags().StringP("user_id", "u", "", "User ID associated with the session (required)")
	destroySessionCmd.Flags().StringP("kasm_id", "k", "", "Kasm ID of the session to be destroyed (required)")
	destroySessionCmd.MarkFlagRequired("user_id")
	destroySessionCmd.MarkFlagRequired("kasm_id")
}

// Update User Command
var updateUserCmd = &cobra.Command{
	Use:   "update-user",
	Short: "Update an existing user in Kasm",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		firstName, _ := cmd.Flags().GetString("first_name")
		lastName, _ := cmd.Flags().GetString("last_name")

		checkRequiredParams(map[string]string{
			"user_id": userID,
		})

		request := api.UpdateUserRequest{
			APIKey:       apiKey,
			APIKeySecret: apiSecret,
			TargetUser: api.UserInfo{
				UserID:    userID,
				FirstName: firstName,
				LastName:  lastName,
			},
		}

		response, err := apiInstance.UpdateUser(request)
		if err != nil {
			log.Fatalf("Error updating user: %v", err)
		}

		fmt.Printf("User updated successfully: %+v\n", response)
	},
}

func init() {
	updateUserCmd.Flags().StringP("user_id", "u", "", "User ID to update (required)")
	updateUserCmd.Flags().StringP("first_name", "f", "", "First name of the user")
	updateUserCmd.Flags().StringP("last_name", "l", "", "Last name of the user")
	updateUserCmd.MarkFlagRequired("user_id")
}

// SSH Connect Command
var sshConnectCmd = &cobra.Command{
	Use:   "ssh-connect",
	Short: "Initiate SSH connection to a running Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		kasmID, _ := cmd.Flags().GetString("kasm_id")
		sshUser, _ := cmd.Flags().GetString("ssh_user")
		keyFile, _ := cmd.Flags().GetString("key_file")

		// Check required parameters
		if userID == "" || kasmID == "" {
			log.Fatalf("User ID and Kasm ID are required fields")
		}

		// Fetch the session status to get the connection details
		sessionStatus, err := apiInstance.GetKasmStatus(userID, kasmID)
		if err != nil {
			log.Fatalf("Error retrieving Kasm session status: %v", err)
		}

		// Check if the session is in a running state
		if sessionStatus.OperationalStatus != "running" {
			log.Fatalf("Session is not in running state. Current status: %s", sessionStatus.OperationalStatus)
		}

		// Assuming the host and port information are available in the session status response
		host := sessionStatus.OperationalMessage // Replace with the correct field that contains the host
		port := 22                               // Assuming the standard SSH port

		// Create the SSH command
		sshCommand := []string{"ssh", fmt.Sprintf("%s@%s", sshUser, host), "-p", fmt.Sprintf("%d", port)}
		if keyFile != "" {
			sshCommand = append(sshCommand, "-i", keyFile)
		}

		// Execute the SSH command
		execCmd := exec.Command(sshCommand[0], sshCommand[1:]...)
		execCmd.Stdout = log.Writer()
		execCmd.Stderr = log.Writer()

		log.Printf("Initiating SSH connection to %s@%s:%d...", sshUser, host, port)
		if err := execCmd.Run(); err != nil {
			log.Fatalf("Failed to establish SSH connection: %v", err)
		}
	},
}

func init() {
	sshConnectCmd.Flags().StringP("user_id", "u", "", "User ID associated with the Kasm session (required)")
	sshConnectCmd.Flags().StringP("kasm_id", "k", "", "Kasm ID of the session to connect via SSH (required)")
	sshConnectCmd.Flags().StringP("ssh_user", "s", "kasm-user", "SSH username for connection (default: kasm-user)")
	sshConnectCmd.Flags().StringP("key_file", "i", "", "Path to SSH private key file (optional)")
	sshConnectCmd.MarkFlagRequired("user_id")
	sshConnectCmd.MarkFlagRequired("kasm_id")
}

// Docker Run Command to Execute Docker Commands over SSH
var dockerRunCmd = &cobra.Command{
	Use:   "docker-run",
	Short: "Execute Docker commands over SSH on a remote server",
	Long:  "Allows you to run Docker commands on a remote server via SSH, such as starting, stopping, and inspecting Docker containers.",
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		keyFile, _ := cmd.Flags().GetString("key_file")
		dockerCommand, _ := cmd.Flags().GetString("docker_command")

		if host == "" || dockerCommand == "" {
			log.Fatalf("Host and Docker command are required fields")
		}

		// Prepare the SSH command to execute the Docker command on the remote server
		sshCommand := []string{"ssh", fmt.Sprintf("%s@%s", user, host), "-p", "22"}
		if keyFile != "" {
			sshCommand = append(sshCommand, "-i", keyFile)
		}

		// Append the Docker command to the SSH command
		sshCommand = append(sshCommand, dockerCommand)

		// Execute the SSH command with the Docker command
		execCmd := exec.Command(sshCommand[0], sshCommand[1:]...)
		execCmd.Stdout = log.Writer()
		execCmd.Stderr = log.Writer()

		log.Printf("Executing Docker command on %s: %s", host, dockerCommand)
		if err := execCmd.Run(); err != nil {
			log.Fatalf("Failed to execute Docker command: %v", err)
		}
	},
}

func init() {
	dockerRunCmd.Flags().StringP("host", "H", "", "Hostname or IP address of the remote server (required)")
	dockerRunCmd.Flags().StringP("user", "u", "root", "SSH username for connection (default: root)")
	dockerRunCmd.Flags().StringP("key_file", "k", "", "Path to SSH private key file (optional)")
	dockerRunCmd.Flags().StringP("docker_command", "d", "", "Docker command to execute on the remote server (required)")
	dockerRunCmd.MarkFlagRequired("host")
	dockerRunCmd.MarkFlagRequired("docker_command")
}
