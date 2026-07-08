package main

import (
	"fmt"
	"runtime/debug"
)

func versionString() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "pawnfmt (unknown version)"
	}

	version := info.Main.Version
	if version == "" || version == "(devel)" {
		version = "devel"
	}

	var revision string

	dirty := false

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}

	if revision == "" {
		return "pawnfmt " + version
	}

	if len(revision) > 12 {
		revision = revision[:12]
	}

	if dirty {
		revision += "-dirty"
	}

	return fmt.Sprintf("pawnfmt %s (%s)", version, revision)
}
