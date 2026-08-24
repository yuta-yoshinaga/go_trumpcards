//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPiedmonteseTarotForTest は指定席数の卓を配り終えた状態で返す。
func newPiedmonteseTarotForTest(t *testing.T, seats int) *PiedmonteseTarot {
	t.Helper()
	cfg := DefaultPiedmonteseTarotConfig()
	cfg.Seats = seats
	cfg.TargetDeals = 2
	g := NewPiedmonteseTarot(newPiedmonteseTarotPlayers(seats), cfg)
	g.Reset()
	return g
}

// piedmonteseTarotPlayHand は 1 ディールを最後まで打つ。合法手の先頭を出し続ける
// 乱暴な打ち方だが、**規則を書き直さずに** 1 ディールを通せる唯一の打ち方でもある。
func piedmonteseTarotPlayHand(t *testing.T, g *PiedmonteseTarot) {
	t.Helper()
	for steps := 0; steps < 4000; steps++ {
		switch g.GetPhase() {
		case PiedmonteseTarotPhaseScarto:
			if g.IsHumanScartoTurn() {
				require.NoError(t, g.PlayerScarto(g.cpuSelectScarto(g.GetDealerIdx())))
				continue
			}
			g.CpuScarto()
		case PiedmonteseTarotPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		case PiedmonteseTarotPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == PiedmonteseTarotPhaseTrickEnd {
				g.NextTrick()
			}
		default:
			return // RoundEnd か GameEnd
		}
	}
	require.FailNow(t, "ディールが終わらない")
}

// **配りは席数が決める。** 78 ÷ 席数で割り切ってしまうと、3 人卓は 26 枚配りで
// タロンが 0 枚になり、スカルト (親の捨て札) が空回りする。
func TestPiedmonteseTarot_DealSizesPerSeatCount(t *testing.T) {
	t.Parallel()
	for seats, want := range map[int]struct{ hand, talon int }{
		3: {25, 3},
		4: {19, 2},
	} {
		assert.Equal(t, want.hand, PiedmonteseTarotHandSize(seats), "%d 人卓の手札枚数", seats)
		assert.Equal(t, want.talon, PiedmonteseTarotTalonSize(seats), "%d 人卓のタロン", seats)
		assert.Equal(t, Tarot78DeckSize, want.hand*seats+want.talon, "%d 人卓で 78 枚に合わない", seats)
		assert.Positive(t, want.talon, "%d 人卓のタロンが 0 枚だとスカルトが空回りする", seats)
	}
	// 対応していない席数は 0 (呼び出し側が卓を作れないと分かる)。
	assert.Zero(t, PiedmonteseTarotHandSize(5))
	assert.Zero(t, PiedmonteseTarotTalonSize(5))
}

func TestPiedmonteseTarot_ResetDealsEverySeatAndTheTalon(t *testing.T) {
	t.Parallel()
	for _, seats := range PiedmonteseTarotSeatSizes {
		g := newPiedmonteseTarotForTest(t, seats)
		require.Equal(t, PiedmonteseTarotPhaseScarto, g.GetPhase())
		hand := PiedmonteseTarotHandSize(seats)
		for i, p := range g.GetPlayers() {
			want := hand
			if i == g.GetDealerIdx() {
				// **親はタロンを抱えている。** 捨てるまで手札は多い。
				want += PiedmonteseTarotTalonSize(seats)
			}
			assert.Equal(t, want, p.GetCardsSize(), "席 %d の手札枚数", i)
		}
	}
}

// **点は「3 枚組から 2 を引く」規則そのもの。** 1 枚あたり 3×値 − 2 の
// 1/3 単位で持ち、全 78 枚で 234 thirds = 78 点になる。
func TestPiedmonteseTarot_TheWholeDeckIsSeventyEightPoints(t *testing.T) {
	t.Parallel()
	total := 0
	honours, kings := 0, 0
	for _, c := range buildTarot78Deck() {
		total += piedmonteseTarotCardThirds(c)
		if tarot78IsBout(c) {
			honours++
		}
		if !tarot78IsTrump(c) && !tarot78IsExcuse(c) && c.GetValue() == Tarot78KingValue {
			kings++
		}
	}
	assert.Equal(t, PiedmonteseTarotTotalThirds, total)
	assert.Equal(t, 78, total/PiedmonteseTarotThirdsPerPoint, "1/3 単位を点に直すと 78 点")
	assert.Equal(t, 3, honours, "オヌールは Bagatto・Mondo・Matto の 3 枚")
	assert.Equal(t, 4, kings)

	// 値の表そのものも固定する。ここが動くと合計 78 点が崩れる。
	for _, tt := range []struct {
		name string
		card *Card
		want int
	}{
		{"Bagatto", NewCard(Tarot78TrumpDesign, 1, false), 5},
		{"Mondo", NewCard(Tarot78TrumpDesign, Tarot78MaxTrump, false), 5},
		{"Matto", NewCard(Tarot78ExcuseDesign, Tarot78ExcuseValue, false), 5},
		{"ただの切り札", NewCard(Tarot78TrumpDesign, 10, false), 1},
		{"Roi", NewCard(CardDesignHeart, Tarot78KingValue, false), 5},
		{"Dame", NewCard(CardDesignHeart, 13, false), 4},
		{"Cavalier", NewCard(CardDesignHeart, 12, false), 3},
		{"Valet", NewCard(CardDesignHeart, 11, false), 2},
		{"ピップ", NewCard(CardDesignHeart, 7, false), 1},
	} {
		assert.Equal(t, tt.want, piedmonteseTarotCardValue(tt.card), tt.name)
		assert.Equal(t, 3*tt.want-2, piedmonteseTarotCardThirds(tt.card), tt.name+" (1/3 単位)")
	}
}

// **精算はゼロサム。** 誰かの取り分が増えれば同じだけ誰かが減る。
func TestPiedmonteseTarot_DealSettlementIsZeroSum(t *testing.T) {
	t.Parallel()
	for _, thirds := range [][]int{
		{78, 78, 78},
		{100, 60, 74},
		{0, 234, 0},
		{60, 60, 60, 54},
	} {
		scores := piedmonteseTarotSettleDeal(thirds)
		sum := 0
		for _, s := range scores {
			sum += s
		}
		assert.Zero(t, sum, "%v の精算が釣り合っていない", thirds)
		assert.Len(t, scores, len(thirds))
	}
	// 取り分が多い席ほど大きい。
	scores := piedmonteseTarotSettleDeal([]int{100, 60, 74})
	assert.Greater(t, scores[0], scores[2])
	assert.Greater(t, scores[2], scores[1])
}

// **1 ディールで 78 枚すべてがどこかの席に入る。** 親のスカルトも親の獲得に
// 数えるので、合計は必ず 234 thirds になる。
func TestPiedmonteseTarot_EveryCardIsAccountedForAfterADeal(t *testing.T) {
	t.Parallel()
	for _, seats := range PiedmonteseTarotSeatSizes {
		for range 10 {
			g := newPiedmonteseTarotForTest(t, seats)
			piedmonteseTarotPlayHand(t, g)
			require.Contains(t,
				[]PiedmonteseTarotPhase{PiedmonteseTarotPhaseRoundEnd, PiedmonteseTarotPhaseGameEnd},
				g.GetPhase())

			total := 0
			for _, v := range g.CapturedThirds() {
				total += v
			}
			assert.Equal(t, PiedmonteseTarotTotalThirds, total, "%d 人卓で札の取り分が合わない", seats)

			sum := 0
			for _, v := range g.GetDealScores() {
				sum += v
			}
			assert.Zero(t, sum, "%d 人卓のディール精算が釣り合っていない", seats)
		}
	}
}

// **Matto は取られない。** 出した本人の獲得札に残り、トリックは他の札で決まる。
func TestPiedmonteseTarot_TheMattoStaysWithItsOwner(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.phase = PiedmonteseTarotPhasePlay
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(Tarot78ExcuseDesign, Tarot78ExcuseValue, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 9, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 2, false)},
	}
	g.trickNumber = 1
	g.phase = PiedmonteseTarotPhaseTrickEnd
	for _, p := range g.GetPlayers() {
		p.ResetRound()
	}
	g.ResolveTrick()

	assert.Equal(t, 3, countPiedmonteseTarotCards(g.GetPlayers()[2]), "♥9 の席が 3 枚取る")
	assert.Equal(t, 1, countPiedmonteseTarotCards(g.GetPlayers()[1]), "Matto は出した本人に残る")
	assert.Zero(t, countPiedmonteseTarotCards(g.GetPlayers()[0]))
}

// countPiedmonteseTarotCards は獲得トリックの札数を返す。
func countPiedmonteseTarotCards(p *PiedmonteseTarotPlayer) int {
	n := 0
	for _, trick := range p.GetTricksTaken() {
		n += len(trick)
	}
	return n
}

// **最強切り札が取る。** スートに従えなかった席が切り札を出した瞬間、
// リードスートの最強札では勝てなくなる。
func TestPiedmonteseTarot_TrumpsBeatTheLedSuit(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, Tarot78KingValue, false)},
		{PlayerIdx: 1, Card: NewCard(Tarot78TrumpDesign, 2, false)},
		{PlayerIdx: 2, Card: NewCard(Tarot78TrumpDesign, 15, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 3, false)},
	}
	assert.Equal(t, 2, g.trickWinner(), "切り札 15 が取る")
	assert.Equal(t, CardDesignHeart, tarot78LedSuit(g.currentTrick))
	assert.Equal(t, 15, tarot78HighestTrumpInTrick(g.currentTrick))
}

// **点になる札は捨てられない。** 親がオヌールやコート札を捨てられると、
// 自分の取り分をそのまま増やせてしまう。
func TestPiedmonteseTarot_ScartoRefusesScoringCards(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	dealer := g.GetPlayers()[g.GetDealerIdx()]
	dealer.Reset()
	for _, c := range []*Card{
		NewCard(Tarot78TrumpDesign, Tarot78MaxTrump, false), // Mondo
		NewCard(CardDesignHeart, Tarot78KingValue, false),   // Roi
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignClover, 5, false),
	} {
		dealer.AddCard(c)
	}

	assert.ErrorIs(t, g.validateScarto(dealer, []int{0, 2}), ErrInvalidPlay, "オヌールを捨てさせない")
	assert.ErrorIs(t, g.validateScarto(dealer, []int{1, 2}), ErrInvalidPlay, "コート札を捨てさせない")
	assert.NoError(t, g.validateScarto(dealer, []int{2, 3}), "ピップ 2 枚は捨てられる")
}

// **捨てられるピップが足りなければ切り札を許す。** ここを閉じると、親が
// 何も捨てられずディールが始まらない。
func TestPiedmonteseTarot_ScartoFallsBackToTrumpsWhenNoPipsRemain(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	dealer := g.GetPlayers()[g.GetDealerIdx()]
	dealer.Reset()
	for _, c := range []*Card{
		NewCard(Tarot78TrumpDesign, 5, false),
		NewCard(Tarot78TrumpDesign, 6, false),
		NewCard(CardDesignHeart, Tarot78KingValue, false),
		NewCard(CardDesignSpade, 12, false),
	} {
		dealer.AddCard(c)
	}
	assert.NoError(t, g.validateScarto(dealer, []int{0, 1}), "ピップが無ければ切り札を捨てられる")
	assert.ErrorIs(t, g.validateScarto(dealer, []int{0, 2}), ErrInvalidPlay, "それでもコート札は捨てない")
}

// 捨てる枚数はタロンぴったり。
func TestPiedmonteseTarot_ScartoNeedsExactlyTheTalonCount(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.dealerIdx = findHumanIdx(g.GetPlayers())
	require.True(t, g.IsHumanScartoTurn())
	assert.ErrorIs(t, g.PlayerScarto([]int{0}), ErrInvalidCard, "1 枚では足りない")
	assert.ErrorIs(t, g.PlayerScarto([]int{0, 1, 2}), ErrInvalidCard, "3 枚は多い")
	assert.ErrorIs(t, g.PlayerScarto([]int{0, 0}), ErrInvalidCard, "同じ札を 2 回選べない")
	assert.ErrorIs(t, g.PlayerScarto([]int{0, 999}), ErrInvalidCard, "範囲外")
}

// **捨てた札は親の取り分になる。** 数えないと、その札の点がどこにも行かず
// 合計が 78 点に届かない。
func TestPiedmonteseTarot_TheDiscardCountsForTheDealer(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.scarto = []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}
	thirds := g.CapturedThirds()
	assert.Equal(t, 2*piedmonteseTarotCardThirds(g.scarto[0]), thirds[g.GetDealerIdx()])
}

// **フォロー義務。** 持っているのに別のスートを出せてしまうと、規則が無い。
func TestPiedmonteseTarot_MustFollowSuitOrTrump(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	p := g.GetPlayers()[0]
	p.Reset()
	for _, c := range []*Card{
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(Tarot78TrumpDesign, 3, false),
		NewCard(Tarot78ExcuseDesign, Tarot78ExcuseValue, false),
	} {
		p.AddCard(c)
	}
	g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 10, false)}}

	valid := g.GetPlayableIndices(0)
	assert.ElementsMatch(t, []int{0, 3}, valid, "♥ を持っていれば ♥ か Matto だけ")

	// ♥ を抜くと切り札義務になる。
	p.Reset()
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(Tarot78TrumpDesign, 3, false),
		NewCard(Tarot78ExcuseDesign, Tarot78ExcuseValue, false),
	} {
		p.AddCard(c)
	}
	assert.ElementsMatch(t, []int{1, 2}, g.GetPlayableIndices(0), "ボイドなら切り札を出す")
}

// **切り札が出ていれば上位で応じる。** 下位で逃げられると、切り札の駆け引きが
// 成立しない。
func TestPiedmonteseTarot_MustOvertrumpWhenPossible(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	p := g.GetPlayers()[0]
	p.Reset()
	for _, c := range []*Card{
		NewCard(Tarot78TrumpDesign, 2, false),
		NewCard(Tarot78TrumpDesign, 12, false),
		NewCard(CardDesignSpade, 9, false),
	} {
		p.AddCard(c)
	}
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: NewCard(Tarot78TrumpDesign, 8, false)},
	}
	assert.Equal(t, []int{1}, g.GetPlayableIndices(0), "8 を超える切り札があるならそれしか出せない")
}

// 保存の往復で盤面が保たれる。
func TestPiedmonteseTarot_SurvivesASaveAndRestore(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	piedmonteseTarotPlayHand(t, g)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(PiedmonteseTarot)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPlayerScores(), restored.GetPlayerScores())
	assert.Equal(t, g.GetDealScores(), restored.GetDealScores())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, g.CapturedThirds(), restored.CapturedThirds())
}

// **壊れた保存は受け取らない。** 席数と得点表が食い違う盤は、次の 1 手で
// 範囲外を読む。
func TestPiedmonteseTarot_RestoreRejectsATamperedBoard(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	for name, mutate := range map[string]func(map[string]json.RawMessage){
		"席数が設定と違う":  func(m map[string]json.RawMessage) { m["cf"] = json.RawMessage(`{"s":3,"cd":1,"td":2}`) },
		"得点表が短い":    func(m map[string]json.RawMessage) { m["ps"] = json.RawMessage(`[0,0]`) },
		"フェーズが範囲外":  func(m map[string]json.RawMessage) { m["ph"] = json.RawMessage(`9`) },
		"親が範囲外":     func(m map[string]json.RawMessage) { m["di"] = json.RawMessage(`9`) },
		"手番が範囲外":    func(m map[string]json.RawMessage) { m["cp"] = json.RawMessage(`-1`) },
		"ディール番号が 0": func(m map[string]json.RawMessage) { m["rn"] = json.RawMessage(`0`) },
	} {
		t.Run(name, func(t *testing.T) {
			clone := make(map[string]json.RawMessage, len(raw))
			for k, v := range raw {
				clone[k] = v
			}
			mutate(clone)
			body, err := json.Marshal(clone)
			require.NoError(t, err)
			assert.Error(t, json.Unmarshal(body, new(PiedmonteseTarot)))
		})
	}
}

// **設定の検証。** 5 人卓は配り方が無いので作らせない。
func TestPiedmonteseTarotConfig_Validate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultPiedmonteseTarotConfig().Validate())
	for _, seats := range []int{0, 2, 5, 9} {
		cfg := DefaultPiedmonteseTarotConfig()
		cfg.Seats = seats
		assert.Error(t, cfg.Validate(), "%d 人卓を通してしまう", seats)
	}
	cfg := DefaultPiedmonteseTarotConfig()
	cfg.TargetDeals = 0
	assert.Error(t, cfg.Validate())
	cfg = DefaultPiedmonteseTarotConfig()
	cfg.CpuDifficulty = PiedmonteseTarotCpuDifficulty(9)
	assert.Error(t, cfg.Validate())
}

// **マッチは規定ディールで終わる。** 終わらないと、いつまでも次が配られる。
func TestPiedmonteseTarot_MatchEndsAfterTheTargetDeals(t *testing.T) {
	t.Parallel()
	cfg := DefaultPiedmonteseTarotConfig()
	cfg.TargetDeals = 2
	g := NewPiedmonteseTarot(newPiedmonteseTarotPlayers(cfg.Seats), cfg)
	g.Reset()
	for deal := 0; deal < cfg.TargetDeals; deal++ {
		piedmonteseTarotPlayHand(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.Equal(t, PiedmonteseTarotPhaseRoundEnd, g.GetPhase())
		g.NextRound()
	}
	assert.True(t, g.GetGameEndFlag(), "規定ディールを打っても終わらない")
	assert.Equal(t, PiedmonteseTarotPhaseGameEnd, g.GetPhase())
	assert.Contains(t, []PiedmonteseTarotResult{
		PiedmonteseTarotResultWin, PiedmonteseTarotResultLose, PiedmonteseTarotResultNone,
	}, g.GetResult())
}

// **親は 1 席ずつ回る。** 動かないと、同じ席だけがタロンを拾い続ける。
func TestPiedmonteseTarot_TheDealerRotates(t *testing.T) {
	t.Parallel()
	cfg := DefaultPiedmonteseTarotConfig()
	cfg.TargetDeals = 4
	g := NewPiedmonteseTarot(newPiedmonteseTarotPlayers(cfg.Seats), cfg)
	g.Reset()
	first := g.GetDealerIdx()
	piedmonteseTarotPlayHand(t, g)
	require.False(t, g.GetGameEndFlag())
	g.NextRound()
	assert.Equal(t, (first+1)%cfg.Seats, g.GetDealerIdx())
}

// ヒントはフェーズごとに違う手を勧める。
func TestPiedmonteseTarot_HintFollowsThePhase(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.dealerIdx = findHumanIdx(g.GetPlayers())
	require.True(t, g.IsHumanScartoTurn())
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "scarto_weak", hint.Reason)
	assert.Len(t, hint.CardIndices, g.TalonSize(), "捨てる枚数ぶん勧める")

	g.CpuScarto()
	require.NoError(t, g.PlayerScarto(hint.CardIndices))
	for g.GetPhase() == PiedmonteseTarotPhasePlay && !g.IsHumanTurn() {
		g.CpuPlay()
	}
	if g.GetPhase() == PiedmonteseTarotPhasePlay {
		playHint := g.GetHint()
		assert.Len(t, playHint.CardIndices, 1)
		assert.Contains(t, g.GetPlayableIndices(g.GetCurrentPlayerIdx()), playHint.CardIndices[0],
			"勧める札が出せない札だった")
	}
}

// **既定の卓から全部見る。** 登録はこの入口から作るので、ここが配れないと
// ゲーム一覧に並んだ瞬間に落ちる。参照系も画面が読む値なので、まとめて確かめる。
func TestPiedmonteseTarot_DefaultTableAndAccessors(t *testing.T) {
	t.Parallel()
	g := NewDefaultPiedmonteseTarot()
	g.Reset()

	assert.Equal(t, PiedmonteseTarotDefaultSeats, g.GetPlayerCnt())
	assert.Equal(t, DefaultPiedmonteseTarotConfig(), g.GetConfig())
	assert.Equal(t, 19, g.HandSize())
	assert.Equal(t, 2, g.TalonSize())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Zero(t, g.GetTrickNumber(), "スカルトの前はトリックが始まっていない")
	assert.Empty(t, g.GetCurrentTrick())
	assert.Empty(t, g.GetScarto())
	assert.Zero(t, g.GetScartoCount())
	assert.Equal(t, -1, g.GetLastTrickWinner())
	assert.Equal(t, -1, g.GetWinnerPlayer())
	assert.Equal(t, PiedmonteseTarotOutcomeNone, g.GetOutcome())
	assert.Equal(t, PiedmonteseTarotResultNone, g.GetResult())
	assert.False(t, g.GetGameEndFlag())
	assert.Len(t, g.GetPlayerScores(), PiedmonteseTarotDefaultSeats)
	assert.Len(t, g.GetDealScores(), PiedmonteseTarotDefaultSeats)

	// 範囲外は nil / 0 を返す。**落ちない**ことが要件。
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(g.GetPlayerCnt()))
	assert.NotNil(t, g.GetPlayer(0))
	assert.Zero(t, g.GetCardThirds(-1))
	assert.Zero(t, g.GetCardThirds(99))
	assert.Nil(t, g.GetPlayableIndices(-1))
	assert.Nil(t, g.GetPlayableIndices(99))

	// 設定は差し替えられる。
	cfg := g.GetConfig()
	cfg.TargetDeals = 3
	g.SetConfig(cfg)
	assert.Equal(t, 3, g.GetConfig().TargetDeals)
}

// **CPU だけでマッチを終わりまで回せる。** 人間が降りたあとも卓が進むことと、
// CPU の戦略分岐 (取りにいく / 安く落とす / リード) を通す。
func TestPiedmonteseTarot_CpuPlaysAMatchToTheEnd(t *testing.T) {
	t.Parallel()
	for _, seats := range PiedmonteseTarotSeatSizes {
		cfg := DefaultPiedmonteseTarotConfig()
		cfg.Seats = seats
		cfg.TargetDeals = 2
		g := NewPiedmonteseTarot(newPiedmonteseTarotPlayers(seats), cfg)
		g.Reset()

		for step := 0; step < 6000 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case PiedmonteseTarotPhaseScarto:
				if g.IsHumanScartoTurn() {
					require.NoError(t, g.PlayerScarto(g.cpuSelectScarto(g.GetDealerIdx())))
					continue
				}
				g.CpuScarto()
			case PiedmonteseTarotPhasePlay:
				if g.IsHumanTurn() {
					require.NoError(t, g.PlayerPlay(g.cpuSelectPlayCard(g.GetCurrentPlayerIdx())))
					continue
				}
				g.CpuPlay()
			case PiedmonteseTarotPhaseTrickEnd:
				g.ResolveTrick()
				if g.GetPhase() == PiedmonteseTarotPhaseTrickEnd {
					g.NextTrick()
				}
			case PiedmonteseTarotPhaseRoundEnd:
				// **ScoreRound は二度呼んでも増えない。**
				before := append([]int(nil), g.GetPlayerScores()...)
				g.ScoreRound()
				assert.Equal(t, before, g.GetPlayerScores(), "精算が二重に走っている")
				g.NextRound()
			default:
				require.FailNow(t, "unexpected phase")
			}
		}
		require.True(t, g.GetGameEndFlag(), "%d 人卓のマッチが終わらない", seats)
		assert.Equal(t, PiedmonteseTarotPhaseGameEnd, g.GetPhase())
		assert.GreaterOrEqual(t, g.GetLastTrickWinner(), 0, "最後のトリックの勝者が記録されていない")
		assert.Contains(t, []PiedmonteseTarotOutcome{
			PiedmonteseTarotOutcomeWin, PiedmonteseTarotOutcomeLoss, PiedmonteseTarotOutcomeNone,
		}, g.GetOutcome())
	}
}

// **終わったマッチと違うフェーズは断る。** 受け付けると、決着後の盤面が動く。
func TestPiedmonteseTarot_RefusesPlaysOutsideThePhase(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrWrongPhase, "スカルト中に札は出せない")

	g.phase = PiedmonteseTarotPhasePlay
	g.currentPlayerIdx = (findHumanIdx(g.GetPlayers()) + 1) % g.GetPlayerCnt()
	assert.ErrorIs(t, g.PlayerPlay(0), ErrNotHumanTurn)

	g.currentPlayerIdx = findHumanIdx(g.GetPlayers())
	assert.ErrorIs(t, g.PlayerPlay(-1), ErrInvalidCard)
	assert.ErrorIs(t, g.PlayerPlay(999), ErrInvalidCard)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerScarto(nil), ErrGameEnded)
	g.CpuScarto() // 終わった卓では何も起きない
	g.CpuPlay()
	assert.True(t, g.GetGameEndFlag())
}

// **札の綴りは棋譜に出る。** 切り札と Matto を数札と同じ形で書くと、
// 棋譜からどの札か読めない。
func TestPiedmonteseTarot_CardNames(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "??", piedmonteseTarotCardStr(nil))
	assert.Equal(t, "Matto", piedmonteseTarotCardStr(NewCard(Tarot78ExcuseDesign, 0, false)))
	assert.Equal(t, "T21", piedmonteseTarotCardStr(NewCard(Tarot78TrumpDesign, 21, false)))
	assert.Equal(t, "♠5", piedmonteseTarotCardStr(NewCard(CardDesignSpade, 5, false)))
	assert.Equal(t, "♥14", piedmonteseTarotCardStr(NewCard(CardDesignHeart, Tarot78KingValue, false)))
	assert.Equal(t, "?3", piedmonteseTarotCardStr(NewCard(99, 3, false)))
}

// **ヒントは席を見る。** 自分の手番でないときに勧めると、押せないボタンを
// 光らせることになる。
func TestPiedmonteseTarot_HintStaysQuietOffTurn(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	g.dealerIdx = (findHumanIdx(g.GetPlayers()) + 1) % g.GetPlayerCnt()
	assert.Equal(t, "none", g.GetHint().Reason, "CPU のスカルトを人間に勧めている")

	g.phase = PiedmonteseTarotPhasePlay
	g.currentPlayerIdx = (findHumanIdx(g.GetPlayers()) + 1) % g.GetPlayerCnt()
	assert.Equal(t, "none", g.GetHint().Reason)

	g.phase = PiedmonteseTarotPhaseTrickEnd
	assert.Equal(t, "next_trick", g.GetHint().Reason)
	g.phase = PiedmonteseTarotPhaseRoundEnd
	assert.Equal(t, "next_round", g.GetHint().Reason)
	g.gameEndFlag = true
	assert.Equal(t, "none", g.GetHint().Reason)
}

// **切り札がリードされた形も見る。** 上位切り札の義務はそこでしか働かない。
func TestPiedmonteseTarot_TrumpLeadForcesAHigherTrump(t *testing.T) {
	t.Parallel()
	g := newPiedmonteseTarotForTest(t, 4)
	p := g.GetPlayers()[0]
	p.Reset()
	for _, c := range []*Card{
		NewCard(Tarot78TrumpDesign, 3, false),
		NewCard(Tarot78TrumpDesign, 18, false),
		NewCard(CardDesignHeart, 4, false),
	} {
		p.AddCard(c)
	}
	g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(Tarot78TrumpDesign, 10, false)}}
	assert.Equal(t, []int{1}, g.GetPlayableIndices(0), "10 を超える切り札があるならそれだけ")

	// 上位が無ければ持っている切り札を出す。
	p.Reset()
	for _, c := range []*Card{
		NewCard(Tarot78TrumpDesign, 3, false),
		NewCard(CardDesignHeart, 4, false),
	} {
		p.AddCard(c)
	}
	assert.Equal(t, []int{0}, g.GetPlayableIndices(0))
}

// 点の書式は 1/3 単位で読める形になる (画面と棋譜が同じ数を出す)。
func TestPiedmonteseTarot_FormatThirds(t *testing.T) {
	t.Parallel()
	for thirds, want := range map[int]string{
		0: "0", 1: "0 1/3", 2: "0 2/3", 3: "1", 78: "26", 79: "26 1/3", 80: "26 2/3",
		-3: "-1", -5: "-1 2/3",
	} {
		assert.Equal(t, want, PiedmonteseTarotFormatThirds(thirds), "%d thirds", thirds)
	}
}
