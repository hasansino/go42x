package vendoring

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/pkg/config"
	"github.com/go42-dev/go42x/pkg/fspatch"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

// commandSettings assembles service settings at the CLI boundary.
func commandSettings() *vendoring.Settings {
	settings := vendoring.NewSettings()
	settings.RootPath = viper.GetString("root")
	settings.ManifestPath = viper.GetString("manifest")
	settings.LockPath = viper.GetString("lock")
	return settings
}

// validateCommandSettings loads project configuration and validates the effective
// settings before execution. Configuration failures retain structured reports.
func validateCommandSettings(f *cmdutil.Factory, settings *vendoring.Settings, operation string, asJSON bool) error {
	if err := settings.Validate(); err != nil {
		return cmdutil.UsageError(err)
	}
	if err := projectSettings(settings); err != nil {
		return report(f, emptyResult(operation), err, asJSON)
	}
	if err := settings.Validate(); err != nil {
		return report(f, emptyResult(operation), err, asJSON)
	}
	return nil
}

func projectSettings(settings *vendoring.Settings) error {
	root, err := filepath.Abs(settings.RootPath)
	if err != nil {
		return err
	}
	settings.RootPath = root
	cfg, err := config.LoadProjectConfig(filepath.Join(root, ".go42x", "go42x.yaml"), nil)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	settings.TempDir = cfg.Context.WithDefaults().TempDir
	return nil
}

func emptyResult(operation string) *vendoring.Result {
	return &vendoring.Result{Operation: operation, Changes: []fspatch.Change{}, Issues: []vendoring.Issue{}}
}

func report(f *cmdutil.Factory, result *vendoring.Result, runErr error, asJSON bool) error {
	if runErr != nil {
		result.Valid = false
		result.Issues = append(result.Issues, vendoring.Issue{Reason: runErr.Error()})
	}
	var err error
	if asJSON {
		err = cmdutil.JSON(f.Output(), result)
	} else {
		status := "complete"
		if !result.Valid {
			status = "failed"
		} else if result.DryRun {
			status = "dry-run"
		}
		_, err = fmt.Fprintf(
			f.Output(),
			"%s: %s; %d checked, %d planned, %d applied\n",
			result.Operation,
			status,
			result.Checked,
			len(result.Changes),
			result.Applied,
		)
		for _, issue := range result.Issues {
			if err == nil {
				_, err = fmt.Fprintf(f.Output(), "%s: %s\n", issue.Path, issue.Reason)
			}
		}
		if err == nil && result.BackupDir != "" {
			_, err = fmt.Fprintln(f.Output(), "Recovery backup:", result.BackupDir)
		}
	}
	if err != nil {
		return err
	}
	if !result.Valid {
		return &cmdutil.ExitError{Code: 1, Err: fmt.Errorf("vendoring %s failed", result.Operation), Reported: true}
	}
	return nil
}
