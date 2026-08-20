//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// rbState builds a RussianBank in an arbitrary state via JSON restore.
func rbState(t *testing.T, js string) *domain.RussianBank {
	t.Helper()
	g := domain.NewDefaultRussianBank()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

const rbTwoEmptyPlayers = `"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},` +
	`{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],"cf":{"cd":1}`

func TestRussianBankCuiPresenter_Output(t *testing.T) {
	p := new(presenter.RussianBankCuiPresenter)

	t.Run("human turn prompt", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":0}`)
		out := p.Output(g, nil)
		assert.Contains(t, out, "Russian Bank")
		assert.Contains(t, out, "ファウンデーション")
		assert.Contains(t, out, "あなたの手番")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":1}`)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("stop available", func(t *testing.T) {
		// Opponent (seat 1) has an Ace on its reserve playable to an empty foundation.
		js := `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},` +
			`{"n":"CPU","c":true,"s":1,"r":[{"d":0,"v":1,"w":true}],"h":[],"w":[]}],` +
			`"cf":{"cd":1},"ph":1,"cu":0}`
		g := rbState(t, js)
		assert.True(t, g.CanCallStop())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end banners", func(t *testing.T) {
		win := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":2,"cu":0,"wn":0}`)
		assert.Contains(t, p.Output(win, nil), "勝ち")
		draw := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":2,"cu":0,"wn":-1}`)
		assert.Contains(t, p.Output(draw, nil), "引き分け")
	})

	t.Run("error block", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":0}`)
		assert.NotEmpty(t, p.Output(g, assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		assert.NotEmpty(t, p.HintOutput(g))
		// A game not on the human's turn yields the no-hint message.
		noHint := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":1}`)
		assert.NotEmpty(t, p.HintOutput(noHint))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})

	// Each board is crafted so the single best hint comes from a specific source,
	// exercising every branch of rbHintSourceName.
	t.Run("hint source names", func(t *testing.T) {
		const A = `{"d":1,"v":1,"w":true}` // A spade
		cases := []struct {
			name string
			js   string
			want string
		}{
			{"own reserve", `{"pl":[{"n":"You","c":false,"s":0,"r":[{"d":4,"v":1,"w":true}],"h":[],"w":[]},{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],"cf":{"cd":1},"ph":1,"cu":0}`, "自リザーブ"},
			{"own waste", `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[` + A + `]},{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],"cf":{"cd":1},"ph":1,"cu":0}`, "自廃札"},
			{"opp reserve", `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},{"n":"CPU","c":true,"s":1,"r":[{"d":2,"v":1,"w":true}],"h":[],"w":[]}],"cf":{"cd":1},"ph":1,"cu":0}`, "相手リザーブ"},
			{"opp waste", `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[{"d":3,"v":1,"w":true}]}],"cf":{"cd":1},"ph":1,"cu":0}`, "相手廃札"},
			{"tableau", `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],"tb":[[{"d":1,"v":2,"w":true}]],"fd":[[` + A + `]],"cf":{"cd":1},"ph":1,"cu":0}`, "タブロー0"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				g := rbState(t, tc.js)
				assert.Contains(t, p.HintOutput(g), tc.want)
			})
		}
	})
}

// #5677: タブロー列は複数枚重なる。Web は #3574 で埋もれた札のランク・スートを
// カスケード表示するようにしたのに、CUI は各列のトップ 1 枚しか出しておらず、
// 「この列の下に何があるか」に到達できなかった。
func TestRussianBankCuiPresenter_ShowsBuriedTableauCards(t *testing.T) {
	p := new(presenter.RussianBankCuiPresenter)

	// 列0 に 3 枚 (下から ♠5 / ♥7 / ♣9)、列1 は 1 枚、列2・3 は空。
	js := "{" + rbTwoEmptyPlayers + `,"ph":1,"cu":0,"tb":[` +
		`[{"d":1,"v":5,"w":true},{"d":3,"v":7,"w":true},{"d":2,"v":9,"w":true}],` +
		`[{"d":4,"v":2,"w":true}],[],[]]}`

	t.Run("lists every card in a stacked column", func(t *testing.T) {
		out := p.Output(rbState(t, js), nil)

		for _, want := range []string{"SPADE 5", "HEART 7", "CLOVER 9"} {
			assert.Contains(t, out, want, "埋もれた札も出す")
		}
	})

	// **一番上がどれかは分かるようにする。**全部並べただけでは、どちらの端が
	// トップなのか読めない。
	t.Run("keeps the top card identifiable", func(t *testing.T) {
		out := p.Output(rbState(t, js), nil)

		// 既存の 1 枚列と同じく **トップだけが [] で囲まれる**。埋もれた札は
		// その手前に下から順に並ぶ。
		assert.Contains(t, out, "SPADE 5 "+color.Red("HEART 7")+" [CLOVER 9]")
	})

	// 空列の表示は従来どおり (受け入れ条件3)。
	t.Run("an empty column still reads as empty", func(t *testing.T) {
		out := p.Output(rbState(t, js), nil)

		assert.Contains(t, out, "[-]")
	})

	// ファウンデーションはトップだけで足りる (受け入れ条件2)。積み上がる順序が
	// A から固定なので、下に何があるかは自明。
	t.Run("foundations still show only their top", func(t *testing.T) {
		fjs := "{" + rbTwoEmptyPlayers + `,"ph":1,"cu":0,"fd":[` +
			`[{"d":1,"v":1,"w":true},{"d":1,"v":2,"w":true}],[],[],[],[],[],[],[]]}`
		out := p.Output(rbState(t, fjs), nil)

		assert.Contains(t, out, "SPADE 2")
		assert.NotContains(t, out, "SPADE 1", "ファウンデーションの下の札は出さない")
	})
}
