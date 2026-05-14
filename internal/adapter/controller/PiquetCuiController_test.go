//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPiquetInteractor() *mockusecase.MockPiquetInteractor {
	return new(mockusecase.MockPiquetInteractor)
}

func TestPiquetCuiControllerQuit(t *testing.T) {
	c := NewPiquetCuiController(newMockPiquetInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPiquetCuiControllerReset(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("Reset").Return("reset_output")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestPiquetCuiControllerExchangeElder(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("ExchangeElder", []int{0, 1, 2}).Return("elder_out")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "elder_out", c.Exec("e 0,1,2"))
	assert.Equal(t, "elder_out", c.Exec("elder 0,1,2"))
}

func TestPiquetCuiControllerExchangeYoungerZero(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("ExchangeYounger", []int(nil)).Return("younger_zero")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "younger_zero", c.Exec("y"))
}

func TestPiquetCuiControllerExchangeYoungerInvalid(t *testing.T) {
	pi := newMockPiquetInteractor()
	c := NewPiquetCuiController(pi)
	got := c.Exec("y abc")
	assert.Contains(t, got, "abc")
}

func TestPiquetCuiControllerDeclare(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("ResolveDeclaration").Return("decl_out")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "decl_out", c.Exec("d"))
	assert.Equal(t, "decl_out", c.Exec("declare"))
}

func TestPiquetCuiControllerPlay(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("Play", 3).Return("play_out")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "play_out", c.Exec("p 3"))
}

func TestPiquetCuiControllerPlayMissingArg(t *testing.T) {
	pi := newMockPiquetInteractor()
	c := NewPiquetCuiController(pi)
	got := c.Exec("p")
	// PromptRequest returns a non-empty string asking for card index
	assert.NotEmpty(t, got)
}

func TestPiquetCuiControllerNextDeal(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("NextDeal").Return("next_out")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "next_out", c.Exec("nd"))
	assert.Equal(t, "next_out", c.Exec("nextdeal"))
}

func TestPiquetCuiControllerHintAndLog(t *testing.T) {
	pi := newMockPiquetInteractor()
	pi.On("Hint").Return("hint_out")
	pi.On("ActionLog").Return("log_out")
	c := NewPiquetCuiController(pi)
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestParseIndexList(t *testing.T) {
	tests := []struct {
		in      string
		want    []int
		wantErr bool
	}{
		{"", nil, false},
		{"0", []int{0}, false},
		{"1,2,3", []int{1, 2, 3}, false},
		{"0 , 1 , 2", []int{0, 1, 2}, false},
		{"abc", nil, true},
	}
	for _, tt := range tests {
		got, err := parseIndexList(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseIndexList(%q) err=%v wantErr=%v", tt.in, err, tt.wantErr)
		}
		if !tt.wantErr {
			assert.Equal(t, tt.want, got, "input=%q", tt.in)
		}
	}
}
