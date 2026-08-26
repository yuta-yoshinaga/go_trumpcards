//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJulepe(t *testing.T) *Julepe {
	t.Helper()
	r := NewDefaultJulepe()
	r.Reset()
	return r
}

// newTestJulepeWith 指定人数で作る
func newTestJulepeWith(t *testing.T, n int) *Julepe {
	t.Helper()
	r := NewDefaultJulepe()
	r.SetConfig(JulepeConfig{PlayerCnt: n, Rounds: JulepeRoundsDefault})
	r.Reset()
	return r
}

// julepeTotalChips 全員の持ちチップ + ポット。**ポット制の不変量。**
func julepeTotalChips(r *Julepe) int {
	total := r.GetPot()
	for i := range r.GetPlayerCnt() {
		total += r.GetPlayer(i).GetChips()
	}
	return total
}

// --- 可変人数 ---

// **3〜5 人で成立する。** これがフレペの特徴で、Loo（4人固定）との違い。
func TestJulepe_SupportsThreeToFivePlayers(t *testing.T) {
	for n := JulepePlayerCntMin; n <= JulepePlayerCntMax; n++ {
		t.Run(string(rune('0'+n))+" players", func(t *testing.T) {
			r := newTestJulepeWith(t, n)
			assert.Equal(t, n, r.GetPlayerCnt())
			for i := range n {
				assert.Equal(t, JulepeHandSize, r.GetPlayer(i).GetCardsSize(), "player %d", i)
			}
			// 5 人でも 5×5+1=26 枚で 32 枚に収まる。
			assert.NotNil(t, r.GetUpCard(), "切り札を決める 1 枚が残る")
			assert.Equal(t, JulepeStartingChips*n, julepeTotalChips(r))
		})
	}
}

func TestJulepe_ResetDealsAndAntes(t *testing.T) {
	r := newTestJulepe(t)

	for i := range r.GetPlayerCnt() {
		assert.Equal(t, JulepeHandSize, r.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, JulepeStartingChips-JulepeAnte, r.GetPlayer(i).GetChips(), "player %d", i)
	}
	assert.Equal(t, JulepeAnte*r.GetPlayerCnt(), r.GetPot())
	// **配り終えた直後は選択フェーズ。** いきなりプレイに入らない。
	assert.Equal(t, JulepePhaseDecide, r.GetPhase())
	assert.True(t, r.IsDecidePhase())
	assert.Equal(t, 1, r.GetRoundNumber())
}

// **切り札は表向きの 1 枚で決まる。** issue には切り札の規定が無い。
func TestJulepe_TrumpComesFromTheUpCard(t *testing.T) {
	r := newTestJulepe(t)
	up := r.GetUpCard()
	require.NotNil(t, up)
	assert.Equal(t, up.GetDesign(), r.GetTrumpSuit())
}

// --- 参加 / 降りる ---

func TestJulepe_DecideStartsPlay(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(true))

	assert.True(t, r.GetPlayer(0).GetInRound())
	assert.True(t, r.GetPlayer(0).GetDecided())
	// CPU も全員選び終える。
	for i := range r.GetPlayerCnt() {
		assert.True(t, r.GetPlayer(i).GetDecided(), "player %d", i)
	}
	assert.NotEqual(t, JulepePhaseDecide, r.GetPhase())
}

func TestJulepe_PassKeepsYouOut(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(false))

	assert.False(t, r.GetPlayer(0).GetInRound(), "降りたら参加者でない")
	assert.True(t, r.GetPlayer(0).GetDecided())
}

func TestJulepe_DecideRejections(t *testing.T) {
	t.Run("twice", func(t *testing.T) {
		r := newTestJulepe(t)
		require.NoError(t, r.Decide(true))
		assert.Error(t, r.Decide(true), "二度は選べない")
	})
	t.Run("after game end", func(t *testing.T) {
		r := newTestJulepe(t)
		r.GiveUp()
		assert.Error(t, r.Decide(true))
	})
}

// 降りた人はプレイに参加しない。
func TestJulepe_PassedPlayerCannotPlay(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(false))
	r.SetPhaseForTest(JulepePhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, r.PlayerPlay(0), "降りた人は札を出せない")
}

// --- ポット ---

// **参加して 0 トリックだと追加で払う。** これがフレペのリスク。
func TestJulepe_ZeroTrickPlayerPaysPenalty(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 2
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
	}
	r.GetPlayer(0).SetRoundTricks(0)
	r.GetPlayer(1).SetRoundTricks(JulepeTricksPerRound)
	chipsBefore := r.GetPlayer(0).GetChips()

	r.FinishRoundForTest()

	assert.Equal(t, chipsBefore-JulepeMissPenalty, r.GetPlayer(0).GetChips(),
		"参加して 0 トリックなら追加支払い")
}

// **降りた人は 0 トリックでも追加で払わない。** 負のコントロール。
func TestJulepe_PassedPlayerPaysNoPenalty(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 2
	r.GetPlayer(0).SetInRound(false)
	r.GetPlayer(1).SetInRound(true)
	r.GetPlayer(1).SetRoundTricks(JulepeTricksPerRound)
	chipsBefore := r.GetPlayer(0).GetChips()

	r.FinishRoundForTest()

	assert.Equal(t, chipsBefore, r.GetPlayer(0).GetChips(), "降りていれば失うのはアンティだけ")
}

// トリック数に応じてポットを按分する。
func TestJulepe_PotIsSharedByTrickCount(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 2
	r.SetPotForTest(20)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
		r.GetPlayer(i).SetChips(0)
	}
	r.GetPlayer(0).SetRoundTricks(3)
	r.GetPlayer(1).SetRoundTricks(2)
	// 2,3 番は 0 トリックなので追加支払いが発生する。

	r.FinishRoundForTest()

	// ポット 20 + 追加 5×2 = 30、5 トリックで割って 1 トリック 6。
	assert.Equal(t, 18, r.GetPlayer(0).GetChips(), "3 トリック × 6")
	assert.Equal(t, 12, r.GetPlayer(1).GetChips(), "2 トリック × 6")
}

// **端数は次ラウンドへ残す。** 配り切れないぶんを消すとチップが減る。
func TestJulepe_RemainderStaysInThePot(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 2
	r.SetPotForTest(11) // 2 トリックで割ると 5 余り 1
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(false)
	}
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)
	before := julepeTotalChips(r)

	r.FinishRoundForTest()

	assert.Equal(t, 1, r.GetPot(), "端数 1 が残る")
	assert.Equal(t, before, julepeTotalChips(r), "チップの総量は変わらない")
}

// 全員降りたらポットは持ち越し。
func TestJulepe_AllPassCarriesThePot(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 3
	pot := r.GetPot()
	before := julepeTotalChips(r)

	// 人間が降り、CPU も全員降りた状態を作る。
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetDecided(true)
		r.GetPlayer(i).SetInRound(false)
	}
	r.StartPlayIfReadyForTest()

	assert.Equal(t, 0, r.GetActiveCount())
	assert.Equal(t, pot, r.GetPot(), "ポットはそのまま残る")
	assert.Equal(t, before, julepeTotalChips(r))
}

// **チップは生まれも消えもしない。** ポット制の最重要不変量。
func TestJulepe_ChipsAreConserved(t *testing.T) {
	for n := JulepePlayerCntMin; n <= JulepePlayerCntMax; n++ {
		for range 10 {
			r := NewDefaultJulepe()
			r.SetConfig(JulepeConfig{PlayerCnt: n, Rounds: JulepeRoundsDefault})
			r.Reset()
			want := JulepeStartingChips * n
			require.Equal(t, want, julepeTotalChips(r), "配り直後 (%d人)", n)

			guard := 0
			for !r.GetGameEndFlag() && guard < 2000 {
				guard++
				switch {
				case r.IsDecidePhase():
					require.NoError(t, r.Decide(true))
				case r.GetPhase() == JulepePhaseRoundEnd:
					r.NextRound()
				case r.IsHumanTurn():
					valid := r.GetValidPlayIndices(0)
					require.NotEmpty(t, valid)
					require.NoError(t, r.PlayerPlay(valid[0]))
				default:
					r.CpuPlay()
				}
				require.Equal(t, want, julepeTotalChips(r), "1 手ごと (%d人)", n)
			}
			require.True(t, r.GetGameEndFlag(), "%d 人でゲームが終わる", n)
			// **終局時にポットが残っていてはいけない。** 総量が保たれていても、
			// 盤上に取り残されたチップは GetChips() に入らず勝敗に反映されない
			// （保存則の assert だけでは捕まらない）。
			require.Equal(t, 0, r.GetPot(), "終局時にポットは空 (%d人)", n)
		}
	}
}

// **最終ラウンドには持ち越し先が無い。** 端数をそのまま残すと、チップが
// 盤上に取り残されたままゲームが終わる。
func TestJulepe_FinalRoundDrainsTheRemainder(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 1
	r.SetPotForTest(11) // 2 トリックで割ると 5 余り 1
	before := julepeTotalChips(r)
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)

	r.FinishRoundForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, 0, r.GetPot(), "端数も配り切る")
	assert.Equal(t, before, julepeTotalChips(r), "チップの総量は変わらない")
}

// 最終ラウンドで全員降りた場合も、ポットを残さず全員に返す。
func TestJulepe_FinalRoundAllPassSplitsThePot(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 1
	before := julepeTotalChips(r)
	pot := r.GetPot()
	require.Positive(t, pot)

	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetDecided(true)
		r.GetPlayer(i).SetInRound(false)
	}
	r.StartPlayIfReadyForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, 0, r.GetPot(), "誰もトリックを取っていなくても残さない")
	assert.Equal(t, before, julepeTotalChips(r))
}

// 中間ラウンドでは今までどおり持ち越す。負のコントロール。
func TestJulepe_MidGameRoundStillCarriesTheRemainder(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 3
	r.roundNumber = 1
	r.SetPotForTest(11)
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)

	r.FinishRoundForTest()

	assert.False(t, r.GetGameEndFlag())
	assert.Equal(t, 1, r.GetPot(), "まだ先があるので端数は持ち越す")
}

// --- 切り札とフォロー義務 ---

func TestJulepe_TrumpBeatsLeadSuit(t *testing.T) {
	r := newTestJulepe(t)
	r.SetTrumpSuitForTest(CardDesignHeart)
	r.leadPlayerIdx = 0
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
	}
	r.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 7, false)},
	})
	assert.Equal(t, 1, r.trickWinner(), "切り札の 7 がスペードの A に勝つ")
}

func TestJulepeBeats(t *testing.T) {
	trump, lead := CardDesignHeart, CardDesignSpade
	assert.True(t, julepeBeats(NewCard(trump, 7, false), NewCard(lead, 1, false), lead, trump), "切り札が勝つ")
	assert.False(t, julepeBeats(NewCard(lead, 1, false), NewCard(trump, 7, false), lead, trump), "切り札には勝てない")
	assert.True(t, julepeBeats(NewCard(trump, 13, false), NewCard(trump, 12, false), lead, trump), "切り札同士は強さ")
	assert.False(t, julepeBeats(NewCard(CardDesignClover, 1, false), NewCard(lead, 7, false), lead, trump),
		"リードでも切り札でもない札は勝てない")
	assert.True(t, julepeBeats(NewCard(lead, 1, false), NewCard(CardDesignClover, 13, false), lead, trump),
		"リードのスートは無関係のスートに勝つ")
}

func TestJulepeRank_AceIsHighest(t *testing.T) {
	assert.Greater(t, julepeRank(NewCard(CardDesignSpade, 1, false)), julepeRank(NewCard(CardDesignSpade, 13, false)))
	assert.Equal(t, 0, julepeRank(nil))
}

func TestJulepe_MustFollowSuit(t *testing.T) {
	r := newTestJulepe(t)
	p := r.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 8, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	r.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}})

	assert.Equal(t, []int{0}, r.GetValidPlayIndices(1))
}

func TestJulepe_GetValidPlayIndicesOutOfRange(t *testing.T) {
	r := newTestJulepe(t)
	assert.Nil(t, r.GetValidPlayIndices(-1))
	assert.Nil(t, r.GetValidPlayIndices(r.GetPlayerCnt()))
}

// --- プレイ ---

func TestJulepe_PlayerPlayRejections(t *testing.T) {
	t.Run("decide phase", func(t *testing.T) {
		r := newTestJulepe(t)
		assert.Error(t, r.PlayerPlay(0), "選択前は出せない")
	})
	t.Run("not your turn", func(t *testing.T) {
		r := newTestJulepe(t)
		require.NoError(t, r.Decide(true))
		r.SetPhaseForTest(JulepePhasePlay)
		r.SetCurrentPlayerIdxForTest(1)
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		r := newTestJulepe(t)
		r.gameEndFlag = true
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		r := newTestJulepe(t)
		require.NoError(t, r.Decide(true))
		r.SetPhaseForTest(JulepePhasePlay)
		r.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, r.PlayerPlay(99))
		assert.Error(t, r.PlayerPlay(-1))
	})
}

func TestJulepe_CpuPlayIsANoOpOnHumanTurn(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(JulepePhasePlay)
	r.SetCurrentPlayerIdxForTest(0)
	before := r.GetPlayer(0).GetCardsSize()
	r.CpuPlay()
	assert.Equal(t, before, r.GetPlayer(0).GetCardsSize())
}

// --- ゲーム終了 ---

func TestJulepe_MostChipsWins(t *testing.T) {
	r := newTestJulepe(t)
	r.GetPlayer(0).SetChips(30)
	r.GetPlayer(1).SetChips(90)
	r.GetPlayer(2).SetChips(45)
	r.GetPlayer(3).SetChips(45)

	r.FinishGameForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, JulepePhaseGameEnd, r.GetPhase())
	assert.Equal(t, 1, r.GetWinnerIdx())
}

func TestJulepe_TieHasNoWinner(t *testing.T) {
	r := newTestJulepe(t)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetChips(50)
	}
	r.FinishGameForTest()
	assert.Equal(t, -1, r.GetWinnerIdx())
}

func TestJulepe_NextRoundRedealsAndKeepsChips(t *testing.T) {
	r := newTestJulepe(t)
	r.config.Rounds = 3
	r.GetPlayer(0).SetChips(80)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(JulepePhaseRoundEnd)
	dealer := r.GetDealerIdx()

	r.NextRound()

	assert.Equal(t, 2, r.GetRoundNumber())
	// **次のラウンドも選択フェーズから始まる。**
	assert.Equal(t, JulepePhaseDecide, r.GetPhase())
	assert.Equal(t, (dealer+1)%r.GetPlayerCnt(), r.GetDealerIdx(), "ディーラーが回る")
	assert.Equal(t, 80-JulepeAnte, r.GetPlayer(0).GetChips(), "チップは持ち越し、アンティを払う")
	assert.False(t, r.GetPlayer(0).GetInRound(), "参加状態はラウンドごとに消える")
	for i := range r.GetPlayerCnt() {
		assert.Equal(t, JulepeHandSize, r.GetPlayer(i).GetCardsSize())
	}
}

func TestJulepe_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	r := newTestJulepe(t)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())

	r.gameEndFlag = true
	r.SetPhaseForTest(JulepePhaseRoundEnd)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())
}

func TestJulepe_GiveUp(t *testing.T) {
	r := newTestJulepe(t)
	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, JulepePhaseGameEnd, r.GetPhase())
	assert.Equal(t, -1, r.GetWinnerIdx())

	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
}

func TestJulepe_GetPlayerOutOfRange(t *testing.T) {
	r := newTestJulepe(t)
	assert.Nil(t, r.GetPlayer(-1))
	assert.Nil(t, r.GetPlayer(r.GetPlayerCnt()))
}

func TestJulepe_Config(t *testing.T) {
	r := newTestJulepe(t)
	assert.Equal(t, JulepePlayerCntDefault, r.GetConfig().PlayerCnt)
	assert.Equal(t, JulepeRoundsDefault, r.GetConfig().Rounds)

	assert.NoError(t, JulepeConfig{PlayerCnt: 3, Rounds: 1}.Validate())
	assert.NoError(t, JulepeConfig{PlayerCnt: 5, Rounds: 12}.Validate())
	assert.Error(t, JulepeConfig{PlayerCnt: 2, Rounds: 4}.Validate(), "2 人は不可")
	assert.Error(t, JulepeConfig{PlayerCnt: 6, Rounds: 4}.Validate(), "6 人は不可")
	assert.Error(t, JulepeConfig{PlayerCnt: 4, Rounds: 0}.Validate())
	assert.Error(t, JulepeConfig{PlayerCnt: 4, Rounds: 13}.Validate())
}

// --- ヒント ---

// **選択フェーズでは出す札ではなく、参加するかどうかを助言する。**
func TestJulepe_GetHint_DecidePhaseAdvisesPlayOrPass(t *testing.T) {
	r := newTestJulepe(t)
	h := r.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex, "選択フェーズでは札を指さない")
	assert.Contains(t, []string{"julepePlayIn", "julepePassOut"}, h.Reason)
}

// 強い手なら参加、弱い手なら降りるを勧める。両側を踏む。
func TestJulepe_GetHint_DecideBothWays(t *testing.T) {
	strong := newTestJulepe(t)
	strong.SetTrumpSuitForTest(CardDesignHeart)
	p := strong.GetPlayer(0)
	p.Reset()
	for _, v := range []int{1, 13, 12, 11, 10} {
		p.AddCard(NewCard(CardDesignHeart, v, false)) // 切り札ばかり
	}
	assert.Equal(t, "julepePlayIn", strong.GetHint().Reason)

	weak := newTestJulepe(t)
	weak.SetTrumpSuitForTest(CardDesignHeart)
	q := weak.GetPlayer(0)
	q.Reset()
	for _, v := range []int{7, 8, 9, 10, 11} {
		q.AddCard(NewCard(CardDesignSpade, v, false)) // 切り札もエースも無い
	}
	assert.Equal(t, "julepePassOut", weak.GetHint().Reason)
}

func TestJulepe_GetHint_PlayPhaseSuggestsACard(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(JulepePhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	h := r.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, r.GetValidPlayIndices(0), *h.CardIndex)
		assert.Equal(t, "julepeTakeTrick", h.Reason, "まだ 1 トリックも取っていない")
	}
}

// 1 トリック取ったあとは追加支払いを免れているので、狙いが変わる。
// **「安全」の線は規定トリック数であって 1 トリックではない。**
//
// クローン元のラムスは 1 トリックで罰を免れたので、その値で試すと
// 新旧どちらの実装も同じ答えを返し、違いが見えない。既定の 4 人卓では
// 規定は 2 なので、1 トリックは**まだ安全ではない**。
func TestJulepe_GetHint_ChangesOnceSafe(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(JulepePhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	// **参加者を固定する。** 規定トリック数は参加人数から出るので、CPU の
	// 参加判断に任せると配りごとに規定が動く。全員が降りた配りでは人間が
	// 1 人だけ残って規定が 1 になり、この検査自体が落ちる (CI 実測)。
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
	}
	required := r.GetRequiredTricks()
	require.Greater(t, required, 1, "規定が 1 だと `> 0` の実装と区別が付かない")

	// 規定に 1 つ足りない: まだ安全ではない。
	r.GetPlayer(0).SetRoundTricks(required - 1)
	h := r.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "julepeTakeTrick", h.Reason, "規定に届いていないのに安全と言っている")

	// 規定ちょうどで安全になる。
	r.GetPlayer(0).SetRoundTricks(required)
	h = r.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "julepeAlreadySafe", h.Reason)
}

func TestJulepe_GetHint_NilAfterGameEnd(t *testing.T) {
	r := newTestJulepe(t)
	r.GiveUp()
	assert.Nil(t, r.GetHint())
}

// --- JSON 往復 ---

func TestJulepe_JSONRoundTrip(t *testing.T) {
	r := newTestJulepe(t)
	require.NoError(t, r.Decide(true))
	r.GetPlayer(0).SetChips(77)
	r.GetPlayer(1).SetInRound(false)
	r.SetPotForTest(31)
	r.roundNumber = 2
	r.SetTrickNumberForTest(3)

	data, err := json.Marshal(r)
	require.NoError(t, err)

	restored := NewDefaultJulepe()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 77, restored.GetPlayer(0).GetChips())
	assert.True(t, restored.GetPlayer(0).GetInRound(), "参加状態が往復する")
	assert.False(t, restored.GetPlayer(1).GetInRound(), "降りた人は降りたまま")
	assert.Equal(t, 31, restored.GetPot())
	assert.Equal(t, r.GetTrumpSuit(), restored.GetTrumpSuit(), "切り札が往復する")
	assert.Equal(t, 2, restored.GetRoundNumber())
	assert.Equal(t, r.GetConfig().PlayerCnt, restored.GetConfig().PlayerCnt)
	if up := r.GetUpCard(); up != nil {
		require.NotNil(t, restored.GetUpCard())
		assert.Equal(t, up.GetDesign(), restored.GetUpCard().GetDesign())
	}
}

// **5 人の局面を 4 人の器へ復元しても壊れない。** 可変人数ゲーム固有の危険。
func TestJulepe_JSONRoundTripWithFivePlayers(t *testing.T) {
	r := newTestJulepeWith(t, 5)
	data, err := json.Marshal(r)
	require.NoError(t, err)

	restored := NewDefaultJulepe() // 4 人で作られている
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, 5, restored.GetPlayerCnt(), "席数も復元される")
}

func TestJulepe_UnmarshalRejectsGarbage(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("not json"), NewDefaultJulepe()))
}

// 範囲外の席数は壊れた KV。読み込みを拒否する。
func TestJulepe_UnmarshalRejectsBadPlayerCount(t *testing.T) {
	r := newTestJulepe(t)
	data, err := json.Marshal(r)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["pl"] = []any{} // 0 人
	broken, err := json.Marshal(raw)
	require.NoError(t, err)

	assert.Error(t, json.Unmarshal(broken, NewDefaultJulepe()))
}

// 壊れた設定も拒否する。Rounds=0 は「配った直後に終わる」盤面になる。
func TestJulepe_UnmarshalRejectsBadConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"rounds is zero", func(raw map[string]any) {
			raw["cf"] = map[string]any{"pc": 4, "rd": 0}
		}},
		{"player count out of range", func(raw map[string]any) {
			raw["cf"] = map[string]any{"pc": 9, "rd": 4}
		}},
		{"config disagrees with the stored seats", func(raw map[string]any) {
			raw["cf"] = map[string]any{"pc": 5, "rd": 4} // 席は 4 つしか入っていない
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestJulepe(t)
			data, err := json.Marshal(r)
			require.NoError(t, err)
			var raw map[string]any
			require.NoError(t, json.Unmarshal(data, &raw))
			tc.mutate(raw)
			broken, err := json.Marshal(raw)
			require.NoError(t, err)

			assert.Error(t, json.Unmarshal(broken, NewDefaultJulepe()))
		})
	}
}

func TestJulepe_ActionLog(t *testing.T) {
	r := newTestJulepe(t)
	assert.NotEmpty(t, r.GetActionLog())
}

// TestJulepe_RequiredTricksScalesWithTheTable は、規定トリック数が
// **参加人数から導かれる**ことを見る。
//
// クローン元のラムスは「1 トリックも取れなければ罰」という固定の線だった。
// 手で並べた表ではなく手札枚数から計算するので、配り枚数を変えれば追随する。
func TestJulepe_RequiredTricksScalesWithTheTable(t *testing.T) {
	// 5 トリックを n 人で分けた切り上げ。
	for n, want := range map[int]int{2: 3, 3: 2, 4: 2, 5: 1} {
		assert.Equal(t, want, JulepeRequiredTricks(n), "参加 %d 人", n)
	}

	// **人数が増えるほど線は下がる (単調非増加)。** 逆向きになっていないこと。
	for n := 2; n < 8; n++ {
		assert.LessOrEqual(t, JulepeRequiredTricks(n+1), JulepeRequiredTricks(n),
			"参加 %d→%d 人で規定が上がっている", n, n+1)
	}

	// 参加が 1 人なら競争が無いので 1 トリック。
	assert.Equal(t, 1, JulepeRequiredTricks(1))

	// **手札枚数から導いていること。** 定数を並べただけなら、この関係は
	// たまたま成り立っているだけになる。
	assert.Equal(t, JulepeTricksPerRound, JulepeRequiredTricks(1)*JulepeTricksPerRound/
		JulepeRequiredTricks(1))
}

// TestJulepe_BeastDoublesTheNextAnte は、規定トリック数に届かなかった席が
// **次のラウンドで倍のアンティを払う**ことを見る。
//
// これがラムスとの核心の差。ラムスは「そのラウンドで罰金を払って終わり」だが、
// Julepe は状態がラウンドをまたぐ。
func TestJulepe_BeastDoublesTheNextAnte(t *testing.T) {
	r := newTestJulepe(t)
	r.SetPotForTest(0)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
		r.GetPlayer(i).SetRoundTricks(0)
	}
	// 席 0 だけが規定を満たす。
	r.GetPlayer(0).SetRoundTricks(JulepeTricksPerRound)

	r.FinishRoundForTest()
	beast := r.GetBeastForTest()
	require.NotEmpty(t, beast)
	assert.False(t, beast[0], "規定を満たした席は beast にならない")
	assert.True(t, beast[1], "規定に届かなかった席が beast")

	// 次のラウンドで倍払い。
	chipsBefore := r.GetPlayer(1).GetChips()
	normalBefore := r.GetPlayer(0).GetChips()
	r.DealRoundForTest()
	assert.Equal(t, chipsBefore-JulepeAnte*2, r.GetPlayer(1).GetChips(), "beast は倍払い")
	assert.Equal(t, normalBefore-JulepeAnte, r.GetPlayer(0).GetChips(), "beast でない席は通常のアンティ")

	// **倍払いは 1 回きり。** 集め終えたらフラグが消えること。
	assert.Empty(t, r.GetBeastForTest(), "beast が持ち越されている (永久に倍払いになる)")
}

// TestJulepe_BeastOnAPartialHaul は、**トリックを取っていても規定に
// 届かなければ beast になる**ことを見る。
//
// クローン元のラムスは「0 トリック」だけを罰するので、0 トリックの席で
// 試すと新旧どちらの実装でも同じ答えになり、違いが見えない。
// 参加 3 人 (規定 2) で 1 トリックだけ取った席を置く。
func TestJulepe_BeastOnAPartialHaul(t *testing.T) {
	r := newTestJulepe(t)
	r.SetPotForTest(0)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(false)
		r.GetPlayer(i).SetRoundTricks(0)
	}
	// 参加は 3 人 → 規定 2 トリック。
	for _, i := range []int{0, 1, 2} {
		r.GetPlayer(i).SetInRound(true)
	}
	require.Equal(t, 2, JulepeRequiredTricks(3))

	r.GetPlayer(0).SetRoundTricks(3) // 規定を満たす
	r.GetPlayer(1).SetRoundTricks(1) // **取っているが届かない**
	r.GetPlayer(2).SetRoundTricks(1) // 同上

	r.FinishRoundForTest()
	beast := r.GetBeastForTest()
	require.NotEmpty(t, beast)
	assert.False(t, beast[0], "規定を満たした席は beast にならない")
	assert.True(t, beast[1], "1 トリックでも規定 2 に届かなければ beast")
	assert.True(t, beast[2])
}
