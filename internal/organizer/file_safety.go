package organizer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// resolveFuturePath resolves existing ancestors without creating a missing destination.
func resolveFuturePath(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	// A dangling symlink is not a missing directory that we can safely create.
	if _, statErr := os.Lstat(abs); statErr == nil {
		return "", err
	}
	parent := filepath.Dir(abs)
	if parent == abs {
		return "", err
	}
	resolvedParent, err := resolveFuturePath(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(abs)), nil
}

func (o *Organizer) validateDestination(target string) error {
	base := o.config.OutputDir
	if base == "" {
		base = o.config.BaseDir
	}
	root, err := resolveFuturePath(base)
	if err != nil {
		return err
	}
	resolved, err := resolveFuturePath(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("destination %s is outside output directory %s", target, root)
	}
	return nil
}

func requireUnoccupied(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("destination already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// moveNoReplace publishes the destination without replacing an existing file.
// A hard link is atomic on one filesystem; exclusive copying handles other filesystems.
func moveNoReplace(source, target string) error {
	if filepath.Clean(source) == filepath.Clean(target) {
		return nil
	}
	if err := requireUnoccupied(target); err != nil {
		return err
	}
	if err := os.Link(source, target); err == nil {
		if err := os.Remove(source); err != nil {
			return errors.Join(err, os.Remove(target))
		}
		return nil
	} else if os.IsExist(err) {
		return err
	}
	return copyNoReplace(source, target)
}

func copyNoReplace(source, target string) (err error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("cannot copy non-regular file %s", source)
	}
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	complete := false
	defer func() {
		_ = output.Close()
		if !complete {
			err = errors.Join(err, os.Remove(target))
		}
	}()
	if _, err = io.Copy(output, input); err != nil {
		return err
	}
	if err = output.Sync(); err != nil {
		return err
	}
	if err = output.Close(); err != nil {
		return err
	}
	if err = input.Close(); err != nil {
		return err
	}
	if err = os.Remove(source); err != nil {
		return err
	}
	complete = true
	return nil
}

func writeLogAtomic(path string, data []byte) (err error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".abook-log-*")
	if err != nil {
		return err
	}
	defer func() { _ = file.Close(); _ = os.Remove(file.Name()) }()
	if err = file.Chmod(0o644); err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
