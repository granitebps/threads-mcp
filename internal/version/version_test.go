package version

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildVersions(t *testing.T) {
	installed := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.0.0"},
		Deps: []*debug.Module{{Path: "github.com/tamnd/threads-cli", Version: "v0.1.1"}},
	}
	checkout := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.0.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs", Value: "git"},
			{Key: "vcs.revision", Value: "abcdef123456"},
			{Key: "vcs.time", Value: "2026-09-04T00:00:00Z"},
		},
	}
	for _, test := range []struct {
		name                  string
		info                  *debug.BuildInfo
		version, commit, date string
		want                  Info
	}{
		{"versioned install", installed, "dev", "unknown", "unknown", Info{"1.0.0", "unknown", "unknown", "v0.1.1"}},
		{"tagged local checkout", checkout, "dev", "unknown", "unknown", Info{"dev", "abcdef123456", "unknown", "unknown"}},
		{"release checkout overrides", checkout, "1.1.0", "release-sha", "release-date", Info{"1.1.0", "release-sha", "release-date", "unknown"}},
		{"release overrides", installed, "1.1.0", "release-sha", "release-date", Info{"1.1.0", "release-sha", "release-date", "v0.1.1"}},
		{"local checkout", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev", "unknown", "unknown", Info{"dev", "unknown", "unknown", "unknown"}},
		{"no metadata", nil, "dev", "unknown", "unknown", Info{"dev", "unknown", "unknown", "unknown"}},
		{"local replacement", &debug.BuildInfo{Deps: []*debug.Module{{Path: "github.com/tamnd/threads-cli", Version: "v0.1.1", Replace: &debug.Module{Path: "../threads-cli"}}}}, "dev", "unknown", "unknown", Info{"dev", "unknown", "unknown", "dev"}},
		{"versioned replacement", &debug.BuildInfo{Deps: []*debug.Module{{Path: "github.com/tamnd/threads-cli", Version: "v0.1.1", Replace: &debug.Module{Path: "example.com/fork", Version: "v0.2.0"}}}}, "dev", "unknown", "unknown", Info{"dev", "unknown", "unknown", "example.com/fork@v0.2.0"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := resolve(test.info, test.version, test.commit, test.date); got != test.want {
				t.Fatalf("resolve() = %+v, want %+v", got, test.want)
			}
		})
	}
}
