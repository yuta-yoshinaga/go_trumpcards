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

func setupFaroWebMockDefaults(m *interfaces.MockFaroGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FaroPhaseBetting).Maybe()
	m.On("GetTurnsPlayed").Return(0).Maybe()
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal).Maybe()
	m.On("GetRemainingCount").Return(51).Maybe()
	m.On("GetBetRanks").Return(([]int)(nil)).Maybe()
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil)).Maybe()
	m.On("GetSoda").Return((*domain.Card)(nil)).Maybe()
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil)).Maybe()
	m.On("GetCallCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCallOrder").Return(([]int)(nil)).Maybe()
	m.On("GetCallWon").Return(false).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetRemainingByRank").Return(faroDeckCounts(4)).Maybe()
}

// faroDeckCounts はランク 1..13 をすべて n 枚にした残数表を返す。
// 添字 0 は未使用 (A..K が 1..13)。
func faroDeckCounts(n int) [domain.FaroMaxRank + 1]int {
	var out [domain.FaroMaxRank + 1]int
	for r := 1; r <= domain.FaroMaxRank; r++ {
		out[r] = n
	}
	return out
}

func parseFaroOutput(t *testing.T, jsonStr string) *controller.FaroWebOutput {
	t.Helper()
	var out controller.FaroWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestFaroWebPresenter_Output_BettingPhase(t *testing.T) {
	p := new(FaroWebPresenter)
	m := new(interfaces.MockFaroGame)
	setupFaroWebMockDefaults(m)
	r := parseFaroOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.FaroPhaseBetting, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Empty(t, r.Bets)
	assert.Equal(t, domain.FaroTurnsPerDeal, r.TurnsTotal)
	assert.Empty(t, r.Message)
}

func TestFaroWebPresenter_Output_Error(t *testing.T) {
	p := new(FaroWebPresenter)
	m := new(interfaces.MockFaroGame)
	setupFaroWebMockDefaults(m)
	r := parseFaroOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", r.Message)
}

func TestFaroWebPresenter_Output_TurnWithBets(t *testing.T) {
	p := new(FaroWebPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.FaroPhaseTurn)
	m.On("GetTurnsPlayed").Return(2)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(47)
	m.On("GetBetRanks").Return([]int{7})
	m.On("GetBets").Return(map[int]*domain.FaroBet{7: {Amount: 100, Copper: true}})
	m.On("GetSoda").Return(domain.NewCard(domain.CardDesignSpade, 1, false))
	m.On("GetLastTurn").Return(&domain.FaroTurnResult{
		LosingCard:  domain.NewCard(domain.CardDesignSpade, 10, false),
		WinningCard: domain.NewCard(domain.CardDesignHeart, 5, false),
		Split:       false,
	})
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(100)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetRemainingByRank").Return(faroDeckCounts(4))

	r := parseFaroOutput(t, p.Output(m, nil))
	assert.Len(t, r.Bets, 1)
	assert.Equal(t, 7, r.Bets[0].Rank)
	assert.True(t, r.Bets[0].Copper)
	assert.NotNil(t, r.Soda)
	assert.NotNil(t, r.LosingCard)
	assert.NotNil(t, r.WinningCard)
}

func TestFaroWebPresenter_Output_CallWonAndLost(t *testing.T) {
	p := new(FaroWebPresenter)

	won := new(interfaces.MockFaroGame)
	setupFaroWebMockDefaults(won)
	won.ExpectedCalls = nil
	won.On("GetChips").Return(1300)
	won.On("GetPhase").Return(domain.FaroPhaseRoundEnd)
	won.On("GetTurnsPlayed").Return(domain.FaroTurnsPerDeal)
	won.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	won.On("GetRemainingCount").Return(0)
	won.On("GetBetRanks").Return(([]int)(nil))
	won.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	won.On("GetSoda").Return((*domain.Card)(nil))
	won.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	won.On("GetCallCards").Return(([]*domain.Card)(nil))
	won.On("GetCallOrder").Return([]int{3, 9, 12})
	won.On("GetCallWon").Return(true)
	won.On("GetTotalPayout").Return(400)
	won.On("GetGameEndFlag").Return(false)
	won.On("GetRemainingByRank").Return(faroDeckCounts(0))
	r := parseFaroOutput(t, p.Output(won, nil))
	assert.Equal(t, "faro.result.callWon", r.MessageCode)
	assert.True(t, r.CallWon)

	lost := new(interfaces.MockFaroGame)
	lost.On("GetChips").Return(700)
	lost.On("GetPhase").Return(domain.FaroPhaseRoundEnd)
	lost.On("GetTurnsPlayed").Return(domain.FaroTurnsPerDeal)
	lost.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	lost.On("GetRemainingCount").Return(0)
	lost.On("GetBetRanks").Return(([]int)(nil))
	lost.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	lost.On("GetSoda").Return((*domain.Card)(nil))
	lost.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	lost.On("GetCallCards").Return(([]*domain.Card)(nil))
	lost.On("GetCallOrder").Return([]int{9, 3, 12})
	lost.On("GetCallWon").Return(false)
	lost.On("GetTotalPayout").Return(-100)
	lost.On("GetGameEndFlag").Return(false)
	lost.On("GetRemainingByRank").Return(faroDeckCounts(0))
	r2 := parseFaroOutput(t, p.Output(lost, nil))
	assert.Equal(t, "faro.result.callLost", r2.MessageCode)
}

func TestFaroWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(FaroWebPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(0)
	m.On("GetPhase").Return(domain.FaroPhaseGameEnd)
	m.On("GetTurnsPlayed").Return(0)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(0)
	m.On("GetBetRanks").Return(([]int)(nil))
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	m.On("GetSoda").Return((*domain.Card)(nil))
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(0)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetRemainingByRank").Return(faroDeckCounts(0))
	r := parseFaroOutput(t, p.Output(m, nil))
	assert.Equal(t, "faro.result.gameEnd", r.MessageCode)
}

func TestFaroWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(FaroWebPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetGameEndFlag").Return(false)
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

// **ケースキーパーはサーバが数える。**以前は Web の応答にランク別残数が無く、
// クライアントが公開札をローカルに蓄えて組み立てていた ── ラウンド途中で
// ページを再読み込みするとその蓄えが空に戻り、実際には配られた札を
// 「全ランク満数」と嘘をついていた (#6471)。CUI は最初からドメインの
// `GetRemainingByRank()` を読んでいる。
func TestFaroWebPresenter_ServesTheRemainingByRank(t *testing.T) {
	p := &FaroWebPresenter{}
	m := new(interfaces.MockFaroGame)
	// **既定より先に登録する。**testify は最初に一致した期待値を使うので、
	// `setupFaroWebMockDefaults` の満数を後から上書きすることはできない。
	counts := faroDeckCounts(4)
	counts[2] = 0 // 2 は出きっている
	counts[9] = 1
	m.On("GetRemainingByRank").Return(counts)
	setupFaroWebMockDefaults(m)

	out := parseFaroOutput(t, p.Output(m, nil))
	// 添字 0 は未使用。1..13 が A..K なので、長さもそのまま渡す。
	assert.Len(t, out.RemainingByRank, domain.FaroMaxRank+1)
	assert.Equal(t, 0, out.RemainingByRank[2])
	assert.Equal(t, 1, out.RemainingByRank[9])
	assert.Equal(t, 4, out.RemainingByRank[5])
}
