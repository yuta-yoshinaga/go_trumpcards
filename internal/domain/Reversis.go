//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ReversisPhase レヴェルシのゲームフェーズ
type ReversisPhase int

// Reversisのフェーズ定数
const (
	// ReversisPhasePlay プレイ中
	ReversisPhasePlay ReversisPhase = iota
	// ReversisPhaseRoundEnd ラウンド終了（配当確定、次ラウンド待ち）
	ReversisPhaseRoundEnd
	// ReversisPhaseGameEnd ゲーム終了
	ReversisPhaseGameEnd
)

// ReversisPlayerCnt プレイヤー数（4 人固定）
const ReversisPlayerCnt = 4

// ReversisHandSize 各プレイヤーの手札枚数
const ReversisHandSize = 12

// ReversisTricksPerRound 1 ラウンドのトリック数
const ReversisTricksPerRound = ReversisHandSize

// ReversisStartingChips 開始時の持ちチップ
const ReversisStartingChips = 50

// ReversisAnte 各ラウンドの開始時に全員がプールへ出すチップ
const ReversisAnte = 5

// ReversisMarkedStake 印付きの札を取ったときにプールへ追加で払うチップ
const ReversisMarkedStake = 5

// ReversisQuinolaSuit キノラのスート（ハート）
const ReversisQuinolaSuit = CardDesignHeart

// ReversisQuinolaValue キノラの値（ジャック）
const ReversisQuinolaValue = 11

// ReversisMarkedPenalty 印付きの札を取ったときの追加失点
const ReversisMarkedPenalty = 5

// reversisMaxSliceLen caps slice sizes during deserialisation.
const reversisMaxSliceLen = 1000

// Reversis レヴェルシ ゲームクラス。
//
// 17〜18 世紀のフランスで流行した**回避型**トリックテイキング。標準 52 枚から
// **10 を 4 枚抜いた 48 枚**を 4 人に 12 枚ずつ配り、12 トリックを戦う。
// **切り札は無い。**
//
// トリックを取ると、そこに含まれる絵札・A が失点になる:
//
//	A = 4 / K = 3 / Q = 2 / J = 1 / その他 = 0
//
// 1 スートで 10 点、4 スートで **40 点**が毎ラウンド動く。
//
// さらに印の付いた 2 枚がある:
//
//   - **キノラ（Quinola）= ♥J**
//   - **♦A**
//
// これを取ると追加で 5 失点し、**プールへ 5 チップを払う**。ゲーム名の由来である
// 「取らないほうが得」という反転に、賭け金の要素が乗る。
//
// 各ラウンドの開始時に全員が 5 チップをプールへ出し、**失点が最も少なかった
// プレイヤーがプールを総取りする**。規定ラウンド後、**チップが最も多い**
// プレイヤーの勝ち。
//
// issue #5235 の仕様案に対して、次の 2 点を補って実装した:
//
//   - **カードの点数配分が書かれていない。** 「そのトリックのカードポイント」と
//     あるだけなので、歴史的な A=4/K=3/Q=2/J=1 を採った。1 ラウンド 40 点ちょうどに
//     なる、この種のゲームで最も一般的な配分。
//   - **キノラと ♦A は「ボーナス」ではなく罰。** issue は「最後に持っていた
//     プレイヤーに特別ボーナス」と書くが、トリックテイキングでは最後に札を
//     持っている人はおらず、また回避型で「取ると得」では筋が通らない。
//     実際のレヴェルシでもキノラは**最も避けたい 1 枚**なので、追加失点＋
//     プールへの支払いとして実装した。
//
// なお歴史的なレヴェルシには「全トリックを取ると評価が反転する（faire
// reversis）」という大技もあるが、issue の仕様に含まれないため実装していない。
type Reversis struct {
	trumpCards *TrumpCards
	players    []*ReversisPlayer
	config     ReversisConfig

	phase       ReversisPhase
	roundNumber int
	trickNumber int
	// pool 場に積まれたチップ
	pool int
	// currentTrick 現在のトリックに出された札
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewReversis コンストラクタ
func NewReversis(trumpCards *TrumpCards, players []*ReversisPlayer, config ReversisConfig) *Reversis {
	return &Reversis{trumpCards: trumpCards, players: players, config: config, winnerIdx: -1}
}

// NewDefaultReversis 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultReversis() *Reversis {
	players := make([]*ReversisPlayer, 0, ReversisPlayerCnt)
	for i := range ReversisPlayerCnt {
		players = append(players, NewReversisPlayer(i == 0))
	}
	return NewReversis(NewTrumpCardsReversis(), players, DefaultReversisConfig())
}

// Reset ゲーム全体を初期化する
func (r *Reversis) Reset() {
	r.phase = ReversisPhasePlay
	r.roundNumber = 1
	r.dealerIdx = 0
	r.gameEndFlag = false
	r.winnerIdx = -1
	r.actionLog = nil
	for _, p := range r.players {
		p.ResetGame()
	}
	r.dealRound()
}

// dealRound 1 ラウンド分を配り、全員からアンティを集める
func (r *Reversis) dealRound() {
	r.trickNumber = 0
	r.currentTrick = nil
	r.pool = 0
	for _, p := range r.players {
		p.ResetRound()
		p.AddChips(-ReversisAnte)
		r.pool += ReversisAnte
	}

	r.trumpCards = NewTrumpCardsReversis()
	r.trumpCards.Shuffle()
	for range ReversisHandSize {
		for i := range ReversisPlayerCnt {
			idx := (r.dealerIdx + 1 + i) % ReversisPlayerCnt
			if c := r.trumpCards.DrawCard(); c != nil {
				r.players[idx].AddCard(c)
			}
		}
	}
	r.leadPlayerIdx = (r.dealerIdx + 1) % ReversisPlayerCnt
	r.currentPlayerIdx = r.leadPlayerIdx
	r.sortAllHands()
	r.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d 開始（プール %d）", r.roundNumber, r.pool), nil)
}

// sortAllHands 手札をスートごと・強さ順に並べる
func (r *Reversis) sortAllHands() {
	for _, p := range r.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return reversisRank(ci) < reversisRank(cj)
		})
	}
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (r *Reversis) PlayerPlay(cardIndex int) error {
	if r.gameEndFlag {
		return errors.New("game has ended")
	}
	if r.phase != ReversisPhasePlay {
		return errors.New("round has ended")
	}
	if r.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return r.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (r *Reversis) CpuPlay() {
	if r.gameEndFlag || r.phase != ReversisPhasePlay || r.currentPlayerIdx == 0 {
		return
	}
	_ = r.play(r.currentPlayerIdx, r.chooseCpuCard(r.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (r *Reversis) play(playerIdx, cardIndex int) error {
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

	if len(r.currentTrick) < ReversisPlayerCnt {
		r.currentPlayerIdx = (playerIdx + 1) % ReversisPlayerCnt
		return nil
	}
	r.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (r *Reversis) canPlay(playerIdx int, card *Card) bool {
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
func (r *Reversis) GetValidPlayIndices(playerIdx int) []int {
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

// resolveTrick トリックを解決し、失点と印付き札の支払いを勝者に科す
func (r *Reversis) resolveTrick() {
	winner := r.trickWinner()
	cards := make([]*Card, 0, len(r.currentTrick))
	penalty := 0
	for _, tc := range r.currentTrick {
		cards = append(cards, tc.Card)
		penalty += ReversisCardPenalty(tc.Card)
		if ReversisIsQuinola(tc.Card) {
			r.players[winner].SetTookQuinola(true)
			r.chargeMarked(winner, "キノラ（♥J）")
		}
		if ReversisIsDiamondAce(tc.Card) {
			r.players[winner].SetTookDiamondAce(true)
			r.chargeMarked(winner, "♦A")
		}
	}
	r.players[winner].AddTrick(cards)
	if penalty > 0 {
		r.players[winner].AddRoundPenalty(penalty)
	}

	r.trickNumber++
	r.currentTrick = nil
	r.leadPlayerIdx = winner
	r.currentPlayerIdx = winner

	if r.trickNumber >= ReversisTricksPerRound {
		r.finishRound()
	}
}

// chargeMarked 印付きの札を取った罰。**追加失点とプールへの支払いの両方。**
func (r *Reversis) chargeMarked(winner int, name string) {
	r.players[winner].AddRoundPenalty(ReversisMarkedPenalty)
	r.players[winner].AddChips(-ReversisMarkedStake)
	r.pool += ReversisMarkedStake
	r.appendLog(winner, "marked",
		fmt.Sprintf("%s を取った（+%d失点、プールへ %d）", name, ReversisMarkedPenalty, ReversisMarkedStake), nil)
}

// ReversisCardPenalty その札の失点を返す。A=4 / K=3 / Q=2 / J=1 / その他=0。
//
// **この配分は issue #5235 に書かれていない。** 「カードポイント」としか
// 書かれていないため、歴史的で最も一般的なこの配分を採った。1 スート 10 点、
// 4 スートで 40 点ちょうどになる。
func ReversisCardPenalty(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 4
	case 13: // K
		return 3
	case 12: // Q
		return 2
	case 11: // J
		return 1
	default:
		return 0
	}
}

// ReversisIsQuinola キノラ（♥J）かどうか
func ReversisIsQuinola(c *Card) bool {
	return c != nil && c.GetDesign() == ReversisQuinolaSuit && c.GetValue() == ReversisQuinolaValue
}

// ReversisIsDiamondAce ♦A かどうか
func ReversisIsDiamondAce(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignDiamond && c.GetValue() == 1
}

// finishRound 失点の最も少ないプレイヤーがプールを総取りする
func (r *Reversis) finishRound() {
	best, bestIdx, tied := r.players[0].GetRoundPenalty(), 0, false
	for i := 1; i < len(r.players); i++ {
		switch pen := r.players[i].GetRoundPenalty(); {
		case pen < best:
			best, bestIdx, tied = pen, i, false
		case pen == best:
			tied = true
		}
	}

	switch {
	case tied && r.roundNumber >= r.config.Rounds:
		// **最終ラウンドの同点は持ち越せない。** NextRound() は二度と走らないので、
		// そのままだとプールのチップが誰のものにもならないまま盤上に取り残される。
		// 総量としては保存されていても、勝敗を決める GetChips() には入らない。
		r.splitPoolAmongTied(best)
	case tied:
		// まだ先があるなら次のラウンドへ持ち越す。プールが膨らむぶん次が重くなる。
		r.appendLog(-1, "pool", fmt.Sprintf("最少失点が同点。プール %d を持ち越し", r.pool), nil)
	default:
		r.players[bestIdx].AddChips(r.pool)
		r.appendLog(bestIdx, "pool", fmt.Sprintf("最少失点（%d）でプール %d を獲得", best, r.pool), nil)
		r.pool = 0
	}

	if r.roundNumber >= r.config.Rounds {
		r.finishGame()
		return
	}
	r.phase = ReversisPhaseRoundEnd
}

// splitPoolAmongTied 最終ラウンドが同点で終わったとき、プールを同点者で分ける。
//
// **1 チップも盤上に残さない。** 割り切れないぶんはディーラーの左隣から順に
// 1 枚ずつ配る。ここで捨ててしまうと「チップは生まれも消えもしない」が
// 総量としては保たれても、勝敗判定からは消える。
func (r *Reversis) splitPoolAmongTied(best int) {
	tiedIdx := make([]int, 0, len(r.players))
	for i := range len(r.players) {
		// ディーラーの左隣から順に見るので、端数の配り先も決定的になる。
		idx := (r.dealerIdx + 1 + i) % len(r.players)
		if r.players[idx].GetRoundPenalty() == best {
			tiedIdx = append(tiedIdx, idx)
		}
	}
	if len(tiedIdx) == 0 {
		return
	}
	share, remainder := r.pool/len(tiedIdx), r.pool%len(tiedIdx)
	for n, idx := range tiedIdx {
		amount := share
		if n < remainder {
			amount++
		}
		r.players[idx].AddChips(amount)
	}
	r.appendLog(-1, "pool",
		fmt.Sprintf("最終ラウンドが同点。プール %d を %d 人で分配", r.pool, len(tiedIdx)), nil)
	r.pool = 0
}

// NextRound 次のラウンドを開始する。持ち越したプールはそのまま残る。
func (r *Reversis) NextRound() {
	if r.gameEndFlag || r.phase != ReversisPhaseRoundEnd {
		return
	}
	carried := r.pool
	r.roundNumber++
	r.dealerIdx = (r.dealerIdx + 1) % ReversisPlayerCnt
	r.phase = ReversisPhasePlay
	r.dealRound()
	r.pool += carried
}

// finishGame チップが最も多いプレイヤーの勝ち
func (r *Reversis) finishGame() {
	r.phase = ReversisPhaseGameEnd
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

// trickWinner 現在のトリックの勝者。切り札が無いので、リードのスートの最強札。
func (r *Reversis) trickWinner() int {
	if len(r.currentTrick) == 0 {
		return r.leadPlayerIdx
	}
	leadSuit := r.currentTrick[0].Card.GetDesign()
	bestIdx, best := r.currentTrick[0].PlayerIdx, r.currentTrick[0].Card
	for _, tc := range r.currentTrick[1:] {
		if tc.Card.GetDesign() != leadSuit {
			continue
		}
		if reversisRank(tc.Card) > reversisRank(best) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// reversisRank 札の強さ。A が最強、2 が最弱。10 はデッキに無い。
func reversisRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return CardValueMax + 1 // A は K より強い
	}
	return c.GetValue()
}

// chooseCpuCard CPU の手を選ぶ。取りたくないゲームなので、取らずに済む一番高い
// 札を捨て、それができないときだけ最弱を出す。
func (r *Reversis) chooseCpuCard(playerIdx int) int {
	valid := r.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := r.players[playerIdx]

	if len(r.currentTrick) == 0 {
		// リード。印付きの札や高い絵札を自分から出すと自分が取りかねない。
		return r.pickSafestLead(p, valid)
	}

	loseIdx, loseRank := -1, -1
	leadSuit := r.currentTrick[0].Card.GetDesign()
	bestSoFar := r.currentBestRank()
	for _, i := range valid {
		c := p.GetCard(i)
		rank := reversisRank(c)
		followsLead := c.GetDesign() == leadSuit
		if (!followsLead || rank < bestSoFar) && rank > loseRank {
			loseIdx, loseRank = i, rank
		}
	}
	if loseIdx >= 0 {
		// 取らずに済むなら、そこで印付きの札を処分できれば理想。
		if mIdx, ok := r.pickDumpableMarked(p, valid); ok {
			return mIdx
		}
		return loseIdx
	}
	return r.pickLowestPenalty(p, valid)
}

// currentBestRank 現在のトリックでリードのスートの最強ランク
func (r *Reversis) currentBestRank() int {
	if len(r.currentTrick) == 0 {
		return -1
	}
	leadSuit := r.currentTrick[0].Card.GetDesign()
	best := -1
	for _, tc := range r.currentTrick {
		if tc.Card.GetDesign() == leadSuit && reversisRank(tc.Card) > best {
			best = reversisRank(tc.Card)
		}
	}
	return best
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (r *Reversis) wouldWin(c *Card) bool {
	if c == nil || len(r.currentTrick) == 0 {
		return true
	}
	if c.GetDesign() != r.currentTrick[0].Card.GetDesign() {
		return false
	}
	return reversisRank(c) > r.currentBestRank()
}

// pickDumpableMarked 取らずに捨てられる印付きの札を探す
func (r *Reversis) pickDumpableMarked(p *ReversisPlayer, valid []int) (int, bool) {
	for _, i := range valid {
		c := p.GetCard(i)
		if (ReversisIsQuinola(c) || ReversisIsDiamondAce(c)) && !r.wouldWin(c) {
			return i, true
		}
	}
	return 0, false
}

// pickSafestLead リード時の札。失点 0 の札のうち一番弱いものを選ぶ。
func (r *Reversis) pickSafestLead(p *ReversisPlayer, valid []int) int {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if ReversisCardPenalty(c) > 0 || ReversisIsQuinola(c) || ReversisIsDiamondAce(c) {
			continue
		}
		if rank := reversisRank(c); bestIdx < 0 || rank < bestRank {
			bestIdx, bestRank = i, rank
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return r.pickLowestPenalty(p, valid)
}

// pickLowestPenalty 失点が最も小さい札。同点ならランクの低いほう。
func (r *Reversis) pickLowestPenalty(p *ReversisPlayer, valid []int) int {
	bestIdx := valid[0]
	bestPen, bestRank := r.markedAwarePenalty(p.GetCard(bestIdx)), reversisRank(p.GetCard(bestIdx))
	for _, i := range valid[1:] {
		c := p.GetCard(i)
		pen, rank := r.markedAwarePenalty(c), reversisRank(c)
		if pen < bestPen || (pen == bestPen && rank < bestRank) {
			bestIdx, bestPen, bestRank = i, pen, rank
		}
	}
	return bestIdx
}

// markedAwarePenalty 印付きの追加失点も含めた 1 枚の重さ
func (r *Reversis) markedAwarePenalty(c *Card) int {
	pen := ReversisCardPenalty(c)
	if ReversisIsQuinola(c) || ReversisIsDiamondAce(c) {
		pen += ReversisMarkedPenalty
	}
	return pen
}

// ReversisHint ヒント情報
type ReversisHint struct {
	// CardIndex 推奨する手札のインデックス
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す。手番でなければ nil。
func (r *Reversis) GetHint() *ReversisHint {
	if !r.IsHumanTurn() || r.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := r.chooseCpuCard(0)
	return &ReversisHint{CardIndex: &idx, Reason: r.hintReason()}
}

// hintReason 現在の狙いを表す理由キーを返す
func (r *Reversis) hintReason() string {
	if len(r.currentTrick) == 0 {
		return "reversisLeadSafe"
	}
	if r.trickHasMarked() {
		return "reversisAvoidMarked"
	}
	if r.trickHasPenalty() {
		return "reversisAvoidPoints"
	}
	return "reversisDumpHigh"
}

// trickHasMarked 現在のトリックに印付きの札が乗っているか
func (r *Reversis) trickHasMarked() bool {
	for _, tc := range r.currentTrick {
		if ReversisIsQuinola(tc.Card) || ReversisIsDiamondAce(tc.Card) {
			return true
		}
	}
	return false
}

// trickHasPenalty 現在のトリックに失点札が乗っているか
func (r *Reversis) trickHasPenalty() bool {
	for _, tc := range r.currentTrick {
		if ReversisCardPenalty(tc.Card) > 0 {
			return true
		}
	}
	return false
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (r *Reversis) GetPhase() ReversisPhase { return r.phase }

// GetConfig 現在の設定
func (r *Reversis) GetConfig() ReversisConfig { return r.config }

// SetConfig 設定を差し替える
func (r *Reversis) SetConfig(c ReversisConfig) { r.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (r *Reversis) GetRoundNumber() int { return r.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (r *Reversis) GetTrickNumber() int { return r.trickNumber }

// GetPool 場に積まれているチップ
func (r *Reversis) GetPool() int { return r.pool }

// GetCurrentTrick 現在のトリック
func (r *Reversis) GetCurrentTrick() []*TrickCard { return r.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (r *Reversis) GetCurrentPlayerIdx() int { return r.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (r *Reversis) GetLeadPlayerIdx() int { return r.leadPlayerIdx }

// GetDealerIdx ディーラー
func (r *Reversis) GetDealerIdx() int { return r.dealerIdx }

// GetPlayerCnt プレイヤー数
func (r *Reversis) GetPlayerCnt() int { return len(r.players) }

// GetPlayer 指定インデックスのプレイヤー
func (r *Reversis) GetPlayer(i int) *ReversisPlayer {
	if i < 0 || i >= len(r.players) {
		return nil
	}
	return r.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (r *Reversis) GetGameEndFlag() bool { return r.gameEndFlag }

// GetWinnerIdx 勝者（-1: 未確定または同点）
func (r *Reversis) GetWinnerIdx() int { return r.winnerIdx }

// IsHumanTurn 人間の手番か
func (r *Reversis) IsHumanTurn() bool {
	return !r.gameEndFlag && r.phase == ReversisPhasePlay && r.currentPlayerIdx == 0
}

// GiveUp 投了する
func (r *Reversis) GiveUp() {
	if r.gameEndFlag {
		return
	}
	r.phase = ReversisPhaseGameEnd
	r.gameEndFlag = true
	r.winnerIdx = -1
	r.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (r *Reversis) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	r.appendLogAt(r.trickNumber, playerIdx, actionType, detail, cards)
}

// reversisJSON is the KV snapshot format for Reversis.
type reversisJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*ReversisPlayer `json:"pl"`
	Config           ReversisConfig    `json:"cf"`
	Phase            ReversisPhase     `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	Pool             int               `json:"po"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	CurrentPlayerIdx int               `json:"cp"`
	LeadPlayerIdx    int               `json:"lp"`
	DealerIdx        int               `json:"di"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (r *Reversis) MarshalJSON() ([]byte, error) {
	return json.Marshal(&reversisJSON{
		TrumpCards:       r.trumpCards,
		Players:          r.players,
		Config:           r.config,
		Phase:            r.phase,
		RoundNumber:      r.roundNumber,
		TrickNumber:      r.trickNumber,
		Pool:             r.pool,
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
func (r *Reversis) UnmarshalJSON(data []byte) error {
	var j reversisJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < ReversisPhasePlay || j.Phase > ReversisPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > ReversisTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 || j.RoundNumber > ReversisRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.Pool < 0 {
		return fmt.Errorf("invalid pool: %d", j.Pool)
	}
	if len(j.ActionLog) > reversisMaxSliceLen {
		return errors.New("reversis: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > ReversisPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= ReversisPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= ReversisPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		r.trumpCards = j.TrumpCards
	}
	if len(j.Players) == ReversisPlayerCnt {
		r.players = j.Players
	}
	r.config = j.Config
	r.phase = j.Phase
	r.roundNumber = j.RoundNumber
	r.trickNumber = j.TrickNumber
	r.pool = j.Pool
	r.currentTrick = j.CurrentTrick
	r.currentPlayerIdx = j.CurrentPlayerIdx
	r.leadPlayerIdx = j.LeadPlayerIdx
	r.dealerIdx = j.DealerIdx
	r.gameEndFlag = j.GameEndFlag
	r.winnerIdx = j.WinnerIdx
	r.actionLog = j.ActionLog
	return nil
}
