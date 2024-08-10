package main

import (
	"runtime"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/version"
)

type VersionCmd struct{}

func (cmd *VersionCmd) Run(ctx *Globals) error {
	type kv struct {
		Key   string `json:"name"`
		Value string `json:"value" writer:",width:60"`
	}
	var metadata = []kv{}
	if version.GitSource != "" {
		metadata = append(metadata, kv{"source", version.GitSource})
	}
	if version.GitBranch != "" {
		metadata = append(metadata, kv{"branch", version.GitBranch})
	}
	if version.GitTag != "" {
		metadata = append(metadata, kv{"tag", version.GitTag})
	}
	if version.GitHash != "" {
		metadata = append(metadata, kv{"hash", version.GitHash})
	}
	if version.GoBuildTime != "" {
		metadata = append(metadata, kv{"build time", version.GoBuildTime})
	}
	metadata = append(metadata, kv{"go version", runtime.Version()})
	metadata = append(metadata, kv{"os", runtime.GOOS + "/" + runtime.GOARCH})

	return ctx.writer.Write(metadata)
}
