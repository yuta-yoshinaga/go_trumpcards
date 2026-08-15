//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockChemindeFerInteractor() *usecase.MockChemindeFerInteractor {
	m := new(usecase.MockChemindeFerInteractor)
	m.On("Reset").Return("reset result")
	m.On("SetStake", 200).Return("staked 200")
	m.On("PlaceBet", 0, 50).Return("bet 50")
	m.On("PlaceBet", 0, 0).Return("passed")
	m.On("PunterDraw").Return("punter drew")
	m.On("PunterStand").Return("punter stood")
	m.On("BankerDraw").Return("banker drew")
	m.On("BankerStand").Return("banker stood")
	m.On("DrawOrStand", true).Return("drew")
	m.On("DrawOrStand", false).Return("stood")
	m.On("PassBank").Return("bank passed")
	m.On("NextRound").Return("next round")
	m.On("GiveUp").Return("gave up")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestChemindeFerCuiController_QuitAndReset(t *testing.T) {
	c := controller.NewChemindeFerCuiController(newMockChemindeFerInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "reset result", c.Exec("r"))
}

func TestChemindeFerCuiController_StakeAndBet(t *testing.T) {
	m := newMockChemindeFerInteractor()
	c := controller.NewChemindeFerCuiController(m)

	assert.Equal(t, "staked 200", c.Exec("stake 200"))
	assert.True(t, msgRejected(c.Exec("stake")))
	assert.True(t, msgRejected(c.Exec("stake xyz")))

	assert.Equal(t, "bet 50", c.Exec("bet 50"))
	// **bet 0 は「降りる」。** 下限を 0 にしていないと降りられない。
	assert.Equal(t, "passed", c.Exec("bet 0"))
	m.AssertCalled(t, "PlaceBet", 0, 0)
	assert.True(t, msgRejected(c.Exec("bet")))
}

// **側を省いたら手番の側へ。** 明示すればその側へ。
func TestChemindeFerCuiController_DrawAndStandRouting(t *testing.T) {
	for _, tt := range []struct {
		command string
		want    string
		method  string
		args    []any
	}{
		{"draw", "drew", "DrawOrStand", []any{true}},
		{"stand", "stood", "DrawOrStand", []any{false}},
		{"draw p", "punter drew", "PunterDraw", nil},
		{"stand p", "punter stood", "PunterStand", nil},
		{"draw b", "banker drew", "BankerDraw", nil},
		{"stand b", "banker stood", "BankerStand", nil},
		{"draw punter", "punter drew", "PunterDraw", nil},
		{"stand banker", "banker stood", "BankerStand", nil},
	} {
		t.Run(tt.command, func(t *testing.T) {
			m := newMockChemindeFerInteractor()
			c := controller.NewChemindeFerCuiController(m)
			assert.Equal(t, tt.want, c.Exec(tt.command))
			m.AssertCalled(t, tt.method, tt.args...)
			// **側を明示したら手番まかせの経路は通らない。**
			if tt.method != "DrawOrStand" {
				m.AssertNotCalled(t, "DrawOrStand", true)
				m.AssertNotCalled(t, "DrawOrStand", false)
			}
		})
	}
}

func TestChemindeFerCuiController_RemainingCommands(t *testing.T) {
	c := controller.NewChemindeFerCuiController(newMockChemindeFerInteractor())
	assert.Equal(t, "bank passed", c.Exec("pass"))
	assert.Equal(t, "next round", c.Exec("next"))
	assert.Equal(t, "gave up", c.Exec("giveup"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestChemindeFerCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewChemindeFerCuiController(newMockChemindeFerInteractor())
	assert.NotEmpty(t, c.Exec("zzz"))
	assert.NotEmpty(t, c.Exec(""))
}
