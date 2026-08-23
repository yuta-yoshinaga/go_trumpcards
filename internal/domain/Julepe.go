//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// JulepePhase フレペのゲームフェーズ
type JulepePhase int

// Julepeのフェーズ定数
const (
	// JulepePhaseDecide 参加するか降りるかの選択中
	JulepePhaseDecide JulepePhase = iota
	// JulepePhasePlay プレイ中
	JulepePhasePlay
	// JulepePhaseRoundEnd ラウンド終了（配当確定、次ラウンド待ち）
	JulepePhaseRoundEnd
	// JulepePhaseGameEnd ゲーム終了
	JulepePhaseGameEnd
)

// JulepeHandSize 各プレイヤーの手札枚数
const JulepeHandSize = 5

// JulepeTricksPerRound 1 ラウンドのトリック数
const JulepeTricksPerRound = JulepeHandSize

// JulepeStartingChips 開始時の持ちチップ
const JulepeStartingChips = 60

// JulepeAnte 各ラウンドの開始時に全員がポットへ出すチップ
const JulepeAnte = 3

// JulepeMissPenalty 規定トリック数に届かなかったときの追加支払い
const JulepeMissPenalty = 5

// GetRequiredTricks は現在の参加人数に対する規定トリック数を返す。
//
// 選択フェーズがまだ終わっていない (誰も参加を決めていない) 場合は
// 席数で計算する ── 画面は決める前からこの数を出す必要がある。
func (r *Julepe) GetRequiredTricks() int {
	n := 0
	for _, p := range r.players {
		if p.GetInRound() {
			n++
		}
	}
	if n == 0 {
		n = len(r.players)
	}
	return JulepeRequiredTricks(n)
}

// GetBeast は次ラウンドのアンティが倍になる席を返す。
func (r *Julepe) GetBeast() []bool { return r.beast }

// GetBeastForTest は beast フラグを返す (テスト用)。
func (r *Julepe) GetBeastForTest() []bool { return r.beast }

// DealRoundForTest は次ラウンドの配りを実行する (テスト用)。
func (r *Julepe) DealRoundForTest() { r.dealRound() }

// JulepeRequiredTricks は参加人数 n に対する規定トリック数を返す。
//
// **人数で変わる。** クローン元のラムスは「1 トリックも取れなければ罰」
// という固定の線だったが、こちらは参加者で山分けした取り分に届いたかを見る。
// 5 トリックを n 人で分けた切り上げ (最低 1)。
//
//	n=2 → 3 / n=3 → 2 / n=4 → 2 / n=5 → 1
//
// **手で並べた表ではなく手札枚数から導く**ので、配り枚数を変えれば追随する。
//
// 参加者が 1 人しかいないラウンドは競争が無い ── 他が全員降りた以上、
// その 1 人はポットを取る側なので、規定は 1 トリックに落とす
// (全トリックを要求すると、技術的な理由だけで beast にできてしまう)。
func JulepeRequiredTricks(n int) int {
	if n <= 1 {
		return 1
	}
	req := (JulepeTricksPerRound + n - 1) / n
	if req < 1 {
		return 1
	}
	return req
}

// julepeMaxSliceLen caps slice sizes during deserialisation.
const julepeMaxSliceLen = 1000

// Julepe フレペ ゲームクラス。
//
// ドイツ・フランスで遊ばれる**ポット式**トリックテイキング。ピケット 32 枚
// (A,7,8,9,10,J,Q,K × 4 スート) を使い、**3〜5 人**という可変人数で遊ぶ。
// 各プレイヤーに **5 枚**配り、次の 1 枚を表向きにして**そのスートを切り札**
// とする。
//
// 手順:
//
//  1. 全員がポットへアンティを出す
//  2. 手札を見て**参加する (play) か降りる (pass) か**を選ぶ
//  3. 参加者だけでフォロー義務のあるトリックテイキングを行う
//  4. **参加したのに 1 トリックも取れなかった人はポットへ追加支払い**
//  5. トリックを取った人は、獲得数に応じてポットを分ける
//
// 降りれば失うのはアンティだけ。参加して 0 トリックだと余分に払う。**弱い手で
// 参加する価値があるかどうか**が唯一の判断で、それがこのゲームの全部になる。
//
// issue #5236 の仕様案に対して、次の 2 点を補って実装した:
//
//   - **手札の枚数が書かれていない。** フレペは 5 枚が標準で、流用元に指定された
//     Loo も 5 枚 (`LooHandSize`)。5 人でも 25 枚しか要らず 32 枚に収まる。
//   - **切り札の規定が無い。** フレペは配り切ったあとの 1 枚を表向きにして切り札を
//     決める。切り札が無いと、降りるか参加するかを手札から判断する材料が乏しくなり、
//     このゲームの唯一の判断が成立しない。
//
// Loo との違いは**人数が固定でない**ことで、そのため Loo が
// `[LooPlayerCnt]bool` の固定長配列で持っている状態はすべてスライスにしている。
type Julepe struct {
	trumpCards *TrumpCards
	players    []*JulepePlayer
	config     JulepeConfig

	phase       JulepePhase
	roundNumber int
	trickNumber int
	// pot 場に積まれたチップ
	pot int
	// beast は規定トリック数に届かなかった席。**次ラウンドのアンティが倍**に
	// なるので、ラウンドをまたいで持ち越す盤面の一部。
	beast []bool
	// trumpSuit 表向きにした 1 枚で決まる切り札
	trumpSuit int
	// upCard 切り札を決めた表向きの 1 枚
	upCard *Card

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewJulepe コンストラクタ
func NewJulepe(trumpCards *TrumpCards, players []*JulepePlayer, config JulepeConfig) *Julepe {
	return &Julepe{trumpCards: trumpCards, players: players, config: config, winnerIdx: -1}
}

// NewDefaultJulepe 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultJulepe() *Julepe {
	cfg := DefaultJulepeConfig()
	return NewJulepe(NewTrumpCardsBelote(), newJulepePlayers(cfg.PlayerCnt), cfg)
}

// newJulepePlayers 指定人数のプレイヤーを作る（0 番だけが人間）
func newJulepePlayers(n int) []*JulepePlayer {
	players := make([]*JulepePlayer, 0, n)
	for i := range n {
		players = append(players, NewJulepePlayer(i == 0))
	}
	return players
}

// Reset ゲーム全体を初期化する
func (r *Julepe) Reset() {
	// 人数設定が変わっていれば席を作り直す。
	if len(r.players) != r.config.PlayerCnt {
		r.players = newJulepePlayers(r.config.PlayerCnt)
	}
	r.roundNumber = 1
	r.dealerIdx = 0
	r.gameEndFlag = false
	r.winnerIdx = -1
	r.pot = 0
	r.actionLog = nil
	for _, p := range r.players {
		p.ResetGame()
	}
	r.dealRound()
}

// dealRound 1 ラウンド分を配り、アンティを集めて選択フェーズに入る
func (r *Julepe) dealRound() {
	r.phase = JulepePhaseDecide
	r.trickNumber = 0
	r.currentTrick = nil
	for i, p := range r.players {
		p.ResetRound()
		// **beast は倍払い。** 前ラウンドで規定トリック数に届かなかった席。
		ante := JulepeAnte
		if i < len(r.beast) && r.beast[i] {
			ante *= 2
			r.appendLog(i, "beast_ante",
				fmt.Sprintf("beast のためアンティ倍払い (%d)", ante), nil)
		}
		p.AddChips(-ante)
		r.pot += ante
	}

	// **倍払いは 1 回きり。** 集め終えたら消す ── 消し忘れると beast が
	// 永久に倍払いを続ける。
	r.beast = nil

	r.trumpCards = NewTrumpCardsBelote()
	r.trumpCards.Shuffle()
	for range JulepeHandSize {
		for i := range len(r.players) {
			idx := (r.dealerIdx + 1 + i) % len(r.players)
			if c := r.trumpCards.DrawCard(); c != nil {
				r.players[idx].AddCard(c)
			}
		}
	}
	// **配り切ったあとの 1 枚が切り札を決める。**
	r.upCard = r.trumpCards.DrawCard()
	if r.upCard != nil {
		r.trumpSuit = r.upCard.GetDesign()
	}
	r.leadPlayerIdx = (r.dealerIdx + 1) % len(r.players)
	r.currentPlayerIdx = r.leadPlayerIdx
	r.sortAllHands()
	r.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d 開始（ポット %d）", r.roundNumber, r.pot), nil)
}

// sortAllHands 手札を切り札を最後に、スートごと・強さ順に並べる
func (r *Julepe) sortAllHands() {
	for _, p := range r.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			iTrump := ci.GetDesign() == r.trumpSuit
			jTrump := cj.GetDesign() == r.trumpSuit
			if iTrump != jTrump {
				return !iTrump
			}
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return julepeRank(ci) < julepeRank(cj)
		})
	}
}

// Decide 人間プレイヤーが参加 (play) するか降りる (pass) かを選ぶ
func (r *Julepe) Decide(play bool) error {
	if r.gameEndFlag {
		return errors.New("game has ended")
	}
	if r.phase != JulepePhaseDecide {
		return errors.New("not the decision phase")
	}
	if r.players[0].GetDecided() {
		return errors.New("already decided")
	}
	r.setDecision(0, play)
	r.cpuDecide()
	r.startPlayIfReady()
	return nil
}

// setDecision 1 人分の選択を記録する
func (r *Julepe) setDecision(idx int, play bool) {
	p := r.players[idx]
	p.SetDecided(true)
	p.SetInRound(play)
	action := "pass"
	detail := "降りた"
	if play {
		action, detail = "play", "参加した"
	}
	r.appendLog(idx, action, detail, nil)
}

// cpuDecide 未決定の CPU に選択させる
func (r *Julepe) cpuDecide() {
	for i := 1; i < len(r.players); i++ {
		if !r.players[i].GetDecided() {
			r.setDecision(i, r.cpuWantsToPlay(i))
		}
	}
}

// cpuWantsToPlay CPU の参加判断。**切り札の高い札があるかどうか**で決める。
// 1 トリックも取れないと追加で払わされるので、取れる見込みが要る。
func (r *Julepe) cpuWantsToPlay(idx int) bool {
	p := r.players[idx]
	strength := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		switch {
		case c.GetDesign() == r.trumpSuit && julepeRank(c) >= CardValueMax:
			strength += 3 // 切り札の A / K
		case c.GetDesign() == r.trumpSuit:
			strength++
		case julepeRank(c) > CardValueMax:
			strength += 2 // 平のエース
		}
	}
	return strength >= 3
}

// startPlayIfReady 全員が選び終えていればプレイに入る
func (r *Julepe) startPlayIfReady() {
	for _, p := range r.players {
		if !p.GetDecided() {
			return
		}
	}
	if r.activeCount() == 0 {
		// 全員降りた。ポットは次ラウンドへ持ち越す。
		r.appendLog(-1, "allpass", fmt.Sprintf("全員降りた。ポット %d を持ち越し", r.pot), nil)
		r.finishRound()
		return
	}
	r.phase = JulepePhasePlay
	r.leadPlayerIdx = r.firstActiveFrom(r.leadPlayerIdx)
	r.currentPlayerIdx = r.leadPlayerIdx
}

// activeCount 参加者の数
func (r *Julepe) activeCount() int {
	n := 0
	for _, p := range r.players {
		if p.GetInRound() {
			n++
		}
	}
	return n
}

// firstActiveFrom idx から時計回りに最初の参加者を返す
func (r *Julepe) firstActiveFrom(idx int) int {
	for i := range len(r.players) {
		cand := (idx + i) % len(r.players)
		if r.players[cand].GetInRound() {
			return cand
		}
	}
	return idx
}

// nextActiveAfter idx の次の参加者を返す
func (r *Julepe) nextActiveAfter(idx int) int {
	return r.firstActiveFrom((idx + 1) % len(r.players))
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (r *Julepe) PlayerPlay(cardIndex int) error {
	if r.gameEndFlag {
		return errors.New("game has ended")
	}
	if r.phase != JulepePhasePlay {
		return errors.New("not the play phase")
	}
	if r.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	if !r.players[0].GetInRound() {
		return errors.New("you passed this round")
	}
	return r.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (r *Julepe) CpuPlay() {
	if r.gameEndFlag || r.phase != JulepePhasePlay || r.currentPlayerIdx == 0 {
		return
	}
	_ = r.play(r.currentPlayerIdx, r.chooseCpuCard(r.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (r *Julepe) play(playerIdx, cardIndex int) error {
	p := r.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !r.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	r.currentTrick = append(r.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	r.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(r.currentTrick) < r.activeCount() {
		r.currentPlayerIdx = r.nextActiveAfter(playerIdx)
		return nil
	}
	r.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (r *Julepe) canPlay(playerIdx int, card *Card) bool {
	if len(r.currentTrick) == 0 {
		return true
	}
	leadSuit := r.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := r.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (r *Julepe) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(r.players) {
		return nil
	}
	p := r.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if r.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決する
func (r *Julepe) resolveTrick() {
	winner := r.trickWinner()
	cards := make([]*Card, 0, len(r.currentTrick))
	for _, tc := range r.currentTrick {
		cards = append(cards, tc.Card)
	}
	r.players[winner].AddTrick(cards)
	r.players[winner].AddRoundTrick()

	r.trickNumber++
	r.currentTrick = nil
	r.leadPlayerIdx = winner
	r.currentPlayerIdx = winner

	if r.trickNumber >= JulepeTricksPerRound {
		r.finishRound()
	}
}

// finishRound ポットを精算する。
//
// **1 トリックも取れなかった参加者は追加で払う。** 取った人はポットを
// 獲得トリック数で按分する。割り切れないぶんは次ラウンドへ残す
// （チップを勝手に増やさないため）。
func (r *Julepe) finishRound() {
	// **規定トリック数に届かなかった参加者が beast。** クローン元のラムスは
	// 「0 トリック」だけを罰したが、こちらは人数で決まる線に届いたかで見る。
	// beast は次ラウンドのアンティが倍になるので、ここでフラグを立てる。
	participants := 0
	for _, p := range r.players {
		if p.GetInRound() {
			participants++
		}
	}
	required := JulepeRequiredTricks(participants)
	r.beast = make([]bool, len(r.players))
	for i, p := range r.players {
		if p.GetInRound() && p.GetRoundTricks() < required {
			p.AddChips(-JulepeMissPenalty)
			r.pot += JulepeMissPenalty
			r.beast[i] = true
			r.appendLog(i, "penalty",
				fmt.Sprintf("規定 %d トリックに届かず (%d) — beast、%d 支払い",
					required, p.GetRoundTricks(), JulepeMissPenalty), nil)
		}
	}

	totalTricks := 0
	for _, p := range r.players {
		totalTricks += p.GetRoundTricks()
	}
	if totalTricks > 0 {
		share := r.pot / totalTricks
		paid := 0
		for i, p := range r.players {
			if n := p.GetRoundTricks(); n > 0 {
				amount := share * n
				p.AddChips(amount)
				paid += amount
				r.appendLog(i, "payout", fmt.Sprintf("%d トリックで %d 獲得", n, amount), nil)
			}
		}
		// **端数は持ち越す。** 配り切れないぶんを消すとチップが減る。
		r.pot -= paid
	}

	if r.roundNumber >= r.config.Rounds {
		// **最終ラウンドには持ち越し先が無い。** ここで空にしておかないと、
		// 端数や全員降りたぶんのチップが盤上に取り残される。総量としては
		// 保たれていても、勝敗を決める GetChips() には入らない。
		r.drainPot()
		r.finishGame()
		return
	}
	r.phase = JulepePhaseRoundEnd
}

// drainPot 最終ラウンドで残ったポットを配り切る。
//
// 受け取るのは**そのラウンドでトリックを取った人**。誰も取っていなければ
// （全員降りた場合）全員で分ける。端数はディーラーの左隣から 1 枚ずつ配るので、
// **1 チップも盤上に残らない**。
func (r *Julepe) drainPot() {
	if r.pot <= 0 {
		return
	}
	recipients := make([]int, 0, len(r.players))
	// ディーラーの左隣から見るので、端数の配り先も決定的になる。
	for i := range len(r.players) {
		idx := (r.dealerIdx + 1 + i) % len(r.players)
		if r.players[idx].GetRoundTricks() > 0 {
			recipients = append(recipients, idx)
		}
	}
	if len(recipients) == 0 {
		for i := range len(r.players) {
			recipients = append(recipients, (r.dealerIdx+1+i)%len(r.players))
		}
	}

	share, remainder := r.pot/len(recipients), r.pot%len(recipients)
	for n, idx := range recipients {
		amount := share
		if n < remainder {
			amount++
		}
		r.players[idx].AddChips(amount)
	}
	r.appendLog(-1, "pot", fmt.Sprintf("最終ラウンド。残りポット %d を %d 人で分配", r.pot, len(recipients)), nil)
	r.pot = 0
}

// NextRound 次のラウンドを開始する。持ち越したポットはそのまま残る。
func (r *Julepe) NextRound() {
	if r.gameEndFlag || r.phase != JulepePhaseRoundEnd {
		return
	}
	r.roundNumber++
	r.dealerIdx = (r.dealerIdx + 1) % len(r.players)
	r.dealRound()
}

// finishGame チップが最も多いプレイヤーの勝ち
func (r *Julepe) finishGame() {
	r.phase = JulepePhaseGameEnd
	r.gameEndFlag = true

	best, bestIdx, tied := r.players[0].GetChips(), 0, false
	for i := 1; i < len(r.players); i++ {
		switch chips := r.players[i].GetChips(); {
		case chips > best:
			best, bestIdx, tied = chips, i, false
		case chips == best:
			tied = true
		}
	}
	if tied {
		r.winnerIdx = -1
		r.appendLog(-1, "result", "同点で決着つかず", nil)
		return
	}
	r.winnerIdx = bestIdx
	r.appendLog(bestIdx, "result", fmt.Sprintf("勝者（%dチップ）", best), nil)
}

// trickWinner 現在のトリックの勝者。切り札が最強、次いでリードのスート。
func (r *Julepe) trickWinner() int {
	if len(r.currentTrick) == 0 {
		return r.leadPlayerIdx
	}
	leadSuit := r.currentTrick[0].Card.GetDesign()
	bestIdx, best := r.currentTrick[0].PlayerIdx, r.currentTrick[0].Card
	for _, tc := range r.currentTrick[1:] {
		if julepeBeats(tc.Card, best, leadSuit, r.trumpSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// julepeBeats challenger が currentBest に勝つか
func julepeBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cTrump := challenger.GetDesign() == trumpSuit
	bTrump := currentBest.GetDesign() == trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if cTrump && bTrump {
		return julepeRank(challenger) > julepeRank(currentBest)
	}
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return julepeRank(challenger) > julepeRank(currentBest)
}

// julepeRank 札の強さ。A が最強、7 が最弱。
func julepeRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// chooseCpuCard CPU の手を選ぶ。**取りに行くゲーム**なので、取れるなら取る。
func (r *Julepe) chooseCpuCard(playerIdx int) int {
	valid := r.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := r.players[playerIdx]

	if len(r.currentTrick) == 0 {
		// リードは一番強い札で。1 トリック取れば追加支払いを避けられる。
		bestIdx, bestRank := valid[0], julepeRank(p.GetCard(valid[0]))
		for _, i := range valid[1:] {
			if rank := julepeRank(p.GetCard(i)); rank > bestRank {
				bestIdx, bestRank = i, rank
			}
		}
		return bestIdx
	}

	// 取れるなら一番安く取る。取れないなら一番弱い札を捨てる。
	if idx, ok := r.pickWinning(p, valid); ok {
		return idx
	}
	bestIdx, bestRank := valid[0], julepeRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if rank := julepeRank(p.GetCard(i)); rank < bestRank {
			bestIdx, bestRank = i, rank
		}
	}
	return bestIdx
}

// pickWinning トリックを取れる札のうち一番安いもの
func (r *Julepe) pickWinning(p *JulepePlayer, valid []int) (int, bool) {
	leadSuit := r.currentTrick[0].Card.GetDesign()
	best := r.currentTrick[0].Card
	for _, tc := range r.currentTrick[1:] {
		if julepeBeats(tc.Card, best, leadSuit, r.trumpSuit) {
			best = tc.Card
		}
	}
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !julepeBeats(c, best, leadSuit, r.trumpSuit) {
			continue
		}
		if rank := julepeRank(c); bestIdx < 0 || rank < bestRank {
			bestIdx, bestRank = i, rank
		}
	}
	return bestIdx, bestIdx >= 0
}

// JulepeHint ヒント情報
type JulepeHint struct {
	// CardIndex 推奨する手札のインデックス（選択フェーズでは nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す。手番でなければ nil。
func (r *Julepe) GetHint() *JulepeHint {
	if r.gameEndFlag {
		return nil
	}
	// **選択フェーズでは出す札ではなく、参加するかどうかを助言する。**
	if r.phase == JulepePhaseDecide && !r.players[0].GetDecided() {
		if r.cpuWantsToPlay(0) {
			return &JulepeHint{Reason: "julepePlayIn"}
		}
		return &JulepeHint{Reason: "julepePassOut"}
	}
	if !r.IsHumanTurn() || r.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := r.chooseCpuCard(0)
	reason := "julepeTakeTrick"
	if r.players[0].GetRoundTricks() > 0 {
		reason = "julepeAlreadySafe"
	}
	return &JulepeHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (r *Julepe) GetPhase() JulepePhase { return r.phase }

// GetConfig 現在の設定
func (r *Julepe) GetConfig() JulepeConfig { return r.config }

// SetConfig 設定を差し替える
func (r *Julepe) SetConfig(c JulepeConfig) { r.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (r *Julepe) GetRoundNumber() int { return r.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (r *Julepe) GetTrickNumber() int { return r.trickNumber }

// GetPot 場に積まれているチップ
func (r *Julepe) GetPot() int { return r.pot }

// GetTrumpSuit 切り札のスート
func (r *Julepe) GetTrumpSuit() int { return r.trumpSuit }

// GetUpCard 切り札を決めた表向きの 1 枚
func (r *Julepe) GetUpCard() *Card { return r.upCard }

// GetCurrentTrick 現在のトリック
func (r *Julepe) GetCurrentTrick() []*TrickCard { return r.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (r *Julepe) GetCurrentPlayerIdx() int { return r.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (r *Julepe) GetLeadPlayerIdx() int { return r.leadPlayerIdx }

// GetDealerIdx ディーラー
func (r *Julepe) GetDealerIdx() int { return r.dealerIdx }

// GetActiveCount このラウンドの参加者数
func (r *Julepe) GetActiveCount() int { return r.activeCount() }

// GetPlayerCnt プレイヤー数
func (r *Julepe) GetPlayerCnt() int { return len(r.players) }

// GetPlayer 指定インデックスのプレイヤー
func (r *Julepe) GetPlayer(i int) *JulepePlayer {
	if i < 0 || i >= len(r.players) {
		return nil
	}
	return r.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (r *Julepe) GetGameEndFlag() bool { return r.gameEndFlag }

// GetWinnerIdx 勝者（-1: 未確定または同点）
func (r *Julepe) GetWinnerIdx() int { return r.winnerIdx }

// IsHumanTurn 人間の手番か
func (r *Julepe) IsHumanTurn() bool {
	return !r.gameEndFlag && r.phase == JulepePhasePlay && r.currentPlayerIdx == 0 && r.players[0].GetInRound()
}

// IsDecidePhase 参加するかどうかの選択中か
func (r *Julepe) IsDecidePhase() bool {
	return !r.gameEndFlag && r.phase == JulepePhaseDecide && !r.players[0].GetDecided()
}

// GiveUp 投了する
func (r *Julepe) GiveUp() {
	if r.gameEndFlag {
		return
	}
	r.phase = JulepePhaseGameEnd
	r.gameEndFlag = true
	r.winnerIdx = -1
	r.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (r *Julepe) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	r.appendLogAt(r.trickNumber, playerIdx, actionType, detail, cards)
}

// julepeJSON is the KV snapshot format for Julepe.
type julepeJSON struct {
	TrumpCards  *TrumpCards     `json:"tc"`
	Players     []*JulepePlayer `json:"pl"`
	Config      JulepeConfig    `json:"cf"`
	Phase       JulepePhase     `json:"ph"`
	RoundNumber int             `json:"rn"`
	TrickNumber int             `json:"tn"`
	Pot         int             `json:"po"`
	// **beast はラウンドをまたぐ。** 落とすと復元後に倍払いが消え、
	// 規定に届かなかった席が普通のアンティで済んでしまう。
	Beast            []bool            `json:"bs"`
	TrumpSuit        int               `json:"ts"`
	UpCard           *Card             `json:"uc"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	CurrentPlayerIdx int               `json:"cp"`
	LeadPlayerIdx    int               `json:"lp"`
	DealerIdx        int               `json:"di"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (r *Julepe) MarshalJSON() ([]byte, error) {
	return json.Marshal(&julepeJSON{
		TrumpCards:       r.trumpCards,
		Players:          r.players,
		Config:           r.config,
		Phase:            r.phase,
		RoundNumber:      r.roundNumber,
		TrickNumber:      r.trickNumber,
		Pot:              r.pot,
		Beast:            r.beast,
		TrumpSuit:        r.trumpSuit,
		UpCard:           r.upCard,
		CurrentTrick:     r.currentTrick,
		CurrentPlayerIdx: r.currentPlayerIdx,
		LeadPlayerIdx:    r.leadPlayerIdx,
		DealerIdx:        r.dealerIdx,
		GameEndFlag:      r.gameEndFlag,
		WinnerIdx:        r.winnerIdx,
		ActionLog:        r.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた
// 任意のバイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (r *Julepe) UnmarshalJSON(data []byte) error {
	var j julepeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < JulepePhaseDecide || j.Phase > JulepePhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > JulepeTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 || j.RoundNumber > JulepeRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.Pot < 0 {
		return fmt.Errorf("invalid pot: %d", j.Pot)
	}
	// **設定そのものも検証する。** Rounds が 0 の壊れた blob を通すと、
	// finishRound() の `roundNumber >= config.Rounds` が 1 ラウンド目で真になり、
	// 配った直後にゲームが終わる。Loo.UnmarshalJSON も同じ検証をしている。
	if err := j.Config.Validate(); err != nil {
		return err
	}
	// **人数は可変なので、まず席数を検証してから席番号を見る。**
	n := len(j.Players)
	if n < JulepePlayerCntMin || n > JulepePlayerCntMax {
		return fmt.Errorf("invalid player count: %d", n)
	}
	// 設定の人数と実際の席数が食い違う blob も拒否する。
	if j.Config.PlayerCnt != n {
		return fmt.Errorf("config says %d players but %d seats were stored", j.Config.PlayerCnt, n)
	}
	if len(j.ActionLog) > julepeMaxSliceLen {
		return errors.New("julepe: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > n {
		return fmt.Errorf("current trick holds %d cards for %d players", len(j.CurrentTrick), n)
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= n {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		r.trumpCards = j.TrumpCards
	}
	r.players = j.Players
	r.config = j.Config
	r.phase = j.Phase
	r.roundNumber = j.RoundNumber
	r.trickNumber = j.TrickNumber
	r.pot = j.Pot
	r.beast = j.Beast
	r.trumpSuit = j.TrumpSuit
	r.upCard = j.UpCard
	r.currentTrick = j.CurrentTrick
	r.currentPlayerIdx = j.CurrentPlayerIdx
	r.leadPlayerIdx = j.LeadPlayerIdx
	r.dealerIdx = j.DealerIdx
	r.gameEndFlag = j.GameEndFlag
	r.winnerIdx = j.WinnerIdx
	r.actionLog = j.ActionLog
	return nil
}
