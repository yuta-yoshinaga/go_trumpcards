package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"sort"
)

// BinokelPlayerCnt ビノクルプレイヤー数 (3人)
const BinokelPlayerCnt = 3

// BinokelHandSize 各プレイヤーの手札枚数 (15枚)
const BinokelHandSize = 15

// BinokelDabbSize 伏せ札(Dabb)の枚数 (3枚)
// 48枚デッキで3人プレイのため、48 = 15×3 + 3 でDabbは3枚となる。
// (4枚のDabbは7を除いた40枚版(3×12+4=40)の規則であり、本実装は48枚を用いるため3枚が正しい)
const BinokelDabbSize = 3

// BinokelMinBid 最低ビッド (150点)
const BinokelMinBid = 150

// BinokelBidStep ビッドの最小刻み幅 (10点)
const BinokelBidStep = 10

// BinokelTotalCardPoints ラウンドあたりの合計カードポイント (最終トリックボーナス含む)
const BinokelTotalCardPoints = 250

// BinokelLastTrickBonus 最終トリックボーナス
const BinokelLastTrickBonus = 10

// BinokelPhase ゲームフェーズ
type BinokelPhase int

// ビノクルのフェーズ定数
const (
	// BinokelPhaseBid ビッドフェーズ
	BinokelPhaseBid BinokelPhase = 0
	// BinokelPhaseDabb 伏せ札(Dabb)交換フェーズ
	BinokelPhaseDabb BinokelPhase = 1
	// BinokelPhaseTrump トランプ宣言フェーズ
	BinokelPhaseTrump BinokelPhase = 2
	// BinokelPhaseMeld メルドフェーズ
	BinokelPhaseMeld BinokelPhase = 3
	// BinokelPhasePlay トリックプレイフェーズ
	BinokelPhasePlay BinokelPhase = 4
	// BinokelPhaseTrickEnd トリック終了フェーズ
	BinokelPhaseTrickEnd BinokelPhase = 5
	// BinokelPhaseRoundEnd ラウンド終了フェーズ
	BinokelPhaseRoundEnd BinokelPhase = 6
	// BinokelPhaseGameEnd ゲーム終了フェーズ
	BinokelPhaseGameEnd BinokelPhase = 7
)

// BinokelMeldType メルドの種類
type BinokelMeldType int

// メルド種類定数
const (
	BinokelMeldDix                BinokelMeldType = iota // 7 of trump (10)
	BinokelMeldCommonMarriage                            // K-Q of non-trump (20)
	BinokelMeldRoyalMarriage                             // K-Q of trump (40)
	BinokelMeldBinokel                                   // J♦ + Q♠ (40)
	BinokelMeldJacksAround                               // One J of each suit (40)
	BinokelMeldQueensAround                              // One Q of each suit (60)
	BinokelMeldKingsAround                               // One K of each suit (80)
	BinokelMeldAcesAround                                // One A of each suit (100)
	BinokelMeldNonTrumpRun                               // A-10-K-Q-J of non-trump (100)
	BinokelMeldRun                                       // A-10-K-Q-J of trump (150)
	BinokelMeldRundgang                                  // K-Q in all 4 suits (240)
	BinokelMeldDoubleBinokel                             // 2x J♦ + 2x Q♠ (300)
	BinokelMeldDoubleJacksAround                         // 2x J of each suit (400)
	BinokelMeldDoubleQueensAround                        // 2x Q of each suit (600)
	BinokelMeldDoubleKingsAround                         // 2x K of each suit (800)
	BinokelMeldDoubleAcesAround                          // 2x A of each suit (1000)
	BinokelMeldDoubleNonTrumpRun                         // 2x A-10-K-Q-J of non-trump (1000)
	BinokelMeldDoubleRun                                 // 2x A-10-K-Q-J of trump (1500)
)

// 別名定義 (ファミリー呼び出し対応)
const (
	BinokelMeldNonTrumpFamily       = BinokelMeldNonTrumpRun
	BinokelMeldDoubleNonTrumpFamily = BinokelMeldDoubleNonTrumpRun
)

// BinokelMeld 検出されたメルド
type BinokelMeld struct {
	Type   BinokelMeldType `json:"t"`
	Points int             `json:"p"`
	Cards  []*Card         `json:"c"`
}

// binokelMeldPoints メルド種類ごとのポイント
var binokelMeldPoints = map[BinokelMeldType]int{
	BinokelMeldDix:                10,
	BinokelMeldCommonMarriage:     20,
	BinokelMeldRoyalMarriage:      40,
	BinokelMeldBinokel:            40,
	BinokelMeldJacksAround:        40,
	BinokelMeldQueensAround:       60,
	BinokelMeldKingsAround:        80,
	BinokelMeldAcesAround:         100,
	BinokelMeldNonTrumpRun:        100,
	BinokelMeldRun:                150,
	BinokelMeldRundgang:           240,
	BinokelMeldDoubleBinokel:      300,
	BinokelMeldDoubleJacksAround:  400,
	BinokelMeldDoubleQueensAround: 600,
	BinokelMeldDoubleKingsAround:  800,
	BinokelMeldDoubleAcesAround:   1000,
	BinokelMeldDoubleNonTrumpRun:  1000,
	BinokelMeldDoubleRun:          1500,
}

// BinokelMeldTableEntry は早見表の1行。メルドの種類とその点数。
type BinokelMeldTableEntry struct {
	Type   BinokelMeldType
	Points int
}

// BinokelMeldTable はメルド種類の点数一覧を安い順で返す。同点のときは
// 種類の定義順。
//
// **表示側が点数を書き写さないためにある。**書き写した表は binokelMeldPoints
// を1つ直した瞬間に黙って食い違い、プレイヤーはビッドの見積もりを誤った表で
// 立てることになる (#5519)。
func BinokelMeldTable() []BinokelMeldTableEntry {
	table := make([]BinokelMeldTableEntry, 0, len(binokelMeldPoints))
	for t, p := range binokelMeldPoints {
		table = append(table, BinokelMeldTableEntry{Type: t, Points: p})
	}
	sort.Slice(table, func(i, j int) bool {
		if table[i].Points != table[j].Points {
			return table[i].Points < table[j].Points
		}
		return table[i].Type < table[j].Type
	})
	return table
}

// BinokelHint ヒント情報
type BinokelHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	BidAmount *int   // 推奨ビッド額 (ビッド時)
	Pass      *bool  // パスすべきか
	Suit      *int   // 推奨スート (トランプ宣言時)
	Reason    string // ヒント理由キー
}

// Binokel ビノクルゲームクラス
type Binokel struct {
	trumpCards       *TrumpCards
	players          []*BinokelPlayer
	config           BinokelConfig
	phase            BinokelPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	trumpSuit        int // 切り札スート (CardDesignSpade等)
	highestBid       int // 現在の最高ビッド
	highestBidder    int // 最高ビッダーのインデックス
	bidPlayerIdx     int // 現在のビッド手番
	scores           [BinokelPlayerCnt]int
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerPlayer     int // 勝利プレイヤー (-1 = 未確定)
	playerMelds      [BinokelPlayerCnt][]*BinokelMeld
	actionLog        []*ActionLogEntry
	dabb             []*Card // 配られたDabb (3枚)
	dabbDiscarded    []*Card // 落札者が捨てたDabb (3枚)
	lastBidSpeaker   int     // 最後にビッドで発言したプレイヤー
}

// NewBinokel コンストラクタ
func NewBinokel(trumpCards *TrumpCards, players []*BinokelPlayer, config BinokelConfig) *Binokel {
	return &Binokel{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerPlayer:   -1,
		roundNumber:    0,
		dealerIdx:      0,
		highestBidder:  -1,
		lastBidSpeaker: -1,
	}
}

// NewDefaultBinokel returns Binokel with the standard 3-player individual setup
// (human player 0, CPU 1, CPU 2) and DefaultBinokelConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBinokel() *Binokel {
	players := []*BinokelPlayer{
		NewBinokelPlayer(true),
		NewBinokelPlayer(false),
		NewBinokelPlayer(false),
	}
	return NewBinokel(NewTrumpCardsBinokel(), players, DefaultBinokelConfig())
}

// Reset ゲーム初期化
func (p *Binokel) Reset() {
	p.gameEndFlag = false
	p.winnerPlayer = -1
	p.roundNumber = 1
	p.trickNumber = 0
	p.dealerIdx = 0
	p.scores = [BinokelPlayerCnt]int{}
	p.actionLog = nil
	p.trumpSuit = 0
	p.highestBid = 0
	p.highestBidder = -1
	p.playerMelds = [BinokelPlayerCnt][]*BinokelMeld{}
	p.dabb = nil
	p.dabbDiscarded = nil
	p.lastBidSpeaker = -1

	for _, pl := range p.players {
		pl.ResetRound()
	}

	p.dealRound()
	p.phase = BinokelPhaseBid
	p.bidPlayerIdx = (p.dealerIdx + 1) % BinokelPlayerCnt
}

// NextRound 次のラウンドを開始する
func (p *Binokel) NextRound() {
	if p.phase != BinokelPhaseRoundEnd {
		return
	}
	p.roundNumber++
	p.trickNumber = 0
	p.dealerIdx = (p.dealerIdx + 1) % BinokelPlayerCnt
	p.trumpSuit = 0
	p.highestBid = 0
	p.highestBidder = -1
	p.playerMelds = [BinokelPlayerCnt][]*BinokelMeld{}
	p.dabb = nil
	p.dabbDiscarded = nil
	p.lastBidSpeaker = -1

	for _, pl := range p.players {
		pl.ResetRound()
	}

	p.dealRound()
	p.phase = BinokelPhaseBid
	p.bidPlayerIdx = (p.dealerIdx + 1) % BinokelPlayerCnt
}

// dealRound ラウンドのカードを配る
func (p *Binokel) dealRound() {
	p.trumpCards.Shuffle()
	for range BinokelHandSize {
		for j := range BinokelPlayerCnt {
			card := p.trumpCards.DrawCard()
			if card != nil {
				p.players[j].AddCard(card)
			}
		}
	}
	p.dabb = make([]*Card, 0, BinokelDabbSize)
	for range BinokelDabbSize {
		card := p.trumpCards.DrawCard()
		if card != nil {
			p.dabb = append(p.dabb, card)
		}
	}
	p.dabbDiscarded = nil
	p.currentTrick = nil
}

// ─── Getters ────────────────────────────────────────────

// GetPhase フェーズを取得
func (p *Binokel) GetPhase() BinokelPhase { return p.phase }

// GetRoundNumber ラウンド番号を取得
func (p *Binokel) GetRoundNumber() int { return p.roundNumber }

// GetTrickNumber トリック番号を取得
func (p *Binokel) GetTrickNumber() int { return p.trickNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得
func (p *Binokel) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// GetCurrentTrick 現在のトリックを取得
func (p *Binokel) GetCurrentTrick() []*TrickCard { return p.currentTrick }

// GetDealerIdx ディーラーインデックスを取得
func (p *Binokel) GetDealerIdx() int { return p.dealerIdx }

// GetTrumpSuit 切り札スートを取得
func (p *Binokel) GetTrumpSuit() int { return p.trumpSuit }

// GetHighestBid 最高ビッド額を取得
func (p *Binokel) GetHighestBid() int { return p.highestBid }

// GetHighestBidder 最高ビッダーインデックスを取得
func (p *Binokel) GetHighestBidder() int { return p.highestBidder }

// GetBidPlayerIdx ビッド手番インデックスを取得
func (p *Binokel) GetBidPlayerIdx() int { return p.bidPlayerIdx }

// GetScores 全プレイヤーのスコアを取得
func (p *Binokel) GetScores() [BinokelPlayerCnt]int { return p.scores }

// GetScore 指定プレイヤーのスコアを取得
func (p *Binokel) GetScore(playerIdx int) int {
	if playerIdx >= 0 && playerIdx < len(p.scores) {
		return p.scores[playerIdx]
	}
	return 0
}

// GetWinnerPlayer 勝利プレイヤーインデックスを取得 (-1 = 未確定)
func (p *Binokel) GetWinnerPlayer() int { return p.winnerPlayer }

// GetLeadPlayerIdx リードプレイヤーインデックスを取得
func (p *Binokel) GetLeadPlayerIdx() int { return p.leadPlayerIdx }

// GetGameEndFlag ゲーム終了フラグを取得
func (p *Binokel) GetGameEndFlag() bool { return p.gameEndFlag }

// IsHumanTurn 現在の手番が人間かを返す
func (p *Binokel) IsHumanTurn() bool {
	if p.phase == BinokelPhaseDabb {
		return p.IsHumanDabbTurn()
	}
	if p.phase == BinokelPhaseBid {
		return p.IsHumanBidTurn()
	}
	if p.currentPlayerIdx < 0 || p.currentPlayerIdx >= len(p.players) {
		return false
	}
	return p.players[p.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かを返す
func (p *Binokel) IsHumanBidTurn() bool {
	if p.bidPlayerIdx < 0 || p.bidPlayerIdx >= len(p.players) {
		return false
	}
	return p.players[p.bidPlayerIdx].GetIsHuman()
}

// IsHumanDabbTurn 現在のDabb手番が人間かを返す (フェーズも見ること)
func (p *Binokel) IsHumanDabbTurn() bool {
	if p.phase != BinokelPhaseDabb {
		return false
	}
	if p.currentPlayerIdx < 0 || p.currentPlayerIdx >= len(p.players) {
		return false
	}
	return p.players[p.currentPlayerIdx].GetIsHuman()
}

// GetDabb 配られたDabb(3枚)を取得
func (p *Binokel) GetDabb() []*Card { return p.dabb }

// GetDabbDiscarded Dabbに捨てられたカード(3枚)を取得
func (p *Binokel) GetDabbDiscarded() []*Card { return p.dabbDiscarded }

// GetPlayerCnt プレイヤー数を取得する
func (p *Binokel) GetPlayerCnt() int { return BinokelPlayerCnt }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (p *Binokel) GetPlayer(i int) *BinokelPlayer { return p.players[i] }

// GetTeamScore チームスコアを取得する (外側の層・interfaces との後方互換性のためのスタブ)
// 個人戦への移行に伴い、team インデックスをプレイヤーインデックスとしてスコアを返す。
func (p *Binokel) GetTeamScore(team int) int { return p.GetScore(team) }

// GetWinnerTeam 勝利チームを取得 (外側の層・interfaces との後方互換性のためのスタブ)
// 個人戦への移行に伴い、勝利プレイヤーインデックスを返す。
func (p *Binokel) GetWinnerTeam() int { return p.winnerPlayer }

// GetPlayers プレイヤーを取得
func (p *Binokel) GetPlayers() []*BinokelPlayer { return p.players }

// GetConfig 設定を取得
func (p *Binokel) GetConfig() BinokelConfig { return p.config }

// SetConfig 設定を変更
func (p *Binokel) SetConfig(config BinokelConfig) { p.config = config }

// GetPlayerMelds プレイヤーのメルドを取得
func (p *Binokel) GetPlayerMelds() [BinokelPlayerCnt][]*BinokelMeld { return p.playerMelds }

// GetActionLog アクションログを取得
func (p *Binokel) GetActionLog() []*ActionLogEntry { return p.actionLog }

// addLog アクションログを追加
func (p *Binokel) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: p.trickNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ─── Card Ranking & Point Values ────────────────────────

// binokelRankValue ビノクルでのカードランク値 (A > 10 > K > Q > J > 7)
func binokelRankValue(value int) int {
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
	case 7: // 7 (Dix)
		return 1
	default:
		return 0
	}
}

// cardRank トリック比較用のカードランクを返す
// trump > lead suit > other suits, 同スートならランク順
func (p *Binokel) cardRank(card *Card) int {
	base := binokelRankValue(card.GetValue())
	suit := card.GetDesign()

	if suit == p.trumpSuit {
		return 400 + base
	}
	if len(p.currentTrick) > 0 && suit == p.currentTrick[0].Card.GetDesign() {
		return 200 + base
	}
	return 100 + base
}

// binokelCardPointValue カードのポイント値を返す
func binokelCardPointValue(card *Card) int {
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
	default: // 7
		return 0
	}
}

// ─── Meld Evaluation ────────────────────────────────────

// evaluateBinokelMelds 手札からメルドを検出する
func evaluateBinokelMelds(hand []*Card, trumpSuit int) []*BinokelMeld {
	type sv struct{ suit, value int }
	counts := make(map[sv]int)
	cardMap := make(map[sv][]*Card)
	for _, c := range hand {
		key := sv{c.GetDesign(), c.GetValue()}
		counts[key]++
		cardMap[key] = append(cardMap[key], c)
	}

	var melds []*BinokelMeld
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}

	// ── Runs (Trump: DoubleRun 1500 / Run 150; Non-Trump: DoubleRun 1000 / Run 100) ──
	runValues := []int{1, 10, 13, 12, 11} // A, 10, K, Q, J
	runCounts := make(map[int]int)
	for _, s := range suits {
		minC := 3
		for _, v := range runValues {
			c := counts[sv{s, v}]
			if c < minC {
				minC = c
			}
		}
		runCounts[s] = minC
		if s == trumpSuit {
			if minC >= 2 {
				var cards []*Card
				for _, v := range runValues {
					cards = append(cards, cardMap[sv{s, v}]...)
				}
				melds = append(melds, &BinokelMeld{
					Type:   BinokelMeldDoubleRun,
					Points: binokelMeldPoints[BinokelMeldDoubleRun],
					Cards:  cards,
				})
			} else if minC >= 1 {
				var cards []*Card
				for _, v := range runValues {
					cards = append(cards, cardMap[sv{s, v}][0])
				}
				melds = append(melds, &BinokelMeld{
					Type:   BinokelMeldRun,
					Points: binokelMeldPoints[BinokelMeldRun],
					Cards:  cards,
				})
			}
		} else {
			if minC >= 2 {
				var cards []*Card
				for _, v := range runValues {
					cards = append(cards, cardMap[sv{s, v}]...)
				}
				melds = append(melds, &BinokelMeld{
					Type:   BinokelMeldDoubleNonTrumpRun,
					Points: binokelMeldPoints[BinokelMeldDoubleNonTrumpRun],
					Cards:  cards,
				})
			} else if minC >= 1 {
				var cards []*Card
				for _, v := range runValues {
					cards = append(cards, cardMap[sv{s, v}][0])
				}
				melds = append(melds, &BinokelMeld{
					Type:   BinokelMeldNonTrumpRun,
					Points: binokelMeldPoints[BinokelMeldNonTrumpRun],
					Cards:  cards,
				})
			}
		}
	}

	// ── Double Around / Around (Aces 1000/100, Kings 800/80, Queens 600/60, Jacks 400/40) ──
	type aroundDef struct {
		value      int
		singleType BinokelMeldType
		doubleType BinokelMeldType
	}
	arounds := []aroundDef{
		{1, BinokelMeldAcesAround, BinokelMeldDoubleAcesAround},
		{13, BinokelMeldKingsAround, BinokelMeldDoubleKingsAround},
		{12, BinokelMeldQueensAround, BinokelMeldDoubleQueensAround},
		{11, BinokelMeldJacksAround, BinokelMeldDoubleJacksAround},
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
			melds = append(melds, &BinokelMeld{
				Type:   a.doubleType,
				Points: binokelMeldPoints[a.doubleType],
				Cards:  cards,
			})
		} else if minCount >= 1 {
			var cards []*Card
			for _, s := range suits {
				cards = append(cards, cardMap[sv{s, a.value}][0])
			}
			melds = append(melds, &BinokelMeld{
				Type:   a.singleType,
				Points: binokelMeldPoints[a.singleType],
				Cards:  cards,
			})
		}
	}

	// ── Double Binokel (300) / Binokel (40): J♦ + Q♠ ──
	jdCount := counts[sv{CardDesignDiamond, 11}]
	qsCount := counts[sv{CardDesignSpade, 12}]
	binokelCount := min(jdCount, qsCount)
	if binokelCount >= 2 {
		var cards []*Card
		cards = append(cards, cardMap[sv{CardDesignDiamond, 11}]...)
		cards = append(cards, cardMap[sv{CardDesignSpade, 12}]...)
		melds = append(melds, &BinokelMeld{
			Type:   BinokelMeldDoubleBinokel,
			Points: binokelMeldPoints[BinokelMeldDoubleBinokel],
			Cards:  cards,
		})
	} else if binokelCount >= 1 {
		melds = append(melds, &BinokelMeld{
			Type:   BinokelMeldBinokel,
			Points: binokelMeldPoints[BinokelMeldBinokel],
			Cards:  []*Card{cardMap[sv{CardDesignDiamond, 11}][0], cardMap[sv{CardDesignSpade, 12}][0]},
		})
	}

	// ── Rundgang (240) & Marriages (Royal 40 / Common 20) ──
	// Rundgang = 4スートすべてに K+Q の組。
	// Rundgang が成立するとき、その 4つの Paar は二重計上しない。
	// また、同一スートの Run に含まれる K+Q は Run に包含されるため個別の Paar は加算しない。
	totalMarriages := make(map[int]int)
	for _, s := range suits {
		kc := counts[sv{s, 13}]
		qc := counts[sv{s, 12}]
		totalMarriages[s] = min(kc, qc)
	}

	rundgangCount := 3
	for _, s := range suits {
		if totalMarriages[s] < rundgangCount {
			rundgangCount = totalMarriages[s]
		}
	}

	for r := 0; r < rundgangCount; r++ {
		var cards []*Card
		for _, s := range suits {
			cards = append(cards, cardMap[sv{s, 13}][r], cardMap[sv{s, 12}][r])
		}
		melds = append(melds, &BinokelMeld{
			Type:   BinokelMeldRundgang,
			Points: binokelMeldPoints[BinokelMeldRundgang],
			Cards:  cards,
		})
	}

	// Rundgang で消費されなかった残りの Marriage を評価
	for _, s := range suits {
		remaining := totalMarriages[s] - rundgangCount
		subsumedByRun := max(0, runCounts[s]-rundgangCount)
		unsubsumed := max(0, remaining-subsumedByRun)

		for i := 0; i < unsubsumed; i++ {
			pairIdx := rundgangCount + subsumedByRun + i
			meldType := BinokelMeldCommonMarriage
			if s == trumpSuit {
				meldType = BinokelMeldRoyalMarriage
			}
			var cards []*Card
			if pairIdx < len(cardMap[sv{s, 13}]) {
				cards = append(cards, cardMap[sv{s, 13}][pairIdx])
			}
			if pairIdx < len(cardMap[sv{s, 12}]) {
				cards = append(cards, cardMap[sv{s, 12}][pairIdx])
			}
			melds = append(melds, &BinokelMeld{
				Type:   meldType,
				Points: binokelMeldPoints[meldType],
				Cards:  cards,
			})
		}
	}

	// ── Dix (10): 7 of trump ──
	dixCount := counts[sv{trumpSuit, 7}]
	for i := range dixCount {
		melds = append(melds, &BinokelMeld{
			Type:   BinokelMeldDix,
			Points: binokelMeldPoints[BinokelMeldDix],
			Cards:  []*Card{cardMap[sv{trumpSuit, 7}][i]},
		})
	}

	return melds
}

// binokelMeldTotalPoints メルドの合計ポイントを計算
func binokelMeldTotalPoints(melds []*BinokelMeld) int {
	total := 0
	for _, m := range melds {
		total += m.Points
	}
	return total
}

// ─── Bidding ─────────────────────────────────────────────

// PlayerBid 人間プレイヤーがビッドする
func (p *Binokel) PlayerBid(amount int) error {
	if p.phase != BinokelPhaseBid {
		return NewDomainError(ErrWrongPhase, "ビッドフェーズではありません")
	}
	if !p.players[p.bidPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doBid(p.bidPlayerIdx, amount)
}

// PlayerPass 人間プレイヤーがパスする
func (p *Binokel) PlayerPass() error {
	if p.phase != BinokelPhaseBid {
		return NewDomainError(ErrWrongPhase, "ビッドフェーズではありません")
	}
	if !p.players[p.bidPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doPass(p.bidPlayerIdx)
}

// doBid ビッドを実行
func (p *Binokel) doBid(playerIdx, amount int) error {
	if amount < BinokelMinBid {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("ビッドは%d以上でなければなりません", BinokelMinBid))
	}
	if p.highestBid > 0 && amount <= p.highestBid {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("現在のビッド%dより大きくなければなりません", p.highestBid))
	}
	if (amount-BinokelMinBid)%BinokelBidStep != 0 {
		return NewDomainError(ErrInvalidAmount, fmt.Sprintf("ビッドは%d刻みでなければなりません", BinokelBidStep))
	}

	p.lastBidSpeaker = playerIdx
	p.players[playerIdx].SetBid(amount)
	p.highestBid = amount
	p.highestBidder = playerIdx
	p.addLog(playerIdx, "bid", fmt.Sprintf("%sがビッド: %d", playerName(p.players, playerIdx), amount), nil)
	p.advanceBidder()
	return nil
}

// doPass パスを実行
func (p *Binokel) doPass(playerIdx int) error {
	p.lastBidSpeaker = playerIdx
	p.players[playerIdx].SetHasPassed(true)
	p.addLog(playerIdx, "pass", fmt.Sprintf("%sがパスしました", playerName(p.players, playerIdx)), nil)
	p.advanceBidder()
	return nil
}

// advanceBidder 次のビッダーに進む
func (p *Binokel) advanceBidder() {
	activeBidders := 0
	lastActive := -1
	for i := range BinokelPlayerCnt {
		if !p.players[i].GetHasPassed() {
			activeBidders++
			lastActive = i
		}
	}

	if activeBidders == 1 && p.highestBid > 0 {
		// ビッド終了 → 落札者決定、Dabbフェーズへ
		p.finishBidding(lastActive)
		return
	}

	// 全員パスした場合 (ビッドなし) → 最後に発言した席が最低額 150 で強制的に落札者になる。
	// (再配りにしないのは決定的でテスト可能な動作を保証するため)
	if activeBidders == 0 {
		forcedBidder := p.lastBidSpeaker
		if forcedBidder < 0 || forcedBidder >= BinokelPlayerCnt {
			forcedBidder = p.dealerIdx
		}
		p.highestBid = BinokelMinBid
		p.players[forcedBidder].SetBid(BinokelMinBid)
		p.addLog(forcedBidder, "forced_bid", fmt.Sprintf("全員パスのため最後の発言者が強制落札: %d点", BinokelMinBid), nil)
		p.finishBidding(forcedBidder)
		return
	}

	// 次のビッダーに進む
	next := (p.bidPlayerIdx + 1) % BinokelPlayerCnt
	for p.players[next].GetHasPassed() {
		next = (next + 1) % BinokelPlayerCnt
	}
	p.bidPlayerIdx = next
}

// finishBidding ビッド終了時の落札者確定とDabb公開
func (p *Binokel) finishBidding(bidder int) {
	p.highestBidder = bidder
	// Dabbの3枚を全員に見える形で公開する (ログにも残す)
	p.addLog(bidder, "dabb_reveal", fmt.Sprintf("Dabb公開: %d枚", len(p.dabb)), p.dabb)
	// 落札者はDabbの3枚を手札に加える (15枚 + 3枚 = 18枚)
	for _, c := range p.dabb {
		p.players[bidder].AddCard(c)
	}
	p.phase = BinokelPhaseDabb
	p.currentPlayerIdx = bidder
}

// ─── Dabb Phase ──────────────────────────────────────────

// PlayerDiscardToDabb 人間落札者がDabbへ3枚伏せて捨てる
func (p *Binokel) PlayerDiscardToDabb(cardIndices []int) error {
	if p.phase != BinokelPhaseDabb {
		return NewDomainError(ErrWrongPhase, "Dabbフェーズではありません")
	}
	if !p.IsHumanDabbTurn() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doDiscardToDabb(p.currentPlayerIdx, cardIndices)
}

// CpuDiscardToDabb CPU落札者がDabbへ3枚捨てる
func (p *Binokel) CpuDiscardToDabb() {
	if p.phase != BinokelPhaseDabb {
		return
	}
	playerIdx := p.currentPlayerIdx
	if p.players[playerIdx].GetIsHuman() {
		return
	}
	indices := p.cpuChooseDabbDiscards(playerIdx)
	_ = p.doDiscardToDabb(playerIdx, indices)
}

func (p *Binokel) doDiscardToDabb(playerIdx int, cardIndices []int) error {
	if len(cardIndices) != BinokelDabbSize {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("%d枚のカードを選択してください", BinokelDabbSize))
	}
	player := p.players[playerIdx]
	handSize := player.GetCardsSize()

	seen := make(map[int]bool)
	for _, idx := range cardIndices {
		if idx < 0 || idx >= handSize {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードが複数回指定されています")
		}
		seen[idx] = true
	}

	sortedIndices := make([]int, len(cardIndices))
	copy(sortedIndices, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIndices)))

	discarded := make([]*Card, 0, BinokelDabbSize)
	for _, idx := range sortedIndices {
		c := player.GetCard(idx)
		discarded = append(discarded, c)
		player.RemoveCard(idx)
	}
	p.dabbDiscarded = discarded

	p.addLog(playerIdx, "dabb_discard", fmt.Sprintf("%sがDabbへ%d枚伏せて捨てました", playerName(p.players, playerIdx), BinokelDabbSize), nil)

	p.phase = BinokelPhaseTrump
	p.currentPlayerIdx = p.highestBidder
	return nil
}

func (p *Binokel) cpuChooseDabbDiscards(playerIdx int) []int {
	player := p.players[playerIdx]
	bestSuit, _ := p.cpuEvalBestTrump(playerIdx)

	type scoredCard struct {
		idx   int
		score int
	}
	var scored []scoredCard
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		score := 0
		if card.GetDesign() == bestSuit {
			score += 100
		}
		switch card.GetValue() {
		case 1:
			score += 40
		case 13, 12:
			score += 30
		case 10:
			score += 20
		case 11:
			score += 10
		case 7:
			score += 0
		}
		scored = append(scored, scoredCard{idx: i, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score < scored[j].score
	})

	indices := make([]int, BinokelDabbSize)
	for i := 0; i < BinokelDabbSize; i++ {
		indices[i] = scored[i].idx
	}
	return indices
}

// ─── CPU Bid Evaluation ──────────────────────────────────

// CpuBid CPUがビッドまたはパスする
func (p *Binokel) CpuBid() {
	if p.phase != BinokelPhaseBid {
		return
	}
	playerIdx := p.bidPlayerIdx
	if p.players[playerIdx].GetIsHuman() {
		return
	}

	switch p.config.CpuDifficulty {
	case BinokelCpuDifficultyEasy:
		p.cpuBidEasy(playerIdx)
	case BinokelCpuDifficultyNormal:
		p.cpuBidNormal(playerIdx)
	case BinokelCpuDifficultyHard:
		p.cpuBidHard(playerIdx)
	}
}

func (p *Binokel) cpuBidEasy(playerIdx int) {
	totalEstimate := p.cpuEstimateBid(playerIdx)
	if totalEstimate < BinokelMinBid {
		_ = p.doPass(playerIdx)
		return
	}
	if rand.Intn(100) < 70 {
		nextBid := BinokelMinBid
		if p.highestBid > 0 {
			nextBid = p.highestBid + BinokelBidStep
		}
		if nextBid <= totalEstimate {
			_ = p.doBid(playerIdx, nextBid)
			return
		}
	}
	_ = p.doPass(playerIdx)
}

func (p *Binokel) cpuBidNormal(playerIdx int) {
	totalEstimate := p.cpuEstimateBid(playerIdx)

	if totalEstimate < BinokelMinBid {
		_ = p.doPass(playerIdx)
		return
	}

	nextBid := BinokelMinBid
	if p.highestBid > 0 {
		nextBid = p.highestBid + BinokelBidStep
	}

	if nextBid > totalEstimate {
		_ = p.doPass(playerIdx)
		return
	}

	_ = p.doBid(playerIdx, nextBid)
}

func (p *Binokel) cpuBidHard(playerIdx int) {
	totalEstimate := p.cpuEstimateBid(playerIdx)

	if totalEstimate < BinokelMinBid {
		_ = p.doPass(playerIdx)
		return
	}

	nextBid := BinokelMinBid
	if p.highestBid > 0 {
		nextBid = p.highestBid + BinokelBidStep
	}

	maxBid := totalEstimate
	if totalEstimate >= 200 {
		maxBid = int(float64(totalEstimate) * 1.05)
	}

	if nextBid > maxBid {
		_ = p.doPass(playerIdx)
		return
	}

	_ = p.doBid(playerIdx, nextBid)
}

// cpuEstimateBid プレイヤーの手札からビッド見積もり額 (メルド点 + 期待トリック点) を計算する
func (p *Binokel) cpuEstimateBid(playerIdx int) int {
	bestSuit, bestMeldPoints := p.cpuEvalBestTrump(playerIdx)
	trickEstimate := p.cpuEstimateTrickPoints(playerIdx, bestSuit)
	return bestMeldPoints + trickEstimate
}

// cpuEvalBestTrump CPUが最適なトランプスートを評価する
func (p *Binokel) cpuEvalBestTrump(playerIdx int) (bestSuit, bestMeldPoints int) {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	hand := make([]*Card, 0, p.players[playerIdx].GetCardsSize())
	for i := 0; i < p.players[playerIdx].GetCardsSize(); i++ {
		hand = append(hand, p.players[playerIdx].GetCard(i))
	}

	bestSuit = suits[0]
	bestMeldPoints = -1
	maxTrumps := -1

	for _, s := range suits {
		melds := evaluateBinokelMelds(hand, s)
		total := binokelMeldTotalPoints(melds)
		trumps := 0
		for _, c := range hand {
			if c.GetDesign() == s {
				trumps++
			}
		}
		if total > bestMeldPoints || (total == bestMeldPoints && trumps > maxTrumps) {
			bestMeldPoints = total
			bestSuit = s
			maxTrumps = trumps
		}
	}
	return
}

// cpuEstimateTrickPoints CPUがトリックポイントを推定する
func (p *Binokel) cpuEstimateTrickPoints(playerIdx, trumpSuit int) int {
	player := p.players[playerIdx]
	points := 0 // トリック獲得ポイント推定値
	trumpCount := 0

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		val := card.GetValue()
		if card.GetDesign() == trumpSuit {
			trumpCount++
			switch val {
			case 1: // Trump A
				points += 11
			case 10: // Trump 10
				points += 10
			case 13: // Trump K
				points += 4
			case 12: // Trump Q
				points += 3
			case 11: // Trump J
				points += 2
			}
		} else {
			switch val {
			case 1: // Offsuit Ace
				points += 8
			case 10: // Offsuit 10
				points += 3
			}
		}
	}

	if trumpCount >= 5 {
		points += (trumpCount - 4) * 8
	}

	if points > BinokelTotalCardPoints {
		points = BinokelTotalCardPoints
	}
	return points
}

// ─── Trump Declaration ──────────────────────────────────

// PlayerCallTrump 人間プレイヤーがトランプスートを宣言する
func (p *Binokel) PlayerCallTrump(suit int) error {
	if p.phase != BinokelPhaseTrump {
		return NewDomainError(ErrWrongPhase, "トランプ宣言フェーズではありません")
	}
	if !p.players[p.currentPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doCallTrump(p.currentPlayerIdx, suit)
}

// doCallTrump トランプ宣言を実行
func (p *Binokel) doCallTrump(playerIdx, suit int) error {
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

	// メルドフェーズへ移行 (15枚の手札に対してメルド評価)
	p.evaluateAllMelds()
	p.phase = BinokelPhaseMeld
	return nil
}

// CpuCallTrump CPUがトランプを宣言する
func (p *Binokel) CpuCallTrump() {
	if p.phase != BinokelPhaseTrump {
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
func (p *Binokel) evaluateAllMelds() {
	for i := range BinokelPlayerCnt {
		hand := make([]*Card, 0, p.players[i].GetCardsSize())
		for j := 0; j < p.players[i].GetCardsSize(); j++ {
			hand = append(hand, p.players[i].GetCard(j))
		}
		melds := evaluateBinokelMelds(hand, p.trumpSuit)
		p.playerMelds[i] = melds
		p.players[i].SetMeldScore(binokelMeldTotalPoints(melds))
		p.addLog(i, "meld", fmt.Sprintf("メルド: %d点", p.players[i].GetMeldScore()), nil)
	}
}

// ConfirmMelds メルドを確認してプレイフェーズに進む
func (p *Binokel) ConfirmMelds() {
	if p.phase != BinokelPhaseMeld {
		return
	}
	p.trickNumber = 1
	p.leadPlayerIdx = p.highestBidder
	p.currentPlayerIdx = p.leadPlayerIdx
	p.currentTrick = nil
	p.phase = BinokelPhasePlay
}

// ─── Trick-Taking ───────────────────────────────────────

// getValidPlayIndices プレイ可能なカードのインデックスを返す
// ビノクルのルール:
// 1. リードスートに従う (must follow suit)
// 2. リードスートがなければトランプを出す (must trump)
// 3. 可能なら現在の勝者を上回るカードを出す (must win)
func (p *Binokel) getValidPlayIndices(playerIdx int) []int {
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
func (p *Binokel) PlayerPlay(cardIndex int) error {
	if p.phase != BinokelPhasePlay {
		return NewDomainError(ErrWrongPhase, "プレイフェーズではありません")
	}
	if !p.players[p.currentPlayerIdx].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "人間プレイヤーのターンではありません")
	}
	return p.doPlay(p.currentPlayerIdx, cardIndex)
}

// doPlay カードプレイを実行
func (p *Binokel) doPlay(playerIdx, cardIndex int) error {
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
	if len(p.currentTrick) == BinokelPlayerCnt {
		p.phase = BinokelPhaseTrickEnd
	} else {
		p.currentPlayerIdx = (p.currentPlayerIdx + 1) % BinokelPlayerCnt
	}
	return nil
}

// CpuPlay CPUがカードをプレイする
func (p *Binokel) CpuPlay() {
	if p.phase != BinokelPhasePlay {
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
	case BinokelCpuDifficultyEasy:
		cardIndex = validIndices[rand.Intn(len(validIndices))]
	case BinokelCpuDifficultyNormal:
		cardIndex = p.cpuPlayNormal(playerIdx, validIndices)
	case BinokelCpuDifficultyHard:
		cardIndex = p.cpuPlayHard(playerIdx, validIndices)
	default:
		cardIndex = validIndices[rand.Intn(len(validIndices))]
	}

	_ = p.doPlay(playerIdx, cardIndex)
}

// cpuPlayNormal Normal: 基本戦略
func (p *Binokel) cpuPlayNormal(playerIdx int, validIndices []int) int {
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
func (p *Binokel) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := p.players[playerIdx]
	if len(p.currentTrick) == 0 {
		// リード: ポイントカードのAceやトランプをリードする
		bestIdx := validIndices[0]
		bestScore := -1
		for _, vi := range validIndices {
			card := player.GetCard(vi)
			score := binokelRankValue(card.GetValue())
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

	return p.cpuPlayNormal(playerIdx, validIndices)
}

// trickWinnerIdx 現在のトリックの勝者のインデックス (currentTrick内)を返す
func (p *Binokel) trickWinnerIdx() int {
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
func (p *Binokel) trickWinner() int {
	return p.currentTrick[p.trickWinnerIdx()].PlayerIdx
}

// ResolveTrick トリックを解決する
func (p *Binokel) ResolveTrick() {
	if p.phase != BinokelPhaseTrickEnd || len(p.currentTrick) != BinokelPlayerCnt {
		return
	}

	winner := p.trickWinner()

	// トリックのカードポイントを計算
	trickPoints := 0
	trickCards := make([]*Card, 0, BinokelPlayerCnt)
	for _, tc := range p.currentTrick {
		trickPoints += binokelCardPointValue(tc.Card)
		trickCards = append(trickCards, tc.Card)
	}

	// 最終トリックボーナス
	if p.trickNumber == BinokelHandSize {
		trickPoints += BinokelLastTrickBonus
	}

	p.players[winner].AddTrick(trickCards)
	p.players[winner].AddTrickPoints(trickPoints)

	p.addLog(winner, "trick_win", fmt.Sprintf("トリック獲得: %d点", trickPoints), trickCards)

	// 次のトリックまたはラウンド終了
	if p.trickNumber >= BinokelHandSize {
		p.phase = BinokelPhaseRoundEnd
		p.scoreRound()
	} else {
		p.leadPlayerIdx = winner
		p.currentPlayerIdx = winner
	}
}

// NextTrick 次のトリックを開始する
// 注: ResolveTrick() は呼び出し元 (Interactor) が先に実行する。
func (p *Binokel) NextTrick() {
	if p.phase != BinokelPhaseTrickEnd {
		return
	}
	p.currentTrick = nil
	p.currentPlayerIdx = p.leadPlayerIdx
	p.trickNumber++
	p.phase = BinokelPhasePlay
}

// ─── Scoring ────────────────────────────────────────────

// scoreRound ラウンドのスコアを計算する
func (p *Binokel) scoreRound() {
	if p.highestBidder < 0 || p.highestBidder >= BinokelPlayerCnt {
		return
	}

	// 捨てた Dabb 3枚のカードポイントを落札者のトリック点に加算
	dabbPoints := 0
	for _, c := range p.dabbDiscarded {
		dabbPoints += binokelCardPointValue(c)
	}
	p.players[p.highestBidder].AddTrickPoints(dabbPoints)

	// 落札者: メルド点 + トリック点 (+ 捨て札の点) が入札額以上ならその合計を加点、届かなければ入札額を減点 (-bid)
	// 落札者以外: 自分のメルド点 + 自分のトリック点をそのまま加点
	for i := range BinokelPlayerCnt {
		total := p.players[i].GetMeldScore() + p.players[i].GetTrickPoints()
		if i == p.highestBidder {
			if total >= p.highestBid {
				p.scores[i] += total
			} else {
				p.scores[i] -= p.highestBid
			}
		} else {
			p.scores[i] += total
		}
	}

	p.addLog(-1, "round_score",
		fmt.Sprintf("ラウンド終了: 累計スコア=[P0:%d, P1:%d, P2:%d]",
			p.scores[0], p.scores[1], p.scores[2]),
		nil)

	// ゲーム終了チェック
	hasReached := false
	for i := range BinokelPlayerCnt {
		if p.scores[i] >= p.config.PointLimit {
			hasReached = true
			break
		}
	}

	if hasReached {
		p.gameEndFlag = true
		p.phase = BinokelPhaseGameEnd
		winner := 0
		for i := 1; i < BinokelPlayerCnt; i++ {
			if p.scores[i] > p.scores[winner] {
				winner = i
			} else if p.scores[i] == p.scores[winner] && i == p.highestBidder {
				winner = i
			}
		}
		p.winnerPlayer = winner
		p.addLog(winner, "game_end", fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), nil)
	}
}

// ─── Hint ───────────────────────────────────────────────

// Hint ヒントを返す
func (p *Binokel) Hint() *BinokelHint {
	// **人間の手番でなければ答えない。**hintBid / hintTrump / hintPlay は
	// bidPlayerIdx / currentPlayerIdx をそのまま使うので、席が誰かを見ない。
	// Output() が毎レスポンスでこれを呼ぶようになった以上、ガードが無いと
	// **CPU の手番に CPU 自身の手を「推奨手」として人間に見せる** (#4585 のレビュー指摘)。
	switch p.phase {
	case BinokelPhaseBid:
		if !p.IsHumanBidTurn() {
			return nil
		}
		return p.hintBid()
	case BinokelPhaseTrump, BinokelPhasePlay:
		if !p.IsHumanTurn() {
			return nil
		}
		if p.phase == BinokelPhaseTrump {
			return p.hintTrump()
		}
		return p.hintPlay()
	default:
		return nil
	}
}

// GetHint ヒントを取得する
func (p *Binokel) GetHint() *BinokelHint { return p.Hint() }

func (p *Binokel) hintBid() *BinokelHint {
	playerIdx := p.bidPlayerIdx
	totalEstimate := p.cpuEstimateBid(playerIdx)

	if totalEstimate < BinokelMinBid {
		pass := true
		return &BinokelHint{Pass: &pass, Reason: "hint_pass"}
	}

	nextBid := BinokelMinBid
	if p.highestBid > 0 {
		nextBid = p.highestBid + BinokelBidStep
	}

	if nextBid > totalEstimate {
		pass := true
		return &BinokelHint{Pass: &pass, Reason: "hint_pass"}
	}

	return &BinokelHint{BidAmount: &nextBid, Reason: "hint_bid"}
}

func (p *Binokel) hintTrump() *BinokelHint {
	bestSuit, _ := p.cpuEvalBestTrump(p.currentPlayerIdx)
	return &BinokelHint{Suit: &bestSuit, Reason: "hint_trump"}
}

func (p *Binokel) hintPlay() *BinokelHint {
	playerIdx := p.currentPlayerIdx
	validIndices := p.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return nil
	}
	cardIndex := p.cpuPlayHard(playerIdx, validIndices)
	return &BinokelHint{CardIndex: &cardIndex, Reason: "hint_play"}
}

// ─── JSON Serialization ─────────────────────────────────

// binokelJSON is the JSON wire format for Binokel.
type binokelJSON struct {
	TrumpCards       *TrumpCards                      `json:"tc"`
	Players          []*BinokelPlayer                 `json:"ps"`
	Config           BinokelConfig                    `json:"cf"`
	Phase            BinokelPhase                     `json:"ph"`
	RoundNumber      int                              `json:"rn"`
	TrickNumber      int                              `json:"tn"`
	CurrentPlayerIdx int                              `json:"ci"`
	CurrentTrick     []*TrickCard                     `json:"ct"`
	DealerIdx        int                              `json:"di"`
	TrumpSuit        int                              `json:"ts"`
	HighestBid       int                              `json:"hb"`
	HighestBidder    int                              `json:"hd"`
	BidPlayerIdx     int                              `json:"bi"`
	Scores           [BinokelPlayerCnt]int            `json:"sc"`
	LeadPlayerIdx    int                              `json:"li"`
	GameEndFlag      bool                             `json:"ge"`
	WinnerPlayer     int                              `json:"wp"`
	PlayerMelds      [BinokelPlayerCnt][]*BinokelMeld `json:"pm"`
	ActionLog        []*ActionLogEntry                `json:"al"`
	Dabb             []*Card                          `json:"db"`
	DabbDiscarded    []*Card                          `json:"dd"`
	LastBidSpeaker   int                              `json:"ls"`
}

// MarshalJSON implements json.Marshaler.
func (p *Binokel) MarshalJSON() ([]byte, error) {
	return json.Marshal(binokelJSON{
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
		Scores:           p.scores,
		LeadPlayerIdx:    p.leadPlayerIdx,
		GameEndFlag:      p.gameEndFlag,
		WinnerPlayer:     p.winnerPlayer,
		PlayerMelds:      p.playerMelds,
		ActionLog:        p.actionLog,
		Dabb:             p.dabb,
		DabbDiscarded:    p.dabbDiscarded,
		LastBidSpeaker:   p.lastBidSpeaker,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Binokel) UnmarshalJSON(data []byte) error {
	var j binokelJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCardsBinokel()
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
	p.scores = j.Scores
	p.leadPlayerIdx = j.LeadPlayerIdx
	p.gameEndFlag = j.GameEndFlag
	p.winnerPlayer = j.WinnerPlayer
	p.playerMelds = j.PlayerMelds
	p.actionLog = j.ActionLog
	p.dabb = j.Dabb
	p.dabbDiscarded = j.DabbDiscarded
	p.lastBidSpeaker = j.LastBidSpeaker
	return nil
}

// ─── Sort ───────────────────────────────────────────────

// SortHand 手札をソートする (スート順 → ランク順)
// ReorderCards を使用してインプレースで並び替えることで Reset+AddCard の副作用を回避する
func (p *Binokel) SortHand(playerIdx int) {
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
		return binokelRankValue(ca.GetValue()) > binokelRankValue(cb.GetValue())
	})
	_ = player.ReorderCards(indices)
}

// GetValidPlayIndices プレイ可能なカードのインデックスを返す (外部公開用)
func (p *Binokel) GetValidPlayIndices(playerIdx int) []int {
	return p.getValidPlayIndices(playerIdx)
}
