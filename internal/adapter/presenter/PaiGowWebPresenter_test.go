package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPaiGowWebMockDefaults(m *interfaces.MockPaiGowGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseBet).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHint").Return((*domain.PaiGowHint)(nil)).Maybe()
	m.On("GetBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parsePaiGowOutput(t *testing.T, jsonStr string) *controller.PaiGowWebOutput {
	t.Helper()
	var out controller.PaiGowWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestPaiGowWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	setupPaiGowWebMockDefaults(m)

	result := parsePaiGowOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.PaiGowPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerCards)
	assert.Empty(t, result.Message)
}

func TestPaiGowWebPresenter_Output_Error(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	setupPaiGowWebMockDefaults(m)

	result := parsePaiGowOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

// #5526: ファウルのエラーは i18n キーを名乗るようになった。Error() をそのまま
// Message に入れると、画面に "paigow.foulHighMustBeat" が出る。
func TestPaiGowWebPresenter_Output_CodedErrorGoesOutAsACode(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	setupPaiGowWebMockDefaults(m)

	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "paigow.foulHighMustBeat", nil)
	result := parsePaiGowOutput(t, p.Output(m, err))
	assert.Equal(t, "paigow.foulHighMustBeat", result.MessageCode)
	assert.Empty(t, result.Message, "生のキーを message に入れない")
}

func TestPaiGowWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1190).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetHint").Return((*domain.PaiGowHint)(nil)).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(190).Maybe()
	m.On("GetCommission").Return(10).Maybe()
	m.On("GetPlayerHighRank").Return(1).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parsePaiGowOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "paigow.result.playerWins", result.MessageCode)
}

func TestPaiGowWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetHint").Return((*domain.PaiGowHint)(nil)).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(1).Maybe()
	m.On("GetDealerLowRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parsePaiGowOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "paigow.result.dealerWins", result.MessageCode)
}

func TestPaiGowWebPresenter_Output_Push(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetHint").Return((*domain.PaiGowHint)(nil)).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(100).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(1).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parsePaiGowOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "paigow.result.push", result.MessageCode)
}

func TestPaiGowWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(PaiGowWebPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}

func TestPaiGowWebPresenter_Hint(t *testing.T) {
	p := new(PaiGowWebPresenter)
	hint := &domain.PaiGowHint{LowIdx0: 2, LowIdx1: 3, LowIsPair: true, Reason: "house_way_pair"}

	// **受動ヒントは Output() でも埋める。**HintOutput() は command:"hint" 専用の
	// レスポンスで、ページの state にはマージされない。ここが nil のままだと
	// フロントの state.hint は常に undefined になる (#4483)。
	t.Run("Output carries the hint into the page state", func(t *testing.T) {
		m := new(interfaces.MockPaiGowGame)
		// **先に登録した期待が勝つ。**defaults の GetHint(nil) より前に置く。
		m.On("GetHint").Return(hint)
		setupPaiGowWebMockDefaults(m)

		out := parsePaiGowOutput(t, p.Output(m, nil))
		if assert.NotNil(t, out.Hint) {
			assert.Equal(t, 2, out.Hint.LowIdx0)
			assert.Equal(t, 3, out.Hint.LowIdx1)
			assert.True(t, out.Hint.LowIsPair)
			assert.Equal(t, "house_way_pair", out.Hint.Reason)
		}
	})

	t.Run("Output omits the hint outside the set hands phase", func(t *testing.T) {
		m := new(interfaces.MockPaiGowGame)
		setupPaiGowWebMockDefaults(m)

		assert.Nil(t, parsePaiGowOutput(t, p.Output(m, nil)).Hint)
	})

	t.Run("HintOutput returns the split", func(t *testing.T) {
		m := new(interfaces.MockPaiGowGame)
		// **先に登録した期待が勝つ。**defaults の GetHint(nil) より前に置く。
		m.On("GetHint").Return(hint)
		setupPaiGowWebMockDefaults(m)

		out := parsePaiGowOutput(t, p.HintOutput(m))
		if assert.NotNil(t, out.Hint) {
			assert.Equal(t, 2, out.Hint.LowIdx0)
		}
	})

	t.Run("HintOutput reports when there is nothing to suggest", func(t *testing.T) {
		m := new(interfaces.MockPaiGowGame)
		setupPaiGowWebMockDefaults(m)

		out := parsePaiGowOutput(t, p.HintOutput(m))
		assert.Nil(t, out.Hint)
		assert.Equal(t, "paigow.hintNone", out.MessageCode)
	})
}
