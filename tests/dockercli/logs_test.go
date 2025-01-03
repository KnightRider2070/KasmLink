package dockercli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/dockercli"
)

func TestPrintBuildLogs(t *testing.T) {
	t.Run("Valid Build Logs", func(t *testing.T) {
		logs := `{"stream":"Step 1/3: FROM alpine:latest\n"}
{"stream":"Step 2/3: COPY . /app\n"}
{"stream":"Step 3/3: CMD [\"/bin/sh\"]\n"}`

		reader := bytes.NewReader([]byte(logs))
		err := dockercli.PrintBuildLogs(reader)
		assert.NoError(t, err)
	})

	t.Run("Build Logs with Error", func(t *testing.T) {
		logs := `{"error":"Something went wrong during the build"}`
		reader := bytes.NewReader([]byte(logs))

		err := dockercli.PrintBuildLogs(reader)
		assert.NoError(t, err)
	})

	t.Run("Invalid JSON Format", func(t *testing.T) {
		logs := `{"stream":"Step 1/3: FROM alpine:latest\n"
{"stream":"Step 2/3: COPY . /app\n"}` // Missing closing bracket

		reader := bytes.NewReader([]byte(logs))
		err := dockercli.PrintBuildLogs(reader)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding build logs")
	})

	t.Run("Empty Logs", func(t *testing.T) {
		logs := ``
		reader := bytes.NewReader([]byte(logs))

		err := dockercli.PrintBuildLogs(reader)
		assert.NoError(t, err)
	})
}
