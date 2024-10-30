package embedfiles

import "embed"

//go:embed workspace-core-image/*
var EmbeddedKasmDirectory embed.FS

//go:embed dockerImages/*
var EmbeddedDockerImagesDirectory embed.FS
