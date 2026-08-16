//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

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

// 勝因が KV スナップショットを往復すること。
//
// 決着の表示は「決着した**次の**リクエスト」で読まれる。載せ忘れると
// 復元時点で勝因が消え、presenter が通常勝ちに寄せるので、テストが全部緑でも
// 本番でだけ直っていない状態になる (#5765)。
func TestLingerLonger_WinReasonSurvivesTheSnapshot(t *testing.T) {
	l := newLingerLongerForReasonTest(t)
	// **脱落席は手札を持てない** (復元時の検査がそれを見る)。全員が同時に
	// 出し切った局面を作るので、手札も空にしてから脱落させる。
	// 出し切った札は捨て札に移る。復元時の検査は 52 枚が
	// 手札 + 場 + 山札 + 捨て札に揃っていることを見るので、数を合わせる。
	for i := range l.players {
		l.discarded += l.players[i].GetCardsSize()
		l.GiveHandForTest(i)
		l.players[i].SetEliminatedAt(i + 1)
	}
	// 脱落の順番は 1..eliminatedCnt の並べ替えでなければ復元が拒む。
	l.eliminatedCnt = len(l.players)
	if !l.checkGameEnd(2) {
		t.Fatal("zero active seats should end the game")
	}

	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back LingerLonger
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := back.GetWinReason(); got != LingerLongerWinLastTrick {
		t.Errorf("restored GetWinReason() = %q, want %q", got, LingerLongerWinLastTrick)
	}
	if back.GetWinnerIdx() != l.GetWinnerIdx() {
		t.Errorf("restored winner = %d, want %d", back.GetWinnerIdx(), l.GetWinnerIdx())
	}
}

// 決着済みなのに勝因が無いスナップショットは拒む。載せ忘れが「通常勝ち」として
// 黙って通ると、直したはずの誤表示がそのまま戻る。
func TestLingerLonger_SnapshotRejectsAMissingOrUnknownWinReason(t *testing.T) {
	l := newLingerLongerForReasonTest(t)
	for i := 1; i < l.config.PlayerCnt; i++ {
		l.discarded += l.players[i].GetCardsSize()
		l.GiveHandForTest(i)
		l.players[i].SetEliminatedAt(i)
	}
	l.eliminatedCnt = l.config.PlayerCnt - 1
	if !l.checkGameEnd(0) {
		t.Fatal("one active seat should end the game")
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, tc := range []struct{ name, replace string }{
		{"missing", `"wr":""`},
		{"unknown", `"wr":"heldOn"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tampered := strings.Replace(string(data),
				`"wr":"`+LingerLongerWinLasted+`"`, tc.replace, 1)
			if tampered == string(data) {
				t.Fatal("改竄が効いていない。フィールド名が変わったらここも直すこと")
			}
			var back LingerLonger
			if err := json.Unmarshal([]byte(tampered), &back); err == nil {
				t.Errorf("%s win reason was accepted", tc.name)
			}
		})
	}
}
