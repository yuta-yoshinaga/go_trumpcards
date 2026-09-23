//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDoubleExposurePresenterBlackJack(t *testing.T) *domain.BlackJack {
	t.Helper()
	bj := domain.NewDefaultBlackJack()
	cfg := bj.GetConfig()
	cfg.Variant = domain.BJVariantDoubleExposure
	require.NoError(t, bj.SetConfig(cfg))
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	bj.SetPhase(domain.BJPhaseAction)
	return bj
}

func TestBlackJackDoubleExposureWebPresenterShowsDealerHoleCard(t *testing.T) {
	bj := setupDoubleExposurePresenterBlackJack(t)
	output := (&BlackJackWebPresenter{}).Output(bj, nil)

	var response struct {
		Dealer struct {
			Score int               `json:"score"`
			Cards []json.RawMessage `json:"cards"`
		} `json:"dealer"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	assert.Equal(t, 18, response.Dealer.Score)
	assert.Len(t, response.Dealer.Cards, 2)
}

func TestBlackJackDoubleExposureCuiPresenterShowsDealerHoleCard(t *testing.T) {
	bj := setupDoubleExposurePresenterBlackJack(t)
	output := (&BlackJackCuiPresenter{}).Output(bj, nil)

	assert.NotContains(t, output, "hiddenCard")
	assert.NotContains(t, output, "??")
}

func TestBlackJackStandardPresentersStillHideDealerHoleCard(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	bj.SetPhase(domain.BJPhaseAction)

	var response struct {
		Dealer struct {
			Score int `json:"score"`
		} `json:"dealer"`
	}
	webOutput := (&BlackJackWebPresenter{}).Output(bj, nil)
	require.NoError(t, json.Unmarshal([]byte(webOutput), &response))
	assert.Zero(t, response.Dealer.Score)

	output := (&BlackJackCuiPresenter{}).Output(bj, nil)
	assert.Contains(t, output, "??")
}

func TestBlackJackDoubleExposureCuiPresenterShowsDealerScore(t *testing.T) {
	bj := setupDoubleExposurePresenterBlackJack(t)

	output := (&BlackJackCuiPresenter{}).Output(bj, nil)

	assert.Contains(t, output, "スコア 18")
}
