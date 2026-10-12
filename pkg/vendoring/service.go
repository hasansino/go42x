// Package vendoring maintains selected upstream files using reproducible locks.
package vendoring

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"github.com/go42-dev/go42x/pkg/fspatch"
	"github.com/go42-dev/go42x/pkg/vendoring/source"
	"github.com/go42-dev/go42x/pkg/vendoring/source/download"
	gitsource "github.com/go42-dev/go42x/pkg/vendoring/source/git"
	httpsource "github.com/go42-dev/go42x/pkg/vendoring/source/http"
	"github.com/go42-dev/go42x/pkg/vendoring/source/npm"
)

type Service struct {
	cliVersion string
	settings   Settings
	logger     *slog.Logger
	adapters   map[string]source.Adapter
}

type Option func(*Service)

func WithVersion(version string) Option {
	return func(s *Service) { s.cliVersion = version }
}

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) { s.logger = logger }
}

type OperationOptions struct {
	DryRun bool
	Force  bool
}

type Issue struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type Result struct {
	Operation string           `json:"operation"`
	Valid     bool             `json:"valid"`
	DryRun    bool             `json:"dry_run"`
	Checked   int              `json:"checked"`
	Applied   int              `json:"applied"`
	Changes   []fspatch.Change `json:"changes"`
	Issues    []Issue          `json:"issues"`
	BackupDir string           `json:"backup_dir,omitempty"`
}

func newResult(operation string) *Result {
	return &Result{Operation: operation, Changes: []fspatch.Change{}, Issues: []Issue{}}
}

func NewService(settings *Settings, options ...Option) (*Service, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	resolved := *settings
	if err := resolved.resolve(); err != nil {
		return nil, err
	}
	client := download.NewClient(resolved.Timeout)
	s := &Service{
		cliVersion: "unknown",
		settings:   resolved,
		logger:     slog.Default(),
		adapters: map[string]source.Adapter{
			"git": gitsource.New(client), "http": httpsource.New(client), "npm": npm.New(client),
		},
	}
	for _, option := range options {
		option(s)
	}
	if s.logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return s, nil
}

type inputs struct {
	manifest *Manifest
	lock     *Lock
	plan     *fspatch.Plan
	owned    map[string]string
}

func (s *Service) load(requireLock bool) (*inputs, error) {
	p := fspatch.NewPlan(
		s.logger,
		s.settings.RootPath,
		filepath.Join(s.settings.RootPath, ".go42x", "backups", "vendor"),
	)
	for _, filename := range []string{s.settings.ManifestPath, s.settings.LockPath} {
		if err := safePath(s.settings.RootPath, filename); err != nil {
			return nil, err
		}
		if s.reserved(filename) {
			return nil, fmt.Errorf("metadata collides with a control directory: %s", filename)
		}
	}
	data, exists, err := p.Read(s.settings.ManifestPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("manifest is missing: %s", s.settings.ManifestPath)
	}
	m, err := ParseManifest(data)
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	in := &inputs{manifest: m, plan: p, owned: map[string]string{}}
	data, exists, err = p.Read(s.settings.LockPath)
	if err != nil {
		return nil, err
	}
	if exists {
		in.lock, err = ParseLock(data)
		if err != nil {
			return nil, fmt.Errorf("lock: %w", err)
		}
		owner, err := s.lockOwner(s.settings.LockPath, in.lock.Manifest)
		if err != nil {
			return nil, err
		}
		if owner != s.settings.ManifestPath {
			return nil, fmt.Errorf("lock belongs to a different manifest")
		}
		for _, d := range in.lock.Dependencies {
			for destination, checksum := range d.Files {
				in.owned[destination] = checksum
			}
		}
	} else if requireLock {
		return nil, fmt.Errorf("lock is missing; run vendor update")
	}
	ownedNames := make(map[string]string, len(in.owned))
	for name := range in.owned {
		ownedNames[strings.ToLower(name)] = name
	}
	for _, d := range m.Dependencies {
		for _, f := range d.Files {
			if previous, ok := ownedNames[strings.ToLower(f.To)]; ok && previous != f.To {
				return nil, fmt.Errorf("case-only destination changes are unsupported: %s -> %s", previous, f.To)
			}
			if err := s.validateTarget(f.To); err != nil {
				return nil, err
			}
		}
	}
	for name := range in.owned {
		if err := s.validateTarget(name); err != nil {
			return nil, err
		}
	}
	return in, nil
}

func (s *Service) target(relative string) string {
	return filepath.Join(filepath.Dir(s.settings.ManifestPath), filepath.FromSlash(relative))
}

func (s *Service) validateTarget(relative string) error {
	filename := s.target(relative)
	if overlap(filename, s.settings.ManifestPath) || overlap(filename, s.settings.LockPath) ||
		s.reserved(filename) {
		return fmt.Errorf("destination collides with control paths: %s", relative)
	}
	return safePath(filepath.Dir(s.settings.ManifestPath), filename)
}

func within(root, filename string) bool {
	rel, err := filepath.Rel(root, filename)
	return err == nil && (rel == "." || filepath.IsLocal(rel))
}

func overlap(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	return within(a, b) || within(b, a)
}

func (s *Service) reserved(filename string) bool {
	root := s.settings.RootPath
	controls := []string{
		filepath.Join(root, ".go42x", "go42x.yaml"), filepath.Join(root, ".go42x", "go42x.local.yaml"),
	}
	for _, control := range append(controls, s.controlDirectories()...) {
		if overlap(control, filename) {
			return true
		}
	}
	return false
}

func (s *Service) controlDirectories() []string {
	return []string{
		filepath.Join(s.settings.RootPath, ".git"),
		filepath.Join(s.settings.RootPath, ".go42x", "backups"),
		filepath.Join(s.settings.TempDir, "vendoring"),
	}
}

func (s *Service) skipDirectory(filename string) bool {
	for _, directory := range s.controlDirectories() {
		if within(strings.ToLower(directory), strings.ToLower(filename)) {
			return true
		}
	}
	return false
}

func (s *Service) lockOwner(filename, pointer string) (string, error) {
	if filepath.IsAbs(pointer) || strings.ContainsAny(pointer, "\\\x00") {
		return "", fmt.Errorf("invalid lock manifest path")
	}
	owner := filepath.Clean(filepath.Join(filepath.Dir(filename), filepath.FromSlash(pointer)))
	if !within(s.settings.RootPath, owner) || owner == s.settings.RootPath {
		return "", fmt.Errorf("locked manifest escapes project root")
	}
	return owner, nil
}

func (s *Service) Check(ctx context.Context) (*Result, error) {
	result := newResult("check")
	if err := ctx.Err(); err != nil {
		return result, err
	}
	in, err := s.load(true)
	if err != nil {
		return result, err
	}
	if err := in.lock.Matches(in.manifest); err != nil {
		result.Issues = append(result.Issues, Issue{s.settings.LockPath, err.Error()})
	}
	for _, name := range slices.Sorted(maps.Keys(in.owned)) {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		data, exists, err := in.plan.Read(s.target(name))
		if err != nil {
			return result, err
		}
		result.Checked++
		if !exists {
			result.Issues = append(result.Issues, Issue{name, "missing owned file"})
		} else if hash(data) != in.owned[name] {
			result.Issues = append(result.Issues, Issue{name, "content differs from lock"})
		}
	}
	result.Valid = len(result.Issues) == 0
	return result, nil
}

func (s *Service) Sync(ctx context.Context, options OperationOptions) (*Result, error) {
	return s.run(ctx, "sync", options)
}
func (s *Service) Update(ctx context.Context, options OperationOptions) (*Result, error) {
	return s.run(ctx, "update", options)
}

func (s *Service) run(ctx context.Context, operation string, options OperationOptions) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, s.settings.Timeout)
	defer cancel()
	result := newResult(operation)
	result.DryRun = options.DryRun
	// Lock the existing root directory: one writer per project, independent of
	// manifest and temp-dir choices, with no new state files during a dry-run.
	writer := flock.New(s.settings.RootPath, flock.SetFlag(os.O_RDONLY))
	locked, err := writer.TryLockContext(ctx, 20*time.Millisecond)
	if err != nil {
		return result, err
	}
	if !locked {
		return result, fmt.Errorf("project writer lock unavailable")
	}
	defer func() { _ = writer.Close() }()
	in, err := s.load(operation == "sync")
	if err != nil {
		return result, err
	}
	if operation == "sync" {
		if err := in.lock.Matches(in.manifest); err != nil {
			return result, err
		}
	}
	if err := s.checkOwnership(ctx, in); err != nil {
		return result, err
	}
	if err := s.snapshot(in, options.Force); err != nil {
		return result, err
	}
	workspaceRoot := filepath.Join(s.settings.TempDir, "vendoring")
	if err := os.MkdirAll(workspaceRoot, 0700); err != nil {
		return result, err
	}
	workspace, err := os.MkdirTemp(workspaceRoot, "operation-*")
	if err != nil {
		return result, err
	}
	defer func() { _ = os.RemoveAll(workspace) }()
	candidate, contents, err := s.fetch(ctx, in, operation, workspace)
	if err != nil {
		return result, err
	}
	if err := s.prepare(in, candidate, contents, operation, options.Force); err != nil {
		return result, err
	}
	result.Changes = in.plan.Changes()
	result.Checked = len(contents)
	if err := in.plan.Validate(); err != nil {
		return result, err
	}
	if options.DryRun {
		result.Valid = true
		return result, nil
	}
	for name := range in.owned {
		if err := s.validateTarget(name); err != nil {
			return result, err
		}
	}
	for name := range contents {
		if err := s.validateTarget(name); err != nil {
			return result, err
		}
	}
	if len(result.Changes) > 0 {
		backupProbe := filepath.Join(s.settings.RootPath, ".go42x", "backups", "vendor", "updates", ".probe")
		if err := safePath(s.settings.RootPath, backupProbe); err != nil {
			return result, err
		}
	}
	applied, err := in.plan.ApplyUpdate(ctx, s.cliVersion)
	if applied != nil {
		result.Applied = applied.Applied
		result.BackupDir = applied.BackupDir
	}
	if err != nil {
		return result, fmt.Errorf(
			"application stopped after %d changes; recovery backup %s: %w",
			result.Applied,
			result.BackupDir,
			err,
		)
	}
	result.Valid = true
	return result, nil
}

func (s *Service) snapshot(in *inputs, force bool) error {
	paths := map[string]bool{}
	for name := range in.owned {
		paths[name] = true
	}
	for _, d := range in.manifest.Dependencies {
		for _, f := range d.Files {
			paths[f.To] = true
		}
	}
	for _, name := range slices.Sorted(maps.Keys(paths)) {
		data, exists, err := in.plan.Read(s.target(name))
		if err != nil {
			return err
		}
		if in.owned[name] != "" {
			if err := checkConflict(name, data, exists, in.owned[name], nil, force); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) fetch(
	ctx context.Context,
	in *inputs,
	operation, workspace string,
) (*Lock, map[string][]byte, error) {
	digest, err := in.manifest.Digest()
	if err != nil {
		return nil, nil, err
	}
	pointer, err := filepath.Rel(filepath.Dir(s.settings.LockPath), s.settings.ManifestPath)
	if err != nil {
		return nil, nil, err
	}
	lock := &Lock{
		Manifest:       filepath.ToSlash(pointer),
		ManifestSHA256: digest,
		Dependencies:   map[string]LockedDependency{},
	}
	contents := map[string][]byte{}
	for _, id := range slices.Sorted(maps.Keys(in.manifest.Dependencies)) {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		d := in.manifest.Dependencies[id]
		var pinned *source.Resolved
		if operation == "sync" {
			identity := in.lock.Dependencies[id].Resolved
			pinned = &identity
		}
		adapter := s.adapters[d.Source.Type]
		if adapter == nil {
			return nil, nil, fmt.Errorf("source adapter is missing: %s", d.Source.Type)
		}
		fetched, err := adapter.Fetch(ctx, d.Source, pinned, filepath.Join(workspace, id))
		if err != nil {
			return nil, nil, fmt.Errorf("fetch %s: %w", id, err)
		}
		entry := LockedDependency{
			Resolved: fetched.Resolved,
			Files:    make(map[string]string, len(d.Files)),
		}
		for _, f := range d.Files {
			data, err := selectFile(fetched.Root, f)
			if err != nil {
				return nil, nil, fmt.Errorf("select %s:%s: %w", id, f.From, err)
			}
			hash := hash(data)
			if operation == "sync" && hash != in.owned[f.To] {
				return nil, nil, fmt.Errorf("locked source file hash differs: %s", f.From)
			}
			entry.Files[f.To] = hash
			contents[f.To] = data
		}
		lock.Dependencies[id] = entry
	}
	return lock, contents, lock.Validate()
}

func (s *Service) prepare(
	in *inputs,
	candidate *Lock,
	contents map[string][]byte,
	operation string,
	force bool,
) error {
	for _, name := range slices.Sorted(maps.Keys(contents)) {
		data, exists, err := in.plan.Read(s.target(name))
		if err != nil {
			return err
		}
		if err := checkConflict(name, data, exists, in.owned[name], contents[name], force); err != nil {
			return err
		}
		if err := in.plan.Write(s.target(name), contents[name], fspatch.Sources, false); err != nil {
			return err
		}
	}
	for _, name := range slices.Sorted(maps.Keys(in.owned)) {
		if _, ok := contents[name]; !ok {
			if err := in.plan.Remove(s.target(name), fspatch.Sources); err != nil {
				return err
			}
		}
	}
	if operation == "update" {
		data, err := candidate.Marshal()
		if err != nil {
			return err
		}
		if in.lock != nil {
			original, err := in.lock.Marshal()
			if err != nil {
				return err
			}
			if bytes.Equal(original, data) {
				return nil
			}
		}
		return in.plan.Write(s.settings.LockPath, data, fspatch.Settings, false)
	}
	return nil
}
