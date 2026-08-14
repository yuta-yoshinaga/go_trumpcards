//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。** 手を一度も指していないセッションでもページが落ちません。
func TestBotifarraWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.BotifarraWebOutput{
		Players:      make([]*controller.BotifarraWebOutputPlayer, 0),
		ValidPlays:   make([]int, 0),
		CurrentTrick: make([]*controller.WebOutputTrickCard, 0),
		LastTrick:    make([]*controller.WebOutputTrickCard, 0),
		RoundPoints:  []int{0, 0},
		Scores:       []int{0, 0},
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"players", "validPlays", "currentTrick", "lastTrick"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func TestBotifarraWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"dealerIdx":0,"declarerIdx":-1,` +
		`"trumpSuit":-1,"multiplier":1,"currentTurn":0,"isHumanTurn":true,"currentTrick":[],` +
		`"lastTrick":[],"lastTrickWinner":-1,"trickCount":0,"roundPoints":[0,0],"scores":[0,0],` +
		`"gameEndFlag":false,"winnerTeam":-1,"message":""}`

	m := new(usecase.MockBotifarraInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("PlayCard", 2).Return(mockOutput)
	m.On("Declare", domain.CardDesignHeart).Return(mockOutput)
	m.On("Declare", domain.BotifarraNoTrump).Return(mockOutput)
	m.On("Delegate").Return(mockOutput)
	m.On("Double").Return(mockOutput)
	m.On("PassDouble").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("GiveUp").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	factory := func() uc.BotifarraInteractorIF { return m }
	ctrl := controller.NewBotifarraWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"play", `{"command":"play","cardIndex":2,"sessionId":"s2"}`},
		{"declare heart", `{"command":"declare","suit":3,"sessionId":"s3"}`},
		{"declare no trump", `{"command":"declare","suit":-1,"sessionId":"s4"}`},
		{"delegate", `{"command":"delegate","sessionId":"s5"}`},
		{"double", `{"command":"double","sessionId":"s6"}`},
		{"passdouble", `{"command":"passdouble","sessionId":"s7"}`},
		{"next", `{"command":"next","sessionId":"s8"}`},
		{"giveup", `{"command":"giveup","sessionId":"s9"}`},
		{"hint", `{"command":"hint","sessionId":"s10"}`},
		{"log", `{"command":"log","sessionId":"s11"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.BotifarraWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.BotifarraWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(strings.TrimSpace(recorded.Body.String()), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		execRequest(t, ctrl.Exec, strings.NewReader("{invalid")).CodeIs(http.StatusBadRequest)
	})

	// **suit を省いた declare は 400。** スートは 1..4 で 0 は無効値なので、
	// 省略を 0 に落とすと「送らなかった」と区別できません。他ゲームと同じ形で返します。
	t.Run("declare without suit is rejected", func(t *testing.T) {
		var input controller.BotifarraWebInput
		_ = json.Unmarshal([]byte(`{"command":"declare","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "suit is required") {
			t.Errorf("expected 'suit is required', got: %s", recorded.Body.String())
		}
		m.AssertNotCalled(t, "Declare", 0)
	})
}

// **切り札なし (-1) がワイヤを渡って届く。**
//
// `suit` をポインタで受けているのは、**-1 が有効な値**だからです。値型だと
// 「送らなかった」と「スペード (0)」が区別できません。
func TestBotifarraWebController_NoTrumpCrossesTheWire(t *testing.T) {
	m := new(usecase.MockBotifarraInteractor)
	m.On("Declare", domain.BotifarraNoTrump).Return(`{}`)
	m.On("Declare", domain.CardDesignSpade).Return(`{}`)

	factory := func() uc.BotifarraInteractorIF { return m }
	ctrl := controller.NewBotifarraWebController(factory)
	defer ctrl.Stop()

	var noTrump controller.BotifarraWebInput
	if err := json.Unmarshal([]byte(`{"command":"declare","suit":-1,"sessionId":"n1"}`), &noTrump); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &noTrump).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Declare", domain.BotifarraNoTrump)

	// **スートは 1..4 なので 0 は無効。** それでも「送らなかった」とは別物として
	// 届き、ドメインが弾きます——ポインタで受けているのはこの区別のためです。
	var spade controller.BotifarraWebInput
	if err := json.Unmarshal([]byte(`{"command":"declare","suit":1,"sessionId":"n2"}`), &spade); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &spade).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Declare", domain.CardDesignSpade)
}
