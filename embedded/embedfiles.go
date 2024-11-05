package embedfiles

import "embed"

//go:embed workspace-core-image/*
var EmbeddedKasmDirectory embed.FS

//go:embed dockerfiles/*
var EmbeddedDockerImagesDirectory embed.FS

//go:embed templates/*
var EmbeddedTemplateFS embed.FS
