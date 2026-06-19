//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spCard は domain.NewCard の薄いラッパ (テストヘルパ)。
func spCard(d, v int) *Card { return NewCard(d, v, true) }

// spSetHand はプレイヤーの手札を明示的に設定する (Reset 後の確定状態を作る)。
func spSetHand(p *SpoonsPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newTestSpoons は決定的乱数源を持つ標準セットアップの Spoons を返す。
func newTestSpoons() *Spoons {
	g := NewDefaultSpoons()
	g.SetRand(rand.New(rand.NewSource(1)))
	return g
}

// 以下は同一パッケージ (package domain) のテストから内部フィールドを直接
// 操作するためのヘルパ。Reset() のシャッフル結果に依存しない確定状態を
// 作るために用いる。

func spSetPhase(g *Spoons, ph SpoonsPhase)  { g.phase = ph }
func spSetCurrentPlayer(g *Spoons, idx int) { g.currentPlayerIdx = idx }
func spSetPassedCard(g *Spoons, c *Card)    { g.passedCard = c }
func spSetGrabWindow(g *Spoons, open bool)  { g.grabWindowOpen = open }
func spSetSpoonsRemaining(g *Spoons, n int) { g.spoonsRemaining = n }
func spSetGameEnd(g *Spoons)                { g.gameEndFlag = true; g.phase = SpoonsPhaseGameEnd }

func TestSpoons_DefaultConstruction(t *testing.T) {
	g := NewDefaultSpoons()
	assert.Equal(t, SpoonsPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < SpoonsPlayerCnt; i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman())
	}
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(SpoonsPlayerCnt))
}

func TestSpoons_ResetDealsHandsAndSpoons(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	assert.Equal(t, SpoonsPhasePass, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, SpoonsPlayerCnt-1, g.GetSpoonsRemaining())
	for i := 0; i < SpoonsPlayerCnt; i++ {
		assert.Equal(t, SpoonsHandSize, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetLetters())
		assert.False(t, g.GetPlayer(i).GetEliminated())
	}
	// 配り手と最初の手番は最初の生存者 (0)。
	assert.Equal(t, 0, g.GetFeederIdx())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
}

func TestSpoonsPlayer_HasFourOfAKind(t *testing.T) {
	p := NewSpoonsPlayer(true)
	spSetHand(p, spCard(CardDesignSpade, 7), spCard(CardDesignHeart, 7),
		spCard(CardDesignClover, 7), spCard(CardDesignDiamond, 7))
	assert.True(t, p.HasFourOfAKind())

	spSetHand(p, spCard(CardDesignSpade, 7), spCard(CardDesignHeart, 7),
		spCard(CardDesignClover, 7), spCard(CardDesignDiamond, 8))
	assert.False(t, p.HasFourOfAKind())

	// 4 枚未満は常に false。
	spSetHand(p, spCard(CardDesignSpade, 7), spCard(CardDesignHeart, 7))
	assert.False(t, p.HasFourOfAKind())
}

func TestSpoonsPlayer_AddLetterElimination(t *testing.T) {
	p := NewSpoonsPlayer(false)
	for i := 0; i < SpoonsMaxLetters-1; i++ {
		assert.False(t, p.AddLetter())
		assert.Equal(t, i+1, p.GetLetters())
		assert.False(t, p.GetEliminated())
	}
	// 6 文字目で脱落。
	assert.True(t, p.AddLetter())
	assert.True(t, p.GetEliminated())
	assert.True(t, p.GetIsFinished())
	assert.Equal(t, SpoonsMaxLetters, p.GetLetters())
}

// TestSpoons_HumanPassCompletesFourOfAKind は、人間が受け取った札で
// フォーオブアカインドを揃え、即座にスプーンを掴みグラブウィンドウが
// 開くことを確認する。
func TestSpoons_HumanPassCompletesFourOfAKind(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	// 人間 (0): 7 が 3 枚 + 余分な 1 枚。incoming に 7 を流して 4 枚目を作る。
	spSetHand(g.GetPlayer(0), spCard(CardDesignSpade, 7), spCard(CardDesignHeart, 7),
		spCard(CardDesignClover, 7), spCard(CardDesignDiamond, 9))
	// passedCard を 4 枚目の 7 にする (人間が受け取る札)。
	spSetPassedCard(g, spCard(CardDesignDiamond, 7))
	spSetPhase(g, SpoonsPhasePass)
	spSetCurrentPlayer(g, 0)

	// 9 (index 3) を渡せば 7 が 4 枚残りフォーオブアカインド成立。
	require.NoError(t, g.PlayerPass(3))
	// 人間がフォーオブアカインドを揃えた瞬間にスプーンを掴み、最初の取得者となる
	// (どちらも決定的)。
	assert.Equal(t, 0, g.GetFirstGrabberIdx())
	assert.True(t, g.GetPlayer(0).GetHasSpoon())
	// 人間が掴んだ後、残りスプーンを CPU が奪い合う。取りこぼしは乱数なので、まだ
	// グラブ中か既にラウンドが解決済みかのどちらも正当。
	assert.Contains(t, []SpoonsPhase{SpoonsPhaseGrab, SpoonsPhaseRoundEnd}, g.GetPhase())
}

func TestSpoons_PlayerPassErrors(t *testing.T) {
	g := newTestSpoons()
	g.Reset()

	// 不正インデックス。
	spSetCurrentPlayer(g, 0)
	spSetPhase(g, SpoonsPhasePass)
	assert.ErrorIs(t, g.PlayerPass(999), ErrInvalidCard)

	// 手番でない。
	spSetCurrentPlayer(g, 1)
	assert.ErrorIs(t, g.PlayerPass(0), ErrInvalidPlay)

	// 誤フェーズ。
	spSetPhase(g, SpoonsPhaseGrab)
	assert.ErrorIs(t, g.PlayerPass(0), ErrWrongPhase)

	// ゲーム終了。
	spSetGameEnd(g)
	assert.ErrorIs(t, g.PlayerPass(0), ErrGameEnded)
}

func TestSpoons_GrabSpoonAndLetterAccrual(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	// グラブウィンドウを開いた状態にする: スプーン残り 1、CPU1 が firstGrabber。
	spSetPhase(g, SpoonsPhaseGrab)
	spSetGrabWindow(g, true)
	spSetSpoonsRemaining(g, 1)
	g.GetPlayer(1).SetHasSpoon(true) // CPU1 が既に 1 本

	// 人間がスプーンを掴む → 残り 0 → ラウンド締め。
	require.NoError(t, g.PlayerGrabSpoon())
	assert.True(t, g.GetPlayer(0).GetHasSpoon())
	assert.Equal(t, SpoonsPhaseRoundEnd, g.GetPhase())

	// スプーンを取れなかった 2 と 3 に文字が付く。
	assert.Equal(t, 1, g.GetPlayer(2).GetLetters())
	assert.Equal(t, 1, g.GetPlayer(3).GetLetters())
	assert.Equal(t, 0, g.GetPlayer(0).GetLetters())
	assert.Equal(t, 0, g.GetPlayer(1).GetLetters())
}

func TestSpoons_GrabSpoonErrors(t *testing.T) {
	g := newTestSpoons()
	g.Reset()

	// 誤フェーズ (Pass)。
	spSetPhase(g, SpoonsPhasePass)
	assert.ErrorIs(t, g.PlayerGrabSpoon(), ErrWrongPhase)

	// 既に掴んでいる。
	spSetPhase(g, SpoonsPhaseGrab)
	spSetSpoonsRemaining(g, 1)
	g.GetPlayer(0).SetHasSpoon(true)
	assert.ErrorIs(t, g.PlayerGrabSpoon(), ErrInvalidPlay)

	// ゲーム終了。
	spSetGameEnd(g)
	assert.ErrorIs(t, g.PlayerGrabSpoon(), ErrGameEnded)
}

// TestSpoons_FullCpuGameTerminates は、難易度を変えてフルCPU対戦を回し、
// 必ず単一の勝者で終了することを確認する (停止保証 + コスト)。
func TestSpoons_FullCpuGameTerminates(t *testing.T) {
	for _, diff := range []SpoonsCpuDifficulty{SpoonsCpuEasy, SpoonsCpuNormal, SpoonsCpuHard} {
		diff := diff
		t.Run("difficulty", func(t *testing.T) {
			players := []*SpoonsPlayer{
				NewSpoonsPlayer(false),
				NewSpoonsPlayer(false),
				NewSpoonsPlayer(false),
				NewSpoonsPlayer(false),
			}
			g := NewSpoons(NewTrumpCards(0), players, SpoonsConfig{CpuDifficulty: diff})
			g.SetRand(rand.New(rand.NewSource(42)))
			g.Reset()

			steps := 0
			for !g.GetGameEndFlag() {
				steps++
				require.Less(t, steps, 2_000_000, "game did not terminate")
				switch g.GetPhase() {
				case SpoonsPhasePass, SpoonsPhaseGrab:
					g.CpuPlay()
				case SpoonsPhaseRoundEnd:
					g.NextRound()
				default:
				}
			}
			winner := g.GetWinnerIdx()
			require.GreaterOrEqual(t, winner, 0)
			// 勝者以外は全員脱落している。
			alive := 0
			for i := 0; i < SpoonsPlayerCnt; i++ {
				if !g.GetPlayer(i).GetEliminated() {
					alive++
				}
			}
			assert.Equal(t, 1, alive)
			assert.False(t, g.GetPlayer(winner).GetEliminated())
		})
	}
}

// TestSpoons_CpuPassAdvancesAndFeeds は、CPU パスがドローパイルを供給しつつ
// 手番を進めることを確認する。
func TestSpoons_CpuPassAdvancesAndFeeds(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	spSetCurrentPlayer(g, 1) // CPU1 の番
	before := g.GetCurrentPlayerIdx()
	g.CpuPlay()
	// 手番が前進している (または four で grab に遷移)。
	assert.True(t, g.GetCurrentPlayerIdx() != before || g.GetPhase() == SpoonsPhaseGrab)
}

func TestSpoons_NextRoundNoOpAfterGameEnd(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	spSetGameEnd(g)
	round := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, round, g.GetRoundNumber())
}

func TestSpoons_JSONRoundTrip(t *testing.T) {
	g := newTestSpoons()
	g.Reset()
	g.GetPlayer(2).SetLetters(3)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored Spoons
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetSpoonsRemaining(), restored.GetSpoonsRemaining())
	assert.Equal(t, 3, restored.GetPlayer(2).GetLetters())
}

func TestSpoons_UnmarshalRejectsBadState(t *testing.T) {
	cases := []string{
		`{"cf":{"cd":99}}`,                                  // bad config
		`{"cf":{"cd":1},"ph":99}`,                           // bad phase
		`{"cf":{"cd":1},"ph":0,"ps":[null,null,null,null]}`, // nil players
	}
	for _, c := range cases {
		var g Spoons
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}
}

func TestSpoonsConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultSpoonsConfig().Validate())
	assert.Error(t, SpoonsConfig{CpuDifficulty: SpoonsCpuDifficulty(99)}.Validate())
	assert.Greater(t, SpoonsConfig{CpuDifficulty: SpoonsCpuEasy}.GrabMissChance(),
		SpoonsConfig{CpuDifficulty: SpoonsCpuHard}.GrabMissChance())
}
