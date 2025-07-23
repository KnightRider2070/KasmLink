> [!WARNING]
> This project has been archived and will no longer be developed, as `KasmOrchestrator` serves as its successor.


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
For using kubernetes
- **kind**
- **helm**
- **kubectl**



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

````There is no bug in the code, it's a configuration issue ;)````

KasmLink interacts with your Kasm server using an API key and secret. You can either:

1. Update the `commands` package code to include your server URL, API key, and secret.
2. Or pass these values dynamically when executing commands:

   ```sh
   kasmlink --webApi-key your_api_key --webApi-secret your_api_secret
   ```

## Command Usage Guide

### 1. Initializing Folder Structures with `kasmlink init`

The `kasmlink init` command allows you to create various initial folder structures that can help set up your project.

#### Usage

```sh
kasmlink init [option] [folderPathForInit]
```

#### Options

- **all-templates**:
    - Creates a folder with preconfigured example files, including Dockerfiles and service templates.
- **dockerfiles-templates**:
    - Creates a folder with only preconfigured Dockerfiles.
- **empty-structure**:
    - Creates all required folders with their basic structure but leaves them empty.
- **service-templates**:
    - Creates a folder with only preconfigured service templates.

#### Arguments

- **folderPathForInit**: The path to the root folder where the specified initialization should occur.

#### Examples

1. **Initialize all folders with example templates:**
   ```sh
   kasmlink init all-templates /path/to/project
   ```
   This command will create `/path/to/project` with a complete set of folders containing Dockerfiles and service
   templates.

2. **Create an empty folder structure:**
   ```sh
   kasmlink init empty-structure /path/to/empty/project
   ```
   This will create `/path/to/empty/project` with the necessary folders (`services`, `dockerfiles`, etc.) but without
   any preconfigured content.

3. **Create a folder with service templates only:**
   ```sh
   kasmlink init service-templates /path/to/project
   ```
   This command creates the `services` folder with example service templates at `/path/to/project`.

### 2. Generating Docker Compose Files with `kasmlink compose generate`

The `kasmlink compose generate` command helps you create or update a Docker Compose file using a specified template. It
allows you to generate multiple instances of a service based on a given template.

#### Usage

```sh
kasmlink compose generate [composeFilePath] [templateFolderPath] [templateName] [count] [serviceNames...]
```

#### Arguments

- **composeFilePath**: The file path to the new or existing Docker Compose file that will be populated.
    - If the file doesn't exist, it will be created.
    - If the file exists, it will be updated with the newly generated services.

- **templateFolderPath**: The path to the root directory that contains the `services` and `dockerfiles` folders.
    - This directory should include all the available templates that can be used for service generation.

- **templateName**: The name of the service YAML file (within the `services` folder) that should be used for service
  generation.
    - Example: If you have a service template named `web-service.yaml`, specify `web-service` as the `templateName`.

- **count**: The number of services that should be generated from the specified template.
    - Example: `3` will generate three copies of the service template.

- **serviceNames...**: (Optional) Specific names for each generated service.
    - If you provide fewer names than the specified `count`, the remaining services will be auto-named by adding a
      suffix like `-1`, `-2`, etc.
    - If only one name is provided, it will be serialized (e.g., `service-1`, `service-2`).

#### Examples

1. **Create a new Docker Compose file with multiple services:**
   ```sh
   kasmlink compose generate /path/to/docker-compose.yaml /path/to/templates web-service 3
   ```
   This command will generate three instances of `web-service` in `/path/to/docker-compose.yaml`. Each service will have
   a generic serialized name, like `web-service-1`, `web-service-2`, etc.

2. **Update an existing Docker Compose file with specific service names:**
   ```sh
   kasmlink compose generate /path/to/docker-compose.yaml /path/to/templates webApi-service 2 webApi-v1 webApi-v2
   ```
   This command will add two instances of the `api-service` template, named `api-v1` and `api-v2`, to the existing
   compose file located at `/path/to/docker-compose.yaml`.

3. **Create services with partial custom naming:**
   ```sh
   kasmlink compose generate /path/to/docker-compose.yaml /path/to/templates database 3 db-master
   ```
   This command will create three services from the `database` template:
    - `db-master`
    - `db-master-1`
    - `db-master-2`

### Additional Notes:

- **Template Structure**:
    - Templates should be organized in the provided `templateFolderPath` under `services` and `dockerfiles` folders.
    - The `templateName` should match the name of a YAML file in the `services` directory.

- **Generating a Docker Compose File**:
    - The `generate` command will either create a new compose file or update an existing one.
    - If `composeFilePath` points to a non-existent file, a new one will be created with the specified version (`3.8`)
      and populated with the new services.

- **Atomic Updates**:
    - Updates to the compose file are done atomically by writing first to a temporary file and then renaming it to
      ensure that the operation is safe and won't leave the compose file in a corrupted state if something goes wrong.

- **Handling Errors**:
    - In case of invalid inputs or errors during file operations (such as file permission issues), meaningful error
      messages are logged to help troubleshoot the problem.
    - Ensure you have the correct permissions to create or modify files in the specified paths.

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
