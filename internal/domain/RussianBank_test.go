//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// rbCard 表向きカードを生成するテストヘルパ。
func rbCard(design, value int) *Card { return NewCard(design, value, true) }

// newRbGame Reset 済みのゲームを返す。
func newRbGame() *RussianBank {
	g := NewDefaultRussianBank()
	g.Reset()
	return g
}

// rbClearBoard 配り済みの盤面を空にして決定的なテスト状態を作る。
func rbClearBoard(g *RussianBank) {
	for i := range g.tableau {
		g.tableau[i] = nil
	}
	for i := range g.foundations {
		g.foundations[i] = nil
	}
	for _, p := range g.players {
		p.resetPiles()
	}
}

func TestRussianBank_ResetDeals(t *testing.T) {
	g := newRbGame()
	if g.GetPhase() != RussianBankPhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	if len(g.GetPlayers()) != RussianBankPlayerCnt {
		t.Fatalf("players = %d, want %d", len(g.GetPlayers()), RussianBankPlayerCnt)
	}
	for i, p := range g.GetPlayers() {
		if p.ReserveSize() != RussianBankReserveSize {
			t.Errorf("seat %d reserve = %d, want %d", i, p.ReserveSize(), RussianBankReserveSize)
		}
		if p.HandSize() != 52-RussianBankReserveSize {
			t.Errorf("seat %d hand = %d, want %d", i, p.HandSize(), 52-RussianBankReserveSize)
		}
		if p.WasteSize() != 0 {
			t.Errorf("seat %d waste = %d, want 0", i, p.WasteSize())
		}
	}
	if g.GetCurrentPlayer() != 0 {
		t.Errorf("current = %d, want 0 (human first)", g.GetCurrentPlayer())
	}
}

func TestRussianBank_TableauRules(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	// 空列は任意のカードを受け入れる。
	if !g.rbCanPlaceTableau(rbCard(CardDesignHeart, 7), 0) {
		t.Error("empty tableau column should accept any card")
	}
	// 交互色・降順のみ許可。
	g.tableau[0] = []*Card{rbCard(CardDesignSpade, 8)} // 黒8
	if !g.rbCanPlaceTableau(rbCard(CardDesignHeart, 7), 0) {
		t.Error("red 7 on black 8 should be legal")
	}
	if g.rbCanPlaceTableau(rbCard(CardDesignClover, 7), 0) {
		t.Error("black 7 on black 8 should be illegal (same color)")
	}
	if g.rbCanPlaceTableau(rbCard(CardDesignHeart, 6), 0) {
		t.Error("red 6 on black 8 should be illegal (not consecutive)")
	}
}

func TestRussianBank_FoundationRules(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	// 空本は A のみ。
	if !g.rbCanPlaceFoundation(rbCard(CardDesignHeart, 1), 0) {
		t.Error("empty foundation should accept Ace")
	}
	if g.rbCanPlaceFoundation(rbCard(CardDesignHeart, 2), 0) {
		t.Error("empty foundation should reject non-Ace")
	}
	// 同スート昇順。
	g.foundations[0] = []*Card{rbCard(CardDesignHeart, 1)}
	if !g.rbCanPlaceFoundation(rbCard(CardDesignHeart, 2), 0) {
		t.Error("heart 2 on heart A should be legal")
	}
	if g.rbCanPlaceFoundation(rbCard(CardDesignSpade, 2), 0) {
		t.Error("spade 2 on heart A should be illegal (suit)")
	}
}

func TestRussianBank_MoveReserveToFoundationAndWin(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	g.current = 0
	// リザーブにエース1枚だけ → ファウンデーションへ出すと空になり勝利。
	g.players[0].pushReserve(rbCard(CardDesignDiamond, 1))
	if err := g.MoveToFoundation(RussianBankSource{Zone: RussianBankZoneReserve}); err != nil {
		t.Fatalf("MoveToFoundation: %v", err)
	}
	if g.GetPhase() != RussianBankPhaseGameEnd {
		t.Fatalf("phase = %d, want GameEnd", g.GetPhase())
	}
	if g.GetWinner() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinner())
	}
}

func TestRussianBank_MoveToTableauIllegal(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	g.current = 0
	g.players[0].pushWaste(rbCard(CardDesignHeart, 6))
	g.tableau[0] = []*Card{rbCard(CardDesignHeart, 8)} // 同色・非連続
	if err := g.MoveToTableau(RussianBankSource{Zone: RussianBankZoneWaste}, 0); err == nil {
		t.Error("expected illegal-move error for red 6 on red 8")
	}
}

func TestRussianBank_DiscardEndsTurn(t *testing.T) {
	g := newRbGame()
	if g.GetCurrentPlayer() != 0 {
		t.Fatal("expected human first")
	}
	handBefore := g.players[0].HandSize()
	if err := g.Discard(); err != nil {
		t.Fatalf("Discard: %v", err)
	}
	if g.players[0].HandSize() != handBefore-1 {
		t.Errorf("hand = %d, want %d", g.players[0].HandSize(), handBefore-1)
	}
	if g.players[0].WasteSize() != 1 {
		t.Errorf("waste = %d, want 1", g.players[0].WasteSize())
	}
	if g.GetCurrentPlayer() != 1 {
		t.Errorf("current = %d, want 1 after discard", g.GetCurrentPlayer())
	}
}

func TestRussianBank_CallStop(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	g.current = 0
	// 相手 (seat1) のリザーブトップに A があり、ファウンデーションへ出せる状態。
	g.players[1].pushReserve(rbCard(CardDesignClover, 1))
	if err := g.CallStop(); err != nil {
		t.Fatalf("valid stop should succeed: %v", err)
	}
	if g.GetStopPoints(0) != 1 {
		t.Errorf("stop points = %d, want 1", g.GetStopPoints(0))
	}
	// 違反が無い状態では stop はエラー。
	g.players[1].popReserve()
	if err := g.CallStop(); err == nil {
		t.Error("stop without violation should error")
	}
	// CPU 手番中は stop 不可。
	g.current = 1
	if err := g.CallStop(); err == nil {
		t.Error("stop on CPU turn should error")
	}
}

func TestRussianBank_UndoRestores(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	g.current = 0
	g.players[0].pushReserve(rbCard(CardDesignSpade, 5))
	g.players[0].pushReserve(rbCard(CardDesignDiamond, 1))
	if g.CanUndo() {
		t.Error("should not be able to undo before any move")
	}
	if err := g.MoveToFoundation(RussianBankSource{Zone: RussianBankZoneReserve}); err != nil {
		t.Fatalf("MoveToFoundation: %v", err)
	}
	if !g.CanUndo() {
		t.Fatal("should be able to undo after a move")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if g.players[0].ReserveSize() != 2 {
		t.Errorf("reserve after undo = %d, want 2", g.players[0].ReserveSize())
	}
	if len(g.foundations[0]) != 0 {
		t.Errorf("foundation after undo = %d, want 0", len(g.foundations[0]))
	}
}

func TestRussianBank_JSONRoundTrip(t *testing.T) {
	g := newRbGame()
	_ = g.Discard() // 状態を少し進める
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultRussianBank()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g2.GetCurrentPlayer() != g.GetCurrentPlayer() {
		t.Errorf("current mismatch: %d vs %d", g2.GetCurrentPlayer(), g.GetCurrentPlayer())
	}
	if g2.GetPlayer(0).HandSize() != g.GetPlayer(0).HandSize() {
		t.Errorf("hand size mismatch after round-trip")
	}
	if g2.GetPlayer(0).WasteSize() != g.GetPlayer(0).WasteSize() {
		t.Errorf("waste size mismatch after round-trip")
	}
}

func TestRussianBank_UnmarshalRejectsBadInput(t *testing.T) {
	twoPlayers := `{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}`
	cases := map[string]string{
		"wrong player count":   `{"pl":[],"cu":0}`,
		"nil player":           `{"pl":[null,null],"cu":0}`,
		"current out of range": `{"pl":[` + twoPlayers + `],"cu":9}`,
		"invalid config":       `{"pl":[` + twoPlayers + `],"cf":{"cd":99},"cu":0}`,
	}
	for name, js := range cases {
		t.Run(name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(js), NewDefaultRussianBank()); err == nil {
				t.Errorf("expected error for %s", name)
			}
		})
	}
}

func TestRussianBank_Hint(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	g.current = 0
	g.players[0].pushReserve(rbCard(CardDesignHeart, 1)) // A → ファウンデーション手があるはず
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected a hint when a foundation move exists")
	}
	if !h.ToFoundation {
		t.Error("best hint should be the foundation move")
	}
}

func TestRussianBank_CpuFullGameTerminates(t *testing.T) {
	for _, diff := range []RussianBankCpuDifficulty{
		RussianBankCpuDifficultyEasy, RussianBankCpuDifficultyNormal, RussianBankCpuDifficultyHard,
	} {
		g := NewRussianBank(RussianBankConfig{CpuDifficulty: diff})
		g.Reset()
		// 人間も CPU と同じ貪欲ロジックで自動プレイし、決着まで進める。
		for i := 0; i < 100000 && g.GetPhase() == RussianBankPhasePlaying; i++ {
			if g.GetCurrentPlayer() == 0 {
				if m, ok := g.bestCpuMove(); ok {
					g.applyCpuMove(m)
					continue
				}
				_ = g.Discard()
			} else {
				g.RunCpuTurn()
			}
		}
		if g.GetPhase() != RussianBankPhaseGameEnd {
			t.Errorf("difficulty %d: game did not terminate", diff)
		}
	}
}

func TestRussianBank_StalemateEndsGame(t *testing.T) {
	g := newRbGame()
	rbClearBoard(g)
	// 両者とも手札・盤面手なし → 連続パスで停滞決着。
	g.current = 0
	g.players[0].pushReserve(rbCard(CardDesignSpade, 13))  // 出せない K
	g.players[1].pushReserve(rbCard(CardDesignClover, 12)) // 出せない Q×2枚
	g.players[1].pushReserve(rbCard(CardDesignClover, 11))
	_ = g.Discard() // seat0 パス
	_ = g.Discard() // seat1 パス → 停滞
	if g.GetPhase() != RussianBankPhaseGameEnd {
		t.Fatalf("phase = %d, want GameEnd (stalemate)", g.GetPhase())
	}
	if g.GetWinner() != 0 {
		t.Errorf("winner = %d, want 0 (fewer reserve cards)", g.GetWinner())
	}
}
