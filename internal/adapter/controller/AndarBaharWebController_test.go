//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustAndarBaharOutputJSON(msg string) string {
	out := &controller.AndarBaharWebOutput{
		AndarCards:    make([]*controller.WebOutputCard, 0),
		BaharCards:    make([]*controller.WebOutputCard, 0),
		History:       make([]int, 0),
		Winner:        -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAndarBaharOutputJSON: %v", err))
	}
	return string(b)
}

// **既定の出力も配列で返る。** 手を一度も指していないセッションでもページが落ちません。
func TestAndarBaharWebController_DefaultOutputHasArrays(t *testing.T) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mustAndarBaharOutputJSON("bye.")), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"andarCards", "baharCards", "history"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func TestAndarBaharWebController_Method(t *testing.T) {
	mockOutput := `{"andarCards":[],"baharCards":[],"firstColumn":0,"dealtCount":0,` +
		`"phase":1,"chips":1000,"betAmount":0,"betTarget":0,"sideAmount":0,"sideBand":-1,` +
		`"winner":-1,"result":0,"payout":0,"history":[],"message":""}`

	abMock := new(usecase.MockAndarBaharInteractor)
	abMock.On("Reset").Return(mockOutput)
	abMock.On("Bet", 100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone).Return(mockOutput)
	abMock.On("Bet", 100, domain.AndarBaharBetBahar, 50, domain.AndarBaharSide6To10).Return(mockOutput)
	abMock.On("ClearHistory").Return(mockOutput)
	abMock.On("Hint").Return(mockOutput)
	abMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.AndarBaharInteractorIF { return abMock }
	ctrl := controller.NewAndarBaharWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
	}{
		{"quit q", `{"command":"q","sessionId":"s1"}`},
		{"quit long", `{"command":"quit","sessionId":"s2"}`},
		{"reset", `{"command":"reset","sessionId":"s3"}`},
		{"reset r", `{"command":"r","sessionId":"s4"}`},
		{"bet", `{"command":"bet","amount":100,"target":0,"sessionId":"s5"}`},
		{"bet b", `{"command":"b","amount":100,"target":0,"sessionId":"s6"}`},
		{"bet with side", `{"command":"b","amount":100,"target":1,"sideAmount":50,"sideBand":2,"sessionId":"s7"}`},
		{"clear", `{"command":"clear","sessionId":"s8"}`},
		{"hint", `{"command":"hint","sessionId":"s9"}`},
		{"log", `{"command":"log","sessionId":"s10"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.AndarBaharWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustAndarBaharOutputJSON("bye."))
			} else {
				recorded.BodyIs(mockOutput)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.AndarBaharWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(recorded.Body.String())
		if !strings.Contains(body, "Unsupported command") {
			t.Errorf("expected Unsupported command in body, got: %s", body)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
	})
}

// **4 つのベット引数がワイヤを渡って届く。** ページのモックはワイヤを通らないので、
// ここで綴りと並びを固定します。
func TestAndarBaharWebController_BetArgumentsCrossTheWire(t *testing.T) {
	abMock := new(usecase.MockAndarBaharInteractor)
	abMock.On("Bet", 30, domain.AndarBaharBetBahar, 20, domain.AndarBaharSide11To15).Return(`{}`)

	factory := func() uc.AndarBaharInteractorIF { return abMock }
	ctrl := controller.NewAndarBaharWebController(factory)
	defer ctrl.Stop()

	var input controller.AndarBaharWebInput
	body := `{"command":"bet","amount":30,"target":1,"sideAmount":20,"sideBand":3,"sessionId":"w1"}`
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	abMock.AssertCalled(t, "Bet", 30, domain.AndarBaharBetBahar, 20, domain.AndarBaharSide11To15)
}

// **サイドベットを省いた普通のベットが弾かれない。**
//
// JSON で `sideBand` を省くと 0 になり、0 は「1 枚目の帯」という有効な値なので、
// 素通しすると賭け金 0 のサイドベットとしてドメインに拒否されます。
func TestAndarBaharWebController_BetWithoutSideBetSendsNone(t *testing.T) {
	abMock := new(usecase.MockAndarBaharInteractor)
	abMock.On("Bet", 100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone).Return(`{}`)

	factory := func() uc.AndarBaharInteractorIF { return abMock }
	ctrl := controller.NewAndarBaharWebController(factory)
	defer ctrl.Stop()

	var input controller.AndarBaharWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"n1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	abMock.AssertCalled(t, "Bet", 100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone)

	// 帯を送っても金額が 0 ならサイドベットにしない。
	var withBand controller.AndarBaharWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","amount":100,"sideBand":4,"sessionId":"n2"}`), &withBand); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &withBand).CodeIs(http.StatusOK)
	abMock.AssertNumberOfCalls(t, "Bet", 2)
}
