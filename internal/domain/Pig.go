//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// PigPhase はゲームフェーズ。
type PigPhase int

// ピッグのフェーズ定数
const (
	// PigPhasePass 全員が同時に 1 枚ずつ左へ回している状態
	PigPhasePass PigPhase = 0
	// PigPhaseSignal 誰かが 4 枚揃えて合図を出し、残りが気づく番になった状態
	PigPhaseSignal PigPhase = 1
	// PigPhaseRoundEnd ラウンド終了 (最後に気づいた 1 人へ文字が付いた)
	PigPhaseRoundEnd PigPhase = 2
	// PigPhaseGameEnd ゲーム終了 (生存者が 1 人になった)
	PigPhaseGameEnd PigPhase = 3
)

// PigPhaseMin / PigPhaseMax はフェーズ列挙の範囲 (復元時の検証用)。
const (
	PigPhaseMin = PigPhasePass
	PigPhaseMax = PigPhaseGameEnd
)

// pigMaxSliceLen は復元時に受け付けるスライス長の上限。
const pigMaxSliceLen = 4000

// PigHint は人間への助言。
type PigHint struct {
	// CardIndex は渡すべき手札の位置。合図の場面では nil。
	CardIndex *int
	// Reason は助言の理由。
	Reason string
}

// Pig はピッグ (Pig / ドンキー) のゲームクラス。
//
// アメリカ・イギリスの子供向けパーティゲーム。**同じランク 4 枚を揃えた人は
// 声を出さず、黙って手を鼻に当てます。** 他の人はそれに気づいて真似るだけで、
// **最後まで気づかなかった 1 人**が P・I・G の文字を 1 つ受け取ります。3 文字
// 揃うと脱落し、最後まで残った 1 人が勝ちです。
//
// # 取り合いではなく、気づき
//
// [Spoons] とよく似ていますが、罰の理由が違います。スプーンは**物を取り合う**
// ので取れなかった人が負け、ピッグは**誰も何も取らない**——遅れた人だけが罰を
// 受けます。合図に気づくのが遅い、それだけが負けの条件です。
//
// # 同時パスのモデル化
//
// 本来は全員が同時に 1 枚を左へ回しますが、サーバ側では決定的な手番として
// モデル化します。**枚数は最初から最後まで 4 枚のまま**——渡すのと受け取るのが
// 同時だからです。人間が `PlayerPass` で渡す札を選び、CPU は `CpuPlay` で
// 自動的に選び、全員が選び終わった時点で一斉に左へ移します。
//
// # デッキ
//
// **人数と同じ種類のランクを 4 スート揃えて使います。** 4 人なら 4 ランク ×
// 4 スート = 16 枚。**新しい Card / Deck 型は要りません**——標準 52 枚から
// ランクを絞った部分集合です（issue の「非52枚デッキなので新しい型が必要」は
// 誤り）。
//
// # 停止保証
//
// 毎ラウンド必ず 1 人に文字が付くので文字は単調に増え、いずれ脱落者が出て
// 生存者が減ります。加えて、揃わないまま回り続ける配りに備えて
// [PigMaxPassesPerRound] と [PigMaxRounds] のガードを置いています。
type Pig struct {
	players []*PigPlayer
	config  PigConfig
	phase   PigPhase

	// pendingPass[i] は席 i が渡すと決めた札 (nil = まだ決めていない)。
	//
	// **同時に渡すので、全員が決まるまで盤面は動きません。**
	pendingPass []*Card

	currentPlayerIdx int
	// signallerIdx は 4 枚揃えて最初に合図した席 (-1 = 合図なし)。
	signallerIdx int
	// noticedCnt は合図に気づいた人数 (合図した本人を含む)。
	noticedCnt int
	// roundLoserIdx は直近ラウンドで文字が付いた席 (-1 = 未)。
	roundLoserIdx int
	passCount     int
	roundNumber   int
	gameEndFlag   bool
	winnerIdx     int
	actionLogBase

	// rng CPU の気づき抽選用 (テストで差し替え可能)
	rng *rand.Rand
}

// NewPig はコンストラクタ。
func NewPig(players []*PigPlayer, config PigConfig) *Pig {
	if config.PlayerCnt < PigPlayerCntMin || config.PlayerCnt > PigPlayerCntMax {
		config.PlayerCnt = PigDefaultPlayerCnt
	}
	if players == nil {
		players = newPigSeats(config.PlayerCnt)
	}
	return &Pig{
		players:       players,
		config:        config,
		winnerIdx:     -1,
		signallerIdx:  -1,
		roundLoserIdx: -1,
		rng:           rand.New(rand.NewSource(rand.Int63())),
	}
}

// newPigSeats は席 0 を人間、以降を CPU とした座席を作る。
func newPigSeats(n int) []*PigPlayer {
	seats := make([]*PigPlayer, n)
	for i := range seats {
		seats[i] = NewPigPlayer(i == 0)
	}
	return seats
}

// NewDefaultPig は既定設定の Pig を返す。
func NewDefaultPig() *Pig {
	cfg := DefaultPigConfig()
	return NewPig(newPigSeats(cfg.PlayerCnt), cfg)
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Pig) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲーム全体を初期化して新しいゲームを開始する。
func (g *Pig) Reset() {
	for _, p := range g.players {
		p.SetLetters(0)
		p.SetEliminated(false)
	}
	g.roundNumber = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundLoserIdx = -1
	g.actionLog = nil
	g.addLog(-1, "start", fmt.Sprintf("ピッグを開始しました（%d 人、%d 枚）",
		g.config.PlayerCnt, PigDeckSize(g.config.PlayerCnt)), nil)
	g.dealRound()
}

// dealRound は生存者へ手札を配り、新しいラウンドのパスフェーズを開始する。
func (g *Pig) dealRound() {
	g.roundNumber++
	g.phase = PigPhasePass
	g.signallerIdx = -1
	g.noticedCnt = 0
	g.passCount = 0

	active := g.activeSeats()
	g.pendingPass = make([]*Card, len(g.players))
	for _, p := range g.players {
		p.Reset()
		p.SetHasSignalled(false)
		p.SetNoticedOrder(0)
	}

	// **生存者と同じ種類のランクを 4 スート揃えて配る。** 標準 52 枚の部分集合。
	deck := NewTrumpCards(0)
	deck.Shuffle()
	ranks := pigRanksFor(len(active))
	want := make(map[int]bool, len(ranks))
	for _, r := range ranks {
		want[r] = true
	}
	byRank := make(map[int][]*Card, len(ranks))
	for {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		if want[c.GetValue()] {
			byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
		}
	}
	pool := make([]*Card, 0, len(active)*PigHandSize)
	for _, r := range ranks {
		pool = append(pool, byRank[r]...)
	}
	// スートごとに固まっているので、配る前に混ぜ直す。
	g.rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	for n := range pool {
		g.players[active[n%len(active)]].AddCard(pool[n])
	}
	g.sortActiveHands()

	g.currentPlayerIdx = active[0]
	g.addLog(-1, "deal", fmt.Sprintf("ラウンド %d を配りました（%d 人）", g.roundNumber, len(active)), nil)
}

// pigRanksFor は n 人卓で使うランクを返す。A を先頭に K から降順で詰める。
func pigRanksFor(n int) []int {
	ranks := make([]int, 0, n)
	ranks = append(ranks, 1)
	for v := 13; len(ranks) < n; v-- {
		ranks = append(ranks, v)
	}
	return ranks
}

// sortActiveHands は生存者の手札を整える。
//
// **同じランクが隣り合うように並べます。** 揃いかけが目で見て分かるかどうかが、
// このゲームでは操作性そのものです。
func (g *Pig) sortActiveHands() {
	for _, i := range g.activeSeats() {
		sortPlayerHand(g.players[i], func(ci, cj *Card) bool {
			if ci.GetValue() != cj.GetValue() {
				return ci.GetValue() < cj.GetValue()
			}
			return ci.GetDesign() < cj.GetDesign()
		})
	}
}

// activeSeats は脱落していない席の番号を返す。
func (g *Pig) activeSeats() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetEliminated() {
			out = append(out, i)
		}
	}
	return out
}

// nextActive は i の次の生存席を返す。
func (g *Pig) nextActive(i int) int {
	n := len(g.players)
	for step := 1; step <= n; step++ {
		j := (i + step) % n
		if !g.players[j].GetEliminated() {
			return j
		}
	}
	return i
}

// IsHumanTurn は現在の手番が人間かを返す。
func (g *Pig) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case PigPhasePass:
		return g.players[g.currentPlayerIdx].GetIsHuman()
	case PigPhaseSignal:
		// **合図の場面では手番が回ってきません。** 気づいた人から順に名乗るだけ。
		human := g.humanSeat()
		return human >= 0 && !g.players[human].GetEliminated() && !g.players[human].GetHasSignalled()
	case PigPhaseRoundEnd:
		// **罰の結果を読む時間。** 押されるまで次は配りません。
		return true
	default:
		return false
	}
}

// humanSeat は人間の席を返す (-1 = 無し)。
func (g *Pig) humanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// GetValidPassIndices は渡せる手札のインデックスを返す。
//
// **どの札でも渡せます。** 制限はありませんが、Web/CUI が同じ形で扱えるように
// 明示的に返します。
func (g *Pig) GetValidPassIndices(playerIdx int) []int {
	if g.phase != PigPhasePass || playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	if p.GetEliminated() || g.pendingPass[playerIdx] != nil {
		return nil
	}
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// PlayerPass は人間が渡す札を選ぶ。
func (g *Pig) PlayerPass(cardIndex int) error {
	if g.phase != PigPhasePass {
		return errors.New("いまは札を渡す場面ではありません")
	}
	idx := g.humanSeat()
	if idx < 0 || g.players[idx].GetEliminated() {
		return errors.New("あなたは脱落しています")
	}
	return g.choosePass(idx, cardIndex)
}

// PlayerSignal は人間が合図に気づいたことを伝える。
func (g *Pig) PlayerSignal() error {
	if g.phase != PigPhaseSignal {
		// **合図が出ていないのに鼻に手を当てるのは反則。** 早とちりは受けません。
		return errors.New("まだ誰も合図していません")
	}
	idx := g.humanSeat()
	if idx < 0 || g.players[idx].GetEliminated() {
		return errors.New("あなたは脱落しています")
	}
	if g.players[idx].GetHasSignalled() {
		return errors.New("すでに合図しています")
	}
	g.notice(idx)
	return nil
}

// choosePass は席 idx が渡す札を決める。
func (g *Pig) choosePass(idx, cardIndex int) error {
	p := g.players[idx]
	if g.pendingPass[idx] != nil {
		return errors.New("すでに渡す札を選んでいます")
	}
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
	}
	c := p.GetCard(cardIndex)
	if c == nil {
		return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
	}
	p.RemoveCard(cardIndex)
	g.pendingPass[idx] = c
	g.advanceChooser()
	g.resolvePassIfReady()
	return nil
}

// advanceChooser は次に選ぶ席へ進める。
func (g *Pig) advanceChooser() {
	next := g.currentPlayerIdx
	for step := 0; step < len(g.players); step++ {
		next = g.nextActive(next)
		if g.pendingPass[next] == nil {
			g.currentPlayerIdx = next
			return
		}
	}
}

// resolvePassIfReady は全員が選び終わっていれば一斉に左へ回す。
func (g *Pig) resolvePassIfReady() {
	active := g.activeSeats()
	for _, i := range active {
		if g.pendingPass[i] == nil {
			return
		}
	}

	// **同時に渡す。** 受け取る先は左隣の生存席。
	for _, i := range active {
		g.players[g.nextActive(i)].AddCard(g.pendingPass[i])
	}
	for i := range g.pendingPass {
		g.pendingPass[i] = nil
	}
	g.sortActiveHands()
	g.passCount++
	g.currentPlayerIdx = active[0]

	if idx := g.findFourOfAKind(); idx >= 0 {
		g.openSignal(idx)
		return
	}
	// **揃わないまま回り続ける配りがある。** 上限で打ち切って合図を開く。
	if g.passCount >= PigMaxPassesPerRound {
		g.openSignal(active[0])
	}
}

// findFourOfAKind は 4 枚揃えた最初の生存席を返す (-1 = 無し)。
func (g *Pig) findFourOfAKind() int {
	for _, i := range g.activeSeats() {
		if g.players[i].HasFourOfAKind() {
			return i
		}
	}
	return -1
}

// openSignal は合図フェーズを開く。揃えた本人が最初に気づいた扱いになる。
func (g *Pig) openSignal(idx int) {
	g.phase = PigPhaseSignal
	g.signallerIdx = idx
	g.noticedCnt = 0
	g.addLog(idx, "signal", "黙って手を鼻に当てました", nil)
	g.notice(idx)
}

// notice は席 idx が気づいたことを記録し、最後の 1 人が残ったら罰を確定する。
func (g *Pig) notice(idx int) {
	p := g.players[idx]
	if p.GetHasSignalled() {
		return
	}
	g.noticedCnt++
	p.SetHasSignalled(true)
	p.SetNoticedOrder(g.noticedCnt)
	if idx != g.signallerIdx {
		g.addLog(idx, "notice", fmt.Sprintf("%d 番目に気づきました", g.noticedCnt), nil)
	}

	// **最後の 1 人が残った時点で決まり。** 全員が気づくのを待ちません。
	if remaining := g.notNoticed(); len(remaining) == 1 {
		g.finishRound(remaining[0])
	}
}

// notNoticed はまだ気づいていない生存席を返す。
func (g *Pig) notNoticed() []int {
	out := make([]int, 0, len(g.players))
	for _, i := range g.activeSeats() {
		if !g.players[i].GetHasSignalled() {
			out = append(out, i)
		}
	}
	return out
}

// finishRound は最後に気づいた席へ文字を付け、ラウンドを終える。
//
// **ここで配り直しません。** 罰は 1 ラウンドに 1 回しか起きない出来事で、盤面に
// 痕跡が残らないので、すぐ配り直すと「誰が文字をもらったのか」が画面に出る前に
// 消えます。[PigPhaseRoundEnd] で止め、[Pig.NextRound] で次を配ります。
func (g *Pig) finishRound(loser int) {
	g.phase = PigPhaseRoundEnd
	g.roundLoserIdx = loser
	out := g.players[loser].AddLetter()
	g.addLog(loser, "letter", fmt.Sprintf("気づくのが最後でした（%s）",
		g.players[loser].GetLetterWord()), nil)
	if out {
		// **脱落した席は合図の記録も落とす。** 「脱落しているのに気づいた」は
		// codec が受け付けない状態で、実際そこへ落ちる経路がありました。
		g.players[loser].SetHasSignalled(false)
		g.players[loser].SetNoticedOrder(0)
		g.recountNoticed()
		g.addLog(loser, "eliminated", "PIG が揃って脱落しました", nil)
	}

	if g.checkGameEnd() {
		return
	}
	if g.roundNumber >= PigMaxRounds {
		g.finish(g.leaderIdx())
	}
}

// NextRound はラウンド終了状態から次のラウンドを配る。
func (g *Pig) NextRound() error {
	if g.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if g.phase != PigPhaseRoundEnd {
		return errors.New("いまはラウンドの区切りではありません")
	}
	g.dealRound()
	return nil
}

// recountNoticed は気づいた人数を数え直す。
func (g *Pig) recountNoticed() {
	n := 0
	for _, p := range g.players {
		if p.GetHasSignalled() {
			n++
		}
	}
	g.noticedCnt = n
}

// checkGameEnd は生存者が 1 人以下になったかを見る。
func (g *Pig) checkGameEnd() bool {
	active := g.activeSeats()
	if len(active) <= 1 {
		if len(active) == 1 {
			g.finish(active[0])
		} else {
			g.finish(g.leaderIdx())
		}
		return true
	}
	return false
}

// leaderIdx は文字のいちばん少ない席を返す (同数なら若い席)。
func (g *Pig) leaderIdx() int {
	best := 0
	for i, p := range g.players {
		if p.GetLetters() < g.players[best].GetLetters() {
			best = i
		}
	}
	return best
}

// finish は勝者を決めて終局する。
func (g *Pig) finish(winner int) {
	g.phase = PigPhaseGameEnd
	g.gameEndFlag = true
	g.winnerIdx = winner
	g.addLog(winner, "result", "最後まで残りました", nil)
}

// GiveUp は投了する。
func (g *Pig) GiveUp() {
	if g.gameEndFlag {
		return
	}
	human := g.humanSeat()
	// 人間以外で文字のいちばん少ない席を勝ちにする。
	best := -1
	for i, p := range g.players {
		if i == human {
			continue
		}
		if best < 0 || p.GetLetters() < g.players[best].GetLetters() {
			best = i
		}
	}
	if best < 0 {
		best = human
	}
	g.phase = PigPhaseGameEnd
	g.gameEndFlag = true
	g.winnerIdx = best
	g.addLog(human, "giveup", "投了しました", nil)
}

// CpuPlay は CPU の手を 1 つ進める。
func (g *Pig) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case PigPhasePass:
		idx := g.currentPlayerIdx
		if g.players[idx].GetIsHuman() || g.players[idx].GetEliminated() {
			return
		}
		_ = g.choosePass(idx, g.chooseCpuCard(idx))
	case PigPhaseSignal:
		g.cpuNotice()
	}
}

// cpuNotice は気づいていない CPU に 1 回ずつ抽選させる。
//
// **人間が黙っていると、CPU が順に気づいて最後に人間だけが残ります。**
func (g *Pig) cpuNotice() {
	miss := g.config.NoticeMissChance()
	for _, i := range g.notNoticed() {
		if g.phase != PigPhaseSignal {
			return
		}
		if g.players[i].GetIsHuman() {
			continue
		}
		if g.rng.Float64() >= miss {
			g.notice(i)
		}
	}
}

// passCandidates は手放して惜しくない札の位置を返す。
//
// **いちばん少ないランクを手放します。** 揃いかけている札は残すので、同数の
// 候補が複数出ることがあります。
func (g *Pig) passCandidates(playerIdx int) []int {
	p := g.players[playerIdx]
	counts := make(map[int]int, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			counts[c.GetValue()]++
		}
	}
	fewest := 0
	for _, n := range counts {
		if fewest == 0 || n < fewest {
			fewest = n
		}
	}
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && counts[c.GetValue()] == fewest {
			out = append(out, i)
		}
	}
	if len(out) == 0 {
		return []int{0}
	}
	return out
}

// chooseCpuCard は CPU が渡す札を選ぶ。
//
// **同数のときは無作為に選びます。決定的に選ぶと卓が固まります。** 全員が 1 枚
// ずつバラバラに持つ配りでは「いちばん少ないランク」が全員にとって同じ位置に
// なり、毎回同じ札が同じ向きに回り続けて、誰も 4 枚を揃えられません。実測では
// **ラウンドの約 40%** がこの循環に落ちて上限で打ち切られていました（3 人卓
// 603/1342）。無作為の同着崩しを入れると上限到達は 0 になります。
func (g *Pig) chooseCpuCard(playerIdx int) int {
	cand := g.passCandidates(playerIdx)
	return cand[g.rng.Intn(len(cand))]
}

// hintPassCard は人間への助言に使う札を返す。
//
// **助言は無作為に揺れてはいけない**ので、CPU と違って先頭を返します。
func (g *Pig) hintPassCard(playerIdx int) int { return g.passCandidates(playerIdx)[0] }

// GetHint は人間への助言を返す。
func (g *Pig) GetHint() *PigHint {
	if g.gameEndFlag {
		return nil
	}
	human := g.humanSeat()
	if human < 0 || g.players[human].GetEliminated() {
		return nil
	}
	switch g.phase {
	case PigPhaseSignal:
		if g.players[human].GetHasSignalled() {
			return nil
		}
		// **合図が出ています。** 気づいた時点で罰は無く、遅れることだけが負け。
		return &PigHint{Reason: "pigSignal"}
	case PigPhasePass:
		if g.pendingPass[human] != nil {
			return nil
		}
		idx := g.hintPassCard(human)
		reason := "pigDiscardOdd"
		if g.players[human].GetCardsSize() > 0 {
			if c := g.players[human].GetCard(idx); c != nil && g.countRank(human, c.GetValue()) >= 2 {
				// 全部が同じ枚数のときは、崩しても被害の小さい札になる。
				reason = "pigNoSingleton"
			}
		}
		return &PigHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// countRank は席 idx の手札に含まれる rank の枚数を返す。
func (g *Pig) countRank(idx, rank int) int {
	n := 0
	p := g.players[idx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetValue() == rank {
			n++
		}
	}
	return n
}

// addLog は棋譜に 1 行足す。
func (g *Pig) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// GetConfig は設定を返す。
func (g *Pig) GetConfig() PigConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Pig) SetConfig(cfg PigConfig) {
	if err := cfg.Validate(); err != nil {
		return
	}
	if cfg.PlayerCnt != g.config.PlayerCnt {
		g.players = newPigSeats(cfg.PlayerCnt)
	}
	g.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (g *Pig) GetPhase() PigPhase { return g.phase }

// GetGameEndFlag は終局フラグを返す。
func (g *Pig) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayerCnt は人数を返す。
func (g *Pig) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (g *Pig) GetPlayer(i int) *PigPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetCurrentPlayerIdx は渡す札を選ぶ番の席を返す。
func (g *Pig) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetSignallerIdx は最初に合図した席を返す (-1 = 合図なし)。
func (g *Pig) GetSignallerIdx() int { return g.signallerIdx }

// GetNoticedCnt は合図に気づいた人数を返す。
func (g *Pig) GetNoticedCnt() int { return g.noticedCnt }

// GetRoundLoserIdx は直近ラウンドで文字が付いた席を返す (-1 = 未)。
func (g *Pig) GetRoundLoserIdx() int { return g.roundLoserIdx }

// GetRoundNumber はラウンド数を返す。
func (g *Pig) GetRoundNumber() int { return g.roundNumber }

// GetPassCount は当該ラウンドのパス回数を返す。
func (g *Pig) GetPassCount() int { return g.passCount }

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (g *Pig) GetWinnerIdx() int { return g.winnerIdx }

// HasChosenPass は席 i が渡す札を選び終えたかを返す。
func (g *Pig) HasChosenPass(i int) bool {
	if i < 0 || i >= len(g.pendingPass) {
		return false
	}
	return g.pendingPass[i] != nil
}

// GetDeckSize はこの卓で使うデッキ枚数を返す。
func (g *Pig) GetDeckSize() int { return PigDeckSize(len(g.activeSeats())) }

// pigJSON is the JSON wire format for Pig.
type pigJSON struct {
	Players       []*PigPlayer      `json:"pl"`
	Config        PigConfig         `json:"cf"`
	Phase         PigPhase          `json:"ph"`
	PendingPass   []*Card           `json:"pp"`
	CurrentIdx    int               `json:"ci"`
	SignallerIdx  int               `json:"si"`
	NoticedCnt    int               `json:"nc"`
	RoundLoserIdx int               `json:"rl"`
	PassCount     int               `json:"pc"`
	RoundNumber   int               `json:"rn"`
	GameEndFlag   bool              `json:"ge"`
	WinnerIdx     int               `json:"wi"`
	ActionLog     []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Pig) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigJSON{
		Players:       g.players,
		Config:        g.config,
		Phase:         g.phase,
		PendingPass:   g.pendingPass,
		CurrentIdx:    g.currentPlayerIdx,
		SignallerIdx:  g.signallerIdx,
		NoticedCnt:    g.noticedCnt,
		RoundLoserIdx: g.roundLoserIdx,
		PassCount:     g.passCount,
		RoundNumber:   g.roundNumber,
		GameEndFlag:   g.gameEndFlag,
		WinnerIdx:     g.winnerIdx,
		ActionLog:     g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Pig) UnmarshalJSON(data []byte) error {
	var j pigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("seat count %d does not match the configured %d", len(j.Players), j.Config.PlayerCnt)
	}
	if j.Phase < PigPhaseMin || j.Phase > PigPhaseMax {
		return fmt.Errorf("phase out of range: %d", j.Phase)
	}
	if len(j.PendingPass) != 0 && len(j.PendingPass) != len(j.Players) {
		return fmt.Errorf("pending pass has %d slots for %d seats", len(j.PendingPass), len(j.Players))
	}
	if len(j.ActionLog) > pigMaxSliceLen {
		return fmt.Errorf("action log too long: %d", len(j.ActionLog))
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= len(j.Players) {
		return fmt.Errorf("current player index out of range: %d", j.CurrentIdx)
	}
	if j.SignallerIdx < -1 || j.SignallerIdx >= len(j.Players) {
		return fmt.Errorf("signaller index out of range: %d", j.SignallerIdx)
	}
	if j.RoundLoserIdx < -1 || j.RoundLoserIdx >= len(j.Players) {
		return fmt.Errorf("round loser index out of range: %d", j.RoundLoserIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= len(j.Players) {
		return fmt.Errorf("winner index out of range: %d", j.WinnerIdx)
	}
	if j.GameEndFlag != (j.Phase == PigPhaseGameEnd) {
		return fmt.Errorf("the game-end flag and the phase disagree (flag=%v, phase=%d)", j.GameEndFlag, j.Phase)
	}
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("a finished game has a winner and an unfinished one does not (flag=%v, winner=%d)",
			j.GameEndFlag, j.WinnerIdx)
	}
	if j.PassCount < 0 || j.PassCount > PigMaxPassesPerRound {
		return fmt.Errorf("pass count out of range: %d", j.PassCount)
	}
	if j.RoundNumber < 0 || j.RoundNumber > PigMaxRounds {
		return fmt.Errorf("round number out of range: %d", j.RoundNumber)
	}

	// **合図の有無とフェーズは同じ事実の裏表。**
	if (j.Phase == PigPhaseSignal) && j.SignallerIdx < 0 {
		return errors.New("the signal phase needs a seat that signalled")
	}
	if j.Phase == PigPhasePass && j.SignallerIdx >= 0 {
		return errors.New("nobody has signalled while cards are still being passed")
	}
	// **ラウンドの区切りは、誰かが文字をもらったという事実そのもの。**
	if j.Phase == PigPhaseRoundEnd && j.RoundLoserIdx < 0 {
		return errors.New("a finished round needs the seat that took the letter")
	}

	noticed, active := 0, 0
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("seat %d is missing", i)
		}
		if !p.GetEliminated() {
			active++
		}
		if p.GetHasSignalled() {
			noticed++
		}
		// **脱落した席は合図にも加わりません。**
		if p.GetEliminated() && p.GetHasSignalled() {
			return fmt.Errorf("seat %d is out but is recorded as having signalled", i)
		}
	}
	if j.NoticedCnt != noticed {
		return fmt.Errorf("noticed count %d does not match the %d seats that signalled", j.NoticedCnt, noticed)
	}
	// **気づいた人が居るのは合図が出たあとだけ。**
	if noticed > 0 && j.SignallerIdx < 0 {
		return fmt.Errorf("%d seats signalled but no seat is recorded as the signaller", noticed)
	}
	if j.Phase != PigPhaseGameEnd && active < 2 {
		return fmt.Errorf("a game in progress needs at least 2 seats, found %d", active)
	}

	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.pendingPass = j.PendingPass
	if len(g.pendingPass) == 0 {
		g.pendingPass = make([]*Card, len(j.Players))
	}
	g.currentPlayerIdx = j.CurrentIdx
	g.signallerIdx = j.SignallerIdx
	g.noticedCnt = j.NoticedCnt
	g.roundLoserIdx = j.RoundLoserIdx
	g.passCount = j.PassCount
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
