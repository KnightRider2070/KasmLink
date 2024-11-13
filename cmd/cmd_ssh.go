package cmd

import (
	"fmt"
	shadowssh "kasmlink/pkg/ssh"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	// Define the "shadowssh" command
	shadowSSHCMD := &cobra.Command{
		Use:   "shadowssh",
		Short: "Manage SSH connections to remote nodes",
		Long:  `Use this command to manage SSH-related functionalities including establishing SSH connections, executing commands, and copying files via SCP.`,
	}

	// Add subcommands for various SSH utilities
	shadowSSHCMD.AddCommand(
		createSSHClientCommand(),
		createSSHExecuteCommand(),
		createSSHExecuteWithOutputCommand(),
	)

	// Add "shadowssh" to the root command
	RootCmd.AddCommand(shadowSSHCMD)
}

// createSSHClientCommand creates a command to establish an SSH connection
func createSSHClientCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect [username] [password] [nodeAddress] [knownHostsFile]",
		Short: "Establish an SSH connection to a remote node",
		Long: `Establish an SSH connection using specified username, password, node address, and known_hosts file.
If connection is successful, the user can use the same client to execute commands remotely.`,
		Args: cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			sshConfig := &shadowssh.SSHConfig{
				Username:          args[0],
				Password:          args[1],
				NodeAddress:       args[2],
				KnownHostsFile:    args[3],
				ConnectionTimeout: 10 * time.Second,
			}

			client, err := shadowssh.NewSSHClient(sshConfig)
			if err != nil {
				HandleError(err)
				return
			}
			defer func() {
				if err := client.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to close SSH client")
				}
			}()

			log.Info().Msg("SSH connection established successfully")
		},
	}
}

// createSSHExecuteCommand creates a command to execute a remote command over SSH
func createSSHExecuteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec [username] [password] [nodeAddress] [knownHostsFile] [command]",
		Short: "Execute a command over SSH on a remote node",
		Long: `Execute a command on a remote node using SSH.
Provide username, password, node address, known_hosts file, and the command to execute.`,
		Args: cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			sshConfig := &shadowssh.SSHConfig{
				Username:          args[0],
				Password:          args[1],
				NodeAddress:       args[2],
				KnownHostsFile:    args[3],
				ConnectionTimeout: 10 * time.Second,
			}

			client, err := shadowssh.NewSSHClient(sshConfig)
			if err != nil {
				HandleError(err)
				return
			}
			defer func() {
				if err := client.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to close SSH client")
				}
			}()

			command := args[4]
			output, err := shadowssh.ExecuteCommand(client, command)
			if err != nil {
				HandleError(err)
				return
			}

			fmt.Println("Command Output:")
			fmt.Println(output)
		},
	}
}

// createSSHExecuteWithOutputCommand creates a command to execute a command with a specific logging duration
func createSSHExecuteWithOutputCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec-with-output [username] [password] [nodeAddress] [knownHostsFile] [command] [logDuration]",
		Short: "Execute a command over SSH and log output for a specified duration",
		Long: `Execute a command on a remote node using SSH and log the output for a specified amount of time.
Provide username, password, node address, known_hosts file, the command to execute, and the logging duration.`,
		Args: cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			sshConfig := &shadowssh.SSHConfig{
				Username:          args[0],
				Password:          args[1],
				NodeAddress:       args[2],
				KnownHostsFile:    args[3],
				ConnectionTimeout: 10 * time.Second,
			}

			client, err := shadowssh.NewSSHClient(sshConfig)
			if err != nil {
				HandleError(err)
				return
			}
			defer func() {
				if err := client.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to close SSH client")
				}
			}()

			command := args[4]
			logDuration := ParseDuration(args[5], 30*time.Second)

			output, err := shadowssh.ShadowExecuteCommandWithOutput(client, command, logDuration)
			if err != nil {
				HandleError(err)
				return
			}

			fmt.Println("Command Output:")
			fmt.Println(output)
		},
	}
}
