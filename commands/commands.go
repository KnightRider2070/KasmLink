package commands

import (
	"fmt"
	"log"

	"kasmlink/api"
	"kasmlink/ssh"

	"github.com/spf13/cobra"
)

// Base KasmAPI instance used across commands
var apiInstance *api.KasmAPI

// Flags for API credentials
var (
	apiKey    string
	apiSecret string
	baseURL   string = "http://localhost:8080"
	skipTLS   bool
)

// Declare the rootCmd as a global variable
var rootCmd = &cobra.Command{
	Use:   "kasmlink",
	Short: "KasmLink CLI for interacting with Kasm API",
	Long:  `KasmLink CLI allows you to interact with Kasm API to manage users, images, sessions, and more.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize API instance with values from flags
		skipTLS = api.AskUserToSkipTLS()
		apiInstance = api.NewKasmAPI(baseURL, apiKey, apiSecret)
		apiInstance.SkipTLSVerification = skipTLS
	},
}

func init() {
	// Add persistent flags for API credentials
	rootCmd.PersistentFlags().StringVar(&apiKey, "api_key", "", "API key for Kasm API (required)")
	rootCmd.PersistentFlags().StringVar(&apiSecret, "api_secret", "", "API secret for Kasm API (required)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base_url", baseURL, "Base URL for the Kasm API")

	// Add command groups
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(imageCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(sshCmd)
}

// Execute initializes the CLI commands and handles their execution.
func Execute() {
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

// USER COMMANDS GROUP

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
	Long:  "Manage Kasm users, including creating, deleting, updating, and listing users.",
}

var createUserCmd = &cobra.Command{
	Use:   "create",
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

		user := api.TargetUser{
			Username:  username,
			FirstName: firstName,
			LastName:  lastName,
			Password:  password,
		}

		response, err := apiInstance.CreateUser(user)
		if err != nil {
			log.Fatalf("Error creating user: %v", err)
		}

		fmt.Printf("User created successfully: %+v\n", response.UserID)
	},
}

var logoutUserCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out all sessions for an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
		})

		if err := apiInstance.LogoutUser(userID); err != nil {
			log.Fatalf("Error logging out user: %v", err)
		}

		fmt.Printf("User %s successfully logged out\n", userID)
	},
}

var getUserAttributesCmd = &cobra.Command{
	Use:   "get-attributes",
	Short: "Retrieve the attribute (preferences) settings for an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
		})

		attributes, err := apiInstance.GetUserAttributes(userID)
		if err != nil {
			log.Fatalf("Error retrieving user attributes: %v", err)
		}

		fmt.Printf("User attributes for user %s: %+v\n", userID, attributes)
	},
}

var updateUserAttributesCmd = &cobra.Command{
	Use:   "update-attributes",
	Short: "Update the attribute (preferences) settings for an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		attributeKey, _ := cmd.Flags().GetString("attribute_key")
		attributeValue, _ := cmd.Flags().GetString("attribute_value")

		checkRequiredParams(map[string]string{
			"user_id":         userID,
			"attribute_key":   attributeKey,
			"attribute_value": attributeValue,
		})

		attributes := api.UserAttributes{
			UserID: userID,
			// Assuming we only update certain fields (e.g. ShowTips)
			ShowTips: attributeValue == "true",
		}

		if err := apiInstance.UpdateUserAttributes(attributes); err != nil {
			log.Fatalf("Error updating user attributes: %v", err)
		}

		fmt.Printf("User attributes for user %s successfully updated\n", userID)
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users in the Kasm system",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := apiInstance.GetUsers()
		if err != nil {
			log.Fatalf("Error retrieving users: %v", err)
		}

		fmt.Println("List of users:")
		for _, user := range users {
			// Handle null fields by checking if they're nil
			firstName := "N/A"
			if user.FirstName != nil {
				firstName = *user.FirstName
			}

			lastName := "N/A"
			if user.LastName != nil {
				lastName = *user.LastName
			}

			fmt.Printf("User ID: %s, Username: %s, First Name: %s, Last Name: %s\n",
				user.UserID, user.Username, firstName, lastName)
		}
	},
}

func init() {
	// Add subcommands to userCmd group
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(logoutUserCmd)
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(updateUserAttributesCmd)
	userCmd.AddCommand(getUserAttributesCmd)

	// Flags for createUserCmd
	createUserCmd.Flags().StringP("username", "u", "", "Username for the new user (required)")
	createUserCmd.Flags().StringP("first_name", "f", "", "First name of the user")
	createUserCmd.Flags().StringP("last_name", "l", "", "Last name of the user")
	createUserCmd.Flags().StringP("password", "p", "", "Password for the new user (required)")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("password")

	// Flags for logoutUserCmd
	logoutUserCmd.Flags().StringP("user_id", "u", "", "User ID to log out (required)")
	logoutUserCmd.MarkFlagRequired("user_id")

	// Flags for getUserAttributesCmd
	getUserAttributesCmd.Flags().StringP("user_id", "u", "", "User ID to retrieve attributes for (required)")
	getUserAttributesCmd.MarkFlagRequired("user_id")

	// Flags for updateUserAttributesCmd
	updateUserAttributesCmd.Flags().StringP("user_id", "u", "", "User ID to update attributes for (required)")
	updateUserAttributesCmd.Flags().StringP("attribute_key", "k", "", "Attribute key to update (required)")
	updateUserAttributesCmd.Flags().StringP("attribute_value", "v", "", "New value for the attribute (required)")
	updateUserAttributesCmd.MarkFlagRequired("user_id")
	updateUserAttributesCmd.MarkFlagRequired("attribute_key")
	updateUserAttributesCmd.MarkFlagRequired("attribute_value")
}

// IMAGE COMMANDS GROUP

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image management commands",
	Long:  "Manage Kasm images, including listing and deploying images.",
}

var listImagesCmd = &cobra.Command{
	Use:   "list",
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

func init() {
	// Add subcommands to imageCmd group
	imageCmd.AddCommand(listImagesCmd)
}

// SESSION COMMANDS GROUP

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Session management commands",
	Long:  "Manage Kasm sessions, including requesting and destroying sessions.",
}

var requestSessionCmd = &cobra.Command{
	Use:   "create",
	Short: "Request a new Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		imageID, _ := cmd.Flags().GetString("image_id")

		checkRequiredParams(map[string]string{
			"user_id":  userID,
			"image_id": imageID,
		})

		request := api.RequestKasmRequest{
			APIKey:       apiKey,
			APIKeySecret: apiSecret,
			UserID:       userID,
			ImageID:      imageID,
		}

		response, err := apiInstance.RequestKasmSession(request)
		if err != nil {
			log.Fatalf("Error requesting session: %v", err)
		}

		fmt.Printf("Session created successfully: %+v\n", response)
	},
}

var destroySessionCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a Kasm session",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user_id")
		kasmID, _ := cmd.Flags().GetString("kasm_id")

		checkRequiredParams(map[string]string{
			"user_id": userID,
			"kasm_id": kasmID,
		})

		request := api.DestroyKasmRequest{
			APIKey:       apiKey,
			APIKeySecret: apiSecret,
			UserID:       userID,
			KasmID:       kasmID,
		}

		if err := apiInstance.DestroyKasmSession(request); err != nil {
			log.Fatalf("Error destroying session: %v", err)
		}

		fmt.Printf("Session with ID %s successfully destroyed\n", kasmID)
	},
}

func init() {
	// Add subcommands to sessionCmd group
	sessionCmd.AddCommand(requestSessionCmd)
	sessionCmd.AddCommand(destroySessionCmd)

	requestSessionCmd.Flags().StringP("user_id", "u", "", "User ID to create a session for (required)")
	requestSessionCmd.Flags().StringP("image_id", "i", "", "Image ID to use for the session (required)")
	requestSessionCmd.MarkFlagRequired("user_id")
	requestSessionCmd.MarkFlagRequired("image_id")

	destroySessionCmd.Flags().StringP("user_id", "u", "", "User ID associated with the session (required)")
	destroySessionCmd.Flags().StringP("kasm_id", "k", "", "Kasm ID of the session to be destroyed (required)")
	destroySessionCmd.MarkFlagRequired("user_id")
	destroySessionCmd.MarkFlagRequired("kasm_id")
}

// SSH command group
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH management commands",
	Long:  "Manage SSH connections and file transfers for Kasm sessions.",
}

// SSH subcommand: Connect to a remote server
var sshConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a remote server via SSH",
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		privateKey, _ := cmd.Flags().GetString("private_key")

		checkRequiredParams(map[string]string{
			"host": host,
			"user": user,
		})

		// Create new SSH client
		client, err := ssh.NewSSHClient(host, port, user, password, privateKey)
		if err != nil {
			log.Fatalf("Error creating SSH client: %v", err)
		}
		defer client.Disconnect()

		fmt.Printf("Connected to %s:%d as %s\n", host, port, user)
	},
}

// SSH subcommand: Upload a file to the remote server
var sshUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local file to the remote server via SCP",
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		privateKey, _ := cmd.Flags().GetString("private_key")
		localPath, _ := cmd.Flags().GetString("local")
		remotePath, _ := cmd.Flags().GetString("remote")

		checkRequiredParams(map[string]string{
			"host":   host,
			"user":   user,
			"local":  localPath,
			"remote": remotePath,
		})

		// Create new SSH client
		client, err := ssh.NewSSHClient(host, port, user, password, privateKey)
		if err != nil {
			log.Fatalf("Error creating SSH client: %v", err)
		}
		defer client.Disconnect()

		// Upload file
		err = client.UploadFile(localPath, remotePath)
		if err != nil {
			log.Fatalf("Error uploading file: %v", err)
		}

		fmt.Printf("File uploaded to %s\n", remotePath)
	},
}

// SSH subcommand: Download a file from the remote server
var sshDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a file from the remote server via SCP",
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		privateKey, _ := cmd.Flags().GetString("private_key")
		localPath, _ := cmd.Flags().GetString("local")
		remotePath, _ := cmd.Flags().GetString("remote")

		checkRequiredParams(map[string]string{
			"host":   host,
			"user":   user,
			"local":  localPath,
			"remote": remotePath,
		})

		// Create new SSH client
		client, err := ssh.NewSSHClient(host, port, user, password, privateKey)
		if err != nil {
			log.Fatalf("Error creating SSH client: %v", err)
		}
		defer client.Disconnect()

		// Download file
		err = client.DownloadFile(remotePath, localPath)
		if err != nil {
			log.Fatalf("Error downloading file: %v", err)
		}

		fmt.Printf("File downloaded to %s\n", localPath)
	},
}

func init() {
	// Add subcommands to sshCmd group
	sshCmd.AddCommand(sshConnectCmd)
	sshCmd.AddCommand(sshUploadCmd)
	sshCmd.AddCommand(sshDownloadCmd)

	// Common SSH flags for all SSH subcommands
	sshCmd.PersistentFlags().StringP("host", "H", "", "SSH server hostname (required)")
	sshCmd.PersistentFlags().IntP("port", "P", 22, "SSH server port (default: 22)")
	sshCmd.PersistentFlags().StringP("user", "u", "", "SSH username (required)")
	sshCmd.PersistentFlags().StringP("password", "p", "", "SSH password (optional)")
	sshCmd.PersistentFlags().StringP("private_key", "i", "", "Path to SSH private key file (optional)")

	// Flags for upload command
	sshUploadCmd.Flags().StringP("local", "l", "", "Local file path to upload (required)")
	sshUploadCmd.Flags().StringP("remote", "r", "", "Remote destination path (required)")
	sshUploadCmd.MarkFlagRequired("local")
	sshUploadCmd.MarkFlagRequired("remote")

	// Flags for download command
	sshDownloadCmd.Flags().StringP("local", "l", "", "Local destination path to download file to (required)")
	sshDownloadCmd.Flags().StringP("remote", "r", "", "Remote file path to download (required)")
	sshDownloadCmd.MarkFlagRequired("local")
	sshDownloadCmd.MarkFlagRequired("remote")
}
