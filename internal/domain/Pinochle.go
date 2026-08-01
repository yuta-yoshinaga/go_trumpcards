package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"sort"
)

// PinochlePlayerCnt ピノクルプレイヤー数
const PinochlePlayerCnt = 4

// PinochleHandSize 各プレイヤーの手札枚数
const PinochleHandSize = 12

// PinochleTeamCnt チーム数
const PinochleTeamCnt = 2

// PinochleMinBid 最低ビッド
const PinochleMinBid = 20

// PinochleTotalCardPoints ラウンドあたりの合計カードポイント (最終トリックボーナス含む)
const PinochleTotalCardPoints = 250

// PinochleLastTrickBonus 最終トリックボーナス
const PinochleLastTrickBonus = 10

// PinochlePhase ゲームフェーズ
type PinochlePhase int

// ピノクルのフェーズ定数
const (
	// PinochlePhaseBid ビッドフェーズ
	PinochlePhaseBid PinochlePhase = 0
	// PinochlePhaseTrump トランプ宣言フェーズ
	PinochlePhaseTrump PinochlePhase = 1
	// PinochlePhaseMeld メルドフェーズ
	PinochlePhaseMeld PinochlePhase = 2
	// PinochlePhasePlay トリックプレイフェーズ
	PinochlePhasePlay PinochlePhase = 3
	// PinochlePhaseTrickEnd トリック終了フェーズ
	PinochlePhaseTrickEnd PinochlePhase = 4
	// PinochlePhaseRoundEnd ラウンド終了フェーズ
	PinochlePhaseRoundEnd PinochlePhase = 5
	// PinochlePhaseGameEnd ゲーム終了フェーズ
	PinochlePhaseGameEnd PinochlePhase = 6
)

// PinochleMeldType メルドの種類
type PinochleMeldType int

// メルド種類定数
const (
	PinochleMeldDix                PinochleMeldType = iota // 9 of trump (10)
	PinochleMeldCommonMarriage                             // K-Q of non-trump (20)
	PinochleMeldRoyalMarriage                              // K-Q of trump (40)
	PinochleMeldPinochle                                   // J♦ + Q♠ (40)
	PinochleMeldJacksAround                                // One J of each suit (40)
	PinochleMeldQueensAround                               // One Q of each suit (60)
	PinochleMeldKingsAround                                // One K of each suit (80)
	PinochleMeldAcesAround                                 // One A of each suit (100)
	PinochleMeldRun                                        // A-10-K-Q-J of trump (150)
	PinochleMeldDoublePinochle                             // 2x J♦ + 2x Q♠ (300)
	PinochleMeldDoubleJacksAround                          // 2x J of each suit (400)
	PinochleMeldDoubleQueensAround                         // 2x Q of each suit (600)
	PinochleMeldDoubleKingsAround                          // 2x K of each suit (800)
	PinochleMeldDoubleAcesAround                           // 2x A of each suit (1000)
	PinochleMeldDoubleRun                                  // 2x A-10-K-Q-J of trump (1500)
)

// PinochleMeld 検出されたメルド
type PinochleMeld struct {
	Type   PinochleMeldType `json:"t"`
	Points int              `json:"p"`
	Cards  []*Card          `json:"c"`
}

// pinochleMeldPoints メルド種類ごとのポイント
var pinochleMeldPoints = map[PinochleMeldType]int{
	PinochleMeldDix:                10,
	PinochleMeldCommonMarriage:     20,
	PinochleMeldRoyalMarriage:      40,
	PinochleMeldPinochle:           40,
	PinochleMeldJacksAround:        40,
	PinochleMeldQueensAround:       60,
	PinochleMeldKingsAround:        80,
	PinochleMeldAcesAround:         100,
	PinochleMeldRun:                150,
	PinochleMeldDoublePinochle:     300,
	PinochleMeldDoubleJacksAround:  400,
	PinochleMeldDoubleQueensAround: 600,
	PinochleMeldDoubleKingsAround:  800,
	PinochleMeldDoubleAcesAround:   1000,
	PinochleMeldDoubleRun:          1500,
}

// PinochleHint ヒント情報
type PinochleHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	BidAmount *int   // 推奨ビッド額 (ビッド時)
	Pass      *bool  // パスすべきか
	Suit      *int   // 推奨スート (トランプ宣言時)
	Reason    string // ヒント理由キー
}

// Pinochle ピノクルゲームクラス
type Pinochle struct {
	trumpCards       *TrumpCards
	players          []*PinochlePlayer
	config           PinochleConfig
	phase            PinochlePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	trumpSuit        int // 切り札スート (CardDesignSpade等)
	highestBid       int // 現在の最高ビッド
	highestBidder    int // 最高ビッダーのインデックス
	bidPlayerIdx     int // 現在のビッド手番
	teamScores       [PinochleTeamCnt]int
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerTeam       int // 勝利チーム (-1 = 未確定)
	playerMelds      [PinochlePlayerCnt][]*PinochleMeld
	actionLog        []*ActionLogEntry
}

// NewPinochle コンストラクタ
func NewPinochle(trumpCards *TrumpCards, players []*PinochlePlayer, config PinochleConfig) *Pinochle {
	return &Pinochle{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerTeam:    -1,
		roundNumber:   0,
		dealerIdx:     0,
		highestBidder: -1,
	}
}

// NewDefaultPinochle returns Pinochle with the standard 4-player team setup
// (human team 0, alternating CPU teams) and DefaultPinochleConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultPinochle() *Pinochle {
	players := []*PinochlePlayer{
		NewPinochlePlayer(true, 0),
		NewPinochlePlayer(false, 1),
		NewPinochlePlayer(false, 0),
		NewPinochlePlayer(false, 1),
	}
	return NewPinochle(NewTrumpCardsPinochle(), players, DefaultPinochleConfig())
}

// Reset ゲーム初期化
func (p *Pinochle) Reset() {
	p.gameEndFlag = false
	p.winnerTeam = -1
	p.roundNumber = 1
	p.trickNumber = 0
	p.dealerIdx = 0
	p.teamScores = [PinochleTeamCnt]int{}
	p.actionLog = nil
	p.trumpSuit = 0
	p.highestBid = 0
	p.highestBidder = -1
	p.playerMelds = [PinochlePlayerCnt][]*PinochleMeld{}

	for _, pl := range p.players {
		pl.ResetRound()
	}

	p.dealRound()
	p.phase = PinochlePhaseBid
	p.bidPlayerIdx = (p.dealerIdx + 1) % PinochlePlayerCnt
}

// NextRound 次のラウンドを開始する
func (p *Pinochle) NextRound() {
	if p.phase != PinochlePhaseRoundEnd {
		return
	}
	p.roundNumber++
	p.trickNumber = 0
	p.dealerIdx = (p.dealerIdx + 1) % PinochlePlayerCnt
	p.trumpSuit = 0
	p.highestBid = 0
	p.highestBidder = -1
	p.playerMelds = [PinochlePlayerCnt][]*PinochleMeld{}

	for _, pl := range p.players {
		pl.ResetRound()
	}

	p.dealRound()
	p.phase = PinochlePhaseBid
	p.bidPlayerIdx = (p.dealerIdx + 1) % PinochlePlayerCnt
}

// dealRound ラウンドのカードを配る
func (p *Pinochle) dealRound() {
	p.trumpCards.Shuffle()
	for range PinochleHandSize {
		for j := range PinochlePlayerCnt {
			card := p.trumpCards.DrawCard()
			if card != nil {
				p.players[j].AddCard(card)
			}
		}
	}
	p.currentTrick = nil
}

// ─── Getters ────────────────────────────────────────────

// GetPhase フェーズを取得
func (p *Pinochle) GetPhase() PinochlePhase { return p.phase }

// GetRoundNumber ラウンド番号を取得
func (p *Pinochle) GetRoundNumber() int { return p.roundNumber }

// GetTrickNumber トリック番号を取得
func (p *Pinochle) GetTrickNumber() int { return p.trickNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得
func (p *Pinochle) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// GetCurrentTrick 現在のトリックを取得
func (p *Pinochle) GetCurrentTrick() []*TrickCard { return p.currentTrick }

// GetDealerIdx ディーラーインデックスを取得
func (p *Pinochle) GetDealerIdx() int { return p.dealerIdx }

// GetTrumpSuit 切り札スートを取得
func (p *Pinochle) GetTrumpSuit() int { return p.trumpSuit }

// GetHighestBid 最高ビッド額を取得
func (p *Pinochle) GetHighestBid() int { return p.highestBid }

// GetHighestBidder 最高ビッダーインデックスを取得
func (p *Pinochle) GetHighestBidder() int { return p.highestBidder }

// GetBidPlayerIdx ビッド手番インデックスを取得
func (p *Pinochle) GetBidPlayerIdx() int { return p.bidPlayerIdx }

// GetTeamScores チームスコアを取得
func (p *Pinochle) GetTeamScores() [PinochleTeamCnt]int { return p.teamScores }

// GetLeadPlayerIdx リードプレイヤーインデックスを取得
func (p *Pinochle) GetLeadPlayerIdx() int { return p.leadPlayerIdx }

// GetGameEndFlag ゲーム終了フラグを取得
func (p *Pinochle) GetGameEndFlag() bool { return p.gameEndFlag }

// IsHumanTurn 現在の手番が人間かを返す
func (p *Pinochle) IsHumanTurn() bool {
	return p.players[p.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かを返す
func (p *Pinochle) IsHumanBidTurn() bool {
	return p.players[p.bidPlayerIdx].GetIsHuman()
}

// GetPlayerCnt プレイヤー数を取得する
func (p *Pinochle) GetPlayerCnt() int { return PinochlePlayerCnt }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (p *Pinochle) GetPlayer(i int) *PinochlePlayer { return p.players[i] }

// GetTeamScore チームスコアを取得する
func (p *Pinochle) GetTeamScore(team int) int { return p.teamScores[team] }

// GetHint ヒントを取得する
func (p *Pinochle) GetHint() *PinochleHint { return p.Hint() }

// GetWinnerTeam 勝利チームを取得
func (p *Pinochle) GetWinnerTeam() int { return p.winnerTeam }

// GetPlayers プレイヤーを取得
func (p *Pinochle) GetPlayers() []*PinochlePlayer { return p.players }

// GetConfig 設定を取得
func (p *Pinochle) GetConfig() PinochleConfig { return p.config }

// SetConfig 設定を変更
func (p *Pinochle) SetConfig(config PinochleConfig) { p.config = config }

// GetPlayerMelds プレイヤーのメルドを取得
func (p *Pinochle) GetPlayerMelds() [PinochlePlayerCnt][]*PinochleMeld { return p.playerMelds }

// GetActionLog アクションログを取得
func (p *Pinochle) GetActionLog() []*ActionLogEntry { return p.actionLog }

// addLog アクションログを追加
func (p *Pinochle) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: p.trickNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ─── Card Ranking & Point Values ────────────────────────

// pinochleCardRankOrder ピノクルでのカードランク順序 (A > 10 > K > Q > J > 9)
// value → 比較用ランク値
func pinochleRankValue(value int) int {
	switch value {
	case 1: // Ace
		return 6
	case 10:
		return 5
	case 13: // King
		return 4
	case 12: // Queen
		return 3
	case 11: // Jack
		return 2
	case 9:
		return 1
	default:
		return 0
	}
}

// cardRank トリック比較用のカードランクを返す
// trump > lead suit > other suits, 同スートならランク順
func (p *Pinochle) cardRank(card *Card) int {
	base := pinochleRankValue(card.GetValue())
	suit := card.GetDesign()

	if suit == p.trumpSuit {
		return 400 + base
	}
	if len(p.currentTrick) > 0 && suit == p.currentTrick[0].Card.GetDesign() {
		return 200 + base
	}
	return 100 + base
}

// pinochleCardPointValue カードのポイント値を返す
func pinochleCardPointValue(card *Card) int {
	switch card.GetValue() {
	case 1: // Ace
		return 11
	case 10:
		return 10
	case 13: // King
		return 4
	case 12: // Queen
		return 3
	case 11: // Jack
		return 2
	default:
		return 0
	}
}

// ─── Meld Evaluation ────────────────────────────────────

// evaluateMelds 手札からメルドを検出する
func evaluateMelds(hand []*Card, trumpSuit int) []*PinochleMeld {
	// カードを {suit, value} ペアでカウント (最大2)
	type sv struct{ suit, value int }
	counts := make(map[sv]int)
	cardMap := make(map[sv][]*Card)
	for _, c := range hand {
		key := sv{c.GetDesign(), c.GetValue()}
		counts[key]++
		cardMap[key] = append(cardMap[key], c)
	}

	var melds []*PinochleMeld
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}

	// ── Double Run (1500) / Run (150) ──
	runValues := []int{1, 10, 13, 12, 11} // A,10,K,Q,J
	minRunCount := 3                      // each {suit,value} pair has at most 2 copies, so 3 is unreachable
	for _, v := range runValues {
		c := counts[sv{trumpSuit, v}]
		if c < minRunCount {
			minRunCount = c
		}
	}
	if minRunCount >= 2 {
		var cards []*Card
		for _, v := range runValues {
			cards = append(cards, cardMap[sv{trumpSuit, v}]...)
		}
		melds = append(melds, &PinochleMeld{
			Type:   PinochleMeldDoubleRun,
			Points: pinochleMeldPoints[PinochleMeldDoubleRun],
			Cards:  cards,
		})
	} else if minRunCount >= 1 {
		var cards []*Card
		for _, v := range runValues {
			cards = append(cards, cardMap[sv{trumpSuit, v}][0])
		}
		melds = append(melds, &PinochleMeld{
			Type:   PinochleMeldRun,
			Points: pinochleMeldPoints[PinochleMeldRun],
			Cards:  cards,
		})
	}

	// ── Double Around / Around (Aces, Kings, Queens, Jacks) ──
	type aroundDef struct {
		value      int
		singleType PinochleMeldType
		doubleType PinochleMeldType
	}
	arounds := []aroundDef{
		{1, PinochleMeldAcesAround, PinochleMeldDoubleAcesAround},
		{13, PinochleMeldKingsAround, PinochleMeldDoubleKingsAround},
		{12, PinochleMeldQueensAround, PinochleMeldDoubleQueensAround},
		{11, PinochleMeldJacksAround, PinochleMeldDoubleJacksAround},
	}
	for _, a := range arounds {
		minCount := 3
		for _, s := range suits {
			c := counts[sv{s, a.value}]
			if c < minCount {
				minCount = c
			}
		}
		if minCount >= 2 {
			var cards []*Card
			for _, s := range suits {
				cards = append(cards, cardMap[sv{s, a.value}]...)
			}
			melds = append(melds, &PinochleMeld{
				Type:   a.doubleType,
				Points: pinochleMeldPoints[a.doubleType],
				Cards:  cards,
			})
		} else if minCount >= 1 {
			var cards []*Card
			for _, s := range suits {
				cards = append(cards, cardMap[sv{s, a.value}][0])
			}
			melds = append(melds, &PinochleMeld{
				Type:   a.singleType,
				Points: pinochleMeldPoints[a.singleType],
				Cards:  cards,
			})
		}
	}

	// ── Double Pinochle (300) / Pinochle (40): J♦ + Q♠ ──
	jdCount := counts[sv{CardDesignDiamond, 11}]
	qsCount := counts[sv{CardDesignSpade, 12}]
	pinochleCount := min(jdCount, qsCount)
	if pinochleCount >= 2 {
		var cards []*Card
		cards = append(cards, cardMap[sv{CardDesignDiamond, 11}]...)
		cards = append(cards, cardMap[sv{CardDesignSpade, 12}]...)
		melds = append(melds, &PinochleMeld{
			Type:   PinochleMeldDoublePinochle,
			Points: pinochleMeldPoints[PinochleMeldDoublePinochle],
			Cards:  cards,
		})
	} else if pinochleCount >= 1 {
		melds = append(melds, &PinochleMeld{
			Type:   PinochleMeldPinochle,
			Points: pinochleMeldPoints[PinochleMeldPinochle],
			Cards:  []*Card{cardMap[sv{CardDesignDiamond, 11}][0], cardMap[sv{CardDesignSpade, 12}][0]},
		})
	}

	// ── Royal Marriage (40) / Common Marriage (20): K-Q of same suit ──
	// RunにはK-Qが含まれるが、メルドは重複計上可能なので独立して評価する
	for _, s := range suits {
		kc := counts[sv{s, 13}]
		qc := counts[sv{s, 12}]
		marriageCount := min(kc, qc)
		// Runが検出された場合、trump suitのmarriageはRunに含まれるがルール上は
		// 独立して数えられるのでそのまま加算する (ただしRunのポイントにマリッジ分は含まれている
		// 伝統ルールではRunにマリッジを含むためmarriageは加算しない)
		// 伝統ルール: Runはマリッジを含む → trumpスートでRunがある場合はmarriage非加算
		hasRun := minRunCount >= 1 && s == trumpSuit
		if hasRun {
			// RunにはRoyal Marriageが含まれるため非加算
			continue
		}
		for i := range marriageCount {
			meldType := PinochleMeldCommonMarriage
			if s == trumpSuit {
				meldType = PinochleMeldRoyalMarriage
			}
			var cards []*Card
			if i < len(cardMap[sv{s, 13}]) {
				cards = append(cards, cardMap[sv{s, 13}][i])
			}
			if i < len(cardMap[sv{s, 12}]) {
				cards = append(cards, cardMap[sv{s, 12}][i])
			}
			melds = append(melds, &PinochleMeld{
				Type:   meldType,
				Points: pinochleMeldPoints[meldType],
				Cards:  cards,
			})
		}
	}

	// ── Dix (10): 9 of trump ──
	dixCount := counts[sv{trumpSuit, 9}]
	for i := range dixCount {
		melds = append(melds, &PinochleMeld{
			Type:   PinochleMeldDix,
			Points: pinochleMeldPoints[PinochleMeldDix],
			Cards:  []*Card{cardMap[sv{trumpSuit, 9}][i]},
		})
	}

	return melds
}

// meldTotalPoints メルドの合計ポイントを計算
func meldTotalPoints(melds []*PinochleMeld) int {
	total := 0
	for _, m := range melds {
		total += m.Points
	}
	return total
}

// ─── Bidding ─────────────────────────────────────────────

// PlayerBid 人間プレイヤーがビッドする
func (p *Pinochle) PlayerBid(amount int) error {
	if p.phase != PinochlePhaseBid {
		return NewDomainError(ErrWrongPhase, "ビッドフェーズではありません")
	}
	if !p.players[p.bidPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doBid(p.bidPlayerIdx, amount)
}

// PlayerPass 人間プレイヤーがパスする
func (p *Pinochle) PlayerPass() error {
	if p.phase != PinochlePhaseBid {
		return NewDomainError(ErrWrongPhase, "ビッドフェーズではありません")
	}
	if !p.players[p.bidPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doPass(p.bidPlayerIdx)
}

// doBid ビッドを実行
func (p *Pinochle) doBid(playerIdx, amount int) error {
	if amount < PinochleMinBid {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("ビッドは%d以上でなければなりません", PinochleMinBid))
	}
	if p.highestBid > 0 && amount <= p.highestBid {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("現在のビッド%dより大きくなければなりません", p.highestBid))
	}

	p.players[playerIdx].SetBid(amount)
	p.highestBid = amount
	p.highestBidder = playerIdx
	p.addLog(playerIdx, "bid", fmt.Sprintf("ビッド: %d", amount), nil)
	p.advanceBidder()
	return nil
}

// doPass パスを実行
func (p *Pinochle) doPass(playerIdx int) error {
	// 最後の一人はパスできない (ビッドを強制)
	activeBidders := 0
	for i := range PinochlePlayerCnt {
		if !p.players[i].GetHasPassed() {
			activeBidders++
		}
	}
	if activeBidders <= 1 && p.highestBid > 0 {
		return NewDomainError(ErrCannotPass, "最後のビッダーはパスできません")
	}

	p.players[playerIdx].SetHasPassed(true)
	p.addLog(playerIdx, "pass", "パス", nil)
	p.advanceBidder()
	return nil
}

// advanceBidder 次のビッダーに進む
func (p *Pinochle) advanceBidder() {
	// アクティブなビッダーが1人だけになったらビッド終了
	activeBidders := 0
	lastActive := -1
	for i := range PinochlePlayerCnt {
		if !p.players[i].GetHasPassed() {
			activeBidders++
			lastActive = i
		}
	}

	if activeBidders == 1 && p.highestBid > 0 {
		// ビッド終了 → トランプ宣言フェーズへ
		p.highestBidder = lastActive
		p.phase = PinochlePhaseTrump
		p.currentPlayerIdx = p.highestBidder
		return
	}

	// 全員パスした場合 (ビッドなし) → ディーラーが最低ビッドを強制
	if activeBidders == 0 || (activeBidders == 1 && p.highestBid == 0) {
		if p.highestBid == 0 {
			p.highestBid = PinochleMinBid
			p.highestBidder = p.dealerIdx
			p.players[p.dealerIdx].SetBid(PinochleMinBid)
			p.addLog(p.dealerIdx, "forced_bid", fmt.Sprintf("強制ビッド: %d", PinochleMinBid), nil)
		}
		p.phase = PinochlePhaseTrump
		p.currentPlayerIdx = p.highestBidder
		return
	}

	// 次のビッダーに進む
	next := (p.bidPlayerIdx + 1) % PinochlePlayerCnt
	for p.players[next].GetHasPassed() {
		next = (next + 1) % PinochlePlayerCnt
	}
	p.bidPlayerIdx = next
}

// CpuBid CPUがビッドまたはパスする
func (p *Pinochle) CpuBid() {
	if p.phase != PinochlePhaseBid {
		return
	}
	playerIdx := p.bidPlayerIdx
	if p.players[playerIdx].GetIsHuman() {
		return
	}

	switch p.config.CpuDifficulty {
	case PinochleCpuDifficultyEasy:
		p.cpuBidEasy(playerIdx)
	case PinochleCpuDifficultyNormal:
		p.cpuBidNormal(playerIdx)
	case PinochleCpuDifficultyHard:
		p.cpuBidHard(playerIdx)
	}
}

// cpuBidEasy Easy: 30%の確率でビッド
func (p *Pinochle) cpuBidEasy(playerIdx int) {
	if rand.Intn(100) < 30 {
		bid := PinochleMinBid
		if p.highestBid > 0 {
			bid = p.highestBid + 1
		}
		_ = p.doBid(playerIdx, bid)
	} else {
		_ = p.doPass(playerIdx)
	}
}

// cpuBidNormal Normal: メルドポイント + 推定トリックポイントでビッド判断
func (p *Pinochle) cpuBidNormal(playerIdx int) {
	bestSuit, bestMeldPoints := p.cpuEvalBestTrump(playerIdx)
	trickEstimate := p.cpuEstimateTrickPoints(playerIdx, bestSuit)
	totalEstimate := bestMeldPoints + trickEstimate

	bid := totalEstimate
	if bid < PinochleMinBid {
		_ = p.doPass(playerIdx)
		return
	}
	if p.highestBid > 0 && bid <= p.highestBid {
		_ = p.doPass(playerIdx)
		return
	}
	if p.highestBid > 0 {
		bid = p.highestBid + 1
	}
	if bid > totalEstimate {
		_ = p.doPass(playerIdx)
		return
	}
	_ = p.doBid(playerIdx, bid)
}

// cpuBidHard Hard: より精密なビッド評価
func (p *Pinochle) cpuBidHard(playerIdx int) {
	bestSuit, bestMeldPoints := p.cpuEvalBestTrump(playerIdx)
	trickEstimate := p.cpuEstimateTrickPoints(playerIdx, bestSuit)
	totalEstimate := bestMeldPoints + trickEstimate

	// ハードAIは少し積極的にビッドする
	bid := totalEstimate
	if bid < PinochleMinBid {
		_ = p.doPass(playerIdx)
		return
	}
	if p.highestBid > 0 && bid <= p.highestBid {
		// 推定値が現在のビッドの110%以内ならまだビッド
		if totalEstimate > int(float64(p.highestBid)*1.1) {
			bid = p.highestBid + 1
		} else {
			_ = p.doPass(playerIdx)
			return
		}
	} else if p.highestBid > 0 {
		bid = p.highestBid + 1
	}
	_ = p.doBid(playerIdx, bid)
}

// cpuEvalBestTrump CPUが最適なトランプスートを評価する
func (p *Pinochle) cpuEvalBestTrump(playerIdx int) (bestSuit, bestMeldPoints int) {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	bestSuit = suits[0]
	bestMeldPoints = 0

	hand := make([]*Card, 0, p.players[playerIdx].GetCardsSize())
	for i := 0; i < p.players[playerIdx].GetCardsSize(); i++ {
		hand = append(hand, p.players[playerIdx].GetCard(i))
	}

	for _, s := range suits {
		melds := evaluateMelds(hand, s)
		total := meldTotalPoints(melds)
		if total > bestMeldPoints {
			bestMeldPoints = total
			bestSuit = s
		}
	}
	return
}

// cpuEstimateTrickPoints CPUがトリックポイントを推定する
func (p *Pinochle) cpuEstimateTrickPoints(playerIdx, trumpSuit int) int {
	points := 0
	for i := 0; i < p.players[playerIdx].GetCardsSize(); i++ {
		card := p.players[playerIdx].GetCard(i)
		if card.GetDesign() == trumpSuit {
			// トランプカードはポイントを獲得する可能性が高い
			points += pinochleCardPointValue(card)
		} else if card.GetValue() == 1 { // 非トランプのAce
			points += 5 // 半分くらいの確率で取れる見積もり
		}
	}
	return points
}

// ─── Trump Declaration ──────────────────────────────────

// PlayerCallTrump 人間プレイヤーがトランプスートを宣言する
func (p *Pinochle) PlayerCallTrump(suit int) error {
	if p.phase != PinochlePhaseTrump {
		return NewDomainError(ErrWrongPhase, "トランプ宣言フェーズではありません")
	}
	if !p.players[p.currentPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doCallTrump(p.currentPlayerIdx, suit)
}

// doCallTrump トランプ宣言を実行
func (p *Pinochle) doCallTrump(playerIdx, suit int) error {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}
	p.trumpSuit = suit
	suitNames := map[int]string{
		CardDesignSpade:   "スペード",
		CardDesignClover:  "クラブ",
		CardDesignHeart:   "ハート",
		CardDesignDiamond: "ダイヤ",
	}
	p.addLog(playerIdx, "trump", fmt.Sprintf("トランプ宣言: %s", suitNames[suit]), nil)

	// メルドフェーズへ移行
	p.evaluateAllMelds()
	p.phase = PinochlePhaseMeld
	return nil
}

// CpuCallTrump CPUがトランプを宣言する
func (p *Pinochle) CpuCallTrump() {
	if p.phase != PinochlePhaseTrump {
		return
	}
	playerIdx := p.currentPlayerIdx
	if p.players[playerIdx].GetIsHuman() {
		return
	}

	bestSuit, _ := p.cpuEvalBestTrump(playerIdx)
	_ = p.doCallTrump(playerIdx, bestSuit)
}

// evaluateAllMelds 全プレイヤーのメルドを評価
func (p *Pinochle) evaluateAllMelds() {
	for i := range PinochlePlayerCnt {
		hand := make([]*Card, 0, p.players[i].GetCardsSize())
		for j := 0; j < p.players[i].GetCardsSize(); j++ {
			hand = append(hand, p.players[i].GetCard(j))
		}
		melds := evaluateMelds(hand, p.trumpSuit)
		p.playerMelds[i] = melds
		p.players[i].SetMeldScore(meldTotalPoints(melds))
		p.addLog(i, "meld", fmt.Sprintf("メルド: %d点", p.players[i].GetMeldScore()), nil)
	}
}

// ─── Meld Phase ─────────────────────────────────────────

// ConfirmMelds メルドを確認してプレイフェーズに進む
func (p *Pinochle) ConfirmMelds() {
	if p.phase != PinochlePhaseMeld {
		return
	}
	p.trickNumber = 1
	p.leadPlayerIdx = p.highestBidder
	p.currentPlayerIdx = p.leadPlayerIdx
	p.currentTrick = nil
	p.phase = PinochlePhasePlay
}

// ─── Trick-Taking ───────────────────────────────────────

// getValidPlayIndices プレイ可能なカードのインデックスを返す
// ピノクルのルール:
// 1. リードスートに従う (must follow suit)
// 2. リードスートがなければトランプを出す (must trump)
// 3. 可能なら現在の勝者を上回るカードを出す (must win)
func (p *Pinochle) getValidPlayIndices(playerIdx int) []int {
	player := p.players[playerIdx]
	handSize := player.GetCardsSize()
	if handSize == 0 {
		return nil
	}

	// リードがない場合は全てのカードが有効
	if len(p.currentTrick) == 0 {
		indices := make([]int, handSize)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	leadSuit := p.currentTrick[0].Card.GetDesign()
	currentWinnerRank := 0
	for _, tc := range p.currentTrick {
		r := p.cardRank(tc.Card)
		if r > currentWinnerRank {
			currentWinnerRank = r
		}
	}

	// 1. リードスートのカードを収集
	var leadSuitIndices []int
	var leadSuitHigherIndices []int
	for i := range handSize {
		card := player.GetCard(i)
		if card.GetDesign() == leadSuit {
			leadSuitIndices = append(leadSuitIndices, i)
			if p.cardRank(card) > currentWinnerRank {
				leadSuitHigherIndices = append(leadSuitHigherIndices, i)
			}
		}
	}

	if len(leadSuitIndices) > 0 {
		// リードスートがある場合
		if len(leadSuitHigherIndices) > 0 {
			// 勝てるカードがある場合はそれのみ
			return leadSuitHigherIndices
		}
		// 勝てなくてもリードスートを出す
		return leadSuitIndices
	}

	// 2. リードスートがない場合 → トランプを出す
	if leadSuit != p.trumpSuit {
		var trumpIndices []int
		var trumpHigherIndices []int
		for i := range handSize {
			card := player.GetCard(i)
			if card.GetDesign() == p.trumpSuit {
				trumpIndices = append(trumpIndices, i)
				if p.cardRank(card) > currentWinnerRank {
					trumpHigherIndices = append(trumpHigherIndices, i)
				}
			}
		}
		if len(trumpIndices) > 0 {
			if len(trumpHigherIndices) > 0 {
				return trumpHigherIndices
			}
			return trumpIndices
		}
	}

	// 3. リードスートもトランプもない場合は何でも出せる
	indices := make([]int, handSize)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (p *Pinochle) PlayerPlay(cardIndex int) error {
	if p.phase != PinochlePhasePlay {
		return NewDomainError(ErrWrongPhase, "プレイフェーズではありません")
	}
	if !p.players[p.currentPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doPlay(p.currentPlayerIdx, cardIndex)
}

// doPlay カードプレイを実行
func (p *Pinochle) doPlay(playerIdx, cardIndex int) error {
	player := p.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "無効なカードインデックスです")
	}

	// バリデーション
	validIndices := p.getValidPlayIndices(playerIdx)
	if !slices.Contains(validIndices, cardIndex) {
		return NewDomainError(ErrInvalidPlay, "このカードはプレイできません")
	}

	card := player.RemoveCard(cardIndex)
	p.currentTrick = append(p.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	p.addLog(playerIdx, "play", "カードプレイ", []*Card{card})

	// トリック完了チェック
	if len(p.currentTrick) == PinochlePlayerCnt {
		p.phase = PinochlePhaseTrickEnd
	} else {
		p.currentPlayerIdx = (p.currentPlayerIdx + 1) % PinochlePlayerCnt
	}
	return nil
}

// CpuPlay CPUがカードをプレイする
func (p *Pinochle) CpuPlay() {
	if p.phase != PinochlePhasePlay {
		return
	}
	playerIdx := p.currentPlayerIdx
	if p.players[playerIdx].GetIsHuman() {
		return
	}

	validIndices := p.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return
	}

	var cardIndex int
	switch p.config.CpuDifficulty {
	case PinochleCpuDifficultyEasy:
		cardIndex = validIndices[rand.Intn(len(validIndices))]
	case PinochleCpuDifficultyNormal:
		cardIndex = p.cpuPlayNormal(playerIdx, validIndices)
	case PinochleCpuDifficultyHard:
		cardIndex = p.cpuPlayHard(playerIdx, validIndices)
	default:
		cardIndex = validIndices[rand.Intn(len(validIndices))]
	}

	_ = p.doPlay(playerIdx, cardIndex)
}

// cpuPlayNormal Normal: 基本戦略
func (p *Pinochle) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := p.players[playerIdx]
	if len(p.currentTrick) == 0 {
		// リード: 最も高いカードを出す
		bestIdx := validIndices[0]
		bestRank := p.cardRank(player.GetCard(bestIdx))
		for _, vi := range validIndices[1:] {
			r := p.cardRank(player.GetCard(vi))
			if r > bestRank {
				bestRank = r
				bestIdx = vi
			}
		}
		return bestIdx
	}

	// フォロー: 勝てる最小のカードを出す、勝てなければ最小のカードを出す
	currentWinnerRank := 0
	for _, tc := range p.currentTrick {
		r := p.cardRank(tc.Card)
		if r > currentWinnerRank {
			currentWinnerRank = r
		}
	}

	var winningIndices []int
	for _, vi := range validIndices {
		if p.cardRank(player.GetCard(vi)) > currentWinnerRank {
			winningIndices = append(winningIndices, vi)
		}
	}

	if len(winningIndices) > 0 {
		// 勝てる最小のカードを出す
		bestIdx := winningIndices[0]
		bestRank := p.cardRank(player.GetCard(bestIdx))
		for _, vi := range winningIndices[1:] {
			r := p.cardRank(player.GetCard(vi))
			if r < bestRank {
				bestRank = r
				bestIdx = vi
			}
		}
		return bestIdx
	}

	// 勝てない場合は最小のカードを出す
	worstIdx := validIndices[0]
	worstRank := p.cardRank(player.GetCard(worstIdx))
	for _, vi := range validIndices[1:] {
		r := p.cardRank(player.GetCard(vi))
		if r < worstRank {
			worstRank = r
			worstIdx = vi
		}
	}
	return worstIdx
}

// cpuPlayHard Hard: 高度な戦略
func (p *Pinochle) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := p.players[playerIdx]
	if len(p.currentTrick) == 0 {
		// リード: ポイントカードのAceやトランプをリードする
		bestIdx := validIndices[0]
		bestScore := -1
		for _, vi := range validIndices {
			card := player.GetCard(vi)
			score := pinochleRankValue(card.GetValue())
			if card.GetDesign() == p.trumpSuit {
				score += 10
			}
			if score > bestScore {
				bestScore = score
				bestIdx = vi
			}
		}
		return bestIdx
	}

	// 最後のプレイヤー: パートナーが勝っていればポイントカードを、そうでなければ最小カードを
	currentWinnerIdx := p.trickWinnerIdx()
	currentWinnerTeam := p.players[p.currentTrick[currentWinnerIdx].PlayerIdx].GetTeam()
	myTeam := p.players[playerIdx].GetTeam()

	if len(p.currentTrick) == PinochlePlayerCnt-1 && currentWinnerTeam == myTeam {
		// パートナーが勝っている → ポイントの高いカードを出す
		bestIdx := validIndices[0]
		bestPoints := pinochleCardPointValue(player.GetCard(bestIdx))
		for _, vi := range validIndices[1:] {
			pts := pinochleCardPointValue(player.GetCard(vi))
			if pts > bestPoints {
				bestPoints = pts
				bestIdx = vi
			}
		}
		return bestIdx
	}

	// Normal戦略にフォールバック
	return p.cpuPlayNormal(playerIdx, validIndices)
}

// trickWinnerIdx 現在のトリックの勝者のインデックス (currentTrick内)を返す
func (p *Pinochle) trickWinnerIdx() int {
	winnerIdx := 0
	winnerRank := p.cardRank(p.currentTrick[0].Card)
	for i, tc := range p.currentTrick[1:] {
		r := p.cardRank(tc.Card)
		if r > winnerRank {
			winnerRank = r
			winnerIdx = i + 1
		}
	}
	return winnerIdx
}

// trickWinner 現在のトリックの勝者のプレイヤーインデックスを返す
func (p *Pinochle) trickWinner() int {
	return p.currentTrick[p.trickWinnerIdx()].PlayerIdx
}

// ResolveTrick トリックを解決する
func (p *Pinochle) ResolveTrick() {
	if p.phase != PinochlePhaseTrickEnd || len(p.currentTrick) != PinochlePlayerCnt {
		return
	}

	winner := p.trickWinner()

	// トリックのカードポイントを計算
	trickPoints := 0
	trickCards := make([]*Card, 0, PinochlePlayerCnt)
	for _, tc := range p.currentTrick {
		trickPoints += pinochleCardPointValue(tc.Card)
		trickCards = append(trickCards, tc.Card)
	}

	// 最終トリックボーナス
	if p.trickNumber == PinochleHandSize {
		trickPoints += PinochleLastTrickBonus
	}

	p.players[winner].AddTrick(trickCards)
	p.players[winner].AddTrickPoints(trickPoints)

	p.addLog(winner, "trick_win", fmt.Sprintf("トリック獲得: %d点", trickPoints), trickCards)

	// 次のトリックまたはラウンド終了
	if p.trickNumber >= PinochleHandSize {
		p.phase = PinochlePhaseRoundEnd
		p.scoreRound()
	} else {
		p.leadPlayerIdx = winner
		p.currentPlayerIdx = winner
	}
}

// NextTrick 次のトリックを開始する
// 注: ResolveTrick() は呼び出し元 (Interactor) が先に実行する。
func (p *Pinochle) NextTrick() {
	if p.phase != PinochlePhaseTrickEnd {
		return
	}
	p.currentTrick = nil
	p.currentPlayerIdx = p.leadPlayerIdx
	p.trickNumber++
	p.phase = PinochlePhasePlay
}

// ─── Scoring ────────────────────────────────────────────

// scoreRound ラウンドのスコアを計算する
func (p *Pinochle) scoreRound() {
	// チームごとのトリックポイントとメルドポイントを集計
	var teamTrickPoints [PinochleTeamCnt]int
	var teamMeldPoints [PinochleTeamCnt]int
	var teamTrickCount [PinochleTeamCnt]int

	for i := range PinochlePlayerCnt {
		team := p.players[i].GetTeam()
		teamTrickPoints[team] += p.players[i].GetTrickPoints()
		teamMeldPoints[team] += p.players[i].GetMeldScore()
		teamTrickCount[team] += p.players[i].GetTrickCount()
	}

	if p.highestBidder < 0 || p.highestBidder >= PinochlePlayerCnt {
		return
	}
	bidderTeam := p.players[p.highestBidder].GetTeam()
	defenderTeam := 1 - bidderTeam

	// メルド没収: トリックを1つも取れなかったチームのメルドは没収
	for t := range PinochleTeamCnt {
		if teamTrickCount[t] == 0 {
			teamMeldPoints[t] = 0
		}
	}

	// ビッドチームの得点計算
	bidderTotal := teamTrickPoints[bidderTeam] + teamMeldPoints[bidderTeam]
	if bidderTotal >= p.highestBid {
		p.teamScores[bidderTeam] += bidderTotal
	} else {
		// ビッド失敗: ビッド額を失う
		p.teamScores[bidderTeam] -= p.highestBid
	}

	// ディフェンダーチームは常に得点を加算
	defenderTotal := teamTrickPoints[defenderTeam] + teamMeldPoints[defenderTeam]
	p.teamScores[defenderTeam] += defenderTotal

	p.addLog(-1, "round_score",
		fmt.Sprintf("ラウンドスコア: チーム0=%d, チーム1=%d (累計: チーム0=%d, チーム1=%d)",
			func() int {
				if bidderTeam == 0 {
					if bidderTotal >= p.highestBid {
						return bidderTotal
					}
					return -p.highestBid
				}
				return defenderTotal
			}(),
			func() int {
				if bidderTeam == 1 {
					if bidderTotal >= p.highestBid {
						return bidderTotal
					}
					return -p.highestBid
				}
				return defenderTotal
			}(),
			p.teamScores[0], p.teamScores[1]),
		nil)

	// ゲーム終了チェック
	for t := range PinochleTeamCnt {
		if p.teamScores[t] >= p.config.PointLimit {
			p.gameEndFlag = true
			p.winnerTeam = t
			// 両チームが到達した場合はビッドチームが優先
			if p.teamScores[1-t] >= p.config.PointLimit {
				p.winnerTeam = bidderTeam
			}
			p.phase = PinochlePhaseGameEnd
			return
		}
	}
}

// ─── Hint ───────────────────────────────────────────────

// Hint ヒントを返す
func (p *Pinochle) Hint() *PinochleHint {
	// **人間の手番でなければ答えない。**hintBid / hintTrump / hintPlay は
	// bidPlayerIdx / currentPlayerIdx をそのまま使うので、席が誰かを見ない。
	// Output() が毎レスポンスでこれを呼ぶようになった以上、ガードが無いと
	// **CPU の手番に CPU 自身の手を「推奨手」として人間に見せる** (#4585 のレビュー指摘)。
	switch p.phase {
	case PinochlePhaseBid:
		if !p.IsHumanBidTurn() {
			return nil
		}
		return p.hintBid()
	case PinochlePhaseTrump, PinochlePhasePlay:
		if !p.IsHumanTurn() {
			return nil
		}
		if p.phase == PinochlePhaseTrump {
			return p.hintTrump()
		}
		return p.hintPlay()
	default:
		return nil
	}
}

func (p *Pinochle) hintBid() *PinochleHint {
	playerIdx := p.bidPlayerIdx
	bestSuit, bestMeldPoints := p.cpuEvalBestTrump(playerIdx)
	trickEstimate := p.cpuEstimateTrickPoints(playerIdx, bestSuit)
	totalEstimate := bestMeldPoints + trickEstimate

	if totalEstimate < PinochleMinBid || (p.highestBid > 0 && totalEstimate <= p.highestBid) {
		pass := true
		return &PinochleHint{Pass: &pass, Reason: "hint_pass"}
	}
	bid := totalEstimate
	if p.highestBid > 0 && bid <= p.highestBid {
		bid = p.highestBid + 1
	}
	return &PinochleHint{BidAmount: &bid, Reason: "hint_bid"}
}

func (p *Pinochle) hintTrump() *PinochleHint {
	bestSuit, _ := p.cpuEvalBestTrump(p.currentPlayerIdx)
	return &PinochleHint{Suit: &bestSuit, Reason: "hint_trump"}
}

func (p *Pinochle) hintPlay() *PinochleHint {
	playerIdx := p.currentPlayerIdx
	validIndices := p.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return nil
	}
	cardIndex := p.cpuPlayHard(playerIdx, validIndices)
	return &PinochleHint{CardIndex: &cardIndex, Reason: "hint_play"}
}

// ─── JSON Serialization ─────────────────────────────────

// pinochleJSON is the JSON wire format for Pinochle.
type pinochleJSON struct {
	TrumpCards       *TrumpCards                        `json:"tc"`
	Players          []*PinochlePlayer                  `json:"ps"`
	Config           PinochleConfig                     `json:"cf"`
	Phase            PinochlePhase                      `json:"ph"`
	RoundNumber      int                                `json:"rn"`
	TrickNumber      int                                `json:"tn"`
	CurrentPlayerIdx int                                `json:"ci"`
	CurrentTrick     []*TrickCard                       `json:"ct"`
	DealerIdx        int                                `json:"di"`
	TrumpSuit        int                                `json:"ts"`
	HighestBid       int                                `json:"hb"`
	HighestBidder    int                                `json:"hd"`
	BidPlayerIdx     int                                `json:"bi"`
	TeamScores       [PinochleTeamCnt]int               `json:"sc"`
	LeadPlayerIdx    int                                `json:"li"`
	GameEndFlag      bool                               `json:"ge"`
	WinnerTeam       int                                `json:"wt"`
	PlayerMelds      [PinochlePlayerCnt][]*PinochleMeld `json:"pm"`
	ActionLog        []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *Pinochle) MarshalJSON() ([]byte, error) {
	return json.Marshal(pinochleJSON{
		TrumpCards:       p.trumpCards,
		Players:          p.players,
		Config:           p.config,
		Phase:            p.phase,
		RoundNumber:      p.roundNumber,
		TrickNumber:      p.trickNumber,
		CurrentPlayerIdx: p.currentPlayerIdx,
		CurrentTrick:     p.currentTrick,
		DealerIdx:        p.dealerIdx,
		TrumpSuit:        p.trumpSuit,
		HighestBid:       p.highestBid,
		HighestBidder:    p.highestBidder,
		BidPlayerIdx:     p.bidPlayerIdx,
		TeamScores:       p.teamScores,
		LeadPlayerIdx:    p.leadPlayerIdx,
		GameEndFlag:      p.gameEndFlag,
		WinnerTeam:       p.winnerTeam,
		PlayerMelds:      p.playerMelds,
		ActionLog:        p.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Pinochle) UnmarshalJSON(data []byte) error {
	var j pinochleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCardsPinochle()
	}
	p.players = j.Players
	p.config = j.Config
	p.phase = j.Phase
	p.roundNumber = j.RoundNumber
	p.trickNumber = j.TrickNumber
	p.currentPlayerIdx = j.CurrentPlayerIdx
	p.currentTrick = j.CurrentTrick
	p.dealerIdx = j.DealerIdx
	p.trumpSuit = j.TrumpSuit
	p.highestBid = j.HighestBid
	p.highestBidder = j.HighestBidder
	p.bidPlayerIdx = j.BidPlayerIdx
	p.teamScores = j.TeamScores
	p.leadPlayerIdx = j.LeadPlayerIdx
	p.gameEndFlag = j.GameEndFlag
	p.winnerTeam = j.WinnerTeam
	p.playerMelds = j.PlayerMelds
	p.actionLog = j.ActionLog
	return nil
}

// ─── Sort ───────────────────────────────────────────────

// SortHand 手札をソートする (スート順 → ランク順)
// ReorderCards を使用してインプレースで並び替えることで Reset+AddCard の副作用を回避する
func (p *Pinochle) SortHand(playerIdx int) {
	player := p.players[playerIdx]
	n := player.GetCardsSize()
	if n <= 1 {
		return
	}

	// ソート順のインデックスを生成
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		ca := player.GetCard(indices[a])
		cb := player.GetCard(indices[b])
		if ca.GetDesign() != cb.GetDesign() {
			return ca.GetDesign() < cb.GetDesign()
		}
		return pinochleRankValue(ca.GetValue()) > pinochleRankValue(cb.GetValue())
	})
	_ = player.ReorderCards(indices)
}

// GetValidPlayIndices プレイ可能なカードのインデックスを返す (外部公開用)
func (p *Pinochle) GetValidPlayIndices(playerIdx int) []int {
	return p.getValidPlayIndices(playerIdx)
}
