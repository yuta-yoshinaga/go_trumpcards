//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSixCardGolfWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	p := new(presenter.SixCardGolfWebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(g, nil), p.HintOutput(g))
}
