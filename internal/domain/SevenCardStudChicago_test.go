//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// chicagoCard is a shorthand constructor for a face-up card.
func chicagoCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// setChicagoHole replaces a seat's face-down cards.
func setChicagoHole(p *domain.SevenCardStudPlayer, cards ...*domain.Card) {
	p.ClearCards()
	for _, c := range cards {
		p.AddHoleCard(c)
	}
}

func TestChicago_ModeFlagsAreDistinct(t *testing.T) {
	g := domain.NewDefaultSevenCardStudChicago()
	assert.True(t, g.GetIsChicago())
	// **Hi-Lo とは別のモード。** 同じフラグを立て回すと、8 以下のローでも
	// 半分が動いてしまう。
	assert.False(t, g.GetIsHiLo())
	assert.False(t, g.GetIsLowball())

	assert.False(t, domain.NewDefaultSevenCardStud().GetIsChicago())
	assert.False(t, domain.NewDefaultSevenCardStudHiLo().GetIsChicago())
	assert.True(t, domain.NewDefaultSevenCardStudHiLo().GetIsHiLo())
}

// **エースは 1 だがスペードでは最高。** 額面で比べると A♠ が 2♠ に負ける。
func TestChicago_AceOfSpadesOutranksEveryOtherSpade(t *testing.T) {
	p := domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG)
	setChicagoHole(p,
		chicagoCard(domain.CardDesignSpade, 13),
		chicagoCard(domain.CardDesignSpade, 1),
		chicagoCard(domain.CardDesignSpade, 12))
	require.True(t, p.EvalChicagoSpade())
	require.NotNil(t, p.GetChicagoSpade())
	assert.Equal(t, 1, p.GetChicagoSpade().GetValue(), "A♠ が最高")
}

func TestChicago_PicksTheHighestSpadeAndIgnoresOtherSuits(t *testing.T) {
	p := domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG)
	setChicagoHole(p,
		chicagoCard(domain.CardDesignHeart, 1),
		chicagoCard(domain.CardDesignSpade, 9),
		chicagoCard(domain.CardDesignDiamond, 13),
		chicagoCard(domain.CardDesignSpade, 4))
	require.True(t, p.EvalChicagoSpade())
	assert.Equal(t, 9, p.GetChicagoSpade().GetValue())
	assert.Equal(t, domain.CardDesignSpade, p.GetChicagoSpade().GetDesign())
}

func TestChicago_NoSpadeInTheHoleDoesNotQualify(t *testing.T) {
	p := domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG)
	setChicagoHole(p,
		chicagoCard(domain.CardDesignHeart, 1),
		chicagoCard(domain.CardDesignClover, 13),
		chicagoCard(domain.CardDesignDiamond, 12))
	assert.False(t, p.EvalChicagoSpade())
	assert.Nil(t, p.GetChicagoSpade())
}

// **表札のスペードは数えない。** 卓に見えている札を数えると、伏せ札の読み合いが
// そのまま消える。
func TestChicago_FaceUpSpadesDoNotCount(t *testing.T) {
	p := domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG)
	p.ClearCards()
	p.AddHoleCard(chicagoCard(domain.CardDesignHeart, 5))
	p.AddDoorCard(chicagoCard(domain.CardDesignSpade, 1))
	assert.False(t, p.EvalChicagoSpade(), "表札の A♠ は資格にならない")
	assert.Nil(t, p.GetChicagoSpade())
}

// **評価は配り直しで消える。** 消さないと、スペードを 1 枚も貰わなかった席が
// 前のディールの札で半分を取り続ける。
func TestChicago_SpadeIsClearedBetweenDeals(t *testing.T) {
	p := domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG)
	setChicagoHole(p, chicagoCard(domain.CardDesignSpade, 1))
	require.True(t, p.EvalChicagoSpade())
	p.ClearCards()
	assert.Nil(t, p.GetChicagoSpade())
}

func TestChicago_JSONRoundTripKeepsTheMode(t *testing.T) {
	g := domain.NewDefaultSevenCardStudChicago()
	require.NoError(t, g.Reset())
	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := new(domain.SevenCardStud)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.True(t, restored.GetIsChicago(), "モードを落とすと復元後はただのスタッドになる")
	assert.False(t, restored.GetIsHiLo())
}
