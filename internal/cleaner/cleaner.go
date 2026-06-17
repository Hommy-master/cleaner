package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cleaner/internal/config"
	"cleaner/internal/util"
)

// Cleaner performs configured file and directory cleanup.
type Cleaner struct {
	cfg    *config.Config
	logger *util.Logger
}

// New creates a Cleaner with the given configuration and logger.
func New(cfg *config.Config, logger *util.Logger) *Cleaner {
	if logger == nil {
		logger = util.GetLogger()
	}
	return &Cleaner{cfg: cfg, logger: logger}
}

// Run executes one cleanup cycle for all configured files and directories.
func (c *Cleaner) Run() {
	for _, file := range c.cfg.Files {
		c.cleanFile(file)
	}
	for _, dir := range c.cfg.Dirs {
		c.cleanDir(dir)
	}
}

func (c *Cleaner) cleanFile(path string) {
	kind, err := util.ClassifyPath(path)
	if err != nil {
		c.logger.Printf("ERROR: classify file %q: %v", path, err)
		return
	}

	switch kind {
	case util.PathMissing:
		c.logger.Printf("ERROR: file %q does not exist, skipped", path)
	case util.PathDirectory:
		c.logger.Printf("ERROR: %q is a directory, not a file, skipped", path)
	case util.PathFile:
		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			c.logger.Printf("ERROR: resolve absolute path for %q: %v", path, absErr)
			return
		}
		if err := os.Remove(path); err != nil {
			c.logger.Printf("ERROR: delete file %q: %v", absPath, err)
			return
		}
		c.logger.Printf("deleted file: %s", absPath)
	}
}

func (c *Cleaner) cleanDir(dirCfg config.DirConfig) {
	path := dirCfg.Path
	kind, err := util.ClassifyPath(path)
	if err != nil {
		c.logger.Printf("ERROR: classify directory %q: %v", path, err)
		return
	}

	switch kind {
	case util.PathMissing:
		c.logger.Printf("ERROR: directory %q does not exist, skipped", path)
		return
	case util.PathFile:
		c.logger.Printf("ERROR: %q is a file, not a directory, skipped", path)
		return
	case util.PathDirectory:
		rules, err := util.BuildIgnoreRules(path, dirCfg.Ignore)
		if err != nil {
			c.logger.Printf("ERROR: resolve ignore rules for %q: %v", path, err)
			return
		}
		if err := c.removeDirContents(path, rules, dirCfg.MinAgeSeconds); err != nil {
			c.logger.Printf("ERROR: clean directory %q: %v", path, err)
			return
		}
		c.logger.Debugf("cleaned directory: %s", path)
	}
}

func (c *Cleaner) removeDirContents(root string, rules []util.IgnoreRule, minAgeSeconds int) error {
	root = filepath.Clean(root)
	now := time.Now()

	if err := c.deleteEligibleFiles(root, rules, minAgeSeconds, now); err != nil {
		return err
	}
	return c.forceRemoveDirectories(root, rules)
}

func (c *Cleaner) deleteEligibleFiles(root string, rules []util.IgnoreRule, minAgeSeconds int, now time.Time) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relative path for %q: %w", path, err)
		}

		if util.IsIgnored(rel, rules) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			return fmt.Errorf("resolve absolute path for %q: %w", path, absErr)
		}

		shouldDelete, ageErr := util.ShouldDeleteFileByAge(path, minAgeSeconds, now)
		if ageErr != nil {
			return fmt.Errorf("check file age for %q: %w", absPath, ageErr)
		}
		if !shouldDelete {
			return nil
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove file %q: %w", absPath, err)
		}
		c.logger.Printf("deleted file: %s", absPath)
		return nil
	})
}

func (c *Cleaner) forceRemoveDirectories(root string, rules []util.IgnoreRule) error {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root || !d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relative path for %q: %w", path, err)
		}

		if util.IsIgnored(rel, rules) {
			return filepath.SkipDir
		}
		if util.CanForceRemoveDir(rel, rules) {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirDepth(root, dirs[i]) > dirDepth(root, dirs[j])
	})

	for _, dirPath := range dirs {
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		absPath, absErr := filepath.Abs(dirPath)
		if absErr != nil {
			return fmt.Errorf("resolve absolute path for %q: %w", dirPath, absErr)
		}

		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("remove directory %q: %w", absPath, err)
		}
		c.logger.Printf("deleted directory: %s", absPath)
	}

	return nil
}

func dirDepth(root, dirPath string) int {
	rel, err := filepath.Rel(root, dirPath)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}
