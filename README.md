# KasmLink

<div align="center">
  <img src="logo.webp" alt="KasmLink Logo"  width="200" />
</div>
KasmLink is a powerful command-line interface (CLI) tool that enables developers to seamlessly deploy GPU-accelerated containers and isolated environments for clients using Docker and Kasm technologies.

## Features

- **User Management**: Easily create, update, delete, and manage users within the Kasm environment.
- **Session Management**: Request, destroy, and monitor sessions with ease.
- **Execute Commands**: Run arbitrary commands inside a Kasm session.
- **SSH Connectivity**: Connect to running Kasm sessions over SSH for direct interaction.
- **Image Management**: List all available Docker images within the Kasm system.

## Requirements

- **Go**: Version 1.20 or later.
- **Docker** and **Kasm**: You need a running Docker environment and a Kasm server to interact with.

## Installation

### Build from Source

To build KasmLink from the source:

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/kasmlink.git
   cd kasmlink
   ```

2. Build the binary:
   ```sh
   go build -o kasmlink
   ```

3. (Optional) Add `kasmlink` to your system's `PATH` for easier access:
   ```sh
   mv kasmlink /usr/local/bin/
   ```

### Running KasmLink

You can now use the CLI by running:

```sh
kasmlink --help
```

This command will provide you with a list of all available subcommands and their usage.

## Configuration

KasmLink interacts with your Kasm server using an API key and secret. You can either:

1. Update the `commands` package code to include your server URL, API key, and secret.
2. Or pass these values dynamically when executing commands:

   ```sh
   kasmlink --api-key your_api_key --api-secret your_api_secret
   ```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have ideas or improvements.

## Testing

To run the unit tests:

```sh
go test ./...
```

Make sure to include tests when adding new functionality.

## CI/CD Pipeline

This project uses GitHub Actions to run tests and build the application on every push to the `main` branch. Pull
requests must pass all tests before being merged.

### Triggering a Release

To create a new version release:

1. Tag your commit with a version number:
   ```sh
   git tag v1.0.0
   git push origin v1.0.0
   ```

   This will trigger a build and release action to package KasmLink binaries for Windows and Linux.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Developed by KnightRider2070. Feel free to reach out if you have questions or feedback!

## Support

For any issues, please open an issue in the [GitHub repository](https://github.com/yourusername/kasmlink/issues).

## Acknowledgements

- **Kasm Technologies** for providing the foundation for containerized desktops.
- **Docker** for containerization support.
