//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBiribaWebMock() *interfaces.MockBiribaGame {
	m := new(interfaces.MockBiribaGame)
	m.On("GetRoundNumber").Return(1)
	// ヒントは既定で「なし」。値そのものを見るテストは本物のドメインで確かめる。
	m.On("GetHint").Return((*domain.BiribaHint)(nil)).Maybe()
	m.On("GetDrawPileCount").Return(54)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetPozzettoCount").Return(2)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetDiscardPile").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BiribaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultBiribaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeBiribaPlayers() []*domain.BiribaPlayer {
	return []*domain.BiribaPlayer{
		domain.NewBiribaPlayer(true),
		domain.NewBiribaPlayer(false),
	}
}

func setupBiribaWebMockWithPlayers() (*interfaces.MockBiribaGame, []*domain.BiribaPlayer) {
	m := setupBiribaWebMock()
	players := makeBiribaPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestBiribaWebPresenter_Output(t *testing.T) {
	p := new(presenter.BiribaWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.BiribaWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 54, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Nil(t, resObj.DiscardTop)
		assert.False(t, resObj.IsFrozen)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden in draw phase", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0)
	})

	t.Run("CPU cards shown in round end phase", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BiribaPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("CPU cards shown in game end phase", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.BiribaPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("discard top populated", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, "HEART", resObj.DiscardTop.Design)
		assert.Equal(t, 7, resObj.DiscardTop.Value)
	})

	t.Run("discard pile populated in oldest-first order", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardPile")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardPileCount")
		pile := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		m.On("GetDiscardPile").Return(pile)
		m.On("GetDiscardPileCount").Return(len(pile))

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.DiscardPile, 3)
		assert.Equal(t, 3, resObj.DiscardPileCount)
		assert.Equal(t, "SPADE", resObj.DiscardPile[0].Design)
		assert.Equal(t, 3, resObj.DiscardPile[0].Value)
		assert.Equal(t, "HEART", resObj.DiscardPile[1].Design)
		assert.Equal(t, "CLOVER", resObj.DiscardPile[2].Design)
		assert.Equal(t, 10, resObj.DiscardPile[2].Value)
	})

	t.Run("empty discard pile yields empty array", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.DiscardPile)
		assert.Len(t, resObj.DiscardPile, 0)
	})

	t.Run("frozen pile flag", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.IsFrozen)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.BiribaConfig{
			CpuDifficulty: domain.BiribaCpuDifficultyHard,
			PointLimit:    7500,
		})

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.BiribaCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 7500, resObj.Config.PointLimit)
	})

	t.Run("player scores", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		players[1].SetCumulativeScore(300)
		players[1].SetRoundScore(100)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 300, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 100, resObj.Players[1].RoundScore)
	})

	t.Run("player with meld", func(t *testing.T) {
		m, players := setupBiribaWebMockWithPlayers()
		meld := &domain.BiribaMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		players[0].SetMelds([]*domain.BiribaMeld{meld})

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Players[0].Melds, 1)
		assert.Len(t, resObj.Players[0].Melds[0].Cards, 3)
		assert.True(t, resObj.Players[0].Melds[0].IsNatural)
		assert.Equal(t, 7, resObj.Players[0].Melds[0].Rank)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetPhase").Return(domain.BiribaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "biriba.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		m.On("GetPhase").Return(domain.BiribaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 1")
		assert.Equal(t, "biriba.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "1"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupBiribaWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.BiribaPlayer)(nil))
		m.On("GetPhase").Return(domain.BiribaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "biriba.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("draw phase messageCode", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "biriba.drawPhase", resObj.MessageCode)
	})

	t.Run("meld phase messageCode", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BiribaPhaseMeld)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "biriba.meldPhase", resObj.MessageCode)
	})

	t.Run("discard phase messageCode", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BiribaPhaseDiscard)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "biriba.discardPhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BiribaPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "biriba.roundEnd", resObj.MessageCode)
	})

	t.Run("game end phase no messageCode", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BiribaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupBiribaWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BiribaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.BiribaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.BiribaDefaultPointLimit, resObj.Config.PointLimit)
	})
}

func TestBiribaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BiribaWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockBiribaGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew from stock", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"draw_stock"`)
		assert.Contains(t, result, `"detail":"drew from stock"`)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockBiribaGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBiribaGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

// #5628: CUI は GetHint() を使って「どちらの山から引くか」「どのカードで
// メルドできるか」を**インデックス付きの理由込み**で返すのに、Web は
// フロントの大まかな推定 (フェーズ別のアクション名だけ) を使っていた。
func TestBiribaWebPresenterCarriesTheDomainHint(t *testing.T) {
	g := domain.NewDefaultBiriba()
	g.Reset()

	var out controller.BiribaWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.BiribaWebPresenter).HintOutput(g)), &out))

	want := g.GetHint()
	require.NotNil(t, want, "配り直後は人間の引きフェーズなのでヒントが出る")
	require.NotNil(t, out.Hint)
	assert.Equal(t, want.Action, out.Hint.Action)
	assert.Equal(t, want.Indices, out.Hint.Indices)

	// **理由は「そのまま」ではなくキーに直して運ぶ。**ドメインが返すのは
	// `draw_discard_pair` のような内部識別子で、素通しするとフロントは存在
	// しないキーを引き、翻訳の代わりに識別子が画面に出る。
	assert.NotEqual(t, want.Reason, out.Hint.Reason, "内部識別子を素通ししていない")
	// CUI が同じ盤面で出す文言と一致すること = 2 つの画面が同じ説明をする。
	assert.Equal(t, i18n.T("biriba."+out.Hint.Reason), i18n.T(biribaCuiReasonKeyForTest(want.Reason)))
	assert.NotEqual(t, "biriba."+out.Hint.Reason, i18n.T("biriba."+out.Hint.Reason),
		"翻訳が存在する (キーがそのまま返っていない)")
}

// CPU の手番などヒントが無い場面ではフィールドごと出さない。
// 空のオブジェクトだと「行動できない」と読める。
func TestBiribaWebPresenterOmitsTheHintWhenThereIsNone(t *testing.T) {
	g := domain.NewDefaultBiriba()
	g.Reset()
	g.SetCurrentPlayerIdx(1)

	var out controller.BiribaWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.BiribaWebPresenter).HintOutput(g)), &out))
	assert.Nil(t, out.Hint)
}

// biribaCuiReasonKeyForTest returns the key the CUI presenter uses for a
// domain reason, so the web output can be compared against the same sentence.
func biribaCuiReasonKeyForTest(reason string) string {
	return map[string]string{
		"draw_discard_pair": "biriba.hintReasonDrawDiscard",
		"draw_stock_safe":   "biriba.hintReasonDrawStock",
		"meld_available":    "biriba.hintReasonMeld",
		"no_meld":           "biriba.hintReasonNoMeld",
		"discard_safe":      "biriba.hintReasonDiscard",
	}[reason]
}

// **ドメインが返しうる理由をすべて変換できること。**表に載っていない値は
// 空文字になり、フロントは `hint.` だけのキーを引いて何も出せなくなる。
func TestBiribaWebPresenterTranslatesEveryHintReason(t *testing.T) {
	for _, reason := range []string{"draw_discard_pair", "draw_stock_safe", "meld_available", "no_meld", "discard_safe"} {
		key := presenter.BiribaWebHintReasonKeyForTest(reason)
		assert.NotEmpty(t, key, "reason %q has no web key", reason)
		// ロケールに実在すること (i18n.T はキーが無いとキー自身を返す)。
		assert.NotEqual(t, "biriba."+key, i18n.T("biriba."+key), "biriba.%s is missing from the locale", key)
	}
}
