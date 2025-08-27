package collector

import (
	"context"
	"os"
	"os/user"
	"runtime"
	"time"
)

const EnvironmentCollectorName = "environment"

// EnvironmentCollector collects runtime environment information
type EnvironmentCollector struct {
	BaseCollector
	envVars []string
}

func NewEnvironmentCollector(envVars []string) *EnvironmentCollector {
	return &EnvironmentCollector{
		BaseCollector: NewBaseCollector(EnvironmentCollectorName, 20),
		envVars:       envVars,
	}
}

func (c *EnvironmentCollector) Collect(_ context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	result["is_ci"] = os.Getenv("CI") == "true"
	result["ci_mode"] = os.Getenv("CI")
	result["os"] = runtime.GOOS
	result["arch"] = runtime.GOARCH
	result["go_version"] = runtime.Version()
	result["num_cpu"] = runtime.NumCPU()
	result["timestamp"] = time.Now().Unix()
	result["timestamp_iso"] = time.Now().Format(time.RFC3339)

	if wd, err := os.Getwd(); err == nil {
		result["working_dir"] = wd
	}
	if hostname, err := os.Hostname(); err == nil {
		result["hostname"] = hostname
	}
	if u, err := user.Current(); err == nil {
		result["user"] = u.Username
		result["user_home"] = u.HomeDir
	}

	// collect user provided environment variables
	if len(c.envVars) > 0 {
		for _, key := range c.envVars {
			if value := os.Getenv(key); value != "" {
				result[key] = value
			}
		}
	}

	return result, nil
}
