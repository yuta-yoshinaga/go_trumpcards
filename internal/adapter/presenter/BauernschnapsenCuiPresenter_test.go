package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBauernschnapsenCuiMock() *interfaces.MockBauernschnapsenGame {
	m := new(interfaces.MockBauernschnapsenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BauernschnapsenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetContract").Return(domain.BauernschnapsenContractRufer)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeBauernschnapsenPlayers() []*domain.BauernschnapsenPlayer {
	return []*domain.BauernschnapsenPlayer{
		domain.NewBauernschnapsenPlayer(true, 0),
		domain.NewBauernschnapsenPlayer(false, 1),
		domain.NewBauernschnapsenPlayer(false, 0),
		domain.NewBauernschnapsenPlayer(false, 1),
	}
}

func setupBauernschnapsenCuiMockWithPlayers() (*interfaces.MockBauernschnapsenGame, []*domain.BauernschnapsenPlayer) {
	m := setupBauernschnapsenCuiMock()
	players := makeBauernschnapsenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBauernschnapsenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BauernschnapsenCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBauernschnapsenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "バウエルンシュナプセン")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "切り札: SPADE")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	// 切り札が未確定のときにスート名を出すと、番人のいない -1 が
	// "UNKNOWN" として画面に出る。切り札なしの行に落ちること。
	t.Run("no trump named while the suit is undecided", func(t *testing.T) {
		m, _ := setupBauernschnapsenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(domain.BauernschnapsenNoTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.Tf("bauernschnapsen.contractLineNoTrump",
			"contract", i18n.T("bauernschnapsen.contractRufer"), "name", "CPU 1"))
		assert.NotContains(t, result, "UNKNOWN")
	})

	t.Run("marriage available prompt lists candidate cards", func(t *testing.T) {
		m, players := setupBauernschnapsenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageIndices")
		m.On("GetMarriageIndices", 0).Return([]int{0, 1})

		result := p.Output(m, nil)
		assert.Contains(t, result, "マリアージュ")
		// The human's K/Q candidate indices are enumerated.
		assert.Contains(t, result, strings.Split(i18n.T("bauernschnapsen.promptMarriageCards"), "{{")[0])
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "[1]")
	})

	t.Run("cpu turn does not leak marriage cards", func(t *testing.T) {
		m, players := setupBauernschnapsenCuiMockWithPlayers()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageIndices")
		m.On("GetCurrentPlayerIdx").Return(1) // a CPU
		m.On("GetMarriageIndices", 1).Return([]int{0, 1})

		result := p.Output(m, nil)
		// Generic hint may show, but the CPU's specific candidate cards must not.
		assert.NotContains(t, result, strings.Split(i18n.T("bauernschnapsen.promptMarriageCards"), "{{")[0])
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupBauernschnapsenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BauernschnapsenPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック完了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupBauernschnapsenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BauernschnapsenPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド完了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupBauernschnapsenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestBauernschnapsenCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BauernschnapsenCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupBauernschnapsenCuiMock()
		m.On("GetHint").Return((*domain.BauernschnapsenHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupBauernschnapsenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.BauernschnapsenHint{CardIndex: &idx, Reason: "follow_cut"})
		assert.Contains(t, p.HintOutput(m), "ヒント")
	})

	t.Run("marriage hint", func(t *testing.T) {
		m, players := setupBauernschnapsenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		idx := 0
		m.On("GetHint").Return(&domain.BauernschnapsenHint{CardIndex: &idx, Reason: "marriage", IsMarriage: true})
		assert.Contains(t, p.HintOutput(m), "マリアージュ")
	})
}

func TestBauernschnapsenCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupBauernschnapsenCuiMock()
	p := new(presenter.BauernschnapsenCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(m))
}

// 切り札は**表向きの札ではなく宣言**で決まる。クローン元のガイゲルにあった
// 「山札の下の表向き 1 枚」も山札そのものも、20 枚を配り切るこのゲームには無い。
// 代わりに契約と宣言者が読めることを見る。
func TestBauernschnapsenCuiPresenter_ShowsTheContract(t *testing.T) {
	p := new(presenter.BauernschnapsenCuiPresenter)

	build := func(phase domain.BauernschnapsenPhase, c domain.BauernschnapsenContract,
		declarer int) *interfaces.MockBauernschnapsenGame {
		m, _ := setupBauernschnapsenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(phase)
		m.On("GetContract").Return(c)
		m.On("GetDeclarerIdx").Return(declarer)
		return m
	}

	t.Run("names the contract and the trump", func(t *testing.T) {
		out := p.Output(build(domain.BauernschnapsenPhasePlay,
			domain.BauernschnapsenContractRufer, 1), nil)

		assert.Contains(t, out, i18n.T("bauernschnapsen.contractRufer"))
		// **反対の契約名は出ない。** 出ていたら契約を読まずに全部並べている。
		assert.NotContains(t, out, i18n.T("bauernschnapsen.contractBettel"))
		assert.NotContains(t, out, i18n.T("bauernschnapsen.contractFarbenzwang"))
	})

	// ベテルは切り札を取らないので、切り札スート名を出してはいけない。
	t.Run("omits the trump for Bettel", func(t *testing.T) {
		out := p.Output(build(domain.BauernschnapsenPhasePlay,
			domain.BauernschnapsenContractBettel, 2), nil)

		assert.Contains(t, out, i18n.T("bauernschnapsen.contractBettel"))
		assert.NotContains(t, out, i18n.T("bauernschnapsen.contractRufer"))
	})

	// 宣言中はまだ契約も切り札も無い。
	t.Run("says bidding while the contract phase is open", func(t *testing.T) {
		out := p.Output(build(domain.BauernschnapsenPhaseContract,
			domain.BauernschnapsenContractNone, -1), nil)

		assert.Contains(t, out, i18n.T("bauernschnapsen.contractPending"))
		assert.Contains(t, out, i18n.T("bauernschnapsen.promptContract"))
	})
}
