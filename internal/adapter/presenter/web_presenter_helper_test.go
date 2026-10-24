//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testHumanSeat bool

func (s testHumanSeat) GetIsHuman() bool { return bool(s) }

type testTurnGame struct {
	players []testHumanSeat
	current int
	valid   []int
}

func (g testTurnGame) GetPlayerCnt() int             { return len(g.players) }
func (g testTurnGame) GetCurrentPlayerIdx() int      { return g.current }
func (g testTurnGame) GetPlayer(i int) testHumanSeat { return g.players[i] }
func (g testTurnGame) GetValidPlayIndices(int) []int { return g.valid }

func TestHumanTurnValidPlays(t *testing.T) {
	tests := []struct {
		name   string
		game   testTurnGame
		inPlay bool
		want   []int
	}{
		{name: "human to act", game: testTurnGame{players: []testHumanSeat{false, true}, current: 1, valid: []int{2, 4}}, inPlay: true, want: []int{2, 4}},
		{name: "CPU to act", game: testTurnGame{players: []testHumanSeat{false, true}, current: 0, valid: []int{2}}, inPlay: true, want: []int{}},
		{name: "outside play phase", game: testTurnGame{players: []testHumanSeat{true}, current: 0, valid: []int{1}}, inPlay: false, want: []int{}},
		{name: "current index out of range", game: testTurnGame{players: []testHumanSeat{true}, current: 1, valid: []int{1}}, inPlay: true, want: []int{}},
		{name: "nil domain result", game: testTurnGame{players: []testHumanSeat{true}, current: 0}, inPlay: true, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanTurnValidPlays[testHumanSeat](tt.game, tt.inPlay)
			if got == nil || len(got) != len(tt.want) {
				t.Fatalf("humanTurnValidPlays() = %v, want non-nil %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("humanTurnValidPlays() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestBuildWinnerResultMessage_Human(t *testing.T) {
	msg := buildWinnerResultMessage(0, true)
	assert.Equal(t, "ゲーム終了！ あなたの勝ち！", msg)
}

func TestBuildWinnerResultMessage_CPU(t *testing.T) {
	msg := buildWinnerResultMessage(2, false)
	assert.Equal(t, "ゲーム終了！ CPU 2の勝ち！", msg)
}

func TestBuildWinnerWebMessage_Human(t *testing.T) {
	resultMsg, code, params := buildWinnerWebMessage("hearts", 0, true)
	assert.Equal(t, "ゲーム終了！ あなたの勝ち！", resultMsg)
	assert.Equal(t, "hearts.result.humanWin", code)
	assert.Nil(t, params)
}

func TestBuildWinnerWebMessage_CPU(t *testing.T) {
	resultMsg, code, params := buildWinnerWebMessage("spades", 2, false)
	assert.Equal(t, "ゲーム終了！ CPU 2の勝ち！", resultMsg)
	assert.Equal(t, "spades.result.cpuWin", code)
	assert.Equal(t, map[string]string{"cpuId": "2"}, params)
}

func TestStringSliceOrEmpty(t *testing.T) {
	assert.Equal(t, []string{}, stringSliceOrEmpty(nil))
	input := []string{"x"}
	assert.Equal(t, input, stringSliceOrEmpty(input))
}
