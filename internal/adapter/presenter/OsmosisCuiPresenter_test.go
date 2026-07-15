//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupOsmosisCuiMockDefaults(og *interfaces.MockOsmosisGame) {
	og.On("GetPhase").Return(domain.OsmosisPhasePlaying).Maybe()
	og.On("GetMoveCount").Return(0).Maybe()
	og.On("GetStockCount").Return(34).Maybe()
	og.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	og.On("GetBaseRank").Return(7).Maybe()

	var reserve [domain.OsmosisReserveCnt][]*domain.Card
	for i := 0; i < domain.OsmosisReserveCnt; i++ {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	og.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.OsmosisFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	og.On("GetFoundation").Return(foundation).Maybe()
}

func TestOsmosisCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "オズモシス")
		assert.Contains(t, result, "ベースランク: 7")
		assert.Contains(t, result, "組札0段:")
		assert.Contains(t, result, "リザーブ0列:")
		assert.Contains(t, result, "ストック: 34枚")
		assert.Contains(t, result, "ウェイスト: [空]")
		assert.Contains(t, result, "手数: 0")
		// Row 0 has cards -> any rank; row 1 (empty, row 0 seeded) -> base rank 7.
		assert.Contains(t, result, "任意ランク")
		assert.Contains(t, result, "[置ける: 7]")
	})

	t.Run("lower row lists ranks present in the row above", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetFoundation")
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		// Row 0 holds 7,8,9; row 1 holds only 7 -> row 1 may still take 8 and 9.
		f[0] = []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignSpade, 8, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		f[1] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)}
		og.On("GetFoundation").Return(f)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "[置ける: 8 9]")
	})

	t.Run("waste card", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetWaste")
		og.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ウェイスト: HEART 5")
	})

	t.Run("empty foundation row", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetFoundation")
		var empty [domain.OsmosisFoundationCnt][]*domain.Card
		og.On("GetFoundation").Return(empty)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty reserve column", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetReserve")
		var empty [domain.OsmosisReserveCnt][]*domain.Card
		og.On("GetReserve").Return(empty)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("error", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		setupOsmosisCuiMockDefaults(og)
		og.ExpectedCalls = filterCalls(og.ExpectedCalls, "GetPhase")
		og.On("GetPhase").Return(domain.OsmosisPhaseGameOver)
		p := new(OsmosisCuiPresenter)
		result := p.Output(og, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestOsmosisCuiPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return((*domain.OsmosisHint)(nil))
		p := new(OsmosisCuiPresenter)
		assert.Contains(t, p.HintOutput(og), "ヒントはありません")
	})

	t.Run("reserve to foundation shows allowed ranks", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "reserve", FromCol: 0, ToCol: 1})
		og.On("GetBaseRank").Return(7)
		var foundation [domain.OsmosisFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		og.On("GetFoundation").Return(foundation)
		p := new(OsmosisCuiPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, "リザーブ0列")
		assert.Contains(t, result, "組札1段")
		// Row 1 is empty with a seeded base row → only the base rank (7) is placeable.
		assert.Contains(t, result, "配置可能: 7")
	})

	t.Run("lower row with partial pile lists ranks from the row above", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "reserve", FromCol: 0, ToCol: 1})
		og.On("GetBaseRank").Return(7)
		var foundation [domain.OsmosisFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		foundation[1] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)}
		og.On("GetFoundation").Return(foundation)
		p := new(OsmosisCuiPresenter)
		result := p.HintOutput(og)
		// Row above has 7 and 9; row 1 already has 7 → only 9 is placeable.
		assert.Contains(t, result, "配置可能: 9")
	})

	t.Run("waste to base row shows any rank", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetHint").Return(&domain.OsmosisHint{FromZone: "waste", FromCol: -1, ToCol: 0})
		og.On("GetBaseRank").Return(7)
		var foundation [domain.OsmosisFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		og.On("GetFoundation").Return(foundation)
		p := new(OsmosisCuiPresenter)
		result := p.HintOutput(og)
		assert.Contains(t, result, "ウェイスト")
		// Base row already seeded → any rank may be added.
		assert.Contains(t, result, "任意ランク")
	})
}

func TestOsmosisAllowedRanks(t *testing.T) {
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }

	t.Run("empty row returns the base rank", func(t *testing.T) {
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{card(7)}
		assert.Equal(t, []int{7}, osmosisAllowedRanks(f, 7, 1))
	})

	t.Run("empty row with empty row above returns nothing", func(t *testing.T) {
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		assert.Nil(t, osmosisAllowedRanks(f, 7, 2))
	})

	t.Run("base row lists every rank not yet present", func(t *testing.T) {
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{card(7), card(13)}
		got := osmosisAllowedRanks(f, 7, 0)
		assert.NotContains(t, got, 7)
		assert.NotContains(t, got, 13)
		assert.Contains(t, got, 1)
		assert.Len(t, got, 11)
	})

	t.Run("lower row intersects the row above minus its own ranks", func(t *testing.T) {
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{card(7), card(9), card(11)}
		f[1] = []*domain.Card{card(7)}
		assert.Equal(t, []int{9, 11}, osmosisAllowedRanks(f, 7, 1))
	})
}

func TestOsmosisRankLabel(t *testing.T) {
	assert.Equal(t, "A", osmosisRankLabel(1))
	assert.Equal(t, "10", osmosisRankLabel(10))
	assert.Equal(t, "J", osmosisRankLabel(11))
	assert.Equal(t, "Q", osmosisRankLabel(12))
	assert.Equal(t, "K", osmosisRankLabel(13))
}

func TestOsmosisCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhasePlaying)
		p := new(OsmosisCuiPresenter)
		assert.Contains(t, p.ActionLogOutput(og), "棋譜はありません")
	})

	t.Run("after clear", func(t *testing.T) {
		og := new(interfaces.MockOsmosisGame)
		og.On("GetPhase").Return(domain.OsmosisPhaseGameClear)
		og.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "test"}})
		p := new(OsmosisCuiPresenter)
		result := p.ActionLogOutput(og)
		assert.Contains(t, result, "draw")
	})
}
