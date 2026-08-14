//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newRollingStoneForWeb(t *testing.T) *domain.RollingStone {
	t.Helper()
	r := domain.NewDefaultRollingStone()
	r.Reset()
	return r
}

func decodeRollingStone(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestRollingStoneWebPresenterOutput(t *testing.T) {
	p := new(RollingStoneWebPresenter)
	m := decodeRollingStone(t, p.Output(newRollingStoneForWeb(t), nil))

	assert.Equal(t, float64(domain.RollingStonePhasePlay), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["lastPickupIdx"])
	assert.Zero(t, m["discarded"])
	assert.False(t, m["mustPickUp"].(bool), "場が空なら引き取りは起きない")
	// **デッキ枚数は人数で変わる。** 4 人なら 32 枚。
	assert.Equal(t, float64(domain.RollingStoneDeckSize(domain.RollingStoneDefaultPlayerCnt)), m["deckSize"])

	players := m["players"].([]any)
	require.Len(t, players, domain.RollingStoneDefaultPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.RollingStoneHandSize), human["cardCount"], "1 人 8 枚")
	assert.Zero(t, human["pickups"])
	assert.Zero(t, human["finishedAt"])
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **出せる札が無いことを別のフラグで出す。**
//
// 空の `validPlays` は「まだ手番でない」とも読めてしまいます。
func TestRollingStoneWebPresenterFlagsWhenAPickUpIsForced(t *testing.T) {
	p := new(RollingStoneWebPresenter)
	r := newRollingStoneForWeb(t)
	r.SetLeadPlayerIdxForTest(1)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignHeart, 8, false))
	r.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})

	m := decodeRollingStone(t, p.Output(r, nil))
	assert.True(t, m["mustPickUp"].(bool))
	assert.Empty(t, m["validPlays"])
	assert.Equal(t, "rollingstone.pickup", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["n"])

	// **負のコントロール: フォローできるなら立たない。**
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false))
	m = decodeRollingStone(t, p.Output(r, nil))
	assert.False(t, m["mustPickUp"].(bool))
	assert.NotEmpty(t, m["validPlays"])
}

// **引き取り回数と上がり順位はワイヤに載る。** どちらも盤面に痕跡が残らない。
func TestRollingStoneWebPresenterCarriesPickupsAndFinishing(t *testing.T) {
	p := new(RollingStoneWebPresenter)
	r := newRollingStoneForWeb(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	// **席 0 に 2 枚持たせる。** 1 枚だと出した時点で上がって終局し、引き取りに進めない。
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false))
	r.GiveHandForTest(1, domain.NewCard(domain.CardDesignHeart, 8, false))
	require.NoError(t, r.PlayForTest(0, 0))
	require.NoError(t, r.PickUpForTest(1))

	m := decodeRollingStone(t, p.Output(r, nil))
	assert.Equal(t, float64(1), m["players"].([]any)[1].(map[string]any)["pickups"])
	assert.Equal(t, float64(1), m["lastPickupIdx"])

	// 上がりは別に確かめる（1 枚だけ持たせて出し切らせる）。
	fin := newRollingStoneForWeb(t)
	fin.SetLeadPlayerIdxForTest(0)
	fin.SetCurrentPlayerIdxForTest(0)
	fin.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false))
	require.NoError(t, fin.PlayForTest(0, 0))
	assert.Equal(t, float64(1),
		decodeRollingStone(t, p.Output(fin, nil))["players"].([]any)[0].(map[string]any)["finishedAt"])
}

func TestRollingStoneWebPresenterMessages(t *testing.T) {
	p := new(RollingStoneWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeRollingStone(t, p.Output(newRollingStoneForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("プレイ中は手札の枚数を出す", func(t *testing.T) {
		m := decodeRollingStone(t, p.Output(newRollingStoneForWeb(t), nil))
		assert.Equal(t, "rollingstone.play", m["messageCode"])
		assert.Equal(t, "8", m["messageParams"].(map[string]any)["n"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		r := newRollingStoneForWeb(t)
		r.GiveUp()
		m := decodeRollingStone(t, p.Output(r, nil))
		// 手札が残ったままなので「決着せず」扱い。
		assert.Contains(t, []any{"rollingstone.result.cpu", "rollingstone.result.stalemate"}, m["messageCode"])
	})

	// **上限で切った局は「上がった」わけではない。** 言い分ける。
	t.Run("上がりで決着した局", func(t *testing.T) {
		r := newRollingStoneForWeb(t)
		for i := range r.GetPlayerCnt() {
			r.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 7+i, false))
		}
		r.SetLeadPlayerIdxForTest(0)
		r.SetCurrentPlayerIdxForTest(0)
		for i := range r.GetPlayerCnt() {
			if r.GetGameEndFlag() {
				break
			}
			require.NoError(t, r.PlayForTest(i, 0))
		}
		require.True(t, r.GetGameEndFlag())
		m := decodeRollingStone(t, p.Output(r, nil))
		assert.Equal(t, "rollingstone.result.you", m["messageCode"])
	})
}

func TestRollingStoneWebPresenterHintOutput(t *testing.T) {
	p := new(RollingStoneWebPresenter)
	r := newRollingStoneForWeb(t)
	r.SetCurrentPlayerIdxForTest(0)

	hint := decodeRollingStone(t, p.HintOutput(r))["hint"].(map[string]any)
	assert.Equal(t, "rollingstoneLead", hint["reason"])
	assert.NotNil(t, hint["cardIndex"])

	// **引き取るしかない場面は札を指さない。**
	r.SetLeadPlayerIdxForTest(1)
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignHeart, 8, false))
	r.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})
	hint = decodeRollingStone(t, p.HintOutput(r))["hint"].(map[string]any)
	assert.Equal(t, "rollingstonePickUp", hint["reason"])
	assert.Nil(t, hint["cardIndex"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeRollingStone(t, p.Output(r, nil))["hint"])

	r.GiveUp()
	assert.Nil(t, decodeRollingStone(t, p.HintOutput(r))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestRollingStoneWebPresenterActionLogOutput(t *testing.T) {
	p := new(RollingStoneWebPresenter)
	r := newRollingStoneForWeb(t)
	assert.Empty(t, decodeRollingStone(t, p.ActionLogOutput(r))["entries"])

	r.GiveUp()
	assert.NotEmpty(t, decodeRollingStone(t, p.ActionLogOutput(r))["entries"])
}
