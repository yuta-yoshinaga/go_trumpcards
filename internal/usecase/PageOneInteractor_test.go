//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newPageOneMocks() (*interfaces.MockPageOneGame, *presenter.MockPageOnePresenter) {
	return new(interfaces.MockPageOneGame), new(presenter.MockPageOnePresenter)
}

// humanIdleMocks sets up mock expectations so that runCpuTurns exits on the human's turn.
func humanIdleMocks(gm *interfaces.MockPageOneGame) {
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)
}

func TestNewPageOneInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockPageOnePresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PageOneInteractor: g must not be nil", func() {
			usecase.NewPageOneInteractor(nil, pMock)
		})
	})
	t.Run("panics when gp is nil", func(t *testing.T) {
		gm := new(interfaces.MockPageOneGame)
		assert.PanicsWithValue(t, "PageOneInteractor: gp must not be nil", func() {
			usecase.NewPageOneInteractor(gm, nil)
		})
	})
}

func TestPageOneInteractor_Reset(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("Reset").Return()
	humanIdleMocks(gm)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.Reset())
	gm.AssertCalled(t, "Reset")
}

func TestPageOneInteractor_ResetWithConfig_Valid(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	cfg := domain.PageOneConfig{CpuDifficulty: domain.PageOneCpuDifficultyHard, PointLimit: 300}
	gm.On("SetConfig", cfg).Return()
	gm.On("Reset").Return()
	humanIdleMocks(gm)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.ResetWithConfig(cfg))
	gm.AssertCalled(t, "SetConfig", cfg)
}

func TestPageOneInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", gm, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
	ci := usecase.NewPageOneInteractor(gm, pm)
	bad := domain.PageOneConfig{CpuDifficulty: domain.PageOneCpuDifficulty(99), PointLimit: 100}
	assert.Equal(t, "err", ci.ResetWithConfig(bad))
	gm.AssertNotCalled(t, "Reset")
}

func TestPageOneInteractor_Play(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0).Return(nil)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.Play(0))
	gm.AssertCalled(t, "PlayerPlay", 0)
}

func TestPageOneInteractor_Play_Error(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", gm, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0).Return(errors.New("boom"))

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "err", ci.Play(0))
}

func TestPageOneInteractor_Draw(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerDraw").Return(nil)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.Draw())
}

func TestPageOneInteractor_Declare(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerDeclare").Return(nil)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.Declare())
	gm.AssertCalled(t, "PlayerDeclare")
}

func TestPageOneInteractor_Declare_Error(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", gm, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerDeclare").Return(errors.New("boom"))

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "err", ci.Declare())
}

func TestPageOneInteractor_SkipDeclare(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerSkipDeclare").Return(nil)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.SkipDeclare())
}

func TestPageOneInteractor_NextRound(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("ScoreRound").Return()
	gm.On("GetGameEndFlag").Return(false)
	gm.On("NextRound").Return()
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(true)

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "ok", ci.NextRound())
	gm.AssertCalled(t, "NextRound")
}

func TestPageOneInteractor_GetConfigAndLog(t *testing.T) {
	gm, pm := newPageOneMocks()
	cfg := domain.DefaultPageOneConfig()
	gm.On("GetConfig").Return(cfg)
	pm.On("ActionLogOutput", mock.Anything).Return("log")

	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestPageOneInteractor_runCpuTurns_Declare(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("Reset").Return()

	// First: gameEnd=false, phase=MustDeclare, IsHumanTurn=false → CpuDeclare
	// Then: gameEnd=false, phase=Play, IsHumanTurn=true → break
	gameEnd := []bool{false, false}
	phases := []domain.PageOnePhase{domain.PageOnePhaseMustDeclare, domain.PageOnePhasePlay}
	humans := []bool{false, true}
	gi, pi, hi := 0, 0, 0
	gm.On("GetGameEndFlag").Return(func() bool {
		v := gameEnd[gi]
		gi++
		return v
	}())
	_ = phases
	_ = humans
	_ = pi
	_ = hi

	// Simpler: use sequence via mock.Called + once. Actually testify allows sequenced return values through Once().
	gm2, pm2 := newPageOneMocks()
	pm2.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm2.On("Reset").Return()
	gm2.On("GetGameEndFlag").Return(false)
	gm2.On("GetPhase").Return(domain.PageOnePhaseMustDeclare).Once()
	gm2.On("IsHumanTurn").Return(false).Once()
	gm2.On("CpuDeclare").Return().Once()
	gm2.On("GetPhase").Return(domain.PageOnePhasePlay).Once()
	gm2.On("IsHumanTurn").Return(true).Once()

	ci := usecase.NewPageOneInteractor(gm2, pm2)
	_ = ci.Reset()
	gm2.AssertCalled(t, "CpuDeclare")
}

func TestPageOneInteractor_runCpuTurns_CpuPlay(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("Reset").Return()
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhasePlay)
	gm.On("IsHumanTurn").Return(false).Once()
	gm.On("CpuPlay").Return().Once()
	gm.On("IsHumanTurn").Return(true).Once()

	ci := usecase.NewPageOneInteractor(gm, pm)
	_ = ci.Reset()
	gm.AssertCalled(t, "CpuPlay")
}

func TestPageOneInteractor_runCpuTurns_StopsOnRoundEnd(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("Output", mock.Anything, mock.Anything).Return("ok")
	gm.On("Reset").Return()
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetPhase").Return(domain.PageOnePhaseRoundEnd)

	ci := usecase.NewPageOneInteractor(gm, pm)
	_ = ci.Reset()
}

func TestRestorePageOneInteractor(t *testing.T) {
	data := `{"pl":[],"cf":{"cd":1,"pl":200},"ps":0,"ci":0,"dp":[],"wp":[],"ge":false,"wi":-1,"rn":1,"al":[]}`
	_, pm := newPageOneMocks()
	ci, err := usecase.RestorePageOneInteractor([]byte(data), pm)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestPageOneInteractor_Hint(t *testing.T) {
	gm, pm := newPageOneMocks()
	pm.On("HintOutput", mock.Anything).Return("hint")
	ci := usecase.NewPageOneInteractor(gm, pm)
	assert.Equal(t, "hint", ci.Hint())
}
