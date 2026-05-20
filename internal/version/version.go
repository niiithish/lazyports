package version

import (
	"runtime/debug"
	"strings"
)

const (
	// Version is the app release version. Bump when tagging a release.
	Version = "0.1.1"

	// InstallCommand is the recommended one-liner install/update command.
	InstallCommand = "curl -fsSL https://raw.githubusercontent.com/niiithish/lazyports/HEAD/install.sh | bash"
)

// Current returns the version string for this binary.
func Current() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := strings.TrimSpace(info.Main.Version)
		if v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0-") {
			return strings.TrimPrefix(v, "v")
		}
	}
	return Version
}
