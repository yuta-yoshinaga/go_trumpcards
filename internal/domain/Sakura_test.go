//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSakuraForTest は席数を指定してさくらを開始する。
func newSakuraForTest(t *testing.T, seats int) *Sakura {
	t.Helper()
	cfg := SakuraConfig{Seats: seats, Rounds: SakuraDefaultRounds}
	require.NoError(t, cfg.Validate())
	g := NewSakura(NewSakuraPlayersForTable(seats), cfg)
	g.Reset()
	return g
}

// sakuraCardsInPlay は盤上のすべての札を集める (保存則の検査に使う)。
func sakuraCardsInPlay(g *Sakura) []*Card {
	out := make([]*Card, 0, SakuraDeckSize)
	out = append(out, g.GetField()...)
	out = append(out, g.stock...)
	for _, p := range g.GetPlayers() {
		out = append(out, p.GetCards()...)
		out = append(out, p.GetTaken()...)
	}
	return out
}

// assertSakuraDeckIntact は 48 枚が 1 枚ずつ揃っていることを確かめる。
func assertSakuraDeckIntact(t *testing.T, g *Sakura) {
	t.Helper()
	cards := sakuraCardsInPlay(g)
	assert.Len(t, cards, SakuraDeckSize, "盤上の総枚数")
	seen := map[[2]int]bool{}
	for _, c := range cards {
		require.NotNil(t, c)
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "札 %v が重複", key)
		seen[key] = true
	}
}

// playSakuraToEnd は終局まで進める。手数の上限で打ち切る。
func playSakuraToEnd(t *testing.T, g *Sakura) {
	t.Helper()
	for range 2000 {
		if g.GetGameEndFlag() {
			return
		}
		switch {
		case g.GetPhase() == SakuraPhaseRoundEnd:
			assertSakuraDeckIntact(t, g)
			g.NextRound()
		case g.IsHumanTurn():
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
		default:
			before := g.GetPlayers()[g.GetTurn()].GetCardsSize()
			g.CpuPlay()
			assert.Less(t, g.GetPlayers()[g.GetTurn()].GetCardsSize(), before+1,
				"CPU の手番が進んでいない")
		}
	}
	t.Fatal("終局しなかった")
}

// --- 設定 ---

func TestSakuraConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SakuraConfig
		wantErr error
	}{
		{"既定", DefaultSakuraConfig(), nil},
		{"最小", SakuraConfig{Seats: SakuraMinSeats, Rounds: SakuraMinRounds}, nil},
		{"最大", SakuraConfig{Seats: SakuraMaxSeats, Rounds: SakuraMaxRounds}, nil},
		{"席が少なすぎる", SakuraConfig{Seats: SakuraMinSeats - 1, Rounds: 3}, errSakuraSeatsRange},
		{"席が多すぎる", SakuraConfig{Seats: SakuraMaxSeats + 1, Rounds: 3}, errSakuraSeatsRange},
		{"ラウンドが 0", SakuraConfig{Seats: 3, Rounds: 0}, errSakuraRoundsRange},
		{"ラウンドが多すぎる", SakuraConfig{Seats: 3, Rounds: SakuraMaxRounds + 1}, errSakuraRoundsRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// 最大席数でも山札が残ることを、定数から直接確かめる。
// **この不変条件が崩れるとめくりが成り立たない。**
func TestSakuraConfig_MaxSeatsStillLeavesStock(t *testing.T) {
	dealt := SakuraMaxSeats*SakuraHandSize + SakuraFieldSize
	assert.Less(t, dealt, SakuraDeckSize, "配り切ると山が残らない")
	assert.NoError(t, SakuraConfig{Seats: SakuraMaxSeats, Rounds: 1}.Validate())
}

func TestSakuraConfig_RoundTripJSON(t *testing.T) {
	cfg := SakuraConfig{Seats: 4, Rounds: 5}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	var back SakuraConfig
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, cfg, back)
}

func TestSakuraConfig_UnmarshalRejectsBadValues(t *testing.T) {
	var cfg SakuraConfig
	assert.Error(t, json.Unmarshal([]byte(`{"s":9,"r":3}`), &cfg))
	assert.Error(t, json.Unmarshal([]byte(`{"s":3,"r":0}`), &cfg))
	assert.Error(t, json.Unmarshal([]byte(`{"s":"x"}`), &cfg))
}

// --- 点数 ---

// 花札 48 枚の点数構成を数え上げる。**表を写さず自分で数える。**
func TestSakuraCardPoints_DeckComposition(t *testing.T) {
	counts := map[int]int{}
	total := 0
	for _, c := range buildKoiKoiDeck() {
		p := SakuraCardPoints(c)
		counts[p]++
		total += p
	}
	assert.Equal(t, 5, counts[SakuraBrightPoints], "20 点札 (光)")
	assert.Equal(t, 9, counts[SakuraAnimalPoints], "10 点札 (タネ)")
	assert.Equal(t, 10, counts[SakuraRibbonPoints], "5 点札 (短冊)")
	assert.Equal(t, 24, counts[SakuraChaffPoints], "1 点札 (カス)")
	assert.Equal(t, 5*20+9*10+10*5+24*1, total)
	assert.Equal(t, 264, total, "花札 48 枚の総点")
}

func TestSakuraCardPoints_NilIsZero(t *testing.T) {
	assert.Equal(t, 0, SakuraCardPoints(nil))
}

func TestSakuraBonus_NameAndPoints(t *testing.T) {
	assert.Equal(t, "allBrights", SakuraBonusName(SakuraBonusAllBrights))
	assert.Equal(t, "sakuraSake", SakuraBonusName(SakuraBonusSakuraSake))
	assert.Equal(t, "none", SakuraBonusName(SakuraBonusNone))
	assert.Equal(t, 100, SakuraBonusPoints(SakuraBonusAllBrights))
	assert.Equal(t, 30, SakuraBonusPoints(SakuraBonusSakuraSake))
	assert.Equal(t, 0, SakuraBonusPoints(SakuraBonusNone))
	assert.Equal(t, SakuraBonusAllBrights, SakuraBonusMax)
}

// 追加役の札が実在の光札であることを、デッキ側から確かめる。
func TestSakuraBonusCardsExistInTheDeck(t *testing.T) {
	for _, tc := range []struct {
		name         string
		month, index int
	}{
		{"桜に幕", SakuraCurtainMonth, SakuraCurtainIndex},
		{"芒に月", SakuraMoonMonth, SakuraMoonIndex},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCard(tc.month, tc.index, false)
			assert.Equal(t, KoiKoiBright, KoiKoiCardCategory(c))
			assert.Equal(t, SakuraBrightPoints, SakuraCardPoints(c))
		})
	}
	brights := 0
	for _, c := range buildKoiKoiDeck() {
		if KoiKoiCardCategory(c) == KoiKoiBright {
			brights++
		}
	}
	assert.Equal(t, SakuraAllBrightsCount, brights, "光札の枚数と定数が食い違う")
}

// --- プレイヤー ---

func TestSakuraPlayer_CardPointsAndCategories(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	p.AddTaken(
		NewCard(SakuraCurtainMonth, SakuraCurtainIndex, false), // 光 20
		NewCard(8, 2, false), // タネ (芒に雁) 10
		NewCard(1, 2, false), // 短冊 5
		NewCard(1, 3, false), // カス 1
	)
	assert.Equal(t, 36, p.CardPoints())
	counts := p.CountByCategory()
	assert.Equal(t, 1, counts[KoiKoiBright])
	assert.Equal(t, 1, counts[KoiKoiAnimal])
	assert.Equal(t, 1, counts[KoiKoiRibbon])
	assert.Equal(t, 1, counts[KoiKoiChaff])
}

func TestSakuraPlayer_AddTakenSkipsNil(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	p.AddTaken(nil, NewCard(1, 1, false), nil)
	assert.Len(t, p.GetTaken(), 1)
}

func TestSakuraPlayer_BonusSakuraSakeNeedsBothCards(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	p.AddTaken(NewCard(SakuraCurtainMonth, SakuraCurtainIndex, false))
	assert.Empty(t, p.Bonuses(), "幕だけでは成立しない")
	assert.Equal(t, 0, p.BonusPoints())

	p.AddTaken(NewCard(SakuraMoonMonth, SakuraMoonIndex, false))
	assert.Equal(t, []SakuraBonus{SakuraBonusSakuraSake}, p.Bonuses())
	assert.Equal(t, 30, p.BonusPoints())
	assert.Equal(t, 40+30, p.TotalPoints(), "素点 40 に追加役 30")
}

func TestSakuraPlayer_BonusAllBrights(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	for _, c := range buildKoiKoiDeck() {
		if KoiKoiCardCategory(c) == KoiKoiBright {
			p.AddTaken(c)
		}
	}
	// 光 5 枚には幕も月も含まれるので、両方の役が立つ。
	assert.ElementsMatch(t, []SakuraBonus{SakuraBonusAllBrights, SakuraBonusSakuraSake}, p.Bonuses())
	assert.Equal(t, 130, p.BonusPoints())
	assert.Equal(t, 5*SakuraBrightPoints, p.CardPoints())
	assert.Equal(t, 100+130, p.TotalPoints())
}

// 光 4 枚では大役が立たない (境界)。
func TestSakuraPlayer_FourBrightsIsNotAllBrights(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	added := 0
	for _, c := range buildKoiKoiDeck() {
		if KoiKoiCardCategory(c) == KoiKoiBright && added < SakuraAllBrightsCount-1 {
			p.AddTaken(c)
			added++
		}
	}
	assert.NotContains(t, p.Bonuses(), SakuraBonusAllBrights)
}

func TestSakuraPlayer_ResetForRoundClearsHandAndTaken(t *testing.T) {
	p := NewSakuraPlayer("YOU", true)
	p.AddCard(NewCard(1, 1, false))
	p.AddTaken(NewCard(2, 1, false))
	p.SetRoundScore(42)
	p.AddScore(42)
	p.AddRoundWin()

	p.ResetForRound()
	assert.Empty(t, p.GetCards())
	assert.Empty(t, p.GetTaken())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 42, p.GetScore(), "通算得点は残す")
	assert.Equal(t, 1, p.GetRoundWins(), "勝ラウンド数は残す")
}

func TestSakuraPlayer_RoundTripJSON(t *testing.T) {
	p := NewSakuraPlayer("CPU1", false)
	p.AddCard(NewCard(1, 1, false))
	p.AddTaken(NewCard(2, 2, false))
	p.AddScore(15)
	p.SetRoundScore(15)
	p.AddRoundWin()

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var back SakuraPlayer
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, "CPU1", back.GetName())
	assert.False(t, back.GetIsHuman())
	assert.Len(t, back.GetCards(), 1)
	assert.Len(t, back.GetTaken(), 1)
	assert.Equal(t, 15, back.GetScore())
	assert.Equal(t, 15, back.GetRoundScore())
	assert.Equal(t, 1, back.GetRoundWins())
}

func TestSakuraPlayer_UnmarshalRejectsImpossibleState(t *testing.T) {
	tests := []struct {
		name string
		data string
		want error
	}{
		{"手札が配る枚数を超える",
			`{"cd":[{"d":1,"v":1},{"d":1,"v":2},{"d":1,"v":3},{"d":1,"v":4},{"d":2,"v":1},{"d":2,"v":2},{"d":2,"v":3},{"d":2,"v":4}]}`,
			errSakuraHandTooLarge},
		{"勝ラウンド数が負", `{"rw":-1}`, errSakuraNegativeWins},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p SakuraPlayer
			assert.ErrorIs(t, json.Unmarshal([]byte(tt.data), &p), tt.want)
		})
	}
}

// --- 配り ---

func TestSakuraReset_DealsHandFieldAndStock(t *testing.T) {
	for _, seats := range []int{SakuraMinSeats, 3, SakuraMaxSeats} {
		t.Run(sakuraSeatsLabel(seats), func(t *testing.T) {
			g := newSakuraForTest(t, seats)
			for i, p := range g.GetPlayers() {
				assert.Equal(t, SakuraHandSize, p.GetCardsSize(), "席 %d の手札", i)
				assert.Empty(t, p.GetTaken(), "席 %d の獲得札", i)
			}
			assert.Len(t, g.GetField(), SakuraFieldSize)
			assert.Equal(t, SakuraDeckSize-seats*SakuraHandSize-SakuraFieldSize, g.GetStockCount())
			assert.Equal(t, SakuraPhasePlay, g.GetPhase())
			assert.Equal(t, 1, g.GetRound())
			assert.False(t, g.GetGameEndFlag())
			assertSakuraDeckIntact(t, g)
		})
	}
}

func sakuraSeatsLabel(seats int) string {
	return strings.Repeat("席", 1) + string(rune('0'+seats))
}

func TestSakuraReset_HumanIsSeatZeroAndLeadsFirstRound(t *testing.T) {
	g := newSakuraForTest(t, 3)
	assert.True(t, g.GetPlayers()[0].GetIsHuman())
	assert.Equal(t, 0, g.HumanSeat())
	assert.Equal(t, 0, g.GetDealer())
	assert.Equal(t, 0, g.GetTurn())
	assert.True(t, g.IsHumanTurn())
}

func TestSakuraNewDefaultSakura(t *testing.T) {
	g := NewDefaultSakura()
	g.Reset()
	assert.Equal(t, SakuraDefaultSeats, len(g.GetPlayers()))
	assert.Equal(t, DefaultSakuraConfig(), g.GetConfig())
}

// --- 手番 ---

func TestSakuraPlayerPlay_CapturesTheSameMonth(t *testing.T) {
	g := newSakuraForTest(t, 2)
	// 手札の 1 枚目と同月の札だけを場に置く。
	hand := g.GetPlayers()[0].GetCards()[0]
	g.field = []*Card{NewCard(hand.GetDesign(), sakuraOtherIndex(hand.GetValue()), false)}
	g.stock = nil

	require.NoError(t, g.PlayerPlay(0, 0))
	assert.Len(t, g.GetPlayers()[0].GetTaken(), 2, "出した札と場札の 2 枚を取る")
	assert.Empty(t, g.GetField(), "場から取り除かれる")
	assert.Equal(t, SakuraHandSize-1, g.GetPlayers()[0].GetCardsSize())
}

func TestSakuraPlayerPlay_DiscardsWhenNoMatch(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	g.field = []*Card{NewCard(sakuraOtherMonth(hand.GetDesign()), 1, false)}
	g.stock = nil

	require.NoError(t, g.PlayerPlay(0, -1))
	assert.Empty(t, g.GetPlayers()[0].GetTaken())
	assert.Len(t, g.GetField(), 2, "捨てた札が場に残る")
}

// 場に同月が 3 枚あれば 4 枚まとめて取る。
func TestSakuraPlayerPlay_TakesAllFourOfAMonth(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	m := hand.GetDesign()
	g.field = nil
	for i := 1; i <= KoiKoiCardsPerMonth; i++ {
		if i != hand.GetValue() {
			g.field = append(g.field, NewCard(m, i, false))
		}
	}
	g.stock = nil

	require.NoError(t, g.PlayerPlay(0, -1))
	assert.Len(t, g.GetPlayers()[0].GetTaken(), KoiKoiCardsPerMonth)
	assert.Empty(t, g.GetField())
}

// 一致が 2 枚あるときは指定した札を取る。**選択が効いていることを見る。**
func TestSakuraPlayerPlay_HonoursTheChosenFieldCard(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	m := hand.GetDesign()
	idxs := []int{}
	for i := 1; i <= KoiKoiCardsPerMonth; i++ {
		if i != hand.GetValue() && len(idxs) < 2 {
			idxs = append(idxs, i)
		}
	}
	g.field = []*Card{NewCard(m, idxs[0], false), NewCard(m, idxs[1], false)}
	g.stock = nil
	// 点数の低いほう (自動選択が選ばない側) を敢えて指定する。
	pick := 0
	if SakuraCardPoints(g.field[1]) < SakuraCardPoints(g.field[0]) {
		pick = 1
	}
	want := g.field[pick]

	require.NoError(t, g.PlayerPlay(0, pick))
	taken := g.GetPlayers()[0].GetTaken()
	require.Len(t, taken, 2)
	assert.Same(t, want, taken[1], "指定した場札を取っていない")
}

// 指定しなければ点数の高い札を取る。
func TestSakuraPlayerPlay_AutoPicksTheRicherFieldCard(t *testing.T) {
	g := newSakuraForTest(t, 2)
	// 松 (1 月): index 1 が光 20 点、index 3 がカス 1 点。
	hand := NewCard(1, 2, false)
	g.players[0].cards = []*Card{hand}
	g.field = []*Card{NewCard(1, 3, false), NewCard(1, 1, false)}
	g.stock = nil

	require.NoError(t, g.PlayerPlay(0, -1))
	taken := g.GetPlayers()[0].GetTaken()
	require.Len(t, taken, 2)
	assert.Equal(t, SakuraBrightPoints, SakuraCardPoints(taken[1]), "高い札を取っていない")
}

func TestSakuraPlayerPlay_Errors(t *testing.T) {
	t.Run("手札の範囲外", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		assert.Error(t, g.PlayerPlay(-1, -1))
		assert.Error(t, g.PlayerPlay(SakuraHandSize, -1))
	})
	t.Run("違う月の場札を指定", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		hand := g.GetPlayers()[0].GetCards()[0]
		g.field = []*Card{NewCard(sakuraOtherMonth(hand.GetDesign()), 1, false)}
		err := g.PlayerPlay(0, 0)
		require.Error(t, err)
		assert.ErrorContains(t, err, "does not match")
	})
	t.Run("場札の範囲外", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		assert.Error(t, g.PlayerPlay(0, 99))
	})
	t.Run("CPU の手番", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		g.turn = 1
		assert.ErrorIs(t, g.PlayerPlay(0, -1), ErrNotHumanTurn)
	})
	t.Run("プレイフェーズでない", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		g.phase = SakuraPhaseRoundEnd
		err := g.PlayerPlay(0, -1)
		require.Error(t, err)
		assert.ErrorContains(t, err, "play phase")
	})
	t.Run("終局後", func(t *testing.T) {
		g := newSakuraForTest(t, 2)
		g.gameEndFlag = true
		assert.ErrorIs(t, g.PlayerPlay(0, -1), ErrGameEnded)
	})
}

// 手番のたびに山札を 1 枚めくる。
func TestSakuraTurn_FlipsOneStockCard(t *testing.T) {
	g := newSakuraForTest(t, 2)
	before := g.GetStockCount()
	require.NoError(t, g.PlayerPlay(0, -1))
	assert.Equal(t, before-1, g.GetStockCount())
}

// **山が尽きてもめくりだけを飛ばす。** 手番そのものは成立しなければならない。
func TestSakuraTurn_ContinuesWithAnEmptyStock(t *testing.T) {
	g := newSakuraForTest(t, 2)
	g.stock = nil
	require.NoError(t, g.PlayerPlay(0, -1))
	assert.Equal(t, 0, g.GetStockCount())
	assert.Equal(t, SakuraHandSize-1, g.GetPlayers()[0].GetCardsSize())
	assert.Equal(t, 1, g.GetTurn(), "手番が進んでいない")
}

// 4 席では山が手札より先に尽きる。**それでもラウンドは終わる。**
func TestSakuraFourSeats_StockRunsOutBeforeHands(t *testing.T) {
	g := newSakuraForTest(t, SakuraMaxSeats)
	handCards := SakuraMaxSeats * SakuraHandSize
	assert.Less(t, g.GetStockCount(), handCards, "この席数では山のほうが少ないはず")

	for range 500 {
		if g.GetPhase() != SakuraPhasePlay {
			break
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}
	assert.Equal(t, 0, g.GetStockCount(), "山を使い切っていない")
	assert.NotEqual(t, SakuraPhasePlay, g.GetPhase(), "ラウンドが終わっていない")
	assertSakuraDeckIntact(t, g)
}

func TestSakuraCpuPlay_DoesNothingOnHumanTurn(t *testing.T) {
	g := newSakuraForTest(t, 2)
	before := g.GetPlayers()[0].GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayers()[0].GetCardsSize())
	assert.Equal(t, 0, g.GetTurn())
}

// --- 集計 ---

func TestSakuraFinishRound_HighestTotalTakesTheRound(t *testing.T) {
	g := newSakuraForTest(t, 2)
	for _, p := range g.GetPlayers() {
		p.cards = nil
	}
	g.players[0].AddTaken(NewCard(1, 1, false)) // 光 20
	g.players[1].AddTaken(NewCard(1, 3, false)) // カス 1
	g.finishRound()

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, 0, res.Winner)
	assert.Equal(t, 20, res.Seats[0].Total)
	assert.Equal(t, 1, res.Seats[1].Total)
	assert.Equal(t, 20, g.GetPlayers()[0].GetScore())
	assert.Equal(t, 1, g.GetPlayers()[0].GetRoundWins())
	assert.Equal(t, 0, g.GetPlayers()[1].GetRoundWins())
	assert.Equal(t, SakuraPhaseRoundEnd, g.GetPhase())
}

func TestSakuraFinishRound_TieHasNoWinner(t *testing.T) {
	g := newSakuraForTest(t, 2)
	for _, p := range g.GetPlayers() {
		p.cards = nil
	}
	g.players[0].AddTaken(NewCard(1, 1, false))
	g.players[1].AddTaken(NewCard(3, 1, false)) // 同じ光 20 点
	g.finishRound()

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, -1, res.Winner)
	assert.Equal(t, 0, g.GetPlayers()[0].GetRoundWins())
	assert.Equal(t, 0, g.GetPlayers()[1].GetRoundWins())
}

// 追加役はラウンドの結果に乗る。
func TestSakuraFinishRound_BonusCountsTowardTheTotal(t *testing.T) {
	g := newSakuraForTest(t, 2)
	for _, p := range g.GetPlayers() {
		p.cards = nil
	}
	g.players[0].AddTaken(
		NewCard(SakuraCurtainMonth, SakuraCurtainIndex, false),
		NewCard(SakuraMoonMonth, SakuraMoonIndex, false),
	)
	// 相手には素点だけで上回らせる (40 点 < 50 点)。
	for i := 1; i <= 5; i++ {
		g.players[1].AddTaken(NewCard(i, 2, false))
	}
	g.players[1].taken = []*Card{
		NewCard(2, 1, false), NewCard(4, 1, false), NewCard(5, 1, false), // タネ 30
		NewCard(1, 2, false), NewCard(2, 2, false), NewCard(4, 2, false), NewCard(5, 2, false), // 短冊 20
	}
	g.finishRound()

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, 40, res.Seats[0].CardPoints)
	assert.Equal(t, 30, res.Seats[0].BonusPoints)
	assert.Equal(t, 70, res.Seats[0].Total)
	assert.Equal(t, []SakuraBonus{SakuraBonusSakuraSake}, res.Seats[0].Bonuses)
	assert.Equal(t, 50, res.Seats[1].Total)
	assert.Equal(t, 0, res.Winner, "追加役ぶんで逆転していない")
}

func TestSakuraNextRound_RotatesTheDealerAndRedeals(t *testing.T) {
	g := newSakuraForTest(t, 3)
	for _, p := range g.GetPlayers() {
		p.cards = nil
	}
	g.finishRound()
	require.Equal(t, SakuraPhaseRoundEnd, g.GetPhase())

	g.NextRound()
	assert.Equal(t, 2, g.GetRound())
	assert.Equal(t, 1, g.GetDealer(), "親が回っていない")
	assert.Equal(t, 1, g.GetTurn(), "親から打ち始めていない")
	for _, p := range g.GetPlayers() {
		assert.Equal(t, SakuraHandSize, p.GetCardsSize())
		assert.Empty(t, p.GetTaken(), "獲得札が持ち越されている")
	}
	assertSakuraDeckIntact(t, g)
}

func TestSakuraNextRound_IgnoredOutsideRoundEnd(t *testing.T) {
	g := newSakuraForTest(t, 2)
	g.NextRound()
	assert.Equal(t, 1, g.GetRound())

	g.phase = SakuraPhaseRoundEnd
	g.gameEndFlag = true
	g.NextRound()
	assert.Equal(t, 1, g.GetRound(), "終局後に次ラウンドへ進んでいる")
}

// 規定ラウンドを終えたら終局する。
func TestSakuraFinishGame_EndsAfterTheConfiguredRounds(t *testing.T) {
	g := NewSakura(NewSakuraPlayersForTable(2), SakuraConfig{Seats: 2, Rounds: 1})
	g.Reset()
	for _, p := range g.GetPlayers() {
		p.cards = nil
	}
	g.players[0].AddTaken(NewCard(1, 1, false))
	g.finishRound()

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, SakuraPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinner())
}

// 通算得点が同じなら勝ちラウンド数で決める。
func TestSakuraFinishGame_RoundWinsBreakTheTie(t *testing.T) {
	g := NewSakura(NewSakuraPlayersForTable(2), SakuraConfig{Seats: 2, Rounds: 1})
	g.Reset()
	g.players[0].AddScore(50)
	g.players[1].AddScore(50)
	g.players[1].AddRoundWin()
	g.finishGame()
	assert.Equal(t, 1, g.GetWinner())
}

func TestSakuraFinishGame_TieWhenNothingSeparatesTheSeats(t *testing.T) {
	g := NewSakura(NewSakuraPlayersForTable(2), SakuraConfig{Seats: 2, Rounds: 1})
	g.Reset()
	g.players[0].AddScore(50)
	g.players[1].AddScore(50)
	g.finishGame()
	assert.Equal(t, -1, g.GetWinner())
}

// --- 通し ---

func TestSakuraFullGame_Terminates(t *testing.T) {
	for _, seats := range []int{SakuraMinSeats, 3, SakuraMaxSeats} {
		t.Run(sakuraSeatsLabel(seats), func(t *testing.T) {
			g := newSakuraForTest(t, seats)
			playSakuraToEnd(t, g)
			assert.True(t, g.GetGameEndFlag())
			assert.Equal(t, SakuraDefaultRounds, g.GetRound())
			assertSakuraDeckIntact(t, g)
			for _, p := range g.GetPlayers() {
				assert.Empty(t, p.GetCards(), "手札が残っている")
			}
			// 得点は毎ラウンド積み上がるので 0 にはならない。
			total := 0
			for _, p := range g.GetPlayers() {
				total += p.GetScore()
			}
			assert.Positive(t, total)
			assert.NotEmpty(t, g.GetActionLog())
		})
	}
}

// 何度回しても終わる (配り依存で止まらないことの確認)。
func TestSakuraFullGame_TerminatesRepeatedly(t *testing.T) {
	for range 30 {
		g := newSakuraForTest(t, 3)
		playSakuraToEnd(t, g)
		require.True(t, g.GetGameEndFlag())
	}
}

// --- ヒント ---

func TestSakuraGetHint_SuggestsALegalPlay(t *testing.T) {
	g := newSakuraForTest(t, 2)
	h := g.GetHint()
	require.GreaterOrEqual(t, h.CardIndex, 0)
	assert.Less(t, h.CardIndex, SakuraHandSize)
	assert.Contains(t, []string{"capture", "discard"}, h.Reason)
	assert.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex), "ヒントの手が通らない")
}

func TestSakuraGetHint_CaptureWhenAMatchExists(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	g.field = []*Card{NewCard(hand.GetDesign(), sakuraOtherIndex(hand.GetValue()), false)}
	h := g.GetHint()
	assert.Equal(t, 0, h.CardIndex)
	assert.Equal(t, "capture", h.Reason)
}

func TestSakuraGetHint_NoneOutsidePlay(t *testing.T) {
	g := newSakuraForTest(t, 2)
	g.phase = SakuraPhaseRoundEnd
	h := g.GetHint()
	assert.Equal(t, -1, h.CardIndex)
	assert.Equal(t, "none", h.Reason)

	g.phase = SakuraPhasePlay
	g.turn = 1
	assert.Equal(t, "none", g.GetHint().Reason)
}

// **合わせられる = 選ばせる、ではない。** 場に同月が 3 枚あると 4 枚まとめて
// 取るので、どれを押しても結果が変わらない ── 選択を求めてはいけない。
func TestSakuraGetChoiceIndices_OnlyWhenTheChoiceMatters(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	m := hand.GetDesign()
	others := []int{}
	for i := 1; i <= KoiKoiCardsPerMonth; i++ {
		if i != hand.GetValue() {
			others = append(others, i)
		}
	}
	require.Len(t, others, 3)

	// 1 枚一致: 選ばせない。
	g.field = []*Card{NewCard(m, others[0], false)}
	assert.Empty(t, g.GetChoiceIndices()[0])
	assert.Len(t, g.GetValidFieldIndices()[0], 1, "合わせられることは伝える")

	// 2 枚一致: 選ばせる。
	g.field = []*Card{NewCard(m, others[0], false), NewCard(m, others[1], false)}
	assert.Len(t, g.GetChoiceIndices()[0], 2)

	// 3 枚一致: まとめて取るので選ばせない。
	g.field = []*Card{
		NewCard(m, others[0], false), NewCard(m, others[1], false), NewCard(m, others[2], false),
	}
	assert.Empty(t, g.GetChoiceIndices()[0], "3 枚一致で選択を求めている")
	assert.Len(t, g.GetValidFieldIndices()[0], 3, "取れる札は 3 枚とも伝える")

	// 実際にまとめて取れる (どれを指定しても同じ)。
	g.stock = nil
	require.NoError(t, g.PlayerPlay(0, 0))
	assert.Len(t, g.GetPlayers()[0].GetTaken(), KoiKoiCardsPerMonth)
}

func TestSakuraGetValidFieldIndices(t *testing.T) {
	g := newSakuraForTest(t, 2)
	hand := g.GetPlayers()[0].GetCards()[0]
	g.field = []*Card{
		NewCard(sakuraOtherMonth(hand.GetDesign()), 1, false),
		NewCard(hand.GetDesign(), sakuraOtherIndex(hand.GetValue()), false),
	}
	valid := g.GetValidFieldIndices()
	assert.Equal(t, []int{1}, valid[0])

	g.phase = SakuraPhaseRoundEnd
	assert.Empty(t, g.GetValidFieldIndices())
}

// --- 永続化 ---

func TestSakuraRoundTripJSON_MidGame(t *testing.T) {
	g := newSakuraForTest(t, 3)
	for range 6 {
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}

	b, err := json.Marshal(g)
	require.NoError(t, err)
	var back Sakura
	require.NoError(t, json.Unmarshal(b, &back))

	assert.Equal(t, g.GetPhase(), back.GetPhase())
	assert.Equal(t, g.GetTurn(), back.GetTurn())
	assert.Equal(t, g.GetDealer(), back.GetDealer())
	assert.Equal(t, g.GetStockCount(), back.GetStockCount())
	assert.Equal(t, len(g.GetField()), len(back.GetField()))
	assert.Equal(t, g.GetRound(), back.GetRound())
	assert.Equal(t, len(g.GetActionLog()), len(back.GetActionLog()))
	for i, p := range g.GetPlayers() {
		assert.Equal(t, p.GetCardsSize(), back.GetPlayers()[i].GetCardsSize())
		assert.Equal(t, p.CardPoints(), back.GetPlayers()[i].CardPoints())
	}
	assertSakuraDeckIntact(t, &back)
	// 復元後も打ち続けられる。
	playSakuraToEnd(t, &back)
}

// 改竄した保存データは、本物の局面を 1 か所だけ壊して作る。
func sakuraTamper(t *testing.T, mutate func(m map[string]any)) error {
	t.Helper()
	g := newSakuraForTest(t, 3)
	for range 4 {
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)

	var back Sakura
	return json.Unmarshal(tampered, &back)
}

func TestSakuraUnmarshal_RejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 9 }, "invalid phase"},
		{"手番が席数を超える", func(m map[string]any) { m["tn"] = 3 }, "turn out of range"},
		{"親が範囲外", func(m map[string]any) { m["dl"] = -1 }, "dealer out of range"},
		{"ラウンドが 0", func(m map[string]any) { m["rd"] = 0 }, "round 0 out of range"},
		{"ラウンドが設定を超える", func(m map[string]any) { m["rd"] = SakuraDefaultRounds + 1 }, "out of range"},
		{"勝者が範囲外", func(m map[string]any) { m["wn"] = 5 }, "winner out of range"},
		{"設定が不正", func(m map[string]any) {
			m["cf"] = map[string]any{"s": 9, "r": 3}
		}, "seats out of range"},
		{"席数と人数が食い違う", func(m map[string]any) {
			m["pl"] = m["pl"].([]any)[:2]
		}, "does not match seats"},
		{"席が nil", func(m map[string]any) {
			pl := m["pl"].([]any)
			pl[1] = nil
			m["pl"] = pl
		}, "nil player"},
		{"場札が nil", func(m map[string]any) {
			m["fd"] = []any{nil}
		}, "nil card"},
		{"場札の月が範囲外", func(m map[string]any) {
			fd := m["fd"].([]any)
			fd[0].(map[string]any)["d"] = 13
			m["fd"] = fd
		}, "card out of range"},
		{"山札の札が範囲外", func(m map[string]any) {
			st := m["st"].([]any)
			st[0].(map[string]any)["v"] = 0
			m["st"] = st
		}, "card out of range"},
		{"同じ札が 2 枚ある", func(m map[string]any) {
			st := m["st"].([]any)
			fd := m["fd"].([]any)
			st[0] = fd[0]
			m["st"] = st
		}, "duplicate card"},
		{"札の総数が多すぎる", func(m map[string]any) {
			// 山札を丸ごと重ねると 48 枚を超える。
			m["st"] = append(m["st"].([]any), m["st"].([]any)...)
		}, "duplicate card"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sakuraTamper(t, tt.mutate)
			require.Error(t, err, "改竄が素通りしている")
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestSakuraUnmarshal_RejectsOversizedArrays(t *testing.T) {
	big := make([]map[string]int, sakuraMaxSliceLen+1)
	for i := range big {
		big[i] = map[string]int{"d": 1, "v": 1}
	}
	payload, err := json.Marshal(map[string]any{"st": big})
	require.NoError(t, err)
	var g Sakura
	err = json.Unmarshal(payload, &g)
	require.Error(t, err)
	assert.ErrorContains(t, err, "maximum allowed size")
}

func TestSakuraUnmarshal_RejectsMalformedJSON(t *testing.T) {
	var g Sakura
	assert.Error(t, json.Unmarshal([]byte(`{"tn":"x"}`), &g))
}

// --- 補助 ---

// sakuraOtherIndex は同じ月の別のインデックスを返す。
func sakuraOtherIndex(index int) int {
	if index == 1 {
		return 2
	}
	return 1
}

// sakuraOtherMonth は別の月を返す。
func sakuraOtherMonth(month int) int {
	if month == 1 {
		return 2
	}
	return 1
}
