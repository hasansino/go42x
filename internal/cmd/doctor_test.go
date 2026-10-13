package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/doctor"
)

func TestDoctorRuntimeErrorUsesJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	factory := cmdutil.NewFactory(ctx)
	var output, errorOutput bytes.Buffer
	factory.SetIOStreams(bytes.NewReader(nil), &output, &errorOutput)
	settings := doctor.NewSettings()
	settings.RootPath, settings.JSON = t.TempDir(), true
	err := runDoctorCommand(factory, settings)
	if cmdutil.ExitCode(err) != 1 {
		t.Fatalf("unexpected status: %v", err)
	}
	var failure struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(
		output.Bytes(),
		&failure,
	); err != nil ||
		!strings.Contains(failure.Error, "context canceled") {
		t.Fatalf("invalid JSON failure: %s, %v", &output, err)
	}
	if errorOutput.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", &errorOutput)
	}
}
