package i18n

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

// TestLoadTranslations_MissingFile covers the ReadFile error branch
// by loading a language with no locale files.
func TestLoadTranslations_MissingFile(t *testing.T) {
	emptyFS := fstest.MapFS{}
	result := loadTranslations(emptyFS, "ja")
	assert.Empty(t, result)
}

// TestLoadTranslations_InvalidJSON covers the json.Unmarshal error branch.
func TestLoadTranslations_InvalidJSON(t *testing.T) {
	badFS := fstest.MapFS{
		"locales/xx/common.json": &fstest.MapFile{Data: []byte("{invalid json")},
	}
	result := loadTranslations(badFS, "xx")
	assert.Empty(t, result)
}

// TestLoadTranslations_ValidLang covers the happy path for both ja and en.
func TestLoadTranslations_ValidLang(t *testing.T) {
	result := loadTranslations(localesFS, "en")
	assert.NotEmpty(t, result)
	assert.Equal(t, "Goodbye.", result["bye"])
}
