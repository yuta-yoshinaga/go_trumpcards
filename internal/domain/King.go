//go:build !js || !wasm || extra

// Package domain キング (King) のドメインモデル。
//
// King はギリシャ・ブラジルで親しまれるコンペンディウム (オムニバス) 型トリック
// テイキングゲーム。4 人・52 枚デッキで、各ディールの親 (ディーラー) がまだ選ばれて
// いないコントラクトを 1 つ選び、13 トリックをマストフォローでプレイする。全 7 種の
// コントラクトを 1 巡 (計 7 ディール) 行い、累計得点が最も高い (= 失点が最も少ない)
// プレイヤーが勝者となる。
//
// コントラクト:
//   - No Tricks    (負) : 取ったトリック 1 つにつき減点。
//   - No Hearts    (負) : 取ったハート 1 枚につき減点。
//   - No Queens    (負) : 取った Q 1 枚につき減点。
//   - No King♥     (負) : K♥ を取ると大幅減点。
//   - No Last Two  (負) : 最後の 2 トリック (12・13 番目) を取ると減点。
//   - No Men       (負) : 取った J / K 1 枚につき減点。
//   - King (Trump) (正) : 親が切り札を選び、取ったトリック 1 つにつき加点。
//
// 負のコントラクトは切り札なし、King (Trump) のみ親が切り札を指定する。トリック強度は
// A > K > Q > J > 10 > ... > 2 (Barbu 同様 A を最強とみなす)。本実装は extra ワーカーから
// 到達可能なようコントラクト・得点・CPU ロジックをすべてインラインで持つ (Barbu の
// solo タグ実装には依存しない)。
package domain

import (
	"encoding/json"
	"fmt"
)

// KingPlayerCnt はキングのプレイヤー数 (固定 4)。
const KingPlayerCnt = 4

// KingHandSize は各プレイヤーの手札枚数 (52 / 4)。
const KingHandSize = 13

// KingContractCnt はコントラクトの種類数。
const KingContractCnt = 7

// KingTotalDeals は 1 ゲームの総ディール数 (全コントラクトを 1 巡)。
const KingTotalDeals = KingContractCnt

// キングのコントラクト定数。
const (
	// KingContractNoTricks トリックを取るとマイナス。
	KingContractNoTricks = 0
	// KingContractNoHearts ハートを取るとマイナス。
	KingContractNoHearts = 1
	// KingContractNoQueens Q を取るとマイナス。
	KingContractNoQueens = 2
	// KingContractKingHeart K♥ を取ると大幅マイナス。
	KingContractKingHeart = 3
	// KingContractNoLastTwo 最後の 2 トリックを取るとマイナス。
	KingContractNoLastTwo = 4
	// KingContractNoMen J / K を取るとマイナス。
	KingContractNoMen = 5
	// KingContractKingTrump 切り札ありの通常トリックテイキング (プラス)。
	KingContractKingTrump = 6
)

// キングのフェーズ。
const (
	// KingPhaseSelectContract 親がコントラクトを選択中。
	KingPhaseSelectContract = "selectContract"
	// KingPhasePlay トリックプレイ中。
	KingPhasePlay = "play"
	// KingPhaseDealEnd 1 ディール終了 (得点確定、次ディール待ち)。
	KingPhaseDealEnd = "dealEnd"
	// KingPhaseGameEnd ゲーム終了 (全コントラクト完了)。
	KingPhaseGameEnd = "gameEnd"
)

// キングの各コントラクトの得点定数 (1 ディールあたり)。負のコントラクトは減点、
// King (Trump) は加点を与える。
const (
	// KingNoTrickPenalty No Tricks: 取ったトリック 1 つにつき減点。
	KingNoTrickPenalty = 2
	// KingHeartPenalty No Hearts: 取ったハート 1 枚につき減点。
	KingHeartPenalty = 2
	// KingQueenPenalty No Queens: 取った Q 1 枚につき減点。
	KingQueenPenalty = 6
	// KingKingHeartPenalty No King♥: K♥ を取ると大幅減点。
	KingKingHeartPenalty = 20
	// KingLastTwoPenalty No Last Two: 最後の 2 トリックそれぞれにつき減点。
	KingLastTwoPenalty = 10
	// KingMenPenalty No Men: 取った J / K 1 枚につき減点。
	KingMenPenalty = 3
	// KingTrumpReward King (Trump): 取ったトリック 1 つにつき加点。
	KingTrumpReward = 5
)

// KingDealDetail は 1 ディールの結果内訳。
type KingDealDetail struct {
	Contract  int         // 選択されたコントラクト
	TrumpSuit int         // 切り札スート (King Trump のみ。それ以外は -1)
	DealerIdx int         // 親
	Gained    map[int]int // プレイヤー別にこのディールで得た得点
}

// King はキングゲームの状態を保持する集約ルート。
type King struct {
	trumpCards      *TrumpCards
	players         []*KingPlayer
	config          KingConfig
	phase           string
	dealNumber      int // 0..KingTotalDeals (= 現在のディール index)
	dealerIdx       int
	currentContract int // -1 = 未選択
	trumpSuit       int // -1 = 切り札なし
	usedContracts   [KingContractCnt]bool
	currentPlayer   int
	leadPlayer      int
	trickNumber     int          // 1..KingHandSize
	currentTrick    []*TrickCard // 進行中のトリック
	lastTrick       []*TrickCard // 直前に完了したトリック (UI 表示用)
	lastTrickWinner int          // 直前トリックの勝者 (-1 = なし)
	gameEndFlag     bool
	lastDealDetail  *KingDealDetail
	actionLogBase
}

// NewKing はコンストラクタ。
func NewKing(trumpCards *TrumpCards, players []*KingPlayer, config KingConfig) *King {
	return &King{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		phase:           KingPhaseSelectContract,
		currentContract: -1,
		trumpSuit:       -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultKing は標準の 4 人構成 (1 human + 3 CPU) と DefaultKingConfig で King を
// 生成する。CUI / Web / Worker の構築の単一情報源。
func NewDefaultKing() *King {
	players := make([]*KingPlayer, KingPlayerCnt)
	players[0] = NewKingPlayer(true)
	for i := 1; i < KingPlayerCnt; i++ {
		players[i] = NewKingPlayer(false)
	}
	return NewKing(newKingDeck(), players, DefaultKingConfig())
}

// newKingDeck はキング用 52 枚デッキを生成する。NewTrumpCards はビルドタグ無しの
// TrumpCards.go にあり extra ワーカーからも到達可能。
func newKingDeck() *TrumpCards {
	return NewTrumpCards(0)
}

// Reset は新しいゲームを開始する。累計得点もクリアする。
func (g *King) Reset() {
	for _, p := range g.players {
		p.ResetDeal()
		p.ResetTotalScore()
	}
	g.dealNumber = 0
	g.gameEndFlag = false
	g.usedContracts = [KingContractCnt]bool{}
	g.lastDealDetail = nil
	g.actionLog = make([]*ActionLogEntry, 0)
	g.startDeal()
}

// NextDeal は次のディールを開始する。全ディール完了済みなら何もしない。
func (g *King) NextDeal() {
	if g.gameEndFlag || g.phase != KingPhaseDealEnd {
		return
	}
	g.dealNumber++
	if g.dealNumber >= KingTotalDeals {
		g.gameEndFlag = true
		g.phase = KingPhaseGameEnd
		g.appendLog(-1, "gameEnd", "all contracts completed", nil)
		return
	}
	g.startDeal()
}

// startDeal は親を決め、手札を配り、コントラクト選択フェーズへ移る。
func (g *King) startDeal() {
	g.dealerIdx = g.dealNumber % KingPlayerCnt
	g.currentContract = -1
	g.trumpSuit = -1
	g.currentTrick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.trickNumber = 0
	for _, p := range g.players {
		p.ResetDeal()
	}
	g.trumpCards = newKingDeck()
	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	for _, p := range g.players {
		kingSortHand(p)
	}
	g.phase = KingPhaseSelectContract
	g.appendLog(-1, "deal", fmt.Sprintf("deal %d/%d, dealer=%d", g.dealNumber+1, KingTotalDeals, g.dealerIdx), nil)
}

// SelectContract は親がコントラクトを選択する。
// trumpSuit は King (Trump) コントラクトでのみ使用する (1-4)。それ以外は無視。
func (g *King) SelectContract(contract, trumpSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KingPhaseSelectContract {
		return NewDomainError(ErrWrongPhase, "not in contract-selection phase")
	}
	if !g.players[g.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applySelectContract(contract, trumpSuit)
}

// applySelectContract はコントラクト選択の共通処理 (human / CPU)。
func (g *King) applySelectContract(contract, trumpSuit int) error {
	if contract < 0 || contract >= KingContractCnt {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("unknown contract %d", contract))
	}
	if g.usedContracts[contract] {
		return NewDomainError(ErrInvalidPlay, "contract already used")
	}
	if contract == KingContractKingTrump {
		if trumpSuit < CardDesignSpade || trumpSuit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "king (trump) contract requires a valid trump suit")
		}
		g.trumpSuit = trumpSuit
	} else {
		g.trumpSuit = -1
	}
	g.usedContracts[contract] = true
	g.currentContract = contract
	g.phase = KingPhasePlay
	g.leadPlayer = g.dealerIdx
	g.currentPlayer = g.dealerIdx
	g.trickNumber = 1
	g.appendLog(g.dealerIdx, "selectContract",
		fmt.Sprintf("dealer %d selects %s", g.dealerIdx, kingContractName(contract)), nil)
	return nil
}

// PlayerPlay は人間プレイヤーの 1 手。handIdx の手札を出す。
func (g *King) PlayerPlay(handIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KingPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyTrickPlay(g.currentPlayer, handIdx)
}

// CpuPlay は CPU のターンを 1 ステップ進める。
// コントラクト選択フェーズでは親 (CPU) がコントラクトを選び、プレイフェーズでは
// 現在の手番 (CPU) が 1 手プレイする。
func (g *King) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case KingPhaseSelectContract:
		if g.players[g.dealerIdx].GetIsHuman() {
			return
		}
		contract, trump := g.cpuSelectContract()
		_ = g.applySelectContract(contract, trump)
	case KingPhasePlay:
		if g.players[g.currentPlayer].GetIsHuman() {
			return
		}
		idx := g.cpuSelectTrickCard(g.currentPlayer)
		_ = g.applyTrickPlay(g.currentPlayer, idx)
	}
}

// applyTrickPlay はトリックコントラクトでの 1 手を実行する。
func (g *King) applyTrickPlay(playerIdx, handIdx int) error {
	player := g.players[playerIdx]
	card := player.GetCard(handIdx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if err := g.validateTrickPlay(playerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(handIdx)
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: played})
	g.appendLog(playerIdx, "play", fmt.Sprintf("player %d plays %s", playerIdx, cardStr(played)), []*Card{played})

	if len(g.currentTrick) == KingPlayerCnt {
		g.resolveTrick()
	} else {
		g.currentPlayer = (g.currentPlayer + 1) % KingPlayerCnt
	}
	return nil
}

// validateTrickPlay はトリックプレイの合法性 (フォロースート) を検証する。
func (g *King) validateTrickPlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "card is nil")
	}
	if len(g.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "must follow the lead suit")
	}
	return nil
}

// playerHasSuit はプレイヤーが指定スートを持っているか。
func (g *King) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// resolveTrick は完成したトリックの勝者を決定し、次トリックへ進める。
func (g *King) resolveTrick() {
	winner := g.trickWinner()
	cards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		cards[i] = tc.Card
	}
	g.players[winner].AddTrickWithRank(cards, g.trickNumber)
	g.lastTrick = g.currentTrick
	g.lastTrickWinner = winner
	g.appendLog(winner, "trickWin", fmt.Sprintf("player %d wins trick %d", winner, g.trickNumber), cards)

	g.currentTrick = nil
	g.leadPlayer = winner
	g.currentPlayer = winner
	if g.trickNumber >= KingHandSize {
		g.finishDeal()
		return
	}
	g.trickNumber++
}

// trickWinner はトリックの勝者を決定する。King (Trump) では切り札が最強。
// それ以外はリードスートの最高位が勝つ。
func (g *King) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayer
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if g.cardBeats(tc.Card, winnerCard, leadSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// kingCardStrength はトリック比較用の強さを返す。A(1) → 14、それ以外は値そのまま。
func kingCardStrength(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// cardBeats は candidate が current より強いかを判定する。
func (g *King) cardBeats(candidate, current *Card, leadSuit int) bool {
	isTrump := g.trumpSuit >= CardDesignSpade
	candTrump := isTrump && candidate.GetDesign() == g.trumpSuit
	curTrump := isTrump && current.GetDesign() == g.trumpSuit
	switch {
	case candTrump && !curTrump:
		return true
	case !candTrump && curTrump:
		return false
	case candTrump && curTrump:
		return kingCardStrength(candidate) > kingCardStrength(current)
	default:
		// どちらも切り札でない: リードスートのみ勝負に絡む。
		if candidate.GetDesign() != leadSuit {
			return false
		}
		if current.GetDesign() != leadSuit {
			return true
		}
		return kingCardStrength(candidate) > kingCardStrength(current)
	}
}

// finishDeal はディール終了処理: 得点を計算して累計に加算する。
func (g *King) finishDeal() {
	detail := g.scoreDeal()
	g.lastDealDetail = detail
	for i, p := range g.players {
		p.AddScore(detail.Gained[i])
	}
	g.phase = KingPhaseDealEnd
	g.appendLog(-1, "dealEnd",
		fmt.Sprintf("deal %d scored (%s)", g.dealNumber+1, kingContractName(g.currentContract)), nil)
	// 最終ディールが終わったら、NextDeal を待たずにゲーム終了とする。
	if g.dealNumber >= KingTotalDeals-1 {
		g.gameEndFlag = true
		g.phase = KingPhaseGameEnd
		g.appendLog(-1, "gameEnd", "all contracts completed", nil)
	}
}

// kingSortHand はプレイヤーの手札をスート→値の順にソートする。
func kingSortHand(p *KingPlayer) {
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

// kingContractName はコントラクトの英語名を返す (ログ用)。
func kingContractName(c int) string {
	switch c {
	case KingContractNoTricks:
		return "No Tricks"
	case KingContractNoHearts:
		return "No Hearts"
	case KingContractNoQueens:
		return "No Queens"
	case KingContractKingHeart:
		return "No King of Hearts"
	case KingContractNoLastTwo:
		return "No Last Two Tricks"
	case KingContractNoMen:
		return "No Men"
	case KingContractKingTrump:
		return "King (Trump)"
	default:
		return "Unknown"
	}
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の意思決定者が人間かを返す。
func (g *King) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case KingPhaseSelectContract:
		return g.players[g.dealerIdx].GetIsHuman()
	case KingPhasePlay:
		return g.players[g.currentPlayer].GetIsHuman()
	default:
		return false
	}
}

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *King) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *King) GetPhase() string { return g.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *King) SetPhase(phase string) { g.phase = phase }

// GetPlayerCnt はプレイヤー数を返す。
func (g *King) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *King) GetPlayer(i int) *KingPlayer {
	return getPlayer(g.players, i)
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *King) GetCurrentTurn() int { return g.currentPlayer }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *King) SetCurrentTurn(idx int) { g.currentPlayer = idx }

// GetLeadPlayer はリードプレイヤーインデックスを返す。
func (g *King) GetLeadPlayer() int { return g.leadPlayer }

// SetLeadPlayer はリードプレイヤーを設定する (テスト用)。
func (g *King) SetLeadPlayer(idx int) { g.leadPlayer = idx }

// GetDealerIdx は現在の親を返す。
func (g *King) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx は親を設定する (テスト用)。
func (g *King) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetDealNumber は現在のディール番号 (0-indexed) を返す。
func (g *King) GetDealNumber() int { return g.dealNumber }

// SetDealNumber はディール番号を設定する (テスト用)。
func (g *King) SetDealNumber(n int) { g.dealNumber = n }

// GetCurrentContract は現在のコントラクト (-1 = 未選択) を返す。
func (g *King) GetCurrentContract() int { return g.currentContract }

// SetCurrentContract は現在のコントラクトを設定する (テスト用)。
func (g *King) SetCurrentContract(c int) { g.currentContract = c }

// GetTrumpSuit は切り札スート (-1 = なし) を返す。
func (g *King) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit は切り札スートを設定する (テスト用)。
func (g *King) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTrickNumber は現在のトリック番号を返す。
func (g *King) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber はトリック番号を設定する (テスト用)。
func (g *King) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentTrick は進行中のトリックを返す。
func (g *King) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick は進行中のトリックを設定する (テスト用)。
func (g *King) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLastTrick は直前に完了したトリックを返す。
func (g *King) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前トリックの勝者を返す (-1 = なし)。
func (g *King) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetUsedContracts は使用済みのコントラクトを返す。
func (g *King) GetUsedContracts() [KingContractCnt]bool { return g.usedContracts }

// GetLastDealDetail は直前ディールの得点内訳を返す (nil の場合もある)。
func (g *King) GetLastDealDetail() *KingDealDetail { return g.lastDealDetail }

// GetConfig はローカルルール設定を返す。
func (g *King) GetConfig() KingConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *King) SetConfig(config KingConfig) { g.config = config }

// GetPlayableIndices はプレイフェーズでプレイ可能な手札インデックスを返す。
func (g *King) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != KingPhasePlay {
		return nil
	}
	return g.getValidTrickIndices(playerIdx)
}

// GetRoundWinners は (ゲーム終了時) 最高得点プレイヤーのリストを返す。
func (g *King) GetRoundWinners() []int {
	if !g.gameEndFlag {
		return nil
	}
	best := g.players[0].GetTotalScore()
	for _, p := range g.players[1:] {
		if p.GetTotalScore() > best {
			best = p.GetTotalScore()
		}
	}
	winners := make([]int, 0)
	for i, p := range g.players {
		if p.GetTotalScore() == best {
			winners = append(winners, i)
		}
	}
	return winners
}

// --- JSON Serialization ---

// kingJSON is the JSON wire format for King.
type kingJSON struct {
	TrumpCards      *TrumpCards           `json:"tc"`
	Players         []*KingPlayer         `json:"pl"`
	Config          KingConfig            `json:"cf"`
	Phase           string                `json:"ph"`
	DealNumber      int                   `json:"dn"`
	DealerIdx       int                   `json:"di"`
	CurrentContract int                   `json:"cc"`
	TrumpSuit       int                   `json:"ts"`
	UsedContracts   [KingContractCnt]bool `json:"uc"`
	CurrentPlayer   int                   `json:"cp"`
	LeadPlayer      int                   `json:"lp"`
	TrickNumber     int                   `json:"tn"`
	CurrentTrick    []*TrickCard          `json:"ct"`
	LastTrick       []*TrickCard          `json:"lt"`
	LastTrickWinner int                   `json:"lw"`
	GameEndFlag     bool                  `json:"ge"`
	LastDealDetail  *KingDealDetail       `json:"ld"`
	ActionLog       []*ActionLogEntry     `json:"al"`
}

// kingMaxSliceLen caps slice sizes during deserialisation.
const kingMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (g *King) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.phase,
		DealNumber:      g.dealNumber,
		DealerIdx:       g.dealerIdx,
		CurrentContract: g.currentContract,
		TrumpSuit:       g.trumpSuit,
		UsedContracts:   g.usedContracts,
		CurrentPlayer:   g.currentPlayer,
		LeadPlayer:      g.leadPlayer,
		TrickNumber:     g.trickNumber,
		CurrentTrick:    g.currentTrick,
		LastTrick:       g.lastTrick,
		LastTrickWinner: g.lastTrickWinner,
		GameEndFlag:     g.gameEndFlag,
		LastDealDetail:  g.lastDealDetail,
		ActionLog:       g.actionLog,
	})
}

// kingInRange reports whether v is in [0, KingPlayerCnt).
func kingInRange(v int) bool { return v >= 0 && v < KingPlayerCnt }

// kingInRangeOrUnset reports whether v is -1 (unset) or in [0, KingPlayerCnt).
func kingInRangeOrUnset(v int) bool { return v == -1 || kingInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *King) UnmarshalJSON(data []byte) error {
	var j kingJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > kingMaxSliceLen || len(j.CurrentTrick) > kingMaxSliceLen ||
		len(j.LastTrick) > kingMaxSliceLen || len(j.ActionLog) > kingMaxSliceLen {
		return fmt.Errorf("king: input array exceeds maximum allowed size")
	}
	if len(j.Players) != KingPlayerCnt {
		return fmt.Errorf("king: invalid player count %d, expected %d", len(j.Players), KingPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("king: nil player in state")
		}
	}
	// トリックカードは PlayerIdx が範囲内で、Card が非 nil でなければならない。
	if err := kingValidateTrick(j.CurrentTrick); err != nil {
		return err
	}
	if err := kingValidateTrick(j.LastTrick); err != nil {
		return err
	}
	// フェーズ検証。
	switch j.Phase {
	case KingPhaseSelectContract, KingPhasePlay, KingPhaseDealEnd, KingPhaseGameEnd:
	default:
		return fmt.Errorf("king: invalid phase %q", j.Phase)
	}
	// コントラクト enum 検証 (-1 = 未選択 許容)。
	if j.CurrentContract < -1 || j.CurrentContract >= KingContractCnt {
		return fmt.Errorf("king: invalid contract %d", j.CurrentContract)
	}
	// trumpSuit: -1 (なし) 許容、それ以外は [Spade, Diamond]。
	if j.TrumpSuit != -1 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("king: invalid trump suit %d", j.TrumpSuit)
	}
	// dealNumber の範囲。
	if j.DealNumber < 0 || j.DealNumber > KingTotalDeals {
		return fmt.Errorf("king: invalid deal number %d", j.DealNumber)
	}
	// インデックスの範囲 (常に [0, PlayerCnt))。
	if !kingInRange(j.DealerIdx) || !kingInRange(j.CurrentPlayer) || !kingInRange(j.LeadPlayer) {
		return fmt.Errorf("king: index field out of range")
	}
	// lastTrickWinner は -1 (なし) 許容。
	if !kingInRangeOrUnset(j.LastTrickWinner) {
		return fmt.Errorf("king: lastTrickWinner out of range")
	}
	// フェーズが play 以降でコントラクトが確定していなければならない。
	if (j.Phase == KingPhasePlay || j.Phase == KingPhaseDealEnd) &&
		(j.CurrentContract < 0 || j.CurrentContract >= KingContractCnt) {
		return fmt.Errorf("king: contract must be set once play begins")
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newKingDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.dealNumber = j.DealNumber
	g.dealerIdx = j.DealerIdx
	g.currentContract = j.CurrentContract
	g.trumpSuit = j.TrumpSuit
	g.usedContracts = j.UsedContracts
	g.currentPlayer = j.CurrentPlayer
	g.leadPlayer = j.LeadPlayer
	g.trickNumber = j.TrickNumber
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.lastTrick = j.LastTrick
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.lastDealDetail = j.LastDealDetail
	if j.ActionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	} else {
		g.actionLog = j.ActionLog
	}
	return nil
}

// kingValidateTrick は復元したトリック配列の各要素を検証する。
func kingValidateTrick(trick []*TrickCard) error {
	for _, tc := range trick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("king: invalid trick card")
		}
		if !kingInRange(tc.PlayerIdx) {
			return fmt.Errorf("king: trick card player index out of range")
		}
	}
	return nil
}

// kingDealDetailJSON is the JSON wire format for KingDealDetail.
type kingDealDetailJSON struct {
	Contract  int         `json:"co"`
	TrumpSuit int         `json:"ts"`
	DealerIdx int         `json:"di"`
	Gained    map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *KingDealDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingDealDetailJSON{
		Contract:  d.Contract,
		TrumpSuit: d.TrumpSuit,
		DealerIdx: d.DealerIdx,
		Gained:    d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *KingDealDetail) UnmarshalJSON(data []byte) error {
	var j kingDealDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Contract = j.Contract
	d.TrumpSuit = j.TrumpSuit
	d.DealerIdx = j.DealerIdx
	d.Gained = j.Gained
	return nil
}
