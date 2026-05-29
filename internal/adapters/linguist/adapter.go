package linguist

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sazardev/gitto/internal/core/entities"
	"github.com/sazardev/gitto/internal/ports"
)

type Adapter struct {
	repoPath string
}

var skipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	".venv":         true,
	"venv":          true,
	"__pycache__":   true,
	".next":         true,
	".nuxt":         true,
	"dist":          true,
	"build":         true,
	"target":        true,
	".tox":          true,
	".mypy_cache":   true,
	".pytest_cache": true,
	".cache":        true,
}

func NewAdapter(repoPath string) *Adapter {
	return &Adapter{repoPath: repoPath}
}

func (a *Adapter) DetectLanguage() (entities.Language, bool) {
	var files []string

	err := filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(a.repoPath, path)
		if err != nil {
			return nil
		}

		base := filepath.Base(rel)
		if strings.HasPrefix(base, ".") && base != ".dockerfile" {
			return nil
		}

		files = append(files, rel)
		return nil
	})

	if err != nil || len(files) == 0 {
		return entities.Language{}, false
	}

	lang := entities.DetectLanguage(files)
	if lang.Name == "Unknown" {
		return lang, false
	}
	return lang, true
}

var _ ports.LanguageProvider = (*Adapter)(nil)
