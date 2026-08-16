//go:build test

package domain

import "testing"

// 勝因を区別する。
//
// checkGameEnd には 2 通りの勝ちがあり、山札が尽きて**全員が同時に手札 0 枚**に
// なったときは「最後まで持ち続けた人」が存在しないので、最後のトリックを取った
// 人を勝ちにしている。ところが finish() のログも presenter の文言も常に
// 「最後まで手札を持ち続けました」で、このゲームの主題そのものである規則に
// 反した勝因を表示していた (#5765)。
func TestLingerLonger_WinReasonDistinguishesTheTwoWins(t *testing.T) {
	t.Run("the last player holding cards lasted", func(t *testing.T) {
		l := newLingerLongerForReasonTest(t)
		// 席 0 以外を全部脱落させる。
		for i := 1; i < l.config.PlayerCnt; i++ {
			l.players[i].SetEliminatedAt(i)
		}
		if !l.checkGameEnd(2) {
			t.Fatal("one active seat should end the game")
		}
		if got := l.GetWinReason(); got != LingerLongerWinLasted {
			t.Errorf("GetWinReason() = %q, want %q", got, LingerLongerWinLasted)
		}
		if l.GetWinnerIdx() != 0 {
			t.Errorf("winner = %d, want the surviving seat 0", l.GetWinnerIdx())
		}
	})

	t.Run("everyone emptying at once wins on the last trick", func(t *testing.T) {
		l := newLingerLongerForReasonTest(t)
		for i := range l.players {
			// eliminatedAt は 1 以上で「脱落済み」なので 0 は使えない。
			l.players[i].SetEliminatedAt(i + 1)
		}
		if !l.checkGameEnd(2) {
			t.Fatal("zero active seats should end the game")
		}
		if got := l.GetWinReason(); got != LingerLongerWinLastTrick {
			t.Errorf("GetWinReason() = %q, want %q", got, LingerLongerWinLastTrick)
		}
		// 勝者は最後のトリックを取った席であって、席 0 ではない。
		if l.GetWinnerIdx() != 2 {
			t.Errorf("winner = %d, want the last-trick winner 2", l.GetWinnerIdx())
		}
	})

	t.Run("giving up is neither", func(t *testing.T) {
		l := newLingerLongerForReasonTest(t)
		l.GiveUp()
		if got := l.GetWinReason(); got != LingerLongerWinGiveUp {
			t.Errorf("GetWinReason() = %q, want %q", got, LingerLongerWinGiveUp)
		}
	})

	// 決着前に勝因を訊かれたら空。ここが既定値で埋まっていると、presenter は
	// 決着していない盤面にも勝因を出せてしまう。
	t.Run("no reason before the game ends", func(t *testing.T) {
		l := newLingerLongerForReasonTest(t)
		if got := l.GetWinReason(); got != "" {
			t.Errorf("GetWinReason() = %q, want empty before the end", got)
		}
	})
}

func newLingerLongerForReasonTest(t *testing.T) *LingerLonger {
	t.Helper()
	l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: 4})
	l.Reset()
	return l
}
