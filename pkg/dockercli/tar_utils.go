package dockercli

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"io"
	"kasmlink/pkg/shadowssh"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// FileSystem abstracts file system operations to support local and remote files.
type FileSystem interface {
	Open(path string) (io.ReadCloser, error)
	Walk(root string, walkFn filepath.WalkFunc) error
}

// LocalFileSystem implements FileSystem for local file operations.
type LocalFileSystem struct{}

func NewLocalFileSystem() *LocalFileSystem { return &LocalFileSystem{} }

// Open opens a local file for reading.
func (l *LocalFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Walk walks the file tree rooted at root, calling walkFn for each file or directory.
func (l *LocalFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// RemoteFileSystem implements FileSystem for remote file systems over SSH.
type RemoteFileSystem struct {
	client *shadowssh.Client
}

func NewRemoteFileSystem(client *shadowssh.Client) *RemoteFileSystem {
	return &RemoteFileSystem{client: client}
}

// Open opens a remote file for reading via SFTP.
func (r *RemoteFileSystem) Open(path string) (io.ReadCloser, error) {
	sftpClient, err := sftp.NewClient(r.client.Client())
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	file, err := sftpClient.Open(path)
	if err != nil {
		sftpClient.Close() // Close client if file cannot be opened.
		return nil, fmt.Errorf("failed to open remote file %s: %w", path, err)
	}

	// Wrap file and client in a struct for proper cleanup.
	return &remoteFileWrapper{
		file:       file,
		sftpClient: sftpClient,
	}, nil
}

// Walk simulates filepath.Walk for remote file systems.
func (r *RemoteFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	sftpClient, err := sftp.NewClient(r.client.Client())
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	return r.walkRemote(sftpClient, root, walkFn)
}

// walkRemote recursively walks through the remote file system.
func (r *RemoteFileSystem) walkRemote(client *sftp.Client, dir string, walkFn filepath.WalkFunc) error {
	entries, err := client.ReadDir(dir)
	if err != nil {
		return walkFn(dir, nil, fmt.Errorf("failed to read directory %s: %w", dir, err))
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		if err := walkFn(fullPath, entry, nil); err != nil {
			if errors.Is(err, filepath.SkipDir) && entry.IsDir() {
				continue
			}
			return err
		}
		if entry.IsDir() {
			if err := r.walkRemote(client, fullPath, walkFn); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateTarWithContext creates a tar archive of the specified directory.
func (dc *DockerClient) CreateTarWithContext(buildContextDir string) (io.Reader, error) {
	if buildContextDir == "" {
		return nil, errors.New("build context directory cannot be empty")
	}

	buffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buffer)
	defer closeWriter(tarWriter)

	log.Info().Str("directory", buildContextDir).Msg("Creating tar archive for build context")

	err := dc.fs.Walk(buildContextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return logError(fmt.Errorf("error accessing %s: %w", path, err))
		}
		return addToTar(tarWriter, path, info, buildContextDir, dc.fs)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create tar archive: %w", err)
	}

	log.Info().Msg("Tar archive created successfully")
	return io.NopCloser(buffer), nil
}

// addToTar adds a file or directory to the tar archive.
func addToTar(tw *tar.Writer, path string, info os.FileInfo, baseDir string, fs FileSystem) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("error creating tar header for %s: %w", path, err)
	}

	header.Name, err = filepath.Rel(baseDir, path)
	if err != nil {
		return fmt.Errorf("error determining relative path for %s: %w", path, err)
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("error writing tar header for %s: %w", path, err)
	}

	if info.IsDir() {
		return nil
	}

	return writeFileToTar(tw, fs, path)
}

// writeFileToTar writes file content to the tar archive.
func writeFileToTar(tw *tar.Writer, fs FileSystem, path string) error {
	file, err := fs.Open(path)
	if err != nil {
		return logError(fmt.Errorf("failed to open %s: %w", path, err))
	}
	defer file.Close()

	if _, err := io.Copy(tw, file); err != nil {
		return logError(fmt.Errorf("failed to copy contents of %s: %w", path, err))
	}

	log.Debug().Str("file", path).Msg("Added to tar archive successfully")
	return nil
}

// closeWriter safely closes a writer.
func closeWriter(w io.Closer) {
	if err := w.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close writer")
	}
}

// logError logs and returns an error.
func logError(err error) error {
	log.Error().Err(err).Msg("Error encountered")
	return err
}

// remoteFileWrapper wraps an SFTP file and its client for cleanup.
type remoteFileWrapper struct {
	file       *sftp.File
	sftpClient *sftp.Client
}

// Read delegates the read operation to the underlying SFTP file.
func (rw *remoteFileWrapper) Read(p []byte) (int, error) {
	return rw.file.Read(p)
}

// Close closes the SFTP file and the client.
func (rw *remoteFileWrapper) Close() error {
	fileErr := rw.file.Close()
	clientErr := rw.sftpClient.Close()

	if fileErr != nil {
		return fileErr
	}
	return clientErr
}
