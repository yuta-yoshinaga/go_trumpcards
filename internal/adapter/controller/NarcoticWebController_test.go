//go:build test

package controller_test

import (
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func narcoticIntPtr(v int) *int { return &v }

func setupNarcoticWebTest(t *testing.T) (*usecase.MockNarcoticInteractor, *controller.NarcoticWebController, string) {
	t.Helper()
	mockOutput := `{"columns":[],"stockCount":0,"discardCount":0,"phase":0,"moveCount":0,"message":""}`
	giMock := new(usecase.MockNarcoticInteractor)
	factory := func() uc.NarcoticInteractorIF { return giMock }
	ctrl := controller.NewNarcoticWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return giMock, ctrl, mockOutput
}

func narcoticPost(t *testing.T, handler http.HandlerFunc, input *controller.NarcoticWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, input)
}

func TestNarcoticWebController_Commands(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)

	giMock.On("Reset").Return(mockOutput)
	giMock.On("Draw").Return(mockOutput)
	giMock.On("GiveUp").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)
	giMock.On("Undo").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", "reset"},
		{"draw", "draw"},
		{"giveup", "giveup"},
		{"hint", "hint"},
		{"log", "log"},
		{"undo", "undo"},
		{"short-r", "r"},
		{"short-d", "d"},
		{"short-g", "g"},
		{"short-u", "u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &controller.NarcoticWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.command, SessionID: "s1"},
			}
			rec := narcoticPost(t, ctrl.Exec, input)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestNarcoticWebController_Remove(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	// **col は要らない。**揃った4枚をまとめて捨てるので、対象は盤面から一意。
	giMock.On("Remove").Return(mockOutput)

	narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s2"},
	})

	rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s2"},
	})
	rec.CodeIs(http.StatusOK)
	giMock.AssertCalled(t, "Remove")
}

func TestNarcoticWebController_Remove_ShortAlias(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Remove").Return(mockOutput)

	narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s5"},
	})

	rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s5"},
	})
	rec.CodeIs(http.StatusOK)
}

// **`move` は列が要る。**クローン元は remove にも列が要ったが、Narcotic で
// 列を取るのは重ねる側だけ。
func TestNarcoticWebController_Move_MissingParams(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("Reset").Return(mockOutput)

	narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s3"},
	})

	rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s3"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestNarcoticWebController_Redeal(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Redeal").Return(mockOutput)

	narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s7"},
	})
	rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "redeal", SessionID: "s7"},
	})
	rec.CodeIs(http.StatusOK)
}

func TestNarcoticWebController_Move(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Move", 1).Return(mockOutput)

	narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s6"},
	})

	t.Run("move with col", func(t *testing.T) {
		rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s6"},
			Col:          narcoticIntPtr(1),
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("move short alias", func(t *testing.T) {
		rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "mv", SessionID: "s6"},
			Col:          narcoticIntPtr(1),
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("move missing col", func(t *testing.T) {
		rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s6"},
		})
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestNarcoticWebController_UndoN(t *testing.T) {
	giMock, ctrl, mockOutput := setupNarcoticWebTest(t)
	giMock.On("UndoN", 3).Return(mockOutput)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		})
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestNarcoticWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupNarcoticWebTest(t)

	rec := narcoticPost(t, ctrl.Exec, &controller.NarcoticWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "unknown", SessionID: "s4"},
	})
	rec.CodeIs(http.StatusBadRequest)
}
