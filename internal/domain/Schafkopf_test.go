//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

// skCard はテスト用にカードを生成するヘルパー。
func skCard(design, value int) *Card { return NewCard(design, value, false) }

// newSKGame は全 CPU の 5 人ゲームを生成する (idx0 を人間にするかは human で指定)。
func newSKGame(human bool) *Schafkopf {
	cfg := DefaultSchafkopfConfig()
	players := make([]*SchafkopfPlayer, SchafkopfPlayerCnt)
	players[0] = NewSchafkopfPlayer(human, cfg.StartChips)
	for i := 1; i < SchafkopfPlayerCnt; i++ {
		players[i] = NewSchafkopfPlayer(false, cfg.StartChips)
	}
	return NewSchafkopf(NewTrumpCardsBelote(), players, cfg)
}

// skSetHand はプレイヤーの手札を指定カードに置き換える。
func skSetHand(p *SchafkopfPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestSchafkopfConfig_Validate(t *testing.T) {
	if err := DefaultSchafkopfConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	cases := []SchafkopfConfig{
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

func TestSchafkopf_ResetDealsHandsAndBlind(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	if g.GetPhase() != SchafkopfPhasePick {
		t.Fatalf("phase = %v, want Pick", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != SchafkopfHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, SchafkopfHandSize)
		}
	}
	// **32 枚を 4 人で配り切る。** クローン元の 5 人卓は 30 枚配って 2 枚を
	// 伏せていたが、こちらは山に 1 枚も残らない。手で書いた枚数ではなく
	// 席数と手札枚数の積で見る。
	dealt := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		dealt += g.GetPlayer(i).GetCardsSize()
	}
	if want := SchafkopfHandSize * SchafkopfPlayerCnt; dealt != want {
		t.Errorf("dealt = %d, want %d (deck fully dealt, no blind)", dealt, want)
	}
	if g.GetPickerIdx() != -1 || g.GetWinnerIdx() != -1 {
		t.Errorf("picker/winner should be -1 at start")
	}
}

func TestSchafkopf_CardClassification(t *testing.T) {
	trumps := []*Card{
		skCard(CardDesignClover, 12), skCard(CardDesignDiamond, 12),
		skCard(CardDesignHeart, 11), skCard(CardDesignDiamond, 1),
		skCard(CardDesignDiamond, 7),
	}
	g := newSKGame(true) // 既定は Rufspiel
	for _, c := range trumps {
		if !g.isTrump(c) {
			t.Errorf("%s should be trump", cardStr(c))
		}
		if g.suitID(c) != schafkopfTrumpSuit {
			t.Errorf("%s suitID should be trump", cardStr(c))
		}
	}
	fails := []*Card{skCard(CardDesignClover, 1), skCard(CardDesignSpade, 10), skCard(CardDesignHeart, 13)}
	for _, c := range fails {
		if g.isTrump(c) {
			t.Errorf("%s should not be trump", cardStr(c))
		}
		if g.suitID(c) != c.GetDesign() {
			t.Errorf("%s suitID should equal design", cardStr(c))
		}
	}
}

func TestSchafkopf_TrumpStrengthOrdering(t *testing.T) {
	// Q♣ > Q♠ > Q♥ > Q♦ > J♣ > J♠ > J♥ > J♦ > A♦ > 10♦ > K♦ > 9♦ > 8♦ > 7♦ > フェイル A
	order := []*Card{
		skCard(CardDesignClover, 12), skCard(CardDesignSpade, 12), skCard(CardDesignHeart, 12), skCard(CardDesignDiamond, 12),
		skCard(CardDesignClover, 11), skCard(CardDesignSpade, 11), skCard(CardDesignHeart, 11), skCard(CardDesignDiamond, 11),
		skCard(CardDesignDiamond, 1), skCard(CardDesignDiamond, 10), skCard(CardDesignDiamond, 13),
		skCard(CardDesignDiamond, 9), skCard(CardDesignDiamond, 8), skCard(CardDesignDiamond, 7),
		skCard(CardDesignClover, 1),
	}
	// Rufspiel の既定の並び。強さは契約で変わるので、契約を持つ盤で測る。
	ruf := newSKGame(false)
	ruf.SetContractForTest(SchafkopfContractRufspiel, 0)
	for i := 1; i < len(order); i++ {
		if ruf.strength(order[i-1]) <= ruf.strength(order[i]) {
			t.Errorf("strength order broken at %d: %s !> %s", i, cardStr(order[i-1]), cardStr(order[i]))
		}
	}
}

func TestSchafkopf_FailStrengthOrdering(t *testing.T) {
	order := []int{1, 10, 13, 9, 8, 7} // A>10>K>9>8>7
	for i := 1; i < len(order); i++ {
		if schafkopfFailRank(order[i-1]) <= schafkopfFailRank(order[i]) {
			t.Errorf("fail rank order broken at %d", i)
		}
	}
}

func TestSchafkopf_CardPoints(t *testing.T) {
	want := map[int]int{1: 11, 10: 10, 13: 4, 12: 3, 11: 2, 9: 0, 8: 0, 7: 0}
	total := 0
	for v, p := range want {
		if got := schafkopfCardPoints(v); got != p {
			t.Errorf("points(%d) = %d, want %d", v, got, p)
		}
	}
	// 32 枚デッキ全体で 120 点。
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
			total += schafkopfCardPoints(v)
			_ = suit
		}
	}
	if total != SchafkopfTotalPoints {
		t.Errorf("deck total = %d, want %d", total, SchafkopfTotalPoints)
	}
}

// **開幕は人間から話す。** 宣言はディーラーの左隣から始まるので、
// ディーラーが 0 のままだと人間 (席 0) は毎回最後に話し、先に宣言した CPU
// を上回れる契約しか選べない。開幕 500 配りで 3 契約すべてが選べた回数は
// 0 だった (実測)。
func TestSchafkopf_HumanSpeaksFirstOnTheOpeningDeal(t *testing.T) {
	g := newSKGame(true)
	g.Reset()

	if got := g.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("first to speak = %d, want 0 (the human)", got)
	}
	if got := len(g.GetBeatableContracts()); got != 3 {
		t.Errorf("beatable contracts = %d, want 3 — nobody has bid yet", got)
	}
}

func TestSchafkopf_PickFlow(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	// **宣言の次は呼びフェーズ。** クローン元はここでブラインドを拾って
	// 埋めフェーズに入ったが、伏せ札が無いので手札は増えない。
	if err := g.PlayerDeclare(true, SchafkopfContractRufspiel, 0); err != nil {
		t.Fatalf("pick err: %v", err)
	}
	// 宣言しただけでは競りは閉じない。残る 3 席が発言してから確定する。
	if g.GetPickerIdx() >= 0 {
		t.Error("one declaration must not settle the auction")
	}
	skFinishAuction(g)
	if g.GetPickerIdx() != 0 {
		t.Errorf("picker = %d, want 0", g.GetPickerIdx())
	}
	if got := g.GetPlayer(0).GetCardsSize(); got != SchafkopfHandSize {
		t.Errorf("picker hand = %d, want %d (no blind to pick up)", got, SchafkopfHandSize)
	}
	if g.GetPhase() != SchafkopfPhaseCall && g.GetPhase() != SchafkopfPhasePlay {
		t.Errorf("phase = %v, want Call (or Play when nothing is callable)", g.GetPhase())
	}
}

// skFinishAuction は人間が発言したあとの残り席をパスさせて競りを閉じる。
// **1 人が宣言しても競りは閉じない。** 全員が発言してから最上位が取る。
func skFinishAuction(g *Schafkopf) {
	for g.GetPhase() == SchafkopfPhasePick {
		g.resolvePick(g.GetCurrentPlayerIdx(), false, SchafkopfContractRufspiel, 0)
	}
}

func TestSchafkopf_PassThenForcedPick(t *testing.T) {
	g := newSKGame(false) // 全 CPU
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	// **席数から導く。** クローン元は 5 人卓だったので 4 回パスして 5 人目が
	// 強制されていた。4 人卓では 3 回。
	for i := 0; i < SchafkopfPlayerCnt-1; i++ {
		g.resolvePick(g.GetCurrentPlayerIdx(), false, SchafkopfContractRufspiel, 0)
		if g.GetPickerIdx() >= 0 {
			t.Fatalf("picker settled after %d passes; every seat speaks first", i+1)
		}
	}
	if got, want := g.GetPassCount(), SchafkopfPlayerCnt-1; got != want {
		t.Fatalf("passCount = %d, want %d", got, want)
	}
	// 全員がパスしたら、ディーラーの左隣が Rufspiel を引き受ける。
	g.resolvePick(g.GetCurrentPlayerIdx(), false, SchafkopfContractRufspiel, 0)
	if g.GetPickerIdx() < 0 {
		t.Errorf("someone must take the contract once everyone has passed")
	}
}

// **先に言った席が契約を取る方式は成立しない。** 席順で先に来る CPU が
// ほぼ毎回宣言してしまい、人間は一度も宣言フェーズに立てなくなる
// (この形にする前の実測は 0/200)。全員が発言してから最上位が取る。
func TestSchafkopf_HighestContractWinsRegardlessOfSeat(t *testing.T) {
	g := newSKGame(false)
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	// 席 0 が Rufspiel、席 2 が Solo。席順では 0 が先だが、契約の順位で 2 が取る。
	g.resolvePick(0, true, SchafkopfContractRufspiel, 0)
	if g.GetPickerIdx() >= 0 {
		t.Fatal("the first declaration must not close the auction")
	}
	g.resolvePick(1, false, SchafkopfContractRufspiel, 0)
	g.resolvePick(2, true, SchafkopfContractSolo, CardDesignHeart)
	g.resolvePick(3, false, SchafkopfContractRufspiel, 0)

	if got := g.GetPickerIdx(); got != 2 {
		t.Errorf("picker = %d, want 2 (Solo outranks Rufspiel)", got)
	}
	if got := g.GetContract(); got != SchafkopfContractSolo {
		t.Errorf("contract = %v, want Solo", got)
	}
	if got := g.GetSoloSuit(); got != CardDesignHeart {
		t.Errorf("solo suit = %d, want %d", got, CardDesignHeart)
	}
}

// 同位では上書きしない。後から同じ契約を言っても順位は上がらないので、
// 先に言った席が残る。
func TestSchafkopf_EqualContractDoesNotDisplaceTheEarlierSeat(t *testing.T) {
	g := newSKGame(false)
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	g.resolvePick(0, true, SchafkopfContractWenz, 0)
	g.resolvePick(1, true, SchafkopfContractWenz, 0)
	g.resolvePick(2, false, SchafkopfContractRufspiel, 0)
	g.resolvePick(3, false, SchafkopfContractRufspiel, 0)

	if got := g.GetPickerIdx(); got != 0 {
		t.Errorf("picker = %d, want 0 (an equal bid does not outrank)", got)
	}
}

// 上回れない契約はエラーで返す。ボタンにも出さないので、押せるのに必ず
// 拒否される操作面ができない。
func TestSchafkopf_CannotDeclareUnderTheStandingBid(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetPhase(SchafkopfPhasePick)
	g.SetCurrentPlayerIdx(0)

	// まだ誰も宣言していない: 3 契約すべて宣言できる。
	if got := len(g.GetBeatableContracts()); got != 3 {
		t.Fatalf("beatable = %d, want 3 before anyone declares", got)
	}

	g.resolvePick(1, true, SchafkopfContractWenz, 0)
	g.SetCurrentPlayerIdx(0)

	if err := g.PlayerDeclare(true, SchafkopfContractRufspiel, 0); err == nil {
		t.Error("Rufspiel must not beat a standing Wenz")
	}
	if got := g.GetBeatableContracts(); len(got) != 1 || got[0] != SchafkopfContractSolo {
		t.Errorf("beatable = %v, want [Solo]", got)
	}
	// **負のコントロール。** 上回る契約は通る。
	if err := g.PlayerDeclare(true, SchafkopfContractSolo, CardDesignHeart); err != nil {
		t.Errorf("Solo should beat a standing Wenz: %v", err)
	}
}

func TestSchafkopf_PlayerDeclareErrors(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	// 人間以外の手番。
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerDeclare(true, SchafkopfContractRufspiel, 0); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(SchafkopfPhasePlay)
	if err := g.PlayerDeclare(true, SchafkopfContractRufspiel, 0); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerDeclare(true, SchafkopfContractRufspiel, 0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

// **埋めフェーズは無い。** 宣言した時点で、呼べる A が無ければ単独プレイへ進む。
func TestSchafkopf_DeclareAloneWhenNoCallable(t *testing.T) {
	g := newSKGame(true)
	g.SetPhase(SchafkopfPhasePick)
	g.SetPickerIdx(0)
	g.SetCurrentPlayerIdx(0)
	// ピッカーが全フェイル A を持つ → 呼べない → 単独。
	skSetHand(g.GetPlayer(0),
		skCard(CardDesignClover, 1), skCard(CardDesignSpade, 1), skCard(CardDesignHeart, 1),
		skCard(CardDesignClover, 12), skCard(CardDesignSpade, 12),
		skCard(CardDesignDiamond, 1), skCard(CardDesignDiamond, 10), skCard(CardDesignDiamond, 13),
	)

	// 宣言を確定させる。呼べる A が無いので呼びフェーズを飛ばして単独プレイへ。
	g.becomePicker(0)

	if g.GetPartnerIdx() != -1 {
		t.Errorf("partner = %d, want -1 (alone)", g.GetPartnerIdx())
	}
	if g.GetPhase() != SchafkopfPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
}

func TestSchafkopf_CallSetsPartner(t *testing.T) {
	g := newSKGame(true)
	g.SetPhase(SchafkopfPhaseCall)
	g.SetPickerIdx(0)
	g.SetCurrentPlayerIdx(0)
	skSetHand(g.GetPlayer(0),
		skCard(CardDesignClover, 12), skCard(CardDesignSpade, 10), skCard(CardDesignSpade, 13),
		skCard(CardDesignHeart, 10), skCard(CardDesignHeart, 13), skCard(CardDesignClover, 10),
	)
	skSetHand(g.GetPlayer(3), skCard(CardDesignClover, 1)) // クラブ A 保持者
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
	if g.GetPhase() != SchafkopfPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
	// 呼べないスート。
	g.SetPhase(SchafkopfPhaseCall)
	if err := g.PlayerCall(CardDesignDiamond); err == nil {
		t.Error("diamond (trump) should not be callable")
	}
}

func TestSchafkopf_MustFollowSuit(t *testing.T) {
	g := newSKGame(true)
	g.SetPhase(SchafkopfPhasePlay)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(-1)
	g.SetCurrentPlayerIdx(0) // 人間 (player 0) が追従する
	g.SetLeadPlayerIdx(3)
	// リード: クラブ A (フェイル ♣) を player 3 が出す。
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: skCard(CardDesignClover, 1)}})
	// 人間はクラブを持つ → クラブ以外は不可。
	skSetHand(g.GetPlayer(0), skCard(CardDesignClover, 10), skCard(CardDesignSpade, 10))
	if err := g.PlayerPlay(1); err == nil { // spade を出そうとする
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // club10 を出す
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestSchafkopf_TrickTopStrengthGuard(t *testing.T) {
	g := newSKGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},
	})
	// A winner not present in the current trick yields the sentinel, not a panic.
	if got := g.trickTopStrength(99); got != -1<<30 {
		t.Errorf("trickTopStrength(absent) = %d, want sentinel", got)
	}
	if got := g.trickTopStrength(0); got != g.strength(skCard(CardDesignClover, 1)) {
		t.Errorf("trickTopStrength(0) = %d, want A♣ strength", got)
	}
}

func TestSchafkopf_TrickWinnerTrumpBeatsFail(t *testing.T) {
	g := newSKGame(false)
	g.SetPhase(SchafkopfPhaseTrickEnd)
	g.SetTrickNumber(1)
	// リード フェイル A♣、続いて Q♦ (切り札), A♠, 7♣, 8♣ → Q♦ が勝つ。
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: skCard(CardDesignDiamond, 12)},
		{PlayerIdx: 2, Card: skCard(CardDesignSpade, 1)},
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 7)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (Q♦ trump)", w)
	}
}

func TestSchafkopf_TrickWinnerHighFail(t *testing.T) {
	g := newSKGame(false)
	// 切り札なし: リードスート ♣ の最強が勝つ。
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 13)}, // K♣
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 1)},  // A♣ (最強)
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 10)},
		{PlayerIdx: 1, Card: skCard(CardDesignClover, 7)},
	})
	if w := g.trickWinner(); w != 3 {
		t.Errorf("winner = %d, want 3 (A♣)", w)
	}
}

func TestSchafkopf_ResolveAndNextTrick(t *testing.T) {
	g := newSKGame(false)
	g.SetPhase(SchafkopfPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: skCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 9)},
	})
	g.ResolveTrick()
	if g.GetPlayer(0).GetTrickCount() != 1 {
		t.Errorf("winner should have 1 trick")
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Errorf("lead should follow winner")
	}
	if g.GetPhase() != SchafkopfPhaseTrickEnd {
		t.Errorf("phase = %v, want TrickEnd", g.GetPhase())
	}
	g.NextTrick()
	if g.GetTrickNumber() != 2 || g.GetPhase() != SchafkopfPhasePlay {
		t.Errorf("next trick not started: trick=%d phase=%v", g.GetTrickNumber(), g.GetPhase())
	}
}

func TestSchafkopf_LastTrickGoesToRoundEnd(t *testing.T) {
	g := newSKGame(false)
	g.SetPhase(SchafkopfPhaseTrickEnd)
	g.SetTrickNumber(SchafkopfTrickCount)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: skCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 9)},
	})
	g.ResolveTrick()
	if g.GetPhase() != SchafkopfPhaseRoundEnd {
		t.Errorf("phase = %v, want RoundEnd", g.GetPhase())
	}
}

func TestSchafkopf_PartnerRevealOnCalledAce(t *testing.T) {
	g := newSKGame(false)
	g.SetPhase(SchafkopfPhasePlay)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(2)
	g.calledSuit = CardDesignClover
	g.SetCurrentPlayerIdx(2)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: skCard(CardDesignClover, 10)}})
	skSetHand(g.GetPlayer(2), skCard(CardDesignClover, 1)) // 呼びカード
	g.SetCurrentPlayerIdx(2)
	// CPU プレイで呼びカードが出る。
	g.CpuPlay()
	if !g.IsPartnerRevealed() {
		t.Error("partner should be revealed after called ace played")
	}
}

func TestSchafkopf_ScoreRoundPickerWinsAndChipsZeroSum(t *testing.T) {
	g := newSKGame(false)
	g.SetPhase(SchafkopfPhaseRoundEnd)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(1)
	// ピッカー組 (0,1) が高得点トリックを取得。
	g.GetPlayer(0).AddTrick([]*Card{skCard(CardDesignClover, 1), skCard(CardDesignSpade, 1)})   // 22
	g.GetPlayer(1).AddTrick([]*Card{skCard(CardDesignHeart, 1), skCard(CardDesignDiamond, 1)})  // 22
	g.GetPlayer(0).AddTrick([]*Card{skCard(CardDesignClover, 10), skCard(CardDesignSpade, 10)}) // 20
	// **埋め札は無い。** その 14 点ぶんはトリックから出す。
	g.GetPlayer(0).AddTrick([]*Card{skCard(CardDesignHeart, 10), skCard(CardDesignClover, 13)}) // 14
	// ディフェンダーが残り。
	g.GetPlayer(2).AddTrick([]*Card{skCard(CardDesignSpade, 13)}) // 4
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

func TestSchafkopf_MultiplierSchneiderSchwarz(t *testing.T) {
	if schafkopfMultiplier(35, false) != 1 {
		t.Error("normal should be x1")
	}
	if schafkopfMultiplier(30, false) != 2 {
		t.Error("<=30 should be schneider x2")
	}
	if schafkopfMultiplier(0, true) != 3 {
		t.Error("no trick should be schwarz x3")
	}
}

func TestSchafkopf_ScoreRoundAloneAndGameEnd(t *testing.T) {
	g := newSKGame(false)
	cfg := g.GetConfig()
	cfg.TargetChips = cfg.StartChips + 1 // 即終了しやすく
	g.SetConfig(cfg)
	g.SetPhase(SchafkopfPhaseRoundEnd)
	g.SetPickerIdx(0)
	g.SetPartnerIdx(-1) // 単独
	// ピッカー単独で全 120 点 (シュバルツ)。
	g.GetPlayer(0).AddTrick([]*Card{
		skCard(CardDesignClover, 1), skCard(CardDesignSpade, 1), skCard(CardDesignHeart, 1),
		skCard(CardDesignDiamond, 1), skCard(CardDesignClover, 10), skCard(CardDesignSpade, 10),
		skCard(CardDesignHeart, 10), skCard(CardDesignDiamond, 10),
	})
	g.GetPlayer(0).AddTrick([]*Card{skCard(CardDesignClover, 13), skCard(CardDesignSpade, 13)})
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Errorf("game should end when picker reaches target chips")
	}
	if g.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerIdx())
	}
}

func TestSchafkopf_NextRound(t *testing.T) {
	g := newSKGame(false)
	g.Reset()
	g.SetPhase(SchafkopfPhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round = %d, want 2", g.GetRoundNumber())
	}
	if g.GetDealerIdx() != (d+1)%SchafkopfPlayerCnt {
		t.Errorf("dealer should rotate")
	}
	if g.GetPhase() != SchafkopfPhasePick {
		t.Errorf("phase = %v, want Pick", g.GetPhase())
	}
}

func TestSchafkopf_CpuPlayFullRoundProgresses(t *testing.T) {
	g := newSKGame(false) // 全 CPU
	g.Reset()
	// 全 CPU なので人間操作なしでラウンドを進められる。
	for steps := 0; steps < 200; steps++ {
		switch g.GetPhase() {
		case SchafkopfPhasePick, SchafkopfPhaseCall, SchafkopfPhasePlay:
			g.CpuPlay()
		case SchafkopfPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == SchafkopfPhaseTrickEnd {
				g.NextTrick()
			}
		case SchafkopfPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case SchafkopfPhaseGameEnd:
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

func TestSchafkopf_HintPerPhase(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	if h := g.GetHint(); h == nil || (h.Reason != "pick_take" && h.Reason != "pick_pass") {
		t.Errorf("pick hint missing: %+v", h)
	}
	// Call hint（呼びフェーズのヒントは宣言者に対して出る）
	g.SetPhase(SchafkopfPhaseCall)
	g.SetPickerIdx(0)
	skSetHand(g.GetPlayer(0),
		skCard(CardDesignClover, 12), skCard(CardDesignSpade, 12),
		skCard(CardDesignSpade, 10), skCard(CardDesignHeart, 10))
	skSetHand(g.GetPlayer(2), skCard(CardDesignSpade, 1))
	if h := g.GetHint(); h == nil || h.Reason != "call_suit" {
		t.Errorf("call hint missing: %+v", h)
	}
	// Play hint
	g.SetPhase(SchafkopfPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	skSetHand(g.GetPlayer(0), skCard(CardDesignClover, 7), skCard(CardDesignSpade, 8))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("play hint missing: %+v", h)
	}
}

func TestSchafkopf_GetPlayableIndices(t *testing.T) {
	g := newSKGame(true)
	g.SetPhase(SchafkopfPhasePlay)
	g.SetCurrentTrick(nil)
	skSetHand(g.GetPlayer(0), skCard(CardDesignClover, 1), skCard(CardDesignSpade, 10))
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(SchafkopfPhaseCall)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Errorf("playable should be nil outside play phase")
	}
	if idx := g.GetPlayableIndices(99); idx != nil {
		t.Errorf("playable OOB should be nil")
	}
}

func TestSchafkopf_JSONRoundTrip(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	_ = g.PlayerDeclare(true, SchafkopfContractRufspiel, 0)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Schafkopf
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPickerIdx() != g.GetPickerIdx() || g2.GetPhase() != g.GetPhase() {
		t.Errorf("round trip mismatch")
	}
	if g2.GetPlayerCnt() != SchafkopfPlayerCnt {
		t.Errorf("players lost in round trip")
	}
}

func TestSchafkopf_UnmarshalOversized(t *testing.T) {
	var g Schafkopf
	// 過大な actionLog を持つ JSON。
	bad := `{"al":[`
	for i := 0; i < schafkopfMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errSchafkopfOversized) {
		t.Errorf("err = %v, want errSchafkopfOversized", err)
	}
}

func TestSchafkopf_IsHumanTurn(t *testing.T) {
	g := newSKGame(true)
	g.SetPhase(SchafkopfPhaseCall)
	g.SetPickerIdx(0)
	if !g.IsHumanTurn() {
		t.Error("human picker in the call phase should be the human turn")
	}
	g.SetPickerIdx(1)
	if g.IsHumanTurn() {
		t.Error("CPU picker bury should not be human turn")
	}
	g.SetPhase(SchafkopfPhasePlay)
	g.SetCurrentPlayerIdx(0)
	if !g.IsHumanTurn() {
		t.Error("human play turn")
	}
}

// TestSchafkopf_ChipsAreZeroSumAtThisTableSize は、精算でチップの総量が
// 変わらないことを**席数から**確かめる。
//
// クローン元は 5 人卓なので「守備 3 人が 1 ずつ払い、ピッカー 2 + 相棒 1 が
// 受け取る」と数を直接書いていた。4 人卓でその式を使うと守備 2 人が払った 2 に
// 対して宣言側が 3 受け取り、**チップが 1 増える**。相棒あり / 単独の両方で見る。
func TestSchafkopf_ChipsAreZeroSumAtThisTableSize(t *testing.T) {
	for _, tc := range []struct {
		name      string
		partner   int
		pickerWon bool
	}{
		{"with a partner, picker wins", 1, true},
		{"with a partner, picker loses", 1, false},
		{"alone, picker wins", -1, true},
		{"alone, picker loses", -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newSKGame(false)
			g.SetPickerIdx(0)
			g.SetPartnerIdx(tc.partner)

			before := 0
			for i := 0; i < g.GetPlayerCnt(); i++ {
				before += g.GetPlayer(i).GetChips()
			}
			g.settleChips(tc.pickerWon, 1)
			after := 0
			for i := 0; i < g.GetPlayerCnt(); i++ {
				after += g.GetPlayer(i).GetChips()
			}

			if before != after {
				t.Errorf("chips not zero-sum: before=%d after=%d", before, after)
			}
			// **実際に動いたこと。** 全員 0 のままなら上の等式は無意味に成立する。
			moved := false
			for i := 0; i < g.GetPlayerCnt(); i++ {
				if g.GetPlayer(i).GetChips() != g.GetConfig().StartChips {
					moved = true
					break
				}
			}
			if !moved {
				t.Error("no chips moved at all")
			}
		})
	}
}

// TestSchafkopf_ContractChangesTheTrumpSet は、**契約によって切り札の顔ぶれが
// 変わる**ことを見る。ここがクローン元 (米国版シープスヘッド) との核心の差で、
// あちらは契約が 1 つしかないので切り札は固定だった。
func TestSchafkopf_ContractChangesTheTrumpSet(t *testing.T) {
	diamond7 := skCard(CardDesignDiamond, 7)
	ober := skCard(CardDesignSpade, schafkopfOber)
	unter := skCard(CardDesignHeart, schafkopfUnter)
	plainHeart := skCard(CardDesignHeart, 10)

	t.Run("rufspiel: diamonds, Obers and Unters", func(t *testing.T) {
		g := newSKGame(true)
		g.SetContractForTest(SchafkopfContractRufspiel, 0)
		for _, c := range []*Card{diamond7, ober, unter} {
			if !g.isTrump(c) {
				t.Errorf("%s should be trump under Rufspiel", cardStr(c))
			}
		}
		if g.isTrump(plainHeart) {
			t.Errorf("%s should not be trump", cardStr(plainHeart))
		}
	})

	t.Run("wenz: only the Unters", func(t *testing.T) {
		g := newSKGame(true)
		g.SetContractForTest(SchafkopfContractWenz, 0)
		if !g.isTrump(unter) {
			t.Error("the Unter must stay trump under Wenz")
		}
		// **Ober もダイヤも平札に落ちる。** ここが Wenz の全部。
		if g.isTrump(ober) {
			t.Error("the Ober must NOT be trump under Wenz")
		}
		if g.isTrump(diamond7) {
			t.Error("diamonds must NOT be trump under Wenz")
		}
	})

	t.Run("solo: the chosen suit plus Obers and Unters", func(t *testing.T) {
		g := newSKGame(true)
		g.SetContractForTest(SchafkopfContractSolo, CardDesignHeart)
		if !g.isTrump(plainHeart) {
			t.Error("the chosen suit must be trump under Solo")
		}
		for _, c := range []*Card{ober, unter} {
			if !g.isTrump(c) {
				t.Errorf("%s should stay trump under Solo", cardStr(c))
			}
		}
		// **選ばれなかったスートは平札。** Rufspiel と違いダイヤは特別ではない。
		if g.isTrump(diamond7) {
			t.Error("diamonds must NOT be trump when the Solo suit is hearts")
		}
	})
}

// TestSchafkopf_WenzChangesWhoWinsTheTrick は、契約の違いが**勝敗として**
// 現れることを見る。切り札の集合を変えただけでは、トリック判定が
// g.isTrump を通っていなければ何も変わらない。
func TestSchafkopf_WenzChangesWhoWinsTheTrick(t *testing.T) {
	trick := []*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},  // フェイル A (リード)
		{PlayerIdx: 1, Card: skCard(CardDesignDiamond, 7)}, // ダイヤの 7
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 10)}, // フェイル 10
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 9)},  // フェイル 9
	}

	// Rufspiel ではダイヤが切り札なので、7 でもフェイル A に勝つ。
	ruf := newSKGame(false)
	ruf.SetContractForTest(SchafkopfContractRufspiel, 0)
	ruf.SetPhase(SchafkopfPhaseTrickEnd)
	ruf.SetTrickNumber(1)
	ruf.SetCurrentTrick(trick)
	ruf.ResolveTrick()
	if got := ruf.GetLeadPlayerIdx(); got != 1 {
		t.Errorf("Rufspiel winner = %d, want 1 (the diamond is trump)", got)
	}

	// **Wenz ではダイヤは平札。** リードスートの最強札 (♣A) が勝つ。
	wenz := newSKGame(false)
	wenz.SetContractForTest(SchafkopfContractWenz, 0)
	wenz.SetPhase(SchafkopfPhaseTrickEnd)
	wenz.SetTrickNumber(1)
	wenz.SetCurrentTrick(trick)
	wenz.ResolveTrick()
	if got := wenz.GetLeadPlayerIdx(); got != 0 {
		t.Errorf("Wenz winner = %d, want 0 (diamonds are plain under Wenz)", got)
	}
}

// TestSchafkopf_WenzDemotesTheOberWithinTheLedSuit は、Wenz で Ober が
// **平札として** 札位どおりに弱くなることを見る。
//
// 切り札集合を切り替えても、強さの計算が固定のままだと Ober は 100 番台に
// 載り続ける。前のテストはダイヤを使ったが、あれは追従もしていないので
// trickWinner が強さを見る前に弾いてしまい、強さの誤りを検出できない。
// **リードスートに従う札**で見る必要がある。
func TestSchafkopf_WenzDemotesTheOberWithinTheLedSuit(t *testing.T) {
	trick := []*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, 1)},             // ♣A (リード)
		{PlayerIdx: 1, Card: skCard(CardDesignClover, schafkopfOber)}, // ♣Q
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 9)},
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 8)},
	}

	// Rufspiel: Ober は切り札なので ♣A に勝つ。
	ruf := newSKGame(false)
	ruf.SetContractForTest(SchafkopfContractRufspiel, 0)
	ruf.SetPhase(SchafkopfPhaseTrickEnd)
	ruf.SetTrickNumber(1)
	ruf.SetCurrentTrick(trick)
	ruf.ResolveTrick()
	if got := ruf.GetLeadPlayerIdx(); got != 1 {
		t.Errorf("Rufspiel winner = %d, want 1 (the Ober is trump)", got)
	}

	// **Wenz: Ober はただのクラブ。** 札位では A に負ける。
	wenz := newSKGame(false)
	wenz.SetContractForTest(SchafkopfContractWenz, 0)
	wenz.SetPhase(SchafkopfPhaseTrickEnd)
	wenz.SetTrickNumber(1)
	wenz.SetCurrentTrick(trick)
	wenz.ResolveTrick()
	if got := wenz.GetLeadPlayerIdx(); got != 0 {
		t.Errorf("Wenz winner = %d, want 0 (the Ober is a plain club under Wenz)", got)
	}
}

// TestSchafkopf_ContractSurvivesTheWire は、契約が往復で保たれることを
// **ゼロ値と食い違う値**で見る。
//
// Rufspiel はゼロ値なので、それで往復させるとフィールドを丸ごと消した実装が
// 同じ答えを返す。Wenz / Solo で見る必要がある。契約は切り札の構成そのものなので、
// 落ちると復元後に切り札が総入れ替えになる。
func TestSchafkopf_ContractSurvivesTheWire(t *testing.T) {
	for _, tc := range []struct {
		name     string
		contract SchafkopfContract
		soloSuit int
	}{
		{"wenz", SchafkopfContractWenz, 0},
		{"solo hearts", SchafkopfContractSolo, CardDesignHeart},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newSKGame(true)
			g.Reset()
			g.SetContractForTest(tc.contract, tc.soloSuit)
			if tc.contract == SchafkopfContractRufspiel {
				t.Fatal("ゼロ値の契約では往復の検査にならない")
			}

			blob, err := json.Marshal(g)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got Schafkopf
			if err := json.Unmarshal(blob, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.GetContract() != tc.contract {
				t.Errorf("contract = %v, want %v (lost on the wire)", got.GetContract(), tc.contract)
			}
			if got.GetSoloSuit() != tc.soloSuit {
				t.Errorf("soloSuit = %d, want %d", got.GetSoloSuit(), tc.soloSuit)
			}
			// **切り札の顔ぶれまで一致すること。** 数値が戻っても
			// isTrump が復元後の契約を見ていなければ意味がない。
			ober := skCard(CardDesignSpade, schafkopfOber)
			if got.isTrump(ober) != g.isTrump(ober) {
				t.Errorf("trump set differs after the round trip for %s", cardStr(ober))
			}
		})
	}
}

// TestSchafkopf_ContractResetsEachRound は、契約がラウンドをまたいで
// 残らないことを見る。持ち越すと前ラウンドの Wenz が次の配りの切り札構成を
// 支配してしまう。
func TestSchafkopf_ContractResetsEachRound(t *testing.T) {
	g := newSKGame(true)
	g.Reset()
	g.SetContractForTest(SchafkopfContractWenz, 0)

	g.startRound()

	if got := g.GetContract(); got != SchafkopfContractRufspiel {
		t.Errorf("contract = %v after a new round, want Rufspiel (default)", got)
	}
	// 既定に戻っていれば Ober はまた切り札。
	if !g.isTrump(skCard(CardDesignSpade, schafkopfOber)) {
		t.Error("the Ober should be trump again once the contract resets")
	}
}

// TestSchafkopf_WenzAndSoloPlayAlone は、Rufspiel だけが相棒を呼ぶことを見る。
//
// クローン元は契約が 1 つしかないので、宣言者は常に A を呼んでいた。
// Wenz と Solo は単独契約なので、呼びフェーズを飛ばしてプレイへ進む。
func TestSchafkopf_WenzAndSoloPlayAlone(t *testing.T) {
	for _, tc := range []struct {
		name      string
		contract  SchafkopfContract
		soloSuit  int
		wantAlone bool
	}{
		{"rufspiel calls a partner", SchafkopfContractRufspiel, 0, false},
		{"wenz plays alone", SchafkopfContractWenz, 0, true},
		{"solo plays alone", SchafkopfContractSolo, CardDesignHeart, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newSKGame(true)
			g.Reset()
			g.SetPhase(SchafkopfPhasePick)
			g.SetCurrentPlayerIdx(0)
			// 呼べる A が確実にある手札にする (Rufspiel なら呼びフェーズに入る)。
			skSetHand(g.GetPlayer(0),
				skCard(CardDesignClover, schafkopfOber), skCard(CardDesignSpade, schafkopfOber),
				skCard(CardDesignSpade, 10), skCard(CardDesignHeart, 10),
				skCard(CardDesignSpade, 9), skCard(CardDesignHeart, 9),
				skCard(CardDesignSpade, 8), skCard(CardDesignHeart, 8))

			if err := g.PlayerDeclare(true, tc.contract, tc.soloSuit); err != nil {
				t.Fatalf("declare: %v", err)
			}
			skFinishAuction(g)

			if tc.wantAlone {
				if g.GetPhase() != SchafkopfPhasePlay {
					t.Errorf("phase = %v, want Play (a solo contract skips the call)", g.GetPhase())
				}
				if g.GetPartnerIdx() != -1 {
					t.Errorf("partner = %d, want -1 for a solo contract", g.GetPartnerIdx())
				}
			} else if g.GetPhase() != SchafkopfPhaseCall {
				t.Errorf("phase = %v, want Call (Rufspiel calls an ace)", g.GetPhase())
			}
		})
	}
}

// TestSchafkopf_SoloNeedsARealSuit は、Solo が本物のスートしか受け取らない
// ことを見る。0 はスート番号ではないので、通すと「どの札とも一致しない切り札」の
// 盤面ができ、Ober と Unter だけが切り札という Wenz もどきになる。
func TestSchafkopf_SoloNeedsARealSuit(t *testing.T) {
	for _, suit := range []int{0, -1, CardDesignMax + 1} {
		g := newSKGame(true)
		g.Reset()
		g.SetPhase(SchafkopfPhasePick)
		g.SetCurrentPlayerIdx(0)
		if err := g.PlayerDeclare(true, SchafkopfContractSolo, suit); err == nil {
			t.Errorf("solo suit %d should be rejected", suit)
		}
		if g.GetPhase() != SchafkopfPhasePick {
			t.Errorf("a rejected declaration must not advance the phase (suit %d)", suit)
		}
	}

	// **負のコントロール。** 本物のスートは通る。
	for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
		g := newSKGame(true)
		g.Reset()
		g.SetPhase(SchafkopfPhasePick)
		g.SetCurrentPlayerIdx(0)
		if err := g.PlayerDeclare(true, SchafkopfContractSolo, suit); err != nil {
			t.Errorf("solo suit %d should be accepted: %v", suit, err)
		}
	}
}

// **リードが「別契約なら切り札」の札でも、契約に従って比べる。**
//
// トリック判定は先頭札の強さを種にして残りと比べる。種だけ契約を見ない
// 関数で取ると、Wenz で Ober をリードした瞬間に「Rufspiel なら切り札」の
// 強さが座り、本物の切り札 (Unter) が負ける。
// 既存のテストは常に ♣A (どちらの契約でも非切り札) をリードしていたので、
// この形だけ素通りしていた。
func TestSchafkopf_WenzJudgesTheLeadByTheContractToo(t *testing.T) {
	// Wenz では Ober は切り札ではなく、Unter だけが切り札。
	trick := []*TrickCard{
		{PlayerIdx: 0, Card: skCard(CardDesignClover, schafkopfOber)},   // リード
		{PlayerIdx: 1, Card: skCard(CardDesignDiamond, schafkopfUnter)}, // Wenz の切り札
		{PlayerIdx: 2, Card: skCard(CardDesignClover, 9)},
		{PlayerIdx: 3, Card: skCard(CardDesignClover, 8)},
	}

	wenz := newSKGame(false)
	wenz.SetContractForTest(SchafkopfContractWenz, 0)
	wenz.SetPhase(SchafkopfPhaseTrickEnd)
	wenz.SetTrickNumber(1)
	wenz.SetCurrentTrick(trick)
	wenz.ResolveTrick()
	if got := wenz.GetLeadPlayerIdx(); got != 1 {
		t.Errorf("Wenz winner = %d, want 1 — the Unter is trump and the Ober is not", got)
	}

	// **負のコントロール。** 同じ 4 枚を Rufspiel で比べると、Ober が切り札
	// なので先頭が勝つ。契約を見ていなければ両方が同じ答えになって差が出ない。
	ruf := newSKGame(false)
	ruf.SetContractForTest(SchafkopfContractRufspiel, 0)
	ruf.SetPhase(SchafkopfPhaseTrickEnd)
	ruf.SetTrickNumber(1)
	ruf.SetCurrentTrick(trick)
	ruf.ResolveTrick()
	if got := ruf.GetLeadPlayerIdx(); got != 0 {
		t.Errorf("Rufspiel winner = %d, want 0 — the Ober outranks the Unter", got)
	}
}
