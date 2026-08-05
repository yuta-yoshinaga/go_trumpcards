package domain

import (
	"encoding/json"
	"testing"
)

// pishtiCard は design / value からカードを作るテストヘルパー (draw=false)。
func pishtiCard(d, v int) *Card { return NewCard(d, v, false) }

// pishtiNewGame は手札・場を空にした 2-4 人ゲームをセットアップする。
// 山札はシャッフル済みだが、テストは手札/場を手動で組むため順序に依存しない。
func pishtiNewGame(playerCnt int) *Pishti {
	cfg := DefaultPishtiConfig()
	cfg.PlayerCnt = playerCnt
	players := makePishtiPlayers(playerCnt)
	g := NewPishti(NewTrumpCards(0), players, cfg)
	// 配札せず空の状態から開始する (手動セットアップ用)。
	g.state.pile = []*Card{}
	return g
}

func TestPishtiConfig_Validate(t *testing.T) {
	if err := DefaultPishtiConfig().Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
	bad := PishtiConfig{PlayerCnt: 1, CpuDifficulty: PishtiDifficultyNormal}
	if bad.Validate() == nil {
		t.Fatalf("player count 1 should be invalid")
	}
	bad2 := PishtiConfig{PlayerCnt: 5, CpuDifficulty: PishtiDifficultyNormal}
	if bad2.Validate() == nil {
		t.Fatalf("player count 5 should be invalid")
	}
	bad3 := PishtiConfig{PlayerCnt: 4, CpuDifficulty: 9}
	if bad3.Validate() == nil {
		t.Fatalf("difficulty 9 should be invalid")
	}
}

func TestPishti_Reset_DealsAndNonJackTop(t *testing.T) {
	g := NewDefaultPishti()
	g.Reset()
	if g.GetPhase() != PishtiPhasePlay {
		t.Fatalf("phase = %s, want play", g.GetPhase())
	}
	if g.GetPlayerCnt() != PishtiDefaultPlayerCnt {
		t.Fatalf("player count = %d", g.GetPlayerCnt())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != PishtiHandSize {
			t.Fatalf("player %d hand = %d, want %d", i, g.GetPlayer(i).GetCardsSize(), PishtiHandSize)
		}
	}
	if top := g.GetPileTop(); top != nil && pishtiIsJack(top) {
		t.Fatalf("initial pile top must not be a Jack, got value %d", top.GetValue())
	}
	if len(g.GetPile()) < PishtiInitialPileSize {
		t.Fatalf("pile should have at least %d cards", PishtiInitialPileSize)
	}
}

func TestPishti_Capture_RankMatch(t *testing.T) {
	g := pishtiNewGame(2)
	// 場: 5♠ (単独以外にするため 2 枚)。
	g.state.pile = []*Card{pishtiCard(CardDesignDiamond, 9), pishtiCard(CardDesignSpade, 5)}
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, 5)) // rank match 5
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].CapturedCount() != 3 {
		t.Fatalf("captured = %d, want 3", g.players[0].CapturedCount())
	}
	if len(g.GetPile()) != 0 {
		t.Fatalf("pile should be empty after capture")
	}
	if g.players[0].GetPistiBonus() != 0 {
		t.Fatalf("multi-card capture should not be a Pişti")
	}
	if g.GetLastCaptureIdx() != 0 {
		t.Fatalf("lastCaptureIdx = %d", g.GetLastCaptureIdx())
	}
}

func TestPishti_Capture_JackWild(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.pile = []*Card{pishtiCard(CardDesignDiamond, 9), pishtiCard(CardDesignSpade, 5)}
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignClover, PishtiJackValue)) // Jack
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].CapturedCount() != 3 {
		t.Fatalf("jack should capture whole pile, got %d", g.players[0].CapturedCount())
	}
	if g.players[0].GetPistiBonus() != 0 {
		t.Fatalf("jack on multi-card pile is not a Pişti")
	}
}

func TestPishti_Pisti_SingleCardBonus(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.pile = []*Card{pishtiCard(CardDesignSpade, 7)} // lone single
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, 7)) // rank match
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].GetPistiBonus() != PishtiBonusSingle {
		t.Fatalf("Pişti bonus = %d, want %d", g.players[0].GetPistiBonus(), PishtiBonusSingle)
	}
	if g.players[0].CapturedCount() != 2 {
		t.Fatalf("captured = %d, want 2", g.players[0].CapturedCount())
	}
}

func TestPishti_Pisti_JackOnSingleNonJack(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.pile = []*Card{pishtiCard(CardDesignSpade, 7)} // lone non-Jack
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, PishtiJackValue))
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].GetPistiBonus() != PishtiBonusSingle {
		t.Fatalf("Jack on lone non-Jack should be +%d, got %d", PishtiBonusSingle, g.players[0].GetPistiBonus())
	}
}

func TestPishti_Pisti_JackOnSingleJack(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.pile = []*Card{pishtiCard(CardDesignSpade, PishtiJackValue)} // lone Jack
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, PishtiJackValue))
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].GetPistiBonus() != PishtiBonusJackOnJack {
		t.Fatalf("Jack on lone Jack should be +%d, got %d", PishtiBonusJackOnJack, g.players[0].GetPistiBonus())
	}
}

func TestPishti_NoCapture_Stacking(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.pile = []*Card{pishtiCard(CardDesignSpade, 7)}
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, 3)) // no match, not jack
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].CapturedCount() != 0 {
		t.Fatalf("should not capture")
	}
	if len(g.GetPile()) != 2 {
		t.Fatalf("pile should grow to 2, got %d", len(g.GetPile()))
	}
	top := g.GetPileTop()
	if top.GetValue() != 3 {
		t.Fatalf("top should be the just-played 3, got %d", top.GetValue())
	}
}

func TestPishti_PlayerPlay_Guards(t *testing.T) {
	g := pishtiNewGame(2)
	g.state.currentTurn = 1 // CPU
	g.players[1].AddCard(pishtiCard(CardDesignHeart, 3))
	if err := g.PlayerPlay(0); err != ErrNotHumanTurn {
		t.Fatalf("want ErrNotHumanTurn, got %v", err)
	}
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, 3))
	if err := g.PlayerPlay(99); err == nil {
		t.Fatalf("out-of-range index should error")
	}
	g.state.gameEndFlag = true
	if err := g.PlayerPlay(0); err != ErrGameEnded {
		t.Fatalf("want ErrGameEnded, got %v", err)
	}
}

func TestPishti_FinalScore_AllBonuses(t *testing.T) {
	g := pishtiNewGame(2)
	// Player 0: most cards + an Ace + 2♣ + 10♦ + a Jack.
	g.players[0].AddCaptured([]*Card{
		pishtiCard(CardDesignSpade, 1),               // Ace +1
		pishtiCard(CardDesignClover, 2),              // 2♣ +2
		pishtiCard(CardDesignDiamond, 10),            // 10♦ +3
		pishtiCard(CardDesignHeart, PishtiJackValue), // Jack +1
		pishtiCard(CardDesignSpade, 4),               // filler
	})
	g.players[0].AddPistiBonus(10)
	// Player 1: fewer cards, one Ace.
	g.players[1].AddCaptured([]*Card{pishtiCard(CardDesignHeart, 1)}) // Ace +1

	scores := g.calcFinalScore()
	// P0: most(3)+ace(1)+2c(2)+10d(3)+jack(1)+pisti(10) = 20
	if scores[0] != 20 {
		t.Fatalf("player0 score = %d, want 20", scores[0])
	}
	// P1: ace(1) = 1
	if scores[1] != 1 {
		t.Fatalf("player1 score = %d, want 1", scores[1])
	}
}

func TestPishti_FinalScore_MostCardsTieNoBonus(t *testing.T) {
	g := pishtiNewGame(2)
	g.players[0].AddCaptured([]*Card{pishtiCard(CardDesignSpade, 4), pishtiCard(CardDesignSpade, 5)})
	g.players[1].AddCaptured([]*Card{pishtiCard(CardDesignHeart, 4), pishtiCard(CardDesignHeart, 5)})
	scores := g.calcFinalScore()
	if scores[0] != 0 || scores[1] != 0 {
		t.Fatalf("tie on cards must award no most-cards bonus: %v", scores)
	}
}

func TestPishti_CardPoints(t *testing.T) {
	cases := []struct {
		c    *Card
		want int
	}{
		{pishtiCard(CardDesignSpade, 1), 1},               // Ace
		{pishtiCard(CardDesignClover, 2), 2},              // 2♣
		{pishtiCard(CardDesignSpade, 2), 0},               // 2♠ no value
		{pishtiCard(CardDesignDiamond, 10), 3},            // 10♦
		{pishtiCard(CardDesignSpade, 10), 0},              // 10♠ no value
		{pishtiCard(CardDesignHeart, PishtiJackValue), 1}, // Jack
		{pishtiCard(CardDesignSpade, 7), 0},               // filler
		{nil, 0},
	}
	for i, tc := range cases {
		if got := pishtiCardPoints(tc.c); got != tc.want {
			t.Fatalf("case %d: points = %d, want %d", i, got, tc.want)
		}
	}
}

func TestPishti_ReDeal(t *testing.T) {
	g := pishtiNewGame(2)
	// 各プレイヤー手札 1 枚、山札を尽きていない状態にする。
	g.trumpCards = NewTrumpCards(0) // 52 枚 (未配布)
	g.state.pile = []*Card{pishtiCard(CardDesignSpade, 9)}
	g.state.currentTurn = 0
	g.players[0].AddCard(pishtiCard(CardDesignHeart, 3))
	g.players[1].AddCard(pishtiCard(CardDesignHeart, 4))
	// P0 plays (no capture), turn -> 1.
	_ = g.PlayerPlay(0)
	// P1 (CPU) plays via direct apply to keep deterministic.
	_ = g.applyPlay(1, 0)
	// 両手札が空になったので再配布が走り、各 4 枚配られているはず。
	if g.players[0].GetCardsSize() != PishtiHandSize || g.players[1].GetCardsSize() != PishtiHandSize {
		t.Fatalf("re-deal failed: p0=%d p1=%d", g.players[0].GetCardsSize(), g.players[1].GetCardsSize())
	}
}

func TestPishti_JSON_RoundTrip(t *testing.T) {
	g := NewDefaultPishti()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Pishti
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Fatalf("player count mismatch after round-trip")
	}
}

func TestPishti_JSON_Rejects(t *testing.T) {
	bad := []string{
		`{"tc":null,"pl":[],"cf":{"pc":4,"di":1}}`,                               // nil trump cards
		`{"tc":{},"pl":[{},{}],"cf":{"pc":1,"di":1}}`,                            // invalid config (pc 1)
		`{"tc":{},"pl":[{}],"cf":{"pc":4,"di":1}}`,                               // player count out of range
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"pc":4,"di":1},"ph":"bogus"}`,         // invalid phase
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"pc":4,"di":1},"ph":"play","ct":9}`,   // current turn out of range
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"pc":4,"di":1},"ph":"play","lc":9}`,   // last capture out of range
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"pc":4,"di":1},"ph":"play","wn":[9]}`, // winner out of range
	}
	for i, s := range bad {
		var g Pishti
		if err := json.Unmarshal([]byte(s), &g); err == nil {
			t.Fatalf("case %d should fail to unmarshal: %s", i, s)
		}
	}
}

// TestPishti_FullCpuGame_Terminates は全 CPU ゲームが必ず終局することを
// 複数の難易度・プレイヤー数で確認する (有限の山札 + 反復上限で保証)。
func TestPishti_FullCpuGame_Terminates(t *testing.T) {
	difficulties := []PishtiCpuDifficulty{
		PishtiDifficultyEasy, PishtiDifficultyNormal, PishtiDifficultyHard,
	}
	for _, pc := range []int{2, 3, 4} {
		for _, di := range difficulties {
			cfg := PishtiConfig{PlayerCnt: pc, CpuDifficulty: di}
			g := NewPishti(NewTrumpCards(0), makePishtiPlayers(pc), cfg)
			g.Reset()
			// applyPlay と chooseCpuCard は座席の人間/CPU 属性に依存しないため、
			// 全座席を自動で進行させて確実に終局するか検証する。
			const maxIter = 5000
			iter := 0
			for !g.GetGameEndFlag() {
				iter++
				if iter > maxIter {
					t.Fatalf("pc=%d di=%d did not terminate", pc, di)
				}
				cur := g.GetCurrentTurn()
				if g.GetPlayer(cur).GetCardsSize() == 0 {
					t.Fatalf("pc=%d di=%d: current player has no cards but game not ended", pc, di)
				}
				idx := g.chooseCpuCard(cur)
				_ = g.applyPlay(cur, idx)
			}
			if len(g.GetWinners()) == 0 {
				t.Fatalf("pc=%d di=%d should have a winner", pc, di)
			}
			if g.GetPhase() != PishtiPhaseGameEnd {
				t.Fatalf("pc=%d di=%d phase=%s", pc, di, g.GetPhase())
			}
			// 終局後は場が空で、捕獲枚数の合計は最大 52 枚に収まる。
			if len(g.GetPile()) != 0 {
				t.Fatalf("pc=%d di=%d pile not empty at game end", pc, di)
			}
			total := 0
			for i := 0; i < g.GetPlayerCnt(); i++ {
				total += g.GetPlayer(i).CapturedCount()
			}
			if total > 52 {
				t.Fatalf("pc=%d di=%d captured total = %d, exceeds 52", pc, di, total)
			}
		}
	}
}

// TestPishti_CpuPlay_HumanGuard は CpuPlay が人間の手番では何もしないことを確認。
func TestPishti_CpuPlay_HumanGuard(t *testing.T) {
	g := NewDefaultPishti()
	g.Reset()
	// seat 0 は人間。CpuPlay は何もしないはず。
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	if g.GetPlayer(0).GetCardsSize() != before {
		t.Fatalf("CpuPlay should not move on human turn")
	}
	// CPU の手番にすると CpuPlay が進める。
	g.state.currentTurn = 1
	beforeCpu := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	if g.GetPlayer(1).GetCardsSize() != beforeCpu-1 {
		t.Fatalf("CpuPlay should play one card on CPU turn")
	}
}

func TestPishti_NextRound_RestartsGame(t *testing.T) {
	g := NewDefaultPishti()
	g.Reset()
	g.state.gameEndFlag = true
	g.NextRound()
	if g.GetGameEndFlag() {
		t.Fatalf("NextRound should restart and clear gameEndFlag")
	}
	if g.GetPhase() != PishtiPhasePlay {
		t.Fatalf("NextRound phase = %s", g.GetPhase())
	}
}

// **同数なら誰にも +3 は付かない。**暫定スコアと最終集計で規則が割れると、
// 途中の順位表示が嘘になる (#4892)。
func TestPishti_ProvisionalScores(t *testing.T) {
	g := NewDefaultPishti()
	g.Reset()
	card := func() *Card { return NewCard(CardDesignSpade, 2, false) }
	give := func(seat, n int) {
		cards := make([]*Card, n)
		for i := range cards {
			cards[i] = card()
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	// 誰も捕獲していなければリーダー無し。
	if got := g.GetProvisionalLeader(); got != -1 {
		t.Fatalf("no captures should mean no leader, got %d", got)
	}

	// 単独リーダーに +3。
	give(0, 5)
	give(1, 3)
	if got := g.GetProvisionalLeader(); got != 0 {
		t.Fatalf("leader = %d, want 0", got)
	}
	prov := g.GetProvisionalScores()
	if prov[0] != PishtiScoreMostCards {
		t.Fatalf("sole leader should get +%d, got %d", PishtiScoreMostCards, prov[0])
	}
	if prov[1] != 0 {
		t.Fatalf("non-leader should get nothing, got %d", prov[1])
	}

	// **同数になったら誰にも付かない** (受け入れ条件2)。
	give(1, 2)
	if got := g.GetProvisionalLeader(); got != -1 {
		t.Fatalf("a tie should have no leader, got %d", got)
	}
	for i, v := range g.GetProvisionalScores() {
		if v != 0 {
			t.Fatalf("seat %d got %d on a tie, want 0", i, v)
		}
	}
}
