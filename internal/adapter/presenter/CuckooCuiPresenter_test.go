//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeCuckooPlayers() []*domain.CuckooPlayer {
	players := make([]*domain.CuckooPlayer, 4)
	players[0] = domain.NewCuckooPlayer(true)
	for i := 1; i < 4; i++ {
		players[i] = domain.NewCuckooPlayer(false)
	}
	for _, p := range players {
		p.SetLives(3)
	}
	return players
}

func setupCuckooCuiMock() (*interfaces.MockCuckooGame, []*domain.CuckooPlayer) {
	m := new(interfaces.MockCuckooGame)
	players := makeCuckooPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetStockCount").Return(40)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CuckooPhaseTurn)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetPendingSwapTo").Return(-1)
	m.On("GetSwapTargetIdx", mock.Anything).Return(1).Maybe()
	m.On("GetRoundLowest").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
		m.On("IsKingRevealed", i).Return(false)
	}
	return m, players
}

func TestCuckooCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CuckooCuiPresenter)

	t.Run("initial turn phase", func(t *testing.T) {
		m, players := setupCuckooCuiMock()
		players[0].SetCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "カッコー")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "ライフ 3")
		// Human card is shown, opponent card hidden.
		assert.Contains(t, result, "SPADE 5")
		assert.Contains(t, result, "非公開")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupCuckooCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})

	t.Run("refuse phase", func(t *testing.T) {
		m, _ := setupCuckooCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPendingSwapTo")
		m.On("GetPhase").Return(domain.CuckooPhaseRefuse)
		m.On("GetPendingSwapTo").Return(1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows lowest", func(t *testing.T) {
		m, _ := setupCuckooCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundLowest")
		m.On("GetPhase").Return(domain.CuckooPhaseRoundEnd)
		m.On("GetRoundLowest").Return(3)
		result := p.Output(m, nil)
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("eliminated player shown OUT", func(t *testing.T) {
		m, players := setupCuckooCuiMock()
		players[3].SetLives(0)
		assert.Contains(t, p.Output(m, nil), "脱落")
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupCuckooCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		assert.Contains(t, p.Output(m, nil), "勝ち")
	})
}

func TestCuckooCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CuckooCuiPresenter)

	m := new(interfaces.MockCuckooGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "keep", Detail: "You keep"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewCuckooPlayer(true)).Maybe()
	assert.Contains(t, p.ActionLogOutput(m), "keep")
}

// **交換の相手が誰か、CUI には一切出ていなかった** ── Web は #5671 でこれを
// 直したのに、CUI の promptTurnHelp は「交換する」としか言わない (#6467)。
// 脱落者は手番から飛ばされるので、席順の隣は答えにならない。
func TestCuckooCuiPresenter_NamesTheSwapTarget(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CuckooCuiPresenter)

	withTarget := func(target int) string {
		m, _ := setupCuckooCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSwapTargetIdx")
		m.On("GetSwapTargetIdx", mock.Anything).Return(target)
		return p.Output(m, nil)
	}

	t.Run("names the seat the swap would reach", func(t *testing.T) {
		// 現在の手番は 0、ディーラーは 3。相手は 2 (1 は飛ばされた席のつもり)。
		out := withTarget(2)
		assert.Contains(t, out, i18n.Tf("cuckoo.promptTurnSwapTarget",
			"name", i18n.Tf("cuiPlayerCpu", "idx", "2")))
		assert.NotContains(t, out, "{{")
	})

	// 相手がいない局面で名前を出すと、存在しない席を名指しすることになる。
	t.Run("says so when nobody is left to swap with", func(t *testing.T) {
		out := withTarget(-1)
		assert.Contains(t, out, i18n.T("cuckoo.promptTurnSwapNobody"))
		// 文言の**接尾**で見る。{{name}} を含む生テンプレートは決して現れないので、
		// キーそのものを NotContains に渡しても何も測らない。
		_, tail, ok := strings.Cut(i18n.T("cuckoo.promptTurnSwapTarget"), "}}")
		assert.True(t, ok)
		assert.NotContains(t, out, tail)
	})

	// ディーラーは山札と交換する。隣の名前を出したら嘘になる。
	t.Run("stays quiet for the dealer, who trades with the stock", func(t *testing.T) {
		m, _ := setupCuckooCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(3) // = GetDealerIdx
		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.T("cuckoo.promptTurnHelpDealer"))
		_, tail, ok := strings.Cut(i18n.T("cuckoo.promptTurnSwapTarget"), "}}")
		assert.True(t, ok)
		assert.NotContains(t, out, tail)
	})
}
