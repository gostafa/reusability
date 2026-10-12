// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package version

import (
	"runtime/debug"
)

// Version returns the module version from build info, or "dev" for local builds.
func Version() string {
	return readVersion(debug.ReadBuildInfo)
}

func readVersion(read func() (*debug.BuildInfo, bool)) string {
	info, ok := read()

	if !ok {
		return fallbackVersion
	}

	return versionFromInfo(info.Main.Version)
}

func versionFromInfo(value string) string {
	if value == emptyVersion || value == develVersion {
		return fallbackVersion
	}

	return value
}
