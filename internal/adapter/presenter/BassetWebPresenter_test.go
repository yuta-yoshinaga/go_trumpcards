package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBassetWebPresenter_OutputUsesBassetJSONFields(t *testing.T) {
	g := domain.NewDefaultBasset()
	out := new(BassetWebPresenter).Output(g, nil)
	assert.Contains(t, out, "remainingByRank")
}
