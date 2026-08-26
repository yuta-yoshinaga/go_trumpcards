//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// caribbeanDrawDrawRecorder is a mock interactor that keeps every slice handed
// to Draw.
//
// **「Draw が呼ばれた」だけでは足りない。** 添字を捨てても、並べ替えても、
// 1 ずれても Draw は呼ばれる。ずれた分だけ別の札が消えるので、実際に届いた
// スライスそのものを見る。
type caribbeanDrawDrawRecorder struct {
	*usecase.MockCaribbeanDrawInteractor
	received [][]int
}

func newCaribbeanDrawDrawRecorder(out string) *caribbeanDrawDrawRecorder {
	r := &caribbeanDrawDrawRecorder{MockCaribbeanDrawInteractor: new(usecase.MockCaribbeanDrawInteractor)}
	r.On("Draw", mock.Anything).Run(func(args mock.Arguments) {
		idx, _ := args.Get(0).([]int)
		r.received = append(r.received, idx)
	}).Return(out)
	r.On("Reset").Return(out).Maybe()
	r.On("Bet", mock.Anything, mock.Anything).Return(out).Maybe()
	r.On("Play").Return(out).Maybe()
	r.On("Fold").Return(out).Maybe()
	r.On("Hint").Return(out).Maybe()
	r.On("ActionLog").Return(out).Maybe()
	return r
}

// only returns the single slice Draw was called with, failing otherwise.
func (r *caribbeanDrawDrawRecorder) only(t *testing.T) []int {
	t.Helper()
	require.Len(t, r.received, 1, "Draw should have been dispatched exactly once")
	return r.received[0]
}

// --- Web controller -------------------------------------------------------

// TestCaribbeanDrawWebController_Draw covers the command the clone source
// (Caribbean Stud) has no equivalent of.
func TestCaribbeanDrawWebController_Draw(t *testing.T) {
	const out = `{"drawn":true}`

	exec := func(t *testing.T, body string) (*caribbeanDrawDrawRecorder, *recorded) {
		t.Helper()
		rec := newCaribbeanDrawDrawRecorder(out)
		ctrl := controller.NewCaribbeanDrawWebController(func() uc.CaribbeanDrawInteractorIF { return rec })
		t.Cleanup(ctrl.Stop)

		var input controller.CaribbeanDrawWebInput
		require.NoError(t, json.Unmarshal([]byte(body), &input))
		return rec, execRequest(t, ctrl.Exec, &input)
	}

	t.Run("d forwards the indices verbatim", func(t *testing.T) {
		rec, resp := exec(t, `{"command":"d","indices":[0,2],"sessionId":"d1"}`)
		resp.CodeIs(http.StatusOK)
		resp.BodyIs(out)
		// **0 始まりのまま渡す。** Web の添字はドメインと同じ 0 始まりなので、
		// ここで足し引きすると隣の札が捨てられる。
		assert.Equal(t, []int{0, 2}, rec.only(t))
	})

	t.Run("draw long form forwards the indices verbatim", func(t *testing.T) {
		rec, resp := exec(t, `{"command":"draw","indices":[1],"sessionId":"d2"}`)
		resp.CodeIs(http.StatusOK)
		resp.BodyIs(out)
		assert.Equal(t, []int{1}, rec.only(t))
	})

	t.Run("keeps the order the client sent", func(t *testing.T) {
		rec, _ := exec(t, `{"command":"d","indices":[4,0],"sessionId":"d3"}`)
		assert.Equal(t, []int{4, 0}, rec.only(t))
	})

	// **省略と null と [] は同じ「交換しない」。** どれか一つでも別扱いに
	// なると、スタンドパットしたつもりの手数料が引かれる。
	t.Run("omitted indices arrive as an empty draw", func(t *testing.T) {
		rec, resp := exec(t, `{"command":"d","sessionId":"d4"}`)
		resp.CodeIs(http.StatusOK)
		assert.Empty(t, rec.only(t))
	})

	t.Run("null indices arrive as an empty draw", func(t *testing.T) {
		rec, resp := exec(t, `{"command":"d","indices":null,"sessionId":"d5"}`)
		resp.CodeIs(http.StatusOK)
		assert.Empty(t, rec.only(t))
	})

	t.Run("empty indices arrive as an empty draw", func(t *testing.T) {
		rec, _ := exec(t, `{"command":"d","indices":[],"sessionId":"d6"}`)
		assert.Empty(t, rec.only(t))
	})

	// The controller is a router: bounds are the domain's job, so an
	// out-of-range index still has to reach it rather than being swallowed.
	t.Run("forwards an out-of-range index for the domain to reject", func(t *testing.T) {
		rec, _ := exec(t, `{"command":"d","indices":[9],"sessionId":"d7"}`)
		assert.Equal(t, []int{9}, rec.only(t))
	})

	t.Run("other commands do not draw", func(t *testing.T) {
		rec, _ := exec(t, `{"command":"p","indices":[0,1],"sessionId":"d8"}`)
		rec.AssertNotCalled(t, "Draw", mock.Anything)
		assert.Empty(t, rec.received)
	})
}

// --- CUI controller -------------------------------------------------------

// TestCaribbeanDrawCuiController_Draw pins the 1-based → 0-based conversion.
//
// **ここが 1 ずれると、黙って別の札が消える。** ドメインは範囲内の添字を
// 疑わないので、エラーも警告も出ないまま間違ったカードが交換される。
func TestCaribbeanDrawCuiController_Draw(t *testing.T) {
	exec := func(t *testing.T, command string) (*caribbeanDrawDrawRecorder, string) {
		t.Helper()
		rec := newCaribbeanDrawDrawRecorder("draw result")
		c := controller.NewCaribbeanDrawCuiController(rec)
		return rec, c.Exec(command)
	}

	t.Run("d 1 3 reaches the interactor as 0 and 2", func(t *testing.T) {
		rec, out := exec(t, "d 1 3")
		assert.Equal(t, "draw result", out)
		assert.Equal(t, []int{0, 2}, rec.only(t))
	})

	t.Run("d 1 reaches the interactor as 0", func(t *testing.T) {
		rec, out := exec(t, "d 1")
		assert.Equal(t, "draw result", out)
		assert.Equal(t, []int{0}, rec.only(t))
	})

	t.Run("d 5 reaches the interactor as 4", func(t *testing.T) {
		rec, _ := exec(t, "d 5")
		assert.Equal(t, []int{4}, rec.only(t))
	})

	t.Run("draw long form converts the same way", func(t *testing.T) {
		rec, out := exec(t, "draw 2 4")
		assert.Equal(t, "draw result", out)
		assert.Equal(t, []int{1, 3}, rec.only(t))
	})

	t.Run("keeps the order the player typed", func(t *testing.T) {
		rec, _ := exec(t, "d 4 2")
		assert.Equal(t, []int{3, 1}, rec.only(t))
	})

	t.Run("a bare d stands pat", func(t *testing.T) {
		rec, out := exec(t, "d")
		assert.Equal(t, "draw result", out)
		assert.Empty(t, rec.only(t), "no card number means no exchange")
	})

	t.Run("a bare draw stands pat", func(t *testing.T) {
		rec, out := exec(t, "draw")
		assert.Equal(t, "draw result", out)
		assert.Empty(t, rec.only(t))
	})
}

// TestCaribbeanDrawCuiController_Draw_Errors checks that a bad card number is
// rejected *before* anything reaches the interactor -- a rejected command must
// not cost the player the exchange fee.
func TestCaribbeanDrawCuiController_Draw_Errors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{"non-numeric", "d abc"},
		{"zero is below the 1-based floor", "d 0"},
		{"negative", "d -1"},
		{"a bad number after a good one", "d 1 xyz"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := newCaribbeanDrawDrawRecorder("draw result")
			c := controller.NewCaribbeanDrawCuiController(rec)

			out := c.Exec(tc.command)

			assert.Contains(t, out, msgStem("invalidCardIndex"))
			assert.True(t, msgRejected(out), "the reply must carry the rejection marker")
			rec.AssertNotCalled(t, "Draw", mock.Anything)
			assert.Empty(t, rec.received, "a rejected command must not be dispatched")
		})
	}
}
