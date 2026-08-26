//go:build test

package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

// advanceSkat は盤面を「初期状態と区別できる」ところまで進める。
//
// **初期状態のまま往復させても意味がない。** 落ちたフィールドが初期値と
// 同じなら、比較は通ってしまう。ビッドを終えて配り・宣言まで進めてから保存する。
func advanceSkat(t *testing.T) *Skat {
	t.Helper()
	s := NewDefaultSkat()
	s.Reset()
	for i := 0; i < 200 && s.GetPhase() != SkatPhasePlay; i++ {
		switch {
		case s.IsHumanBidTurn():
			_ = s.PlayerBid(true)
		case s.GetPhase() == SkatPhaseBid:
			s.CpuBid()
		case s.GetPhase() == SkatPhaseSkatPickup:
			if s.IsHumanDeclarerTurn() {
				_ = s.PlayerPickSkat(true)
			} else {
				s.CpuPickSkat()
			}
		case s.GetPhase() == SkatPhaseDiscard:
			if s.IsHumanDeclarerTurn() {
				_ = s.PlayerDiscard(0, 1)
			} else {
				s.CpuDiscard()
			}
		case s.GetPhase() == SkatPhaseGameDeclaration:
			if s.IsHumanDeclarerTurn() {
				_ = s.PlayerDeclareGame(SkatGameGrand, 0)
			} else {
				s.CpuDeclareGame()
			}
		default:
			return s
		}
	}
	return s
}

// TestSkatJSONRoundTrip は、盤面が JSON を往復しても**指し続けられる**ことを見る。
//
// # なぜこの形か
//
// Skat は非公開フィールドしか持たないので、`MarshalJSON` が無いと
// `encoding/json` は `{}` の 2 バイトしか出さない —— **エラーは出ない**。
// Worker はリクエストごとに KV から盤面を復元するので、保存が空だと毎回
// 初期状態の卓が作り直され、ゲームが進行しない (#6215)。
//
// 「フィールドが等しい」ではなく「復元した盤で指し続けられる」を見るのは、
// 初期値と区別が付かない状態で比べても通ってしまうため。
func TestSkatJSONRoundTrip(t *testing.T) {
	s := advanceSkat(t)

	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// **空でないこと。** これが 2 バイトなら他の検査は全部無意味。
	if len(blob) < 100 {
		t.Fatalf("Skat marshalled to %d bytes (%s) — the board is not being saved", len(blob), blob)
	}

	var got Skat
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.GetPhase() != s.GetPhase() {
		t.Errorf("phase = %v, want %v", got.GetPhase(), s.GetPhase())
	}
	if got.GetPlayerCnt() != s.GetPlayerCnt() {
		t.Fatalf("player count = %d, want %d", got.GetPlayerCnt(), s.GetPlayerCnt())
	}
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if a, b := got.GetPlayer(i).GetCardsSize(), s.GetPlayer(i).GetCardsSize(); a != b {
			t.Errorf("seat %d hand = %d cards, want %d", i, a, b)
		}
	}

	// **ラウンド状態を丸ごと比べる。** 個別のフィールドを列挙すると、
	// 列挙し忘れたフィールドがワイヤ形式から落ちても気づけない ——
	// 実測: DeclarerIdx を MarshalJSON から外しても、フェーズと手札だけ見る
	// 検査は通ってしまった。
	if !reflect.DeepEqual(got.round, s.round) {
		t.Errorf("round state did not survive the round trip\n got: %+v\nwant: %+v", got.round, s.round)
	}

	// **復元した盤で指し続けられること。** ここが本番の経路。
	if got.GetPhase() == SkatPhasePlay {
		valid := got.GetValidPlayIndices(got.GetCurrentPlayerIdx())
		if len(valid) == 0 {
			t.Fatal("restored board offers no legal play — the hand cannot continue")
		}
	}
}

// TestSkatJSONRejectsBadShape は、壊れた payload を弾くことを見る。
func TestSkatJSONRejectsBadShape(t *testing.T) {
	for _, tc := range []struct{ name, payload string }{
		{"wrong player count", `{"ps":[null,null]}`},
		{"empty players", `{"ps":[]}`},
		{"too many trick cards", `{"ps":[null,null,null],"ct":[{"pi":0},{"pi":1},{"pi":2},{"pi":0}]}`},
		// **添字に使う値そのもの。** 長さだけ見ても、席番号が範囲外なら
		// 次のリクエストで s.players[...] が panic する。
		{"currentPlayerIdx out of range", `{"ps":[null,null,null],"ci":3}`},
		// -1 は「まだ誰の手番でもない」の番兵なので通す。
		{"currentPlayerIdx below the -1 sentinel", `{"ps":[null,null,null],"ci":-2}`},
		{"declarerIdx out of range", `{"ps":[null,null,null],"de":9}`},
		{"declarerIdx below the -1 sentinel", `{"ps":[null,null,null],"de":-2}`},
		{"dealerIdx out of range", `{"ps":[null,null,null],"di":7}`},
		{"bidderIdx out of range", `{"ps":[null,null,null],"bi":42}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s Skat
			if err := json.Unmarshal([]byte(tc.payload), &s); err == nil {
				t.Errorf("expected an error for %s", tc.name)
			}
		})
	}
}

// TestSkatJSONCarriesEveryField は、ワイヤ形式から**どのフィールドが落ちても**
// 検出できることを見る。
//
// # なぜ実戦の局面では足りないか
//
// 進めた盤で往復比較しても、値がたまたま**ゼロ値と同じ**フィールドは、
// ワイヤ形式から落としても復元後に一致してしまう。実測: `declarerIdx` を
// MarshalJSON から外しても、席 0 が declarer だった (= 0 = ゼロ値) ため
// `reflect.DeepEqual` が通った。
//
// そこで全フィールドに**ゼロ値と異なる目印**を入れてから往復させる。
// 現実には起きない組み合わせでよい —— 検査しているのはゲームの規則ではなく
// ワイヤ形式だから。
func TestSkatJSONCarriesEveryField(t *testing.T) {
	s := NewDefaultSkat()
	s.Reset()

	card := func(v int) *Card { return NewCard(CardDesignSpade, v, false) }
	s.round = skatRoundState{
		phase:            SkatPhaseTrickEnd,
		roundNumber:      7,
		trickNumber:      5,
		currentPlayerIdx: 2,
		currentTrick:     []*TrickCard{{PlayerIdx: 1, Card: card(10)}},
		leadPlayerIdx:    1,
		dealerIdx:        2,
		forehandIdx:      1,
		middlehandIdx:    2,
		rearhandIdx:      1,
		bidderIdx:        2,
		responderIdx:     1,
		bidStep:          4,
		currentBid:       24,
		auctionRound:     2,
		round1Winner:     1,
		declarerIdx:      2,
		passedAtCall:     [SkatPlayerCnt]bool{false, true, false},
		pickedSkat:       true,
		gameType:         SkatGameNull,
		trumpSuit:        CardDesignHeart,
		skat:             []*Card{card(7), card(8)},
		originalSkat:     []*Card{card(9), card(11)},
		declarerHand:     []*Card{card(12), card(13)},
		gameValue:        48,
		breakdown:        &SkatScoreBreakdown{},
		declarerCardPts:  61,
		defendersCardPts: 59,
		winnerSide:       1,
		gameEndFlag:      true,
	}
	s.round.actionLog = []*ActionLogEntry{{PlayerIdx: 0, ActionType: "bid", Detail: "18"}}

	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Skat
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(got.round, s.round) {
		t.Errorf("a field was lost in the round trip\n got: %+v\nwant: %+v", got.round, s.round)
	}
	if !reflect.DeepEqual(got.config, s.config) {
		t.Errorf("config lost: got %+v want %+v", got.config, s.config)
	}
}
