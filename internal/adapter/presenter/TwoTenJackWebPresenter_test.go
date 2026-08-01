package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTTJWebMock() (*interfaces.MockTwoTenJackGame, []*domain.TwoTenJackPlayer) {
	m := new(interfaces.MockTwoTenJackGame)
	players := makeTTJPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TwoTenJackPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultTwoTenJackConfig())
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m, players
}

func TestTwoTenJackWebPresenter_Output(t *testing.T) {
	p := new(presenter.TwoTenJackWebPresenter)

	t.Run("basic", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		result := p.Output(m, nil)
		assert.Contains(t, result, `"phase":1`)
		assert.Contains(t, result, `"trumpSuit":1`)
	})

	t.Run("error", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		result := p.Output(m, errors.New("bad"))
		assert.Contains(t, result, `"message":"bad"`)
	})

	t.Run("declare phase msg", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare)
		result := p.Output(m, nil)
		assert.Contains(t, result, "twotenjack.declarePhase")
	})

	t.Run("trick end msg", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "twotenjack.trickEnd")
	})

	t.Run("round end msg", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "twotenjack.roundEnd")
	})

	t.Run("play follow msg", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		}
		m.On("GetCurrentTrick").Return(trick)
		result := p.Output(m, nil)
		assert.Contains(t, result, "twotenjack.playPhase.follow")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestTwoTenJackWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TwoTenJackWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		idx := 0
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.TwoTenJackHint{CardIndex: &idx, Reason: "lead"})
		result := p.HintOutput(m)
		assert.Contains(t, result, `"hint"`)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTTJWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.TwoTenJackHint)(nil))
		result := p.HintOutput(m)
		assert.NotContains(t, result, `"hint":{`)
	})
}

func TestTwoTenJackWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TwoTenJackWebPresenter)
	m := new(interfaces.MockTwoTenJackGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。TwoTenJack.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestTwoTenJackWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ttj, _ := setupTTJWebMock()
	ttj.ExpectedCalls = removeMockCall(ttj.ExpectedCalls, "GetHint")
	ttj.On("GetHint").Return(&domain.TwoTenJackHint{CardIndex: &idx, Reason: "lead_trump"})

	result := new(presenter.TwoTenJackWebPresenter).Output(ttj, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
