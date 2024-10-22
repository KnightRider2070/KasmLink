package cmd

import (
	"fmt"
	"kasmlink/pkg/dockercli"

	"github.com/spf13/cobra"
)

var networkName string

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage Docker networks (create, inspect, remove, etc.)",
}

var createNetworkCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Docker network",
	Run: func(cmd *cobra.Command, args []string) {
		// Customize your network options here
		opts := dockercli.NetworkOptions{
			Name:   networkName,
			Driver: dockercli.DriverBridge,
		}
		err := dockercli.CreateDockerNetwork(opts)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	dockerCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(createNetworkCmd)

	// Flags for network commands
	createNetworkCmd.Flags().StringVarP(&networkName, "name", "n", "", "Network name")

	// Make network name required
	createNetworkCmd.MarkFlagRequired("name")
}
