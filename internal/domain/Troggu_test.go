//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTrogguForTest はゲームを開始した状態で返す。
func newTrogguForTest(t *testing.T) *Troggu {
	t.Helper()
	g := NewDefaultTroggu()
	g.Reset()
	return g
}

// trogguTablePlayers は 4 席 (席 0 が人間) を返す。
func trogguTablePlayers() []*TrogguPlayer {
	ps := make([]*TrogguPlayer, TrogguPlayerCnt)
	ps[0] = NewTrogguPlayer(true)
	for i := 1; i < TrogguPlayerCnt; i++ {
		ps[i] = NewTrogguPlayer(false)
	}
	return ps
}

// trogguStep は CPU 側 / 自動進行を 1 歩進める。
func trogguStep(g *Troggu) {
	switch g.GetPhase() {
	case TrogguPhaseBid:
		g.CpuBid()
	case TrogguPhasePlay:
		g.CpuPlayCard()
	case TrogguPhaseTrickEnd:
		g.NextTrick()
	case TrogguPhaseRoundEnd:
		g.NextRound()
	}
}

// trogguPlayOut は終局まで進める (人間は推奨手に従う)。
func trogguPlayOut(t *testing.T, g *Troggu) {
	t.Helper()
	for range 5000 {
		if g.GetGameEndFlag() {
			return
		}
		if !g.IsHumanTurn() {
			trogguStep(g)
			continue
		}
		switch g.GetPhase() {
		case TrogguPhaseBid:
			require.NoError(t, g.PlayerPass())
		case TrogguPhasePlay:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NotNil(t, h.CardIndex)
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
		default:
			trogguStep(g)
		}
	}
	t.Fatal("終局しなかった")
}

// trogguCardsInPlay は盤上のすべての札を集める。
func trogguCardsInPlay(g *Troggu) []*Card {
	out := make([]*Card, 0, TrogguDeckSize)
	out = append(out, g.talon...)
	for _, tc := range g.GetCurrentTrick() {
		out = append(out, tc.Card)
	}
	for _, p := range g.GetPlayers() {
		for i := range p.GetCardsSize() {
			out = append(out, p.GetCard(i))
		}
		for _, trick := range p.GetTricksTaken() {
			out = append(out, trick...)
		}
	}
	return out
}

// assertTrogguDeckIntact は 78 枚が 1 枚ずつ揃っていることを確かめる。
func assertTrogguDeckIntact(t *testing.T, g *Troggu) {
	t.Helper()
	cards := trogguCardsInPlay(g)
	assert.Len(t, cards, TrogguDeckSize, "盤上の総枚数")
	seen := map[[2]int]bool{}
	for _, c := range cards {
		require.NotNil(t, c)
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "札 %v が重複", key)
		seen[key] = true
	}
}

// --- 設定 ---

func TestTrogguConfig_Validate(t *testing.T) {
	tests := []struct {
		name string
		cfg  TrogguConfig
		ok   bool
	}{
		{"既定", DefaultTrogguConfig(), true},
		{"最小ディール", TrogguConfig{TargetDeals: TrogguMinDeals}, true},
		{"最大ディール", TrogguConfig{TargetDeals: TrogguMaxDeals}, true},
		{"ディール 0", TrogguConfig{TargetDeals: 0}, false},
		{"ディール超過", TrogguConfig{TargetDeals: TrogguMaxDeals + 1}, false},
		{"難易度が範囲外", TrogguConfig{CpuDifficulty: 9, TargetDeals: 4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ok {
				assert.NoError(t, tt.cfg.Validate())
				return
			}
			assert.Error(t, tt.cfg.Validate())
		})
	}
}

// --- 配り ---

func TestTrogguReset_DealsEighteenEachAndSixTalon(t *testing.T) {
	g := newTrogguForTest(t)
	for i, p := range g.GetPlayers() {
		assert.Equal(t, TrogguHandSize, p.GetCardsSize(), "席 %d の手札", i)
	}
	assert.Equal(t, TrogguTalonSize, g.GetTalonSize())
	assert.Equal(t, TrogguPhaseBid, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assertTrogguDeckIntact(t, g)
}

// タロー 78 枚の構成をデッキから数える。
func TestTrogguDeckComposition(t *testing.T) {
	suits, trumps, excuse := 0, 0, 0
	for _, c := range buildFrenchTarotDeck() {
		switch {
		case frenchTarotIsExcuse(c):
			excuse++
		case frenchTarotIsTrump(c):
			trumps++
		default:
			suits++
		}
	}
	assert.Equal(t, 56, suits, "スート札")
	assert.Equal(t, 21, trumps, "アトゥ")
	assert.Equal(t, 1, excuse, "エクスキューズ")
	assert.Equal(t, TrogguDeckSize, suits+trumps+excuse)
	assert.Equal(t, TrogguDeckSize, TrogguPlayerCnt*TrogguHandSize+TrogguTalonSize)
}

// --- 入札 ---

func TestTrogguBid_RejectsPassAndLowBids(t *testing.T) {
	g := newTrogguForTest(t)
	for !g.IsHumanTurn() {
		g.CpuBid()
		if g.GetPhase() != TrogguPhaseBid {
			t.Skip("この配りでは人間の入札手番が来なかった")
		}
	}
	assert.Error(t, g.PlayerBid(TrogguBidPass), "パスが入札として通っている")

	high := g.GetHighestBid()
	if high >= TrogguBidMisere {
		t.Skip("既に最高入札が出ている")
	}
	// 現在の最高入札と同じ値は通らない。
	if high > TrogguBidPass {
		assert.Error(t, g.PlayerBid(high))
	}
	require.NoError(t, g.PlayerBid(TrogguBidMisere))
	assert.Equal(t, TrogguBidMisere, g.GetHighestBid())
}

func TestTrogguValidBid(t *testing.T) {
	assert.False(t, TrogguValidBid(TrogguBidPass))
	for _, b := range []TrogguBid{TrogguBidTrois, TrogguBidSolo, TrogguBidPiccolo, TrogguBidMisere} {
		assert.True(t, TrogguValidBid(b), "%s が宣言できない", TrogguBidName(b))
	}
	assert.False(t, TrogguValidBid(TrogguBidMisere+1))
}

func TestTrogguBid_Errors(t *testing.T) {
	t.Run("プレイフェーズでは入札できない", func(t *testing.T) {
		g := newTrogguForTest(t)
		g.phase = TrogguPhasePlay
		assert.Error(t, g.PlayerBid(TrogguBidSolo))
		assert.Error(t, g.PlayerPass())
	})
	t.Run("終局後", func(t *testing.T) {
		g := newTrogguForTest(t)
		g.gameEndFlag = true
		assert.ErrorIs(t, g.PlayerBid(TrogguBidSolo), ErrGameEnded)
		assert.ErrorIs(t, g.PlayerPass(), ErrGameEnded)
	})
	t.Run("CPU の入札手番", func(t *testing.T) {
		g := newTrogguForTest(t)
		g.bidPlayerIdx = 1
		assert.ErrorIs(t, g.PlayerBid(TrogguBidSolo), ErrNotHumanTurn)
		assert.ErrorIs(t, g.PlayerPass(), ErrNotHumanTurn)
	})
}

// **誰も落札しなければ配り直す。** 契約が無いと勝ち方そのものが決まらない。
func TestTrogguAllPass_ThrowsInTheDeal(t *testing.T) {
	g := NewTroggu(trogguTablePlayers(), TrogguConfig{TargetDeals: 2})
	g.Reset()
	for range TrogguPlayerCnt {
		if g.GetPhase() != TrogguPhaseBid {
			break
		}
		g.applyPass(g.GetBidPlayerIdx())
	}
	assert.Equal(t, TrogguPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assert.Equal(t, TrogguBidPass, g.GetContract())
	assert.Equal(t, TrogguOutcomeNone, g.GetOutcome())
	assert.Nil(t, g.GetBreakdown(), "流局に精算がある")
	// 得点は動かない。
	for i := range TrogguPlayerCnt {
		assert.Equal(t, 0, g.GetPlayerScore(i))
	}
	// 次のディールへ進める。
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, TrogguPhaseBid, g.GetPhase())
}

// --- プレイ ---

// trogguAtPlay は指定の契約でプレイフェーズに入った局面を返す。
func trogguAtPlay(t *testing.T, contract TrogguBid) *Troggu {
	t.Helper()
	g := newTrogguForTest(t)
	g.highestBid = contract
	g.highestBidder = 0
	g.finalizeBid()
	require.Equal(t, TrogguPhasePlay, g.GetPhase())
	require.Equal(t, 0, g.GetDeclarerIdx())
	return g
}

func TestTrogguValidPlays_MustFollowTheLead(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	lead := g.GetCurrentPlayerIdx()
	assert.Len(t, g.GetValidPlayIndices(lead), g.GetPlayer(lead).GetCardsSize())

	g.playCard(lead, 0)
	next := g.GetCurrentPlayerIdx()
	led := g.ledSuit()
	p := g.GetPlayer(next)
	hasLed := false
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if frenchTarotIsExcuse(c) {
			continue
		}
		if (led == FrenchTarotTrumpDesign && frenchTarotIsTrump(c)) ||
			(led != FrenchTarotTrumpDesign && !frenchTarotIsTrump(c) && c.GetDesign() == led) {
			hasLed = true
		}
	}
	if !hasLed {
		return
	}
	for _, i := range g.GetValidPlayIndices(next) {
		c := p.GetCard(i)
		if frenchTarotIsExcuse(c) {
			continue // **エクスキューズはいつでも出せる。**
		}
		if led == FrenchTarotTrumpDesign {
			assert.True(t, frenchTarotIsTrump(c), "アトゥのリードに他スートが合法")
			continue
		}
		assert.Equal(t, led, c.GetDesign(), "リードスートを持っているのに他スートが合法")
	}
}

// **エクスキューズはリードスートを決めない。** 次の札がスートを決める。
func TestTrogguLedSuit_ExcuseDoesNotSetTheSuit(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(FrenchTarotExcuseDesign, FrenchTarotExcuseValue, false)},
	}
	assert.Equal(t, -1, g.ledSuit(), "エクスキューズがスートを決めている")

	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: 1, Card: NewCard(2, 5, false)})
	assert.Equal(t, 2, g.ledSuit())
}

func TestTrogguValidPlays_MustTrumpWhenVoid(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	seat := g.GetCurrentPlayerIdx()
	next := (seat + 1) % TrogguPlayerCnt

	lead := g.GetPlayer(seat)
	lead.Reset()
	lead.AddCard(NewCard(1, 4, false))
	p := g.GetPlayer(next)
	p.Reset()
	p.AddCard(NewCard(FrenchTarotTrumpDesign, 5, false))
	p.AddCard(NewCard(2, 3, false))

	g.currentPlayerIdx = seat
	g.playCard(seat, 0)

	valid := g.GetValidPlayIndices(next)
	require.Len(t, valid, 1)
	assert.True(t, frenchTarotIsTrump(p.GetCard(valid[0])), "アトゥを出す義務が働いていない")
}

func TestTrogguPlayerPlayCard_Errors(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	seat := g.HumanSeat()
	g.currentPlayerIdx = seat

	assert.Error(t, g.PlayerPlayCard(-1))
	assert.Error(t, g.PlayerPlayCard(999))

	g.phase = TrogguPhaseBid
	assert.Error(t, g.PlayerPlayCard(0))
	g.phase = TrogguPhasePlay

	g.currentPlayerIdx = (seat + 1) % TrogguPlayerCnt
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrNotHumanTurn)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrGameEnded)
}

func TestTrogguTrick_WinnerTakesTheCards(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	for range TrogguPlayerCnt {
		seat := g.GetCurrentPlayerIdx()
		g.playCard(seat, g.GetValidPlayIndices(seat)[0])
	}
	assert.Equal(t, TrogguPhaseTrickEnd, g.GetPhase())
	assert.Len(t, g.GetLastTrickCards(), TrogguPlayerCnt)
	assertTrogguDeckIntact(t, g)
}

// --- 精算 ---

// 総ハーフポイントは表から写さず数える。
func TestTrogguTotalHalfPoints(t *testing.T) {
	total := trogguTotalHalfPoints()
	assert.Positive(t, total)
	assert.Equal(t, total, trogguTotalHalfPoints())
	// ブー 3 枚と 4 スートのロワ (各 9) が含まれる下限。
	assert.GreaterOrEqual(t, total, 7*9)
}

// **契約ごとに見るものが違う。** ソロだけが点数、他はトリック数。
func TestTrogguScoreDeal_ByContract(t *testing.T) {
	tests := []struct {
		name       string
		contract   TrogguBid
		giveDeck   bool // デクレアラーに全札を渡す
		tricks     int  // 渡すトリック数 (giveDeck=false のとき)
		wantWon    bool
		wantTricks bool
	}{
		{"ソロ: 全札を取れば成功", TrogguBidSolo, true, 0, true, false},
		{"ソロ: 1 枚も取らなければ失敗", TrogguBidSolo, false, 0, false, false},
		{"トロワ: 3 トリックで成功", TrogguBidTrois, false, 3, true, true},
		{"トロワ: 2 トリックでは失敗", TrogguBidTrois, false, 2, false, true},
		{"ピッコロ: ちょうど 1 で成功", TrogguBidPiccolo, false, 1, true, true},
		{"ピッコロ: 0 では失敗", TrogguBidPiccolo, false, 0, false, true},
		{"ピッコロ: 2 でも失敗", TrogguBidPiccolo, false, 2, false, true},
		{"ミゼール: 0 で成功", TrogguBidMisere, false, 0, true, true},
		{"ミゼール: 1 でも失敗", TrogguBidMisere, false, 1, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTrogguForTest(t)
			g.declarerIdx = 0
			g.contract = tt.contract
			if tt.giveDeck {
				g.GetPlayer(0).AddTrick(buildFrenchTarotDeck())
			}
			for range tt.tricks {
				g.GetPlayer(0).AddTrick([]*Card{NewCard(1, 1, false)})
			}

			bd := g.scoreDeal()
			assert.Equal(t, tt.wantWon, bd.Won)
			assert.Equal(t, tt.wantTricks, bd.TargetIsTricks)
			assert.Equal(t, TrogguBidName(tt.contract), bd.ContractName)
			// **精算はゼロサム。**
			sum := 0
			for _, v := range bd.Seats {
				sum += v
			}
			assert.Equal(t, 0, sum, "ゼロサムになっていない")
			// 単独契約なので、デクレアラーは 3 人ぶんを受け取るか払う。
			assert.Equal(t, 3*bd.Base, trogguAbs(bd.Seats[0]))
			if tt.wantWon {
				assert.Positive(t, bd.Seats[0])
				assert.Negative(t, bd.Seats[1])
				return
			}
			assert.Negative(t, bd.Seats[0])
			assert.Positive(t, bd.Seats[1])
		})
	}
}

// そろえにくい契約ほど倍率が高い。
// trogguAbs は絶対値を返す (domain には Bisley の abs が既にある)。
func trogguAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func TestTrogguBidMultiplier_OrderedByDifficulty(t *testing.T) {
	assert.Less(t, TrogguBidMultiplier(TrogguBidTrois), TrogguBidMultiplier(TrogguBidSolo))
	assert.Less(t, TrogguBidMultiplier(TrogguBidSolo), TrogguBidMultiplier(TrogguBidPiccolo))
	assert.Less(t, TrogguBidMultiplier(TrogguBidPiccolo), TrogguBidMultiplier(TrogguBidMisere))
	assert.Equal(t, 1, TrogguBidMultiplier(TrogguBidPass))
}

// **場札はデクレアラーの得点になる。** どこにも足さないと 6 枚ぶんの点が消える。
func TestTrogguFinishRound_TalonGoesToTheDeclarer(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	want := 0
	for _, c := range g.talon {
		want += frenchTarotCardHalfPoints(c)
	}
	require.Positive(t, want)
	before := g.GetCardPoints(0)

	g.finishRound()
	assert.Equal(t, 0, g.GetTalonSize(), "場札が残っている")
	assert.Equal(t, before+want, g.GetCardPoints(0), "場札の点が加わっていない")
}

// --- 通し ---

func TestTrogguFullMatch_Terminates(t *testing.T) {
	g := NewTroggu(trogguTablePlayers(), TrogguConfig{TargetDeals: 2})
	g.Reset()
	trogguPlayOut(t, g)

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, TrogguPhaseGameEnd, g.GetPhase())
	assert.NotEmpty(t, g.GetActionLog())
	// 精算はディールごとにゼロサムなので通算も 0。
	sum := 0
	for i := range TrogguPlayerCnt {
		sum += g.GetPlayerScore(i)
	}
	assert.Equal(t, 0, sum, "通算得点がゼロサムでない")
}

func TestTrogguFullMatch_TerminatesRepeatedly(t *testing.T) {
	for range 20 {
		g := NewTroggu(trogguTablePlayers(), TrogguConfig{TargetDeals: 1})
		g.Reset()
		trogguPlayOut(t, g)
		require.True(t, g.GetGameEndFlag())
	}
}

// 契約が付いたディールは 78 枚すべてが席に収まる。
func TestTrogguRoundEnd_AllCardsAccountedFor(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	for range 5000 {
		if g.GetPhase() == TrogguPhaseRoundEnd || g.GetGameEndFlag() {
			break
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NotNil(t, h)
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
			continue
		}
		trogguStep(g)
	}
	require.NotEqual(t, TrogguPhasePlay, g.GetPhase())

	taken := 0
	for _, p := range g.GetPlayers() {
		assert.Equal(t, 0, p.GetCardsSize(), "手札が残っている")
		for _, tr := range p.GetTricksTaken() {
			taken += len(tr)
		}
	}
	assert.Equal(t, TrogguDeckSize, taken, "78 枚が席に収まっていない")
}

// --- ヒント ---

func TestTrogguGetHint_ByPhase(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidMisere)
	g.currentPlayerIdx = g.HumanSeat()
	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	// ミゼールのデクレアラーは取りたくないので、伏せる助言になる。
	assert.Equal(t, "play_duck", h.Reason)
	assert.NoError(t, g.PlayerPlayCard(*h.CardIndex), "ヒントの手が通らない")

	solo := trogguAtPlay(t, TrogguBidSolo)
	solo.currentPlayerIdx = solo.HumanSeat()
	h = solo.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "play_win", h.Reason)
}

func TestTrogguGetHint_NilOutsideHumanTurns(t *testing.T) {
	g := newTrogguForTest(t)
	g.gameEndFlag = true
	assert.Nil(t, g.GetHint())

	g.gameEndFlag = false
	g.phase = TrogguPhaseRoundEnd
	assert.Nil(t, g.GetHint())
}

// --- 永続化 ---

func TestTrogguRoundTripJSON(t *testing.T) {
	g := trogguAtPlay(t, TrogguBidSolo)
	for range 8 {
		if g.GetPhase() == TrogguPhaseTrickEnd {
			g.NextTrick()
			continue
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
			continue
		}
		g.CpuPlayCard()
	}

	b, err := json.Marshal(g)
	require.NoError(t, err)
	var back Troggu
	require.NoError(t, json.Unmarshal(b, &back))

	assert.Equal(t, g.GetPhase(), back.GetPhase())
	assert.Equal(t, g.GetContract(), back.GetContract())
	assert.Equal(t, g.GetDeclarerIdx(), back.GetDeclarerIdx())
	assert.Equal(t, g.GetTrickNumber(), back.GetTrickNumber())
	assertTrogguDeckIntact(t, &back)
	trogguPlayOut(t, &back)
}

// 改竄した保存データは、本物の局面を 1 か所だけ壊して作る。
func trogguTamper(t *testing.T, mutate func(m map[string]any)) error {
	t.Helper()
	g := trogguAtPlay(t, TrogguBidSolo)
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	var back Troggu
	return json.Unmarshal(tampered, &back)
}

func TestTrogguUnmarshal_RejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 9 }, "invalid phase"},
		{"手番が範囲外", func(m map[string]any) { m["cp"] = 4 }, "current player out of range"},
		{"親が範囲外", func(m map[string]any) { m["di"] = -1 }, "dealer out of range"},
		{"デクレアラーが範囲外", func(m map[string]any) { m["dc"] = 9 }, "declarer out of range"},
		{"ラウンドが 0", func(m map[string]any) { m["rn"] = 0 }, "round 0 out of range"},
		{"トリックが多すぎる", func(m map[string]any) { m["tn"] = TrogguTrickCount + 1 }, "trick"},
		{"契約が不正", func(m map[string]any) { m["co"] = 9 }, "invalid contract"},
		// **範囲チェックだけでは通ってしまう組み合わせ。**
		{"契約があるのにデクレアラーが居ない", func(m map[string]any) { m["dc"] = -1 }, "requires a declarer"},
		{"デクレアラーが居るのに契約が無い", func(m map[string]any) {
			m["co"] = int(TrogguBidPass)
		}, "requires a contract"},
		{"設定が不正", func(m map[string]any) {
			m["cf"] = map[string]any{"cd": 0, "td": 0}
		}, "target deals"},
		{"席数が違う", func(m map[string]any) { m["pl"] = m["pl"].([]any)[:3] }, "expected 4 players"},
		{"席が nil", func(m map[string]any) {
			pl := m["pl"].([]any)
			pl[1] = nil
			m["pl"] = pl
		}, "nil player"},
		{"場札の札が範囲外", func(m map[string]any) {
			m["tl"] = []any{map[string]any{"d": 9, "v": 1}}
		}, "card out of range"},
		{"場札が nil", func(m map[string]any) { m["tl"] = []any{nil} }, "nil card"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := trogguTamper(t, tt.mutate)
			require.Error(t, err, "改竄が素通りしている")
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestTrogguUnmarshal_RejectsOversizedArrays(t *testing.T) {
	big := make([]map[string]int, trogguMaxSliceLen+1)
	for i := range big {
		big[i] = map[string]int{"d": 1, "v": 1}
	}
	payload, err := json.Marshal(map[string]any{"tl": big})
	require.NoError(t, err)
	var g Troggu
	err = json.Unmarshal(payload, &g)
	require.Error(t, err)
	assert.ErrorContains(t, err, "maximum allowed size")
}

func TestTrogguUnmarshal_RejectsMalformedJSON(t *testing.T) {
	var g Troggu
	assert.Error(t, json.Unmarshal([]byte(`{"ph":"x"}`), &g))
}

func TestTrogguPlayer_RoundTripJSON(t *testing.T) {
	p := NewTrogguPlayer(true)
	p.AddCard(NewCard(FrenchTarotTrumpDesign, 21, false))
	p.AddTrick([]*Card{NewCard(1, FrenchTarotKingValue, false)})

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var back TrogguPlayer
	require.NoError(t, json.Unmarshal(b, &back))
	assert.True(t, back.GetIsHuman())
	assert.Equal(t, 1, back.GetCardsSize())
	assert.Equal(t, 1, back.GetTrickCount())
}

func TestTrogguBidName(t *testing.T) {
	assert.Equal(t, "trois", TrogguBidName(TrogguBidTrois))
	assert.Equal(t, "solo", TrogguBidName(TrogguBidSolo))
	assert.Equal(t, "piccolo", TrogguBidName(TrogguBidPiccolo))
	assert.Equal(t, "misere", TrogguBidName(TrogguBidMisere))
	assert.Equal(t, "pass", TrogguBidName(TrogguBidPass))
}
