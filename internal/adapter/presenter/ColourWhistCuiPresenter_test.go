package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// fillColourWhistDefaults は未設定の呼び出しに既定値を与える。
//
// **testify は登録順に照合する**ので、個別の期待は**これより先に**登録します。
func fillColourWhistDefaults(m *interfaces.MockColourWhistGame) {
	m.On("GetPhase").Return(domain.ColourWhistPhasePlay).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsHumanTurn").Return(true).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{0}).Maybe()
	m.On("GetDealerIdx").Return(0).Maybe()
	m.On("GetContract").Return(domain.ColourWhistContractSamen).Maybe()
	m.On("GetDeclarerIdx").Return(0).Maybe()
	m.On("GetPartnerIdx").Return(-1).Maybe()
	m.On("GetCalledCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade).Maybe()
	m.On("IsTroelForced").Return(false).Maybe()
	m.On("GetCurrentTurn").Return(0).Maybe()
	m.On("GetTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrickWinner").Return(-1).Maybe()
	m.On("GetTrickCount").Return(0).Maybe()
	m.On("GetDeclarerTricks").Return(0).Maybe()
	m.On("GetRoundNumber").Return(1).Maybe()
	m.On("GetPlayerCnt").Return(domain.ColourWhistPlayerCnt).Maybe()
	for i := range domain.ColourWhistPlayerCnt {
		m.On("GetPlayer", i).Return((*domain.ColourWhistPlayer)(nil)).Maybe()
		m.On("IsDeclarerSide", i).Return(false).Maybe()
		m.On("IsDeclarerSideVisible", i).Return(false).Maybe()
		m.On("HasPassed", i).Return(false).Maybe()
	}
	m.On("GetWinnerIdx").Return(-1).Maybe()
	m.On("GetConfig").Return(domain.DefaultColourWhistConfig()).Maybe()
	m.On("GetHint").Return((*domain.ColourWhistHint)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

// **翻訳が読めているかを、生キーではなく実際の日本語で検査します。**
func TestColourWhistCuiPresenter_Output(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(m)

	out := new(ColourWhistCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "フェーズ: PLAY")
	assert.Contains(t, out, "ラウンド: 1 / 8")
	assert.Contains(t, out, "サーメン")
	assert.NotContains(t, out, "colourwhist.", "生キーが漏れている")
}

// **Troel は「配りで成立した」ことを明示する。** 競りで選んだのではありません。
func TestColourWhistCuiPresenter_ShowsTroelAsForced(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	m.On("IsTroelForced").Return(true)
	m.On("GetContract").Return(domain.ColourWhistContractTroel)
	m.On("GetDeclarerIdx").Return(2)
	fillColourWhistDefaults(m)

	out := new(ColourWhistCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "トルール")
	assert.Contains(t, out, "エース3枚")
	assert.Contains(t, out, "競りはありません")
}

// **Troel でなければその行は出ない。** 負のコントロールです。
func TestColourWhistCuiPresenter_NoTroelLineWhenBid(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(m)
	assert.NotContains(t, new(ColourWhistCuiPresenter).Output(m, nil), "競りはありません")
}

func TestColourWhistCuiPresenter_ShowsTheTrickAndScores(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	m.On("GetTrick").Return([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
	})
	fillColourWhistDefaults(m)

	out := new(ColourWhistCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "席 3:")
	assert.Contains(t, out, "得点:")
}

func TestColourWhistCuiPresenter_Result(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	m := new(interfaces.MockColourWhistGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetPhase").Return(domain.ColourWhistPhaseGameEnd)
	fillColourWhistDefaults(m)

	out := new(ColourWhistCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "席 0 の勝ちです")
}

func TestColourWhistCuiPresenter_Error(t *testing.T) {
	m := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(m)
	assert.Contains(t, new(ColourWhistCuiPresenter).Output(m, errors.New("その札は出せません")),
		"その札は出せません")
}

func TestColourWhistCuiPresenter_Hint(t *testing.T) {
	p := new(ColourWhistCuiPresenter)

	contract := domain.ColourWhistContractAlleen
	bid := new(interfaces.MockColourWhistGame)
	bid.On("GetHint").Return(&domain.ColourWhistHint{Contract: &contract, Reason: "colourWhistBidStrength"})
	fillColourWhistDefaults(bid)
	assert.Contains(t, p.HintOutput(bid), "アレーン")

	idx := 2
	play := new(interfaces.MockColourWhistGame)
	play.On("GetHint").Return(&domain.ColourWhistHint{CardIndex: &idx, Reason: "colourWhistFollowSuit"})
	fillColourWhistDefaults(play)
	assert.Contains(t, p.HintOutput(play), "2 番目")

	none := new(interfaces.MockColourWhistGame)
	fillColourWhistDefaults(none)
	assert.Contains(t, p.HintOutput(none), "助言できません")
}

func TestColourWhistCuiPresenter_UnknownValues(t *testing.T) {
	p := new(ColourWhistCuiPresenter)
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
	assert.Equal(t, "未定", p.contractStr(domain.ColourWhistContractNone))
	assert.Contains(t, p.contractStr(domain.ColourWhistContractMiserie), "ミゼリー")
	assert.Equal(t, "なし", p.trumpStr(domain.ColourWhistNoTrump))
	assert.Equal(t, "クラブ", p.trumpStr(domain.CardDesignClover))
	assert.Equal(t, "BID", p.phaseStr(domain.ColourWhistPhaseBid))
	assert.Equal(t, "CALL", p.phaseStr(domain.ColourWhistPhaseCall))
}
