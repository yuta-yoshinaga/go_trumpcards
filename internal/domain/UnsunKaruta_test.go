//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnsunForTest(t *testing.T) *UnsunKaruta {
	t.Helper()
	g := NewDefaultUnsunKaruta()
	g.Reset()
	return g
}

// unsunPlayDeal は 1 ディールを最後まで打つ。合法手の先頭を出し続ける。
func unsunPlayDeal(t *testing.T, g *UnsunKaruta) {
	t.Helper()
	for step := 0; step < 2000; step++ {
		switch g.GetPhase() {
		case UnsunKarutaPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
				require.NoError(t, g.PlayerPlay(valid[0], false))
				continue
			}
			g.CpuPlay()
		case UnsunKarutaPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == UnsunKarutaPhaseTrickEnd {
				g.NextTrick()
			}
		default:
			return
		}
	}
	require.FailNow(t, "ディールが終わらない")
}

// **デッキは 5 スート × 15 枚。** 4 スートのままだと、そもそも別のゲームになる。
func TestUnsunKaruta_DeckIsSeventyFiveCards(t *testing.T) {
	t.Parallel()
	deck := buildUnsunKarutaDeck()
	assert.Len(t, deck, UnsunKarutaDeckSize)

	perSuit := map[int]map[int]bool{}
	for _, c := range deck {
		if perSuit[c.GetDesign()] == nil {
			perSuit[c.GetDesign()] = map[int]bool{}
		}
		assert.False(t, perSuit[c.GetDesign()][c.GetValue()], "同じ札が 2 枚ある")
		perSuit[c.GetDesign()][c.GetValue()] = true
	}
	assert.Len(t, perSuit, UnsunKarutaSuitCnt)
	for suit, ranks := range perSuit {
		assert.Len(t, ranks, UnsunKarutaRankCnt, "スート %d の枚数", suit)
	}
	// 第 5 スート「クル」が居ること。
	assert.NotNil(t, perSuit[UnsunKarutaSuitKuru])
}

// **丸物の数札は逆順。** うんすんカルタが残した最古の特徴で、取り違えると
// 丸物のトリックが全部ひっくり返る。
func TestUnsunKaruta_RoundSuitPipsAreInverted(t *testing.T) {
	t.Parallel()
	for _, suit := range []int{UnsunKarutaSuitPao, UnsunKarutaSuitIsu} {
		assert.False(t, UnsunKarutaIsRoundSuit(suit), "%d は長物", suit)
		assert.Greater(t, unsunKarutaStrength(NewCard(suit, 9, false)),
			unsunKarutaStrength(NewCard(suit, 1, false)), "長物は 9 が強い")
	}
	for _, suit := range []int{UnsunKarutaSuitKotsu, UnsunKarutaSuitOru, UnsunKarutaSuitKuru} {
		assert.True(t, UnsunKarutaIsRoundSuit(suit), "%d は丸物", suit)
		assert.Greater(t, unsunKarutaStrength(NewCard(suit, 1, false)),
			unsunKarutaStrength(NewCard(suit, 9, false)), "丸物は 1 が強い")
	}
}

// **絵札の並びは ウン > スン > ソウタ > ロバイ > キリ > ウマ。** どの数札より
// 強い。
func TestUnsunKaruta_CourtOrder(t *testing.T) {
	t.Parallel()
	order := []int{UnsunKarutaUn, UnsunKarutaSun, UnsunKarutaSota,
		UnsunKarutaRobai, UnsunKarutaKiri, UnsunKarutaUma}
	for i := 1; i < len(order); i++ {
		prev := unsunKarutaStrength(NewCard(UnsunKarutaSuitPao, order[i-1], false))
		cur := unsunKarutaStrength(NewCard(UnsunKarutaSuitPao, order[i], false))
		assert.Greater(t, prev, cur, "%s は %s より強い",
			UnsunKarutaRankName(order[i-1]), UnsunKarutaRankName(order[i]))
	}
	weakestCourt := unsunKarutaStrength(NewCard(UnsunKarutaSuitPao, UnsunKarutaUma, false))
	for pip := 1; pip <= 9; pip++ {
		for _, suit := range []int{UnsunKarutaSuitPao, UnsunKarutaSuitKotsu} {
			assert.Greater(t, weakestCourt, unsunKarutaStrength(NewCard(suit, pip, false)),
				"ウマは数札より強い (%d の %d)", suit, pip)
		}
	}
}

// **配りは 9 枚 + 中央 3 枚で、返した 1 枚が切り札。**
func TestUnsunKaruta_DealsNineEachAndTurnsTheTrump(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	for i, p := range g.GetPlayers() {
		assert.Equal(t, UnsunKarutaHandSize, p.GetCardsSize(), "席 %d の手札", i)
	}
	assert.Len(t, g.talon, UnsunKarutaTalonSize)
	require.NotNil(t, g.TrumpCard())
	assert.Equal(t, g.TrumpCard().GetDesign(), g.GetTrumpSuit(), "返した札のスートが切り札")
	assert.Equal(t, UnsunKarutaHandSize*UnsunKarutaPlayerCnt+UnsunKarutaTalonSize, UnsunKarutaDeckSize)
}

// **敵味方が交互に座る。** 隣り合う席が同じチームだと、4 対 4 の駆け引きが
// 成り立たない。
func TestUnsunKaruta_SeatsAlternateBetweenTeams(t *testing.T) {
	t.Parallel()
	for seat := 0; seat < UnsunKarutaPlayerCnt; seat++ {
		next := (seat + 1) % UnsunKarutaPlayerCnt
		assert.NotEqual(t, UnsunKarutaTeamOf(seat), UnsunKarutaTeamOf(next), "席 %d と %d", seat, next)
	}
	assert.Equal(t, 0, UnsunKarutaTeamOf(0), "人間は席 0 でチーム 0")
}

// **切り札は台札に勝つ。** 台札の最強札より、どんな切り札でも強い。
func TestUnsunKaruta_TrumpBeatsTheLedSuit(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	g.trumpSuit = UnsunKarutaSuitKuru
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(UnsunKarutaSuitPao, UnsunKarutaUn, false)}, // 台札の最強
		{PlayerIdx: 1, Card: NewCard(UnsunKarutaSuitKuru, 9, false)},            // 最弱の切り札
		{PlayerIdx: 2, Card: NewCard(UnsunKarutaSuitIsu, UnsunKarutaUn, false)},
		{PlayerIdx: 3, Card: NewCard(UnsunKarutaSuitPao, 9, false)},
	}
	assert.Equal(t, 1, g.trickWinner(), "切り札が取る")

	// 切り札が 2 枚出れば強いほうが取る。
	g.currentTrick = append(g.currentTrick,
		&TrickCard{PlayerIdx: 4, Card: NewCard(UnsunKarutaSuitKuru, 1, false)}) // 丸物の 1 = 最強の数札
	assert.Equal(t, 4, g.trickWinner(), "丸物の 1 が 9 に勝つ")
}

// **台札に付いていない札は勝てない。** 切り札でなければ、どんなに強くても
// トリックには絡まない。
func TestUnsunKaruta_OffSuitCardsCannotWin(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	g.trumpSuit = UnsunKarutaSuitKuru
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(UnsunKarutaSuitPao, 3, false)},
		{PlayerIdx: 1, Card: NewCard(UnsunKarutaSuitIsu, UnsunKarutaUn, false)},
	}
	assert.Equal(t, 0, g.trickWinner(), "台札に付いていない ウン は勝てない")
}

// **フォロー義務は宣言で生まれる。** 宣言が無いトリックでは何を出してもよい。
func TestUnsunKaruta_FollowIsOnlyRequiredAfterADeclaration(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	human := findHumanIdx(g.GetPlayers())
	g.currentPlayerIdx = human
	g.currentTrick = []*TrickCard{
		{PlayerIdx: (human + 7) % UnsunKarutaPlayerCnt, Card: NewCard(UnsunKarutaSuitPao, 5, false)},
	}

	g.mustFollow = false
	assert.Len(t, g.GetPlayableIndices(human), g.GetPlayer(human).GetCardsSize(),
		"宣言が無ければ全部出せる")

	g.mustFollow = true
	valid := g.GetPlayableIndices(human)
	hand := g.GetPlayer(human)
	held := 0
	for i := 0; i < hand.GetCardsSize(); i++ {
		if hand.GetCard(i).GetDesign() == UnsunKarutaSuitPao {
			held++
		}
	}
	if held > 0 {
		assert.Len(t, valid, held, "宣言されたら台札にだけ従う")
		for _, idx := range valid {
			assert.Equal(t, UnsunKarutaSuitPao, hand.GetCard(idx).GetDesign())
		}
		return
	}
	assert.Len(t, valid, hand.GetCardsSize(), "台札を持っていなければ何を出してもよい")
}

// 宣言できるのはリードの席だけ。
func TestUnsunKaruta_OnlyTheLeaderMayDeclare(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	human := findHumanIdx(g.GetPlayers())
	g.currentPlayerIdx = human
	g.currentTrick = []*TrickCard{
		{PlayerIdx: (human + 7) % UnsunKarutaPlayerCnt, Card: NewCard(UnsunKarutaSuitPao, 5, false)},
	}
	assert.False(t, g.CanDeclare())
	assert.ErrorIs(t, g.PlayerPlay(0, true), errUnsunKarutaNotLeader)

	g.currentTrick = nil
	assert.True(t, g.CanDeclare())
	require.NoError(t, g.PlayerPlay(0, true))
	assert.True(t, g.IsMustFollow(), "宣言がフォロー義務を立てていない")
	assert.True(t, g.IsDeclaredThisTrick())
}

// **宣言はトリックごと。** 前のトリックの義務を持ち越すと、宣言していない
// リードでも縛られる。
func TestUnsunKaruta_TheDeclarationDoesNotCarryOver(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	g.mustFollow = true
	g.declaredThisTrick = true
	g.phase = UnsunKarutaPhaseTrickEnd
	g.NextTrick()
	assert.False(t, g.IsMustFollow())
	assert.False(t, g.IsDeclaredThisTrick())
}

// **1 ディールで 9 トリック、手札は使い切る。**
func TestUnsunKaruta_ADealPlaysOutNineTricks(t *testing.T) {
	t.Parallel()
	for range 10 {
		g := newUnsunForTest(t)
		unsunPlayDeal(t, g)
		require.Contains(t, []UnsunKarutaPhase{
			UnsunKarutaPhaseRoundEnd, UnsunKarutaPhaseGameEnd,
		}, g.GetPhase())

		tricks := g.GetTeamTricks()
		assert.Equal(t, UnsunKarutaTrickCount, tricks[0]+tricks[1], "コの合計がトリック数と合わない")
		for i, p := range g.GetPlayers() {
			assert.Zero(t, p.GetCardsSize(), "席 %d に札が残っている", i)
		}
	}
}

// **累計はディールの積み上げ。** マッチは規定ディールで終わる。
func TestUnsunKaruta_MatchEndsAfterTheTargetDeals(t *testing.T) {
	t.Parallel()
	cfg := DefaultUnsunKarutaConfig()
	cfg.TargetDeals = 2
	g := NewUnsunKaruta(newUnsunKarutaPlayers(), cfg)
	g.Reset()

	total := 0
	for deal := 0; deal < cfg.TargetDeals; deal++ {
		unsunPlayDeal(t, g)
		total += UnsunKarutaTrickCount
		if g.GetGameEndFlag() {
			break
		}
		require.Equal(t, UnsunKarutaPhaseRoundEnd, g.GetPhase())
		g.NextRound()
	}
	assert.True(t, g.GetGameEndFlag())
	scores := g.GetTeamScores()
	assert.Equal(t, total, scores[0]+scores[1], "累計が配ったトリック数と合わない")
	assert.Contains(t, []UnsunKarutaResult{
		UnsunKarutaResultWin, UnsunKarutaResultLose, UnsunKarutaResultNone,
	}, g.GetResult())
}

// **親は 1 席ずつ回る。** 動かないと同じ席がずっとリードの隣になる。
func TestUnsunKaruta_TheDealerRotates(t *testing.T) {
	t.Parallel()
	cfg := DefaultUnsunKarutaConfig()
	cfg.TargetDeals = 4
	g := NewUnsunKaruta(newUnsunKarutaPlayers(), cfg)
	g.Reset()
	first := g.GetDealerIdx()
	// 親の左隣がリードする。
	assert.Equal(t, (first+1)%UnsunKarutaPlayerCnt, g.GetLeadPlayerIdx())

	unsunPlayDeal(t, g)
	require.False(t, g.GetGameEndFlag())
	g.NextRound()
	assert.Equal(t, (first+1)%UnsunKarutaPlayerCnt, g.GetDealerIdx())
}

// 保存の往復で盤面が保たれる。
func TestUnsunKaruta_SurvivesASaveAndRestore(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	human := findHumanIdx(g.GetPlayers())
	for !g.IsHumanTurn() && g.GetPhase() == UnsunKarutaPhasePlay {
		g.CpuPlay()
	}
	if g.IsHumanTurn() {
		// **宣言できるのはリードのときだけ。** 人間が 2 番手以降で回ってくる
		// 配りもあるので、宣言はリードのときに限って渡す。
		require.NoError(t, g.PlayerPlay(g.GetPlayableIndices(human)[0], g.CanDeclare()))
	}

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(UnsunKaruta)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, g.IsMustFollow(), restored.IsMustFollow(), "宣言が復元されていない")
	assert.Equal(t, g.GetTeamScores(), restored.GetTeamScores())
	assert.Equal(t, g.GetPlayer(human).GetCardsSize(), restored.GetPlayer(human).GetCardsSize())
}

// **壊れた保存は受け取らない。**
func TestUnsunKaruta_RestoreRejectsATamperedBoard(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	for name, mutate := range map[string]func(map[string]json.RawMessage){
		"切り札が存在しないスート": func(m map[string]json.RawMessage) { m["ts"] = json.RawMessage(`9`) },
		"フェーズが範囲外":     func(m map[string]json.RawMessage) { m["ph"] = json.RawMessage(`9`) },
		"親が範囲外":        func(m map[string]json.RawMessage) { m["di"] = json.RawMessage(`8`) },
		"手番が範囲外":       func(m map[string]json.RawMessage) { m["cp"] = json.RawMessage(`-1`) },
		"ディール番号が 0":    func(m map[string]json.RawMessage) { m["rn"] = json.RawMessage(`0`) },
		"席が足りない":       func(m map[string]json.RawMessage) { m["pl"] = json.RawMessage(`[]`) },
	} {
		t.Run(name, func(t *testing.T) {
			clone := make(map[string]json.RawMessage, len(raw))
			for k, v := range raw {
				clone[k] = v
			}
			mutate(clone)
			body, err := json.Marshal(clone)
			require.NoError(t, err)
			assert.Error(t, json.Unmarshal(body, new(UnsunKaruta)))
		})
	}
}

func TestUnsunKarutaConfig_Validate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultUnsunKarutaConfig().Validate())
	for _, deals := range []int{0, -1, 9} {
		cfg := DefaultUnsunKarutaConfig()
		cfg.TargetDeals = deals
		assert.Error(t, cfg.Validate(), "%d ディールを通してしまう", deals)
	}
	cfg := DefaultUnsunKarutaConfig()
	cfg.CpuDifficulty = UnsunKarutaCpuDifficulty(9)
	assert.Error(t, cfg.Validate())
}

// ヒントはフェーズごとに違う手を勧める。
func TestUnsunKaruta_HintFollowsThePhase(t *testing.T) {
	t.Parallel()
	g := newUnsunForTest(t)
	human := findHumanIdx(g.GetPlayers())
	g.currentPlayerIdx = human
	g.currentTrick = nil

	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "lead_strong", hint.Reason)
	require.Len(t, hint.CardIndices, 1)
	assert.Contains(t, g.GetPlayableIndices(human), hint.CardIndices[0], "勧める札が出せない")

	g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(UnsunKarutaSuitPao, 5, false)}}
	assert.Equal(t, "follow_play", g.GetHint().Reason)

	g.phase = UnsunKarutaPhaseTrickEnd
	assert.Equal(t, "next_trick", g.GetHint().Reason)
	g.phase = UnsunKarutaPhaseRoundEnd
	assert.Equal(t, "next_round", g.GetHint().Reason)
}

// **手札は強さ順に並ぶ。** 値の順に並べると、丸物だけ弱い札が先頭に来る。
func TestUnsunKaruta_HandIsSortedByStrength(t *testing.T) {
	t.Parallel()
	p := NewUnsunKarutaPlayer(true)
	for _, c := range []*Card{
		NewCard(UnsunKarutaSuitKotsu, 1, false),
		NewCard(UnsunKarutaSuitKotsu, 9, false),
		NewCard(UnsunKarutaSuitKotsu, UnsunKarutaUn, false),
		NewCard(UnsunKarutaSuitPao, 1, false),
		NewCard(UnsunKarutaSuitPao, 9, false),
	} {
		p.AddCard(c)
	}
	unsunKarutaSortHand(p)

	var suits, values []int
	for i := 0; i < p.GetCardsSize(); i++ {
		suits = append(suits, p.GetCard(i).GetDesign())
		values = append(values, p.GetCard(i).GetValue())
	}
	assert.Equal(t, []int{UnsunKarutaSuitPao, UnsunKarutaSuitPao,
		UnsunKarutaSuitKotsu, UnsunKarutaSuitKotsu, UnsunKarutaSuitKotsu}, suits)
	// 長物は 9 が先、丸物は 1 が先 (ウンが最初)。
	assert.Equal(t, []int{9, 1, UnsunKarutaUn, 1, 9}, values)
}

// **助言は CPU の難易度に引きずられない。** cpuSelectPlayCard は Easy だと
// 合法手からランダムに選ぶので、GetHint がそれを流用していると「Easy を
// 選んだ人にだけでたらめな札を勧める」ことになる。
func TestUnsunKarutaGetHintIgnoresCpuDifficulty(t *testing.T) {
	easy := NewUnsunKaruta(newUnsunKarutaPlayers(), UnsunKarutaConfig{
		CpuDifficulty: UnsunKarutaCpuDifficultyEasy,
		TargetDeals:   UnsunKarutaDefaultDeals,
	})
	easy.Reset()
	for easy.GetPhase() == UnsunKarutaPhasePlay && !easy.IsHumanTurn() {
		easy.CpuPlay()
	}
	if easy.GetPhase() != UnsunKarutaPhasePlay {
		t.Skip("配りによっては人間の手番の前にトリックが揃う")
	}
	human := easy.GetCurrentPlayerIdx()
	valid := easy.GetPlayableIndices(human)
	// 合法手が 1 枚しか無ければランダムでも同じ札が返る = 何も試していない。
	require.Greater(t, len(valid), 1, "合法手が 1 枚しかない")
	want := easy.cpuPlaySmart(human, valid)

	// 同じ盤面で 20 回訊いても、返るのは毎回その 1 枚。ランダム選択を通して
	// いれば、合法手が複数あるかぎりどこかで別の札が返る。
	for i := 0; i < 20; i++ {
		hint := easy.GetHint()
		require.NotNil(t, hint)
		require.Len(t, hint.CardIndices, 1)
		assert.Equal(t, want, hint.CardIndices[0], "%d 回目でぶれた", i+1)
	}
}
