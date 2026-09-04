//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCanfieldWebMockDefaults(cg *interfaces.MockCanfieldGame) {
	cg.On("GetPhase").Return(domain.CanfieldPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("GetStockCount").Return(34).Maybe()
	cg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	cg.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)}).Maybe()
	cg.On("GetBaseRank").Return(7).Maybe()
	cg.On("CanUndo").Return(false).Maybe()

	var tableau [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	for i := 0; i < domain.CanfieldTableauCnt; i++ {
		tableau[i] = []*domain.CanfieldTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, i+1, false)}}
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CanfieldFoundationCnt][]*domain.Card
	cg.On("GetFoundation").Return(foundation).Maybe()
}

// setupCanfieldOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupCanfieldOutputMock(g *interfaces.MockCanfieldGame) {
	setupCanfieldWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCanfieldWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, `"baseRank":7`)
		assert.Contains(t, result, `"stockCount":34`)
		assert.Contains(t, result, `"messageCode":"canfield.playing"`)
	})

	t.Run("with error", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameClear)
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "canfield.gameClear")
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameOver)
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "canfield.gameOver")
	})

	t.Run("with waste", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetWaste")
		cg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, `"waste"`)
	})

	t.Run("foundation with cards", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldOutputMock(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetFoundation")
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		cg.On("GetFoundation").Return(f)
		p := new(CanfieldWebPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, `"foundation"`)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCanfieldWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CanfieldHint{FromZone: "tableau", FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	cg := new(interfaces.MockCanfieldGame)
	setupCanfieldWebMockDefaults(cg)
	cg.On("GetHint").Return(hint).Maybe()

	result := new(CanfieldWebPresenter).Output(cg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

func TestCanfieldWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldWebMockDefaults(cg)
		cg.On("GetHint").Return(&domain.CanfieldHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: 0})
		p := new(CanfieldWebPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, `"canfield.hintAvailable"`)
	})

	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		setupCanfieldWebMockDefaults(cg)
		cg.On("GetHint").Return((*domain.CanfieldHint)(nil))
		p := new(CanfieldWebPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, `"canfield.noHint"`)
	})
}

func TestCanfieldWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetPhase").Return(domain.CanfieldPhasePlaying)
		cg.On("GetGameEndFlag").Return(false)
		p := new(CanfieldWebPresenter)
		_ = p.ActionLogOutput(cg)
	})

	t.Run("cleared", func(t *testing.T) {
		cg := new(interfaces.MockCanfieldGame)
		cg.On("GetPhase").Return(domain.CanfieldPhaseGameClear)
		cg.On("GetGameEndFlag").Return(true)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw"}})
		p := new(CanfieldWebPresenter)
		_ = p.ActionLogOutput(cg)
	})
}

func TestCanfieldWebPresenter_Errors(t *testing.T) {
	cg := new(interfaces.MockCanfieldGame)
	cg.On("GetBaseRank").Return(1)
	cg.On("GetFoundation").Return([domain.CanfieldFoundationCnt][]*domain.Card{})
	cg.On("GetReserve").Return([]*domain.Card{})
	cg.On("GetStockCount").Return(0)
	cg.On("GetWaste").Return([]*domain.Card{})
	cg.On("GetTableau").Return([domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard{})
	cg.On("GetPhase").Return(domain.CanfieldPhasePlaying)
	cg.On("GetMoveCount").Return(0)
	cg.On("CanUndo").Return(false)
	cg.On("GetHint").Return((*domain.CanfieldHint)(nil)) // For playing phase, it queries GetHint

	p := new(CanfieldWebPresenter)

	t.Run("with domain error code", func(t *testing.T) {
		err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "canfield.errEmptyColumnAutoFillOnly", nil)
		result := p.Output(cg, err)
		assert.Contains(t, result, `"messageCode":"canfield.errEmptyColumnAutoFillOnly"`)
		// キーを message にも入れてしまうと、それを表示する画面には
		// "canfield.err..." が生のまま出る (#5526)。code を出したのなら
		// message 側は空でなければならない。
		assert.NotContains(t, result, `"message":"canfield.`)
	})

	t.Run("with standard errors.New (negative control)", func(t *testing.T) {
		err := errors.New("cannot place card on tableau")
		result := p.Output(cg, err)
		assert.Contains(t, result, `"message":"cannot place card on tableau"`)
		assert.NotContains(t, result, `"messageCode":`)
	})
}
