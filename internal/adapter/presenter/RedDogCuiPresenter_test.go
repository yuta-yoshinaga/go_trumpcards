package presenter

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupRedDogCuiMockDefaults(m *interfaces.MockRedDogGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.RedDogPhaseBet).Maybe()
	m.On("GetInitialCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetThirdCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnte").Return(0).Maybe()
	m.On("GetRaise").Return(0).Maybe()
	m.On("GetSpread").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestRedDogCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "BET")
}

func TestRedDogCuiPresenter_Output_Error(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogCuiMockDefaults(m)
	result := p.Output(m, errors.New("oops"))
	assert.Contains(t, result, "oops")
}

func TestRedDogCuiPresenter_Output_SpreadDecision(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseSpreadDecision)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "SPREAD DECISION")
	assert.Contains(t, result, "INITIAL")
	assert.Contains(t, result, "スプレッド: 4")
	assert.Contains(t, result, "アンテ: 100")
	// Winning ranks strictly between 5 and 10 are 6,7,8,9, plus the raise/stay guide.
	assert.Contains(t, result, "勝てるランク: 6, 7, 8, 9")
	assert.Contains(t, result, "raise")
}

func TestRedDogCuiPresenter_Output_SpreadGuideFaceCards(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 1, false), // Ace high (14)
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseSpreadDecision)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(3)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	// Ranks strictly between 10 and Ace(14): J,Q,K.
	assert.Contains(t, result, "勝てるランク: J, Q, K")
}

func TestRedDogCuiPresenter_Output_EndWin(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	third := domain.NewCard(domain.CardDesignClover, 7, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return(third)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(100)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(400)

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "THIRD")
	assert.Contains(t, result, "レイズ: 100")
	assert.Contains(t, result, "合計払戻し: 400")
	assert.Contains(t, result, "当たりランク: 6, 7, 8, 9 (引いたカード: 7 で的中)")
}

func TestRedDogCuiPresenter_Output_EndLose(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	third := domain.NewCard(domain.CardDesignClover, 12, false)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return(third)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーの負け")
	assert.Contains(t, result, "当たりランク: 6, 7, 8, 9 (外れ)")
	assert.NotContains(t, result, "で的中)")
}

func TestRedDogCuiPresenter_Output_EndPush(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(([]*domain.Card)(nil))
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetTotalPayout").Return(100)
	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
}

func TestRedDogCuiPresenter_Output_End_SpreadZero(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	third := domain.NewCard(domain.CardDesignClover, 12, false)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return(third)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetTotalPayout").Return(100)

	result := p.Output(m, nil)
	assert.NotContains(t, result, "当たりランク:")
}

func TestRedDogCuiPresenter_PhaseStr_AllBranches(t *testing.T) {
	p := new(RedDogCuiPresenter)
	for phase, expect := range map[int]string{
		domain.RedDogPhaseBet:            "BET",
		domain.RedDogPhaseInitialDealt:   "INITIAL DEALT",
		domain.RedDogPhaseSpreadDecision: "SPREAD DECISION",
		domain.RedDogPhaseEnd:            "END",
		999:                              "UNKNOWN",
	} {
		assert.Equal(t, expect, p.phaseStr(phase))
	}
}

func TestRedDogWinningRanksStr(t *testing.T) {
	// Higher card first exercises the lo/hi swap.
	reversed := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}
	assert.Equal(t, "6, 7, 8, 9", redDogWinningRanksStr(reversed))
	// A non-pair of two cards always yields a non-empty list at decision time,
	// but a malformed (non-2) slice is guarded.
	assert.Equal(t, "", redDogWinningRanksStr(nil))
	assert.Equal(t, "", redDogWinningRanksStr([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}))
}

func TestRedDogCuiPresenter_HintOutput_RaiseRecommended(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 11, false), // spread 3..11 -> 7 winning ranks
	}
	m.On("GetPhase").Return(domain.RedDogPhaseSpreadDecision)
	m.On("GetSpread").Return(7)
	m.On("GetInitialCards").Return(cards)

	result := p.HintOutput(m)
	assert.Contains(t, result, "スプレッド 7")
	assert.Contains(t, result, "レイズ推奨")
	assert.Contains(t, result, "勝てるランク: 4, 5, 6, 7, 8, 9, 10")
}

func TestRedDogCuiPresenter_HintOutput_StayRecommended(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	m.On("GetPhase").Return(domain.RedDogPhaseSpreadDecision)
	m.On("GetSpread").Return(4)
	m.On("GetInitialCards").Return(cards)

	result := p.HintOutput(m)
	assert.Contains(t, result, "スプレッド 4")
	assert.Contains(t, result, "ステイ推奨")
}

func TestRedDogCuiPresenter_HintOutput_BetPhase(t *testing.T) {
	p := new(RedDogCuiPresenter)
	for _, phase := range []int{domain.RedDogPhaseBet, domain.RedDogPhaseInitialDealt} {
		m := new(interfaces.MockRedDogGame)
		m.On("GetPhase").Return(phase)
		assert.Contains(t, p.HintOutput(m), "まずベットしてください")
	}
}

func TestRedDogCuiPresenter_HintOutput_GameOver(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	assert.Contains(t, p.HintOutput(m), "ゲームは終了しています")
}

func TestRedDogCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// #5539: Web はベット前に配当表を常設しているのに、CUI はベット額を決める材料が
// スプレッドの広さだけだった。
func TestRedDogCuiPresenter_Output_Paytable(t *testing.T) {
	p := new(RedDogCuiPresenter)

	outputInPhase := func(phase int) string {
		m := new(interfaces.MockRedDogGame)
		m.On("GetPhase").Return(phase)
		setupRedDogCuiMockDefaults(m)
		return p.Output(m, nil)
	}

	betOut := outputInPhase(domain.RedDogPhaseBet)
	assert.Contains(t, betOut, i18n.T("reddog.paytableHeader"))
	// **倍率はドメインの定数から出す。**表示用に書き写すと、配当を直したときに
	// 表だけが古いまま残る。
	for _, tc := range []struct {
		key  string
		mult int
	}{
		{"reddog.paySpread1", domain.RedDogPaySpread1},
		{"reddog.paySpread2", domain.RedDogPaySpread2},
		{"reddog.paySpread3", domain.RedDogPaySpread3},
		{"reddog.paySpread4Plus", domain.RedDogPaySpread4},
		{"reddog.payPair", domain.RedDogPayPair},
	} {
		assert.Contains(t, betOut, i18n.Tf(tc.key, "mult", strconv.Itoa(tc.mult)), tc.key)
	}
	// プッシュになる2ケースも Web と同じく書く。
	assert.Contains(t, betOut, i18n.T("reddog.payPush"))

	// **他フェーズでは出さない。**もう賭け終わっている。
	header := i18n.T("reddog.paytableHeader")
	assert.NotContains(t, outputInPhase(domain.RedDogPhaseSpreadDecision), header)
	assert.NotContains(t, outputInPhase(domain.RedDogPhaseEnd), header)
}

// TestRedDogCuiPresenter_Output_EndPairDealShowsNoGuide はペア初手の局で
// 案内が出ないことを見る。
//
// ペアの局は「間に入ったか」ではなく「ペアと同じランクを引いたか」で決まる
// (RedDog.go:206)。当たりランクの一覧はその規則を説明していないので、勝った局でも
// 出してはいけない。案内側は GetResult() を「間に入ったか」として読んでいるので、
// この除外が外れると、ペアを引いて勝った局に的外れなランク一覧が出る。
func TestRedDogCuiPresenter_Output_EndPairDealShowsNoGuide(t *testing.T) {
	i18n.SetLang("ja")
	rd := new(interfaces.MockRedDogGame)
	rd.On("GetChips").Return(1100).Maybe()
	rd.On("GetPhase").Return(domain.RedDogPhaseEnd).Maybe()
	rd.On("GetAnte").Return(100).Maybe()
	rd.On("GetRaise").Return(0).Maybe()
	rd.On("GetSpread").Return(0).Maybe()
	// 初期2枚が同ランク＝ペア初手。3枚目も同ランクで勝ち。
	rd.On("GetInitialCards").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	rd.On("GetThirdCard").Return(domain.NewCard(domain.CardDesignClover, 5, false)).Maybe()
	rd.On("GetResult").Return(domain.GameResultWin).Maybe()
	rd.On("GetTotalPayout").Return(1200).Maybe()
	rd.On("GetGameEndFlag").Return(true).Maybe()
	rd.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	p := new(RedDogCuiPresenter)
	out := p.Output(rd, nil)

	assert.Contains(t, out, "プレイヤーの勝ち", "前提: ペアを引き当てて勝っている")
	assert.NotContains(t, out, "当たりランク:", "ペアの局に間のランクの案内を出さない")
}
