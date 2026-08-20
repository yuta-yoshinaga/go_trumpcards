//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockDragonTigerInteractor() *usecase.MockDragonTigerInteractor {
	m := new(usecase.MockDragonTigerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, domain.DragonTigerBetDragon).Return("bet dragon")
	m.On("Bet", 100, domain.DragonTigerBetTiger).Return("bet tiger")
	m.On("Bet", 100, domain.DragonTigerBetTie).Return("bet tie")
	m.On("ClearHistory").Return("history cleared")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestDragonTigerCuiController_Quit(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDragonTigerCuiController_Reset(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestDragonTigerCuiController_Bet_Dragon(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet dragon", c.Exec("b 100 d"))
	assert.Equal(t, "bet dragon", c.Exec("bet 100 dragon"))
	assert.Equal(t, "bet dragon", c.Exec("b 100 0"))
}

func TestDragonTigerCuiController_Bet_Tiger(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet tiger", c.Exec("b 100 t"))
	assert.Equal(t, "bet tiger", c.Exec("b 100 tiger"))
	assert.Equal(t, "bet tiger", c.Exec("b 100 1"))
}

func TestDragonTigerCuiController_Bet_Tie(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet tie", c.Exec("b 100 e"))
	assert.Equal(t, "bet tie", c.Exec("b 100 tie"))
	assert.Equal(t, "bet tie", c.Exec("b 100 2"))
}

func TestDragonTigerCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec("b"), msgStem("betAmountAndTypeRequired"))
	assert.Contains(t, c.Exec("b 100"), msgStem("betAmountAndTypeRequired"))
	assert.Contains(t, c.Exec("b abc d"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 100 x"), msgStem("invalidBetTypeDragonTiger"))
}

func TestDragonTigerCuiController_ClearHistory(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "history cleared", c.Exec("clear"))
}

func TestDragonTigerCuiController_ActionLog(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestDragonTigerCuiController_Unknown(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestDragonTigerCuiController_Empty(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

// #5585: Web は 1 クリックで同じ賭けを繰り返せるのに、CUI は毎ラウンド
// `r` のあとフルの `b <額> <種別>` を打ち直す必要があった。
func TestDragonTigerCuiControllerRebet(t *testing.T) {
	for _, alias := range []string{"rb", "rebet"} {
		t.Run(alias, func(t *testing.T) {
			di := newMockDragonTigerInteractor()
			c := controller.NewDragonTigerCuiController(di)
			assert.Equal(t, "bet tiger", c.Exec("b 100 t"))
			// **reset を挟んでから同じ賭けを打ち直す** (Web の handleRebet と同順)。
			assert.Equal(t, "bet tiger", c.Exec(alias))
			di.AssertCalled(t, "Reset")
			di.AssertNumberOfCalls(t, "Bet", 2)
		})
	}
}

// 履歴が無ければ、リセットもせずにエラーを返す。**「リセットだけされた」状態は
// 打ち直したつもりのプレイヤーには気づけない。**
func TestDragonTigerCuiControllerRebetWithoutHistory(t *testing.T) {
	di := newMockDragonTigerInteractor()
	c := controller.NewDragonTigerCuiController(di)

	out := c.Exec("rebet")
	assert.Contains(t, out, i18n.T("dragontiger.noPreviousBet"))
	di.AssertNotCalled(t, "Reset")
	di.AssertNotCalled(t, "Bet", mock.Anything, mock.Anything)
}

// **不正なベットは覚えない。**覚えると、通らない賭けを rb が繰り返す。
func TestDragonTigerCuiControllerDoesNotRememberARejectedBet(t *testing.T) {
	di := newMockDragonTigerInteractor()
	c := controller.NewDragonTigerCuiController(di)

	c.Exec("b 100 zz") // 種別が不正
	out := c.Exec("rb")
	assert.Contains(t, out, i18n.T("dragontiger.noPreviousBet"))
	di.AssertNotCalled(t, "Bet", mock.Anything, mock.Anything)
}

// 賭け直しは**最後の**ベットを繰り返す。
func TestDragonTigerCuiControllerRebetUsesTheLatestBet(t *testing.T) {
	di := newMockDragonTigerInteractor()
	c := controller.NewDragonTigerCuiController(di)
	di.On("Bet", 250, domain.DragonTigerBetTie).Return("bet 250 tie")

	c.Exec("b 100 d")
	c.Exec("b 250 e")
	assert.Equal(t, "bet 250 tie", c.Exec("rb"))
	di.AssertNumberOfCalls(t, "Bet", 3)
}
