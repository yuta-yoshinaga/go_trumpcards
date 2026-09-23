package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBassetCuiPresenter_Output(t *testing.T) {
	g := domain.NewDefaultBasset()
	out := new(BassetCuiPresenter).Output(g, nil)
	assert.Contains(t, out, "1000")
}
