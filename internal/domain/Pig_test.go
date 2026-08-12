//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPigForTest(t *testing.T, n int) *Pig {
	t.Helper()
	cfg := DefaultPigConfig()
	cfg.PlayerCnt = n
	g := NewPig(newPigSeats(n), cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

// **配る枚数は最初から最後まで 4 枚。** 渡すのと受け取るのが同時だからです。
func TestPigResetDealsFourToEverySeat(t *testing.T) {
	for n := PigPlayerCntMin; n <= PigPlayerCntMax; n++ {
		g := newPigForTest(t, n)
		require.Equal(t, PigPhasePass, g.GetPhase())
		total := 0
		for i := range n {
			assert.Equal(t, PigHandSize, g.GetPlayer(i).GetCardsSize(), "%d 人: 席 %d", n, i)
			total += g.GetPlayer(i).GetCardsSize()
		}
		// **デッキは人数 × 4 枚ちょうど。** 余りも不足も出ません。
		assert.Equal(t, PigDeckSize(n), total, "%d 人", n)
	}
}

// **使うのは人数と同じ種類のランクを 4 スート揃えたもの。**
//
// 揃えられなければゲームが成立しないので、**各ランクがちょうど 4 枚**あることが
// 配りの条件そのものです。
func TestPigDealUsesCompleteRankSets(t *testing.T) {
	for n := PigPlayerCntMin; n <= PigPlayerCntMax; n++ {
		g := newPigForTest(t, n)
		counts := map[int]int{}
		designs := map[int]map[int]bool{}
		for i := range n {
			p := g.GetPlayer(i)
			for k := range p.GetCardsSize() {
				c := p.GetCard(k)
				counts[c.GetValue()]++
				if designs[c.GetValue()] == nil {
					designs[c.GetValue()] = map[int]bool{}
				}
				assert.False(t, designs[c.GetValue()][c.GetDesign()],
					"%d 人: ランク %d のスート %d が重複した", n, c.GetValue(), c.GetDesign())
				designs[c.GetValue()][c.GetDesign()] = true
			}
		}
		assert.Len(t, counts, n, "%d 人: ランクの種類", n)
		for rank, cnt := range counts {
			assert.Equal(t, 4, cnt, "%d 人: ランク %d が 4 枚ない", n, rank)
		}
	}
}

// **同時に渡す。** 全員が選び終わるまで札は動きません。
func TestPigPassIsSimultaneous(t *testing.T) {
	g := newPigForTest(t, 4)
	before := make([]int, 4)
	for i := range 4 {
		before[i] = g.GetPlayer(i).GetCardsSize()
	}

	require.NoError(t, g.ChoosePassForTest(0, 0))
	// 選んだ席だけが 1 枚減り、**他の席はまだ受け取っていない**。
	assert.Equal(t, before[0]-1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, before[1], g.GetPlayer(1).GetCardsSize(), "まだ回っていない")
	assert.True(t, g.HasChosenPass(0))

	for i := 1; i < 4; i++ {
		require.NoError(t, g.ChoosePassForTest(i, 0))
	}
	// 全員が選び終わると一斉に回り、**枚数は元に戻る**。
	for i := range 4 {
		assert.Equal(t, PigHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
		assert.False(t, g.HasChosenPass(i))
	}
	assert.Equal(t, 1, g.GetPassCount())
}

// **渡す札は左隣へ行く。**
func TestPigPassGoesToTheLeftNeighbour(t *testing.T) {
	g := newPigForTest(t, 4)
	marker := NewCard(CardDesignSpade, 7, false)
	g.GiveHandForTest(0, marker,
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 4, false))

	require.NoError(t, g.ChoosePassForTest(0, 0))
	for i := 1; i < 4; i++ {
		require.NoError(t, g.ChoosePassForTest(i, 0))
	}

	found := false
	for k := range g.GetPlayer(1).GetCardsSize() {
		c := g.GetPlayer(1).GetCard(k)
		if c.GetDesign() == CardDesignSpade && c.GetValue() == 7 {
			found = true
		}
	}
	assert.True(t, found, "席 0 が渡した札は席 1 に届く")
}

// **4 枚揃うと合図フェーズに移る。** 揃えた本人が最初に気づいた扱い。
func TestPigFourOfAKindOpensTheSignal(t *testing.T) {
	g := newPigForTest(t, 4)
	// 席 1 が受け取った時点で 4 枚揃うように仕込む。
	// **16 枚をちょうど使い切る配り。** 1 枚でも重複すると別の席も揃ってしまい、
	// 「最初に揃えた席」が入れ替わって、この検査は静かに別のことを見ます。
	g.GiveHandForTest(0, NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 13, false))
	g.GiveHandForTest(1, NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignSpade, 12, false))
	g.GiveHandForTest(2, NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignHeart, 12, false))
	g.GiveHandForTest(3, NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignDiamond, 11, false))

	require.NoError(t, g.ChoosePassForTest(0, 0)) // ♠A を席 1 へ
	require.NoError(t, g.ChoosePassForTest(1, 3)) // ♠Q を席 2 へ
	require.NoError(t, g.ChoosePassForTest(2, 3)) // ♥Q を席 3 へ
	require.NoError(t, g.ChoosePassForTest(3, 3)) // ♦J を席 0 へ

	assert.Equal(t, PigPhaseSignal, g.GetPhase())
	assert.Equal(t, 1, g.GetSignallerIdx())
	assert.True(t, g.GetPlayer(1).GetHasSignalled(), "揃えた本人は気づいている")
	assert.Equal(t, 1, g.GetPlayer(1).GetNoticedOrder())
	assert.Equal(t, 1, g.GetNoticedCnt())
}

// **最後の 1 人が残った時点で決まり。** 全員が気づくのを待ちません。
func TestPigLastToNoticeTakesTheLetter(t *testing.T) {
	g := newPigForTest(t, 4)
	g.OpenSignalForTest(2)
	require.Equal(t, PigPhaseSignal, g.GetPhase())

	// 合図した本人がすでに 1 人目。あと 3 人のうち 2 人が気づいた時点で決まる。
	g.NoticeForTest(3)
	assert.Equal(t, PigPhaseSignal, g.GetPhase(), "まだ 2 人残っている")

	g.NoticeForTest(0)
	// 残り 1 人になった時点でラウンドが終わる——全員を待ちません。
	assert.Equal(t, PigPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundLoserIdx())
}

// **気づかなかった 1 人に文字が付く。**
func TestPigLoserGetsALetter(t *testing.T) {
	g := newPigForTest(t, 4)
	g.OpenSignalForTest(0)
	g.NoticeForTest(1)
	g.NoticeForTest(2)
	// 残りは席 3 だけ。この時点でラウンドが終わる。
	assert.Equal(t, 3, g.GetRoundLoserIdx())
	assert.Equal(t, 1, g.GetPlayer(3).GetLetters())
	assert.Equal(t, "P", g.GetPlayer(3).GetLetterWord())

	// **すぐには配り直しません。** 罰は盤面に痕跡が残らないので、読む時間を作る。
	assert.Equal(t, PigPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())

	require.NoError(t, g.NextRound())
	assert.Equal(t, PigPhasePass, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// **誰が文字をもらったかは次の罰まで残す。** 配り直しで消しません。
	assert.Equal(t, 3, g.GetRoundLoserIdx())
}

// **3 文字で脱落し、最後の 1 人が勝つ。**
func TestPigThreeLettersEliminates(t *testing.T) {
	g := newPigForTest(t, 3)
	g.GetPlayer(2).SetLetters(2)
	g.OpenSignalForTest(0)
	g.NoticeForTest(1)

	assert.Equal(t, PigMaxLetters, g.GetPlayer(2).GetLetters())
	assert.True(t, g.GetPlayer(2).GetEliminated())
	assert.Equal(t, "PIG", g.GetPlayer(2).GetLetterWord())
	// **脱落した席は合図の記録も落とす。** codec が受け付けない状態を作らない。
	assert.False(t, g.GetPlayer(2).GetHasSignalled())
	assert.Zero(t, g.GetPlayer(2).GetNoticedOrder())

	// 3 人卓なので、1 人脱落すると残り 2 人。まだ続く。
	assert.False(t, g.GetGameEndFlag())
	require.NoError(t, g.NextRound())

	// もう 1 人落とすと決着。
	g.GetPlayer(1).SetLetters(2)
	g.OpenSignalForTest(0)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

// **脱落した席は配りにも合図にも加わりません。**
func TestPigEliminatedSeatsAreDealtOut(t *testing.T) {
	g := newPigForTest(t, 4)
	g.GetPlayer(2).SetEliminated(true)
	g.SetPhaseForTest(PigPhaseRoundEnd)
	g.SetRoundLoserIdxForTest(2)
	require.NoError(t, g.NextRound())

	assert.Zero(t, g.GetPlayer(2).GetCardsSize(), "脱落した席には配らない")
	// **デッキも 1 人分縮む。** 3 人なら 12 枚。
	assert.Equal(t, PigDeckSize(3), g.GetDeckSize())
	total := 0
	for i := range 4 {
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, PigDeckSize(3), total)
}

// **合図が出ていないのに鼻を触るのは反則。**
func TestPigSignalIsRejectedBeforeAnybodySignals(t *testing.T) {
	g := newPigForTest(t, 4)
	require.Equal(t, PigPhasePass, g.GetPhase())
	assert.Error(t, g.PlayerSignal())

	g.OpenSignalForTest(1)
	assert.NoError(t, g.PlayerSignal())
	assert.Error(t, g.PlayerSignal(), "二度は名乗れない")
}

func TestPigPlayerPassRejectsBadInput(t *testing.T) {
	g := newPigForTest(t, 4)
	assert.Error(t, g.PlayerPass(-1))
	assert.Error(t, g.PlayerPass(PigHandSize))
	require.NoError(t, g.PlayerPass(0))
	assert.Error(t, g.PlayerPass(0), "同じラウンドで二度は渡せない")

	g.OpenSignalForTest(1)
	assert.Error(t, g.PlayerPass(0), "合図の場面では渡せない")
}

func TestPigGiveUp(t *testing.T) {
	g := newPigForTest(t, 4)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, PigPhaseGameEnd, g.GetPhase())
	assert.NotEqual(t, 0, g.GetWinnerIdx(), "投了した人は勝たない")

	g.GiveUp() // 二度目は無視される
	assert.True(t, g.GetGameEndFlag())
}

func TestPigHint(t *testing.T) {
	g := newPigForTest(t, 4)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.NotNil(t, h.CardIndex, "渡す場面では札を指す")

	g.OpenSignalForTest(1)
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex, "合図の場面では札を指さない")
	assert.Equal(t, "pigSignal", h.Reason)

	require.NoError(t, g.PlayerSignal())
	assert.Nil(t, g.GetHint(), "名乗ったあとは助言しない")

	g.GiveUp()
	assert.Nil(t, g.GetHint(), "終局後は助言しない")
}

// **CPU は少ないランクを手放します。**
func TestPigCpuKeepsThePairs(t *testing.T) {
	g := newPigForTest(t, 4)
	g.GiveHandForTest(1,
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 1, false))
	assert.Equal(t, 3, g.CpuChoiceForTest(1), "3 枚揃っているランクは残す")
}

// **全人数を終局まで指して、1 手ごとに保存して読み直す。**
//
// #5316 のレビュー指摘（codec のガードは正しく、破っていたのは書き込み側）から
// 標準化したテスト。範囲検査では出ない「フィールド間の食い違い」を、書き込み側が
// 実際に作れるかどうかで確かめます。
func TestPig_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := PigPlayerCntMin; n <= PigPlayerCntMax; n++ {
		for seed := range 5 {
			cfg := DefaultPigConfig()
			cfg.PlayerCnt = n
			g := NewPig(newPigSeats(n), cfg)
			g.SetRand(rand.New(rand.NewSource(int64(seed) + 1)))
			g.Reset()

			for turns := 0; ; turns++ {
				require.Less(t, turns, 100000, "%d 人 seed %d: 終わらない", n, seed)

				data, err := json.Marshal(g)
				require.NoError(t, err)
				var back Pig
				require.NoError(t, json.Unmarshal(data, &back),
					"%d 人 seed %d %d 手目: 書き込み側が codec の不変条件を破った", n, seed, turns)

				if g.GetGameEndFlag() {
					break
				}
				switch g.GetPhase() {
				case PigPhasePass:
					idx := g.GetCurrentPlayerIdx()
					require.NoError(t, g.ChoosePassForTest(idx, g.CpuChoiceForTest(idx)))
				case PigPhaseSignal:
					// CPU が順に気づく。人間は **本番の入口** から名乗る——
					// notice() を直に叩くと、脱落済みの席にも合図させられてしまい、
					// production では届かない状態を作ってしまいます。
					g.CpuPlay()
					_ = g.PlayerSignal()
				case PigPhaseRoundEnd:
					require.NoError(t, g.NextRound())
				default:
					t.Fatalf("%d 人: 進めないフェーズ %d", n, g.GetPhase())
				}
			}
		}
	}
}
