//go:build !js || !wasm || solo

// Package domain バルブ (Barbu) のドメインモデル。
//
// Barbu はフランス発祥のコンペンディウム (オムニバス) 型トリックテイキング
// ゲーム。4 人・52 枚デッキで、各プレイヤーが順番にディーラーを務め、計 28
// ディール (各プレイヤー 7 回ずつ) を行う。ディーラーは手札を見たうえで、
// 7 つのコントラクト (No Tricks / No Hearts / No Queens / Barbu(K♥) /
// No Last Trick / Trumps / Dominoes) から 1 つを選ぶ。各コントラクトは
// ディーラー 1 人につき 1 回ずつしか選べない。負のコントラクトは失点を、
// Trumps / Dominoes は加点を与え、28 ディール後の累計が最高のプレイヤーが
// 勝者となる。本実装は Hearts / Whist のトリック処理と Sevens のレイアウト
// 処理を統合・再利用する「クリーンアーキテクチャの集大成」である。
package domain

import (
	"encoding/json"
	"fmt"
)

// BarbuPlayerCnt はバルブのプレイヤー数 (固定 4)。
const BarbuPlayerCnt = 4

// BarbuHandSize は各プレイヤーの手札枚数 (52 / 4)。
const BarbuHandSize = 13

// BarbuContractCnt はコントラクトの種類数。
const BarbuContractCnt = 7

// BarbuTotalDeals は 1 ゲームの総ディール数 (各プレイヤー 7 コントラクト)。
const BarbuTotalDeals = BarbuPlayerCnt * BarbuContractCnt

// バルブのコントラクト定数。
const (
	// BarbuContractNoTricks トリックを取るとマイナス。
	BarbuContractNoTricks = 0
	// BarbuContractNoHearts ハートを取るとマイナス。
	BarbuContractNoHearts = 1
	// BarbuContractNoQueens Q を取るとマイナス。
	BarbuContractNoQueens = 2
	// BarbuContractKingHeart K♥ (ひげおじさん) を取ると大幅マイナス。
	BarbuContractKingHeart = 3
	// BarbuContractNoLastTrick 最後のトリックを取るとマイナス。
	BarbuContractNoLastTrick = 4
	// BarbuContractTrumps 切り札ありの通常トリックテイキング (プラス)。
	BarbuContractTrumps = 5
	// BarbuContractDominoes 7 並べ。早く上がるとプラス。
	BarbuContractDominoes = 6
)

// バルブのフェーズ。
const (
	// BarbuPhaseSelectContract ディーラーがコントラクトを選択中。
	BarbuPhaseSelectContract = "selectContract"
	// BarbuPhasePlay プレイ中 (トリックまたは 7 並べ)。
	BarbuPhasePlay = "play"
	// BarbuPhaseDealEnd 1 ディール終了 (得点確定、次ディール待ち)。
	BarbuPhaseDealEnd = "dealEnd"
	// BarbuPhaseGameEnd ゲーム終了 (28 ディール完了)。
	BarbuPhaseGameEnd = "gameEnd"
)

// BarbuDealDetail は 1 ディールの結果内訳。
type BarbuDealDetail struct {
	Contract  int         // 選択されたコントラクト
	TrumpSuit int         // 切り札スート (Trumps のみ。それ以外は -1)
	DealerIdx int         // ディーラー
	Gained    map[int]int // プレイヤー別にこのディールで得た得点
}

// Barbu はバルブゲームの状態を保持する集約ルート。
type Barbu struct {
	trumpCards      *TrumpCards
	players         []*BarbuPlayer
	config          BarbuConfig
	phase           string
	dealNumber      int // 0..BarbuTotalDeals (= 現在のディール index)
	dealerIdx       int
	currentContract int // -1 = 未選択
	trumpSuit       int // -1 = 切り札なし
	usedContracts   [BarbuPlayerCnt][BarbuContractCnt]bool
	currentPlayer   int
	leadPlayer      int
	trickNumber     int          // 1..BarbuHandSize
	currentTrick    []*TrickCard // 進行中のトリック
	lastTrick       []*TrickCard // 直前に完了したトリック (UI 表示用)
	lastTrickWinner int          // 直前トリックの勝者 (-1 = なし)
	tablePlaced     [5]uint16    // Dominoes の場 (index 1-4 = スート)
	passCount       [BarbuPlayerCnt]int
	dominoFinished  int // Dominoes で上がった人数
	gameEndFlag     bool
	lastDealDetail  *BarbuDealDetail
	dealHistory     []*BarbuDealDetail // 完了した各ディールの得点内訳 (最大 BarbuTotalDeals 件)
	actionLogBase
}

// NewBarbu はコンストラクタ。
func NewBarbu(trumpCards *TrumpCards, players []*BarbuPlayer, config BarbuConfig) *Barbu {
	return &Barbu{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		phase:           BarbuPhaseSelectContract,
		currentContract: -1,
		trumpSuit:       -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultBarbu は標準の 4 人構成 (1 human + 3 CPU) と DefaultBarbuConfig で
// Barbu を生成する。CUI / Web / Worker の構築の単一情報源。
func NewDefaultBarbu() *Barbu {
	players := make([]*BarbuPlayer, BarbuPlayerCnt)
	players[0] = NewBarbuPlayer(true)
	for i := 1; i < BarbuPlayerCnt; i++ {
		players[i] = NewBarbuPlayer(false)
	}
	return NewBarbu(NewTrumpCards(0), players, DefaultBarbuConfig())
}

// Reset は新しいゲームを開始する。累計得点もクリアする。
func (b *Barbu) Reset() {
	for _, p := range b.players {
		p.ResetDeal()
		p.ResetTotalScore()
	}
	b.dealNumber = 0
	b.gameEndFlag = false
	b.usedContracts = [BarbuPlayerCnt][BarbuContractCnt]bool{}
	b.lastDealDetail = nil
	b.dealHistory = make([]*BarbuDealDetail, 0, BarbuTotalDeals)
	b.actionLog = make([]*ActionLogEntry, 0)
	b.startDeal()
}

// NextDeal は次のディールを開始する。28 ディール完了済みなら何もしない。
func (b *Barbu) NextDeal() {
	if b.gameEndFlag || b.phase != BarbuPhaseDealEnd {
		return
	}
	b.dealNumber++
	if b.dealNumber >= BarbuTotalDeals {
		b.gameEndFlag = true
		b.phase = BarbuPhaseGameEnd
		b.appendLog(-1, "gameEnd", "all 28 deals completed", nil)
		return
	}
	b.startDeal()
}

// startDeal はディーラーを決め、手札を配り、コントラクト選択フェーズへ移る。
func (b *Barbu) startDeal() {
	b.dealerIdx = b.dealNumber % BarbuPlayerCnt
	b.currentContract = -1
	b.trumpSuit = -1
	b.currentTrick = nil
	b.lastTrick = nil
	b.lastTrickWinner = -1
	b.trickNumber = 0
	b.tablePlaced = [5]uint16{}
	b.passCount = [BarbuPlayerCnt]int{}
	b.dominoFinished = 0
	for _, p := range b.players {
		p.ResetDeal()
	}
	b.trumpCards = NewTrumpCards(0)
	b.trumpCards.Shuffle()
	dealAllCards(b.trumpCards, b.players)
	for _, p := range b.players {
		barbuSortHand(p)
	}
	b.phase = BarbuPhaseSelectContract
	b.appendLog(-1, "deal", fmt.Sprintf("deal %d/%d, dealer=%d", b.dealNumber+1, BarbuTotalDeals, b.dealerIdx), nil)
}

// SelectContract はディーラーがコントラクトを選択する。
// trumpSuit は Trumps コントラクトでのみ使用する (1-4)。それ以外は無視。
func (b *Barbu) SelectContract(contract, trumpSuit int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BarbuPhaseSelectContract {
		return NewDomainError(ErrWrongPhase, "not in contract-selection phase")
	}
	if !b.players[b.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return b.applySelectContract(contract, trumpSuit)
}

// applySelectContract はコントラクト選択の共通処理 (human / CPU)。
func (b *Barbu) applySelectContract(contract, trumpSuit int) error {
	if contract < 0 || contract >= BarbuContractCnt {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("unknown contract %d", contract))
	}
	if b.usedContracts[b.dealerIdx][contract] {
		return NewDomainError(ErrInvalidPlay, "contract already used by this dealer")
	}
	if contract == BarbuContractTrumps {
		if trumpSuit < CardDesignSpade || trumpSuit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "trumps contract requires a valid trump suit")
		}
		b.trumpSuit = trumpSuit
	} else {
		b.trumpSuit = -1
	}
	b.usedContracts[b.dealerIdx][contract] = true
	b.currentContract = contract
	b.phase = BarbuPhasePlay
	b.leadPlayer = b.dealerIdx
	b.currentPlayer = b.dealerIdx
	b.trickNumber = 1
	b.appendLog(b.dealerIdx, "selectContract",
		fmt.Sprintf("dealer %d selects %s", b.dealerIdx, barbuContractName(contract)), nil)
	return nil
}

// PlayerPlay は人間プレイヤーの 1 手。
// トリックコントラクトでは handIdx の手札を出す (tableIdxs は無視)。
// Dominoes では handIdx の手札を場に置く。handIdx == -1 はパス (合法手がない場合のみ)。
func (b *Barbu) PlayerPlay(handIdx int, tableIdxs []int) error {
	_ = tableIdxs // トリック/Dominoes ともに場の選択は不要 (互換のため受け取る)
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BarbuPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !b.players[b.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if b.currentContract == BarbuContractDominoes {
		return b.applyDominoPlay(b.currentPlayer, handIdx)
	}
	return b.applyTrickPlay(b.currentPlayer, handIdx)
}

// CpuPlay は CPU のターンを 1 ステップ進める。
// コントラクト選択フェーズではディーラー (CPU) がコントラクトを選び、
// プレイフェーズでは現在の手番 (CPU) が 1 手プレイする。
func (b *Barbu) CpuPlay() {
	if b.gameEndFlag {
		return
	}
	switch b.phase {
	case BarbuPhaseSelectContract:
		if b.players[b.dealerIdx].GetIsHuman() {
			return
		}
		contract, trump := b.cpuSelectContract(b.dealerIdx)
		_ = b.applySelectContract(contract, trump)
	case BarbuPhasePlay:
		if b.players[b.currentPlayer].GetIsHuman() {
			return
		}
		if b.currentContract == BarbuContractDominoes {
			b.cpuDominoPlay(b.currentPlayer)
		} else {
			idx := b.cpuSelectTrickCard(b.currentPlayer)
			_ = b.applyTrickPlay(b.currentPlayer, idx)
		}
	}
}

// applyTrickPlay はトリックコントラクトでの 1 手を実行する。
func (b *Barbu) applyTrickPlay(playerIdx, handIdx int) error {
	player := b.players[playerIdx]
	card := player.GetCard(handIdx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if err := b.validateTrickPlay(playerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(handIdx)
	b.currentTrick = append(b.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: played})
	b.appendLog(playerIdx, "play", fmt.Sprintf("player %d plays %s", playerIdx, cardStr(played)), []*Card{played})

	if len(b.currentTrick) == BarbuPlayerCnt {
		b.resolveTrick()
	} else {
		b.currentPlayer = (b.currentPlayer + 1) % BarbuPlayerCnt
	}
	return nil
}

// GetPlayableIndices はそのプレイヤーがいま出せる手札の位置を返す。
//
// **判定は validateTrickPlay をそのまま通す。**別のスキャンを書くと「出せる」と
// 見えた札がサーバーに弾かれる。Web はドミノ契約以外でフォロー義務を可視化して
// いなかった (#4804)。
func (b *Barbu) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	p := b.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if b.validateTrickPlay(playerIdx, p.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// validateTrickPlay はトリックプレイの合法性 (フォロースート) を検証する。
func (b *Barbu) validateTrickPlay(playerIdx int, card *Card) error {
	if len(b.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && b.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "must follow the lead suit")
	}
	return nil
}

// playerHasSuit はプレイヤーが指定スートを持っているか。
func (b *Barbu) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(b.players[playerIdx], design)
}

// resolveTrick は完成したトリックの勝者を決定し、次トリックへ進める。
func (b *Barbu) resolveTrick() {
	winner := b.trickWinner()
	cards := make([]*Card, len(b.currentTrick))
	for i, tc := range b.currentTrick {
		cards[i] = tc.Card
	}
	b.players[winner].AddTrick(cards)
	b.lastTrick = b.currentTrick
	b.lastTrickWinner = winner
	b.appendLog(winner, "trickWin", fmt.Sprintf("player %d wins trick %d", winner, b.trickNumber), cards)

	b.currentTrick = nil
	b.leadPlayer = winner
	b.currentPlayer = winner
	if b.trickNumber >= BarbuHandSize {
		b.finishDeal()
		return
	}
	b.trickNumber++
}

// trickWinner はトリックの勝者を決定する。Trumps コントラクトでは切り札が
// 最強。それ以外はリードスートの最高位が勝つ。
func (b *Barbu) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return b.leadPlayer
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerCard := b.currentTrick[0].Card
	for _, tc := range b.currentTrick[1:] {
		if b.cardBeats(tc.Card, winnerCard, leadSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// barbuCardStrength はトリック比較用の強さを返す。A(1) → 14、それ以外は値そのまま。
func barbuCardStrength(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// cardBeats は candidate が current より強いかを判定する。
func (b *Barbu) cardBeats(candidate, current *Card, leadSuit int) bool {
	isTrump := b.trumpSuit >= CardDesignSpade
	candTrump := isTrump && candidate.GetDesign() == b.trumpSuit
	curTrump := isTrump && current.GetDesign() == b.trumpSuit
	switch {
	case candTrump && !curTrump:
		return true
	case !candTrump && curTrump:
		return false
	case candTrump && curTrump:
		return barbuCardStrength(candidate) > barbuCardStrength(current)
	default:
		// どちらも切り札でない: リードスートのみ勝負に絡む。
		if candidate.GetDesign() != leadSuit {
			return false
		}
		if current.GetDesign() != leadSuit {
			return true
		}
		return barbuCardStrength(candidate) > barbuCardStrength(current)
	}
}

// finishDeal はディール終了処理: 得点を計算して累計に加算する。
func (b *Barbu) finishDeal() {
	detail := b.scoreDeal()
	b.lastDealDetail = detail
	b.dealHistory = append(b.dealHistory, detail)
	for i, p := range b.players {
		p.AddScore(detail.Gained[i])
	}
	b.phase = BarbuPhaseDealEnd
	b.appendLog(-1, "dealEnd",
		fmt.Sprintf("deal %d scored (%s)", b.dealNumber+1, barbuContractName(b.currentContract)), nil)
	// 28 ディール目が終わったら、NextDeal を待たずにゲーム終了とする。
	if b.dealNumber >= BarbuTotalDeals-1 {
		b.gameEndFlag = true
		b.phase = BarbuPhaseGameEnd
		b.appendLog(-1, "gameEnd", "all 28 deals completed", nil)
	}
}

// barbuSortHand はプレイヤーの手札をスート→値の順にソートする。
func barbuSortHand(p *BarbuPlayer) {
	n := p.GetCardsSize()
	cards := make([]*Card, n)
	for i := 0; i < n; i++ {
		cards[i] = p.GetCard(i)
	}
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0; j-- {
			a, c := cards[j-1], cards[j]
			less := c.GetDesign() < a.GetDesign() ||
				(c.GetDesign() == a.GetDesign() && c.GetValue() < a.GetValue())
			if !less {
				break
			}
			cards[j-1], cards[j] = cards[j], cards[j-1]
		}
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// barbuContractName はコントラクトの英語名を返す (ログ用)。
func barbuContractName(c int) string {
	switch c {
	case BarbuContractNoTricks:
		return "No Tricks"
	case BarbuContractNoHearts:
		return "No Hearts"
	case BarbuContractNoQueens:
		return "No Queens"
	case BarbuContractKingHeart:
		return "Barbu"
	case BarbuContractNoLastTrick:
		return "No Last Trick"
	case BarbuContractTrumps:
		return "Trumps"
	case BarbuContractDominoes:
		return "Dominoes"
	default:
		return "Unknown"
	}
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の意思決定者が人間かを返す。
func (b *Barbu) IsHumanTurn() bool {
	if b.gameEndFlag {
		return false
	}
	switch b.phase {
	case BarbuPhaseSelectContract:
		return b.players[b.dealerIdx].GetIsHuman()
	case BarbuPhasePlay:
		return b.players[b.currentPlayer].GetIsHuman()
	default:
		return false
	}
}

// GetGameEndFlag はゲーム終了フラグを返す。
func (b *Barbu) GetGameEndFlag() bool { return b.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (b *Barbu) GetPhase() string { return b.phase }

// GetPlayerCnt はプレイヤー数を返す。
func (b *Barbu) GetPlayerCnt() int { return len(b.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (b *Barbu) GetPlayer(i int) *BarbuPlayer {
	return getPlayer(b.players, i)
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (b *Barbu) GetCurrentTurn() int { return b.currentPlayer }

// GetDealerIdx は現在のディーラーを返す。
func (b *Barbu) GetDealerIdx() int { return b.dealerIdx }

// GetDealNumber は現在のディール番号 (0-indexed) を返す。
func (b *Barbu) GetDealNumber() int { return b.dealNumber }

// GetCurrentContract は現在のコントラクト (-1 = 未選択) を返す。
func (b *Barbu) GetCurrentContract() int { return b.currentContract }

// GetTrumpSuit は切り札スート (-1 = なし) を返す。
func (b *Barbu) GetTrumpSuit() int { return b.trumpSuit }

// GetTrickNumber は現在のトリック番号を返す。
func (b *Barbu) GetTrickNumber() int { return b.trickNumber }

// GetCurrentTrick は進行中のトリックを返す。
func (b *Barbu) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// GetLastTrick は直前に完了したトリックを返す。
func (b *Barbu) GetLastTrick() []*TrickCard { return b.lastTrick }

// GetLastTrickWinner は直前トリックの勝者を返す (-1 = なし)。
func (b *Barbu) GetLastTrickWinner() int { return b.lastTrickWinner }

// GetTablePlaced は Dominoes の場の状態を返す (index 1-4 = スートのビットマスク)。
func (b *Barbu) GetTablePlaced() [5]uint16 { return b.tablePlaced }

// GetUsedContracts は指定ディーラーが使用済みのコントラクトを返す。
func (b *Barbu) GetUsedContracts(dealerIdx int) [BarbuContractCnt]bool {
	if dealerIdx < 0 || dealerIdx >= BarbuPlayerCnt {
		return [BarbuContractCnt]bool{}
	}
	return b.usedContracts[dealerIdx]
}

// GetLastDealDetail は直前ディールの得点内訳を返す (nil の場合もある)。
func (b *Barbu) GetLastDealDetail() *BarbuDealDetail { return b.lastDealDetail }

// GetDealHistory は完了した各ディールの得点内訳を古い順に返す。
func (b *Barbu) GetDealHistory() []*BarbuDealDetail { return b.dealHistory }

// GetConfig はローカルルール設定を返す。
func (b *Barbu) GetConfig() BarbuConfig { return b.config }

// SetConfig はローカルルール設定を変更する。
func (b *Barbu) SetConfig(config BarbuConfig) { b.config = config }

// GetRoundWinners は (ゲーム終了時) 最高得点プレイヤーのリストを返す。
func (b *Barbu) GetRoundWinners() []int {
	if !b.gameEndFlag {
		return nil
	}
	return topScorers(b.players)
}

// --- JSON Serialization ---

// barbuJSON is the JSON wire format for Barbu.
type barbuJSON struct {
	TrumpCards      *TrumpCards                            `json:"tc"`
	Players         []*BarbuPlayer                         `json:"pl"`
	Config          BarbuConfig                            `json:"cf"`
	Phase           string                                 `json:"ph"`
	DealNumber      int                                    `json:"dn"`
	DealerIdx       int                                    `json:"di"`
	CurrentContract int                                    `json:"cc"`
	TrumpSuit       int                                    `json:"ts"`
	UsedContracts   [BarbuPlayerCnt][BarbuContractCnt]bool `json:"uc"`
	CurrentPlayer   int                                    `json:"cp"`
	LeadPlayer      int                                    `json:"lp"`
	TrickNumber     int                                    `json:"tn"`
	CurrentTrick    []*TrickCard                           `json:"ct"`
	LastTrick       []*TrickCard                           `json:"lt"`
	LastTrickWinner int                                    `json:"lw"`
	TablePlaced     [5]uint16                              `json:"tb"`
	PassCount       [BarbuPlayerCnt]int                    `json:"pc"`
	DominoFinished  int                                    `json:"df"`
	GameEndFlag     bool                                   `json:"ge"`
	LastDealDetail  *BarbuDealDetail                       `json:"ld"`
	DealHistory     []*BarbuDealDetail                     `json:"dh"`
	ActionLog       []*ActionLogEntry                      `json:"al"`
}

// barbuMaxSliceLen caps slice sizes during deserialisation.
const barbuMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (b *Barbu) MarshalJSON() ([]byte, error) {
	return json.Marshal(barbuJSON{
		TrumpCards:      b.trumpCards,
		Players:         b.players,
		Config:          b.config,
		Phase:           b.phase,
		DealNumber:      b.dealNumber,
		DealerIdx:       b.dealerIdx,
		CurrentContract: b.currentContract,
		TrumpSuit:       b.trumpSuit,
		UsedContracts:   b.usedContracts,
		CurrentPlayer:   b.currentPlayer,
		LeadPlayer:      b.leadPlayer,
		TrickNumber:     b.trickNumber,
		CurrentTrick:    b.currentTrick,
		LastTrick:       b.lastTrick,
		LastTrickWinner: b.lastTrickWinner,
		TablePlaced:     b.tablePlaced,
		PassCount:       b.passCount,
		DominoFinished:  b.dominoFinished,
		GameEndFlag:     b.gameEndFlag,
		LastDealDetail:  b.lastDealDetail,
		DealHistory:     b.dealHistory,
		ActionLog:       b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Barbu) UnmarshalJSON(data []byte) error {
	var j barbuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > barbuMaxSliceLen || len(j.CurrentTrick) > barbuMaxSliceLen ||
		len(j.LastTrick) > barbuMaxSliceLen || len(j.ActionLog) > barbuMaxSliceLen ||
		len(j.DealHistory) > barbuMaxSliceLen {
		return fmt.Errorf("barbu: input array exceeds maximum allowed size")
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("barbu: missing trump cards in state")
	}
	b.trumpCards = j.TrumpCards
	b.players = j.Players
	if b.players == nil {
		b.players = make([]*BarbuPlayer, 0)
	}
	if len(b.players) != BarbuPlayerCnt {
		return fmt.Errorf("barbu: invalid player count %d, expected %d", len(b.players), BarbuPlayerCnt)
	}
	b.config = j.Config
	b.phase = j.Phase
	b.dealNumber = j.DealNumber
	b.dealerIdx = j.DealerIdx
	b.currentContract = j.CurrentContract
	b.trumpSuit = j.TrumpSuit
	b.usedContracts = j.UsedContracts
	b.currentPlayer = j.CurrentPlayer
	b.leadPlayer = j.LeadPlayer
	b.trickNumber = j.TrickNumber
	b.currentTrick = j.CurrentTrick
	b.lastTrick = j.LastTrick
	b.lastTrickWinner = j.LastTrickWinner
	b.tablePlaced = j.TablePlaced
	b.passCount = j.PassCount
	b.dominoFinished = j.DominoFinished
	b.gameEndFlag = j.GameEndFlag
	b.lastDealDetail = j.LastDealDetail
	b.dealHistory = j.dealHistoryOrEmpty()
	b.actionLog = j.actionLogOrEmpty()
	return nil
}

// actionLogOrEmpty returns the action log, defaulting to an empty slice.
func (j barbuJSON) actionLogOrEmpty() []*ActionLogEntry {
	if j.ActionLog == nil {
		return make([]*ActionLogEntry, 0)
	}
	return j.ActionLog
}

// dealHistoryOrEmpty returns the deal history, defaulting to an empty slice.
func (j barbuJSON) dealHistoryOrEmpty() []*BarbuDealDetail {
	if j.DealHistory == nil {
		return make([]*BarbuDealDetail, 0, BarbuTotalDeals)
	}
	return j.DealHistory
}

// barbuTrickCardJSON / BarbuDealDetail serialization ------------------------

// barbuDealDetailJSON is the JSON wire format for BarbuDealDetail.
type barbuDealDetailJSON struct {
	Contract  int         `json:"co"`
	TrumpSuit int         `json:"ts"`
	DealerIdx int         `json:"di"`
	Gained    map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *BarbuDealDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(barbuDealDetailJSON{
		Contract:  d.Contract,
		TrumpSuit: d.TrumpSuit,
		DealerIdx: d.DealerIdx,
		Gained:    d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *BarbuDealDetail) UnmarshalJSON(data []byte) error {
	var j barbuDealDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Contract = j.Contract
	d.TrumpSuit = j.TrumpSuit
	d.DealerIdx = j.DealerIdx
	d.Gained = j.Gained
	return nil
}
