//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newZwanzigerrufenForTest はゲームを開始した状態で返す。
func newZwanzigerrufenForTest(t *testing.T) *Zwanzigerrufen {
	t.Helper()
	g := NewDefaultZwanzigerrufen()
	g.Reset()
	return g
}

// zwanzigerrufenDrive は指定フェーズに達するか終局するまで CPU を進める。
func zwanzigerrufenDrive(t *testing.T, g *Zwanzigerrufen) {
	t.Helper()
	for range 500 {
		if g.GetGameEndFlag() || g.IsHumanTurn() {
			return
		}
		switch g.GetPhase() {
		case ZwanzigerrufenPhaseBid:
			g.CpuBid()
		case ZwanzigerrufenPhaseTalon:
			g.CpuDiscard()
		case ZwanzigerrufenPhasePlay:
			g.CpuPlayCard()
		case ZwanzigerrufenPhaseTrickEnd:
			g.NextTrick()
		case ZwanzigerrufenPhaseRoundEnd:
			g.NextRound()
		default:
			return
		}
	}
	t.Fatal("進行が止まった")
}

// zwanzigerrufenPlayOut はマッチが終わるまで進める (人間は推奨手に従う)。
func zwanzigerrufenPlayOut(t *testing.T, g *Zwanzigerrufen) {
	t.Helper()
	for range 5000 {
		if g.GetGameEndFlag() {
			return
		}
		if !g.IsHumanTurn() {
			zwanzigerrufenStepCpu(g)
			continue
		}
		switch g.GetPhase() {
		case ZwanzigerrufenPhaseBid:
			require.NoError(t, g.PlayerPass())
		case ZwanzigerrufenPhaseTalon:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NoError(t, g.PlayerDiscard(h.DiscardIndices))
		case ZwanzigerrufenPhasePlay:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NotNil(t, h.CardIndex)
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
		default:
			zwanzigerrufenStepCpu(g)
		}
	}
	t.Fatal("終局しなかった")
}

// zwanzigerrufenStepCpu は CPU 側 / 自動進行を 1 歩進める。
func zwanzigerrufenStepCpu(g *Zwanzigerrufen) {
	switch g.GetPhase() {
	case ZwanzigerrufenPhaseBid:
		g.CpuBid()
	case ZwanzigerrufenPhaseTalon:
		g.CpuDiscard()
	case ZwanzigerrufenPhasePlay:
		g.CpuPlayCard()
	case ZwanzigerrufenPhaseTrickEnd:
		g.NextTrick()
	case ZwanzigerrufenPhaseRoundEnd:
		g.NextRound()
	}
}

// zwanzigerrufenCardsInPlay は盤上のすべての札を集める。
func zwanzigerrufenCardsInPlay(g *Zwanzigerrufen) []*Card {
	out := make([]*Card, 0, ZwanzigerrufenDeckSize)
	out = append(out, g.talon...)
	out = append(out, g.stash...)
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

// assertZwanzigerrufenDeckIntact は 54 枚が 1 枚ずつ揃っていることを確かめる。
func assertZwanzigerrufenDeckIntact(t *testing.T, g *Zwanzigerrufen) {
	t.Helper()
	cards := zwanzigerrufenCardsInPlay(g)
	assert.Len(t, cards, ZwanzigerrufenDeckSize, "盤上の総枚数")
	seen := map[[2]int]bool{}
	for _, c := range cards {
		require.NotNil(t, c)
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "札 %v が重複", key)
		seen[key] = true
	}
}

// --- 設定 ---

func TestZwanzigerrufenConfig_Validate(t *testing.T) {
	tests := []struct {
		name string
		cfg  ZwanzigerrufenConfig
		ok   bool
	}{
		{"既定", DefaultZwanzigerrufenConfig(), true},
		{"最小ディール", ZwanzigerrufenConfig{TargetDeals: ZwanzigerrufenMinDeals}, true},
		{"最大ディール", ZwanzigerrufenConfig{TargetDeals: ZwanzigerrufenMaxDeals}, true},
		{"ディール 0", ZwanzigerrufenConfig{TargetDeals: 0}, false},
		{"ディール超過", ZwanzigerrufenConfig{TargetDeals: ZwanzigerrufenMaxDeals + 1}, false},
		{"難易度が範囲外", ZwanzigerrufenConfig{CpuDifficulty: 9, TargetDeals: 4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

// --- 配り ---

func TestZwanzigerrufenReset_DealsTwelveEachAndSixTalon(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	for i, p := range g.GetPlayers() {
		assert.Equal(t, ZwanzigerrufenHandSize, p.GetCardsSize(), "席 %d の手札", i)
	}
	assert.Equal(t, ZwanzigerrufenTalonSize, g.GetTalonSize())
	assert.Equal(t, ZwanzigerrufenPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assert.Equal(t, -1, g.GetCalledTrump())
	assertZwanzigerrufenDeckIntact(t, g)
}

// デッキ 54 枚の中身をタロックの構成そのままで数える。
func TestZwanzigerrufenDeckComposition(t *testing.T) {
	suits, trumps, skus := 0, 0, 0
	for _, c := range buildKoenigrufenDeck() {
		switch {
		case koenigrufenIsSkus(c):
			skus++
		case koenigrufenIsTrump(c):
			trumps++
		default:
			suits++
		}
	}
	assert.Equal(t, 32, suits, "スート札")
	assert.Equal(t, 21, trumps, "切り札")
	assert.Equal(t, 1, skus, "スキュース")
	assert.Equal(t, ZwanzigerrufenDeckSize, suits+trumps+skus)
	// 配り切ると場札が残らない構成になっていないこと。
	assert.Equal(t, ZwanzigerrufenDeckSize,
		ZwanzigerrufenPlayerCnt*ZwanzigerrufenHandSize+ZwanzigerrufenTalonSize)
}

// --- 入札 ---

func TestZwanzigerrufenBid_RejectsUndeclarableAndLowBids(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	for !g.IsHumanTurn() {
		g.CpuBid()
		if g.GetPhase() != ZwanzigerrufenPhaseBid {
			t.Skip("この配りでは人間の入札手番が来なかった")
		}
	}
	// **Trischaken は宣言できない。** 全員パスの結果としてしか成立しない。
	err := g.PlayerBid(ZwanzigerrufenBidTrischaken)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot be declared")
	assert.Error(t, g.PlayerBid(ZwanzigerrufenBidPass))

	// **上回らない入札は通らない。** CPU が既に Rufer を宣言していれば Solo しか残らない。
	if g.GetHighestBid() >= ZwanzigerrufenBidRufer {
		assert.Error(t, g.PlayerBid(ZwanzigerrufenBidRufer), "同額の入札が通っている")
		require.NoError(t, g.PlayerBid(ZwanzigerrufenBidSolo))
		assert.Equal(t, ZwanzigerrufenBidSolo, g.GetHighestBid())
		return
	}
	require.NoError(t, g.PlayerBid(ZwanzigerrufenBidRufer))
	assert.Equal(t, ZwanzigerrufenBidRufer, g.GetHighestBid())
}

func TestZwanzigerrufenBid_Errors(t *testing.T) {
	t.Run("プレイフェーズでは入札できない", func(t *testing.T) {
		g := newZwanzigerrufenForTest(t)
		g.phase = ZwanzigerrufenPhasePlay
		assert.Error(t, g.PlayerBid(ZwanzigerrufenBidRufer))
		assert.Error(t, g.PlayerPass())
	})
	t.Run("終局後", func(t *testing.T) {
		g := newZwanzigerrufenForTest(t)
		g.gameEndFlag = true
		assert.ErrorIs(t, g.PlayerBid(ZwanzigerrufenBidRufer), ErrGameEnded)
		assert.ErrorIs(t, g.PlayerPass(), ErrGameEnded)
	})
	t.Run("CPU の入札手番", func(t *testing.T) {
		g := newZwanzigerrufenForTest(t)
		g.bidPlayerIdx = 1
		assert.ErrorIs(t, g.PlayerBid(ZwanzigerrufenBidRufer), ErrNotHumanTurn)
		assert.ErrorIs(t, g.PlayerPass(), ErrNotHumanTurn)
	})
}

// **全員パスなら Trischaken。** デクレアラーも場札交換も無い。
func TestZwanzigerrufenAllPass_BecomesTrischaken(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	for range ZwanzigerrufenPlayerCnt {
		if g.GetPhase() != ZwanzigerrufenPhaseBid {
			break
		}
		g.applyPass(g.GetBidPlayerIdx())
	}
	assert.Equal(t, ZwanzigerrufenBidTrischaken, g.GetContract())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assert.Equal(t, -1, g.GetCalledTrump())
	assert.Equal(t, ZwanzigerrufenPhasePlay, g.GetPhase(), "場札交換を挟んでいる")
	// **場札は脇へ移る。** 手札に加えないが、置き去りにもしない ── 最終トリックの
	// 勝者が引き取るので、ここで移し忘れると 6 枚ぶんの点が消える。
	assert.Equal(t, 0, g.GetTalonSize(), "場札が場に残っている")
	assert.Len(t, g.stash, ZwanzigerrufenTalonSize, "場札が脇へ移っていない")
	assertZwanzigerrufenDeckIntact(t, g)
}

// **どの契約でも、ディールが終われば 54 枚すべてが席に収まる。**
//
// 脇に置いた札を引き取らせ忘れると総点が合わなくなる ── 手で stash を積むテスト
// では finalizeBid の移し忘れを踏めないので、本物の入札から通す。
func TestZwanzigerrufenTrischaken_TalonReachesASeat(t *testing.T) {
	g := NewZwanzigerrufen(zwanzigerrufenPlayers(), ZwanzigerrufenConfig{TargetDeals: 1})
	g.Reset()
	for range ZwanzigerrufenPlayerCnt {
		if g.GetPhase() != ZwanzigerrufenPhaseBid {
			break
		}
		g.applyPass(g.GetBidPlayerIdx())
	}
	require.Equal(t, ZwanzigerrufenBidTrischaken, g.GetContract())

	zwanzigerrufenPlayOut(t, g)
	taken := 0
	for _, p := range g.GetPlayers() {
		for _, tr := range p.GetTricksTaken() {
			taken += len(tr)
		}
	}
	assert.Equal(t, ZwanzigerrufenDeckSize, taken, "54 枚が席に収まっていない")
	// 点も消えていない。
	total := 0
	for i := range g.GetPlayers() {
		total += g.GetCardPoints(i)
	}
	assert.Equal(t, zwanzigerrufenTotalPoints(), total, "カードポイントが消えている")
}

// --- 呼び札 ---

// 呼ぶのは切り札の 20 番で、その持ち主が秘密のパートナーになる。
func TestZwanzigerrufenCall_CallsTrumpTwenty(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	// 20 番を席 0 以外へ移し、席 0 をデクレアラーにする。
	zwanzigerrufenGiveTrump(t, g, ZwanzigerrufenCallTrump, 2)
	g.declarerIdx = 0
	g.resolveCall()

	assert.Equal(t, ZwanzigerrufenCallTrump, g.GetCalledTrump())
	assert.Equal(t, 2, g.partnerIdx, "20 番の持ち主がパートナー")
	// **明かすまでは答えない。**
	assert.Equal(t, -1, g.GetPartnerIdx())
	assert.False(t, g.GetPartnerRevealed())
}

// **自分が持っている札は呼べない。** 20 番を抱えていたら 19 番へ下げる。
func TestZwanzigerrufenCall_StepsDownWhenTheDeclarerHoldsIt(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	zwanzigerrufenGiveTrump(t, g, ZwanzigerrufenCallTrump, 0)
	zwanzigerrufenGiveTrump(t, g, ZwanzigerrufenCallTrump-1, 3)
	g.declarerIdx = 0
	g.resolveCall()

	assert.Equal(t, ZwanzigerrufenCallTrump-1, g.GetCalledTrump(), "19 番へ下がっていない")
	assert.Equal(t, 3, g.partnerIdx)
}

// 呼べる札を全部抱えていたら単独で戦う。
func TestZwanzigerrufenCall_PlaysAloneWhenEveryCallableTrumpIsHeld(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	for tr := ZwanzigerrufenMinCallTrump; tr <= ZwanzigerrufenCallTrump; tr++ {
		zwanzigerrufenGiveTrump(t, g, tr, 0)
	}
	g.declarerIdx = 0
	g.resolveCall()

	assert.Equal(t, -1, g.GetCalledTrump())
	assert.Equal(t, -1, g.partnerIdx, "自分自身がパートナーになっている")
}

// zwanzigerrufenGiveTrump は指定番号の切り札を席 seat の手札へ移す。
func zwanzigerrufenGiveTrump(t *testing.T, g *Zwanzigerrufen, trump, seat int) {
	t.Helper()
	for i, p := range g.GetPlayers() {
		for k := range p.GetCardsSize() {
			c := p.GetCard(k)
			if koenigrufenIsTrump(c) && c.GetValue() == trump {
				if i == seat {
					return
				}
				moved := p.RemoveCard(k)
				g.GetPlayers()[seat].AddCard(moved)
				return
			}
		}
	}
	// 場札にあるなら、そこから移す。
	for k, c := range g.talon {
		if koenigrufenIsTrump(c) && c.GetValue() == trump {
			g.talon = append(g.talon[:k], g.talon[k+1:]...)
			g.GetPlayers()[seat].AddCard(c)
			return
		}
	}
	t.Fatalf("切り札 %d が見つからない", trump)
}

// パートナーは呼び札が場に出た瞬間に判明する。
func TestZwanzigerrufenPartnerRevealsWhenTheCalledTrumpIsPlayed(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	zwanzigerrufenGiveTrump(t, g, ZwanzigerrufenCallTrump, 1)
	g.declarerIdx = 0
	g.contract = ZwanzigerrufenBidRufer
	g.resolveCall()
	require.Equal(t, 1, g.partnerIdx)
	g.startPlay()

	// 席 1 に呼び札を出させる。
	p := g.GetPlayers()[1]
	idx := -1
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if koenigrufenIsTrump(c) && c.GetValue() == ZwanzigerrufenCallTrump {
			idx = i
		}
	}
	require.GreaterOrEqual(t, idx, 0)
	g.currentPlayerIdx = 1
	g.playCard(1, idx)

	assert.True(t, g.GetPartnerRevealed())
	assert.Equal(t, 1, g.GetPartnerIdx())
}

// --- 場札交換 ---

// **キングとトゥルルは伏せられない。** 5 点札を黙って自分の得点へ移せると、
// 場札交換が「点を配る操作」になる。
func TestZwanzigerrufenDiscard_RejectsKingsAndTrull(t *testing.T) {
	tests := []struct {
		name string
		pick func(c *Card) bool
	}{
		{"キング", koenigrufenIsKing},
		{"トゥルル", koenigrufenIsTrull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := zwanzigerrufenAtTalon(t)
			p := g.GetPlayer(g.GetDeclarerIdx())
			target := -1
			for i := range p.GetCardsSize() {
				if tt.pick(p.GetCard(i)) {
					target = i
					break
				}
			}
			if target < 0 {
				t.Skipf("この配りでは %s が手札に無い", tt.name)
			}
			// target を含む、重複しないちょうど 6 枚を作る。
			indices := []int{target}
			for i := 0; i < p.GetCardsSize() && len(indices) < ZwanzigerrufenTalonSize; i++ {
				if i != target {
					indices = append(indices, i)
				}
			}
			require.Len(t, indices, ZwanzigerrufenTalonSize)

			err := g.PlayerDiscard(indices)
			require.Error(t, err, "%s を伏せられてしまった", tt.name)
			// 6 枚のうち別の禁止札が先に当たることもあるので、どちらかであればよい。
			assert.Contains(t,
				[]string{"a king cannot be discarded", "a trull card cannot be discarded"},
				err.(*DomainError).Message)
		})
	}
}

func TestZwanzigerrufenDiscard_RejectsWrongCountAndDuplicates(t *testing.T) {
	g := zwanzigerrufenAtTalon(t)
	err := g.PlayerDiscard([]int{0, 1, 2})
	require.Error(t, err)
	assert.ErrorContains(t, err, "exactly")

	err = g.PlayerDiscard([]int{0, 0, 1, 2, 3, 4})
	require.Error(t, err)
	assert.ErrorContains(t, err, "twice")

	err = g.PlayerDiscard([]int{0, 1, 2, 3, 4, 999})
	require.Error(t, err)
	assert.ErrorContains(t, err, "out of range")
}

// 伏せた 6 枚は手札から抜け、デクレアラーの得点として脇に残る。
func TestZwanzigerrufenDiscard_BuriesSixCards(t *testing.T) {
	g := zwanzigerrufenAtTalon(t)
	h := g.GetHint()
	require.NotNil(t, h)
	require.Len(t, h.DiscardIndices, ZwanzigerrufenTalonSize)

	require.NoError(t, g.PlayerDiscard(h.DiscardIndices))
	assert.Equal(t, ZwanzigerrufenHandSize, g.GetPlayers()[g.GetDeclarerIdx()].GetCardsSize())
	assert.Equal(t, ZwanzigerrufenPhasePlay, g.GetPhase())
	assertZwanzigerrufenDeckIntact(t, g)
}

// zwanzigerrufenAtTalon は人間がデクレアラーの場札交換フェーズを作る。
func zwanzigerrufenAtTalon(t *testing.T) *Zwanzigerrufen {
	t.Helper()
	g := newZwanzigerrufenForTest(t)
	g.highestBid = ZwanzigerrufenBidRufer
	g.highestBidder = 0
	g.finalizeBid()
	require.Equal(t, ZwanzigerrufenPhaseTalon, g.GetPhase())
	require.Equal(t, 0, g.GetDeclarerIdx())
	return g
}

// --- プレイ ---

// リードスートに従う義務がある。
func TestZwanzigerrufenValidPlays_MustFollowTheLead(t *testing.T) {
	g := zwanzigerrufenAtPlay(t)
	lead := g.GetCurrentPlayerIdx()
	// 先頭は何でも出せる。
	assert.Len(t, g.GetValidPlayIndices(lead), g.GetPlayers()[lead].GetCardsSize())

	g.playCard(lead, 0)
	next := g.GetCurrentPlayerIdx()
	led := g.ledSuit()
	valid := g.GetValidPlayIndices(next)
	p := g.GetPlayers()[next]
	hasLed := false
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if (led == KoenigrufenTrumpDesign && koenigrufenIsTrumpLike(c)) ||
			(led != KoenigrufenTrumpDesign && !koenigrufenIsTrumpLike(c) && c.GetDesign() == led) {
			hasLed = true
		}
	}
	if hasLed {
		for _, i := range valid {
			c := p.GetCard(i)
			if led == KoenigrufenTrumpDesign {
				assert.True(t, koenigrufenIsTrumpLike(c), "リードが切り札なのに他スートが合法")
				continue
			}
			assert.Equal(t, led, c.GetDesign(), "リードスートを持っているのに他スートが合法")
		}
	}
}

// リードスートが無ければ切り札を切る義務がある。
func TestZwanzigerrufenValidPlays_MustTrumpWhenVoid(t *testing.T) {
	g := zwanzigerrufenAtPlay(t)
	seat := g.GetCurrentPlayerIdx()
	next := (seat + 1) % ZwanzigerrufenPlayerCnt
	// 席 next の手札を「1 スートを持たない + 切り札あり」に作り替える。
	p := g.GetPlayers()[next]
	p.Reset()
	p.AddCard(NewCard(KoenigrufenTrumpDesign, 5, false))
	p.AddCard(NewCard(2, 3, false))

	// スート 1 をリードする。
	lead := g.GetPlayers()[seat]
	lead.Reset()
	lead.AddCard(NewCard(1, 4, false))
	g.currentPlayerIdx = seat
	g.playCard(seat, 0)

	valid := g.GetValidPlayIndices(next)
	require.Len(t, valid, 1)
	assert.True(t, koenigrufenIsTrumpLike(p.GetCard(valid[0])), "切り札を切る義務が働いていない")
}

// 従えない札は弾く。
func TestZwanzigerrufenPlayerPlayCard_Errors(t *testing.T) {
	g := zwanzigerrufenAtPlay(t)
	seat := g.HumanSeat()
	g.currentPlayerIdx = seat

	assert.Error(t, g.PlayerPlayCard(-1))
	assert.Error(t, g.PlayerPlayCard(99))

	g.phase = ZwanzigerrufenPhaseBid
	assert.Error(t, g.PlayerPlayCard(0))
	g.phase = ZwanzigerrufenPhasePlay

	g.currentPlayerIdx = (seat + 1) % ZwanzigerrufenPlayerCnt
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrNotHumanTurn)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrGameEnded)
}

// zwanzigerrufenAtPlay は Rufer 契約でプレイフェーズに入った局面を返す。
func zwanzigerrufenAtPlay(t *testing.T) *Zwanzigerrufen {
	t.Helper()
	g := zwanzigerrufenAtTalon(t)
	h := g.GetHint()
	require.NotNil(t, h)
	require.NoError(t, g.PlayerDiscard(h.DiscardIndices))
	require.Equal(t, ZwanzigerrufenPhasePlay, g.GetPhase())
	return g
}

// トリックは 4 枚で決着し、勝者が引き取る。
func TestZwanzigerrufenTrick_WinnerTakesTheCards(t *testing.T) {
	g := zwanzigerrufenAtPlay(t)
	before := 0
	for _, p := range g.GetPlayers() {
		before += p.GetTrickCount()
	}
	for range ZwanzigerrufenPlayerCnt {
		g.playCard(g.GetCurrentPlayerIdx(), g.GetValidPlayIndices(g.GetCurrentPlayerIdx())[0])
	}
	assert.Equal(t, ZwanzigerrufenPhaseTrickEnd, g.GetPhase())
	after := 0
	for _, p := range g.GetPlayers() {
		after += p.GetTrickCount()
	}
	assert.Equal(t, before+1, after)
	assert.Len(t, g.GetLastTrickCards(), ZwanzigerrufenPlayerCnt)
	assertZwanzigerrufenDeckIntact(t, g)
}

// --- 精算 ---

// 総カードポイントは表から写さず数える。
func TestZwanzigerrufenTotalPoints(t *testing.T) {
	total := zwanzigerrufenTotalPoints()
	assert.Positive(t, total)
	// 3 枚のトゥルル (5 点) と 4 人のキング (5 点) が含まれる下限。
	assert.GreaterOrEqual(t, total, 3*5+4*5)
	// もう一度数えても同じ (デッキ構築が状態を持っていない)。
	assert.Equal(t, total, zwanzigerrufenTotalPoints())
}

// **場札はどこかの席に必ず加わる。** 消えると総点が合わなくなる。
func TestZwanzigerrufenStashIsAlwaysCounted(t *testing.T) {
	tests := []struct {
		name  string
		setup func(g *Zwanzigerrufen)
	}{
		{"Rufer は伏せ札がデクレアラーの点", func(g *Zwanzigerrufen) {
			g.declarerIdx = 0
			g.contract = ZwanzigerrufenBidRufer
			g.stashOwner = 0
		}},
		{"Solo は場札が防御側の点", func(g *Zwanzigerrufen) {
			g.declarerIdx = 0
			g.contract = ZwanzigerrufenBidSolo
			g.stashOwner = 1
		}},
		{"Trischaken は最終トリックの勝者へ", func(g *Zwanzigerrufen) {
			g.declarerIdx = -1
			g.contract = ZwanzigerrufenBidTrischaken
			g.stashOwner = -1
			g.lastTrickWinner = 2
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newZwanzigerrufenForTest(t)
			tt.setup(g)
			g.stash = append([]*Card(nil), g.talon...)
			g.talon = nil
			want := 0
			for _, c := range g.stash {
				want += koenigrufenCardPoints(c)
			}
			require.Positive(t, want)

			g.assignStash()
			assert.Empty(t, g.stash, "脇に残っている")
			got := 0
			for i := range g.GetPlayers() {
				got += g.GetCardPoints(i)
			}
			assert.Equal(t, want, got, "場札の点がどこにも加わっていない")
		})
	}
}

// **Trischaken は最も多く取った席が負ける。** 勝ち負けの向きが逆。
func TestZwanzigerrufenScoreTrischaken_MostPointsLoses(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	g.contract = ZwanzigerrufenBidTrischaken
	g.declarerIdx = -1
	// 席 2 にだけ高い札を積む。
	g.GetPlayers()[2].AddTrick([]*Card{NewCard(1, KoenigrufenKingValue, false)})
	g.GetPlayers()[0].AddTrick([]*Card{NewCard(1, 1, false)})

	bd := g.scoreTrischaken()
	assert.Equal(t, 2, bd.Loser)
	assert.Equal(t, -ZwanzigerrufenTrischakenLoss, bd.Seats[2])
	sum := 0
	for _, v := range bd.Seats {
		sum += v
	}
	assert.Equal(t, 0, sum, "ゼロサムになっていない")
}

// 契約の精算はゼロサムで、単独は 3 人ぶんを背負う。
func TestZwanzigerrufenScoreContract_ZeroSum(t *testing.T) {
	tests := []struct {
		name     string
		partner  int
		contract ZwanzigerrufenBid
	}{
		{"組 (2 対 2)", 2, ZwanzigerrufenBidRufer},
		{"単独 (1 対 3)", -1, ZwanzigerrufenBidSolo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newZwanzigerrufenForTest(t)
			g.declarerIdx = 0
			g.partnerIdx = tt.partner
			g.contract = tt.contract
			// デッキの全札をデクレアラーへ (必ず達成する側)。
			g.GetPlayers()[0].AddTrick(buildKoenigrufenDeck())

			bd := g.scoreContract()
			assert.True(t, bd.Won, "全札を取って失敗している")
			sum := 0
			for _, v := range bd.Seats {
				sum += v
			}
			assert.Equal(t, 0, sum, "ゼロサムになっていない")
			assert.Positive(t, bd.Seats[0])
			if tt.partner >= 0 {
				assert.Positive(t, bd.Seats[tt.partner])
				assert.Equal(t, bd.Seats[0], bd.Seats[tt.partner])
			} else {
				assert.Equal(t, 3*bd.Base, bd.Seats[0], "単独が 3 人ぶんを受け取っていない")
			}
		})
	}
}

// 1 点も取れなければ失敗し、符号が反転する。
func TestZwanzigerrufenScoreContract_LossFlipsTheSigns(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	g.declarerIdx = 0
	g.partnerIdx = 2
	g.contract = ZwanzigerrufenBidRufer
	g.GetPlayers()[1].AddTrick(buildKoenigrufenDeck())

	bd := g.scoreContract()
	assert.False(t, bd.Won)
	assert.Negative(t, bd.Seats[0])
	assert.Negative(t, bd.Seats[2])
	assert.Positive(t, bd.Seats[1])
	assert.Positive(t, bd.Seats[3])
}

// --- 通し ---

func TestZwanzigerrufenFullMatch_Terminates(t *testing.T) {
	g := NewZwanzigerrufen(zwanzigerrufenPlayers(), ZwanzigerrufenConfig{TargetDeals: 2})
	g.Reset()
	zwanzigerrufenPlayOut(t, g)

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, ZwanzigerrufenPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.NotEmpty(t, g.GetActionLog())
	// 精算はディールごとにゼロサムなので、通算も 0。
	sum := 0
	for i := range ZwanzigerrufenPlayerCnt {
		sum += g.GetPlayerScore(i)
	}
	assert.Equal(t, 0, sum, "通算得点がゼロサムでない")
}

func TestZwanzigerrufenFullMatch_TerminatesRepeatedly(t *testing.T) {
	for range 20 {
		g := NewZwanzigerrufen(zwanzigerrufenPlayers(), ZwanzigerrufenConfig{TargetDeals: 1})
		g.Reset()
		zwanzigerrufenPlayOut(t, g)
		require.True(t, g.GetGameEndFlag())
		require.NotNil(t, g.GetBreakdown())
	}
}

// 全席が打ち切ると、12 トリック × 4 枚がどこかの席に収まる。
func TestZwanzigerrufenRoundEnd_AllCardsAccountedFor(t *testing.T) {
	g := NewZwanzigerrufen(zwanzigerrufenPlayers(), ZwanzigerrufenConfig{TargetDeals: 1})
	g.Reset()
	zwanzigerrufenPlayOut(t, g)

	taken := 0
	for _, p := range g.GetPlayers() {
		assert.Equal(t, 0, p.GetCardsSize(), "手札が残っている")
		for _, tr := range p.GetTricksTaken() {
			taken += len(tr)
		}
	}
	assert.Equal(t, ZwanzigerrufenDeckSize, taken, "54 枚が席に収まっていない")
}

func zwanzigerrufenPlayers() []*ZwanzigerrufenPlayer {
	ps := make([]*ZwanzigerrufenPlayer, ZwanzigerrufenPlayerCnt)
	ps[0] = NewZwanzigerrufenPlayer(true)
	for i := 1; i < ZwanzigerrufenPlayerCnt; i++ {
		ps[i] = NewZwanzigerrufenPlayer(false)
	}
	return ps
}

// --- ヒント ---

func TestZwanzigerrufenGetHint_ByPhase(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	zwanzigerrufenDrive(t, g)
	if g.GetPhase() == ZwanzigerrufenPhaseBid && g.IsHumanTurn() {
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Contains(t, []string{"pass_weak_hand", "bid_strong_trumps"}, h.Reason)
	}

	talon := zwanzigerrufenAtTalon(t)
	h := talon.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "bury_cheap_cards", h.Reason)
	assert.Len(t, h.DiscardIndices, ZwanzigerrufenTalonSize)

	play := zwanzigerrufenAtPlay(t)
	play.currentPlayerIdx = play.HumanSeat()
	h = play.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.NoError(t, play.PlayerPlayCard(*h.CardIndex), "ヒントの手が通らない")
}

func TestZwanzigerrufenGetHint_NilOutsideHumanTurns(t *testing.T) {
	g := newZwanzigerrufenForTest(t)
	g.gameEndFlag = true
	assert.Nil(t, g.GetHint())

	g.gameEndFlag = false
	g.phase = ZwanzigerrufenPhaseRoundEnd
	assert.Nil(t, g.GetHint())
}

// --- 永続化 ---

func TestZwanzigerrufenRoundTripJSON(t *testing.T) {
	g := zwanzigerrufenAtPlay(t)
	for range 6 {
		if g.GetPhase() == ZwanzigerrufenPhaseTrickEnd {
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
	var back Zwanzigerrufen
	require.NoError(t, json.Unmarshal(b, &back))

	assert.Equal(t, g.GetPhase(), back.GetPhase())
	assert.Equal(t, g.GetContract(), back.GetContract())
	assert.Equal(t, g.GetCalledTrump(), back.GetCalledTrump())
	assert.Equal(t, g.GetDeclarerIdx(), back.GetDeclarerIdx())
	assert.Equal(t, g.GetTrickNumber(), back.GetTrickNumber())
	assert.Equal(t, g.partnerIdx, back.partnerIdx, "秘密のパートナーが保存されていない")
	assertZwanzigerrufenDeckIntact(t, &back)
	zwanzigerrufenPlayOut(t, &back)
}

// 改竄した保存データは、本物の局面を 1 か所だけ壊して作る。
func zwanzigerrufenTamper(t *testing.T, mutate func(m map[string]any)) error {
	t.Helper()
	g := zwanzigerrufenAtPlay(t)
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	var back Zwanzigerrufen
	return json.Unmarshal(tampered, &back)
}

func TestZwanzigerrufenUnmarshal_RejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 9 }, "invalid phase"},
		{"手番が範囲外", func(m map[string]any) { m["cp"] = 4 }, "current player out of range"},
		{"親が範囲外", func(m map[string]any) { m["di"] = -1 }, "dealer out of range"},
		{"デクレアラーが範囲外", func(m map[string]any) { m["dc"] = 9 }, "declarer out of range"},
		{"パートナーが範囲外", func(m map[string]any) { m["pi"] = 9 }, "partner out of range"},
		{"ラウンドが 0", func(m map[string]any) { m["rn"] = 0 }, "round 0 out of range"},
		{"トリックが多すぎる", func(m map[string]any) {
			m["tn"] = ZwanzigerrufenTrickCount + 1
		}, "trick"},
		{"契約が不正", func(m map[string]any) { m["co"] = 7 }, "invalid contract"},
		{"呼び札が 1 番", func(m map[string]any) { m["cl"] = 1 }, "called trump 1 out of range"},
		// **範囲チェックだけでは通ってしまう組み合わせ。**
		{"Trischaken なのに呼び札がある", func(m map[string]any) {
			m["co"] = int(ZwanzigerrufenBidTrischaken)
			m["dc"] = -1
			m["cl"] = ZwanzigerrufenCallTrump
		}, "requires the rufer contract"},
		{"Trischaken なのにデクレアラーが居る", func(m map[string]any) {
			m["co"] = int(ZwanzigerrufenBidTrischaken)
			m["cl"] = -1
			m["dc"] = 1
		}, "trischaken has no declarer"},
		{"呼び札が無いのにパートナーが居る", func(m map[string]any) {
			m["cl"] = -1
			m["pi"] = 2
		}, "requires a called trump"},
		{"自分自身がパートナー", func(m map[string]any) {
			m["dc"] = 1
			m["pi"] = 1
		}, "own partner"},
		{"設定が不正", func(m map[string]any) {
			m["cf"] = map[string]any{"cd": 0, "td": 0}
		}, "target deals"},
		{"席数が違う", func(m map[string]any) {
			m["pl"] = m["pl"].([]any)[:3]
		}, "expected 4 players"},
		{"席が nil", func(m map[string]any) {
			pl := m["pl"].([]any)
			pl[1] = nil
			m["pl"] = pl
		}, "nil player"},
		{"場札の札が範囲外", func(m map[string]any) {
			m["st"] = []any{map[string]any{"d": 9, "v": 1}}
		}, "card out of range"},
		{"場札が nil", func(m map[string]any) {
			m["st"] = []any{nil}
		}, "nil card"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := zwanzigerrufenTamper(t, tt.mutate)
			require.Error(t, err, "改竄が素通りしている")
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestZwanzigerrufenUnmarshal_RejectsOversizedArrays(t *testing.T) {
	big := make([]map[string]int, zwanzigerrufenMaxSliceLen+1)
	for i := range big {
		big[i] = map[string]int{"d": 1, "v": 1}
	}
	payload, err := json.Marshal(map[string]any{"st": big})
	require.NoError(t, err)
	var g Zwanzigerrufen
	err = json.Unmarshal(payload, &g)
	require.Error(t, err)
	assert.ErrorContains(t, err, "maximum allowed size")
}

func TestZwanzigerrufenUnmarshal_RejectsMalformedJSON(t *testing.T) {
	var g Zwanzigerrufen
	assert.Error(t, json.Unmarshal([]byte(`{"ph":"x"}`), &g))
}

func TestZwanzigerrufenPlayer_RoundTripJSON(t *testing.T) {
	p := NewZwanzigerrufenPlayer(true)
	p.AddCard(NewCard(KoenigrufenTrumpDesign, 20, false))
	p.AddTrick([]*Card{NewCard(1, KoenigrufenKingValue, false)})

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var back ZwanzigerrufenPlayer
	require.NoError(t, json.Unmarshal(b, &back))
	assert.True(t, back.GetIsHuman())
	assert.Equal(t, 1, back.GetCardsSize())
	assert.Equal(t, 1, back.GetTrickCount())
}

func TestZwanzigerrufenBidName(t *testing.T) {
	assert.Equal(t, "trischaken", ZwanzigerrufenBidName(ZwanzigerrufenBidTrischaken))
	assert.Equal(t, "rufer", ZwanzigerrufenBidName(ZwanzigerrufenBidRufer))
	assert.Equal(t, "solo", ZwanzigerrufenBidName(ZwanzigerrufenBidSolo))
	assert.Equal(t, "pass", ZwanzigerrufenBidName(ZwanzigerrufenBidPass))
}
