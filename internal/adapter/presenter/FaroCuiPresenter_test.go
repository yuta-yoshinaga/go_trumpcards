package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupFaroCuiMockDefaults(m *interfaces.MockFaroGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FaroPhaseBetting).Maybe()
	m.On("GetTurnsPlayed").Return(0).Maybe()
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal).Maybe()
	m.On("GetRemainingCount").Return(51).Maybe()
	m.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	m.On("GetBetRanks").Return(([]int)(nil)).Maybe()
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil)).Maybe()
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil)).Maybe()
	m.On("GetCallCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCallOrder").Return(([]int)(nil)).Maybe()
	m.On("GetCallWon").Return(false).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestFaroCuiPresenter_Output_BettingPhase(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	setupFaroCuiMockDefaults(m)
	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "BETTING")
	assert.Contains(t, result, "ターン: 0/24")
}

func TestFaroCuiPresenter_Output_FaceRankLabels(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.FaroPhaseBetting)
	m.On("GetTurnsPlayed").Return(0)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(51)
	m.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	m.On("GetBetRanks").Return([]int{1, 11, 12, 13, 8})
	m.On("GetBets").Return(map[int]*domain.FaroBet{
		1: {Amount: 50}, 11: {Amount: 60}, 12: {Amount: 70}, 13: {Amount: 80}, 8: {Amount: 90},
	})
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(0)
	m.On("GetGameEndFlag").Return(false)

	result := p.Output(m, nil)
	assert.Contains(t, result, "ランク A: 50")
	assert.Contains(t, result, "ランク J: 60")
	assert.Contains(t, result, "ランク Q: 70")
	assert.Contains(t, result, "ランク K: 80")
	assert.Contains(t, result, "ランク 8: 90") // 2-10 stay numeric
	assert.NotContains(t, result, "ランク 1:")
	assert.NotContains(t, result, "ランク 13:")
}

func TestFaroCuiPresenter_Output_Error(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	setupFaroCuiMockDefaults(m)
	result := p.Output(m, errors.New("oops"))
	assert.Contains(t, result, "oops")
}

func TestFaroCuiPresenter_Output_WithBetsAndTurn(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.FaroPhaseTurn)
	m.On("GetTurnsPlayed").Return(1)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(49)
	m.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	m.On("GetBetRanks").Return([]int{7})
	m.On("GetBets").Return(map[int]*domain.FaroBet{7: {Amount: 100, Copper: true}})
	m.On("GetLastTurn").Return(&domain.FaroTurnResult{
		LosingCard:  domain.NewCard(domain.CardDesignSpade, 10, false),
		WinningCard: domain.NewCard(domain.CardDesignHeart, 5, false),
		Split:       false,
		Net:         100,
	})
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(100)
	m.On("GetGameEndFlag").Return(false)

	result := p.Output(m, nil)
	assert.Contains(t, result, "ランク 7: 100")
	assert.Contains(t, result, "カッパー")
	assert.Contains(t, result, "TURN")
}

func TestFaroCuiPresenter_Output_SplitAndCall(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(800)
	m.On("GetPhase").Return(domain.FaroPhaseCall)
	m.On("GetTurnsPlayed").Return(domain.FaroTurnsPerDeal)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(3)
	m.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	m.On("GetBetRanks").Return(([]int)(nil))
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	m.On("GetLastTurn").Return(&domain.FaroTurnResult{
		LosingCard:  domain.NewCard(domain.CardDesignSpade, 8, false),
		WinningCard: domain.NewCard(domain.CardDesignHeart, 8, false),
		Split:       true,
	})
	m.On("GetCallCards").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignClover, 12, false),
	})
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(-50)
	m.On("GetGameEndFlag").Return(false)

	result := p.Output(m, nil)
	assert.Contains(t, result, "スプリット")
	assert.Contains(t, result, "CALL")
}

func TestFaroCuiPresenter_Output_RoundEndWonAndLost(t *testing.T) {
	p := new(FaroCuiPresenter)

	won := new(interfaces.MockFaroGame)
	won.On("GetChips").Return(1300)
	won.On("GetPhase").Return(domain.FaroPhaseRoundEnd)
	won.On("GetTurnsPlayed").Return(domain.FaroTurnsPerDeal)
	won.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	won.On("GetRemainingCount").Return(0)
	won.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	won.On("GetBetRanks").Return(([]int)(nil))
	won.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	won.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	won.On("GetCallCards").Return(([]*domain.Card)(nil))
	won.On("GetCallOrder").Return([]int{3, 9, 12})
	won.On("GetCallWon").Return(true)
	won.On("GetTotalPayout").Return(400)
	won.On("GetGameEndFlag").Return(false)
	assert.Contains(t, p.Output(won, nil), "コール成功")

	lost := new(interfaces.MockFaroGame)
	lost.On("GetChips").Return(700)
	lost.On("GetPhase").Return(domain.FaroPhaseRoundEnd)
	lost.On("GetTurnsPlayed").Return(domain.FaroTurnsPerDeal)
	lost.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	lost.On("GetRemainingCount").Return(0)
	lost.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	lost.On("GetBetRanks").Return(([]int)(nil))
	lost.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	lost.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	lost.On("GetCallCards").Return(([]*domain.Card)(nil))
	lost.On("GetCallOrder").Return([]int{9, 3, 12})
	lost.On("GetCallWon").Return(false)
	lost.On("GetTotalPayout").Return(-100)
	lost.On("GetGameEndFlag").Return(false)
	assert.Contains(t, p.Output(lost, nil), "コール失敗")
}

func TestFaroCuiPresenter_Output_GameEnd(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	setupFaroCuiMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(0)
	m.On("GetPhase").Return(domain.FaroPhaseGameEnd)
	m.On("GetTurnsPlayed").Return(0)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(0)
	m.On("GetRemainingByRank").Return([domain.FaroMaxRank + 1]int{}).Maybe()
	m.On("GetBetRanks").Return(([]int)(nil))
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetCallOrder").Return(([]int)(nil))
	m.On("GetCallWon").Return(false)
	m.On("GetTotalPayout").Return(0)
	m.On("GetGameEndFlag").Return(true)
	result := p.Output(m, nil)
	assert.Contains(t, result, "ゲーム終了")
}

func TestFaroCuiPresenter_PhaseStr_AllBranches(t *testing.T) {
	p := new(FaroCuiPresenter)
	for phase, expect := range map[int]string{
		domain.FaroPhaseBetting:  "BETTING",
		domain.FaroPhaseTurn:     "TURN",
		domain.FaroPhaseCall:     "CALL",
		domain.FaroPhaseRoundEnd: "ROUND END",
		domain.FaroPhaseGameEnd:  "GAME END",
		999:                      "UNKNOWN",
	} {
		assert.Equal(t, expect, p.phaseStr(phase))
	}
}

func TestFaroCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(FaroCuiPresenter)
	m := new(interfaces.MockFaroGame)
	m.On("GetGameEndFlag").Return(false)
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

// **ケースキーパーはこのゲームの中核。**Web はランク別の残数を常時出すのに、
// CUI は総枚数しか出していなかった (#4894)。
func TestFaroCuiPresenter_ShowsTheCaseKeeper(t *testing.T) {
	// defaults は GetRemainingByRank を空で先に登録してしまう (testify は最初に
	// 一致した期待値を使う) ので、この検査だけは自前で組む。
	m := new(interfaces.MockFaroGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.FaroPhaseBetting)
	m.On("GetTurnsPlayed").Return(0)
	m.On("GetTurnsTotal").Return(domain.FaroTurnsPerDeal)
	m.On("GetRemainingCount").Return(51)
	m.On("GetBetRanks").Return(([]int)(nil))
	m.On("GetBets").Return((map[int]*domain.FaroBet)(nil))
	m.On("GetLastTurn").Return((*domain.FaroTurnResult)(nil))
	m.On("GetCallCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	var counts [domain.FaroMaxRank + 1]int
	for r := domain.FaroMinRank; r <= domain.FaroMaxRank; r++ {
		counts[r] = 4
	}
	counts[domain.FaroMinRank] = 0 // A は出尽くした
	counts[13] = 2                 // K は 2 枚残り
	m.On("GetRemainingByRank").Return(counts)

	out := new(FaroCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "ケースキーパー")
	// 出尽くしたランクも 0 として出す。**消すと「まだある」と読める。**
	assert.Contains(t, out, "A:0")
	assert.Contains(t, out, "K:2")
	assert.Contains(t, out, "2:4")
}
