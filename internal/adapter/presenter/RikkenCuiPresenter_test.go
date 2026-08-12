package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// fillRikkenDefaults は未設定の呼び出しに既定値を与える。
//
// **testify は登録順に照合する**ので、個別に上書きしたい期待は**これより先に**
// 登録します。
func fillRikkenDefaults(m *interfaces.MockRikkenGame) {
	m.On("GetPhase").Return(domain.RikkenPhasePlay).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsHumanTurn").Return(true).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{0}).Maybe()
	m.On("IsDeclarerSide", 0).Return(true).Maybe()
	m.On("IsDeclarerSide", 1).Return(false).Maybe()
	m.On("IsDeclarerSide", 2).Return(false).Maybe()
	m.On("IsDeclarerSide", 3).Return(false).Maybe()
	m.On("HasPassed", 0).Return(false).Maybe()
	m.On("HasPassed", 1).Return(false).Maybe()
	m.On("HasPassed", 2).Return(false).Maybe()
	m.On("HasPassed", 3).Return(false).Maybe()
	m.On("GetDealerIdx").Return(0).Maybe()
	m.On("GetContract").Return(domain.RikkenContractRik).Maybe()
	m.On("GetDeclarerIdx").Return(0).Maybe()
	m.On("GetPartnerIdx").Return(-1).Maybe()
	m.On("GetCalledCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade).Maybe()
	m.On("GetCurrentTurn").Return(0).Maybe()
	m.On("GetTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrickWinner").Return(-1).Maybe()
	m.On("GetTrickCount").Return(0).Maybe()
	m.On("GetDeclarerTricks").Return(0).Maybe()
	m.On("GetRoundNumber").Return(1).Maybe()
	m.On("GetPlayerCnt").Return(domain.RikkenPlayerCnt).Maybe()
	for i := range domain.RikkenPlayerCnt {
		m.On("GetPlayer", i).Return((*domain.RikkenPlayer)(nil)).Maybe()
	}
	m.On("GetWinnerIdx").Return(-1).Maybe()
	m.On("GetConfig").Return(domain.DefaultRikkenConfig()).Maybe()
	m.On("GetHint").Return((*domain.RikkenHint)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

// **翻訳が読めているかを、生キーではなく実際の日本語で検査します。**
func TestRikkenCuiPresenter_Output(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	fillRikkenDefaults(m)

	out := new(RikkenCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "フェーズ: PLAY")
	assert.Contains(t, out, "ラウンド: 1 / 8")
	assert.Contains(t, out, "リク")
	assert.Contains(t, out, "スペード")
	assert.NotContains(t, out, "rikken.", "生キーが漏れている")
}

// **契約は4種類とも名前で出す。**
func TestRikkenCuiPresenter_ContractNames(t *testing.T) {
	p := new(RikkenCuiPresenter)
	assert.Contains(t, p.contractStr(domain.RikkenContractRik), "リク")
	assert.Contains(t, p.contractStr(domain.RikkenContractMisere), "ミゼール")
	assert.Contains(t, p.contractStr(domain.RikkenContractSolo), "ソロ")
	assert.Contains(t, p.contractStr(domain.RikkenContractOpenMisere), "オープン")
	assert.Equal(t, "未定", p.contractStr(domain.RikkenContractNone))
}

func TestRikkenCuiPresenter_ShowsTheTrickAndScores(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	m.On("GetTrick").Return([]*domain.TrickCard{
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
	})
	fillRikkenDefaults(m)

	out := new(RikkenCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "席 2:")
	assert.Contains(t, out, "得点:")
}

func TestRikkenCuiPresenter_Result(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	m := new(interfaces.MockRikkenGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetPhase").Return(domain.RikkenPhaseGameEnd)
	fillRikkenDefaults(m)

	out := new(RikkenCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "席 0 の勝ちです")
	assert.Contains(t, out, "フェーズ: GAME END")
}

func TestRikkenCuiPresenter_Error(t *testing.T) {
	m := new(interfaces.MockRikkenGame)
	fillRikkenDefaults(m)

	out := new(RikkenCuiPresenter).Output(m, errors.New("その札は出せません"))
	assert.Contains(t, out, "その札は出せません")
}

func TestRikkenCuiPresenter_Hint(t *testing.T) {
	p := new(RikkenCuiPresenter)

	contract := domain.RikkenContractSolo
	bid := new(interfaces.MockRikkenGame)
	bid.On("GetHint").Return(&domain.RikkenHint{Contract: &contract, Reason: "rikkenBidStrength"})
	fillRikkenDefaults(bid)
	assert.Contains(t, p.HintOutput(bid), "ソロ")

	idx := 4
	play := new(interfaces.MockRikkenGame)
	play.On("GetHint").Return(&domain.RikkenHint{CardIndex: &idx, Reason: "rikkenFollowSuit"})
	fillRikkenDefaults(play)
	assert.Contains(t, p.HintOutput(play), "4 番目")

	none := new(interfaces.MockRikkenGame)
	fillRikkenDefaults(none)
	assert.Contains(t, p.HintOutput(none), "助言できません")
}

func TestRikkenCuiPresenter_UnknownValues(t *testing.T) {
	p := new(RikkenCuiPresenter)
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
	assert.Equal(t, "なし", p.trumpStr(domain.RikkenNoTrump))
	assert.Equal(t, "クラブ", p.trumpStr(domain.CardDesignClover))
	assert.Equal(t, "ダイヤ", p.trumpStr(domain.CardDesignDiamond))
	assert.Equal(t, "BID", p.phaseStr(domain.RikkenPhaseBid))
	assert.Equal(t, "CALL", p.phaseStr(domain.RikkenPhaseCall))
	assert.Equal(t, "ROUND END", p.phaseStr(domain.RikkenPhaseRoundEnd))
}

// **CUI もオープンミゼールで宣言者の手札を見せる。**
func TestRikkenCuiPresenter_OpenMisereShowsTheOpenHand(t *testing.T) {
	cpu := domain.NewRikkenPlayer(false)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	m := new(interfaces.MockRikkenGame)
	m.On("GetContract").Return(domain.RikkenContractOpenMisere)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.RikkenNoTrump)
	m.On("GetPlayer", 1).Return(cpu)
	fillRikkenDefaults(m)

	out := new(RikkenCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "公開手札（席 1）")
	assert.NotContains(t, out, "rikken.", "生キーが漏れている")
}

// **ミゼールでは公開しない。** 負のコントロールです。
func TestRikkenCuiPresenter_PlainMisereHasNoOpenHand(t *testing.T) {
	cpu := domain.NewRikkenPlayer(false)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	m := new(interfaces.MockRikkenGame)
	m.On("GetContract").Return(domain.RikkenContractMisere)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.RikkenNoTrump)
	m.On("GetPlayer", 1).Return(cpu)
	fillRikkenDefaults(m)

	assert.NotContains(t, new(RikkenCuiPresenter).Output(m, nil), "公開手札")
}
