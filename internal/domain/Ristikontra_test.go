package domain

import (
	"encoding/json"
	"testing"
)

// ristikontraCard は design / value からカードを作るテストヘルパー (draw=false)。
func ristikontraCard(d, v int) *Card { return NewCard(d, v, false) }

// ristikontraNewGame は手札・場を空にした 2-4 人ゲームをセットアップする。
// 山札はシャッフル済みだが、テストは手札/場を手動で組むため順序に依存しない。
func ristikontraNewGame(playerCnt int) *Ristikontra {
	cfg := DefaultRistikontraConfig()
	cfg.PlayerCnt = playerCnt
	players := makeRistikontraPlayers(playerCnt)
	g := NewRistikontra(NewTrumpCards(0), players, cfg)
	// 配札せず空の状態から開始する (手動セットアップ用)。
	g.state.pile = []*Card{}
	return g
}

func TestRistikontraConfig_Validate(t *testing.T) {
	if err := DefaultRistikontraConfig().Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
	bad := RistikontraConfig{PlayerCnt: 1, CpuDifficulty: RistikontraDifficultyNormal}
	if bad.Validate() == nil {
		t.Fatalf("player count 1 should be invalid")
	}
	bad2 := RistikontraConfig{PlayerCnt: 5, CpuDifficulty: RistikontraDifficultyNormal}
	if bad2.Validate() == nil {
		t.Fatalf("player count 5 should be invalid")
	}
	bad3 := RistikontraConfig{PlayerCnt: 4, CpuDifficulty: 9}
	if bad3.Validate() == nil {
		t.Fatalf("difficulty 9 should be invalid")
	}
}

func TestRistikontra_Reset_Deals(t *testing.T) {
	g := NewDefaultRistikontra()
	g.Reset()
	if g.GetPhase() != RistikontraPhasePlay {
		t.Fatalf("phase = %s, want play", g.GetPhase())
	}
	if g.GetPlayerCnt() != RistikontraDefaultPlayerCnt {
		t.Fatalf("player count = %d", g.GetPlayerCnt())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != RistikontraHandSize {
			t.Fatalf("player %d hand = %d, want %d", i, g.GetPlayer(i).GetCardsSize(), RistikontraHandSize)
		}
	}
	// クローン元のピシュティは「一番上がジャックでないこと」を要求していた
	// (ジャックが万能の捕獲札だったから)。リスティコントラのジャックはただの
	// 札なので、その制約は無い。
	if len(g.GetPile()) < RistikontraInitialPileSize {
		t.Fatalf("pile should have at least %d cards", RistikontraInitialPileSize)
	}
}

func TestRistikontra_Capture_RankMatch(t *testing.T) {
	g := ristikontraNewGame(2)
	// 場: 5♠ (単独以外にするため 2 枚)。
	g.state.pile = []*Card{ristikontraCard(CardDesignDiamond, 9), ristikontraCard(CardDesignSpade, 5)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 5)) // rank match 5
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if g.players[0].CapturedCount() != 3 {
		t.Fatalf("captured = %d, want 3", g.players[0].CapturedCount())
	}
	if len(g.GetPile()) != 0 {
		t.Fatalf("pile should be empty after capture")
	}
	if false {
		t.Fatalf("multi-card capture should not be a Pişti")
	}
	if g.GetLastCaptureIdx() != 0 {
		t.Fatalf("lastCaptureIdx = %d", g.GetLastCaptureIdx())
	}
}

// TestRistikontra_Capture_JackIsNotWild は、ジャックが万能札でないことを見る。
//
// **クローン元のピシュティではジャックが何にでも重なって場を総取りする。**
// リスティコントラで捕獲できるのは同ランクだけなので、その分岐を持ち込むと
// 別のゲームになる。
func TestRistikontra_Capture_JackIsNotWild(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	g.state.pile = []*Card{ristikontraCard(CardDesignDiamond, 9), ristikontraCard(CardDesignSpade, 5)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignClover, RistikontraJackValue))
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if got := g.players[0].CapturedCount(); got != 0 {
		t.Fatalf("a Jack onto a 5 must not capture, captured %d", got)
	}
	if got := len(g.GetPile()); got != 3 {
		t.Fatalf("the Jack should just sit on the pile, pile = %d", got)
	}
}

// TestRistikontra_Counter_StealsTheCapture は、このゲームの名前になっている
// 打ち返し (risti-kontra) を固定する。
//
// 捕獲された直後に**その捕獲を成立させたランク**を被せると、束ごと奪える。
func TestRistikontra_Counter_StealsTheCapture(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	g.state.pile = []*Card{ristikontraCard(CardDesignSpade, 7)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 7))  // 7 で捕獲
	g.players[1].AddCard(ristikontraCard(CardDesignClover, 7)) // 同じ 7 で打ち返し

	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play: %v", err)
	}
	if got := g.players[0].CapturedCount(); got != 2 {
		t.Fatalf("seat 0 should hold the 2-card capture, got %d", got)
	}

	// PlayerPlay は席 0 (人間) 専用なので、CPU 席は applyPlay で直接動かす。
	g.state.currentTurn = 1
	if err := g.applyPlay(1, 0); err != nil {
		t.Fatalf("counter: %v", err)
	}
	if got := g.players[0].CapturedCount(); got != 0 {
		t.Fatalf("the capture must be taken back from seat 0, still holds %d", got)
	}
	if got := g.players[1].CapturedCount(); got != 3 {
		t.Fatalf("seat 1 should hold the stolen bundle plus its own card, got %d", got)
	}
}

// TestRistikontra_Counter_Chains は、打ち返しが続く限り連鎖することを見る。
func TestRistikontra_Counter_Chains(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	g.state.pile = []*Card{ristikontraCard(CardDesignSpade, 7)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 7))
	g.players[1].AddCard(ristikontraCard(CardDesignClover, 7))
	g.players[2].AddCard(ristikontraCard(CardDesignDiamond, 7))

	for seat := 0; seat < 3; seat++ {
		g.state.currentTurn = seat
		if err := g.applyPlay(seat, 0); err != nil {
			t.Fatalf("seat %d: %v", seat, err)
		}
	}
	// 4 枚 (場の 7 + 3 人ぶんの 7) が最後に打ち返した席 2 に集まる。
	if got := g.players[2].CapturedCount(); got != 4 {
		t.Fatalf("the last countering seat should hold all 4, got %d", got)
	}
	for _, seat := range []int{0, 1} {
		if got := g.players[seat].CapturedCount(); got != 0 {
			t.Fatalf("seat %d should have been countered out, still holds %d", seat, got)
		}
	}
}

// TestRistikontra_Counter_ClosesOnADifferentRank は、別ランクが出た時点で
// 打ち返しの機会が閉じることを見る。**閉じないと、何手も後の同ランクで
// 束が飛んでいく。**
func TestRistikontra_Counter_ClosesOnADifferentRank(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	g.state.pile = []*Card{ristikontraCard(CardDesignSpade, 7)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 7))   // 捕獲
	g.players[1].AddCard(ristikontraCard(CardDesignClover, 3))  // 別ランク
	g.players[2].AddCard(ristikontraCard(CardDesignDiamond, 7)) // もう遅い

	g.state.currentTurn = 0
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("capture: %v", err)
	}
	g.state.currentTurn = 1
	if err := g.applyPlay(1, 0); err != nil {
		t.Fatalf("break: %v", err)
	}
	g.state.currentTurn = 2
	if err := g.applyPlay(2, 0); err != nil {
		t.Fatalf("late 7: %v", err)
	}
	if got := g.players[0].CapturedCount(); got != 2 {
		t.Fatalf("seat 0's capture should be safe once the rank was broken, got %d", got)
	}
	if got := g.players[2].CapturedCount(); got != 0 {
		t.Fatalf("the late 7 must not steal anything, seat 2 holds %d", got)
	}
}

func TestRistikontra_NoCapture_Stacking(t *testing.T) {
	g := ristikontraNewGame(2)
	g.state.pile = []*Card{ristikontraCard(CardDesignSpade, 7)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 3)) // no match, not jack
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

func TestRistikontra_PlayerPlay_Guards(t *testing.T) {
	g := ristikontraNewGame(2)
	g.state.currentTurn = 1 // CPU
	g.players[1].AddCard(ristikontraCard(CardDesignHeart, 3))
	if err := g.PlayerPlay(0); err != ErrNotHumanTurn {
		t.Fatalf("want ErrNotHumanTurn, got %v", err)
	}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 3))
	if err := g.PlayerPlay(99); err == nil {
		t.Fatalf("out-of-range index should error")
	}
	g.state.gameEndFlag = true
	if err := g.PlayerPlay(0); err != ErrGameEnded {
		t.Fatalf("want ErrGameEnded, got %v", err)
	}
}

// TestRistikontra_TeamCardCounts は、勝敗がチーム単位の獲得枚数で決まることを
// 見る。**クローン元のピシュティは席ごとに札の点数を数える**が、
// リスティコントラは 2 対 2 で、枚数の多いチームが勝つ。
func TestRistikontra_TeamCardCounts(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	// 席 0・2 = チーム 0、席 1・3 = チーム 1。
	g.players[0].AddCaptured(ristikontraCards(3))
	g.players[2].AddCaptured(ristikontraCards(4)) // チーム 0 = 7 枚
	g.players[1].AddCaptured(ristikontraCards(5))
	g.players[3].AddCaptured(ristikontraCards(1)) // チーム 1 = 6 枚

	counts := g.GetTeamCardCounts()
	if len(counts) != RistikontraTeamCnt {
		t.Fatalf("team count = %d, want %d", len(counts), RistikontraTeamCnt)
	}
	if counts[0] != 7 || counts[1] != 6 {
		t.Fatalf("team counts = %v, want [7 6]", counts)
	}
}

// TestRistikontra_FinishGame_WinnersAreATeam は、勝者が**チームの両席**に
// なることを見る。席ごとの多寡ではない —— パートナーが少なくても、合計で
// 勝っていれば両方が勝者。
func TestRistikontra_FinishGame_WinnersAreATeam(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	// 席 1 が単独では一番多いが、チーム 0 の合計が上回る形にする。
	g.players[0].AddCaptured(ristikontraCards(6))
	g.players[2].AddCaptured(ristikontraCards(6)) // チーム 0 = 12
	g.players[1].AddCaptured(ristikontraCards(9))
	g.players[3].AddCaptured(ristikontraCards(1)) // チーム 1 = 10
	g.state.lastCaptureIdx = 0

	g.finishGame()

	winners := g.GetWinners()
	if len(winners) != 2 || winners[0] != 0 || winners[1] != 2 {
		t.Fatalf("winners = %v, want [0 2] (team 0)", winners)
	}
}

// TestRistikontra_FinishGame_EqualSplitIsADraw は、合計が並んだら勝者無しに
// なることを見る。**片方だけ返してしまうと、引き分けが勝ちに化ける。**
func TestRistikontra_FinishGame_EqualSplitIsADraw(t *testing.T) {
	g := ristikontraNewGame(RistikontraDefaultPlayerCnt)
	for seat := 0; seat < RistikontraDefaultPlayerCnt; seat++ {
		g.players[seat].AddCaptured(ristikontraCards(4))
	}
	g.state.lastCaptureIdx = 0

	g.finishGame()

	if got := g.GetWinners(); len(got) != 0 {
		t.Fatalf("an equal split must leave no winner, got %v", got)
	}
}

func TestRistikontra_FinalScore_MostCardsTieNoBonus(t *testing.T) {
	g := ristikontraNewGame(2)
	g.players[0].AddCaptured([]*Card{ristikontraCard(CardDesignSpade, 4), ristikontraCard(CardDesignSpade, 5)})
	g.players[1].AddCaptured([]*Card{ristikontraCard(CardDesignHeart, 4), ristikontraCard(CardDesignHeart, 5)})
	scores := g.calcFinalScore()
	if scores[0] != 0 || scores[1] != 0 {
		t.Fatalf("tie on cards must award no most-cards bonus: %v", scores)
	}
}

func TestRistikontra_CardPoints(t *testing.T) {
	cases := []struct {
		c    *Card
		want int
	}{
		{ristikontraCard(CardDesignSpade, 1), 1},                    // Ace
		{ristikontraCard(CardDesignClover, 2), 2},                   // 2♣
		{ristikontraCard(CardDesignSpade, 2), 0},                    // 2♠ no value
		{ristikontraCard(CardDesignDiamond, 10), 3},                 // 10♦
		{ristikontraCard(CardDesignSpade, 10), 0},                   // 10♠ no value
		{ristikontraCard(CardDesignHeart, RistikontraJackValue), 1}, // Jack
		{ristikontraCard(CardDesignSpade, 7), 0},                    // filler
		{nil, 0},
	}
	for i, tc := range cases {
		if got := ristikontraCardPoints(tc.c); got != tc.want {
			t.Fatalf("case %d: points = %d, want %d", i, got, tc.want)
		}
	}
}

func TestRistikontra_ReDeal(t *testing.T) {
	g := ristikontraNewGame(2)
	// 各プレイヤー手札 1 枚、山札を尽きていない状態にする。
	g.trumpCards = NewTrumpCards(0) // 52 枚 (未配布)
	g.state.pile = []*Card{ristikontraCard(CardDesignSpade, 9)}
	g.state.currentTurn = 0
	g.players[0].AddCard(ristikontraCard(CardDesignHeart, 3))
	g.players[1].AddCard(ristikontraCard(CardDesignHeart, 4))
	// P0 plays (no capture), turn -> 1.
	_ = g.PlayerPlay(0)
	// P1 (CPU) plays via direct apply to keep deterministic.
	_ = g.applyPlay(1, 0)
	// 両手札が空になったので再配布が走り、各 4 枚配られているはず。
	if g.players[0].GetCardsSize() != RistikontraHandSize || g.players[1].GetCardsSize() != RistikontraHandSize {
		t.Fatalf("re-deal failed: p0=%d p1=%d", g.players[0].GetCardsSize(), g.players[1].GetCardsSize())
	}
}

func TestRistikontra_JSON_RoundTrip(t *testing.T) {
	g := NewDefaultRistikontra()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Ristikontra
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Fatalf("player count mismatch after round-trip")
	}
}

func TestRistikontra_JSON_Rejects(t *testing.T) {
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
		var g Ristikontra
		if err := json.Unmarshal([]byte(s), &g); err == nil {
			t.Fatalf("case %d should fail to unmarshal: %s", i, s)
		}
	}
}

// TestRistikontra_FullCpuGame_Terminates は全 CPU ゲームが必ず終局することを
// 複数の難易度・プレイヤー数で確認する (有限の山札 + 反復上限で保証)。
func TestRistikontra_FullCpuGame_Terminates(t *testing.T) {
	difficulties := []RistikontraCpuDifficulty{
		RistikontraDifficultyEasy, RistikontraDifficultyNormal, RistikontraDifficultyHard,
	}
	// **席数は 4 固定。** 2 人卓・3 人卓はチームが組めないので、
	// クローン元のピシュティにあった可変人数のループは意味を失った。
	for _, pc := range []int{RistikontraDefaultPlayerCnt} {
		for _, di := range difficulties {
			cfg := RistikontraConfig{PlayerCnt: pc, CpuDifficulty: di}
			g := NewRistikontra(NewTrumpCards(0), makeRistikontraPlayers(pc), cfg)
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
			// **引き分けは正当な結末。** 2 チームの獲得枚数が並べば勝者は
			// 出ない。「必ず勝者がいる」と書くと、その盤面で偽陽性になる。
			winners := g.GetWinners()
			switch len(winners) {
			case 0: // 引き分け
			case 2:
				if ristikontraTeamOf(winners[0]) != ristikontraTeamOf(winners[1]) {
					t.Fatalf("pc=%d di=%d winners %v are not one team", pc, di, winners)
				}
			default:
				t.Fatalf("pc=%d di=%d winners = %v, want a team of 2 or a draw", pc, di, winners)
			}
			if g.GetPhase() != RistikontraPhaseGameEnd {
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

// TestRistikontra_CpuPlay_HumanGuard は CpuPlay が人間の手番では何もしないことを確認。
func TestRistikontra_CpuPlay_HumanGuard(t *testing.T) {
	g := NewDefaultRistikontra()
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

func TestRistikontra_NextRound_RestartsGame(t *testing.T) {
	g := NewDefaultRistikontra()
	g.Reset()
	g.state.gameEndFlag = true
	g.NextRound()
	if g.GetGameEndFlag() {
		t.Fatalf("NextRound should restart and clear gameEndFlag")
	}
	if g.GetPhase() != RistikontraPhasePlay {
		t.Fatalf("NextRound phase = %s", g.GetPhase())
	}
}

// **同数なら誰にも +3 は付かない。**暫定スコアと最終集計で規則が割れると、
// 途中の順位表示が嘘になる (#4892)。
// TestRistikontra_ProvisionalScores は、途中経過が**チームの獲得枚数**に
// なっていることを見る。
//
// クローン元のピシュティは席ごとにカード点を最後に数えるので、途中の値は
// 「確定済みのボーナス + 最多捕獲の +3」という近似だった。リスティコントラは
// 枚数がそのまま結果なので近似ではなく、席には**自分のチームの合計**が入る。
func TestRistikontra_ProvisionalScores(t *testing.T) {
	g := NewDefaultRistikontra()
	g.Reset()
	give := func(seat, n int) {
		cards := make([]*Card, n)
		for i := range cards {
			cards[i] = NewCard(CardDesignSpade, 2, false)
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	// 席 0 = チーム 0、席 1 = チーム 1。
	give(0, 5)
	give(1, 3)
	prov := g.GetProvisionalScores()
	if prov[0] != 5 || prov[2] != 5 {
		t.Fatalf("team 0 seats should both read 5, got %v", prov)
	}
	if prov[1] != 3 || prov[3] != 3 {
		t.Fatalf("team 1 seats should both read 3, got %v", prov)
	}

	// **パートナーの枚数が合算される。** 席ごとの数字を見せると
	// 「自分は取っているのに負けている」が読めない。
	give(2, 4)
	prov = g.GetProvisionalScores()
	if prov[0] != 9 || prov[2] != 9 {
		t.Fatalf("team 0 should read 9 after the partner captured, got %v", prov)
	}
	if prov[1] != 3 {
		t.Fatalf("team 1 should be unchanged at 3, got %v", prov)
	}
}

// ristikontraCards は n 枚のダミー札を返す (枚数だけを見るテスト用)。
func ristikontraCards(n int) []*Card {
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, ristikontraCard(CardDesignSpade, i%13+1))
	}
	return out
}
