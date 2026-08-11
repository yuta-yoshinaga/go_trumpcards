//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PolignacPhase ポリニャックのゲームフェーズ
type PolignacPhase int

// Polignacのフェーズ定数
const (
	// PolignacPhaseDeclare capot 宣言の受付中
	PolignacPhaseDeclare PolignacPhase = iota
	// PolignacPhasePlay プレイ中
	PolignacPhasePlay
	// PolignacPhaseRoundEnd ラウンド終了（失点確定、次ラウンド待ち）
	PolignacPhaseRoundEnd
	// PolignacPhaseGameEnd ゲーム終了
	PolignacPhaseGameEnd
)

// PolignacPlayerCnt プレイヤー数（4 人固定）
const PolignacPlayerCnt = 4

// PolignacHandSize 各プレイヤーの手札枚数
const PolignacHandSize = 8

// PolignacTricksPerRound 1 ラウンドのトリック数
const PolignacTricksPerRound = PolignacHandSize

// PolignacJackValue ジャックの値
const PolignacJackValue = 11

// PolignacJackPenalty ジャック 1 枚あたりの失点
const PolignacJackPenalty = 1

// PolignacSpadeJackPenalty スペードのジャック（Polignac）の失点。**他の 2 倍。**
const PolignacSpadeJackPenalty = 2

// PolignacCapotStake capot の成否で動く点数
const PolignacCapotStake = 5

// polignacMaxSliceLen caps slice sizes during deserialisation.
const polignacMaxSliceLen = 1000

// Polignac ポリニャック ゲームクラス。
//
// フランス発祥の**回避型**トリックテイキング。別名 Quatre Valets（4 人のジャック）。
// ピケット 32 枚 (A,7,8,9,10,J,Q,K × 4 スート) を 4 人に 8 枚ずつ配り、8 トリックを
// 戦う。**切り札は無い。**
//
// 回避対象は**ジャック 4 枚だけ**で、取ると失点する:
//
//   - ジャック 1 枚につき **1 点**
//   - ただし**スペードのジャック（Polignac）だけは 2 点**
//
// つまり 1 ラウンドで動く失点は最大 5 点。Hearts と違い危険な札が 4 枚しかないので、
// 「どの 1 枚が誰の手にあるか」の読み合いになる。
//
// **失点は正の数で数え、合計が最も小さいプレイヤーが勝つ。** スロバーハンネスとは
// 符号の向きが逆だが、これは意図的で、それぞれの実際の記譜法に合わせている
// （issue #5234 も「合計失点が最も少ないプレイヤーが勝利」と書いている）。
//
// issue #5234 の仕様案とは 1 点異なり、実際の規則に合わせた:
//   - **capot は「全 8 トリックを取る」宣言。** issue は「全 4 枚のジャックを取る」
//     と書いているが、capot はフランスのカードゲームで「全トリック獲得」を指す語で、
//     ポリニャックでも同じ。ジャック 4 枚だけを狙う宣言にすると、ジャックの無い
//     トリックを捨てられるぶん**桁違いに簡単**になり、賭けとして成立しない。
//     成功すれば宣言者以外の全員に 5 点、失敗すれば宣言者に 5 点。
type Polignac struct {
	trumpCards *TrumpCards
	players    []*PolignacPlayer
	config     PolignacConfig

	phase       PolignacPhase
	roundNumber int
	trickNumber int
	// currentTrick 現在のトリックに出された札
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	// capotIdx capot を宣言したプレイヤー（-1: 宣言なし）
	capotIdx int
	// capotTricks capot 宣言者が取ったトリック数
	capotTricks int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewPolignac コンストラクタ
func NewPolignac(trumpCards *TrumpCards, players []*PolignacPlayer, config PolignacConfig) *Polignac {
	return &Polignac{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		capotIdx:   -1,
		winnerIdx:  -1,
	}
}

// NewDefaultPolignac 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultPolignac() *Polignac {
	players := make([]*PolignacPlayer, 0, PolignacPlayerCnt)
	for i := range PolignacPlayerCnt {
		players = append(players, NewPolignacPlayer(i == 0))
	}
	return NewPolignac(NewTrumpCardsBelote(), players, DefaultPolignacConfig())
}

// Reset ゲーム全体を初期化する
func (p *Polignac) Reset() {
	p.roundNumber = 1
	p.dealerIdx = 0
	p.gameEndFlag = false
	p.winnerIdx = -1
	p.actionLog = nil
	for _, pl := range p.players {
		pl.ResetGame()
	}
	p.dealRound()
}

// dealRound 1 ラウンド分を配る。配り終えたら capot 宣言の受付から始まる。
func (p *Polignac) dealRound() {
	p.phase = PolignacPhaseDeclare
	p.trickNumber = 0
	p.currentTrick = nil
	p.capotIdx = -1
	p.capotTricks = 0
	for _, pl := range p.players {
		pl.ResetRound()
	}

	// **ピケット 32 枚。** Skat の newSkatDeck は 32 枚だが中身が
	// ♠1-13/♣1-13/♥1-6/♦なし で、デッキとして成立していない (#5296)。
	p.trumpCards = NewTrumpCardsBelote()
	p.trumpCards.Shuffle()
	for range PolignacHandSize {
		for i := range PolignacPlayerCnt {
			idx := (p.dealerIdx + 1 + i) % PolignacPlayerCnt
			if c := p.trumpCards.DrawCard(); c != nil {
				p.players[idx].AddCard(c)
			}
		}
	}
	p.leadPlayerIdx = (p.dealerIdx + 1) % PolignacPlayerCnt
	p.currentPlayerIdx = p.leadPlayerIdx
	p.sortAllHands()
	p.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", p.roundNumber), nil)
}

// sortAllHands 手札をスートごと・強さ順に並べる
func (p *Polignac) sortAllHands() {
	for _, pl := range p.players {
		sortPlayerHand(pl, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return polignacRank(ci) < polignacRank(cj)
		})
	}
}

// DeclareCapot 人間プレイヤーが capot を宣言する。宣言せず進める場合は Pass。
func (p *Polignac) DeclareCapot() error {
	if p.phase != PolignacPhaseDeclare {
		return errors.New("not the declaration phase")
	}
	if p.capotIdx >= 0 {
		return errors.New("capot already declared")
	}
	p.capotIdx = 0
	p.players[0].SetDeclaredCapot(true)
	p.appendLog(0, "capot", "capot（全トリック獲得）を宣言", nil)
	p.startPlay()
	return nil
}

// PassDeclaration 宣言せずにプレイへ進む。CPU は capot を宣言しない
// （成功率が低く、読みの材料も無いため）。
func (p *Polignac) PassDeclaration() error {
	if p.phase != PolignacPhaseDeclare {
		return errors.New("not the declaration phase")
	}
	p.appendLog(0, "pass", "宣言なし", nil)
	p.startPlay()
	return nil
}

// startPlay 宣言フェーズを終えてプレイに入る
func (p *Polignac) startPlay() {
	p.phase = PolignacPhasePlay
	p.currentPlayerIdx = p.leadPlayerIdx
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (p *Polignac) PlayerPlay(cardIndex int) error {
	if p.gameEndFlag {
		return errors.New("game has ended")
	}
	if p.phase != PolignacPhasePlay {
		return errors.New("not the play phase")
	}
	if p.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return p.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (p *Polignac) CpuPlay() {
	if p.gameEndFlag || p.phase != PolignacPhasePlay || p.currentPlayerIdx == 0 {
		return
	}
	_ = p.play(p.currentPlayerIdx, p.chooseCpuCard(p.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (p *Polignac) play(playerIdx, cardIndex int) error {
	pl := p.players[playerIdx]
	if cardIndex < 0 || cardIndex >= pl.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := pl.GetCard(cardIndex)
	if !p.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	pl.RemoveCard(cardIndex)
	p.currentTrick = append(p.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	p.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(p.currentTrick) < PolignacPlayerCnt {
		p.currentPlayerIdx = (playerIdx + 1) % PolignacPlayerCnt
		return nil
	}
	p.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (p *Polignac) canPlay(playerIdx int, card *Card) bool {
	if len(p.currentTrick) == 0 {
		return true
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	pl := p.players[playerIdx]
	for i := range pl.GetCardsSize() {
		if pl.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (p *Polignac) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(p.players) {
		return nil
	}
	pl := p.players[playerIdx]
	valid := make([]int, 0, pl.GetCardsSize())
	for i := range pl.GetCardsSize() {
		if p.canPlay(playerIdx, pl.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、含まれるジャックの失点を勝者に科す
func (p *Polignac) resolveTrick() {
	winner := p.trickWinner()
	cards := make([]*Card, 0, len(p.currentTrick))
	penalty := 0
	for _, tc := range p.currentTrick {
		cards = append(cards, tc.Card)
		penalty += PolignacCardPenalty(tc.Card)
	}
	p.players[winner].AddTrick(cards)
	if penalty > 0 {
		p.players[winner].AddRoundPenalty(penalty)
		p.appendLog(winner, "penalty", fmt.Sprintf("ジャックを取って +%d 失点", penalty), nil)
	}
	if winner == p.capotIdx {
		p.capotTricks++
	}

	p.trickNumber++
	p.currentTrick = nil
	p.leadPlayerIdx = winner
	p.currentPlayerIdx = winner

	if p.trickNumber >= PolignacTricksPerRound {
		p.finishRound()
	}
}

// PolignacCardPenalty その札を取ったときの失点を返す。
// **ジャックだけが失点し、スペードのジャックはその 2 倍。**
func PolignacCardPenalty(c *Card) int {
	if c == nil || c.GetValue() != PolignacJackValue {
		return 0
	}
	if c.GetDesign() == CardDesignSpade {
		return PolignacSpadeJackPenalty
	}
	return PolignacJackPenalty
}

// finishRound ラウンドの失点を確定させる
func (p *Polignac) finishRound() {
	if p.capotIdx >= 0 {
		p.settleCapot()
	} else {
		for i, pl := range p.players {
			if n := pl.GetRoundPenalty(); n > 0 {
				pl.AddScore(n)
				p.appendLog(i, "score", fmt.Sprintf("失点 %d", n), nil)
			}
		}
	}

	if p.roundNumber >= p.config.Rounds {
		p.finishGame()
		return
	}
	p.phase = PolignacPhaseRoundEnd
}

// settleCapot capot の成否を判定して精算する。
//
// **成功したらジャックの失点は帳消し。** 全トリック取ったのだからジャックも
// 全部持っており、そのまま加算すると成功しても 5 点しか得しないことになる。
func (p *Polignac) settleCapot() {
	declarer := p.players[p.capotIdx]
	if p.capotTricks >= PolignacTricksPerRound {
		for i, pl := range p.players {
			if i != p.capotIdx {
				pl.AddScore(PolignacCapotStake)
			}
		}
		p.appendLog(p.capotIdx, "capot", "capot 成功。他の全員に 5 失点", nil)
		return
	}
	declarer.AddScore(PolignacCapotStake)
	p.appendLog(p.capotIdx, "capot", "capot 失敗。宣言者に 5 失点", nil)
	// 失敗した場合、そのラウンドのジャックの失点も通常どおり科す。
	for i, pl := range p.players {
		if n := pl.GetRoundPenalty(); n > 0 {
			pl.AddScore(n)
			p.appendLog(i, "score", fmt.Sprintf("失点 %d", n), nil)
		}
	}
}

// NextRound 次のラウンドを開始する
func (p *Polignac) NextRound() {
	if p.gameEndFlag || p.phase != PolignacPhaseRoundEnd {
		return
	}
	p.roundNumber++
	p.dealerIdx = (p.dealerIdx + 1) % PolignacPlayerCnt
	p.dealRound()
}

// finishGame 勝者を決めて終了する。**失点が最も少ないプレイヤーが勝つ。**
func (p *Polignac) finishGame() {
	p.phase = PolignacPhaseGameEnd
	p.gameEndFlag = true

	best := p.players[0].GetScore()
	bestIdx := 0
	tied := false
	for i := 1; i < len(p.players); i++ {
		switch score := p.players[i].GetScore(); {
		case score < best:
			best, bestIdx, tied = score, i, false
		case score == best:
			tied = true
		}
	}
	if tied {
		p.winnerIdx = -1
		p.appendLog(-1, "result", "同点で決着つかず", nil)
		return
	}
	p.winnerIdx = bestIdx
	p.appendLog(bestIdx, "result", fmt.Sprintf("勝者（失点 %d）", best), nil)
}

// trickWinner 現在のトリックの勝者。切り札が無いので、リードのスートの最強札。
func (p *Polignac) trickWinner() int {
	if len(p.currentTrick) == 0 {
		return p.leadPlayerIdx
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestIdx := p.currentTrick[0].PlayerIdx
	best := p.currentTrick[0].Card
	for _, tc := range p.currentTrick[1:] {
		if tc.Card.GetDesign() != leadSuit {
			continue
		}
		if polignacRank(tc.Card) > polignacRank(best) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// polignacRank 札の強さ。A が最強、7 が最弱。
func polignacRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return CardValueMax + 1 // A は K より強い
	}
	return c.GetValue()
}

// chooseCpuCard CPU の手を選ぶ。
//
// capot 宣言者がいるときは**全力で邪魔する**（1 トリック取れば宣言は潰れる）。
// それ以外は回避に徹し、ジャックの乗ったトリックを取らないようにする。
func (p *Polignac) chooseCpuCard(playerIdx int) int {
	valid := p.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	pl := p.players[playerIdx]

	// capot を潰しに行く場面では、取れるなら取る。
	if p.capotIdx >= 0 && p.capotIdx != playerIdx && len(p.currentTrick) > 0 {
		if idx, ok := p.pickWinning(pl, valid); ok {
			return idx
		}
	}

	if len(p.currentTrick) == 0 {
		// リード。ジャックを持っていても自分から出すと自分が取りかねないので、
		// 低い非ジャックでリードする。
		return p.pickLowestNonJack(pl, valid)
	}

	// ジャックが乗っていれば何としても取らない。乗っていなければ、
	// 取らずに済む一番高い札を捨てて手を軽くする。
	loseIdx, loseRank := -1, -1
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestSoFar := p.currentBestRank()
	for _, i := range valid {
		c := pl.GetCard(i)
		r := polignacRank(c)
		followsLead := c.GetDesign() == leadSuit
		if (!followsLead || r < bestSoFar) && r > loseRank {
			loseIdx, loseRank = i, r
		}
	}
	// 取らずに済むなら、そこでジャックを処分できれば理想。
	if loseIdx >= 0 {
		if jIdx, ok := p.pickDumpableJack(pl, valid); ok {
			return jIdx
		}
		return loseIdx
	}
	return p.pickLowest(pl, valid)
}

// currentBestRank 現在のトリックでリードのスートの最強ランク
func (p *Polignac) currentBestRank() int {
	if len(p.currentTrick) == 0 {
		return -1
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	best := -1
	for _, tc := range p.currentTrick {
		if tc.Card.GetDesign() == leadSuit && polignacRank(tc.Card) > best {
			best = polignacRank(tc.Card)
		}
	}
	return best
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (p *Polignac) wouldWin(c *Card) bool {
	if c == nil || len(p.currentTrick) == 0 {
		return true
	}
	if c.GetDesign() != p.currentTrick[0].Card.GetDesign() {
		return false
	}
	return polignacRank(c) > p.currentBestRank()
}

// pickWinning トリックを取れる札のうち一番安いものを選ぶ
func (p *Polignac) pickWinning(pl *PolignacPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := pl.GetCard(i)
		if !p.wouldWin(c) {
			continue
		}
		if r := polignacRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// pickDumpableJack 取らずに捨てられるジャックを探す（スペードを優先して処分）
func (p *Polignac) pickDumpableJack(pl *PolignacPlayer, valid []int) (int, bool) {
	bestIdx, bestPenalty := -1, 0
	for _, i := range valid {
		c := pl.GetCard(i)
		pen := PolignacCardPenalty(c)
		if pen == 0 || p.wouldWin(c) {
			continue
		}
		if pen > bestPenalty {
			bestIdx, bestPenalty = i, pen
		}
	}
	return bestIdx, bestIdx >= 0
}

// pickLowestNonJack 一番弱い非ジャックを選ぶ。全部ジャックなら一番弱い札。
func (p *Polignac) pickLowestNonJack(pl *PolignacPlayer, valid []int) int {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := pl.GetCard(i)
		if PolignacCardPenalty(c) > 0 {
			continue
		}
		if r := polignacRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return p.pickLowest(pl, valid)
}

// pickLowest 合法手のなかで一番弱い札
func (p *Polignac) pickLowest(pl *PolignacPlayer, valid []int) int {
	bestIdx, bestRank := valid[0], polignacRank(pl.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if r := polignacRank(pl.GetCard(i)); r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// PolignacHint ヒント情報
type PolignacHint struct {
	// CardIndex 推奨する手札のインデックス
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す。手番でなければ nil。
func (p *Polignac) GetHint() *PolignacHint {
	if !p.IsHumanTurn() || p.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := p.chooseCpuCard(0)
	return &PolignacHint{CardIndex: &idx, Reason: p.hintReason()}
}

// hintReason 現在の狙いを表す理由キーを返す
func (p *Polignac) hintReason() string {
	if p.capotIdx >= 0 && p.capotIdx != 0 {
		return "polignacBlockCapot"
	}
	if len(p.currentTrick) == 0 {
		return "polignacLeadSafe"
	}
	if p.trickHasJack() {
		return "polignacAvoidJack"
	}
	return "polignacDumpJack"
}

// trickHasJack 現在のトリックにジャックが乗っているか
func (p *Polignac) trickHasJack() bool {
	for _, tc := range p.currentTrick {
		if PolignacCardPenalty(tc.Card) > 0 {
			return true
		}
	}
	return false
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (p *Polignac) GetPhase() PolignacPhase { return p.phase }

// GetConfig 現在の設定
func (p *Polignac) GetConfig() PolignacConfig { return p.config }

// SetConfig 設定を差し替える
func (p *Polignac) SetConfig(c PolignacConfig) { p.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (p *Polignac) GetRoundNumber() int { return p.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (p *Polignac) GetTrickNumber() int { return p.trickNumber }

// GetCurrentTrick 現在のトリック
func (p *Polignac) GetCurrentTrick() []*TrickCard { return p.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (p *Polignac) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (p *Polignac) GetLeadPlayerIdx() int { return p.leadPlayerIdx }

// GetDealerIdx ディーラー
func (p *Polignac) GetDealerIdx() int { return p.dealerIdx }

// GetCapotIdx capot 宣言者（-1: 宣言なし）
func (p *Polignac) GetCapotIdx() int { return p.capotIdx }

// GetCapotTricks capot 宣言者が取ったトリック数
func (p *Polignac) GetCapotTricks() int { return p.capotTricks }

// GetPlayerCnt プレイヤー数
func (p *Polignac) GetPlayerCnt() int { return len(p.players) }

// GetPlayer 指定インデックスのプレイヤー
func (p *Polignac) GetPlayer(i int) *PolignacPlayer {
	if i < 0 || i >= len(p.players) {
		return nil
	}
	return p.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (p *Polignac) GetGameEndFlag() bool { return p.gameEndFlag }

// GetWinnerIdx 勝者（-1: 未確定または同点）
func (p *Polignac) GetWinnerIdx() int { return p.winnerIdx }

// IsHumanTurn 人間の手番か
func (p *Polignac) IsHumanTurn() bool {
	return !p.gameEndFlag && p.phase == PolignacPhasePlay && p.currentPlayerIdx == 0
}

// IsDeclarePhase capot 宣言の受付中か
func (p *Polignac) IsDeclarePhase() bool {
	return !p.gameEndFlag && p.phase == PolignacPhaseDeclare
}

// GiveUp 投了する
func (p *Polignac) GiveUp() {
	if p.gameEndFlag {
		return
	}
	p.phase = PolignacPhaseGameEnd
	p.gameEndFlag = true
	p.winnerIdx = -1
	p.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (p *Polignac) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.appendLogAt(p.trickNumber, playerIdx, actionType, detail, cards)
}

// polignacJSON is the KV snapshot format for Polignac.
type polignacJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PolignacPlayer `json:"pl"`
	Config           PolignacConfig    `json:"cf"`
	Phase            PolignacPhase     `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	CurrentPlayerIdx int               `json:"cp"`
	LeadPlayerIdx    int               `json:"lp"`
	DealerIdx        int               `json:"di"`
	CapotIdx         int               `json:"ci"`
	CapotTricks      int               `json:"ck"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (p *Polignac) MarshalJSON() ([]byte, error) {
	return json.Marshal(&polignacJSON{
		TrumpCards:       p.trumpCards,
		Players:          p.players,
		Config:           p.config,
		Phase:            p.phase,
		RoundNumber:      p.roundNumber,
		TrickNumber:      p.trickNumber,
		CurrentTrick:     p.currentTrick,
		CurrentPlayerIdx: p.currentPlayerIdx,
		LeadPlayerIdx:    p.leadPlayerIdx,
		DealerIdx:        p.dealerIdx,
		CapotIdx:         p.capotIdx,
		CapotTricks:      p.capotTricks,
		GameEndFlag:      p.gameEndFlag,
		WinnerIdx:        p.winnerIdx,
		ActionLog:        p.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた
// 任意のバイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (p *Polignac) UnmarshalJSON(data []byte) error {
	var j polignacJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < PolignacPhaseDeclare || j.Phase > PolignacPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > PolignacTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 || j.RoundNumber > PolignacRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > polignacMaxSliceLen {
		return errors.New("polignac: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > PolignacPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= PolignacPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.CapotIdx < -1 || j.CapotIdx >= PolignacPlayerCnt {
		return fmt.Errorf("invalid capot declarer: %d", j.CapotIdx)
	}
	if j.CapotTricks < 0 || j.CapotTricks > PolignacTricksPerRound {
		return fmt.Errorf("invalid capot tricks: %d", j.CapotTricks)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= PolignacPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		p.trumpCards = j.TrumpCards
	}
	if len(j.Players) == PolignacPlayerCnt {
		p.players = j.Players
	}
	p.config = j.Config
	p.phase = j.Phase
	p.roundNumber = j.RoundNumber
	p.trickNumber = j.TrickNumber
	p.currentTrick = j.CurrentTrick
	p.currentPlayerIdx = j.CurrentPlayerIdx
	p.leadPlayerIdx = j.LeadPlayerIdx
	p.dealerIdx = j.DealerIdx
	p.capotIdx = j.CapotIdx
	p.capotTricks = j.CapotTricks
	p.gameEndFlag = j.GameEndFlag
	p.winnerIdx = j.WinnerIdx
	p.actionLog = j.ActionLog
	return nil
}
