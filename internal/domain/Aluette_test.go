//go:build test

package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func aluetteCard(design, value int) *Card { return NewCard(design, value, false) }

func aluetteTrick(entries ...[2]int) []*TrickCard {
	trick := make([]*TrickCard, 0, len(entries))
	for seat, e := range entries {
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: aluetteCard(e[0], e[1])})
	}
	return trick
}

func TestAluette_DeckIsFortyEightDistinctCards(t *testing.T) {
	deck := buildAluetteDeck()
	assert.Len(t, deck, AluetteDeckSize)
	assert.Equal(t, 48, AluetteDeckSize)

	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		// **8 と 9 を含む。**40 枚デッキ (Tute/Briscola) の感覚で作ると、
		// 9 のうち 2 枚がリュエットなので序列表が壊れる。
		assert.NotEqual(t, 10, c.GetValue(), "the Latin deck has no 10")
	}
	for _, v := range []int{8, 9} {
		assert.True(t, seen[[2]int{1, v}], "value %d must exist in every suit", v)
	}
}

// **これがこのゲームの核心。**強さはランクではなく「特定の 1 枚」で決まる。
// 同じ value 3 でも、金貨の3 (Monsieur) は最強で、剣の3 は普通の 3 でしかない。
func TestAluette_RankIsPerCardNotPerValue(t *testing.T) {
	monsieur := aluetteCard(4, 3)   // 金貨の3
	plainThree := aluetteCard(1, 3) // 剣の3
	assert.Equal(t, "Monsieur", AluetteLuetteName(monsieur))
	assert.Empty(t, AluetteLuetteName(plainThree))
	assert.Greater(t, AluetteRank(monsieur), AluetteRank(plainThree),
		"the same value must not imply the same strength")
}

// リュエット 6 枚は名前の順に最上位を占める。
func TestAluette_LuettesOutrankEverythingInOrder(t *testing.T) {
	order := []*Card{
		aluetteCard(4, 3), // Monsieur
		aluetteCard(3, 3), // Madame
		aluetteCard(3, 2), // Borgne
		aluetteCard(4, 2), // Vache
		aluetteCard(3, 9), // GrandNeuf
		aluetteCard(4, 9), // PetitNeuf
	}
	names := []string{"Monsieur", "Madame", "Borgne", "Vache", "GrandNeuf", "PetitNeuf"}
	for i, c := range order {
		assert.Equal(t, names[i], AluetteLuetteName(c))
		if i > 0 {
			assert.Less(t, AluetteRank(c), AluetteRank(order[i-1]),
				"%s must be weaker than %s", names[i], names[i-1])
		}
	}
	// 最弱のリュエットでも、最強の通常札 (剣の3) より強い。
	assert.Greater(t, AluetteRank(order[len(order)-1]), AluetteRank(aluetteCard(1, 3)))
}

// 通常札はスートを一切見ず、値だけで 3 > 2 > A > 王 > 騎 > 従 > 9 > 8 … と並ぶ。
func TestAluette_OrdinaryOrderIgnoresSuit(t *testing.T) {
	desc := []int{3, 2, 1, 13, 12, 11, 9, 8, 7, 6, 5, 4}
	for i := 1; i < len(desc); i++ {
		// 剣 (design 1) はリュエットを 1 枚も含まないので、通常序列の確認に使える。
		hi, lo := aluetteCard(1, desc[i-1]), aluetteCard(1, desc[i])
		assert.Greater(t, AluetteRank(hi), AluetteRank(lo),
			"value %d must outrank %d", desc[i-1], desc[i])
	}
	// スートが違っても同値なら同ランク。
	assert.Equal(t, AluetteRank(aluetteCard(1, 7)), AluetteRank(aluetteCard(2, 7)))
}

func TestAluette_TrickWinner(t *testing.T) {
	cases := []struct {
		name  string
		trick []*TrickCard
		want  int
	}{
		{"the strongest card wins regardless of what was led", aluetteTrick(
			[2]int{1, 4}, [2]int{2, 13}, [2]int{1, 5}, [2]int{2, 6},
		), 1},
		{"a luette beats every ordinary card", aluetteTrick(
			[2]int{1, 3}, [2]int{2, 3}, [2]int{4, 9}, [2]int{1, 2},
		), 2},
		{"Monsieur beats Madame", aluetteTrick(
			[2]int{3, 3}, [2]int{4, 3}, [2]int{1, 1}, [2]int{2, 1},
		), 1},
		// **同ランクは先に出した方が勝つ。**後から同じ強さを重ねても奪えない。
		{"an equal rank played later does not take it", aluetteTrick(
			[2]int{1, 7}, [2]int{2, 7}, [2]int{3, 7}, [2]int{4, 7},
		), 0},
		{"an empty trick does not panic", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, aluetteTrickWinnerOf(tc.trick))
		})
	}
}

// 切り札は存在しない。スートは強さに一切関与しない。
func TestAluette_NoTrumpSuitExists(t *testing.T) {
	for suit := 1; suit <= AluetteSuitCnt; suit++ {
		// 同じ 4 (最弱の通常札) はどのスートでも同じ強さ。
		assert.Equal(t, AluetteRank(aluetteCard(1, 4)), AluetteRank(aluetteCard(suit, 4)),
			"suit %d must not confer any strength", suit)
	}
}

func TestAluette_TeamsAreOpposite(t *testing.T) {
	assert.Equal(t, AluetteTeamOf(0), AluetteTeamOf(2))
	assert.Equal(t, AluetteTeamOf(1), AluetteTeamOf(3))
	assert.NotEqual(t, AluetteTeamOf(0), AluetteTeamOf(1))
}

func TestAluette_DealArithmetic(t *testing.T) {
	assert.Equal(t, 5, AluetteHandSize)
	assert.Equal(t, AluetteHandSize, AluetteTrickCount)
	// 5 戦 3 勝。過半数であることを式で固定しておく。
	assert.Equal(t, 3, AluetteTricksToWin)
	assert.Greater(t, AluetteTricksToWin*2, AluetteTrickCount, "winning must require a majority")
	// 配り切らない。残りはそのメーヌでは使わない。
	assert.Less(t, AluettePlayerCnt*AluetteHandSize, AluetteDeckSize)
}

func TestAluette_RankHandlesNil(t *testing.T) {
	assert.Equal(t, -1, AluetteRank(nil))
	assert.Empty(t, AluetteLuetteName(nil))
}

// newTestAluette は配り札を固定した Aluette を返す (#4467)。
func newTestAluette(t *testing.T) *Aluette {
	t.Helper()
	g := NewDefaultAluette()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

// **SetRand を呼ばずに作る。**helper が上書きすると、コンストラクタが乱数源を
// 入れ忘れていても全テストが通ってしまう (Ganjifa #4661)。
func TestAluette_ProductionConstructorShufflesTheDeck(t *testing.T) {
	handOf := func(g *Aluette) string {
		p := g.GetPlayer(0)
		var b strings.Builder
		for i := 0; i < p.GetCardsSize(); i++ {
			fmt.Fprintf(&b, "%d-%d,", p.GetCard(i).GetDesign(), p.GetCard(i).GetValue())
		}
		return b.String()
	}
	first, second := NewDefaultAluette(), NewDefaultAluette()
	first.Reset()
	second.Reset()
	assert.NotEqual(t, handOf(first), handOf(second),
		"two fresh games dealt identical hands -- the constructor did not seed rng")
}

func TestAluette_ResetDealsTheHandSizeAndLeavesTheRest(t *testing.T) {
	g := newTestAluette(t)
	assert.Equal(t, AluettePhasePlay, g.GetPhase())
	for i := 0; i < AluettePlayerCnt; i++ {
		assert.Equal(t, AluetteHandSize, g.GetPlayer(i).GetCardsSize(), "seat %d", i)
	}
}

// 手札は値順ではなく強さ順に並ぶ。値順だと金貨の3 が剣の3 と同じ位置に来る。
func TestAluette_HandIsSortedByStrengthNotValue(t *testing.T) {
	g := newTestAluette(t)
	p := g.GetPlayer(0)
	for i := 1; i < p.GetCardsSize(); i++ {
		assert.GreaterOrEqual(t, AluetteRank(p.GetCard(i-1)), AluetteRank(p.GetCard(i)),
			"the hand must read strongest-first")
	}
}

// **フォロー義務は無い。**どの札もいつでも出せる。
func TestAluette_EveryCardIsAlwaysPlayable(t *testing.T) {
	g := newTestAluette(t)
	g.SetCurrentPlayerIdx(0)
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: aluetteCard(1, 6)}}
	valid := g.GetValidPlayIndices(0)
	assert.Len(t, valid, g.GetPlayer(0).GetCardsSize(),
		"no follow obligation exists, so the whole hand stays legal")
	assert.NoError(t, g.PlayerPlay(valid[len(valid)-1]),
		"even the weakest card must be legal while a suit is led")
}

func TestAluette_PlayGuards(t *testing.T) {
	g := newTestAluette(t)
	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(999))

	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrNotHumanTurn)

	g.SetPhase(AluettePhaseTrickEnd)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrWrongPhase)

	g.SetPhase(AluettePhasePlay)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
}

func TestAluette_CpuPlayWithAnEmptyHandIsANoOp(t *testing.T) {
	g := newTestAluette(t)
	g.SetCurrentPlayerIdx(1)
	p := g.GetPlayer(1)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	assert.NotPanics(t, func() { g.CpuPlay() })
	assert.Empty(t, g.GetCurrentTrick())
}

func TestAluette_FullMeneAccountsForEveryTrick(t *testing.T) {
	g := newTestAluette(t)
	for range AluetteTrickCount {
		for range AluettePlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		}
		require.Equal(t, AluettePhaseTrickEnd, g.GetPhase())
		g.ResolveTrick()
		if g.GetPhase() == AluettePhaseTrickEnd {
			g.NextTrick()
		}
	}
	require.Equal(t, AluettePhaseRoundEnd, g.GetPhase())
	tricks := 0
	for _, n := range g.GetRoundTricks() {
		tricks += n
	}
	assert.Equal(t, AluetteTrickCount, tricks)
	for i := 0; i < AluettePlayerCnt; i++ {
		assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "seat %d should be out of cards", i)
	}
	// **メーヌを取った側に 1 点。**4-1 でも 3-2 でも 1 点で変わらない。
	assert.Equal(t, 1, g.GetTeamScores()[0]+g.GetTeamScores()[1])
}

// 過半数に届かないメーヌ (同点) は誰にも点が入らない。
func TestAluette_ASplitMeneScoresForNobody(t *testing.T) {
	g := newTestAluette(t)
	g.SetPhase(AluettePhaseTrickEnd)
	g.trickNumber = AluetteTrickCount
	// 2-2 で 1 トリック足りない状態を作る (5 戦のうち 4 戦しか決着していない)。
	g.roundTricks = [AluettePlayerCnt]int{1, 1, 1, 1}
	g.settleRound()
	assert.Equal(t, 0, g.GetTeamScores()[0]+g.GetTeamScores()[1],
		"neither side reached the majority, so no point is awarded")
}

func TestAluette_MatchEndsAtTheTargetPoints(t *testing.T) {
	g := newTestAluette(t)
	g.SetPhase(AluettePhaseRoundEnd)
	g.teamScores = [2]int{g.GetConfig().TargetPoints, 0}
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
}

func TestAluette_MatchWinnerIsATeam(t *testing.T) {
	cases := map[string]struct {
		scores [2]int
		want   int
	}{
		"team 0 ahead":         {[2]int{6, 4}, 0},
		"team 1 ahead":         {[2]int{4, 6}, 1},
		"a draw has no winner": {[2]int{5, 5}, -1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newTestAluette(t)
			g.teamScores = tc.scores
			g.finishMatch()
			assert.Equal(t, tc.want, g.GetWinnerTeam())
			assert.True(t, g.GetGameEndFlag())
		})
	}
}

func TestAluette_NextRoundRotatesTheDealer(t *testing.T) {
	g := newTestAluette(t)
	g.SetPhase(AluettePhaseRoundEnd)
	dealer, round := g.GetDealerIdx(), g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, round+1, g.GetRoundNumber())
	assert.Equal(t, (dealer+1)%AluettePlayerCnt, g.GetDealerIdx())
	assert.Equal(t, AluettePhasePlay, g.GetPhase())
}

func TestAluette_NoOpsOutsideTheirPhase(t *testing.T) {
	g := newTestAluette(t)
	g.ResolveTrick()
	g.NextTrick()
	assert.Equal(t, AluettePhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())

	g.NextRound()
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestAluette_HintOnlyOnTheHumanTurn(t *testing.T) {
	g := newTestAluette(t)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint(), "no hint on a CPU turn")

	g.SetCurrentPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)

	g.SetPhase(AluettePhaseRoundEnd)
	assert.Nil(t, g.GetHint())
}

// リュエットを勧めるときは専用の理由を返す。名前つきの 6 枚は見た目が他と
// 変わらないので、「なぜその札か」を言わないと助言が読めない。
func TestAluette_HintNamesALuette(t *testing.T) {
	g := newTestAluette(t)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(aluetteCard(4, 3)) // Monsieur
	g.currentTrick = nil
	assert.Equal(t, "play_luette", g.playHintReason(0, 0))

	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(aluetteCard(1, 4))
	assert.Equal(t, "lead_low", g.playHintReason(0, 0))
}

func TestAluette_Accessors(t *testing.T) {
	g := newTestAluette(t)
	assert.Equal(t, AluettePlayerCnt, g.GetPlayerCnt())
	assert.Len(t, g.GetPlayers(), AluettePlayerCnt)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(AluettePlayerCnt))
	assert.Equal(t, -1, g.GetLastTrickWinner())
	assert.Equal(t, (g.GetDealerIdx()+1)%AluettePlayerCnt, g.GetLeadPlayerIdx(),
		"the seat left of the dealer leads the first trick")
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.NotEmpty(t, g.GetPlayableIndices(0))

	cfg := AluetteConfig{CpuDifficulty: AluetteCpuDifficultyHard, TargetPoints: 10}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	assert.Empty(t, g.GetActionLog())
}

func TestAluette_ConfigValidation(t *testing.T) {
	assert.NoError(t, DefaultAluetteConfig().Validate())
	assert.Error(t, AluetteConfig{CpuDifficulty: 99, TargetPoints: 6}.Validate())
	assert.Error(t, AluetteConfig{TargetPoints: 0}.Validate())
}

// Worker のセッションは毎リクエスト JSON を往復する。
func TestAluette_SurvivesRoundTrippingEveryRequest(t *testing.T) {
	g := NewDefaultAluette()
	g.SetRand(rand.New(rand.NewSource(7)))
	g.Reset()
	roundTrip := func(src *Aluette) *Aluette {
		t.Helper()
		data, err := src.MarshalJSON()
		require.NoError(t, err)
		var out Aluette
		require.NoError(t, out.UnmarshalJSON(data))
		return &out
	}
	g = roundTrip(g)
	for range AluetteTrickCount {
		for range AluettePlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
			g = roundTrip(g)
		}
		g.ResolveTrick()
		g = roundTrip(g)
		if g.GetPhase() == AluettePhaseTrickEnd {
			g.NextTrick()
			g = roundTrip(g)
		}
	}
	tricks := 0
	for _, n := range g.GetRoundTricks() {
		tricks += n
	}
	require.Equal(t, AluetteTrickCount, tricks, "tricks were lost across the round trips")
}

// **復元後に SetRand を呼ばずに Easy の CPU を回す。**Worker の復元経路には
// その一行が無い (#4663)。
func TestAluette_RestoredGameSurvivesEasyCpu(t *testing.T) {
	src := NewDefaultAluette()
	src.Reset()
	cfg := src.GetConfig()
	cfg.CpuDifficulty = AluetteCpuDifficultyEasy
	src.SetConfig(cfg)
	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var restored Aluette
	require.NoError(t, restored.UnmarshalJSON(data))
	restored.SetCurrentPlayerIdx(1)
	assert.NotPanics(t, func() { restored.CpuPlay() })
	assert.Len(t, restored.GetCurrentTrick(), 1)
}

func TestAluette_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":                 `{`,
		"wrong player count":       `{"pl":[],"rn":1,"tn":1}`,
		"trick number too high":    `{"pl":[{},{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tp":6}}`,
		"nil trick card":           `{"pl":[{},{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tp":6}}`,
		"winner team out of range": `{"pl":[{},{},{},{}],"rn":1,"tn":1,"wt":5,"cfg":{"cd":1,"tp":6}}`,
		"invalid config":           `{"pl":[{},{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tp":0}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Aluette
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

// TestAluette_LuetteTableIsTheSameSourceAsRank はテーブル公開が序列とずれないことを見る。
//
// **UI に配る表とドメインの強さ判定が別々に育つのを防ぐ。**表の順序が
// AluetteRank の降順と一致していなければ、画面の序列表は嘘になる。
func TestAluette_LuetteTableIsTheSameSourceAsRank(t *testing.T) {
	table := AluetteLuetteTable()
	assert.Len(t, table, 6)

	prev := 1 << 30
	for _, l := range table {
		c := NewCard(l.Design, l.Value, true)
		assert.Equal(t, l.Name, AluetteLuetteName(c), "%s の同定がテーブルとずれている", l.Name)
		r := AluetteRank(c)
		assert.Less(t, r, prev, "%s がテーブル順どおりに弱くなっていない", l.Name)
		prev = r
	}

	// 返した表を書き換えても内部表は動かない。
	table[0].Name = "TAMPERED"
	assert.Equal(t, "Monsieur", AluetteLuetteTable()[0].Name)
}

// TestAluette_LeadPlayerIdxFollowsDealer は配り直しごとのリード席を見る。
func TestAluette_LeadPlayerIdxFollowsDealer(t *testing.T) {
	g := NewDefaultAluette()
	g.Reset()
	assert.Equal(t, (g.GetDealerIdx()+1)%AluettePlayerCnt, g.GetLeadPlayerIdx())
	assert.Equal(t, g.GetLeadPlayerIdx(), g.GetCurrentPlayerIdx())
}

// TestAluette_DefaultTargetPointsMatchesTheDocumentedValue は既定の到達点を固定する。
//
// **数値を誰も検査していないと黙ってずれる。**CUI は DefaultAluetteConfig を使う
// 一方、Web はフロントの既定を送るので、片方だけ動かしてもローカルでは気づけない
// (PR #4666 のレビュー指摘)。マニュアルと docs/games.md も 5 点先取と書いている。
func TestAluette_DefaultTargetPointsMatchesTheDocumentedValue(t *testing.T) {
	assert.Equal(t, 5, DefaultAluetteTargetPoints)
	assert.Equal(t, DefaultAluetteTargetPoints, DefaultAluetteConfig().TargetPoints)
	assert.NoError(t, DefaultAluetteConfig().Validate())
}

// TestAluette_HardKeepsItsLuettesForLaterTricks は Hard が Normal と別物であることを見る。
//
// **選べる難易度が何も変えないなら、それは嘘の選択肢。**強さが絶対なので
// リュエットはいつ出しても勝つ ——「3 で足りる場面で Monsieur を切らない」かどうかが
// 唯一の腕の差になる。
func TestAluette_HardKeepsItsLuettesForLaterTricks(t *testing.T) {
	setup := func(diff AluetteCpuDifficulty) *Aluette {
		g := NewDefaultAluette()
		cfg := DefaultAluetteConfig()
		cfg.CpuDifficulty = diff
		g.SetConfig(cfg)
		g.Reset()
		// 席 1 の手札を固定する。序列は 3 > 2 > A > 王 > 騎 > 従 > 9 > … > 4 なので、
		// リードされた 6 に勝てるのは Monsieur と剣の A の 2 枚。剣の 4 は負ける。
		p := g.players[1]
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
		p.AddCard(NewCard(4, 3, false)) // Monsieur — 最強
		p.AddCard(NewCard(1, 1, false)) // 剣の A — 勝てるが安い
		p.AddCard(NewCard(1, 4, false)) // 剣の 4 — 最弱、勝てない
		// 席 0 (相手チーム) が中位の札でリード済みの状態にする。
		g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(1, 6, false)}}
		g.currentPlayerIdx = 1
		return g
	}

	// Hard は 5 で足りるので Monsieur を温存する。
	hard := setup(AluetteCpuDifficultyHard)
	pick := hard.cpuSelectPlayCard(1, []int{0, 1, 2})
	assert.Equal(t, "", AluetteLuetteName(hard.players[1].GetCard(pick)),
		"Hard がリュエットを無駄打ちしている")
	assert.True(t, hard.winsTrick(1, hard.players[1].GetCard(pick)), "それでもトリックは取る")

	// Normal は最強札を出す。ここが両者の差。
	normal := setup(AluetteCpuDifficultyNormal)
	pick = normal.cpuSelectPlayCard(1, []int{0, 1, 2})
	assert.Equal(t, "Monsieur", AluetteLuetteName(normal.players[1].GetCard(pick)))
}

// TestAluette_HardLeadsLowToDrawOutStrength は Hard のリード方針を見る。
func TestAluette_HardLeadsLowToDrawOutStrength(t *testing.T) {
	g := NewDefaultAluette()
	cfg := DefaultAluetteConfig()
	cfg.CpuDifficulty = AluetteCpuDifficultyHard
	g.SetConfig(cfg)
	g.Reset()
	p := g.players[1]
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(4, 3, false)) // Monsieur
	p.AddCard(NewCard(1, 4, false)) // 最弱
	g.currentTrick = nil

	pick := g.cpuSelectPlayCard(1, []int{0, 1})
	assert.Equal(t, 4, p.GetCard(pick).GetValue(), "Hard がリードでリュエットを切っている")

	// Normal は逆に最強札でリードする。
	cfg.CpuDifficulty = AluetteCpuDifficultyNormal
	g.SetConfig(cfg)
	pick = g.cpuSelectPlayCard(1, []int{0, 1})
	assert.Equal(t, "Monsieur", AluetteLuetteName(p.GetCard(pick)))
}

// TestAluette_CpuNeverOverplaysItsPartner は 3 難易度すべてで味方を奪わないことを見る
// (Easy はランダムなので対象外)。
func TestAluette_CpuNeverOverplaysItsPartner(t *testing.T) {
	for _, diff := range []AluetteCpuDifficulty{AluetteCpuDifficultyNormal, AluetteCpuDifficultyHard} {
		g := NewDefaultAluette()
		cfg := DefaultAluetteConfig()
		cfg.CpuDifficulty = diff
		g.SetConfig(cfg)
		g.Reset()
		p := g.players[3]
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
		p.AddCard(NewCard(4, 3, false)) // Monsieur
		p.AddCard(NewCard(1, 4, false))
		// 席 1 (席 3 の味方) が Madame で勝っている。
		g.currentTrick = []*TrickCard{
			{PlayerIdx: 0, Card: NewCard(1, 6, false)},
			{PlayerIdx: 1, Card: NewCard(3, 3, false)},
			{PlayerIdx: 2, Card: NewCard(2, 7, false)},
		}
		g.currentPlayerIdx = 3

		pick := g.cpuSelectPlayCard(3, []int{0, 1})
		assert.Equal(t, 4, p.GetCard(pick).GetValue(),
			"difficulty %d: 味方が勝っているのに強い札を重ねている", diff)
	}
}

// TestAluette_TrickWinnerSurvivesNilEntries はトリックの先頭が nil でも落ちないことを見る。
func TestAluette_TrickWinnerSurvivesNilEntries(t *testing.T) {
	assert.Equal(t, 0, aluetteTrickWinnerOf(nil))
	assert.Equal(t, 0, aluetteTrickWinnerOf([]*TrickCard{nil, nil}))
	assert.Equal(t, 2, aluetteTrickWinnerOf([]*TrickCard{nil, {PlayerIdx: 2, Card: NewCard(1, 4, false)}}))
}
