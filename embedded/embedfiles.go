package embedfiles

import "embed"

//go:embed workspace-core-image/*
var EmbeddedKasmDirectory embed.FS

//go:embed services/*
var EmbeddedServicesDirectory embed.FS

//go:embed dockerfiles/*
var EmbeddedDockerImagesDirectory embed.FS

//go:embed templates/*
var EmbeddedTemplateFS embed.FS

//go:embed compose-spec.json
var ComposeSpec []byte
