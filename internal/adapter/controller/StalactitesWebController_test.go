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
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustStalactitesOutputJSON(msg string) string {
	out := &controller.StalactitesWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		Cells:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustStalactitesOutputJSON: %v", err))
	}
	return string(b)
}

func TestStalactitesWebController(t *testing.T) {
	mockOutput := mustStalactitesOutputJSON("ok")

	fiMock := new(usecase.MockStalactitesInteractor)
	fiMock.On("Reset").Return(mockOutput)
	fiMock.On("GiveUp").Return(mockOutput)
	fiMock.On("Hint").Return(mockOutput)
	fiMock.On("AutoComplete").Return(mockOutput)
	fiMock.On("ActionLog").Return(mockOutput)
	fiMock.On("Undo").Return(mockOutput)
	fiMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	fiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	fiMock.On("MoveTableauToStalactites", 0, 1).Return(mockOutput)
	fiMock.On("MoveStalactitesToTableau", 1, 3).Return(mockOutput)
	fiMock.On("MoveStalactitesToFoundation", 0).Return(mockOutput)

	factory := func() uc.StalactitesInteractorIF { return fiMock }
	ctrl := controller.NewStalactitesWebController(factory)

	expectedBody := mockOutput

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "quit q",
			body:       `{"command":"q","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   mustStalactitesOutputJSON("bye."),
		},
		{
			name:       "quit",
			body:       `{"command":"quit","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   mustStalactitesOutputJSON("bye."),
		},
		{
			name:       "reset r",
			body:       `{"command":"r","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "reset",
			body:       `{"command":"reset","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "giveup g",
			body:       `{"command":"g","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "giveup",
			body:       `{"command":"giveup","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "hint h",
			body:       `{"command":"h","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "hint",
			body:       `{"command":"hint","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "autocomplete ac",
			body:       `{"command":"ac","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "autocomplete",
			body:       `{"command":"autocomplete","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "log",
			body:       `{"command":"log","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "log l",
			body:       `{"command":"l","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "undo u",
			body:       `{"command":"u","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "undo",
			body:       `{"command":"undo","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "unsupported command",
			body:       `{"command":"xyz","sessionId":"s1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustStalactitesOutputJSON("Unsupported command."),
		},
		{
			name:       "empty command",
			body:       `{"command":"","sessionId":"s1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustStalactitesOutputJSON("param error."),
		},
		{
			name:       "empty session",
			body:       `{"command":"r","sessionId":""}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustStalactitesOutputJSON("param error."),
		},
		{
			name:       "session too long",
			body:       fmt.Sprintf(`{"command":"r","sessionId":"%s"}`, strings.Repeat("a", controller.SessionMaxIDLen+1)),
			wantStatus: http.StatusBadRequest,
			wantBody:   mustStalactitesOutputJSON("param error."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, json.RawMessage(tt.body))
			recorded.CodeIs(tt.wantStatus)
			recorded.BodyIs(tt.wantBody)
		})
	}

	// Move tests using struct bodies
	type moveTest struct {
		name       string
		input      controller.StalactitesWebInput
		wantStatus int
		wantBody   string
	}

	moveTests := []moveTest{
		{
			name: "move tableau to tableau",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
				To:           &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(4)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to foundation",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(1)},
				To:           &controller.StalactitesWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to stalactites",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(0)},
				To:           &controller.StalactitesWebZone{Zone: "stalactites", Cell: intPtr(1)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move stalactites to tableau",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "stalactites", Cell: intPtr(1)},
				To:           &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(3)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move stalactites to foundation",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "stalactites", Cell: intPtr(0)},
				To:           &controller.StalactitesWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
	}

	for _, tt := range moveTests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, tt.input)
			recorded.CodeIs(tt.wantStatus)
			recorded.BodyIs(tt.wantBody)
		})
	}
}

func TestStalactitesWebController_UndoN(t *testing.T) {
	mockOutput := mustStalactitesOutputJSON("ok")

	fiMock := new(usecase.MockStalactitesInteractor)
	fiMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.StalactitesInteractorIF { return fiMock }
	ctrl := controller.NewStalactitesWebController(factory)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.StalactitesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.StalactitesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestStalactitesWebController_MoveErrors(t *testing.T) {
	fiMock := new(usecase.MockStalactitesInteractor)
	factory := func() uc.StalactitesInteractorIF { return fiMock }
	ctrl := controller.NewStalactitesWebController(factory)

	tests := []struct {
		name  string
		input controller.StalactitesWebInput
	}{
		{
			name: "move without from/to",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			},
		},
		{
			name: "move tableau to tableau missing params",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau"},
				To:           &controller.StalactitesWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move tableau to foundation no col",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau"},
				To:           &controller.StalactitesWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move tableau to stalactites missing params",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "tableau"},
				To:           &controller.StalactitesWebZone{Zone: "stalactites"},
			},
		},
		{
			name: "move stalactites to tableau missing params",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "stalactites"},
				To:           &controller.StalactitesWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move stalactites to foundation no cell",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "stalactites"},
				To:           &controller.StalactitesWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move invalid zones",
			input: controller.StalactitesWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.StalactitesWebZone{Zone: "invalid"},
				To:           &controller.StalactitesWebZone{Zone: "tableau", Col: intPtr(0)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, tt.input)
			recorded.CodeIs(http.StatusBadRequest)
		})
	}
}

func TestStalactitesWebController_Stop(t *testing.T) {
	fiMock := new(usecase.MockStalactitesInteractor)
	factory := func() uc.StalactitesInteractorIF { return fiMock }
	ctrl := controller.NewStalactitesWebController(factory)
	ctrl.Stop()
	ctrl.Stop() // idempotent
}
