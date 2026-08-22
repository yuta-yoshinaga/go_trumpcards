package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewDramahaInteractor(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)
	assert.NotNil(t, hi)
}

func TestNewDramahaInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockDramahaPresenter)
	assert.Panics(t, func() {
		NewDramahaInteractor(nil, mp)
	})
}

func TestNewDramahaInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	assert.Panics(t, func() {
		NewDramahaInteractor(mg, nil)
	})
}

func TestDramahaInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := hi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestDramahaInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Reset()
	assert.Equal(t, "error output", result)
}

func TestDramahaInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DramahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestDramahaInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DramahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestDramahaInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mp.On("Output", mg, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")
	cfg := domain.DramahaConfig{SmallBlind: 0, BigBlind: 10, BlindLevelHands: 10}
	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
	mg.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestDramahaInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("PlayerAction", domain.DramahaActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := hi.Action(domain.DramahaActionCheck, 0, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.DramahaActionCheck, 0, 0)
}

func TestDramahaInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DefaultDramahaConfig()
	mg.On("GetConfig").Return(cfg)

	result := hi.GetConfig()
	assert.Equal(t, cfg, result)
	mg.AssertCalled(t, "GetConfig")
}

func TestDramahaInteractor_ResetWithConfig_TableSizeChange(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	// **6-max を頼んでも 4 席で組む。** ドラマハは 1 席が最悪 10 枚
	// (ホール 5 + 交換 5) 使い、ボードに 5 枚要るので 10N+5 枚必要 ——
	// 6-max は 65 枚で 52 枚の山に収まらない。NewDramahaPlayersForTable が
	// 4-max へ丸めるので、Resize に渡るのは 4 席。
	cfg := domain.DefaultDramahaConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mg.On("GetPlayerCnt").Return(2)
	mg.On("Resize", mock.MatchedBy(func(players []*domain.DramahaPlayer) bool {
		return len(players) == 4 && players[0].GetIsHuman()
	})).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("resize output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "resize output", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestDramahaInteractor_ResetWithConfig_SameTableSize(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DefaultDramahaConfig()
	cfg.TableSize = domain.HoldemTableSize4
	mg.On("GetPlayerCnt").Return(4)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("no resize output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "no resize output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestDramahaInteractor_ResetWithConfig_TableSizeZero(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DefaultDramahaConfig()
	cfg.TableSize = 0 // not set, should skip resize
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("zero output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "zero output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestDramahaInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	cfg := domain.DramahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	result := hi.ResetWithConfig(cfg, profileData)
	assert.Equal(t, "with profile output", result)
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestDramahaInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)

	hi := NewDramahaInteractor(mg, mp)
	result := hi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	mp.AssertExpectations(t)
}

func TestDramahaInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("test error")
	mg.On("PlayerAction", domain.DramahaActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Action(domain.DramahaActionBet, 50, 0)
	assert.Equal(t, "error output", result)
}

func TestDramahaInteractor_Rebuy(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("Rebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("rebuy output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy output", result)
	mg.AssertCalled(t, "Rebuy")
}

func TestDramahaInteractor_Rebuy_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("rebuy failed")
	mg.On("Rebuy").Return(err)
	mp.On("Output", mg, err).Return("rebuy error output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy error output", result)
}

func TestDramahaInteractor_SkipRebuy(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("SkipRebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip rebuy output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy output", result)
	mg.AssertCalled(t, "SkipRebuy")
}

func TestDramahaInteractor_SkipRebuy_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("skip rebuy failed")
	mg.On("SkipRebuy").Return(err)
	mp.On("Output", mg, err).Return("skip rebuy error output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy error output", result)
}

func TestDramahaInteractor_Addon(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("Addon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("addon output")

	result := hi.Addon()
	assert.Equal(t, "addon output", result)
	mg.AssertCalled(t, "Addon")
}

func TestDramahaInteractor_Addon_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("addon failed")
	mg.On("Addon").Return(err)
	mp.On("Output", mg, err).Return("addon error output")

	result := hi.Addon()
	assert.Equal(t, "addon error output", result)
}

func TestDramahaInteractor_SkipAddon(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("SkipAddon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip addon output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon output", result)
	mg.AssertCalled(t, "SkipAddon")
}

func TestDramahaInteractor_SkipAddon_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("skip addon failed")
	mg.On("SkipAddon").Return(err)
	mp.On("Output", mg, err).Return("skip addon error output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon error output", result)
}

func TestDramahaInteractor_Muck(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("Muck").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("muck output")

	result := hi.Muck()
	assert.Equal(t, "muck output", result)
	mg.AssertCalled(t, "Muck")
}

func TestDramahaInteractor_Muck_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("muck failed")
	mg.On("Muck").Return(err)
	mp.On("Output", mg, err).Return("muck error output")

	result := hi.Muck()
	assert.Equal(t, "muck error output", result)
}

func TestDramahaInteractor_ShowHand(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("ShowHand").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("show hand output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand output", result)
	mg.AssertCalled(t, "ShowHand")
}

func TestDramahaInteractor_ShowHand_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("show hand failed")
	mg.On("ShowHand").Return(err)
	mp.On("Output", mg, err).Return("show hand error output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand error output", result)
}

// ---------------------------------------------------------------------------
// Draw -- the round the clone (Omaha) has no equivalent for.
// ---------------------------------------------------------------------------

// TestDramahaInteractor_Draw_ForwardsIndicesVerbatim asserts the interactor
// passes the indices through untouched. Any renumbering belongs in the
// controllers (the CUI takes 1-based input); doing it twice would shift the
// discard by one.
func TestDramahaInteractor_Draw_ForwardsIndicesVerbatim(t *testing.T) {
	for _, tc := range []struct {
		name    string
		indices []int
	}{
		{"two cards", []int{0, 2}},
		{"the whole hand", []int{0, 1, 2, 3, 4}},
		{"stand pat with nil", nil},
		{"stand pat with an empty slice", []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mg := new(interfaces.MockDramahaGame)
			mp := new(presenter.MockDramahaPresenter)
			hi := NewDramahaInteractor(mg, mp)

			mg.On("Draw", 0, tc.indices).Return(nil)
			mp.On("Output", mg, mock.Anything).Return("draw output")

			assert.Equal(t, "draw output", hi.Draw(tc.indices))
			mg.AssertCalled(t, "Draw", 0, tc.indices)
		})
	}
}

// TestDramahaInteractor_Draw_AlwaysDrawsForTheHumanSeat pins the hard-coded
// seat. Letting the caller choose would let a Web request rebuild a CPU's hand.
func TestDramahaInteractor_Draw_AlwaysDrawsForTheHumanSeat(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	mg.On("Draw", 0, []int{1}).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("draw output")

	hi.Draw([]int{1})

	mg.AssertCalled(t, "Draw", 0, []int{1})
	mg.AssertNotCalled(t, "Draw", 1, mock.Anything)
	mg.AssertNotCalled(t, "Draw", 2, mock.Anything)
	mg.AssertNotCalled(t, "Draw", 3, mock.Anything)
}

func TestDramahaInteractor_Draw_Error(t *testing.T) {
	mg := new(interfaces.MockDramahaGame)
	mp := new(presenter.MockDramahaPresenter)
	hi := NewDramahaInteractor(mg, mp)

	err := errors.New("already drawn")
	mg.On("Draw", 0, []int{0}).Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", hi.Draw([]int{0}),
		"a rejected draw must be rendered, not swallowed")
	mp.AssertCalled(t, "Output", mg, err)
}
