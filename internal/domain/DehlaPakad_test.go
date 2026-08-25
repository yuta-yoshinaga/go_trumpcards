//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDehlaPakadForTest(t *testing.T) *DehlaPakad {
	t.Helper()
	d := NewDefaultDehlaPakad()
	d.Reset()
	return d
}

// dehlaPakadPlayHand は 1 ハンドを最後まで打つ。
func dehlaPakadPlayHand(t *testing.T, d *DehlaPakad) {
	t.Helper()
	for step := 0; step < 2000; step++ {
		switch d.GetPhase() {
		case DehlaPakadPhaseSelectTrump:
			if d.IsHumanTurn() {
				require.NoError(t, d.SelectTrump(CardDesignSpade))
				continue
			}
			d.CpuSelectTrump()
		case DehlaPakadPhasePlay:
			if d.IsHumanTurn() {
				valid := d.GetPlayableIndices(d.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
				require.NoError(t, d.PlayerPlay(valid[0]))
				continue
			}
			d.CpuPlay()
		default:
			return
		}
	}
	t.Fatal("ハンドが終わらない")
}

// **配りは 5 枚 → 宣言 → 残り 8 枚。** 13 枚配ってから訊くと、宣言する人が
// 手札全部を見てから決められてしまう。
func TestDehlaPakad_DealsFiveBeforeTheTrumpIsCalled(t *testing.T) {
	d := newDehlaPakadForTest(t)
	assert.Equal(t, DehlaPakadPhaseSelectTrump, d.GetPhase())
	for i, p := range d.GetPlayers() {
		assert.Equal(t, DehlaPakadFirstBatch, p.GetCardsSize(), "席 %d の手札", i)
	}
	assert.Equal(t, -1, d.GetTrumpSuit())

	require.NoError(t, d.SelectTrump(CardDesignHeart))
	assert.Equal(t, CardDesignHeart, d.GetTrumpSuit())
	assert.Equal(t, DehlaPakadPhasePlay, d.GetPhase())
	total := 0
	for i, p := range d.GetPlayers() {
		assert.Equal(t, DehlaPakadHandSize, p.GetCardsSize(), "席 %d の手札", i)
		total += p.GetCardsSize()
	}
	assert.Equal(t, 52, total)
}

// **切り札を決めるのは親の右隣。** 反時計回りなので次の席。
func TestDehlaPakad_TrumpChooserIsToTheDealersRight(t *testing.T) {
	d := newDehlaPakadForTest(t)
	assert.Equal(t, DehlaPakadNextSeat(d.GetDealerIdx()), d.GetTrumpChooserIdx())
	assert.Equal(t, d.GetTrumpChooserIdx(), d.GetLeadPlayerIdx(), "宣言した席がリードする")
}

func TestDehlaPakad_RejectsAnInvalidTrumpSuit(t *testing.T) {
	d := newDehlaPakadForTest(t)
	assert.Error(t, d.SelectTrump(0))
	assert.Error(t, d.SelectTrump(9))
	assert.Equal(t, -1, d.GetTrumpSuit())
}

// **相方は向かい。** 隣は必ず相手。
func TestDehlaPakadTeamOf(t *testing.T) {
	assert.Equal(t, DehlaPakadTeamOf(0), DehlaPakadTeamOf(2))
	assert.Equal(t, DehlaPakadTeamOf(1), DehlaPakadTeamOf(3))
	assert.NotEqual(t, DehlaPakadTeamOf(0), DehlaPakadTeamOf(1))
}

func TestDehlaPakadCardStrength(t *testing.T) {
	assert.Equal(t, 14, DehlaPakadCardStrength(NewCard(CardDesignSpade, 1, false)), "A が最強")
	assert.Greater(t, DehlaPakadCardStrength(NewCard(CardDesignSpade, 1, false)),
		DehlaPakadCardStrength(NewCard(CardDesignSpade, 13, false)))
	assert.Equal(t, -1, DehlaPakadCardStrength(nil))
}

// **切り札は台札より強く、台札外は勝てない。**
func TestDehlaPakad_TrickWinner(t *testing.T) {
	d := newDehlaPakadForTest(t)
	d.trumpSuit = CardDesignHeart
	d.phase = DehlaPakadPhasePlay
	d.leadPlayer = 0
	d.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},   // 台札のエース
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 2, false)},   // 切り札の 2 が勝つ
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 1, false)}, // 台札外は無力
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)},
	}
	assert.Equal(t, 1, d.trickWinner())
}

// **札は 2 連勝ではじめて手に入る。** これがこのゲームの心臓部で、
// 「取ったのに 10 が手に入らない」局面を作る。
func TestDehlaPakad_CardsAreOnlyGatheredOnTwoConsecutiveTricks(t *testing.T) {
	d := newDehlaPakadForTest(t)
	require.NoError(t, d.SelectTrump(CardDesignSpade))
	d.trickNumber = 1
	d.prevTrickWinner = -1

	// 1 トリック目: 席 0 が取るが、直前がないので山は残る。
	d.currentTrick = dehlaPakadTrickWonBySeat(0, CardDesignHeart, 2, 3)
	d.resolveTrick()
	assert.Len(t, d.GetCentrePile(), DehlaPakadPlayerCnt, "1 回勝っただけで引き取っている")
	assert.Equal(t, 1, d.GetCentrePileTens())
	assert.Equal(t, []int{0, 0}, d.GetTeamTens(), "まだ誰の 10 でもない")

	// 2 トリック目: 席 1 が取る ── 連勝ではないので山はさらに積み上がる。
	d.currentTrick = dehlaPakadTrickWonBySeat(1, CardDesignClover, 2, -1)
	d.resolveTrick()
	assert.Len(t, d.GetCentrePile(), 2*DehlaPakadPlayerCnt)
	assert.Equal(t, []int{0, 0}, d.GetTeamTens())

	// 3 トリック目: 席 1 が続けて取る ── ここで山ごと引き取る。
	d.currentTrick = dehlaPakadTrickWonBySeat(1, CardDesignDiamond, 3, -1)
	d.resolveTrick()
	assert.Empty(t, d.GetCentrePile(), "2 連勝したのに引き取っていない")
	assert.Equal(t, 1, d.GetTeamTens()[DehlaPakadTeamOf(1)], "10 が相手の組に入っている")
	assert.Equal(t, 0, d.GetTeamTens()[DehlaPakadTeamOf(0)])
}

// **最終トリックだけは無条件で引き取る。** そうしないと山が宙に浮く。
func TestDehlaPakad_TheLastTrickTakesThePileRegardless(t *testing.T) {
	d := newDehlaPakadForTest(t)
	require.NoError(t, d.SelectTrump(CardDesignSpade))
	d.trickNumber = DehlaPakadTrickCount
	d.prevTrickWinner = 2 // 直前は席 2 ＝ 連勝ではない
	d.centrePile = []*Card{NewCard(CardDesignHeart, DehlaPakadTenValue, false)}

	d.currentTrick = dehlaPakadTrickWonBySeat(1, CardDesignClover, 5, -1)
	d.resolveTrick()
	assert.Empty(t, d.GetCentrePile(), "最終トリックで山が残っている")
	assert.Equal(t, 1, d.GetTeamTens()[DehlaPakadTeamOf(1)])
	assert.Equal(t, DehlaPakadPhaseHandEnd, d.GetPhase())
}

// **判定は左右非対称。** 親でない組は 10 が 2 枚で勝ち、親側は 3 枚要る。
func TestDehlaPakad_JudgeHandIsAsymmetric(t *testing.T) {
	d := newDehlaPakadForTest(t)
	d.dealerIdx = 0
	dealerTeam := DehlaPakadTeamOf(0)
	other := 1 - dealerTeam

	for _, tt := range []struct {
		name       string
		dealerTens int
		otherTens  int
		want       int
		kot        bool
	}{
		{"2 対 2 は親でない組の勝ち", 2, 2, other, false},
		{"親側は 3 枚で勝つ", 3, 1, dealerTeam, false},
		{"親でない組は 2 枚で勝つ", 1, 3, other, false},
		{"親側が 4 枚ならコート", 4, 0, dealerTeam, true},
		{"親でない組が 4 枚でもコート", 0, 4, other, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d.teamTens[dealerTeam] = tt.dealerTens
			d.teamTens[other] = tt.otherTens
			res := d.judgeHand()
			assert.Equal(t, tt.want, res.WinnerTeam)
			assert.Equal(t, tt.kot, res.Kot)
			if tt.kot {
				assert.Equal(t, "allTens", res.KotReason)
			}
		})
	}
}

// **親が替わるのは親側が勝ったときだけ。** 据え置きだからこそ連勝が起こりうる。
func TestDehlaPakad_DealerMovesOnlyWhenTheDealersTeamWins(t *testing.T) {
	d := newDehlaPakadForTest(t)
	d.dealerIdx = 0
	dealerTeam := DehlaPakadTeamOf(0)

	// 親でない組が勝つ ── 親は据え置き。
	d.teamTens[dealerTeam] = 1
	d.teamTens[1-dealerTeam] = 3
	d.trickNumber = DehlaPakadTrickCount
	d.finishHand()
	assert.Equal(t, 0, d.GetDealerIdx(), "親でない組が勝ったのに親が動いた")

	// 親側が勝つ ── 親は右へ。
	d.teamTens[dealerTeam] = 3
	d.teamTens[1-dealerTeam] = 1
	d.phase = DehlaPakadPhasePlay
	d.finishHand()
	assert.Equal(t, 1, d.GetDealerIdx(), "親側が勝ったのに親が動かない")
}

// **7 連勝もコートになる。** 4 枚の 10 だけがコートの道ではない。
func TestDehlaPakad_SevenConsecutiveHandsIsAlsoAKot(t *testing.T) {
	d := newDehlaPakadForTest(t)
	d.config.TargetKots = DehlaPakadMaxKots // 途中で終局させない
	d.dealerIdx = 0
	dealerTeam := DehlaPakadTeamOf(0)
	other := 1 - dealerTeam

	for hand := 1; hand <= DehlaPakadStreakForKot; hand++ {
		d.dealerIdx = 0 // 親でない組が勝ち続けるので親は据え置き
		d.teamTens[dealerTeam] = 1
		d.teamTens[other] = 3
		d.phase = DehlaPakadPhasePlay
		d.trickNumber = DehlaPakadTrickCount
		d.finishHand()
		if hand < DehlaPakadStreakForKot {
			assert.Equal(t, 0, d.GetTeamKots()[other], "%d 連勝で早すぎるコート", hand)
		}
	}
	assert.Equal(t, 1, d.GetTeamKots()[other], "7 連勝でコートにならない")
	assert.Equal(t, "streak", d.GetLastResult().KotReason)
}

// 連勝はチームが替わると途切れる。
func TestDehlaPakad_StreakResetsWhenTheOtherTeamWins(t *testing.T) {
	d := newDehlaPakadForTest(t)
	d.dealerIdx = 0
	dealerTeam := DehlaPakadTeamOf(0)

	d.teamTens[dealerTeam] = 1
	d.teamTens[1-dealerTeam] = 3
	d.trickNumber = DehlaPakadTrickCount
	d.finishHand()
	assert.Equal(t, 1, d.GetStreakCount())

	d.dealerIdx = 0
	d.teamTens[dealerTeam] = 3
	d.teamTens[1-dealerTeam] = 1
	d.phase = DehlaPakadPhasePlay
	d.finishHand()
	assert.Equal(t, 1, d.GetStreakCount(), "連勝が途切れていない")
	assert.Equal(t, dealerTeam, d.GetStreakTeam())
}

// **リードスート必従。**
func TestDehlaPakad_MustFollowSuit(t *testing.T) {
	d := newDehlaPakadForTest(t)
	require.NoError(t, d.SelectTrump(CardDesignHeart))
	p := d.GetPlayer(0)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 7, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	d.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
	assert.Equal(t, []int{0}, d.GetPlayableIndices(0))

	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignClover, 8, false))
	assert.Equal(t, []int{0, 1}, d.GetPlayableIndices(0), "台札が無ければ何でも出せる")
}

// **1 ハンドで 10 は 4 枚。** 中央に取り残されないことまで見る。
func TestDehlaPakad_AllFourTensAreAccountedForEachHand(t *testing.T) {
	d := newDehlaPakadForTest(t)
	dehlaPakadPlayHand(t, d)
	require.Equal(t, DehlaPakadPhaseHandEnd, d.GetPhase())
	tens := d.GetTeamTens()
	assert.Equal(t, DehlaPakadTenCnt, tens[0]+tens[1], "10 が宙に浮いている")
	assert.Empty(t, d.GetCentrePile(), "中央に札が残っている")
}

// **1 マッチを通しで打てる。**
func TestDehlaPakad_PlaysAMatchThrough(t *testing.T) {
	d := newDehlaPakadForTest(t)
	for hand := 0; hand < 200 && !d.GetGameEndFlag(); hand++ {
		dehlaPakadPlayHand(t, d)
		d.NextHand()
	}
	require.True(t, d.GetGameEndFlag(), "マッチが終わらない")
	assert.Equal(t, DehlaPakadPhaseGameEnd, d.GetPhase())
	assert.GreaterOrEqual(t, d.GetWinnerTeam(), 0)
	assert.GreaterOrEqual(t, d.GetTeamKots()[d.GetWinnerTeam()], d.GetConfig().TargetKots)
}

// **助言は CPU の難易度に引きずられない。** Easy でランダムに選ぶ関数を
// そのまま使うと、Easy を選んだ人にだけでたらめな札を勧めることになる。
func TestDehlaPakad_HintIgnoresCpuDifficulty(t *testing.T) {
	d := NewDehlaPakad(NewTrumpCards(0), dehlaPakadTestPlayers(), DehlaPakadConfig{
		CpuDifficulty: DehlaPakadCpuDifficultyEasy,
		TargetKots:    DehlaPakadDefaultKots,
	})
	d.Reset()

	// 宣言フェーズ: 何度訊いても同じスートを勧める。
	require.Equal(t, DehlaPakadPhaseSelectTrump, d.GetPhase())
	wantSuit := d.GetHint().TrumpSuit
	for i := 0; i < 20; i++ {
		assert.Equal(t, wantSuit, d.GetHint().TrumpSuit, "%d 回目で切り札がぶれた", i+1)
	}

	require.NoError(t, d.SelectTrump(wantSuit))
	// **人間のリードに固定する。** 配り任せだと合法手が 1 枚しかない局面に
	// 当たり、ランダムでも同じ札が返るので何も試していないことになる。
	d.currentPlayer = 0
	d.leadPlayer = 0
	d.currentTrick = nil
	require.Greater(t, len(d.GetPlayableIndices(0)), 1, "リードなら手札すべてが合法手のはず")

	want := d.GetHint().CardIndices[0]
	for i := 0; i < 20; i++ {
		hint := d.GetHint()
		require.Len(t, hint.CardIndices, 1)
		assert.Equal(t, want, hint.CardIndices[0], "%d 回目でぶれた", i+1)
	}
}

func TestDehlaPakad_HintFollowsThePhase(t *testing.T) {
	d := newDehlaPakadForTest(t)
	assert.Equal(t, "call_longest", d.GetHint().Reason)
	require.NoError(t, d.SelectTrump(CardDesignSpade))
	for d.GetPhase() == DehlaPakadPhasePlay && !d.IsHumanTurn() {
		d.CpuPlay()
	}
	if d.GetPhase() == DehlaPakadPhasePlay {
		hint := d.GetHint()
		assert.Contains(t, []string{"take_the_ten", "keep_the_lead"}, hint.Reason)
		require.Len(t, hint.CardIndices, 1)
		assert.Contains(t, d.GetPlayableIndices(0), hint.CardIndices[0])
	}
}

func TestDehlaPakad_RejectsBadInput(t *testing.T) {
	d := newDehlaPakadForTest(t)
	assert.ErrorIs(t, d.PlayerPlay(0), ErrWrongPhase, "宣言フェーズでは出せない")
	require.NoError(t, d.SelectTrump(CardDesignSpade))
	assert.ErrorIs(t, d.SelectTrump(CardDesignHeart), ErrWrongPhase)
	for d.GetPhase() == DehlaPakadPhasePlay && !d.IsHumanTurn() {
		d.CpuPlay()
	}
	if d.GetPhase() != DehlaPakadPhasePlay {
		t.Skip("配りによっては人間の手番の前にトリックが揃う")
	}
	assert.Error(t, d.PlayerPlay(99))
	assert.Error(t, d.PlayerPlay(-1))
}

// **保存した盤で指し続けられる。**
func TestDehlaPakad_SaveRestoreKeepsPlaying(t *testing.T) {
	d := newDehlaPakadForTest(t)
	require.NoError(t, d.SelectTrump(CardDesignSpade))
	for d.GetPhase() == DehlaPakadPhasePlay && !d.IsHumanTurn() {
		d.CpuPlay()
	}
	data, err := json.Marshal(d)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored := new(DehlaPakad)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, d.GetPhase(), restored.GetPhase())
	assert.Equal(t, d.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, d.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, d.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, d.GetPrevTrickWinner(), restored.GetPrevTrickWinner(), "2 連勝の記憶が消えている")
	assert.Len(t, restored.GetCentrePile(), len(d.GetCentrePile()))
	for i := range d.GetPlayers() {
		assert.Equal(t, d.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	// 復元した盤で最後まで打てる。
	for hand := 0; hand < 200 && !restored.GetGameEndFlag(); hand++ {
		dehlaPakadPlayHand(t, restored)
		restored.NextHand()
	}
	assert.True(t, restored.GetGameEndFlag())
}

func TestDehlaPakad_RejectsTamperedSnapshot(t *testing.T) {
	restored := new(DehlaPakad)
	assert.Error(t, restored.UnmarshalJSON([]byte("{")))
	assert.Error(t, restored.UnmarshalJSON([]byte(`{"pl":[]}`)))
}

func TestDehlaPakadConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultDehlaPakadConfig().Validate())
	assert.Error(t, DehlaPakadConfig{CpuDifficulty: -1, TargetKots: 2}.Validate())
	assert.Error(t, DehlaPakadConfig{CpuDifficulty: 1, TargetKots: 0}.Validate())
	assert.Error(t, DehlaPakadConfig{CpuDifficulty: 1, TargetKots: DehlaPakadMaxKots + 1}.Validate())
}

func TestDehlaPakadSuitName(t *testing.T) {
	seen := map[string]bool{}
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		name := DehlaPakadSuitName(suit)
		assert.NotEqual(t, "unknown", name)
		assert.False(t, seen[name], "%q が重複している", name)
		seen[name] = true
	}
	assert.Equal(t, "unknown", DehlaPakadSuitName(99))
}

// dehlaPakadTestPlayers は席 0 が人間の 4 人を返す。
func dehlaPakadTestPlayers() []*DehlaPakadPlayer {
	players := make([]*DehlaPakadPlayer, DehlaPakadPlayerCnt)
	players[0] = NewDehlaPakadPlayer(true)
	for i := 1; i < DehlaPakadPlayerCnt; i++ {
		players[i] = NewDehlaPakadPlayer(false)
	}
	return players
}

// dehlaPakadTrickWonBySeat は winner が取るトリックを組む。
//
// leadSuit の札を 4 枚出し、winner だけがエースを持つ。tenSeat に席を渡すと
// その席だけが 10 を出す (-1 で 10 なし)。**10 の枚数は数え違えると全部の
// 検算が狂う**ので、明示的に 1 枚だけ置けるようにしてある。
func dehlaPakadTrickWonBySeat(winner, leadSuit, otherValue, tenSeat int) []*TrickCard {
	trick := make([]*TrickCard, 0, DehlaPakadPlayerCnt)
	for seat := 0; seat < DehlaPakadPlayerCnt; seat++ {
		value := otherValue
		switch seat {
		case winner:
			value = 1 // エース
		case tenSeat:
			value = DehlaPakadTenValue
		}
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: NewCard(leadSuit, value, false)})
	}
	return trick
}
