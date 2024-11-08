package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercli"
	"time"
)

func init() {
	// Define the "dockercli" command
	dockerCliCmd := &cobra.Command{
		Use:   "dockercli",
		Short: "Interact with Docker CLI commands",
		Long:  `Use this command to interact with Docker-related functionalities in the Docker CLI`,
	}

	// Define and add "compose" commands
	composeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Interact with Docker Compose",
		Long:  `This command allows you to manage Docker Compose commands such as up, down, build, start, stop, etc.`,
	}
	composeCmd.AddCommand(
		createSimpleSubCommand("up", "Bring up Docker Compose services", dockercli.ComposeUp),
		createSimpleSubCommand("down", "Stop and remove Docker Compose services", dockercli.ComposeDown),
		createSimpleSubCommand("start", "Start existing Docker Compose services", dockercli.ComposeStart),
		createSimpleSubCommand("stop", "Stop running Docker Compose services", dockercli.ComposeStop),
		createSimpleSubCommand("build", "Build Docker Compose services", dockercli.ComposeBuild),
		createSimpleSubCommandWithOutput("logs", "Fetch Docker Compose logs", dockercli.ComposeLogs),
		createListServicesCommand("list-services", "List Docker Compose services", dockercli.ListComposeServices),
		createSimpleSubCommandWithOutput("inspect-service", "Inspect Docker container by ID", dockercli.InspectComposeService),
	)

	// Define and add "container" commands
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Manage Docker containers",
		Long:  `This command allows you to manage Docker containers, including start, stop, remove, restart, copy files, etc.`,
	}
	containerCmd.AddCommand(
		createTimedSubCommand("start", "Start a Docker container", dockercli.StartContainer),
		createTimedSubCommand("stop", "Stop a Docker container", dockercli.StopContainer),
		createTimedSubCommand("remove", "Remove a Docker container", dockercli.RemoveContainer),
		createTimedSubCommand("restart", "Restart a Docker container", dockercli.RestartContainer),
		createCopySubCommand("copy-from", "Copy files from Docker container to host", dockercli.CopyFromContainer),
		createCopySubCommand("copy-to", "Copy files from host to Docker container", dockercli.CopyToContainer),
	)

	// Define and add "image" commands
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage Docker images",
		Long:  `This command allows you to manage Docker images, including pulling, pushing, removing, exporting, and importing.`,
	}
	imageCmd.AddCommand(
		createImageSubCommand("pull", "Pull a Docker image from the registry", dockercli.PullImage, 5*time.Minute),
		createImageSubCommand("push", "Push a Docker image to the registry", dockercli.PushImage, 5*time.Minute),
		createImageSubCommand("remove", "Remove a Docker image by name or ID", dockercli.RemoveImage, 2*time.Minute),
		createImageSubCommandWithOutput("list-images", "List all Docker images on the host", dockercli.ListImages, 1*time.Minute),
		createSimpleImageCommand("update-all-images", "Pull the latest version of all present Docker images", dockercli.UpdateAllImages, 10*time.Minute),
	)

	// Define and add "network" commands
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage Docker networks",
		Long:  `This command allows you to manage Docker networks, including creating, inspecting, and removing them.`,
	}
	networkCmd.AddCommand(
		createNetworkSubCommand("create", "Create a Docker network", dockercli.CreateDockerNetwork),
		createNetworkInspectCommand("inspect", "Inspect a Docker network", dockercli.InspectNetwork),
		createNetworkRemoveCommand("remove", "Remove a Docker network", dockercli.RemoveNetwork),
	)

	// Define and add "volume" commands
	volumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage Docker volumes",
		Long:  `This command allows you to manage Docker volumes, including creating, inspecting, and removing them.`,
	}
	volumeCmd.AddCommand(
		createSimpleSubCommand("create", "Create a Docker volume", dockercli.CreateVolume),
		createSimpleSubCommandWithOutput("inspect", "Inspect a Docker volume by name", dockercli.InspectVolume),
		createSimpleSubCommand("remove", "Remove a Docker volume by name", dockercli.RemoveVolume),
	)

	// Add all command categories to dockerCliCmd
	dockerCliCmd.AddCommand(composeCmd, containerCmd, imageCmd, networkCmd, volumeCmd)

	// Add "dockercli" to root command
	RootCmd.AddCommand(dockerCliCmd)
}

func createSimpleImageCommand(use, shortDesc string, action func(context.Context, int) error, timeout time.Duration) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [retries]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			retries := ParseInt(args[0], 3)
			ctx, cancel := CreateContextWithTimeout(timeout)
			defer cancel()

			err := action(ctx, retries)
			HandleError(err)
		},
	}
}

// Reusable command creators
func createSimpleSubCommand(use, shortDesc string, action func(string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [arg]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := action(args[0])
			HandleError(err)
		},
	}
}

func createSimpleSubCommandWithOutput(use, shortDesc string, action func(string) (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [arg]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			output, err := action(args[0])
			HandleError(err)
			fmt.Println(output)
		},
	}
}

func createListServicesCommand(use, shortDesc string, action func(string) (map[string]map[string]string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [composeFilePath]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			services, err := action(args[0])
			HandleError(err)
			for name, info := range services {
				fmt.Printf("Service: %s, Info: %+v\n", name, info)
			}
		},
	}
}

func createTimedSubCommand(use, shortDesc string, action func(string, int, time.Duration) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [containerID] [retries] [timeout]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			containerID := args[0]
			retries := ParseInt(args[1], 3)
			timeout := ParseDuration(args[2], 30*time.Second)
			err := action(containerID, retries, timeout)
			HandleError(err)
		},
	}
}

func createCopySubCommand(use, shortDesc string, action func(string, string, string, int, time.Duration) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [containerID] [containerPath] [hostPath] [retries] [timeout]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			containerID := args[0]
			containerPath := args[1]
			hostPath := args[2]
			retries := ParseInt(args[3], 3)
			timeout := ParseDuration(args[4], 30*time.Second)
			err := action(containerID, containerPath, hostPath, retries, timeout)
			HandleError(err)
		},
	}
}

func createImageSubCommand(use, shortDesc string, action func(context.Context, int, string) error, timeout time.Duration) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [imageName] [retries]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			imageName := args[0]
			retries := ParseInt(args[1], 3)
			ctx, cancel := CreateContextWithTimeout(timeout)
			defer cancel()

			err := action(ctx, retries, imageName)
			HandleError(err)
		},
	}
}

func createImageSubCommandWithOutput(use, shortDesc string, action func(context.Context, int) ([]string, error), timeout time.Duration) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [retries]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			retries := ParseInt(args[0], 3)
			ctx, cancel := CreateContextWithTimeout(timeout)
			defer cancel()

			images, err := action(ctx, retries)
			HandleError(err)

			for _, image := range images {
				fmt.Println(image)
			}
		},
	}
}

func createNetworkSubCommand(use, shortDesc string, action func(dockercli.NetworkOptions) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [networkName] [driver] [retries] [retryDelay] [timeout]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			networkName := args[0]
			driver := args[1]
			retries := ParseInt(args[2], 3)
			retryDelay := ParseDuration(args[3], 2*time.Second)
			timeout := ParseDuration(args[4], 30*time.Second)

			opts := dockercli.NetworkOptions{
				Name:       networkName,
				Driver:     dockercli.NetworkDriver(driver),
				RetryCount: retries,
				RetryDelay: retryDelay,
				Timeout:    timeout,
			}

			err := action(opts)
			HandleError(err)
		},
	}
}

func createNetworkInspectCommand(use, shortDesc string, action func(string, int, time.Duration, time.Duration) (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [networkName] [retries] [retryDelay] [timeout]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			networkName := args[0]
			retries := ParseInt(args[1], 3)
			retryDelay := ParseDuration(args[2], 2*time.Second)
			timeout := ParseDuration(args[3], 30*time.Second)

			output, err := action(networkName, retries, retryDelay, timeout)
			HandleError(err)

			fmt.Println(output)
		},
	}
}

func createNetworkRemoveCommand(use, shortDesc string, action func(string, int, time.Duration, time.Duration) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " [networkName] [retries] [retryDelay] [timeout]",
		Short: shortDesc,
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			networkName := args[0]
			retries := ParseInt(args[1], 3)
			retryDelay := ParseDuration(args[2], 2*time.Second)
			timeout := ParseDuration(args[3], 30*time.Second)

			err := action(networkName, retries, retryDelay, timeout)
			HandleError(err)
		},
	}
}
