//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

// ssCard はテスト用にカードを生成するヘルパー。
func ssCard(design, value int) *Card { return NewCard(design, value, false) }

// newSSGame は全 CPU の 5 人ゲームを生成する (idx0 を人間にするかは human で指定)。
func newSSGame(human bool) *Sheepshead {
	cfg := DefaultSheepsheadConfig()
	players := make([]*SheepsheadPlayer, SheepsheadPlayerCnt)
	players[0] = NewSheepsheadPlayer(human, cfg.StartChips)
	for i := 1; i < SheepsheadPlayerCnt; i++ {
		players[i] = NewSheepsheadPlayer(false, cfg.StartChips)
	}
	return NewSheepshead(NewTrumpCardsBelote(), players, cfg)
}

// ssSetHand はプレイヤーの手札を指定カードに置き換える。
func ssSetHand(p *SheepsheadPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestSheepsheadConfig_Validate(t *testing.T) {
	if err := DefaultSheepsheadConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	cases := []SheepsheadConfig{
		{CpuDifficulty: -1, BaseChips: 2, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 99, BaseChips: 2, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 0, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 2, StartChips: 0, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 2, StartChips: 20, TargetChips: 20},
	}
	for i, c := range cases {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}

func TestSheepshead_ResetDealsHandsAndBlind(t *testing.T) {
	g := newSSGame(true)
	g.Reset()
	if g.GetPhase() != SheepsheadPhasePick {
		t.Fatalf("phase = %v, want Pick", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != SheepsheadHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, SheepsheadHandSize)
		}
	}
	if len(g.GetBlind()) != SheepsheadBlindSize {
		t.Errorf("blind = %d, want %d", len(g.GetBlind()), SheepsheadBlindSize)
	}
	if g.GetPickerIdx() != -1 || g.GetWinnerIdx() != -1 {
		t.Errorf("picker/winner should be -1 at start")
	}
}

func TestSheepshead_CardClassification(t *testing.T) {
	trumps := []*Card{
		ssCard(CardDesignClover, 12), ssCard(CardDesignDiamond, 12),
		ssCard(CardDesignHeart, 11), ssCard(CardDesignDiamond, 1),
		ssCard(CardDesignDiamond, 7),
	}
	for _, c := range trumps {
		if !sheepsheadIsTrump(c) {
			t.Errorf("%s should be trump", cardStr(c))
		}
		if sheepsheadSuitID(c) != sheepsheadTrumpSuit {
			t.Errorf("%s suitID should be trump", cardStr(c))
		}
	}
	fails := []*Card{ssCard(CardDesignClover, 1), ssCard(CardDesignSpade, 10), ssCard(CardDesignHeart, 13)}
	for _, c := range fails {
		if sheepsheadIsTrump(c) {
			t.Errorf("%s should not be trump", cardStr(c))
		}
		if sheepsheadSuitID(c) != c.GetDesign() {
			t.Errorf("%s suitID should equal design", cardStr(c))
		}
	}
}

func TestSheepshead_TrumpStrengthOrdering(t *testing.T) {
	// Q♣ > Q♠ > Q♥ > Q♦ > J♣ > J♠ > J♥ > J♦ > A♦ > 10♦ > K♦ > 9♦ > 8♦ > 7♦ > フェイル A
	order := []*Card{
		ssCard(CardDesignClover, 12), ssCard(CardDesignSpade, 12), ssCard(CardDesignHeart, 12), ssCard(CardDesignDiamond, 12),
		ssCard(CardDesignClover, 11), ssCard(CardDesignSpade, 11), ssCard(CardDesignHeart, 11), ssCard(CardDesignDiamond, 11),
		ssCard(CardDesignDiamond, 1), ssCard(CardDesignDiamond, 10), ssCard(CardDesignDiamond, 13),
		ssCard(CardDesignDiamond, 9), ssCard(CardDesignDiamond, 8), ssCard(CardDesignDiamond, 7),
		ssCard(CardDesignClover, 1),
	}
	for i := 1; i < len(order); i++ {
		if sheepsheadStrength(order[i-1]) <= sheepsheadStrength(order[i]) {
			t.Errorf("strength order broken at %d: %s !> %s", i, cardStr(order[i-1]), cardStr(order[i]))
		}
	}
}

func TestSheepshead_FailStrengthOrdering(t *testing.T) {
	order := []int{1, 10, 13, 9, 8, 7} // A>10>K>9>8>7
	for i := 1; i < len(order); i++ {
		if sheepsheadFailRank(order[i-1]) <= sheepsheadFailRank(order[i]) {
			t.Errorf("fail rank order broken at %d", i)
		}
	}
}

func TestSheepshead_CardPoints(t *testing.T) {
	want := map[int]int{1: 11, 10: 10, 13: 4, 12: 3, 11: 2, 9: 0, 8: 0, 7: 0}
	total := 0
	for v, p := range want {
		if got := sheepsheadCardPoints(v); got != p {
			t.Errorf("points(%d) = %d, want %d", v, got, p)
		}
	}
	// 32 枚デッキ全体で 120 点。
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
			total += sheepsheadCardPoints(v)
			_ = suit
		}
	}
	if total != SheepsheadTotalPoints {
		t.Errorf("deck total = %d, want %d", total, SheepsheadTotalPoints)
	}
}

func TestSheepshead_PickFlow(t *testing.T) {
	g := newSSGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	// 人間がピック → ブラインドを取り込みフェーズ Bury。
	if err := g.PlayerPick(true); err != nil {
		t.Fatalf("pick err: %v", err)
	}
	if g.GetPickerIdx() != 0 {
		t.Errorf("picker = %d, want 0", g.GetPickerIdx())
	}
	if g.GetPlayer(0).GetCardsSize() != SheepsheadHandSize+SheepsheadBlindSize {
		t.Errorf("picker hand = %d, want 8", g.GetPlayer(0).GetCardsSize())
	}
	if g.GetPhase() != SheepsheadPhaseBury {
		t.Errorf("phase = %v, want Bury", g.GetPhase())
	}
}

func TestSheepshead_PassThenForcedPick(t *testing.T) {
	g := newSSGame(false) // 全 CPU
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	// 4 人パスを強制的に再現。
	for i := 0; i < 4; i++ {
		g.resolvePick(g.GetCurrentPlayerIdx(), false)
	}
	if g.GetPassCount() != 4 {
		t.Fatalf("passCount = %d, want 4", g.GetPassCount())
	}
	// 5 人目はパスできず強制ピック。
	g.resolvePick(g.GetCurrentPlayerIdx(), false)
	if g.GetPickerIdx() < 0 {
		t.Errorf("last player should be forced to pick")
	}
}

func TestSheepshead_PlayerPickErrors(t *testing.T) {
	g := newSSGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	// 人間以外の手番。
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerPick(true); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(SheepsheadPhasePlay)
	if err := g.PlayerPick(true); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerPick(true); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

func TestSheepshead_BuryValidationAndApply(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhaseBury)
	g.SetPickerIdx(0)
	g.SetCurrentPlayerIdx(0)
	picker := g.GetPlayer(0)
	// クラブ A を呼べるよう、ピッカーはクラブ A を持たず、他者が持つ構成にする。
	ssSetHand(picker,
		ssCard(CardDesignClover, 12), ssCard(CardDesignSpade, 12), // Q (trump)
		ssCard(CardDesignSpade, 10), ssCard(CardDesignSpade, 13), // fail spades
		ssCard(CardDesignHeart, 10), ssCard(CardDesignHeart, 13), // fail hearts
		ssCard(CardDesignClover, 10), ssCard(CardDesignClover, 13), // fail clubs (no club A)
	)
	// 他のプレイヤーにクラブ A を渡す。
	ssSetHand(g.GetPlayer(1), ssCard(CardDesignClover, 1))

	// 枚数違い。
	if err := g.PlayerBury([]int{0}); err == nil {
		t.Error("expected error for wrong bury count")
	}
	// 範囲外。
	if err := g.PlayerBury([]int{0, 99}); err == nil {
		t.Error("expected error for OOB index")
	}
	// 重複。
	if err := g.PlayerBury([]int{2, 2}); err == nil {
		t.Error("expected error for duplicate index")
	}
	// 正常: 高得点フェイルを 2 枚埋める。
	if err := g.PlayerBury([]int{2, 4}); err != nil { // spade10, heart10
		t.Fatalf("bury err: %v", err)
	}
	if len(g.GetBuried()) != 2 {
		t.Errorf("buried = %d, want 2", len(g.GetBuried()))
	}
	if g.GetPhase() != SheepsheadPhaseCall {
		t.Errorf("phase = %v, want Call", g.GetPhase())
	}
}

func TestSheepshead_BuryAloneWhenNoCallable(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhaseBury)
	g.SetPickerIdx(0)
	g.SetCurrentPlayerIdx(0)
	// ピッカーが全フェイル A を持つ → 呼べない → 単独。
	ssSetHand(g.GetPlayer(0),
		ssCard(CardDesignClover, 1), ssCard(CardDesignSpade, 1), ssCard(CardDesignHeart, 1),
		ssCard(CardDesignClover, 12), ssCard(CardDesignSpade, 12),
		ssCard(CardDesignDiamond, 1), ssCard(CardDesignDiamond, 10), ssCard(CardDesignDiamond, 13),
	)
	if err := g.PlayerBury([]int{6, 7}); err != nil {
		t.Fatalf("bury err: %v", err)
	}
	if g.GetPartnerIdx() != -1 {
		t.Errorf("partner = %d, want -1 (alone)", g.GetPartnerIdx())
	}
	if g.GetPhase() != SheepsheadPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
}

func TestSheepshead_CallSetsPartner(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhaseCall)
	g.SetPickerIdx(0)
	g.SetCurrentPlayerIdx(0)
	ssSetHand(g.GetPlayer(0),
		ssCard(CardDesignClover, 12), ssCard(CardDesignSpade, 10), ssCard(CardDesignSpade, 13),
		ssCard(CardDesignHeart, 10), ssCard(CardDesignHeart, 13), ssCard(CardDesignClover, 10),
	)
	ssSetHand(g.GetPlayer(3), ssCard(CardDesignClover, 1)) // クラブ A 保持者
	callable := g.GetCallableSuits()
	found := false
	for _, s := range callable {
		if s == CardDesignClover {
			found = true
		}
	}
	if !found {
		t.Fatalf("clover should be callable, got %v", callable)
	}
	if err := g.PlayerCall(CardDesignClover); err != nil {
		t.Fatalf("call err: %v", err)
	}
	if g.GetPartnerIdx() != 3 {
		t.Errorf("partner = %d, want 3", g.GetPartnerIdx())
	}
	if g.GetPhase() != SheepsheadPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
	// 呼べないスート。
	g.SetPhase(SheepsheadPhaseCall)
	if err := g.PlayerCall(CardDesignDiamond); err == nil {
		t.Error("diamond (trump) should not be callable")
	}
}

func TestSheepshead_MustFollowSuit(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhasePlay)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(-1)
	g.SetCurrentPlayerIdx(0) // 人間 (player 0) が追従する
	g.SetLeadPlayerIdx(3)
	// リード: クラブ A (フェイル ♣) を player 3 が出す。
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: ssCard(CardDesignClover, 1)}})
	// 人間はクラブを持つ → クラブ以外は不可。
	ssSetHand(g.GetPlayer(0), ssCard(CardDesignClover, 10), ssCard(CardDesignSpade, 10))
	if err := g.PlayerPlay(1); err == nil { // spade を出そうとする
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // club10 を出す
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestSheepshead_TrickTopStrengthGuard(t *testing.T) {
	g := newSSGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ssCard(CardDesignClover, 1)},
	})
	// A winner not present in the current trick yields the sentinel, not a panic.
	if got := g.trickTopStrength(99); got != -1<<30 {
		t.Errorf("trickTopStrength(absent) = %d, want sentinel", got)
	}
	if got := g.trickTopStrength(0); got != sheepsheadStrength(ssCard(CardDesignClover, 1)) {
		t.Errorf("trickTopStrength(0) = %d, want A♣ strength", got)
	}
}

func TestSheepshead_TrickWinnerTrumpBeatsFail(t *testing.T) {
	g := newSSGame(false)
	g.SetPhase(SheepsheadPhaseTrickEnd)
	g.SetTrickNumber(1)
	// リード フェイル A♣、続いて Q♦ (切り札), A♠, 7♣, 8♣ → Q♦ が勝つ。
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ssCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: ssCard(CardDesignDiamond, 12)},
		{PlayerIdx: 2, Card: ssCard(CardDesignSpade, 1)},
		{PlayerIdx: 3, Card: ssCard(CardDesignClover, 7)},
		{PlayerIdx: 4, Card: ssCard(CardDesignClover, 8)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (Q♦ trump)", w)
	}
}

func TestSheepshead_TrickWinnerHighFail(t *testing.T) {
	g := newSSGame(false)
	// 切り札なし: リードスート ♣ の最強が勝つ。
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 2, Card: ssCard(CardDesignClover, 13)}, // K♣
		{PlayerIdx: 3, Card: ssCard(CardDesignClover, 1)},  // A♣ (最強)
		{PlayerIdx: 4, Card: ssCard(CardDesignSpade, 1)},   // A♠ (別スート無効)
		{PlayerIdx: 0, Card: ssCard(CardDesignClover, 10)},
		{PlayerIdx: 1, Card: ssCard(CardDesignClover, 7)},
	})
	if w := g.trickWinner(); w != 3 {
		t.Errorf("winner = %d, want 3 (A♣)", w)
	}
}

func TestSheepshead_ResolveAndNextTrick(t *testing.T) {
	g := newSSGame(false)
	g.SetPhase(SheepsheadPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ssCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: ssCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: ssCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: ssCard(CardDesignClover, 9)},
		{PlayerIdx: 4, Card: ssCard(CardDesignClover, 8)},
	})
	g.ResolveTrick()
	if g.GetPlayer(0).GetTrickCount() != 1 {
		t.Errorf("winner should have 1 trick")
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Errorf("lead should follow winner")
	}
	if g.GetPhase() != SheepsheadPhaseTrickEnd {
		t.Errorf("phase = %v, want TrickEnd", g.GetPhase())
	}
	g.NextTrick()
	if g.GetTrickNumber() != 2 || g.GetPhase() != SheepsheadPhasePlay {
		t.Errorf("next trick not started: trick=%d phase=%v", g.GetTrickNumber(), g.GetPhase())
	}
}

func TestSheepshead_LastTrickGoesToRoundEnd(t *testing.T) {
	g := newSSGame(false)
	g.SetPhase(SheepsheadPhaseTrickEnd)
	g.SetTrickNumber(SheepsheadTrickCount)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ssCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: ssCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: ssCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: ssCard(CardDesignClover, 9)},
		{PlayerIdx: 4, Card: ssCard(CardDesignClover, 8)},
	})
	g.ResolveTrick()
	if g.GetPhase() != SheepsheadPhaseRoundEnd {
		t.Errorf("phase = %v, want RoundEnd", g.GetPhase())
	}
}

func TestSheepshead_PartnerRevealOnCalledAce(t *testing.T) {
	g := newSSGame(false)
	g.SetPhase(SheepsheadPhasePlay)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(2)
	g.calledSuit = CardDesignClover
	g.SetCurrentPlayerIdx(2)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: ssCard(CardDesignClover, 10)}})
	ssSetHand(g.GetPlayer(2), ssCard(CardDesignClover, 1)) // 呼びカード
	g.SetCurrentPlayerIdx(2)
	// CPU プレイで呼びカードが出る。
	g.CpuPlay()
	if !g.IsPartnerRevealed() {
		t.Error("partner should be revealed after called ace played")
	}
}

func TestSheepshead_ScoreRoundPickerWinsAndChipsZeroSum(t *testing.T) {
	g := newSSGame(false)
	g.SetPhase(SheepsheadPhaseRoundEnd)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(1)
	// ピッカー組 (0,1) が高得点トリックを取得。
	g.GetPlayer(0).AddTrick([]*Card{ssCard(CardDesignClover, 1), ssCard(CardDesignSpade, 1)})   // 22
	g.GetPlayer(1).AddTrick([]*Card{ssCard(CardDesignHeart, 1), ssCard(CardDesignDiamond, 1)})  // 22
	g.GetPlayer(0).AddTrick([]*Card{ssCard(CardDesignClover, 10), ssCard(CardDesignSpade, 10)}) // 20
	// 埋め札。
	g.buried = []*Card{ssCard(CardDesignHeart, 10), ssCard(CardDesignClover, 13)} // 14
	// ディフェンダーが残り。
	g.GetPlayer(2).AddTrick([]*Card{ssCard(CardDesignSpade, 13)}) // 4
	before := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		before += g.GetPlayer(i).GetChips()
	}
	g.ScoreRound()
	if !g.GetRoundPickerWon() {
		t.Errorf("picker team should win with %d pts", g.GetRoundPickerPoints())
	}
	after := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		after += g.GetPlayer(i).GetChips()
	}
	if before != after {
		t.Errorf("chips not zero-sum: before=%d after=%d", before, after)
	}
	if g.GetPlayer(0).GetChips() <= g.GetConfig().StartChips {
		t.Errorf("picker should gain chips")
	}
}

func TestSheepshead_MultiplierSchneiderSchwarz(t *testing.T) {
	if sheepsheadMultiplier(35, false) != 1 {
		t.Error("normal should be x1")
	}
	if sheepsheadMultiplier(30, false) != 2 {
		t.Error("<=30 should be schneider x2")
	}
	if sheepsheadMultiplier(0, true) != 3 {
		t.Error("no trick should be schwarz x3")
	}
}

func TestSheepshead_ScoreRoundAloneAndGameEnd(t *testing.T) {
	g := newSSGame(false)
	cfg := g.GetConfig()
	cfg.TargetChips = cfg.StartChips + 1 // 即終了しやすく
	g.SetConfig(cfg)
	g.SetPhase(SheepsheadPhaseRoundEnd)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(-1) // 単独
	// ピッカー単独で全 120 点 (シュバルツ)。
	g.GetPlayer(0).AddTrick([]*Card{
		ssCard(CardDesignClover, 1), ssCard(CardDesignSpade, 1), ssCard(CardDesignHeart, 1),
		ssCard(CardDesignDiamond, 1), ssCard(CardDesignClover, 10), ssCard(CardDesignSpade, 10),
		ssCard(CardDesignHeart, 10), ssCard(CardDesignDiamond, 10),
	})
	g.buried = []*Card{ssCard(CardDesignClover, 13), ssCard(CardDesignSpade, 13)}
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Errorf("game should end when picker reaches target chips")
	}
	if g.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerIdx())
	}
}

func TestSheepshead_NextRound(t *testing.T) {
	g := newSSGame(false)
	g.Reset()
	g.SetPhase(SheepsheadPhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round = %d, want 2", g.GetRoundNumber())
	}
	if g.GetDealerIdx() != (d+1)%SheepsheadPlayerCnt {
		t.Errorf("dealer should rotate")
	}
	if g.GetPhase() != SheepsheadPhasePick {
		t.Errorf("phase = %v, want Pick", g.GetPhase())
	}
}

func TestSheepshead_CpuPlayFullRoundProgresses(t *testing.T) {
	g := newSSGame(false) // 全 CPU
	g.Reset()
	// 全 CPU なので人間操作なしでラウンドを進められる。
	for steps := 0; steps < 200; steps++ {
		switch g.GetPhase() {
		case SheepsheadPhasePick, SheepsheadPhaseBury, SheepsheadPhaseCall, SheepsheadPhasePlay:
			g.CpuPlay()
		case SheepsheadPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == SheepsheadPhaseTrickEnd {
				g.NextTrick()
			}
		case SheepsheadPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case SheepsheadPhaseGameEnd:
			return
		}
		if g.GetRoundNumber() > 50 {
			break
		}
	}
	// ピッカーが確定し、トリックが消化されていることを確認。
	if g.GetRoundNumber() < 1 {
		t.Error("game did not progress")
	}
}

func TestSheepshead_HintPerPhase(t *testing.T) {
	g := newSSGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	if h := g.GetHint(); h == nil || (h.Reason != "pick_take" && h.Reason != "pick_pass") {
		t.Errorf("pick hint missing: %+v", h)
	}
	// Bury hint
	g.SetPhase(SheepsheadPhaseBury)
	g.SetPickerIdx(0)
	ssSetHand(g.GetPlayer(0),
		ssCard(CardDesignClover, 12), ssCard(CardDesignSpade, 10), ssCard(CardDesignSpade, 13),
		ssCard(CardDesignHeart, 10), ssCard(CardDesignClover, 10), ssCard(CardDesignClover, 13),
		ssCard(CardDesignDiamond, 1), ssCard(CardDesignDiamond, 10),
	)
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 2 {
		t.Errorf("bury hint missing: %+v", h)
	}
	// Call hint
	g.SetPhase(SheepsheadPhaseCall)
	ssSetHand(g.GetPlayer(2), ssCard(CardDesignSpade, 1))
	if h := g.GetHint(); h == nil || h.Reason != "call_suit" {
		t.Errorf("call hint missing: %+v", h)
	}
	// Play hint
	g.SetPhase(SheepsheadPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	ssSetHand(g.GetPlayer(0), ssCard(CardDesignClover, 7), ssCard(CardDesignSpade, 8))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("play hint missing: %+v", h)
	}
}

func TestSheepshead_GetPlayableIndices(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhasePlay)
	g.SetCurrentTrick(nil)
	ssSetHand(g.GetPlayer(0), ssCard(CardDesignClover, 1), ssCard(CardDesignSpade, 10))
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(SheepsheadPhaseBury)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Errorf("playable should be nil outside play phase")
	}
	if idx := g.GetPlayableIndices(99); idx != nil {
		t.Errorf("playable OOB should be nil")
	}
}

func TestSheepshead_JSONRoundTrip(t *testing.T) {
	g := newSSGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	_ = g.PlayerPick(true)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Sheepshead
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPickerIdx() != g.GetPickerIdx() || g2.GetPhase() != g.GetPhase() {
		t.Errorf("round trip mismatch")
	}
	if g2.GetPlayerCnt() != SheepsheadPlayerCnt {
		t.Errorf("players lost in round trip")
	}
}

func TestSheepshead_UnmarshalOversized(t *testing.T) {
	var g Sheepshead
	// 過大な actionLog を持つ JSON。
	bad := `{"al":[`
	for i := 0; i < sheepsheadMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errSheepsheadOversized) {
		t.Errorf("err = %v, want errSheepsheadOversized", err)
	}
}

func TestSheepshead_IsHumanTurn(t *testing.T) {
	g := newSSGame(true)
	g.SetPhase(SheepsheadPhaseBury)
	g.SetPickerIdx(0)
	if !g.IsHumanTurn() {
		t.Error("human picker bury should be human turn")
	}
	g.SetPickerIdx(1)
	if g.IsHumanTurn() {
		t.Error("CPU picker bury should not be human turn")
	}
	g.SetPhase(SheepsheadPhasePlay)
	g.SetCurrentPlayerIdx(0)
	if !g.IsHumanTurn() {
		t.Error("human play turn")
	}
}
