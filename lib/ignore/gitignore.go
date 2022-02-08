package ignore

import (
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/logger"
	"github.com/syncthing/syncthing/lib/sha256"
	"github.com/syncthing/syncthing/lib/sync"
)

var l = logger.DefaultLogger.NewFacility("gitignore", "gitignore support")

type gitMatcher struct {
	filesystem fs.Filesystem
	matcher    gitignore.Matcher
	hash       string
	mutex      sync.Mutex
}

func NewGitignore(fs fs.Filesystem, opts ...Option) *gitMatcher {
	return &gitMatcher{filesystem: fs, matcher: nil, mutex: sync.NewMutex()}
}

func (m *gitMatcher) Load(_ string) error {
	// TODO
	ps, err := gitignore.ReadPatterns(osfs.New(m.filesystem.URI()), []string{})
	if err != nil {
		return err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.matcher = gitignore.NewMatcher(ps)

	// TODO this is extremely inefficient
	hasher := sha256.New()
	for _, p := range ps {
		l.Debugf("pattern: %#v", p)
		_, _ = hasher.Write([]byte(fmt.Sprintf("%#v", p)))
	}
	m.hash = string(hex.EncodeToString(hasher.Sum([]byte{})))
	l.Debugf("loaded %d patterns, hash=%s", len(ps), m.hash)
	return nil
}

func (m *gitMatcher) Match(filename string) Result {
	if m.ShouldIgnore(filename) {
		return resultInclude
	}
	return resultNotMatched
}

func (m *gitMatcher) Hash() string {
	// TODO
	return m.hash
}

func (m *gitMatcher) ShouldIgnore(path string) bool {
	// TODO this is slow as f
	info, err := m.filesystem.Stat(path)
	isDir := err == nil && info.IsDir()
	parts := filepath.SplitList(path)
	// TODO not pretty
	if parts[0] == ".git" {
		return true
	}
	shouldIgnore := m.matcher.Match(parts, isDir)
	l.Debugf("ShouldIgnore(%s, isDir=%v) = %v", path, isDir, shouldIgnore)
	return shouldIgnore
}

func (m *gitMatcher) SkipIgnoredDirs() bool {
	return true
}

func (m *gitMatcher) Lines() []string {
	//TODO
	return []string{}
}

func (m *gitMatcher) Patterns() []string {
	//TODO
	return []string{}
}
