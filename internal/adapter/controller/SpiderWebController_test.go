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

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func mustSpiderOutputJSON(msg string) string {
	out := &controller.SpiderWebOutput{
		Tableau:       [][]*controller.SpiderWebOutputTableauCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSpiderOutputJSON: %v", err))
	}
	return string(b)
}

func TestSpiderWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"score":0,"difficulty":0,"message":""}`
	expectedBody := mockOutput

	siMock := new(usecase.MockSpiderInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)

	factory := func() uc.SpiderInteractorIF { return siMock }
	ctrl := controller.NewSpiderWebController(factory)
	defer ctrl.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/spider/exec", ctrl.Exec),
	)
	api.SetApp(router)

	t.Run("quit q", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustSpiderOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustSpiderOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("deal d", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("deal", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"deal","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"autocomplete","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Move command
	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.SpiderWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.SpiderWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
			To:           &controller.SpiderWebZone{Zone: "tableau", Col: intPtr(4)},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("unsupported command", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustSpiderOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustSpiderOutputJSON("param error."))
	})

	t.Run("empty session", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustSpiderOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.SpiderWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustSpiderOutputJSON("param error."))
	})
}

func TestSpiderWebController_MoveErrors(t *testing.T) {
	siMock := new(usecase.MockSpiderInteractor)
	factory := func() uc.SpiderInteractorIF { return siMock }
	ctrl := controller.NewSpiderWebController(factory)
	defer ctrl.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/spider/exec", ctrl.Exec))
	api.SetApp(router)

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.SpiderWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		input := controller.SpiderWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.SpiderWebZone{Zone: "tableau"},
			To:           &controller.SpiderWebZone{Zone: "tableau"},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.SpiderWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.SpiderWebZone{Zone: "waste"},
			To:           &controller.SpiderWebZone{Zone: "tableau"},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestSpiderWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"score":0,"difficulty":2,"message":""}`

	siMock := new(usecase.MockSpiderInteractor)
	siMock.On("ResetWithConfig", domain.SpiderConfig{Difficulty: domain.SpiderDifficulty2Suit}).Return(mockOutput)

	factory := func() uc.SpiderInteractorIF { return siMock }
	ctrl := controller.NewSpiderWebController(factory)
	defer ctrl.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/spider/exec", ctrl.Exec))
	api.SetApp(router)

	difficulty := 2
	input := controller.SpiderWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.SpiderWebConfig{Difficulty: &difficulty},
	}
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/spider/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestSpiderWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockSpiderInteractor)
	factory := func() uc.SpiderInteractorIF { return siMock }
	c := controller.NewSpiderWebController(factory)
	c.Stop()
	c.Stop()
}
