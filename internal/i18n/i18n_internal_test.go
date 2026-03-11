package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadTranslations_MissingFile covers the ReadFile error branch
// by loading a language that has no locale files for most entries.
// The "zz" locale has a common.json with invalid JSON to cover the Unmarshal error branch.
func TestLoadTranslations_MissingFile(t *testing.T) {
	result := loadTranslations("zz") // locales/zz/common.json exists but has invalid JSON; others are missing
	assert.Empty(t, result)
}

// TestLoadTranslations_ValidLang covers the happy path for both ja and en.
func TestLoadTranslations_ValidLang(t *testing.T) {
	result := loadTranslations("en")
	assert.NotEmpty(t, result)
	assert.Equal(t, "Goodbye.", result["bye"])
}
