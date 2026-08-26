//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSutdaForTest(t *testing.T) *Sutda {
	t.Helper()
	s := NewDefaultSutda()
	s.Reset()
	return s
}

// sutdaCard は月と複製番号から札を作る。複製 1 が光札 (1・3・8 月のみ)。
func sutdaCard(month, copyIdx int) *Card { return NewCard(month, copyIdx, false) }

// sutdaPlain は光でないほうの札を作る。
func sutdaPlain(month int) *Card { return sutdaCard(month, 2) }

// **20 枚 = 1〜10 月 × 2 枚。** 1 枚でも欠けると役の確率が変わる。
func TestSutda_DeckIsTwentyCards(t *testing.T) {
	deck := buildSutdaDeck()
	assert.Len(t, deck, 20)
	seen := map[[2]int]bool{}
	months := map[int]int{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "重複した札 %v", key)
		seen[key] = true
		months[c.GetDesign()]++
	}
	assert.Len(t, months, SutdaMonthCnt)
	for m, n := range months {
		assert.Equal(t, SutdaCopiesPerMonth, n, "%d 月の枚数", m)
	}
}

// **光札は 1・3・8 月だけ。** 花札の光は 1・3・8・11・12 月で、11 と 12 は
// 20 枚デッキに入らない ── だから光ッタンは 3 通りしか存在しない。
func TestSutda_OnlyThreeMonthsHaveAGwang(t *testing.T) {
	gwang := 0
	for _, c := range buildSutdaDeck() {
		if SutdaIsGwang(c) {
			gwang++
			assert.Contains(t, []int{1, 3, 8}, c.GetDesign(), "光でない月に光がある")
		}
	}
	assert.Equal(t, 3, gwang)
	// 同じ月のもう 1 枚は光ではない。
	assert.False(t, SutdaIsGwang(sutdaPlain(1)))
	assert.False(t, SutdaIsGwang(sutdaPlain(3)))
	assert.False(t, SutdaIsGwang(sutdaPlain(8)))
	// 2 月には光札そのものが無い。
	assert.False(t, SutdaIsGwang(sutdaCard(2, 1)))
	assert.False(t, SutdaIsGwang(nil))
}

// **光ッタンは 38 > 18 > 13 の 3 通り。**
func TestSutda_GwangTtaengOrder(t *testing.T) {
	g38 := SutdaEvaluate(sutdaCard(3, 1), sutdaCard(8, 1))
	g18 := SutdaEvaluate(sutdaCard(1, 1), sutdaCard(8, 1))
	g13 := SutdaEvaluate(sutdaCard(1, 1), sutdaCard(3, 1))
	assert.Equal(t, "gwang38", g38.Name)
	assert.Equal(t, "gwang18", g18.Name)
	assert.Equal(t, "gwang13", g13.Name)
	assert.Greater(t, g38.Rank, g18.Rank)
	assert.Greater(t, g18.Rank, g13.Rank)

	// **光でないほうの札では成立しない。** 3+8 でも光でなければただの 1 끗。
	plain := SutdaEvaluate(sutdaPlain(3), sutdaPlain(8))
	assert.Equal(t, "kkeut1", plain.Name, "光でない 3+8 が光ッタンになっている")
	assert.Less(t, plain.Rank, g13.Rank)
}

// **땡 は同じ月の 2 枚。** 장땡 (10) が最強で 1땡 が最弱。
func TestSutda_TtaengOrder(t *testing.T) {
	prev := 0
	for m := 1; m <= SutdaMonthCnt; m++ {
		h := SutdaEvaluate(sutdaCard(m, 1), sutdaCard(m, 2))
		assert.Greater(t, h.Rank, prev, "%d땡 が前より弱い", m)
		prev = h.Rank
	}
	assert.Equal(t, "jangttaeng", SutdaEvaluate(sutdaCard(10, 1), sutdaCard(10, 2)).Name)
	assert.Equal(t, "ttaeng1", SutdaEvaluate(sutdaCard(1, 1), sutdaCard(1, 2)).Name)

	// 1땡 でも光ッタンの下、特殊役の上。
	ttaeng1 := SutdaEvaluate(sutdaCard(1, 1), sutdaCard(1, 2))
	assert.Less(t, ttaeng1.Rank, SutdaEvaluate(sutdaCard(1, 1), sutdaCard(3, 1)).Rank)
	assert.Greater(t, ttaeng1.Rank, SutdaEvaluate(sutdaCard(1, 1), sutdaPlain(2)).Rank)
}

// **特殊役は 알리 > 독사 > 구삥 > 장삥 > 장사 > 세륙。**
func TestSutda_SpecialOrder(t *testing.T) {
	order := []struct {
		a, b int
		name string
	}{
		{1, 2, "ali"},
		{1, 4, "doksa"},
		{1, 9, "gupping"},
		{1, 10, "jangpping"},
		{4, 10, "jangsa"},
		{4, 6, "seryuk"},
	}
	prev := 1 << 30
	for _, tt := range order {
		h := SutdaEvaluate(sutdaPlain(tt.a), sutdaPlain(tt.b))
		assert.Equal(t, tt.name, h.Name, "%d+%d", tt.a, tt.b)
		assert.Less(t, h.Rank, prev, "%s が前より強い", tt.name)
		prev = h.Rank
	}
	// 並べる順を逆にしても同じ役。
	assert.Equal(t, "ali", SutdaEvaluate(sutdaPlain(2), sutdaPlain(1)).Name)
	// **一番弱い特殊役でも、一番強い끗 より上。**
	seryuk := SutdaEvaluate(sutdaPlain(4), sutdaPlain(6))
	gabo := SutdaEvaluate(sutdaPlain(4), sutdaPlain(5)) // 9끗 = 갑오
	assert.Equal(t, "gabo", gabo.Name)
	assert.Greater(t, seryuk.Rank, gabo.Rank, "세륙 が 갑오 に負けている")
}

// **끗 は合計の下 1 桁。** 0 は망통 で最弱。
func TestSutda_KkeutIsTheLastDigitOfTheSum(t *testing.T) {
	for _, tt := range []struct {
		a, b, kkeut int
		name        string
	}{
		{2, 3, 5, "kkeut5"},
		{7, 8, 5, "kkeut5"},
		{9, 10, 9, "gabo"},
		{2, 8, 0, "mangtong"},
		{3, 7, 0, "mangtong"},
		{5, 6, 1, "kkeut1"},
	} {
		h := SutdaEvaluate(sutdaPlain(tt.a), sutdaPlain(tt.b))
		assert.Equal(t, tt.kkeut, h.Kkeut, "%d+%d の끗", tt.a, tt.b)
		assert.Equal(t, tt.name, h.Name)
	}
	// 망통 がいちばん弱い。
	mangtong := SutdaEvaluate(sutdaPlain(2), sutdaPlain(8))
	assert.Less(t, mangtong.Rank, SutdaEvaluate(sutdaPlain(5), sutdaPlain(6)).Rank)
}

func TestSutdaEvaluate_NilIsNoHand(t *testing.T) {
	assert.Equal(t, "none", SutdaEvaluate(nil, sutdaPlain(1)).Name)
	assert.Equal(t, "none", SutdaEvaluate(sutdaPlain(1), nil).Name)
}

// 役の識別名は重複しない (i18n キーに使うため)。
func TestSutda_HandNamesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	deck := buildSutdaDeck()
	for i := 0; i < len(deck); i++ {
		for j := i + 1; j < len(deck); j++ {
			h := SutdaEvaluate(deck[i], deck[j])
			assert.NotEqual(t, "none", h.Name, "評価できない組み合わせ")
			seen[h.Name] = true
		}
	}
	// 3 光ッタン + 10 땡 + 6 特殊 + 10 끗 = 29 種類。
	assert.Len(t, seen, 29)
}

// **配りは 2 枚ずつ。** 参加料は全員から取る。
func TestSutda_DealsTwoEachAndTakesTheAnte(t *testing.T) {
	s := newSutdaForTest(t)
	assert.Equal(t, SutdaPhaseBet, s.GetPhase())
	for i, p := range s.GetPlayers() {
		assert.Equal(t, SutdaHandSize, p.GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, SutdaAnte, p.GetBet(), "席 %d の参加料", i)
		assert.Equal(t, SutdaDefaultChips-SutdaAnte, p.GetChips(), "席 %d のチップ", i)
	}
	assert.Equal(t, SutdaAnte*s.GetPlayerCnt(), s.GetPot())
	assert.Equal(t, SutdaAnte, s.GetCurrentBet())
}

// **開幕は人間の手番。** 親の左隣から始まるので、親を最後の席にしてある ──
// ここを 0 にすると、人間は一度もベットを選べないまま卓が回る。
func TestSutda_HumanActsFirst(t *testing.T) {
	s := newSutdaForTest(t)
	assert.Equal(t, len(s.GetPlayers())-1, s.GetDealerIdx())
	assert.Equal(t, 0, s.GetCurrentPlayerIdx())
	assert.True(t, s.IsHumanTurn())
}

// **レイズは他の席の「行動済み」を取り消す。** 取り消さないと、上げられた側が
// コールもフォールドもしないままハンドが終わる。
func TestSutda_RaiseReopensTheBetting(t *testing.T) {
	s := newSutdaForTest(t)
	before := s.GetCurrentBet()
	require.NoError(t, s.PlayerAction(SutdaActionRaise))
	assert.Equal(t, before+SutdaBetUnit, s.GetCurrentBet())
	assert.Equal(t, 1, s.GetRaiseCount())
	// 席 0 が上げた直後に showdown へ飛んでいないこと。
	assert.Equal(t, SutdaPhaseBet, s.GetPhase(), "上げた本人だけで手が終わっている")
	assert.NotEqual(t, 0, s.GetCurrentPlayerIdx())
}

// **レイズには上限がある。** 無いと CPU 同士が際限なく上げ続ける。
func TestSutda_RaisesAreCapped(t *testing.T) {
	s := newSutdaForTest(t)
	for i := 0; i < SutdaMaxRaises; i++ {
		s.raiseCount = i
		assert.True(t, s.CanRaise(0), "%d 回目のレイズができない", i+1)
	}
	s.raiseCount = SutdaMaxRaises
	assert.False(t, s.CanRaise(0), "上限を超えて上げられる")
	assert.Error(t, s.PlayerAction(SutdaActionRaise))
}

// チップが足りなければ上げられない。
func TestSutda_CannotRaiseWithoutTheChips(t *testing.T) {
	s := newSutdaForTest(t)
	s.GetPlayer(0).SetChips(0)
	assert.False(t, s.CanRaise(0))
}

func TestSutda_RejectsUnknownActionAndOffTurn(t *testing.T) {
	s := newSutdaForTest(t)
	assert.Error(t, s.PlayerAction("zzz"))
	// 人間以外の手番では弾く。
	s.currentPlayer = 1
	assert.ErrorIs(t, s.PlayerAction(SutdaActionCall), ErrNotHumanTurn)
}

// **降りたら以降の手番が回ってこない。**
func TestSutda_FoldingLeavesTheHand(t *testing.T) {
	s := newSutdaForTest(t)
	require.NoError(t, s.PlayerAction(SutdaActionFold))
	assert.True(t, s.GetPlayer(0).IsFolded())
	sutdaRunToShowdown(t, s)
	require.NotNil(t, s.GetLastResult())
	assert.NotContains(t, s.GetLastResult().Winners, 0, "降りた席が勝っている")
}

// **同点はポットを分ける。** 席順で決めると、同じ役なのに座席で負ける。
func TestSutda_TiedHandsSplitThePot(t *testing.T) {
	s := newSutdaForTest(t)
	// 席 0 と 席 1 に同じ役 (5끗) を持たせ、他は降ろす。
	sutdaSetHand(s, 0, sutdaPlain(2), sutdaPlain(3))
	sutdaSetHand(s, 1, sutdaPlain(7), sutdaPlain(8))
	for i := 2; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetFolded(true)
	}
	before0 := s.GetPlayer(0).GetChips()
	before1 := s.GetPlayer(1).GetChips()
	pot := s.GetPot()
	s.showdown()

	res := s.GetLastResult()
	require.NotNil(t, res)
	assert.ElementsMatch(t, []int{0, 1}, res.Winners)
	assert.Equal(t, pot, res.Pot)
	assert.Equal(t, before0+pot/2, s.GetPlayer(0).GetChips())
	assert.Equal(t, before1+pot/2, s.GetPlayer(1).GetChips())
}

// 端数が出る割り方でも、配ったチップの総額はポットと一致する。
func TestSutda_SplitPotLosesNoChips(t *testing.T) {
	s := newSutdaForTest(t)
	sutdaSetHand(s, 0, sutdaPlain(2), sutdaPlain(3))
	sutdaSetHand(s, 1, sutdaPlain(7), sutdaPlain(8))
	for i := 2; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetFolded(true)
	}
	s.pot = 7 // 2 人で割り切れない
	before := s.GetPlayer(0).GetChips() + s.GetPlayer(1).GetChips()
	s.showdown()
	assert.Equal(t, before+7, s.GetPlayer(0).GetChips()+s.GetPlayer(1).GetChips(), "チップが消えている")
}

// **強い役が勝つ。** 光ッタンは땡 にも끗 にも負けない。
func TestSutda_StrongestHandTakesThePot(t *testing.T) {
	s := newSutdaForTest(t)
	sutdaSetHand(s, 0, sutdaPlain(2), sutdaPlain(3))     // 5끗
	sutdaSetHand(s, 1, sutdaCard(3, 1), sutdaCard(8, 1)) // 38광땡
	for i := 2; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetFolded(true)
	}
	s.showdown()
	res := s.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, []int{1}, res.Winners)
	assert.Equal(t, "gwang38", res.Hands[1].Name)
}

// **1 マッチを通しで打てる。** 誰かのチップが尽きるまで回る。
func TestSutda_PlaysAMatchThrough(t *testing.T) {
	s := newSutdaForTest(t)
	for hand := 0; hand < 3000 && !s.GetGameEndFlag(); hand++ {
		sutdaRunToShowdown(t, s)
		s.NextHand()
	}
	require.True(t, s.GetGameEndFlag(), "マッチが終わらない")
	assert.Equal(t, SutdaPhaseGameEnd, s.GetPhase())
	assert.GreaterOrEqual(t, s.GetWinnerIdx(), 0)
}

// **チップの総額はハンドを跨いで変わらない。** 湧いたり消えたりしない。
func TestSutda_ChipsAreConserved(t *testing.T) {
	s := newSutdaForTest(t)
	total := func() int {
		n := s.GetPot()
		for _, p := range s.GetPlayers() {
			n += p.GetChips()
		}
		return n
	}
	want := total()
	for hand := 0; hand < 40 && !s.GetGameEndFlag(); hand++ {
		sutdaRunToShowdown(t, s)
		assert.Equal(t, want, total(), "ハンド %d でチップが合わない", hand+1)
		s.NextHand()
	}
}

func TestSutda_HintFollowsThePhase(t *testing.T) {
	s := newSutdaForTest(t)
	hint := s.GetHint()
	require.NotNil(t, hint)
	assert.Contains(t, []string{SutdaActionCall, SutdaActionRaise, SutdaActionFold}, hint.Action)

	sutdaRunToShowdown(t, s)
	if !s.GetGameEndFlag() {
		assert.Equal(t, "next_hand", s.GetHint().Reason)
	}
}

// **助言は CPU の難易度に引きずられない。** Easy の方策はランダムに降りるので、
// そのまま使うと Easy を選んだ人にだけ雑な助言が出る。
func TestSutda_HintIgnoresCpuDifficulty(t *testing.T) {
	cfg := DefaultSutdaConfig()
	cfg.CpuDifficulty = SutdaCpuDifficultyEasy
	s := NewSutdaWithConfig(cfg)
	s.Reset()
	require.True(t, s.IsHumanTurn())
	// **コールに追加が要る局面に固定する。** 差額 0 だとどの方策もコールを
	// 返すので、ぶれないことを確かめたことにならない。
	s.currentBet = SutdaAnte + SutdaBetUnit
	require.Positive(t, s.GetCallAmount(0))

	want := s.GetHint().Action
	for i := 0; i < 20; i++ {
		assert.Equal(t, want, s.GetHint().Action, "%d 回目でぶれた", i+1)
	}
}

// **保存した盤で指し続けられる。**
func TestSutda_SaveRestoreKeepsPlaying(t *testing.T) {
	s := newSutdaForTest(t)
	require.NoError(t, s.PlayerAction(SutdaActionRaise))

	data, err := json.Marshal(s)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored := new(Sutda)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetPot(), restored.GetPot())
	assert.Equal(t, s.GetCurrentBet(), restored.GetCurrentBet())
	assert.Equal(t, s.GetRaiseCount(), restored.GetRaiseCount(), "レイズ回数が消えている")
	assert.Equal(t, s.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	for i := range s.GetPlayers() {
		assert.Equal(t, s.GetPlayer(i).GetChips(), restored.GetPlayer(i).GetChips(), "席 %d のチップ", i)
		assert.Equal(t, s.GetPlayer(i).GetBet(), restored.GetPlayer(i).GetBet(), "席 %d のベット", i)
	}
	// 復元した盤で最後まで打てる。
	for hand := 0; hand < 3000 && !restored.GetGameEndFlag(); hand++ {
		sutdaRunToShowdown(t, restored)
		restored.NextHand()
	}
	assert.True(t, restored.GetGameEndFlag())
}

func TestSutda_RejectsTamperedSnapshot(t *testing.T) {
	restored := new(Sutda)
	assert.Error(t, restored.UnmarshalJSON([]byte("{")))
	assert.Error(t, restored.UnmarshalJSON([]byte(`{"pl":[]}`)))
}

// **席数を変えたら卓も変わる。** 設定だけ変えて席が据え置きだと食い違う。
func TestSutda_ResetRebuildsTheTableForTheSeatCount(t *testing.T) {
	s := newSutdaForTest(t)
	require.Equal(t, SutdaDefaultSeats, s.GetPlayerCnt())
	cfg := s.GetConfig()
	cfg.Seats = 5
	s.SetConfig(cfg)
	s.Reset()
	assert.Equal(t, 5, s.GetPlayerCnt())
	assert.True(t, s.GetPlayer(0).GetIsHuman())
	for i := 1; i < 5; i++ {
		assert.False(t, s.GetPlayer(i).GetIsHuman(), "席 %d が人間になっている", i)
	}
}

func TestSutdaConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultSutdaConfig().Validate())
	assert.Error(t, SutdaConfig{CpuDifficulty: -1, Seats: 3, StartChips: 1000}.Validate())
	assert.Error(t, SutdaConfig{CpuDifficulty: 1, Seats: 1, StartChips: 1000}.Validate())
	assert.Error(t, SutdaConfig{CpuDifficulty: 1, Seats: 99, StartChips: 1000}.Validate())
	assert.Error(t, SutdaConfig{CpuDifficulty: 1, Seats: 3, StartChips: 1}.Validate())
}

// sutdaSetHand は席に指定の 2 枚を持たせる。
func sutdaSetHand(s *Sutda, playerIdx int, a, b *Card) {
	p := s.GetPlayer(playerIdx)
	p.Reset()
	p.AddCard(a)
	p.AddCard(b)
	p.SetFolded(false)
}

// sutdaRunToShowdown は現在のハンドをショーダウンまで進める。
func sutdaRunToShowdown(t *testing.T, s *Sutda) {
	t.Helper()
	for step := 0; step < 500 && s.GetPhase() == SutdaPhaseBet; step++ {
		if s.IsHumanTurn() {
			require.NoError(t, s.PlayerAction(SutdaActionCall))
			continue
		}
		s.CpuAction()
	}
	require.NotEqual(t, SutdaPhaseBet, s.GetPhase(), "ベッティングが終わらない")
}
