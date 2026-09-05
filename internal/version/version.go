package version

import (
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version, Commit, Date, ProviderVersion string
}

// Current combines release build flags with metadata embedded by the Go toolchain.
func Current() Info {
	info, _ := debug.ReadBuildInfo()
	return resolve(info, Version, Commit, Date)
}

func resolve(info *debug.BuildInfo, version, commit, date string) Info {
	result := Info{version, commit, date, "unknown"}
	if info == nil {
		return result
	}
	localCheckout := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs":
			localCheckout = true
		case "vcs.revision":
			if result.Commit == "unknown" {
				result.Commit = setting.Value
			}
		}
	}
	// Go 1.24+ also stamps versions for local checkouts. Only module installs
	// without local VCS metadata should supply the fallback release version.
	if result.Version == "dev" && !localCheckout && info.Main.Version != "" && info.Main.Version != "(devel)" && info.Main.Replace == nil {
		result.Version = strings.TrimPrefix(info.Main.Version, "v")
	}
	for _, dep := range info.Deps {
		if dep.Path != "github.com/tamnd/threads-cli" {
			continue
		}
		switch {
		case dep.Replace != nil:
			result.ProviderVersion = "dev"
			if dep.Replace.Version != "" && dep.Replace.Version != "(devel)" {
				result.ProviderVersion = dep.Replace.Path + "@" + dep.Replace.Version
			}
		case dep.Version != "":
			result.ProviderVersion = dep.Version
		}
		break
	}
	return result
}
