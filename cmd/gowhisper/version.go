package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	// Packages
	version "github.com/mutablelogic/go-whisper/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func VersionJSON() string {
	// Get executable name
	name := "gowhisper"
	if exe, err := os.Executable(); err == nil {
		name = filepath.Base(exe)
	}

	metadata := map[string]string{
		"name":       name,
		"compiler":   runtime.Version(),
		"source":     version.GitSource,
		"tag":        version.GitTag,
		"branch":     version.GitBranch,
		"hash":       version.GitHash,
		"build_time": version.GoBuildTime,
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
