package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// fillBotifarraDefaults は未設定の呼び出しに既定値を与える。
//
// **testify は登録順に照合する**ので、個別に上書きしたい期待は**これより先に**
// 登録します。後から On を足しても既定の `.Maybe()` が先に一致してしまいます。
func fillBotifarraDefaults(m *interfaces.MockBotifarraGame) {
	m.On("GetPhase").Return(domain.BotifarraPhasePlay).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsHumanTurn").Return(true).Maybe()
	m.On("GetValidPlayIndices", 0).Return([]int{0}).Maybe()
	m.On("GetDealerIdx").Return(0).Maybe()
	m.On("GetDeclarerIdx").Return(0).Maybe()
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade).Maybe()
	m.On("GetMultiplier").Return(domain.BotifarraMultiplierNone).Maybe()
	m.On("GetCurrentTurn").Return(0).Maybe()
	m.On("GetTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrick").Return(([]*domain.TrickCard)(nil)).Maybe()
	m.On("GetLastTrickWinner").Return(-1).Maybe()
	m.On("GetTrickCount").Return(0).Maybe()
	m.On("GetRoundPoints", 0).Return(0).Maybe()
	m.On("GetRoundPoints", 1).Return(0).Maybe()
	m.On("GetScore", 0).Return(0).Maybe()
	m.On("GetScore", 1).Return(0).Maybe()
	m.On("GetWinnerTeam").Return(-1).Maybe()
	m.On("GetPlayerCnt").Return(domain.BotifarraPlayerCnt).Maybe()
	m.On("GetPlayer", 0).Return((*domain.BotifarraPlayer)(nil)).Maybe()
	m.On("GetPlayer", 1).Return((*domain.BotifarraPlayer)(nil)).Maybe()
	m.On("GetPlayer", 2).Return((*domain.BotifarraPlayer)(nil)).Maybe()
	m.On("GetPlayer", 3).Return((*domain.BotifarraPlayer)(nil)).Maybe()
	m.On("GetConfig").Return(domain.DefaultBotifarraConfig()).Maybe()
	m.On("GetHint").Return((*domain.BotifarraHint)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

// **翻訳が読めているかを、生キーではなく実際の日本語で検査します。**
func TestBotifarraCuiPresenter_Output(t *testing.T) {
	m := new(interfaces.MockBotifarraGame)
	fillBotifarraDefaults(m)

	out := new(BotifarraCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "フェーズ: PLAY")
	assert.Contains(t, out, "切り札: スペード")
	assert.Contains(t, out, "101 点で上がり")
	assert.Contains(t, out, "合計 72")
	assert.NotContains(t, out, "botifarra.", "生キーが漏れている")
}

// **出せる札に印が付く。** 勝つ義務があるので、出せない札がふつうに残ります。
func TestBotifarraCuiPresenter_MarksLegalCards(t *testing.T) {
	game := domain.NewDefaultBotifarra()
	game.Reset()

	out := new(BotifarraCuiPresenter).Output(game, nil)
	assert.Contains(t, out, "手札:")
	assert.Contains(t, out, "勝てるなら勝つ義務")
}

func TestBotifarraCuiPresenter_ShowsTheTrick(t *testing.T) {
	m := new(interfaces.MockBotifarraGame)
	m.On("GetTrick").Return([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})
	fillBotifarraDefaults(m)

	out := new(BotifarraCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "席 1:")
}

func TestBotifarraCuiPresenter_Result(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	m := new(interfaces.MockBotifarraGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerTeam").Return(0)
	m.On("GetPhase").Return(domain.BotifarraPhaseGameEnd)
	fillBotifarraDefaults(m)

	out := new(BotifarraCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "チーム 0 の勝ちです")
	assert.Contains(t, out, "フェーズ: GAME END")
}

func TestBotifarraCuiPresenter_Error(t *testing.T) {
	m := new(interfaces.MockBotifarraGame)
	fillBotifarraDefaults(m)

	out := new(BotifarraCuiPresenter).Output(m, errors.New("その札は出せません"))
	assert.Contains(t, out, "その札は出せません")
}

func TestBotifarraCuiPresenter_Hint(t *testing.T) {
	p := new(BotifarraCuiPresenter)

	suit := domain.CardDesignHeart
	declare := new(interfaces.MockBotifarraGame)
	declare.On("GetHint").Return(&domain.BotifarraHint{Suit: &suit, Reason: "botifarraDeclareLongest"})
	fillBotifarraDefaults(declare)
	out := p.HintOutput(declare)
	assert.Contains(t, out, "ハート")
	assert.NotContains(t, out, "botifarra.", "生キーが漏れている")

	idx := 3
	play := new(interfaces.MockBotifarraGame)
	play.On("GetHint").Return(&domain.BotifarraHint{CardIndex: &idx, Reason: "botifarraMustWin"})
	fillBotifarraDefaults(play)
	assert.Contains(t, p.HintOutput(play), "3 番目")

	none := new(interfaces.MockBotifarraGame)
	fillBotifarraDefaults(none)
	assert.Contains(t, p.HintOutput(none), "助言できません")
}

func TestBotifarraCuiPresenter_UnknownValues(t *testing.T) {
	p := new(BotifarraCuiPresenter)
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
	assert.Equal(t, "切り札なし", p.trumpStr(domain.BotifarraNoTrump))
	assert.Equal(t, "クラブ", p.trumpStr(domain.CardDesignClover))
	assert.Equal(t, "ダイヤ", p.trumpStr(domain.CardDesignDiamond))
	assert.Equal(t, "DECLARE", p.phaseStr(domain.BotifarraPhaseDeclare))
	assert.Equal(t, "DELEGATED", p.phaseStr(domain.BotifarraPhaseDelegated))
	assert.Equal(t, "DOUBLE", p.phaseStr(domain.BotifarraPhaseDouble))
	assert.Equal(t, "ROUND END", p.phaseStr(domain.BotifarraPhaseRoundEnd))
}
