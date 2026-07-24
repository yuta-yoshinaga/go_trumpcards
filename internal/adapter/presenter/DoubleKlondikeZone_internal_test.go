//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestDoubleKlondikeZoneName(t *testing.T) {
	i18n.SetLang("en")
	defer i18n.SetLang("ja")

	assert.Equal(t, "Waste", doubleKlondikeZoneName("waste"))
	assert.Equal(t, "Tableau", doubleKlondikeZoneName("tableau"))
	assert.Equal(t, "Foundation", doubleKlondikeZoneName("foundation"))
	// Unknown identifiers fall back to the raw string without panicking.
	assert.Equal(t, "stock", doubleKlondikeZoneName("stock"))
	assert.Equal(t, "", doubleKlondikeZoneName(""))
}
