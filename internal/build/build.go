package build

import (
	"runtime/debug"
)

var (
	xBuildCommit  = "dev"
	xBuildVersion = "dev"
)

func GetCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return xBuildCommit
}

func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}
	return xBuildVersion
}
