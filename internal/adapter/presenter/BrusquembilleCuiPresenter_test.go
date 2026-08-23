//go:build test && (!js || !wasm || classic)

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBrusquembilleCuiMock(trumpCard *domain.Card) *interfaces.MockBrusquembilleGame {
	m := new(interfaces.MockBrusquembilleGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BrusquembillePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetStockRemaining").Return(33)
	m.On("IsFollowRequired").Return(false).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1, 2}).Maybe()
	m.On("GetWinnerIdx").Return(-1)
	return m
}

func setupBrusquembilleCuiMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBrusquembilleGame, []*domain.BrusquembillePlayer) {
	m := setupBrusquembilleCuiMock(trumpCard)
	players := []*domain.BrusquembillePlayer{
		domain.NewBrusquembillePlayer(true),
		domain.NewBrusquembillePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(15)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestBrusquembilleCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.BrusquembilleCuiPresenter)

	t.Run("initial state shows header, trump, points", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupBrusquembilleCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, "Brusquembille")
		assert.Contains(t, out, "トリック: 1")
		assert.Contains(t, out, "山札: 33枚")
		// **全席ぶん並べる。** 3〜5 人卓では席 0/1 だけだと残りが消える。
		assert.Contains(t, out, "得点: 15-5")
		assert.Contains(t, out, "あなた:")
		assert.Contains(t, out, "CPU 1:")
		assert.Contains(t, out, "play <idx>")
	})

	t.Run("trump card exhausted", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
		out := p.Output(m, nil)
		assert.Contains(t, out, "使い切り")
	})

	t.Run("error is rendered", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, errors.New("kaboom"))
		assert.Contains(t, out, "kaboom")
	})

	t.Run("trick-end prompt", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BrusquembillePhaseTrickEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, "トリック終了")
	})

	t.Run("game end p0 banner", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(70)
		m.On("GetPlayerPoints", 1).Return(50)

		out := p.Output(m, nil)
		assert.Contains(t, out, "あなたの勝利")
		assert.Contains(t, out, "70-50")
	})

	t.Run("game end p1 banner", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(40)
		m.On("GetPlayerPoints", 1).Return(80)

		out := p.Output(m, nil)
		assert.Contains(t, out, "CPU 1 の勝利")
	})

	t.Run("game end tie banner", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(-1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(60)
		m.On("GetPlayerPoints", 1).Return(60)

		out := p.Output(m, nil)
		assert.Contains(t, out, "引き分け")
	})
}

func TestBrusquembilleCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BrusquembilleCuiPresenter)

	t.Run("hint shows card and reason", func(t *testing.T) {
		m, players := setupBrusquembilleCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.BrusquembilleHint{CardIndex: &idx, Reason: "follow_cut"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "トランプでカット")
	})

	t.Run("hint nil falls back to hintNone", func(t *testing.T) {
		m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
		m.On("GetHint").Return((*domain.BrusquembilleHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})

	t.Run("hint with unknown reason falls back to shared lookup", func(t *testing.T) {
		m, players := setupBrusquembilleCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.BrusquembilleHint{CardIndex: &idx, Reason: "unknown_reason"})
		out := p.HintOutput(m)
		assert.NotEmpty(t, out)
	})
}

func TestBrusquembilleCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BrusquembilleCuiPresenter)
	m, _ := setupBrusquembilleCuiMockWithPlayers(nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotNil(t, out)
}

// TestBrusquembilleCuiPresenter_MarksLegalCardsAfterTheStockRuns は、
// 後半に出せる札だけへ印が付くことを見る。
//
// **前半は印を付けない。** 山札がある間はどの札も合法なので、全部に印を
// 付けても情報が増えない。山札が尽きて追従義務が生まれた時点から、
// 出せる札を示す —— 素の一覧だと毎ターン手札とリードスートを暗算させる。
func TestBrusquembilleCuiPresenter_MarksLegalCardsAfterTheStockRuns(t *testing.T) {
	trumpCard := domain.NewCard(domain.CardDesignSpade, 1, false)
	render := func(followRequired bool, legal []int) string {
		m := new(interfaces.MockBrusquembilleGame)
		m.On("GetTrickNumber").Return(1)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.BrusquembillePhasePlay)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
		m.On("GetTrumpCard").Return(trumpCard)
		m.On("GetStockRemaining").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("IsFollowRequired").Return(followRequired)
		m.On("GetValidPlayIndices", 0).Return(legal)

		human := domain.NewBrusquembillePlayer(true)
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignClover, 13, false),
		} {
			human.AddCard(c)
		}
		cpu := domain.NewBrusquembillePlayer(false)
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(human)
		m.On("GetPlayer", 1).Return(cpu)
		m.On("GetPlayerPoints", 0).Return(0)
		m.On("GetPlayerPoints", 1).Return(0)

		p := new(presenter.BrusquembilleCuiPresenter)
		return p.Output(m, nil)
	}

	// 前半: 全部合法 → 印は付かない。
	early := render(false, []int{0, 1, 2})
	assert.NotContains(t, early, presenter.CuiLegalMark,
		"前半は全部出せるので印は不要:\n%s", early)

	// 後半: 追従できる札だけ → その 1 枚に印。
	// **凡例の行にも `*` が入る**ので、手札の行だけを数える。
	late := render(true, []int{1})
	handLine := ""
	for _, line := range strings.Split(late, "\n") {
		if strings.Contains(line, "[0]") {
			handLine = line
			break
		}
	}
	assert.NotEmpty(t, handLine, "手札の行が見つからない:\n%s", late)
	assert.Equal(t, 1, strings.Count(handLine, presenter.CuiLegalMark),
		"追従できる 1 枚だけに印が付く: %q", handLine)
	assert.Contains(t, late, i18n.T("brusquembille.followLegend"),
		"印の意味を説明する:\n%s", late)
}

// TestBrusquembilleCuiPresenter_ReportsEverySeat は、得点行と結果バナーが
// **卓の全席**を出すことを見る。
//
// **席 0 と席 1 だけ読むと、3〜5 人卓では残りの席が結果画面から消える** ——
// しかも勝者がその席のときに。ドメインの勝者判定を全席にしても、表示側が
// 二人ぶんのままだと画面の中で矛盾する。
func TestBrusquembilleCuiPresenter_ReportsEverySeat(t *testing.T) {
	m := new(interfaces.MockBrusquembilleGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetPhase").Return(domain.BrusquembillePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return((*domain.Card)(nil))
	m.On("GetStockRemaining").Return(0)
	m.On("IsFollowRequired").Return(false).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{}).Maybe()
	m.On("GetGameEndFlag").Return(true)
	// 席 3 (CPU 3) が最多で勝つ 4 人卓。
	m.On("GetWinnerIdx").Return(3)
	m.On("GetPlayerCnt").Return(4)
	pts := []int{10, 20, 30, 60}
	for i, v := range pts {
		m.On("GetPlayer", i).Return(domain.NewBrusquembillePlayer(i == 0))
		m.On("GetPlayerPoints", i).Return(v)
	}

	out := new(presenter.BrusquembilleCuiPresenter).Output(m, nil)

	// 4 席ぶんの得点が全部出る。
	assert.Contains(t, out, "10-20-30-60",
		"全席の得点が出る (席 0/1 だけだと 30 と 60 が消える):\n%s", out)
	// 勝者は席 3。「CPU の勝利」ではどの CPU か分からない。
	assert.Contains(t, out, "CPU 3 の勝利", "勝った席を名指しする:\n%s", out)
}
