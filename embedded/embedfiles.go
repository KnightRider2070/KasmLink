package embedfiles

import "embed"

//go:embed workspace-core-image/*
var DockerFilesKasm embed.FS
