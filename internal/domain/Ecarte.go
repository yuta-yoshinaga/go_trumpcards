//go:build !js || !wasm || casino

// Package domain エカルテ (Écarté) のドメインモデル。
//
// Écarté はフランス発祥の 2 人用トリックゲーム。32 枚デッキ (Belote デッキ = A,7,8,9,10,J,Q,K
// × 4 スート) を使い、各プレイヤーに 5 枚配って残りをストックとし、その上札を表向きにして
// 切り札を決める。表向きが King なら親に 1 点。
//
// 交換フェーズ: 非親 (elder) が「提案 (propose)」か「勝負 (stand)」を選ぶ。提案された親は
// 「承諾 (accept)」か「拒否 (refuse)」を選ぶ。承諾なら両者が任意枚数を捨てて山札から引き直し、
// elder が再度提案/勝負を選ぶ (山札が尽きるまで繰り返せる)。拒否した親がそのディールで 3 トリック
// 未満なら elder に 1 点が追加される (拒否ペナルティ)。
//
// プレイフェーズ: 5 トリックを切り札ありで争う。マストフォロー (フォロー→勝てるなら勝つ→
// 出せないなら切り札) が厳格に課される。ランクは K>Q>J>A>10>9>8>7。切り札の King を手札に
// 持つプレイヤーは 1 点 (自動宣言)。3 トリック以上で 1 点、5 トリック総取り (Vole) で 2 点。
// 累積が目標点 (既定 5) に達した側が試合に勝利する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// EcarteHandSize 各プレイヤーの手札枚数
const EcarteHandSize = 5

// EcarteTricksToWin ディール勝利に必要なトリック数 (過半数)
const EcarteTricksToWin = 3

// EcartePhase ゲームフェーズ
type EcartePhase int

// Écarté のフェーズ定数
const (
	// EcartePhaseExchange 交換ネゴシエーションフェーズ
	EcartePhaseExchange EcartePhase = iota
	// EcartePhasePlay トリックプレイフェーズ
	EcartePhasePlay
	// EcartePhaseRoundEnd ディール終了フェーズ
	EcartePhaseRoundEnd
	// EcartePhaseGameEnd ゲーム終了フェーズ
	EcartePhaseGameEnd
)

// EcarteNegStep 交換フェーズ内の手番ステップ
type EcarteNegStep int

// 交換ステップ定数
const (
	// EcarteNegElderDecide elder が提案/勝負を決める
	EcarteNegElderDecide EcarteNegStep = iota
	// EcarteNegDealerRespond 親が承諾/拒否を決める
	EcarteNegDealerRespond
	// EcarteNegElderDiscard elder が捨て札を選ぶ
	EcarteNegElderDiscard
	// EcarteNegDealerDiscard 親が捨て札を選ぶ
	EcarteNegDealerDiscard
)

// EcarteHint ヒント情報
type EcarteHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイフェーズ)
	Action    string // 推奨アクション (交換フェーズ: "propose"/"stand"/"accept"/"refuse"/"discard")
	Reason    string // ヒント理由キー
}

// EcarteRankOrder カードのスート内順位を返す (大きいほど強い; K>Q>J>A>10>9>8>7)。
func EcarteRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 13: // K
		return 8
	case 12: // Q
		return 7
	case 11: // J
		return 6
	case 1: // A
		return 5
	case 10: // 10
		return 4
	default: // 9,8,7
		return c.GetValue() - 6 // 9→3, 8→2, 7→1
	}
}

// Ecarte エカルテゲームクラス
type Ecarte struct {
	trumpCards       *TrumpCards
	players          []*EcartePlayer
	config           EcarteConfig
	phase            EcartePhase
	negStep          EcarteNegStep
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpCard        *Card // 表向きの切り札表示カード
	trumpSuit        int
	dealerIdx        int
	leadPlayerIdx    int
	dealPoints       []int
	matchScore       []int
	refusalByDealer  bool
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定
	actionLogBase
}

// NewEcarte コンストラクタ
func NewEcarte(trumpCards *TrumpCards, players []*EcartePlayer, config EcarteConfig) *Ecarte {
	return &Ecarte{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
		dealPoints: make([]int, len(players)),
		matchScore: make([]int, len(players)),
	}
}

// NewDefaultEcarte 標準の 2 人対戦セットアップを返す (人間 idx 0 + CPU idx 1)。
func NewDefaultEcarte() *Ecarte {
	players := []*EcartePlayer{
		NewEcartePlayer(true),
		NewEcartePlayer(false),
	}
	return NewEcarte(NewTrumpCardsBelote(), players, DefaultEcarteConfig())
}

// elderIdx 非親 (先手) のインデックスを返す。
func (e *Ecarte) elderIdx() int { return (e.dealerIdx + 1) % EcartePlayerCnt }

// Reset ゲーム初期化
func (e *Ecarte) Reset() {
	e.gameEndFlag = false
	e.winnerIdx = -1
	e.roundNumber = 1
	e.dealerIdx = 0
	e.matchScore = make([]int, len(e.players))
	e.actionLog = nil
	e.startDeal()
}

// NextRound 次のディールを開始する
func (e *Ecarte) NextRound() {
	if e.phase != EcartePhaseRoundEnd {
		return
	}
	e.roundNumber++
	e.dealerIdx = (e.dealerIdx + 1) % EcartePlayerCnt
	e.startDeal()
}

// startDeal 1 ディールを開始する。
func (e *Ecarte) startDeal() {
	e.dealPoints = make([]int, len(e.players))
	e.currentTrick = nil
	e.trumpCard = nil
	e.trumpSuit = 0
	e.refusalByDealer = false

	for _, p := range e.players {
		p.ResetGame()
	}

	e.trumpCards.Shuffle()
	for range EcarteHandSize {
		for i := range EcartePlayerCnt {
			player := e.players[(e.dealerIdx+1+i)%EcartePlayerCnt]
			if c := e.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	e.trumpCard = e.trumpCards.DrawCard()
	if e.trumpCard != nil {
		e.trumpSuit = e.trumpCard.GetDesign()
		e.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(e.trumpCard)), []*Card{e.trumpCard})
		// 表向きが切り札 King なら親に 1 点。
		if e.trumpCard.GetValue() == 13 {
			e.dealPoints[e.dealerIdx]++
			e.appendLog(e.dealerIdx, "king_turn",
				fmt.Sprintf("%s scores the turned King (+1)", playerName(e.players, e.dealerIdx)), nil)
		}
	}
	e.sortAllHands()

	e.phase = EcartePhaseExchange
	e.negStep = EcarteNegElderDecide
	e.currentPlayerIdx = e.elderIdx()
}

// --- Exchange phase actions ---

// PlayerPropose elder が交換を提案する。
func (e *Ecarte) PlayerPropose() error {
	if err := e.checkNeg(EcarteNegElderDecide); err != nil {
		return err
	}
	e.appendLog(e.currentPlayerIdx, "propose", fmt.Sprintf("%s proposes", playerName(e.players, e.currentPlayerIdx)), nil)
	e.negStep = EcarteNegDealerRespond
	e.currentPlayerIdx = e.dealerIdx
	return nil
}

// PlayerStand elder が交換せずに勝負する。
func (e *Ecarte) PlayerStand() error {
	if err := e.checkNeg(EcarteNegElderDecide); err != nil {
		return err
	}
	e.appendLog(e.currentPlayerIdx, "stand", fmt.Sprintf("%s stands", playerName(e.players, e.currentPlayerIdx)), nil)
	e.startPlay()
	return nil
}

// PlayerRespond 親が提案に承諾 (accept=true) か拒否 (accept=false) する。
func (e *Ecarte) PlayerRespond(accept bool) error {
	if err := e.checkNeg(EcarteNegDealerRespond); err != nil {
		return err
	}
	if !accept {
		e.refusalByDealer = true
		e.appendLog(e.currentPlayerIdx, "refuse", fmt.Sprintf("%s refuses", playerName(e.players, e.currentPlayerIdx)), nil)
		e.startPlay()
		return nil
	}
	e.appendLog(e.currentPlayerIdx, "accept", fmt.Sprintf("%s accepts", playerName(e.players, e.currentPlayerIdx)), nil)
	e.negStep = EcarteNegElderDiscard
	e.currentPlayerIdx = e.elderIdx()
	return nil
}

// PlayerDiscard 現在の手番プレイヤーが指定インデックスのカードを捨てて引き直す。
func (e *Ecarte) PlayerDiscard(indices []int) error {
	if e.phase != EcartePhaseExchange {
		return ErrWrongPhase
	}
	if e.negStep != EcarteNegElderDiscard && e.negStep != EcarteNegDealerDiscard {
		return ErrWrongPhase
	}
	if !e.players[e.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return e.applyDiscard(e.currentPlayerIdx, indices)
}

// checkNeg は交換フェーズの指定ステップで人間の手番かを検証する。
func (e *Ecarte) checkNeg(step EcarteNegStep) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EcartePhaseExchange || e.negStep != step {
		return ErrWrongPhase
	}
	if !e.players[e.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// applyDiscard 捨て札と引き直しを実行し、次のネゴシエーションステップへ進める。
func (e *Ecarte) applyDiscard(playerIdx int, indices []int) error {
	p := e.players[playerIdx]
	if len(indices) > e.trumpCards.GetRemainingCount() {
		return NewDomainError(ErrInvalidCard, "山札の残り枚数を超えて交換できません")
	}
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= p.GetCardsSize() || seen[idx] {
			return NewDomainError(ErrInvalidCard, "捨て札のインデックスが不正です")
		}
		seen[idx] = true
	}
	// 大きいインデックスから取り除く (インデックスずれ防止)。
	order := make([]int, 0, len(indices))
	order = append(order, indices...)
	sort.Sort(sort.Reverse(sort.IntSlice(order)))
	n := 0
	for _, idx := range order {
		p.RemoveCard(idx)
		n++
	}
	for i := 0; i < n; i++ {
		if c := e.trumpCards.DrawCard(); c != nil {
			p.AddCard(c)
		}
	}
	e.sortHand(p)
	e.appendLog(playerIdx, "discard",
		fmt.Sprintf("%s exchanges %d card(s)", playerName(e.players, playerIdx), n), nil)

	if e.negStep == EcarteNegElderDiscard {
		e.negStep = EcarteNegDealerDiscard
		e.currentPlayerIdx = e.dealerIdx
		return nil
	}
	// 親の捨て札が終わった: 山札が残っていれば再交渉、無ければプレイ開始。
	if e.trumpCards.GetRemainingCount() > 0 {
		e.negStep = EcarteNegElderDecide
		e.currentPlayerIdx = e.elderIdx()
	} else {
		e.startPlay()
	}
	return nil
}

// startPlay プレイフェーズを開始する。King ボーナスを付与する。
func (e *Ecarte) startPlay() {
	for i, p := range e.players {
		if e.handHasTrumpKing(p) {
			e.dealPoints[i]++
			e.appendLog(i, "king", fmt.Sprintf("%s declares the King of trumps (+1)", playerName(e.players, i)), nil)
		}
	}
	e.phase = EcartePhasePlay
	e.trickNumber = 1
	e.currentTrick = nil
	e.leadPlayerIdx = e.elderIdx()
	e.currentPlayerIdx = e.elderIdx()
}

// handHasTrumpKing 手札に切り札 King を持つか
func (e *Ecarte) handHasTrumpKing(p *EcartePlayer) bool {
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == e.trumpSuit && c.GetValue() == 13 {
			return true
		}
	}
	return false
}

// --- Play phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (e *Ecarte) PlayerPlay(cardIndex int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EcartePhasePlay {
		return ErrWrongPhase
	}
	if !e.players[e.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := e.players[e.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := e.validatePlay(e.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	e.playCard(e.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する (プレイフェーズ)
func (e *Ecarte) CpuPlay() {
	if e.gameEndFlag || e.phase != EcartePhasePlay {
		return
	}
	idx := e.currentPlayerIdx
	if e.players[idx].GetIsHuman() {
		return
	}
	cardIdx := e.cpuSelectPlayCard(idx)
	played := e.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	e.playCard(idx, played)
}

// CpuExchange 現在の交換手番が CPU の場合に 1 アクション実行する。
func (e *Ecarte) CpuExchange() {
	if e.gameEndFlag || e.phase != EcartePhaseExchange {
		return
	}
	idx := e.currentPlayerIdx
	if e.players[idx].GetIsHuman() {
		return
	}
	switch e.negStep {
	case EcarteNegElderDecide:
		if e.cpuWantsExchange(idx) {
			_ = e.PlayerProposeCPU()
		} else {
			e.appendLog(idx, "stand", fmt.Sprintf("%s stands", playerName(e.players, idx)), nil)
			e.startPlay()
		}
	case EcarteNegDealerRespond:
		if e.cpuWantsExchange(idx) {
			e.appendLog(idx, "accept", fmt.Sprintf("%s accepts", playerName(e.players, idx)), nil)
			e.negStep = EcarteNegElderDiscard
			e.currentPlayerIdx = e.elderIdx()
		} else {
			e.refusalByDealer = true
			e.appendLog(idx, "refuse", fmt.Sprintf("%s refuses", playerName(e.players, idx)), nil)
			e.startPlay()
		}
	case EcarteNegElderDiscard, EcarteNegDealerDiscard:
		_ = e.applyDiscard(idx, e.cpuChooseDiscards(idx))
	}
}

// PlayerProposeCPU は CPU 用の propose 内部実装 (検証なし)。
func (e *Ecarte) PlayerProposeCPU() error {
	e.appendLog(e.currentPlayerIdx, "propose", fmt.Sprintf("%s proposes", playerName(e.players, e.currentPlayerIdx)), nil)
	e.negStep = EcarteNegDealerRespond
	e.currentPlayerIdx = e.dealerIdx
	return nil
}

// playCard カードをプレイする共通処理。2枚出そろったらトリックを解決する。
func (e *Ecarte) playCard(playerIdx int, card *Card) {
	e.currentTrick = append(e.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	e.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(e.players, playerIdx), cardStr(card)), []*Card{card})
	if len(e.currentTrick) == EcartePlayerCnt {
		e.resolveTrick()
		return
	}
	e.currentPlayerIdx = (e.currentPlayerIdx + 1) % EcartePlayerCnt
}

// resolveTrick トリックを解決する。
func (e *Ecarte) resolveTrick() {
	winnerIdx := e.trickWinner()
	trickCards := make([]*Card, len(e.currentTrick))
	for i, tc := range e.currentTrick {
		trickCards[i] = tc.Card
	}
	e.players[winnerIdx].AddTrick(trickCards)
	e.leadPlayerIdx = winnerIdx
	e.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(e.players, winnerIdx), e.trickNumber), trickCards)

	if e.allHandsEmpty() {
		e.scoreDeal()
		return
	}
	e.currentTrick = nil
	e.trickNumber++
	e.currentPlayerIdx = winnerIdx
}

// scoreDeal ディールを集計して累積し、ゲーム終了を判定する。
func (e *Ecarte) scoreDeal() {
	tricks := [EcartePlayerCnt]int{}
	for i, p := range e.players {
		tricks[i] = p.GetTrickCount()
	}
	winner := 0
	if tricks[1] >= EcarteTricksToWin {
		winner = 1
	}
	pts := 1
	if tricks[winner] == EcarteHandSize {
		pts = 2 // Vole
	}
	// 拒否ペナルティ: 親が拒否してそのディールに敗れたら elder に +1。
	if e.refusalByDealer && winner == e.elderIdx() {
		pts++
	}
	e.dealPoints[winner] += pts

	for i := range e.players {
		e.matchScore[i] += e.dealPoints[i]
		e.players[i].SetRoundScore(e.dealPoints[i])
		e.players[i].CommitRoundScore()
	}
	e.appendLog(-1, "deal_score",
		fmt.Sprintf("Deal %d: tricks %d-%d, deal %d-%d (match %d-%d)", e.roundNumber,
			tricks[0], tricks[1], e.dealPoints[0], e.dealPoints[1], e.matchScore[0], e.matchScore[1]), nil)

	if e.matchScore[0] >= e.config.TargetScore || e.matchScore[1] >= e.config.TargetScore {
		e.finishGame()
		return
	}
	e.phase = EcartePhaseRoundEnd
}

// finishGame ゲームを終了させ勝者を決定する。
func (e *Ecarte) finishGame() {
	e.gameEndFlag = true
	e.phase = EcartePhaseGameEnd
	switch {
	case e.matchScore[0] > e.matchScore[1]:
		e.winnerIdx = 0
	case e.matchScore[1] > e.matchScore[0]:
		e.winnerIdx = 1
	default:
		e.winnerIdx = e.leadPlayerIdx
	}
	e.appendLog(-1, "game_end", fmt.Sprintf("Game end: %d-%d", e.matchScore[0], e.matchScore[1]), nil)
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (e *Ecarte) allHandsEmpty() bool {
	return allHandsEmpty(e.players)
}

// validatePlay マストフォロー (フォロー→勝てるなら勝つ→出せないなら切り札) を検証する。
func (e *Ecarte) validatePlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードが nil です")
	}
	if len(e.currentTrick) == 0 {
		return nil
	}
	if !e.cardSatisfiesFollow(playerIdx, card) {
		return NewDomainError(ErrInvalidCard, "フォロールール (勝てるなら勝つ・切り札) に従ってください")
	}
	return nil
}

// cardSatisfiesFollow 追随カードが合法かを返す。
func (e *Ecarte) cardSatisfiesFollow(playerIdx int, card *Card) bool {
	player := e.players[playerIdx]
	leadCard := e.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	if ecartePlayerHasSuit(player, leadSuit) {
		if card.GetDesign() != leadSuit {
			return false
		}
		if ecartePlayerHasSuitWinner(player, leadCard, leadSuit, e.trumpSuit) {
			return ecarteBeats(card, leadCard, leadSuit, e.trumpSuit)
		}
		return true
	}
	if ecartePlayerHasSuit(player, e.trumpSuit) {
		return card.GetDesign() == e.trumpSuit
	}
	return true
}

// legalPlayIndices validatePlay を満たすカードのインデックス集合を返す。
func (e *Ecarte) legalPlayIndices(playerIdx int) []int {
	return validPlayIndices(e.players[playerIdx], func(c *Card) bool { return e.validatePlay(playerIdx, c) == nil })
}

func ecartePlayerHasSuit(player *EcartePlayer, suit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

func ecartePlayerHasSuitWinner(player *EcartePlayer, leadCard *Card, leadSuit, trumpSuit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != leadSuit {
			continue
		}
		if ecarteBeats(c, leadCard, leadSuit, trumpSuit) {
			return true
		}
	}
	return false
}

// trickWinner 現在のトリックの勝者インデックスを決定する
func (e *Ecarte) trickWinner() int {
	if len(e.currentTrick) == 0 {
		return 0
	}
	leadSuit := e.currentTrick[0].Card.GetDesign()
	winnerIdx := e.currentTrick[0].PlayerIdx
	winnerCard := e.currentTrick[0].Card
	for _, tc := range e.currentTrick[1:] {
		if ecarteBeats(tc.Card, winnerCard, leadSuit, e.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// ecarteBeats challenger が currentBest に勝つかを判定する。
func ecarteBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit
	switch {
	case cIsTrump && bIsTrump:
		return EcarteRankOrder(challenger) > EcarteRankOrder(currentBest)
	case cIsTrump:
		return true
	case bIsTrump:
		return false
	}
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return EcarteRankOrder(challenger) > EcarteRankOrder(currentBest)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (e *Ecarte) GetPhase() EcartePhase { return e.phase }

// SetPhase フェーズ設定 (テスト用)
func (e *Ecarte) SetPhase(phase EcartePhase) { e.phase = phase }

// GetNegStep 現在の交換ステップ取得
func (e *Ecarte) GetNegStep() EcarteNegStep { return e.negStep }

// SetNegStep 交換ステップ設定 (テスト用)
func (e *Ecarte) SetNegStep(step EcarteNegStep) { e.negStep = step }

// GetRoundNumber 現在のディール番号取得
func (e *Ecarte) GetRoundNumber() int { return e.roundNumber }

// SetRoundNumber ディール番号設定 (テスト用)
func (e *Ecarte) SetRoundNumber(n int) { e.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (e *Ecarte) GetTrickNumber() int { return e.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (e *Ecarte) SetTrickNumber(n int) { e.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (e *Ecarte) GetCurrentPlayerIdx() int { return e.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (e *Ecarte) SetCurrentPlayerIdx(idx int) { e.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (e *Ecarte) GetCurrentTrick() []*TrickCard { return e.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (e *Ecarte) SetCurrentTrick(trick []*TrickCard) { e.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (e *Ecarte) GetTrumpSuit() int { return e.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (e *Ecarte) SetTrumpSuit(suit int) { e.trumpSuit = suit }

// GetTrumpCard 表向きの切り札表示カードを取得
func (e *Ecarte) GetTrumpCard() *Card { return e.trumpCard }

// SetTrumpCard 切り札表示カード設定 (テスト用)
func (e *Ecarte) SetTrumpCard(c *Card) { e.trumpCard = c }

// GetDealerIdx ディーラーインデックス取得
func (e *Ecarte) GetDealerIdx() int { return e.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (e *Ecarte) SetDealerIdx(idx int) { e.dealerIdx = idx }

// GetElderIdx 非親 (先手) インデックス取得
func (e *Ecarte) GetElderIdx() int { return e.elderIdx() }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (e *Ecarte) GetLeadPlayerIdx() int { return e.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (e *Ecarte) SetLeadPlayerIdx(idx int) { e.leadPlayerIdx = idx }

// IsRefusalByDealer 親が交換を拒否したか
func (e *Ecarte) IsRefusalByDealer() bool { return e.refusalByDealer }

// GetDealPoints プレイヤーの当ディール得点取得
func (e *Ecarte) GetDealPoints(i int) int {
	if i < 0 || i >= len(e.dealPoints) {
		return 0
	}
	return e.dealPoints[i]
}

// SetDealPoints プレイヤーの当ディール得点設定 (テスト用)
func (e *Ecarte) SetDealPoints(i, points int) {
	if i >= 0 && i < len(e.dealPoints) {
		e.dealPoints[i] = points
	}
}

// GetMatchScore プレイヤーの試合累積得点取得
func (e *Ecarte) GetMatchScore(i int) int {
	if i < 0 || i >= len(e.matchScore) {
		return 0
	}
	return e.matchScore[i]
}

// SetMatchScore プレイヤーの試合累積得点設定 (テスト用)
func (e *Ecarte) SetMatchScore(i, points int) {
	if i >= 0 && i < len(e.matchScore) {
		e.matchScore[i] = points
	}
}

// GetGameEndFlag ゲーム終了フラグ取得
func (e *Ecarte) GetGameEndFlag() bool { return e.gameEndFlag }

// GetWinnerIdx 勝者プレイヤーインデックス (-1: 未確定)
func (e *Ecarte) GetWinnerIdx() int { return e.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (e *Ecarte) GetPlayerCnt() int { return len(e.players) }

// GetPlayer プレイヤー取得
func (e *Ecarte) GetPlayer(i int) *EcartePlayer {
	return getPlayer(e.players, i)
}

// GetStockRemaining 山札の残り枚数
func (e *Ecarte) GetStockRemaining() int { return e.trumpCards.GetRemainingCount() }

// IsHumanTurn 現在の手番が人間かどうか
func (e *Ecarte) IsHumanTurn() bool {
	return isHumanTurn(e.players, e.currentPlayerIdx)
}

// GetConfig 設定取得
func (e *Ecarte) GetConfig() EcarteConfig { return e.config }

// SetConfig 設定変更
func (e *Ecarte) SetConfig(cfg EcarteConfig) { e.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (e *Ecarte) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(e.players) || e.phase != EcartePhasePlay {
		return nil
	}
	return e.legalPlayIndices(playerIdx)
}

// GetHint 人間プレイヤーへのヒントを取得する
func (e *Ecarte) GetHint() *EcarteHint {
	if e.currentPlayerIdx != 0 {
		return nil
	}
	switch e.phase {
	case EcartePhaseExchange:
		switch e.negStep {
		case EcarteNegElderDecide:
			if e.cpuWantsExchange(0) {
				return &EcarteHint{Action: "propose", Reason: "weak_hand"}
			}
			return &EcarteHint{Action: "stand", Reason: "strong_hand"}
		case EcarteNegDealerRespond:
			if e.cpuWantsExchange(0) {
				return &EcarteHint{Action: "accept", Reason: "weak_hand"}
			}
			return &EcarteHint{Action: "refuse", Reason: "strong_hand"}
		default:
			return &EcarteHint{Action: "discard", Reason: "exchange_weak"}
		}
	case EcartePhasePlay:
		if e.players[0].GetCardsSize() == 0 {
			return nil
		}
		idx := e.cpuSelectPlayCard(0)
		return &EcarteHint{CardIndex: &idx, Reason: e.playHintReason(0, idx)}
	}
	return nil
}

// --- Sorting / helpers ---

func (e *Ecarte) sortAllHands() {
	for _, p := range e.players {
		e.sortHand(p)
	}
}

func (e *Ecarte) sortHand(p *EcartePlayer) {
	trumpSuit := e.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return EcarteRankOrder(ci) < EcarteRankOrder(cj)
	})
}

func (e *Ecarte) playHintReason(playerIdx, chosenIdx int) string {
	card := e.players[playerIdx].GetCard(chosenIdx)
	if len(e.currentTrick) == 0 {
		if card.GetDesign() == e.trumpSuit {
			return "lead_trump"
		}
		return "lead_high"
	}
	leadCard := e.currentTrick[0].Card
	if ecarteBeats(card, leadCard, leadCard.GetDesign(), e.trumpSuit) {
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI ---

// cpuWantsExchange CPU が手札を弱いと判断して交換したいかを返す。
// 切り札枚数 + 高位札 (K,Q,J) が 2 未満なら交換を望む。
func (e *Ecarte) cpuWantsExchange(playerIdx int) bool {
	return e.handStrength(playerIdx) < 2
}

// handStrength 切り札 + 高位札 (K,Q,J) のおおまかな強さ。
func (e *Ecarte) handStrength(playerIdx int) int {
	p := e.players[playerIdx]
	score := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == e.trumpSuit {
			score++
		}
		if c.GetValue() == 13 || c.GetValue() == 12 || c.GetValue() == 11 {
			score++
		}
	}
	return score
}

// cpuChooseDiscards 切り札と高位札を残し、それ以外を捨てるインデックスを返す。
func (e *Ecarte) cpuChooseDiscards(playerIdx int) []int {
	p := e.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		keep := c.GetDesign() == e.trumpSuit || c.GetValue() == 13 || c.GetValue() == 12 || c.GetValue() == 11
		if !keep {
			out = append(out, i)
		}
	}
	// 山札の残り枚数を超えては引き直せない。
	if avail := e.trumpCards.GetRemainingCount(); len(out) > avail {
		out = out[:avail]
	}
	return out
}

// cpuSelectPlayCard CPU が出すカードのインデックスを選ぶ (合法手から)。
func (e *Ecarte) cpuSelectPlayCard(playerIdx int) int {
	legal := e.legalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return 0
	}
	if len(legal) == 1 {
		return legal[0]
	}
	p := e.players[playerIdx]
	if len(e.currentTrick) == 0 {
		// リード: 最も強い非切り札、無ければ最強。
		return e.pickStrongest(p, legal)
	}
	// 追随: 勝てる最弱、無ければ最弱を捨てる。
	leadCard := e.currentTrick[0].Card
	winIdx, dumpIdx := -1, legal[0]
	for _, i := range legal {
		c := p.GetCard(i)
		if ecarteBeats(c, leadCard, leadCard.GetDesign(), e.trumpSuit) {
			if winIdx < 0 || EcarteRankOrder(c) < EcarteRankOrder(p.GetCard(winIdx)) {
				winIdx = i
			}
		}
		if EcarteRankOrder(c) < EcarteRankOrder(p.GetCard(dumpIdx)) {
			dumpIdx = i
		}
	}
	if winIdx >= 0 {
		return winIdx
	}
	return dumpIdx
}

func (e *Ecarte) pickStrongest(p *EcartePlayer, legal []int) int {
	best := legal[0]
	for _, i := range legal[1:] {
		if EcarteRankOrder(p.GetCard(i)) > EcarteRankOrder(p.GetCard(best)) {
			best = i
		}
	}
	return best
}

// --- JSON ---

type ecarteJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*EcartePlayer   `json:"ps"`
	Config           EcarteConfig      `json:"cf"`
	Phase            EcartePhase       `json:"ph"`
	NegStep          EcarteNegStep     `json:"ns"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	TrumpCard        *Card             `json:"tu"`
	TrumpSuit        int               `json:"ts"`
	DealerIdx        int               `json:"di"`
	LeadPlayerIdx    int               `json:"li"`
	DealPoints       []int             `json:"dp"`
	MatchScore       []int             `json:"ms"`
	RefusalByDealer  bool              `json:"rf"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (e *Ecarte) MarshalJSON() ([]byte, error) {
	return json.Marshal(ecarteJSON{
		TrumpCards:       e.trumpCards,
		Players:          e.players,
		Config:           e.config,
		Phase:            e.phase,
		NegStep:          e.negStep,
		RoundNumber:      e.roundNumber,
		TrickNumber:      e.trickNumber,
		CurrentPlayerIdx: e.currentPlayerIdx,
		CurrentTrick:     e.currentTrick,
		TrumpCard:        e.trumpCard,
		TrumpSuit:        e.trumpSuit,
		DealerIdx:        e.dealerIdx,
		LeadPlayerIdx:    e.leadPlayerIdx,
		DealPoints:       e.dealPoints,
		MatchScore:       e.matchScore,
		RefusalByDealer:  e.refusalByDealer,
		GameEndFlag:      e.gameEndFlag,
		WinnerIdx:        e.winnerIdx,
		ActionLog:        e.actionLog,
	})
}

const ecarteMaxSliceLen = 5000

var errEcarteSnapshot = errors.New("ecarte: invalid serialised game state")

func ecarteIdxInRange(i int) bool { return i >= 0 && i < EcartePlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (e *Ecarte) UnmarshalJSON(data []byte) error {
	var j ecarteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != EcartePlayerCnt || len(j.CurrentTrick) > EcartePlayerCnt ||
		(j.DealPoints != nil && len(j.DealPoints) != EcartePlayerCnt) ||
		(j.MatchScore != nil && len(j.MatchScore) != EcartePlayerCnt) ||
		len(j.ActionLog) > ecarteMaxSliceLen ||
		!ecarteIdxInRange(j.CurrentPlayerIdx) || !ecarteIdxInRange(j.DealerIdx) ||
		!ecarteIdxInRange(j.LeadPlayerIdx) ||
		j.TrumpSuit < 0 || j.TrumpSuit > CardDesignDiamond ||
		j.RoundNumber < 1 ||
		j.Phase < EcartePhaseExchange || j.Phase > EcartePhaseGameEnd ||
		j.NegStep < EcarteNegElderDecide || j.NegStep > EcarteNegDealerDiscard {
		return errEcarteSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errEcarteSnapshot
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || !ecarteIdxInRange(tc.PlayerIdx) {
			return errEcarteSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errEcarteSnapshot
		}
	}
	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCardsBelote()
	}
	e.players = j.Players
	e.config = j.Config
	e.phase = j.Phase
	e.negStep = j.NegStep
	e.roundNumber = j.RoundNumber
	e.trickNumber = j.TrickNumber
	e.currentPlayerIdx = j.CurrentPlayerIdx
	e.currentTrick = j.CurrentTrick
	if e.currentTrick == nil {
		e.currentTrick = make([]*TrickCard, 0)
	}
	e.trumpCard = j.TrumpCard
	e.trumpSuit = j.TrumpSuit
	e.dealerIdx = j.DealerIdx
	e.leadPlayerIdx = j.LeadPlayerIdx
	e.dealPoints = ecarteEnsureLen(j.DealPoints)
	e.matchScore = ecarteEnsureLen(j.MatchScore)
	e.refusalByDealer = j.RefusalByDealer
	e.gameEndFlag = j.GameEndFlag
	e.winnerIdx = j.WinnerIdx
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

func ecarteEnsureLen(s []int) []int {
	if len(s) == EcartePlayerCnt {
		return s
	}
	return make([]int, EcartePlayerCnt)
}
