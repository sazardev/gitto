package ports

import "github.com/sazardev/gitto/internal/core/entities"

type LanguageProvider interface {
	DetectLanguage() (entities.Language, bool)
}
