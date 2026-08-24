//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eightGameTotalChips は正本の総量を返す。
func eightGameTotalChips(g *Horse) int {
	total := 0
	for i := range g.GetSeatCount() {
		total += g.GetSeatChips(i)
	}
	return total
}

// eightGamePlayHand は 1 ハンドを最後まで打つ。**ベットは全部コール、
// 引き直しは 3 枚**という乱暴な打ち方だが、種目ごとの規則を書き直さずに
// 8 種目すべてを通せる唯一の打ち方でもある。
//
// 打った手数を返す。0 を返したなら「ハンドを打った」ことにはなっていない。
func eightGamePlayHand(t *testing.T, g *Horse) int {
	t.Helper()
	steps := 0
	for ; g.GetPhase() == HorsePhaseHand; steps++ {
		require.Less(t, steps, 200, "ハンドが終わらない (%s)", HorseDisciplineName(g.GetDiscipline()))
		switch {
		case g.IsDrawPhase():
			require.NoError(t, g.PlayerExchange([]int{0, 1, 2}))
		case g.IsHumanTurn():
			if err := g.PlayerAction(HoldemActionCall, 0, 0); err != nil {
				require.NoError(t, g.PlayerAction(HoldemActionCheck, 0, 0))
			}
		default:
			require.FailNow(t, "打てる手が 1 つも無いまま盤面が止まっている",
				"種目=%s 卓のフェーズ=%d 手番=%d",
				HorseDisciplineName(g.GetDiscipline()), g.GetTablePhase(), g.GetCurrentTurn())
		}
	}
	return steps
}

// **8 種目すべてが実際に打てる。** 並びに載っているだけで卓が作れない種目が
// あると、そのハンドでマッチが理由も出さずに終わる。
//
// **自然なローテーションを 8 ハンド待たない。** 席が飛べば `NextHand` が
// そこでマッチを畳むので、「8 連続で誰も飛ばない配り」を待つ試験は配りに
// 依存して落ちる (実測: 同じパッケージにテストを足しただけで発現した)。
// 種目は直接指定して 1 ハンドずつ打ち切る。
func TestEightGame_EveryDisciplineCanBePlayed(t *testing.T) {
	t.Parallel()
	for _, d := range HorseRotation(HorseVariantEightGame) {
		g := NewEightGame(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
		g.Reset()
		g.discipline = d
		g.startHand()
		require.Equal(t, HorsePhaseHand, g.GetPhase(), "%s の卓が作れていない", HorseDisciplineName(d))
		require.Positive(t, eightGamePlayHand(t, g), "%s のハンドが打てない", HorseDisciplineName(d))
		assert.Equal(t, d, g.GetDiscipline(), "ハンドの途中で種目が変わった")
	}
}

// **並びの順に進む。** ローテーションが 1 つでもずれると、8 種目のうちどれかが
// 二度回るか一度も来ない。
func TestEightGame_AdvancesInRotationOrder(t *testing.T) {
	t.Parallel()
	rotation := HorseRotation(HorseVariantEightGame)
	g := NewEightGame(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()
	got := make([]HorseDiscipline, 0, len(rotation)+1)
	for range len(rotation) + 1 {
		got = append(got, g.GetDiscipline())
		g.discipline = g.nextDiscipline()
	}
	assert.Equal(t, append(append([]HorseDiscipline(nil), rotation...), rotation[0]), got,
		"一周して先頭へ戻るまでが並び")
}

// **チップは種目をまたいでも湧かないし消えない。** 追加の 3 種目は卓の作り方が
// 他と違う (リミットの差し替え / ドロー系のプレイヤー生成) ので、正本への
// 配りと回収がそこだけずれていないかを見る。
func TestEightGame_EveryDisciplineConservesChips(t *testing.T) {
	t.Parallel()
	for range 25 {
		g := NewEightGame(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
		g.Reset()
		for range len(HorseRotation(HorseVariantEightGame)) {
			if g.GetGameEndFlag() {
				break
			}
			d := g.GetDiscipline()
			before := eightGameTotalChips(g)
			eightGamePlayHand(t, g)
			assert.Equal(t, before, eightGameTotalChips(g),
				"%s のハンドで総量が変わった", HorseDisciplineName(d))
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
}

// **H.O.R.S.E. は 5 種目のまま。** 種目の値を 8 つに増やしたので、
// `(d+1) % 種目数` のまま進めていると H.O.R.S.E. がノーリミットまで回る。
func TestHorse_RotationStillStopsAtFiveDisciplines(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()
	seen := map[HorseDiscipline]bool{}
	for range 12 {
		seen[g.GetDiscipline()] = true
		g.discipline = g.nextDiscipline()
	}
	assert.Len(t, seen, HorseDisciplineCount)
	for _, d := range []HorseDiscipline{HorseNLHoldem, HorsePLOmaha, HorseTripleDraw} {
		assert.False(t, seen[d], "H.O.R.S.E. が %s を回している", HorseDisciplineName(d))
	}
}

// ローテーションは定義そのもの。並びが変わるとゲームが変わる。
func TestEightGame_RotationOrder(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []HorseDiscipline{
		HorseHoldem, HorseOmahaHiLo, HorseRazz, HorseStud, HorseStudHiLo,
		HorseNLHoldem, HorsePLOmaha, HorseTripleDraw,
	}, HorseRotation(HorseVariantEightGame))
	assert.Equal(t, []HorseDiscipline{
		HorseHoldem, HorseOmahaHiLo, HorseRazz, HorseStud, HorseStudHiLo,
	}, HorseRotation(HorseVariantHorse))

	// **返すのは複製。** 呼び出し側が並び替えても卓は変わらない。
	rot := HorseRotation(HorseVariantEightGame)
	rot[0] = HorseTripleDraw
	assert.Equal(t, HorseHoldem, HorseRotation(HorseVariantEightGame)[0])
}

// **リミットの違いが種目の違い。** 差し替えを忘れると、8 種目のうち 2 つが
// 先に回したリミット種目とまったく同じ卓になる。
func TestEightGame_NoLimitAndPotLimitTablesUseTheirLimits(t *testing.T) {
	t.Parallel()
	g := NewEightGame(DefaultEightGameConfig())
	g.Reset()

	g.discipline = HorseNLHoldem
	g.startHand()
	h, ok := g.table.(*Holdem)
	require.True(t, ok, "ノーリミットはホールデムの卓")
	assert.Equal(t, BettingLimitNoLimit, h.GetConfig().BettingLimit)

	g.discipline = HorsePLOmaha
	g.startHand()
	o, ok := g.table.(*Omaha)
	require.True(t, ok, "ポットリミットはオマハの卓")
	assert.Equal(t, BettingLimitPotLimit, o.GetConfig().BettingLimit)
	assert.False(t, o.GetIsHiLo(), "PLO はハイのみ。Hi-Lo は別の種目として既に回している")

	// リミットホールデムは据え置き。
	g.discipline = HorseHoldem
	g.startHand()
	limit, ok := g.table.(*Holdem)
	require.True(t, ok)
	assert.Equal(t, BettingLimitFixed, limit.GetConfig().BettingLimit)
}

// **引き直しはドロー系の種目でしか受け付けない。** 受け付けてしまうと、
// ベットの最中に手札が入れ替わる。
func TestEightGame_ExchangeOnlyWorksInTheDrawDiscipline(t *testing.T) {
	t.Parallel()
	g := NewEightGame(DefaultEightGameConfig())
	g.Reset()
	require.Equal(t, HorseHoldem, g.GetDiscipline())
	assert.False(t, g.IsDrawPhase(), "ホールデムに引き直しは無い")
	assert.ErrorIs(t, g.PlayerExchange([]int{0}), errHorseNoDraw)
	assert.Zero(t, g.GetDrawIndex())
}

// **引き直しの番は「打てる手がある番」。** ここが false のままだと、
// ドローの盤面でベットの面しか出ず、6 種目目でマッチが止まる。
func TestEightGame_TripleDrawReachesADrawTurn(t *testing.T) {
	t.Parallel()
	drew := false
	for range 20 {
		g := NewEightGame(DefaultEightGameConfig())
		g.Reset()
		g.discipline = HorseTripleDraw
		g.startHand()
		for steps := 0; steps < 200 && g.GetPhase() == HorsePhaseHand; steps++ {
			if g.IsDrawPhase() {
				assert.Contains(t, []int{1, 2, 3}, g.GetDrawIndex(), "引き直しは 3 回まで")
				before := g.GetSeatCards(g.GetHumanSeat())
				require.Len(t, before, DeuceToSevenHandSize)
				require.NoError(t, g.PlayerExchange([]int{0}))
				drew = true
				continue
			}
			if g.IsHumanTurn() {
				if err := g.PlayerAction(HoldemActionCall, 0, 0); err != nil {
					require.NoError(t, g.PlayerAction(HoldemActionCheck, 0, 0))
				}
				continue
			}
			require.FailNow(t, "ドローの卓で打てる手が無くなった")
		}
		if drew {
			break
		}
	}
	assert.True(t, drew, "20 ハンド打っても引き直しの番が来なかった")
}

// **相手の 5 枚は 1 枚も見せない。** ドロー系に表向きの札は無いので、
// 表示に混ぜると相手の手が全部読める。
func TestEightGame_TripleDrawHidesTheOpponentsHand(t *testing.T) {
	t.Parallel()
	g := NewEightGame(DefaultEightGameConfig())
	g.Reset()
	g.discipline = HorseTripleDraw
	g.startHand()

	assert.Len(t, g.GetSeatCards(g.GetHumanSeat()), DeuceToSevenHandSize, "自分の手は全部見える")
	for seat := range g.GetSeatCount() {
		if seat == g.GetHumanSeat() {
			continue
		}
		assert.Empty(t, g.GetSeatCards(seat), "席 %d の札が見えている", seat)
	}
	assert.Empty(t, g.GetCommunityCards(), "ドロー系に共有札は無い")
}

// **Eight-Game Mix は 4 人卓しか作れない。** 2-7 Triple Draw が 4 席までで、
// 6 人や 9 人を選べるようにすると 6 ハンド目で理由も出さずに終わる。
func TestEightGame_OnlyFourSeatsAreAccepted(t *testing.T) {
	t.Parallel()
	for _, seats := range []int{6, 9} {
		cfg := DefaultEightGameConfig()
		cfg.Seats = seats
		assert.ErrorIs(t, cfg.Validate(), errHorseSeatsRange, "%d 人卓を通してしまう", seats)
	}
	assert.NoError(t, DefaultEightGameConfig().Validate())

	// H.O.R.S.E. のほうは今までどおり 4 / 6 / 9。
	for _, seats := range HorseSeatSizes {
		cfg := DefaultHorseConfig()
		cfg.Seats = seats
		assert.NoError(t, cfg.Validate(), "H.O.R.S.E. の %d 人卓が弾かれた", seats)
	}
}

// **保存の往復でローテーションが変わらない。** バリアントが落ちると、
// 復元した卓は 5 種目しか回さなくなる (8 種目のはずが H.O.R.S.E. になる)。
func TestEightGame_SurvivesASaveAndRestore(t *testing.T) {
	t.Parallel()
	g := NewEightGame(DefaultEightGameConfig())
	g.Reset()
	g.discipline = HorseTripleDraw
	g.startHand()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := new(Horse)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, HorseVariantEightGame, restored.GetVariant())
	assert.Equal(t, HorseTripleDraw, restored.GetDiscipline())
	assert.Len(t, restored.GetRotation(), 8)
	assert.Len(t, restored.GetSeatCards(restored.GetHumanSeat()), DeuceToSevenHandSize,
		"打ちかけの手が消えている")
}

// **回さない種目からは復元しない。** 保存を書き換えれば、5 種目の卓を
// 2-7 Triple Draw の途中から復元できてしまう。
func TestHorse_RestoreRejectsADisciplineOutsideTheRotation(t *testing.T) {
	t.Parallel()
	g := NewEightGame(DefaultEightGameConfig())
	g.Reset()
	g.discipline = HorseTripleDraw
	g.startHand()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	// バリアントだけ H.O.R.S.E. に書き換える。種目は 2-7 のまま。
	raw["cf"] = json.RawMessage(`{"s":4,"c":1000,"h":2,"v":0}`)
	tampered, err := json.Marshal(raw)
	require.NoError(t, err)

	assert.ErrorContains(t, json.Unmarshal(tampered, new(Horse)), "invalid discipline")
}

// **ショーダウンで盤面が凍らない。** 人間がコールして負けた手は種目側が
// マック待ちで留まるが、ミックスゲームにはその入力が無い ── 公開して閉じる
// 経路が無いと、打てる手が 1 つも無い盤面のまま次のハンドへも進めない。
func TestHorse_AShowdownLossDoesNotFreezeTheMatch(t *testing.T) {
	t.Parallel()
	froze := 0
	for range 40 {
		g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
		g.Reset()
		for steps := 0; steps < 200 && g.GetPhase() == HorsePhaseHand; steps++ {
			if !g.IsHumanTurn() {
				froze++
				break
			}
			if err := g.PlayerAction(HoldemActionCall, 0, 0); err != nil {
				require.NoError(t, g.PlayerAction(HoldemActionCheck, 0, 0))
			}
		}
	}
	assert.Zero(t, froze, "コールし続けただけで盤面が止まった")
}

// **既定の卓は 8 種目の 4 人卓。** 登録はこの入口から作るので、ここが
// H.O.R.S.E. を返していると、ゲーム一覧に並ぶのは名前だけ違う同じ卓になる。
func TestEightGame_DefaultTableIsAFourHandedEightGame(t *testing.T) {
	t.Parallel()
	g := NewDefaultEightGame()
	assert.Equal(t, HorseVariantEightGame, g.GetVariant())
	assert.Equal(t, HorseEightGameSeatSizes[0], g.GetConfig().Seats)
	assert.Len(t, g.GetRotation(), 8)
	g.Reset()
	assert.Equal(t, HorseHoldem, g.GetDiscipline(), "並びの先頭はリミットホールデム")
}

// **ドロー系の卓でも画面の数字は答えられる。** ここが 0 のままだと、
// 2-7 の番だけコール額も残高も出ない画面になる。
func TestEightGame_DrawTableAnswersTheBoardQuestions(t *testing.T) {
	t.Parallel()
	g := NewDefaultEightGame()
	g.Reset()
	g.discipline = HorseTripleDraw
	g.startHand()

	human := g.GetHumanSeat()
	assert.Empty(t, g.GetCommunityCards(), "ドロー系に共有札は無い")
	assert.GreaterOrEqual(t, g.GetToCall(), 0)
	assert.GreaterOrEqual(t, g.GetMinRaise(), 0)
	assert.Positive(t, g.GetSeatLiveChips(human), "打っている最中の残高が出ない")
	assert.Positive(t, g.GetPot(), "アンティを置いた卓のポットが 0")

	// 卓の外の席は nil。**範囲外を読むと 1 手で落ちる。**
	table, ok := g.table.(*DeuceToSeven)
	require.True(t, ok)
	assert.Nil(t, horseDrawPlayer(table, -1))
	assert.Nil(t, horseDrawPlayer(table, g.GetSeatCount()))
	assert.NotNil(t, horseDrawPlayer(table, 0))
}

// 引き直しも終わったマッチや打てない局面では断る。
func TestEightGame_ExchangeRefusesOutsideAHand(t *testing.T) {
	t.Parallel()
	g := NewDefaultEightGame()
	g.Reset()
	g.discipline = HorseTripleDraw
	g.startHand()

	g.phase = HorsePhaseHandEnd
	assert.ErrorIs(t, g.PlayerExchange(nil), errHorseWrongPhase)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerExchange(nil), errHorseFinished)
	assert.False(t, g.IsDrawPhase(), "終わったマッチが引き直しを名乗っている")
	assert.Zero(t, g.GetDrawIndex())
}

// **知らないバリアントは H.O.R.S.E. として扱う。** ゼロ値の卓を組み立てる経路
// (復元の途中など) でローテーションの参照が落ちると、次の 1 手で index out of
// range になる。
func TestHorse_RotationHelpersRejectAnUnknownVariant(t *testing.T) {
	t.Parallel()
	assert.Nil(t, HorseRotation(HorseVariant(-1)))
	assert.Nil(t, HorseRotation(HorseVariant(99)))
	assert.Equal(t, -1, HorseRotationIndex(HorseVariant(99), HorseHoldem))
	assert.Equal(t, -1, HorseRotationIndex(HorseVariantHorse, HorseTripleDraw),
		"H.O.R.S.E. の並びに 2-7 は無い")

	var zero Horse
	zero.config.Variant = HorseVariant(42)
	assert.Equal(t, HorseVariantHorse, zero.rotationVariant())
	assert.Equal(t, HorseHoldem, zero.rotationAt(0))
	assert.Equal(t, HorseHoldem, zero.rotationAt(HorseDisciplineCount), "一周して先頭へ戻る")
	assert.Equal(t, HorseStudHiLo, zero.rotationAt(-1), "負の位置も並びの中に収める")
}

// **並びの外にいる卓は先頭へ戻す。** 保存の書き換えや将来の種目追加で、
// いまのローテーションに無い種目を指した卓が生まれうる。そこで詰まると
// 「次のハンド」が永遠に同じ種目を配り直す。
func TestHorse_AnOutOfRotationDisciplineFallsBackToTheStart(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()
	g.discipline = HorseTripleDraw // H.O.R.S.E. の並びには無い
	assert.Equal(t, HorseHoldem, g.nextDiscipline())
}

// **本番の経路で 8 種目を回り切る。** 種目を直接指定する試験や
// `nextDiscipline` を直接歩かせる試験は、**`NextHand` から呼ばれていなくても
// 通る** ── レビューで指摘された穴で、実際そこにミューテーションが残っていた。
//
// **席が飛ばないようにしてから回す。** 誰かが飛べば `NextHand` はマッチを畳むので、
// 素直に回すと配り次第で 5 種目目までしか進まない。1 ハンドごとに全員の残高を
// 戻せば、進むのはローテーションだけになる。
func TestEightGame_NextHandWalksTheWholeRotation(t *testing.T) {
	t.Parallel()
	rotation := HorseRotation(HorseVariantEightGame)
	g := NewEightGame(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()

	seen := make([]HorseDiscipline, 0, len(rotation)+1)
	for range len(rotation) + 1 {
		seen = append(seen, g.GetDiscipline())
		require.Positive(t, eightGamePlayHand(t, g), "%s のハンドが打てない", HorseDisciplineName(g.GetDiscipline()))
		for i := range g.GetSeatCount() {
			g.SetSeatChips(i, HorseDefaultChips)
		}
		require.False(t, g.GetGameEndFlag(), "残高を戻したのにマッチが終わった")
		require.NoError(t, g.NextHand())
	}

	assert.Equal(t, append(append([]HorseDiscipline(nil), rotation...), rotation[0]), seen,
		"NextHand が並びの順に進んでいない")
	assert.Contains(t, seen, HorseTripleDraw, "8 種目目に到達していない")
}
