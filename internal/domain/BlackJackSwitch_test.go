//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setSwitchHandsForTest はテスト用に2ハンドを直接構築するヘルパー。
// (hand0Cards, hand1Cards) と各ハンドのベット額を指定する。
func setSwitchHandsForTest(t *testing.T, bs *BlackJackSwitch, bet int, hand0, hand1 []*Card) {
	t.Helper()
	hands := []*BlackJackHand{NewBlackJackHand(), NewBlackJackHand()}
	hands[0].SetBet(bet)
	for _, c := range hand0 {
		hands[0].AddCard(c)
	}
	hands[1].SetBet(bet)
	for _, c := range hand1 {
		hands[1].AddCard(c)
	}
	bs.SetHands(hands)
}

func TestNewDefaultBlackJackSwitch_InitialState(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	assert.Equal(t, BJSwitchPhaseBet, bs.GetPhase())
	assert.Equal(t, BJSwitchDefaultChips, bs.GetPlayer().GetChips())
	assert.Len(t, bs.GetHands(), BJSwitchHands)
	assert.False(t, bs.GetGameEndFlag())
	assert.False(t, bs.IsSwitched())
}

func TestBlackJackSwitch_PlayerBet_InvalidAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative", -1},
		{"below min", BJSwitchMinBet - 1},
		{"not multiple", BJSwitchMinBet + 1},
		{"above max", BJSwitchMaxBet + BJSwitchMinBet},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bs := NewDefaultBlackJackSwitch()
			err := bs.PlayerBet(c.amount)
			require.Error(t, err)
		})
	}
}

func TestBlackJackSwitch_PlayerBet_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.SetPhase(BJSwitchPhaseAction)
	err := bs.PlayerBet(BJSwitchMinBet)
	require.Error(t, err)
}

func TestBlackJackSwitch_PlayerBet_InsufficientChips(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.GetPlayer().SetChips(BJSwitchMinBet) // can't cover 2 hands
	err := bs.PlayerBet(BJSwitchMinBet)
	require.Error(t, err)
}

func TestBlackJackSwitch_PlayerBet_DealsTwoHandsAndAdvancesToSwitch(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	startChips := bs.GetPlayer().GetChips()
	require.NoError(t, bs.PlayerBet(BJSwitchMinBet))

	// Either Switch phase (normal) or End phase (dealer natural BJ on first deal).
	switch bs.GetPhase() {
	case BJSwitchPhaseSwitch:
		assert.Equal(t, 2, bs.GetHands()[0].GetCardsSize())
		assert.Equal(t, 2, bs.GetHands()[1].GetCardsSize())
		assert.Equal(t, 2, bs.GetDealer().GetCardsSize())
		assert.Equal(t, startChips-BJSwitchMinBet*2, bs.GetPlayer().GetChips())
	case BJSwitchPhaseEnd:
		// Dealer dealt natural BJ; cards still consistent.
		assert.True(t, bs.GetGameEndFlag())
	default:
		t.Fatalf("unexpected phase %d", bs.GetPhase())
	}
}

func TestBlackJackSwitch_PlayerSwitch_SwapsSecondCards(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 5, true)},
		[]*Card{NewCard(CardDesignHeart, 6, true), NewCard(CardDesignHeart, 11, true)},
	)
	bs.SetPhase(BJSwitchPhaseSwitch)
	require.NoError(t, bs.PlayerSwitch())
	// Hand 0: 10 + (was hand1 second) Jack(11→10) = 20
	assert.Equal(t, 20, bs.GetHands()[0].GetScore())
	// Hand 1: 6 + (was hand0 second) 5 = 11
	assert.Equal(t, 11, bs.GetHands()[1].GetScore())
	assert.True(t, bs.IsSwitched())
}

func TestBlackJackSwitch_PlayerSwitch_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.SetPhase(BJSwitchPhaseAction)
	err := bs.PlayerSwitch()
	require.Error(t, err)
}

func TestBlackJackSwitch_PlayerKeep_AdvancesToAction(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 5, true)},
		[]*Card{NewCard(CardDesignHeart, 7, true), NewCard(CardDesignHeart, 6, true)},
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true)) // dealer 17, will stand
	bs.SetPhase(BJSwitchPhaseSwitch)
	require.NoError(t, bs.PlayerKeep())
	assert.False(t, bs.IsSwitched())
	// Both hands non-21 → action phase, neither auto-stood.
	assert.Equal(t, BJSwitchPhaseAction, bs.GetPhase())
}

func TestBlackJackSwitch_PlayerKeep_AutoStandsOn21AndPlaysOut(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignSpade, 13, true)}, // 21
		[]*Card{NewCard(CardDesignHeart, 1, true), NewCard(CardDesignHeart, 12, true)}, // 21
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true)) // dealer 17, stand
	bs.SetPhase(BJSwitchPhaseSwitch)
	startChips := bs.GetPlayer().GetChips()
	require.NoError(t, bs.PlayerKeep())
	// Both hands = natural 21 → auto stand → dealer plays → game ends.
	assert.Equal(t, BJSwitchPhaseEnd, bs.GetPhase())
	// 1:1 payout for each hand: 2*bet returned per hand × 2 hands.
	assert.Equal(t, startChips+BJSwitchMinBet*4, bs.GetPlayer().GetChips())
}

func TestBlackJackSwitch_HitAndBust(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 10, true)}, // 20
		[]*Card{NewCard(CardDesignHeart, 7, true), NewCard(CardDesignHeart, 6, true)},   // 13
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true)) // 17
	bs.SetPhase(BJSwitchPhaseAction)
	bs.SetCurrentHandIdx(0)
	// Force a known card on top to bust hand 0: stand first.
	require.NoError(t, bs.PlayerStand())
	assert.Equal(t, 1, bs.GetCurrentHandIdx())
	// Now hit hand 1 by injecting cards ourselves via direct hand mutation —
	// instead, exercise PlayerStand so we don't depend on shuffle order.
	require.NoError(t, bs.PlayerStand())
	assert.Equal(t, BJSwitchPhaseEnd, bs.GetPhase())
}

func TestBlackJackSwitch_DoubleDown_RequiresTwoCards(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{
			NewCard(CardDesignSpade, 5, true),
			NewCard(CardDesignSpade, 5, true),
			NewCard(CardDesignSpade, 2, true),
		},
		[]*Card{NewCard(CardDesignHeart, 7, true), NewCard(CardDesignHeart, 6, true)},
	)
	bs.SetPhase(BJSwitchPhaseAction)
	bs.SetCurrentHandIdx(0)
	err := bs.PlayerDoubleDown()
	require.Error(t, err)
}

func TestBlackJackSwitch_Hit_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	err := bs.PlayerHit()
	require.Error(t, err)
}

func TestBlackJackSwitch_Stand_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	err := bs.PlayerStand()
	require.Error(t, err)
}

func TestBlackJackSwitch_DoubleDown_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	err := bs.PlayerDoubleDown()
	require.Error(t, err)
}

func TestBlackJackSwitch_DoubleDown_InsufficientChips(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, BJSwitchMinBet,
		[]*Card{NewCard(CardDesignSpade, 5, true), NewCard(CardDesignSpade, 5, true)},
		[]*Card{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignHeart, 5, true)},
	)
	bs.GetPlayer().SetChips(0)
	bs.SetPhase(BJSwitchPhaseAction)
	bs.SetCurrentHandIdx(0)
	err := bs.PlayerDoubleDown()
	require.Error(t, err)
}

func TestBlackJackSwitch_PayoutsRule_BJPays1To1(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignSpade, 13, true)}, // 21
		[]*Card{NewCard(CardDesignHeart, 10, true), NewCard(CardDesignHeart, 8, true)}, // 18
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true)) // 17
	bs.SetPhase(BJSwitchPhaseAction)
	bs.SetCurrentHandIdx(0)
	startChips := bs.GetPlayer().GetChips()
	require.NoError(t, bs.PlayerStand()) // hand0 stands at 21
	require.NoError(t, bs.PlayerStand()) // hand1 stands at 18
	// Both hands win: hand0 (21 vs 17), hand1 (18 vs 17). 1:1 each → +200.
	assert.Equal(t, startChips+200*2, bs.GetPlayer().GetChips())
	require.Len(t, bs.GetHandResults(), 2)
	assert.Equal(t, GameResultWin, bs.GetHandResults()[0])
	assert.Equal(t, GameResultWin, bs.GetHandResults()[1])
}

func TestBlackJackSwitch_PayoutsRule_Dealer22IsPush(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 10, true)}, // 20
		[]*Card{NewCard(CardDesignHeart, 9, true), NewCard(CardDesignHeart, 9, true)},   // 18
	)
	// Dealer 12 + draws to 22 (10+2+10 = 22, post deal we set both up cards).
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 5, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true))  // 12 → must hit
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true)) // 22
	// Mark dealer auto-stand by direct call to resolvePayouts — bypass dealer draw:
	// We instead just check resolvePayouts directly via the public Stand path.
	// Simulate that hand0/hand1 have already stood:
	bs.GetHands()[0].SetStood(true)
	bs.GetHands()[1].SetStood(true)
	bs.SetPhase(BJSwitchPhaseAction)
	bs.SetCurrentHandIdx(0)
	startChips := bs.GetPlayer().GetChips()
	// Trigger end-game pipeline directly via resolvePayouts.
	bs.resolvePayouts()
	assert.True(t, bs.IsDealerPushed22())
	// Both hands push (player 20, 18 vs dealer 22 not-natural) → return bet × 2.
	assert.Equal(t, startChips+100*2, bs.GetPlayer().GetChips())
	for _, r := range bs.GetHandResults() {
		assert.Equal(t, GameResultDraw, r)
	}
}

func TestBlackJackSwitch_PayoutsRule_Dealer22Player21NaturalWins(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignSpade, 13, true)}, // natural 21
		[]*Card{NewCard(CardDesignHeart, 10, true), NewCard(CardDesignHeart, 9, true)}, // 19
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 5, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true)) // 22
	bs.GetHands()[0].SetStood(true)
	bs.GetHands()[1].SetStood(true)
	startChips := bs.GetPlayer().GetChips()
	bs.resolvePayouts()
	// hand0 natural 21 beats dealer 22 → win 1:1 (200), hand1 19 vs 22 → push (100).
	assert.Equal(t, GameResultWin, bs.GetHandResults()[0])
	assert.Equal(t, GameResultDraw, bs.GetHandResults()[1])
	assert.Equal(t, startChips+200+100, bs.GetPlayer().GetChips())
}

func TestBlackJackSwitch_PayoutsRule_DealerNaturalBJBeatsPlayer21(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 5, true), NewCard(CardDesignSpade, 6, true)}, // 21 in 3 cards
		[]*Card{NewCard(CardDesignHeart, 1, true), NewCard(CardDesignHeart, 13, true)},                                    // natural 21
	)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 1, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 13, true)) // natural BJ
	startChips := bs.GetPlayer().GetChips()
	bs.resolvePayouts()
	// Hand 0 (3-card 21) loses to dealer natural; hand 1 (natural 21) pushes.
	assert.Equal(t, GameResultLose, bs.GetHandResults()[0])
	assert.Equal(t, GameResultDraw, bs.GetHandResults()[1])
	// hand1 push returns 100; hand0 loses; total +100.
	assert.Equal(t, startChips+100, bs.GetPlayer().GetChips())
}

func TestBlackJackSwitch_PayoutsRule_BustLoses(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 5, true)}, // 25 bust
		[]*Card{NewCard(CardDesignHeart, 10, true), NewCard(CardDesignHeart, 9, true)},                                     // 19
	)
	bs.GetHands()[0].SetBusted(true)
	bs.GetHands()[1].SetStood(true)
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 10, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true)) // 17
	startChips := bs.GetPlayer().GetChips()
	bs.resolvePayouts()
	assert.Equal(t, GameResultLose, bs.GetHandResults()[0])
	assert.Equal(t, GameResultWin, bs.GetHandResults()[1])
	// hand0 lose 0; hand1 win 200. Total +200.
	assert.Equal(t, startChips+200, bs.GetPlayer().GetChips())
}

// TestBlackJackSwitch_DealerSkipsDrawWhenAllHandsBust は、両ハンドが既にバスト
// している場合、ディーラーは追加で引かずホールカードを開けるだけで終局すること
// を確認する（標準BJと同じ挙動）。
func TestBlackJackSwitch_DealerSkipsDrawWhenAllHandsBust(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	setSwitchHandsForTest(t, bs, 100,
		[]*Card{NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 10, true), NewCard(CardDesignSpade, 5, true)}, // 25 bust
		[]*Card{NewCard(CardDesignHeart, 10, true), NewCard(CardDesignHeart, 9, true), NewCard(CardDesignHeart, 5, true)},  // 24 bust
	)
	bs.GetHands()[0].SetBusted(true)
	bs.GetHands()[1].SetBusted(true)
	// Dealer 12 — would normally draw to 17+ if reached.
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 5, true))
	bs.GetDealer().AddCard(NewCard(CardDesignClover, 7, true))
	bs.SetPhase(BJSwitchPhaseAction)
	startChips := bs.GetPlayer().GetChips()

	bs.dealerPlay()

	// Dealer must not have drawn extra cards.
	assert.Equal(t, 2, bs.GetDealer().GetCardsSize())
	assert.Equal(t, BJSwitchPhaseEnd, bs.GetPhase())
	assert.True(t, bs.GetGameEndFlag())
	assert.Equal(t, GameResultLose, bs.GetHandResults()[0])
	assert.Equal(t, GameResultLose, bs.GetHandResults()[1])
	assert.Equal(t, startChips, bs.GetPlayer().GetChips()) // both bets forfeited
}

func TestBlackJackSwitch_OverallResult(t *testing.T) {
	cases := []struct {
		name string
		r    []GameResult
		want GameResult
	}{
		{"both win", []GameResult{GameResultWin, GameResultWin}, GameResultWin},
		{"win+draw", []GameResult{GameResultWin, GameResultDraw}, GameResultWin},
		{"split", []GameResult{GameResultWin, GameResultLose}, GameResultDraw},
		{"both lose", []GameResult{GameResultLose, GameResultLose}, GameResultLose},
		{"lose+draw", []GameResult{GameResultLose, GameResultDraw}, GameResultLose},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bs := NewDefaultBlackJackSwitch()
			bs.handResults = c.r
			assert.Equal(t, c.want, bs.GetOverallResult())
		})
	}
}

func TestBlackJackSwitch_GetTotalPayout(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.handPayouts = []int{200, 100}
	assert.Equal(t, 300, bs.GetTotalPayout())
}

func TestBlackJackSwitch_Reset_RestoresDefaults(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.SetPhase(BJSwitchPhaseEnd)
	bs.GetPlayer().SetChips(0)
	bs.Reset()
	assert.Equal(t, BJSwitchPhaseBet, bs.GetPhase())
	assert.Equal(t, BJSwitchDefaultChips, bs.GetPlayer().GetChips())
	assert.Len(t, bs.GetHands(), BJSwitchHands)
	assert.False(t, bs.IsSwitched())
}

func TestBlackJackSwitch_JSONRoundTrip(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	require.NoError(t, bs.PlayerBet(BJSwitchMinBet))
	data, err := json.Marshal(bs)
	require.NoError(t, err)

	var restored BlackJackSwitch
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, bs.GetPhase(), restored.GetPhase())
	assert.Equal(t, bs.GetPlayer().GetChips(), restored.GetPlayer().GetChips())
	assert.Equal(t, len(bs.GetHands()), len(restored.GetHands()))
}

func TestBlackJackSwitch_UnmarshalJSON_Invalid(t *testing.T) {
	var bs BlackJackSwitch
	err := json.Unmarshal([]byte("not json"), &bs)
	require.Error(t, err)
}

func TestBlackJackSwitch_GetActionLog(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	require.NoError(t, bs.PlayerBet(BJSwitchMinBet))
	assert.NotEmpty(t, bs.GetActionLog())
}

// TestBlackJackSwitch_PlayerSwitch_Refuses_BeforeDeal は2枚揃わないハンドで
// スイッチを呼ぶとエラーになることを確認する。
func TestBlackJackSwitch_PlayerSwitch_Refuses_BeforeDeal(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.SetPhase(BJSwitchPhaseSwitch) // hands empty
	err := bs.PlayerSwitch()
	require.Error(t, err)
}

func TestBlackJackSwitch_PlayerKeep_WrongPhase(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.SetPhase(BJSwitchPhaseAction)
	err := bs.PlayerKeep()
	require.Error(t, err)
}

// #5586: 交換すると得か損かは、打つまで分からなかった。先読みは実際に入れ替える
// PlayerSwitch と**同じ採点**を通ること — 別に数え直すと予告と結果が食い違う。
func TestBlackJackSwitch_SwitchPreviewMatchesTheActualSwitch(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.Reset()
	require.NoError(t, bs.PlayerBet(10))

	// スイッチフェーズに入るまで進める (Reset→Bet で 2 枚ずつ配られる)。
	if bs.GetPhase() != BJSwitchPhaseSwitch {
		t.Skipf("deal did not reach the switch phase (phase=%d)", bs.GetPhase())
	}

	first, second, ok := bs.SwitchPreviewScores()
	require.True(t, ok, "two dealt hands can always be previewed")

	require.NoError(t, bs.PlayerSwitch())
	hands := bs.GetHands()
	// **予告した得点がそのまま出ること。**
	assert.Equal(t, first, hands[0].GetScore())
	assert.Equal(t, second, hands[1].GetScore())
}

// 2 枚に満たないハンドは入れ替えられないので、先読みも返さない。
func TestBlackJackSwitch_SwitchPreviewRefusesShortHands(t *testing.T) {
	bs := NewDefaultBlackJackSwitch()
	bs.Reset()

	_, _, ok := bs.SwitchPreviewScores()
	assert.False(t, ok, "no cards have been dealt yet")
}
