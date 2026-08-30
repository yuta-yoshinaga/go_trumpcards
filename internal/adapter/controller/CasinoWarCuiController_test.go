//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockCasinoWarInteractor() *usecase.MockCasinoWarInteractor {
	m := new(usecase.MockCasinoWarInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Surrender").Return("surrender result")
	m.On("War").Return("war result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestCasinoWarCuiController_Quit(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCasinoWarCuiController_Reset(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestCasinoWarCuiController_Bet(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "bet result", c.Exec("bet 100"))
}

func TestCasinoWarCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec("b"), msgBetAmountRequired())
	assert.Contains(t, c.Exec("b abc"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 0"), msgInvalidBetAmountPrefix())
}

func TestCasinoWarCuiController_Surrender(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "surrender result", c.Exec("surrender"))
}

func TestCasinoWarCuiController_War(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "war result", c.Exec("war"))
}

func TestCasinoWarCuiController_ActionLog(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestCasinoWarCuiController_Unknown(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
	assert.Contains(t, c.Exec("rebe"), "rebet")
}

func TestCasinoWarCuiController_Empty(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

// #6379: Web は 1 クリックで同じ賭けを繰り返せるのに、CUI は毎ラウンド
// `r` のあとフルの `b <額>` を打ち直す必要があった。
func TestCasinoWarCuiControllerRebet(t *testing.T) {
	for _, alias := range []string{"rb", "rebet"} {
		t.Run(alias, func(t *testing.T) {
			ci := newMockCasinoWarInteractor()
			c := controller.NewCasinoWarCuiController(ci)
			assert.Equal(t, "bet result", c.Exec("b 100"))
			// **reset を挟んでから同じ賭けを打ち直す** (Web の handleRebet と同順)。
			assert.Equal(t, "bet result", c.Exec(alias))
			ci.AssertCalled(t, "Reset")
			ci.AssertNumberOfCalls(t, "Bet", 2)
			if assert.Len(t, ci.Calls, 3) {
				assert.Equal(t, "Bet", ci.Calls[0].Method)
				assert.Equal(t, "Reset", ci.Calls[1].Method)
				assert.Equal(t, "Bet", ci.Calls[2].Method)
				assert.Equal(t, 100, ci.Calls[2].Arguments.Int(0))
			}
		})
	}
}

// 履歴が無ければ、リセットもせずにエラーを返す。**「リセットだけされた」状態は
// 打ち直したつもりのプレイヤーには気づけない。**
func TestCasinoWarCuiControllerRebetWithoutHistory(t *testing.T) {
	ci := newMockCasinoWarInteractor()
	c := controller.NewCasinoWarCuiController(ci)

	out := c.Exec("rebet")
	assert.Contains(t, out, i18n.T("casinowar.noPreviousBet"))
	ci.AssertNotCalled(t, "Reset")
	ci.AssertNotCalled(t, "Bet", mock.Anything)
}

// **不正なベットは覚えない。**覚えると、通らない賭けを rb が繰り返す。
// 不正な額 (b abc) の後に rb を打っても、直前の正しいベットが繰り返される。
func TestCasinoWarCuiControllerDoesNotRememberARejectedBet(t *testing.T) {
	ci := newMockCasinoWarInteractor()
	c := controller.NewCasinoWarCuiController(ci)

	// 最初は履歴なしで不正入力
	c.Exec("b abc")
	out := c.Exec("rb")
	assert.Contains(t, out, i18n.T("casinowar.noPreviousBet"))
	ci.AssertNotCalled(t, "Bet", mock.Anything)

	// 正しいベットを行った後に不正入力を与えても、直前の正しいベットを維持する
	assert.Equal(t, "bet result", c.Exec("b 100"))
	c.Exec("b abc")
	assert.Equal(t, "bet result", c.Exec("rb"))
	ci.AssertNumberOfCalls(t, "Bet", 2)
}

// 賭け直しは**最後の**ベットを繰り返す。
func TestCasinoWarCuiControllerRebetUsesTheLatestBet(t *testing.T) {
	ci := newMockCasinoWarInteractor()
	c := controller.NewCasinoWarCuiController(ci)
	ci.On("Bet", 250).Return("bet 250")

	c.Exec("b 100")
	c.Exec("b 250")
	assert.Equal(t, "bet 250", c.Exec("rb"))
	ci.AssertNumberOfCalls(t, "Bet", 3)
}
