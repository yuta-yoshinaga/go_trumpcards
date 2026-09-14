//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
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

func TestSixCardGolfWebPresenter_CodedError(t *testing.T) {
	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "sixcardgolf.errCannotFlip", nil)
	result := new(presenter.SixCardGolfWebPresenter).Output(g, err)
	var output controller.SixCardGolfWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &output))
	assert.Empty(t, output.Message)
	assert.Equal(t, "sixcardgolf.errCannotFlip", output.MessageCode)
}
