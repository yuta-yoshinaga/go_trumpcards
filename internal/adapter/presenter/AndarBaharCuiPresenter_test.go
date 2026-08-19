package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// fillAndarBaharCuiDefaults は未設定の呼び出しに既定値を与える。
//
// **testify は登録順に照合する**ので、個別に上書きしたい期待は**これより先に**
// 登録します。後から `On` を足しても既定の `.Maybe()` が先に一致してしまいます。
func fillAndarBaharCuiDefaults(m *interfaces.MockAndarBaharGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.AndarBaharPhaseBet).Maybe()
	m.On("GetJoker").Return((*domain.Card)(nil)).Maybe()
	m.On("GetAndarCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBaharCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetFirstColumn").Return(domain.AndarBaharBetAndar).Maybe()
	m.On("DealtCount").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBetTarget").Return(domain.AndarBaharBetAndar).Maybe()
	m.On("GetSideAmount").Return(0).Maybe()
	m.On("GetSideBand").Return(domain.AndarBaharSideNone).Maybe()
	m.On("GetWinner").Return(-1).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetMainPayout").Return(0).Maybe()
	m.On("GetSidePayout").Return(0).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetHint").Return("andarBaharHintAndarFirst").Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

// **翻訳が読めているかを、生キーではなく実際の日本語で検査します。**
// `i18n.T(key)` と突き合わせると、ロケールが落ちていても両辺が生キーになって通ります。
func TestAndarBaharCuiPresenter_Output_BetPhase(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
	// **先に配る列は賭ける前に見えている必要がある。** 配当が下がる側だからです。
	assert.Contains(t, result, "先に配る列: アンダー")
	assert.Contains(t, result, "0.9:1")
	assert.NotContains(t, result, "払い戻し:", "決着前に払戻額は出さない")
	assert.NotContains(t, result, "andarbahar.", "生キーが漏れている")
}

func TestAndarBaharCuiPresenter_Output_ShowsTheJokerAndColumns(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	m.On("GetJoker").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetFirstColumn").Return(domain.AndarBaharBetBahar)
	m.On("GetAndarCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	m.On("GetBaharCards").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	})
	m.On("DealtCount").Return(3)
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "基準札:")
	assert.Contains(t, result, "アンダー:")
	assert.Contains(t, result, "バハール:")
	assert.Contains(t, result, "配った枚数: 3")
	assert.Contains(t, result, "先に配る列: バハール", "赤の基準札はバハールから")
}

func TestAndarBaharCuiPresenter_Output_Result(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	m := new(interfaces.MockAndarBaharGame)
	m.On("GetPhase").Return(domain.AndarBaharPhaseEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinner").Return(domain.AndarBaharBetBahar)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetPayout").Return(190)
	m.On("GetBetAmount").Return(100)
	m.On("GetSideBand").Return(domain.AndarBaharSide2To5)
	m.On("GetSideAmount").Return(50)
	// **内訳は合計より前に登録する。** testify は先に一致した期待を返すので、
	// fillAndarBaharCuiDefaults の 0 が先だと内訳が出ない。
	m.On("GetMainPayout").Return(190)
	m.On("GetSidePayout").Return(0)
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "バハール に基準札と同じランクが出ました")
	assert.Contains(t, result, "払い戻し: 190")
	assert.Contains(t, result, "ベット: 100 (アンダー)")
	assert.Contains(t, result, "サイドベット: 50 (2〜5 枚)")
	// **サイドベットは別の賭け** (#5770)。メインで取ってサイドを外した回だと
	// 分かる形で内訳を出す。
	assert.Contains(t, result, i18n.Tf("andarbahar.payoutBreakdownLine", "main", "190", "side", "0"))
}

// **負のコントロール: 張っていない回に内訳は出ない** (受け入れ条件3)。
func TestAndarBaharCuiPresenter_Output_NoBreakdownWithoutASideBet(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	m.On("GetPhase").Return(domain.AndarBaharPhaseEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinner").Return(domain.AndarBaharBetBahar)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetPayout").Return(190)
	m.On("GetMainPayout").Return(190)
	m.On("GetSidePayout").Return(0)
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "払い戻し: 190")
	assert.NotContains(t, result, fixedPart("andarbahar.payoutBreakdownLine"))
}

func TestAndarBaharCuiPresenter_Output_Error(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, errors.New("Insufficient chips."))
	assert.Contains(t, result, "Insufficient chips.")
}

func TestAndarBaharCuiPresenter_Output_History(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	m := new(interfaces.MockAndarBaharGame)
	m.On("GetHistory").Return([]int{
		domain.AndarBaharBetAndar, domain.AndarBaharBetBahar, domain.AndarBaharBetAndar,
	})
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "罫線: A B A")
	assert.Contains(t, result, "アンダー 2 / バハール 1")
}

// **長い罫線と長い列は末尾だけに切り詰める。** 端末の行からあふれさせないためです。
func TestAndarBaharCuiPresenter_Output_TruncatesLongRuns(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	history := make([]int, andarBaharHistoryMaxShown+5)
	cards := make([]*domain.Card, andarBaharColumnMaxShown+4)
	for i := range cards {
		cards[i] = domain.NewCard(domain.CardDesignSpade, i%13+1, false)
	}

	m := new(interfaces.MockAndarBaharGame)
	m.On("GetHistory").Return(history)
	m.On("GetAndarCards").Return(cards)
	m.On("DealtCount").Return(len(cards))
	fillAndarBaharCuiDefaults(m)

	result := new(AndarBaharCuiPresenter).Output(m, nil)
	assert.Contains(t, result, "... ", "切り詰めたことが分かる")
	// 表示は切り詰めても、件数は全部を数える。
	assert.Contains(t, result, "アンダー 25 / バハール 0")
}

func TestAndarBaharCuiPresenter_HintAndActionLog(t *testing.T) {
	m := new(interfaces.MockAndarBaharGame)
	fillAndarBaharCuiDefaults(m)

	p := new(AndarBaharCuiPresenter)
	hint := p.HintOutput(m)
	assert.Contains(t, hint, "51.50%")
	assert.NotContains(t, hint, "andarbahar.", "生キーが漏れている")

	assert.NotEmpty(t, p.ActionLogOutput(m))
}

func TestAndarBaharCuiPresenter_UnknownValues(t *testing.T) {
	p := new(AndarBaharCuiPresenter)
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
	assert.Equal(t, "UNKNOWN", p.columnStr(99))
	assert.Equal(t, "UNKNOWN", p.bandStr(99))
	assert.Equal(t, "1 枚ちょうど", p.bandStr(domain.AndarBaharSideFirst))
	assert.Equal(t, "(なし)", p.columnCards(nil))
}
