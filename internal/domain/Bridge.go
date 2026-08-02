//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// BridgePlayerCnt ブリッジプレイヤー数
const BridgePlayerCnt = 4

// BridgeHandSize 各プレイヤーの手札枚数
const BridgeHandSize = 13

// BridgeTeamCnt チーム数
const BridgeTeamCnt = 2

// BridgeTotalTricks 1ラウンドの総トリック数
const BridgeTotalTricks = 13

// BridgePhase ゲームフェーズ
type BridgePhase int

// ブリッジのフェーズ定数
const (
	// BridgePhaseBid オークション（ビッド）フェーズ
	BridgePhaseBid BridgePhase = 0
	// BridgePhasePlay トリックプレイフェーズ
	BridgePhasePlay BridgePhase = 1
	// BridgePhaseTrickEnd トリック終了フェーズ
	BridgePhaseTrickEnd BridgePhase = 2
	// BridgePhaseRoundEnd ラウンド終了フェーズ
	BridgePhaseRoundEnd BridgePhase = 3
	// BridgePhaseGameEnd ゲーム終了フェーズ
	BridgePhaseGameEnd BridgePhase = 4
)

// BridgeBidType ビッドの種類
type BridgeBidType int

// ビッド種類定数
const (
	// BridgeBidPass パス
	BridgeBidPass BridgeBidType = 0
	// BridgeBidNormal 通常ビッド (レベル+スート)
	BridgeBidNormal BridgeBidType = 1
	// BridgeBidDouble ダブル
	BridgeBidDouble BridgeBidType = 2
	// BridgeBidRedouble リダブル
	BridgeBidRedouble BridgeBidType = 3
)

// BridgeSuitNoTrump ノートランプを表すスート値
const BridgeSuitNoTrump = 0

// ビッドスート定数 (ランク順: Clubs < Diamonds < Hearts < Spades < NoTrump)
const (
	BridgeBidSuitClub    = 1
	BridgeBidSuitDiamond = 2
	BridgeBidSuitHeart   = 3
	BridgeBidSuitSpade   = 4
	BridgeBidSuitNT      = 5
)

// BridgeBidEntry オークション中の1ビッドエントリ
type BridgeBidEntry struct {
	PlayerIdx int           `json:"pi"`
	BidType   BridgeBidType `json:"bt"`
	Level     int           `json:"lv"` // 1-7 (BidNormal時のみ)
	Suit      int           `json:"su"` // BridgeBidSuitClub-BridgeBidSuitNT (BidNormal時のみ)
}

// BridgeHint ヒント情報
type BridgeHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	BidType   *int   // 推奨ビッド種類
	BidLevel  *int   // 推奨ビッドレベル
	BidSuit   *int   // 推奨ビッドスート
	Reason    string // ヒント理由キー
}

// Bridge ブリッジゲームクラス
type Bridge struct {
	trumpCards       *TrumpCards
	players          []*BridgePlayer
	config           BridgeConfig
	phase            BridgePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	bidPlayerIdx     int // 現在のビッド手番
	bidHistory       []*BridgeBidEntry
	passCount        int // 連続パス数
	contractLevel    int // コントラクトレベル (1-7, 0=未確定)
	contractSuit     int // コントラクトスート (BridgeBidSuitClub-BridgeBidSuitNT)
	doubled          int // 0=なし, 1=ダブル, 2=リダブル
	declarerIdx      int // デクレアラーのインデックス
	dummyIdx         int // ダミーのインデックス
	lastBidderIdx    int // 最後にビッドしたプレイヤー (-1=なし)
	lastBidTeam      int // 最後にビッドしたチーム (-1=なし)
	openingLeadDone  bool
	trumpSuit        int // カードスート値 (-1=NoTrump, CardDesignSpade等)
	vulnerability    [BridgeTeamCnt]bool
	teamScores       [BridgeTeamCnt]int // ラバー累計スコア
	gamesWon         [BridgeTeamCnt]int // 勝利ゲーム数 (2で勝ち)
	belowLine        [BridgeTeamCnt]int // ライン以下 (コントラクトポイント)
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerTeam       int // 勝利チーム (-1 = 未確定)
	actionLog        []*ActionLogEntry
}

// NewBridge コンストラクタ
func NewBridge(trumpCards *TrumpCards, players []*BridgePlayer, config BridgeConfig) *Bridge {
	return &Bridge{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerTeam:    -1,
		declarerIdx:   -1,
		dummyIdx:      -1,
		lastBidderIdx: -1,
		lastBidTeam:   -1,
	}
}

// NewDefaultBridge returns Bridge with the standard 4-player setup
// (North human team 0, East CPU team 1, South CPU team 0, West CPU team 1)
// and DefaultBridgeConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultBridge() *Bridge {
	players := []*BridgePlayer{
		NewBridgePlayer(true, 0),
		NewBridgePlayer(false, 1),
		NewBridgePlayer(false, 0),
		NewBridgePlayer(false, 1),
	}
	return NewBridge(NewTrumpCards(0), players, DefaultBridgeConfig())
}

// Reset ゲーム初期化
func (b *Bridge) Reset() {
	b.gameEndFlag = false
	b.winnerTeam = -1
	b.roundNumber = 1
	b.trickNumber = 0
	b.dealerIdx = 0
	b.teamScores = [BridgeTeamCnt]int{}
	b.gamesWon = [BridgeTeamCnt]int{}
	b.belowLine = [BridgeTeamCnt]int{}
	b.vulnerability = [BridgeTeamCnt]bool{}
	b.actionLog = nil

	for _, p := range b.players {
		p.ResetRound()
	}

	b.dealRound()
	b.startBidPhase()
}

// NextRound 次のラウンドを開始する
func (b *Bridge) NextRound() {
	if b.phase != BridgePhaseRoundEnd {
		return
	}

	b.roundNumber++
	b.dealerIdx = (b.dealerIdx + 1) % BridgePlayerCnt
	b.trickNumber = 0
	b.currentTrick = nil
	b.leadPlayerIdx = -1

	for _, p := range b.players {
		p.ResetRound()
	}

	b.dealRound()
	b.startBidPhase()
}

// dealRound カードを配る (13枚ずつ)
func (b *Bridge) dealRound() {
	b.trumpCards.Shuffle()
	for range BridgeHandSize {
		for j := range BridgePlayerCnt {
			card := b.trumpCards.DrawCard()
			if card != nil {
				b.players[j].AddCard(card)
			}
		}
	}
	b.sortAllHands()
}

// startBidPhase ビッドフェーズを開始する
func (b *Bridge) startBidPhase() {
	b.phase = BridgePhaseBid
	b.bidPlayerIdx = (b.dealerIdx + 1) % BridgePlayerCnt
	b.bidHistory = nil
	b.passCount = 0
	b.contractLevel = 0
	b.contractSuit = 0
	b.doubled = 0
	b.declarerIdx = -1
	b.dummyIdx = -1
	b.lastBidderIdx = -1
	b.lastBidTeam = -1
	b.openingLeadDone = false
	b.trumpSuit = -1
}

// --- Bid Phase ---

// PlayerBid 人間プレイヤーがビッドする
func (b *Bridge) PlayerBid(bidType int, level int, suit int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BridgePhaseBid {
		return ErrWrongPhase
	}
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	return b.executeBid(humanIdx, BridgeBidType(bidType), level, suit)
}

// CpuBid CPUプレイヤーがビッドする
func (b *Bridge) CpuBid() {
	if b.gameEndFlag || b.phase != BridgePhaseBid {
		return
	}
	if b.players[b.bidPlayerIdx].GetIsHuman() {
		return
	}

	bidType, level, suit := b.cpuSelectBid(b.bidPlayerIdx)
	_ = b.executeBid(b.bidPlayerIdx, bidType, level, suit)
}

// executeBid ビッドを実行する
func (b *Bridge) executeBid(playerIdx int, bidType BridgeBidType, level int, suit int) error {
	switch bidType {
	case BridgeBidPass:
		return b.doBidPass(playerIdx)
	case BridgeBidNormal:
		return b.doBidNormal(playerIdx, level, suit)
	case BridgeBidDouble:
		return b.doBidDouble(playerIdx)
	case BridgeBidRedouble:
		return b.doBidRedouble(playerIdx)
	default:
		return NewDomainError(ErrInvalidPlay, "無効なビッド種類です")
	}
}

// doBidPass パスする
func (b *Bridge) doBidPass(playerIdx int) error {
	b.bidHistory = append(b.bidHistory, &BridgeBidEntry{
		PlayerIdx: playerIdx,
		BidType:   BridgeBidPass,
	})
	b.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes", b.playerName(playerIdx)), nil)
	b.passCount++

	// 4連続パス (誰もビッドしていない) = リディール
	if b.passCount >= 4 && b.contractLevel == 0 {
		b.appendLog(-1, "redeal", "All pass — redeal", nil)
		b.dealerIdx = (b.dealerIdx + 1) % BridgePlayerCnt
		for _, p := range b.players {
			p.ResetRound()
		}
		b.dealRound()
		b.startBidPhase()
		return nil
	}

	// 最後のビッドの後に3連続パス = オークション終了
	if b.passCount >= 3 && b.contractLevel > 0 {
		b.finishAuction()
		return nil
	}

	b.advanceBidPlayer()
	return nil
}

// doBidNormal 通常ビッドする
func (b *Bridge) doBidNormal(playerIdx int, level int, suit int) error {
	if level < 1 || level > 7 {
		return NewDomainError(ErrInvalidPlay, "ビッドレベルは1-7です")
	}
	if suit < BridgeBidSuitClub || suit > BridgeBidSuitNT {
		return NewDomainError(ErrInvalidPlay, "無効なビッドスートです")
	}

	// 現在のコントラクトより高いビッドのみ有効
	if !b.isHigherBid(level, suit) {
		return NewDomainError(ErrInvalidPlay, "現在のコントラクトより高いビッドが必要です")
	}

	b.bidHistory = append(b.bidHistory, &BridgeBidEntry{
		PlayerIdx: playerIdx,
		BidType:   BridgeBidNormal,
		Level:     level,
		Suit:      suit,
	})

	b.contractLevel = level
	b.contractSuit = suit
	b.doubled = 0
	b.lastBidderIdx = playerIdx
	b.lastBidTeam = b.players[playerIdx].GetTeam()
	b.passCount = 0

	suitName := b.bidSuitName(suit)
	b.appendLog(playerIdx, "bid",
		fmt.Sprintf("%s bids %d%s", b.playerName(playerIdx), level, suitName), nil)

	b.advanceBidPlayer()
	return nil
}

// doBidDouble ダブルする
func (b *Bridge) doBidDouble(playerIdx int) error {
	if b.contractLevel == 0 {
		return NewDomainError(ErrInvalidPlay, "ビッドがない状態ではダブルできません")
	}
	if b.doubled != 0 {
		return NewDomainError(ErrInvalidPlay, "既にダブル/リダブルされています")
	}
	// 相手チームのビッドのみダブル可能
	if b.players[playerIdx].GetTeam() == b.lastBidTeam {
		return NewDomainError(ErrInvalidPlay, "自分のチームのビッドはダブルできません")
	}

	b.bidHistory = append(b.bidHistory, &BridgeBidEntry{
		PlayerIdx: playerIdx,
		BidType:   BridgeBidDouble,
	})
	b.doubled = 1
	b.passCount = 0

	b.appendLog(playerIdx, "double",
		fmt.Sprintf("%s doubles", b.playerName(playerIdx)), nil)

	b.advanceBidPlayer()
	return nil
}

// doBidRedouble リダブルする
func (b *Bridge) doBidRedouble(playerIdx int) error {
	if b.doubled != 1 {
		return NewDomainError(ErrInvalidPlay, "ダブルされていない状態ではリダブルできません")
	}
	// ダブルされた側のチームのみリダブル可能
	if b.players[playerIdx].GetTeam() != b.lastBidTeam {
		return NewDomainError(ErrInvalidPlay, "相手チームのダブルにのみリダブルできます")
	}

	b.bidHistory = append(b.bidHistory, &BridgeBidEntry{
		PlayerIdx: playerIdx,
		BidType:   BridgeBidRedouble,
	})
	b.doubled = 2
	b.passCount = 0

	b.appendLog(playerIdx, "redouble",
		fmt.Sprintf("%s redoubles", b.playerName(playerIdx)), nil)

	b.advanceBidPlayer()
	return nil
}

// isHigherBid 指定されたビッドが現在のコントラクトより高いか
func (b *Bridge) isHigherBid(level int, suit int) bool {
	if b.contractLevel == 0 {
		return true
	}
	if level > b.contractLevel {
		return true
	}
	if level == b.contractLevel && suit > b.contractSuit {
		return true
	}
	return false
}

// advanceBidPlayer 次のビッドプレイヤーへ
func (b *Bridge) advanceBidPlayer() {
	b.bidPlayerIdx = (b.bidPlayerIdx + 1) % BridgePlayerCnt
}

// finishAuction オークション終了処理
func (b *Bridge) finishAuction() {
	// デクレアラー = 勝利チームで最初にそのスートをビッドしたプレイヤー
	b.declarerIdx = b.findFirstBidder(b.lastBidTeam, b.contractSuit)
	b.dummyIdx = (b.declarerIdx + 2) % BridgePlayerCnt // パートナー

	// 切り札スートをカードデザインに変換
	b.trumpSuit = b.bidSuitToCardDesign(b.contractSuit)

	suitName := b.bidSuitName(b.contractSuit)
	b.appendLog(-1, "contract",
		fmt.Sprintf("Contract: %d%s by %s", b.contractLevel, suitName, b.playerName(b.declarerIdx)), nil)
	switch b.doubled {
	case 1:
		b.appendLog(-1, "doubled", "Doubled", nil)
	case 2:
		b.appendLog(-1, "redoubled", "Redoubled", nil)
	}

	// オープニングリード: デクレアラーの左隣
	b.leadPlayerIdx = (b.declarerIdx + 1) % BridgePlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber = 1
	b.currentTrick = nil
	b.openingLeadDone = false
	b.phase = BridgePhasePlay
}

// findFirstBidder チームで最初に指定スートをビッドしたプレイヤーを返す
func (b *Bridge) findFirstBidder(team int, suit int) int {
	for _, entry := range b.bidHistory {
		if entry.BidType == BridgeBidNormal && entry.Suit == suit &&
			b.players[entry.PlayerIdx].GetTeam() == team {
			return entry.PlayerIdx
		}
	}
	return b.lastBidderIdx
}

// --- Play Phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Bridge) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BridgePhasePlay {
		return ErrWrongPhase
	}

	// デクレアラーはダミーのカードもプレイできる
	actingPlayer := b.currentPlayerIdx
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}

	// 人間がデクレアラーの場合、ダミーの手番でもプレイ可能
	if actingPlayer == b.dummyIdx && humanIdx == b.declarerIdx {
		// OK: デクレアラーがダミーのカードをプレイ
	} else if actingPlayer != humanIdx {
		return ErrNotHumanTurn
	}

	player := b.players[actingPlayer]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := b.validatePlay(actingPlayer, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	b.playCard(actingPlayer, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行
func (b *Bridge) CpuPlay() {
	if b.gameEndFlag || b.phase != BridgePhasePlay {
		return
	}

	actingPlayer := b.currentPlayerIdx

	// ダミーの手番: デクレアラーがCPUの場合CPUがプレイ
	if actingPlayer == b.dummyIdx {
		if b.players[b.declarerIdx].GetIsHuman() {
			return // 人間デクレアラーがダミーを操作する
		}
		// CPUデクレアラーがダミーのカードを選ぶ
		cardIdx := b.cpuSelectPlayCard(actingPlayer)
		played := b.players[actingPlayer].RemoveCard(cardIdx)
		b.playCard(actingPlayer, played)
		return
	}

	if b.players[actingPlayer].GetIsHuman() {
		return
	}

	cardIdx := b.cpuSelectPlayCard(actingPlayer)
	played := b.players[actingPlayer].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	b.playCard(actingPlayer, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (b *Bridge) ResolveTrick() {
	if b.phase != BridgePhaseTrickEnd || len(b.currentTrick) != BridgePlayerCnt {
		return
	}

	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
	}

	b.players[winnerIdx].AddTrick(trickCards)

	winnerName := b.playerName(winnerIdx)
	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", winnerName, b.trickNumber), trickCards)

	b.leadPlayerIdx = winnerIdx

	if b.trickNumber >= BridgeTotalTricks {
		b.phase = BridgePhaseRoundEnd
	} else {
		b.phase = BridgePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (b *Bridge) NextTrick() {
	if b.phase != BridgePhaseTrickEnd {
		return
	}
	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = BridgePhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (b *Bridge) ScoreRound() {
	if b.phase != BridgePhaseRoundEnd {
		return
	}

	// デクレアラーチームのトリック数を集計
	declarerTeam := b.players[b.declarerIdx].GetTeam()
	defenderTeam := 1 - declarerTeam
	teamTricks := [BridgeTeamCnt]int{}
	for _, p := range b.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}

	declarerTricks := teamTricks[declarerTeam]
	requiredTricks := b.contractLevel + 6 // レベル + ブック (6)

	b.appendLog(-1, "tricks",
		fmt.Sprintf("Declarer team tricks: %d (needed: %d)", declarerTricks, requiredTricks), nil)

	if declarerTricks >= requiredTricks {
		// コントラクト達成
		overtricks := declarerTricks - requiredTricks
		points := b.calcMadeContractScore(declarerTeam, overtricks)
		b.teamScores[declarerTeam] += points
		b.appendLog(-1, "contract_made",
			fmt.Sprintf("Contract made! +%d points for team %d", points, declarerTeam), nil)
	} else {
		// コントラクト失敗
		undertricks := requiredTricks - declarerTricks
		penalty := b.calcUndertrickPenalty(declarerTeam, undertricks)
		b.teamScores[defenderTeam] += penalty
		b.appendLog(-1, "contract_down",
			fmt.Sprintf("Contract down %d! +%d points for team %d", undertricks, penalty, defenderTeam), nil)
	}

	// スコアログ
	for ti := range BridgeTeamCnt {
		b.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points (tricks: %d)", ti, b.teamScores[ti], teamTricks[ti]), nil)
	}

	b.checkGameEnd()
}

// calcMadeContractScore コントラクト達成時のスコアを計算する
func (b *Bridge) calcMadeContractScore(declarerTeam int, overtricks int) int {
	score := 0
	isVul := b.vulnerability[declarerTeam]

	// コントラクトポイント (ライン以下)
	contractPoints := b.calcContractPoints()

	// ダブル/リダブル倍率
	switch b.doubled {
	case 1:
		contractPoints *= 2
	case 2:
		contractPoints *= 4
	}

	b.belowLine[declarerTeam] += contractPoints
	score += contractPoints

	// ゲーム達成判定 (100点以上でゲーム)
	if b.belowLine[declarerTeam] >= 100 {
		b.gamesWon[declarerTeam]++
		// ゲームボーナス
		if isVul {
			score += 500
		} else {
			score += 300
		}
		// 両チームのベロウラインをリセット
		b.belowLine = [BridgeTeamCnt]int{}
		// バルネラビリティ更新
		b.vulnerability[declarerTeam] = true
	}

	// オーバートリックボーナス
	if overtricks > 0 {
		score += b.calcOvertrickBonus(declarerTeam, overtricks)
	}

	// スラムボーナス
	switch b.contractLevel {
	case 6:
		if isVul {
			score += 750
		} else {
			score += 500
		}
	case 7:
		if isVul {
			score += 1500
		} else {
			score += 1000
		}
	}

	// ダブル/リダブルで成功した場合の追加ボーナス
	switch b.doubled {
	case 1:
		score += 50
	case 2:
		score += 100
	}

	return score
}

// calcContractPoints コントラクトポイント（ライン以下）を計算する
func (b *Bridge) calcContractPoints() int {
	switch b.contractSuit {
	case BridgeBidSuitClub, BridgeBidSuitDiamond:
		return b.contractLevel * 20
	case BridgeBidSuitHeart, BridgeBidSuitSpade:
		return b.contractLevel * 30
	case BridgeBidSuitNT:
		return 40 + (b.contractLevel-1)*30
	}
	return 0
}

// calcOvertrickBonus オーバートリックボーナスを計算する
func (b *Bridge) calcOvertrickBonus(declarerTeam int, overtricks int) int {
	isVul := b.vulnerability[declarerTeam]
	switch b.doubled {
	case 0:
		// アンダブル: トリック単価
		switch b.contractSuit {
		case BridgeBidSuitClub, BridgeBidSuitDiamond:
			return overtricks * 20
		default:
			return overtricks * 30
		}
	case 1:
		// ダブル
		if isVul {
			return overtricks * 200
		}
		return overtricks * 100
	case 2:
		// リダブル
		if isVul {
			return overtricks * 400
		}
		return overtricks * 200
	}
	return 0
}

// calcUndertrickPenalty アンダートリックペナルティを計算する
func (b *Bridge) calcUndertrickPenalty(declarerTeam int, undertricks int) int {
	isVul := b.vulnerability[declarerTeam]
	penalty := 0

	switch b.doubled {
	case 0:
		if isVul {
			penalty = undertricks * 100
		} else {
			penalty = undertricks * 50
		}
	case 1:
		for i := 1; i <= undertricks; i++ {
			if isVul {
				if i == 1 {
					penalty += 200
				} else {
					penalty += 300
				}
			} else {
				if i == 1 {
					penalty += 100
				} else if i <= 3 {
					penalty += 200
				} else {
					penalty += 300
				}
			}
		}
	case 2:
		for i := 1; i <= undertricks; i++ {
			if isVul {
				if i == 1 {
					penalty += 400
				} else {
					penalty += 600
				}
			} else {
				if i == 1 {
					penalty += 200
				} else if i <= 3 {
					penalty += 400
				} else {
					penalty += 600
				}
			}
		}
	}

	return penalty
}

// checkGameEnd ゲーム終了判定 (ラバー: 先に2ゲーム勝利)
func (b *Bridge) checkGameEnd() {
	for ti := range BridgeTeamCnt {
		if b.gamesWon[ti] >= 2 {
			b.gameEndFlag = true
			b.phase = BridgePhaseGameEnd

			// ラバーボーナス
			opponentGames := b.gamesWon[1-ti]
			if opponentGames == 0 {
				b.teamScores[ti] += 700
			} else {
				b.teamScores[ti] += 500
			}

			b.winnerTeam = ti
			b.appendLog(-1, "rubber_end",
				fmt.Sprintf("Team %d wins the rubber!", ti), nil)
			return
		}
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (b *Bridge) GetPhase() BridgePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Bridge) SetPhase(phase BridgePhase) { b.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (b *Bridge) GetRoundNumber() int { return b.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (b *Bridge) SetRoundNumber(n int) { b.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (b *Bridge) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Bridge) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Bridge) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Bridge) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Bridge) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Bridge) SetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Bridge) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (b *Bridge) GetWinnerTeam() int { return b.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (b *Bridge) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Bridge) GetPlayer(i int) *BridgePlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Bridge) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Bridge) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (b *Bridge) GetBidPlayerIdx() int { return b.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (b *Bridge) SetBidPlayerIdx(idx int) { b.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Bridge) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Bridge) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得 (-1 = NoTrump)
func (b *Bridge) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (b *Bridge) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetContractLevel コントラクトレベル取得
func (b *Bridge) GetContractLevel() int { return b.contractLevel }

// SetContractLevel コントラクトレベル設定 (テスト用)
func (b *Bridge) SetContractLevel(level int) { b.contractLevel = level }

// GetContractSuit コントラクトスート取得
func (b *Bridge) GetContractSuit() int { return b.contractSuit }

// SetContractSuit コントラクトスート設定 (テスト用)
func (b *Bridge) SetContractSuit(suit int) { b.contractSuit = suit }

// GetDoubled ダブル状態取得 (0=なし, 1=ダブル, 2=リダブル)
func (b *Bridge) GetDoubled() int { return b.doubled }

// SetDoubled ダブル状態設定 (テスト用)
func (b *Bridge) SetDoubled(d int) { b.doubled = d }

// GetDeclarerIdx デクレアラーインデックス取得
func (b *Bridge) GetDeclarerIdx() int { return b.declarerIdx }

// SetDeclarerIdx デクレアラーインデックス設定 (テスト用)
func (b *Bridge) SetDeclarerIdx(idx int) { b.declarerIdx = idx }

// GetDummyIdx ダミーインデックス取得
func (b *Bridge) GetDummyIdx() int { return b.dummyIdx }

// SetDummyIdx ダミーインデックス設定 (テスト用)
func (b *Bridge) SetDummyIdx(idx int) { b.dummyIdx = idx }

// GetBidHistory ビッド履歴取得
func (b *Bridge) GetBidHistory() []*BridgeBidEntry { return b.bidHistory }

// GetVulnerability バルネラビリティ取得
func (b *Bridge) GetVulnerability(team int) bool {
	if team < 0 || team >= BridgeTeamCnt {
		return false
	}
	return b.vulnerability[team]
}

// SetVulnerability バルネラビリティ設定 (テスト用)
func (b *Bridge) SetVulnerability(team int, vul bool) {
	if team >= 0 && team < BridgeTeamCnt {
		b.vulnerability[team] = vul
	}
}

// GetTeamScore チームスコア取得
func (b *Bridge) GetTeamScore(team int) int {
	if team < 0 || team >= BridgeTeamCnt {
		return 0
	}
	return b.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (b *Bridge) SetTeamScore(team, score int) {
	if team >= 0 && team < BridgeTeamCnt {
		b.teamScores[team] = score
	}
}

// GetGamesWon 勝利ゲーム数取得
func (b *Bridge) GetGamesWon(team int) int {
	if team < 0 || team >= BridgeTeamCnt {
		return 0
	}
	return b.gamesWon[team]
}

// SetGamesWon 勝利ゲーム数設定 (テスト用)
func (b *Bridge) SetGamesWon(team, wins int) {
	if team >= 0 && team < BridgeTeamCnt {
		b.gamesWon[team] = wins
	}
}

// GetBelowLine ライン以下スコア取得
func (b *Bridge) GetBelowLine(team int) int {
	if team < 0 || team >= BridgeTeamCnt {
		return 0
	}
	return b.belowLine[team]
}

// SetBelowLine ライン以下スコア設定 (テスト用)
func (b *Bridge) SetBelowLine(team, score int) {
	if team >= 0 && team < BridgeTeamCnt {
		b.belowLine[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Bridge) IsHumanTurn() bool {
	if b.phase != BridgePhasePlay {
		return false
	}
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 {
		return false
	}
	// ダミーの手番かつ人間がデクレアラーの場合
	if b.currentPlayerIdx == b.dummyIdx && humanIdx == b.declarerIdx {
		return true
	}
	return b.currentPlayerIdx == humanIdx
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (b *Bridge) IsHumanBidTurn() bool {
	if b.bidPlayerIdx < 0 || b.bidPlayerIdx >= len(b.players) {
		return false
	}
	return b.players[b.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (b *Bridge) GetConfig() BridgeConfig { return b.config }

// SetConfig 設定変更
func (b *Bridge) SetConfig(cfg BridgeConfig) { b.config = cfg }

// GetActionLog 棋譜取得
func (b *Bridge) GetActionLog() []*ActionLogEntry { return b.actionLog }

// IsOpeningLeadDone オープニングリード完了か
func (b *Bridge) IsOpeningLeadDone() bool { return b.openingLeadDone }

// GetDummyHand ダミーの手札を取得 (オープニングリード後のみ公開)
func (b *Bridge) GetDummyHand() []*Card {
	if !b.openingLeadDone || b.dummyIdx < 0 {
		return nil
	}
	dummy := b.players[b.dummyIdx]
	cards := make([]*Card, dummy.GetCardsSize())
	for i := 0; i < dummy.GetCardsSize(); i++ {
		cards[i] = dummy.GetCard(i)
	}
	return cards
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (b *Bridge) GetValidPlayIndices(playerIdx int) []int {
	return b.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (b *Bridge) GetHint() *BridgeHint {
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}

	switch b.phase {
	case BridgePhaseBid:
		if b.bidPlayerIdx != humanIdx {
			return nil
		}
		bidType, level, suit := b.cpuSelectBid(humanIdx)
		bt := int(bidType)
		return &BridgeHint{BidType: &bt, BidLevel: &level, BidSuit: &suit, Reason: "strategic_bid"}

	case BridgePhasePlay:
		playIdx := b.currentPlayerIdx
		// ダミーの手番かつ人間デクレアラーの場合
		if playIdx == b.dummyIdx && humanIdx == b.declarerIdx {
			playIdx = b.dummyIdx
		} else if playIdx != humanIdx {
			return nil
		}
		validIndices := b.getValidPlayIndices(playIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := b.cpuPlayHard(playIdx, validIndices)
		return &BridgeHint{CardIndex: &idx, Reason: b.playHintReason(playIdx, idx)}
	}
	return nil
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (b *Bridge) findHumanIdx() int {
	for i, p := range b.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// playCard カードをプレイする共通処理
func (b *Bridge) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	b.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", b.playerName(playerIdx), cardStr(card)), []*Card{card})

	// オープニングリード後にダミーの手札を公開
	if !b.openingLeadDone && len(b.currentTrick) == 1 {
		b.openingLeadDone = true
		b.appendLog(-1, "dummy_revealed",
			fmt.Sprintf("Dummy's hand (player %d) is revealed", b.dummyIdx), nil)
	}

	if len(b.currentTrick) == BridgePlayerCnt {
		b.phase = BridgePhaseTrickEnd
	} else {
		b.currentPlayerIdx = (b.currentPlayerIdx + 1) % BridgePlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (b *Bridge) validatePlay(playerIdx int, card *Card) error {
	if len(b.currentTrick) == 0 {
		return nil // リードは自由
	}

	// フォロースート
	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if b.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか
func (b *Bridge) playerHasSuit(playerIdx int, suit int) bool {
	p := b.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (b *Bridge) getValidPlayIndices(playerIdx int) []int {
	player := b.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return b.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// trickWinner トリックの勝者を決定する
func (b *Bridge) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}

	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerRank := b.cardRank(b.currentTrick[0].Card)
	winnerSuit := b.currentTrick[0].Card.GetDesign()
	leadSuit := winnerSuit

	for _, tc := range b.currentTrick[1:] {
		rank := b.cardRank(tc.Card)
		suit := tc.Card.GetDesign()

		if b.trumpSuit >= 0 {
			// 切り札ありの場合
			if suit == b.trumpSuit && winnerSuit != b.trumpSuit {
				winnerIdx = tc.PlayerIdx
				winnerRank = rank
				winnerSuit = suit
			} else if suit == winnerSuit && rank > winnerRank {
				winnerIdx = tc.PlayerIdx
				winnerRank = rank
				winnerSuit = suit
			}
		} else {
			// ノートランプの場合: リードスートの最高ランクが勝つ
			if suit == leadSuit && rank > winnerRank {
				winnerIdx = tc.PlayerIdx
				winnerRank = rank
				winnerSuit = suit
			}
		}
	}
	return winnerIdx
}

// cardRank トリック比較用のカードランク (高い = 強い)
// A(14) > K(13) > Q(12) > J(11) > 10 > ... > 2
func (b *Bridge) cardRank(card *Card) int {
	v := card.GetValue()
	if v == 1 {
		return 14 // Ace is highest
	}
	return v
}

// sortAllHands 全プレイヤーの手札をソートする
func (b *Bridge) sortAllHands() {
	for _, p := range b.players {
		bridgeSortHand(p)
	}
}

// bridgeSortHand プレイヤーの手札をスート→ランクの順にソートする
func bridgeSortHand(p *BridgePlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si := ci.GetDesign()
		sj := cj.GetDesign()
		if si != sj {
			return si < sj
		}
		vi := ci.GetValue()
		vj := cj.GetValue()
		// Aceを最強にするため14に変換
		if vi == 1 {
			vi = 14
		}
		if vj == 1 {
			vj = 14
		}
		return vi < vj
	})
}

// bidSuitToCardDesign ビッドスートをカードデザインに変換する
func (b *Bridge) bidSuitToCardDesign(suit int) int {
	switch suit {
	case BridgeBidSuitClub:
		return CardDesignClover
	case BridgeBidSuitDiamond:
		return CardDesignDiamond
	case BridgeBidSuitHeart:
		return CardDesignHeart
	case BridgeBidSuitSpade:
		return CardDesignSpade
	case BridgeBidSuitNT:
		return -1 // NoTrump
	}
	return -1
}

// bidSuitName ビッドスート名を返す
func (b *Bridge) bidSuitName(suit int) string {
	switch suit {
	case BridgeBidSuitClub:
		return "C"
	case BridgeBidSuitDiamond:
		return "D"
	case BridgeBidSuitHeart:
		return "H"
	case BridgeBidSuitSpade:
		return "S"
	case BridgeBidSuitNT:
		return "NT"
	}
	return "?"
}

// playerName プレイヤー名を返す
func (b *Bridge) playerName(idx int) string {
	if idx < 0 || idx >= len(b.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	positions := []string{"North", "East", "South", "West"}
	if b.players[idx].GetIsHuman() {
		return fmt.Sprintf("You (%s)", positions[idx])
	}
	return fmt.Sprintf("CPU %d (%s)", idx, positions[idx])
}

// appendLog 棋譜にエントリを追加する
func (b *Bridge) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// playHintReason プレイヒントの理由を判定する
func (b *Bridge) playHintReason(playerIdx int, chosenIdx int) string {
	player := b.players[playerIdx]
	card := player.GetCard(chosenIdx)

	if len(b.currentTrick) == 0 {
		if b.trumpSuit >= 0 && card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}

	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if b.trumpSuit >= 0 && card.GetDesign() == b.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (b *Bridge) cpuSelectBid(playerIdx int) (BridgeBidType, int, int) {
	switch b.config.CpuDifficulty {
	case BridgeCpuDifficultyHard:
		return b.cpuBidHard(playerIdx)
	case BridgeCpuDifficultyNormal:
		return b.cpuBidNormal(playerIdx)
	default:
		return b.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムにビッド (50%パス、50%最低ビッド)
func (b *Bridge) cpuBidEasy(playerIdx int) (BridgeBidType, int, int) {
	if rand.Intn(2) == 0 || b.contractLevel >= 4 {
		return BridgeBidPass, 0, 0
	}

	// 最低有効ビッドを探す
	for level := 1; level <= 7; level++ {
		for suit := BridgeBidSuitClub; suit <= BridgeBidSuitNT; suit++ {
			if b.isHigherBid(level, suit) {
				return BridgeBidNormal, level, suit
			}
		}
	}
	return BridgeBidPass, 0, 0
}

// cpuBidNormal ポイントカウント制に基づくビッド
func (b *Bridge) cpuBidNormal(playerIdx int) (BridgeBidType, int, int) {
	hcp := b.calcHCP(playerIdx)
	player := b.players[playerIdx]

	// ダブル判定: 相手チームのビッドで自分がHCP15+ならダブル
	if b.contractLevel > 0 && b.doubled == 0 &&
		b.players[playerIdx].GetTeam() != b.lastBidTeam && hcp >= 15 {
		return BridgeBidDouble, 0, 0
	}

	// リダブル判定: 味方がビッドしてダブルされた場合、HCP10+でリダブル
	if b.doubled == 1 && b.players[playerIdx].GetTeam() == b.lastBidTeam && hcp >= 10 {
		return BridgeBidRedouble, 0, 0
	}

	// HCP12未満ならパス
	if hcp < 12 {
		return BridgeBidPass, 0, 0
	}

	// 最長スートを見つける
	bestSuit, bestLen := b.findLongestSuit(player)

	// ビッドレベル判定
	bidLevel := 1
	if hcp >= 20 {
		bidLevel = 2
	}
	if hcp >= 25 && bestLen >= 6 {
		bidLevel = 3
	}

	// NT判定: バランスハンド (各スート2枚以上)
	bidSuit := b.cardDesignToBidSuit(bestSuit)
	if b.isBalancedHand(player) && hcp >= 15 {
		bidSuit = BridgeBidSuitNT
	}

	// 有効なビッドか確認
	if b.isHigherBid(bidLevel, bidSuit) {
		return BridgeBidNormal, bidLevel, bidSuit
	}

	// レベルを上げてリトライ
	for level := bidLevel; level <= 7; level++ {
		for suit := BridgeBidSuitClub; suit <= BridgeBidSuitNT; suit++ {
			if b.isHigherBid(level, suit) {
				// レベルが高すぎる場合はパス
				if level > 3 && hcp < 20 {
					return BridgeBidPass, 0, 0
				}
				return BridgeBidNormal, level, suit
			}
		}
	}

	return BridgeBidPass, 0, 0
}

// cpuBidHard 高度なビッド戦略 (HCP+分布点、パートナー連携)
func (b *Bridge) cpuBidHard(playerIdx int) (BridgeBidType, int, int) {
	hcp := b.calcHCP(playerIdx)
	distPts := b.calcDistributionPoints(playerIdx)
	totalPts := hcp + distPts
	player := b.players[playerIdx]

	// パートナーフィットボーナス
	partnerSuit := b.partnerBidSuit(playerIdx)
	if partnerSuit > 0 {
		partnerCardDesign := b.bidSuitToCardDesign(partnerSuit)
		if b.countSuitCards(playerIdx, partnerCardDesign) >= 3 {
			totalPts += 3
		}
	}

	// ダブル判定: 相手チームのビッド、totalPts>=14、コントラクトスートに2枚以上
	// NT契約の場合はbidSuitToCardDesignが-1を返すためcountSuitCardsは常に0となりダブルしない
	if b.contractLevel > 0 && b.doubled == 0 &&
		b.players[playerIdx].GetTeam() != b.lastBidTeam && totalPts >= 14 {
		contractCardDesign := b.bidSuitToCardDesign(b.contractSuit)
		if b.countSuitCards(playerIdx, contractCardDesign) >= 2 {
			return BridgeBidDouble, 0, 0
		}
	}

	// リダブル判定: 味方ビッドがダブルされた場合、totalPts>=12
	if b.doubled == 1 && b.players[playerIdx].GetTeam() == b.lastBidTeam && totalPts >= 12 {
		return BridgeBidRedouble, 0, 0
	}

	// totalPts < 12 ならパス
	if totalPts < 12 {
		return BridgeBidPass, 0, 0
	}

	// パートナーのスートを支持 (3枚以上のフィット)
	if partnerSuit > 0 {
		partnerCardDesign := b.bidSuitToCardDesign(partnerSuit)
		if b.countSuitCards(playerIdx, partnerCardDesign) >= 3 {
			// パートナーのスートでレイズ
			bidLevel := b.contractLevel + 1
			if totalPts >= 13 {
				// ゲームレベルへジャンプ
				if partnerSuit == BridgeBidSuitHeart || partnerSuit == BridgeBidSuitSpade {
					bidLevel = 4
				} else {
					bidLevel = 5
				}
			}
			if bidLevel > 7 {
				bidLevel = 7
			}
			if b.isHigherBid(bidLevel, partnerSuit) {
				return BridgeBidNormal, bidLevel, partnerSuit
			}
		}
	}

	// 最長スートを見つける
	bestSuit, bestLen := b.findLongestSuit(player)
	bidSuit := b.cardDesignToBidSuit(bestSuit)

	// メジャースート優先 (同じ長さならハート/スペードを優先)
	if bestLen >= 4 {
		suitCounts := [5]int{}
		for i := 0; i < player.GetCardsSize(); i++ {
			d := player.GetCard(i).GetDesign()
			if d >= CardDesignSpade && d <= CardDesignDiamond {
				suitCounts[d]++
			}
		}
		// メジャースートで同じ長さがあれば優先
		for _, major := range []int{CardDesignSpade, CardDesignHeart} {
			if suitCounts[major] >= bestLen {
				bidSuit = b.cardDesignToBidSuit(major)
				break
			}
		}
	}

	// NT判定: バランスハンド
	if b.isBalancedHand(player) {
		if totalPts >= 20 {
			bidSuit = BridgeBidSuitNT
			if b.isHigherBid(2, bidSuit) {
				return BridgeBidNormal, 2, bidSuit
			}
		} else if totalPts >= 15 {
			bidSuit = BridgeBidSuitNT
			if b.isHigherBid(1, bidSuit) {
				return BridgeBidNormal, 1, bidSuit
			}
		}
	}

	// ビッドレベル判定
	bidLevel := 1
	if totalPts >= 23 {
		// ゲームレベル
		switch bidSuit {
		case BridgeBidSuitHeart, BridgeBidSuitSpade:
			bidLevel = 4
		case BridgeBidSuitNT:
			bidLevel = 3
		default:
			bidLevel = 5
		}
	} else if totalPts >= 20 {
		bidLevel = 2
	}

	// 有効なビッドか確認
	if b.isHigherBid(bidLevel, bidSuit) {
		return BridgeBidNormal, bidLevel, bidSuit
	}

	// レベルを上げてリトライ (まず最強スートで試行)
	for level := bidLevel; level <= 7; level++ {
		if level > 3 && totalPts < 18 {
			return BridgeBidPass, 0, 0
		}
		if level > 5 && totalPts < 25 {
			return BridgeBidPass, 0, 0
		}
		if b.isHigherBid(level, bidSuit) {
			return BridgeBidNormal, level, bidSuit
		}
		for suit := BridgeBidSuitClub; suit <= BridgeBidSuitNT; suit++ {
			if suit == bidSuit {
				continue
			}
			if b.isHigherBid(level, suit) {
				return BridgeBidNormal, level, suit
			}
		}
	}

	return BridgeBidPass, 0, 0
}

// calcHCP ハイカードポイントを計算する (A=4, K=3, Q=2, J=1)
func (b *Bridge) calcHCP(playerIdx int) int {
	player := b.players[playerIdx]
	hcp := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		switch card.GetValue() {
		case 1: // Ace
			hcp += 4
		case 13: // King
			hcp += 3
		case 12: // Queen
			hcp += 2
		case 11: // Jack
			hcp++
		}
	}
	return hcp
}

// findLongestSuit 最長スートとその枚数を返す
func (b *Bridge) findLongestSuit(player *BridgePlayer) (int, int) {
	suitCounts := [5]int{} // index 0=unused, 1-4=CardDesign
	for i := 0; i < player.GetCardsSize(); i++ {
		d := player.GetCard(i).GetDesign()
		if d >= CardDesignSpade && d <= CardDesignDiamond {
			suitCounts[d]++
		}
	}

	bestSuit := CardDesignSpade
	bestLen := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCounts[suit] > bestLen {
			bestLen = suitCounts[suit]
			bestSuit = suit
		}
	}
	return bestSuit, bestLen
}

// isBalancedHand バランスハンドかどうか (各スート2枚以上)
func (b *Bridge) isBalancedHand(player *BridgePlayer) bool {
	suitCounts := [5]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		d := player.GetCard(i).GetDesign()
		if d >= CardDesignSpade && d <= CardDesignDiamond {
			suitCounts[d]++
		}
	}
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCounts[suit] < 2 {
			return false
		}
	}
	return true
}

// calcDistributionPoints 分布点を計算する (ボイド=3, シングルトン=2, ダブルトン=1)
func (b *Bridge) calcDistributionPoints(playerIdx int) int {
	player := b.players[playerIdx]
	suitCounts := [5]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		d := player.GetCard(i).GetDesign()
		if d >= CardDesignSpade && d <= CardDesignDiamond {
			suitCounts[d]++
		}
	}
	pts := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		switch suitCounts[suit] {
		case 0:
			pts += 3
		case 1:
			pts += 2
		case 2:
			pts++
		}
	}
	return pts
}

// countSuitCards 指定スートのカード枚数を返す
func (b *Bridge) countSuitCards(playerIdx int, suit int) int {
	player := b.players[playerIdx]
	count := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			count++
		}
	}
	return count
}

// partnerBidSuit パートナーの最新ビッドスートを返す (0=なし)
func (b *Bridge) partnerBidSuit(playerIdx int) int {
	partnerIdx := (playerIdx + 2) % BridgePlayerCnt
	lastSuit := 0
	for _, entry := range b.bidHistory {
		if entry.PlayerIdx == partnerIdx && entry.BidType == BridgeBidNormal {
			lastSuit = entry.Suit
		}
	}
	return lastSuit
}

// countTrumpsRemaining プレイヤーの手札にある切り札の枚数を返す
func (b *Bridge) countTrumpsRemaining(playerIdx int) int {
	if b.trumpSuit < 0 {
		return 0
	}
	return b.countSuitCards(playerIdx, b.trumpSuit)
}

// findShortestSuitCard ボイド作りのため最も短いスートの最弱カードを返す
func (b *Bridge) findShortestSuitCard(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]
	// 各スートの枚数を数える (validIndicesの中だけ、固定順序配列)
	suitCounts := [5]int{} // index 0=unused, 1-4=CardDesign
	for _, idx := range validIndices {
		d := player.GetCard(idx).GetDesign()
		if d >= CardDesignSpade && d <= CardDesignDiamond {
			suitCounts[d]++
		}
	}

	// 最短スートを見つける (固定順序で決定的)
	shortestLen := 14
	shortestSuit := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCounts[suit] > 0 && suitCounts[suit] < shortestLen {
			shortestLen = suitCounts[suit]
			shortestSuit = suit
		}
	}

	// 最短スートの中で最弱カードを選ぶ
	best := validIndices[0]
	bestRank := b.cardRank(player.GetCard(best))
	bestInShortest := false
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		rank := b.cardRank(card)
		inShortest := card.GetDesign() == shortestSuit
		if !bestInShortest && inShortest {
			best = idx
			bestRank = rank
			bestInShortest = true
		} else if inShortest == bestInShortest && rank < bestRank {
			best = idx
			bestRank = rank
		}
	}
	return best
}

// cardDesignToBidSuit カードデザインをビッドスートに変換する
func (b *Bridge) cardDesignToBidSuit(design int) int {
	switch design {
	case CardDesignClover:
		return BridgeBidSuitClub
	case CardDesignDiamond:
		return BridgeBidSuitDiamond
	case CardDesignHeart:
		return BridgeBidSuitHeart
	case CardDesignSpade:
		return BridgeBidSuitSpade
	}
	return BridgeBidSuitClub
}

// cpuSelectPlayCard CPUがプレイするカードを選択する
func (b *Bridge) cpuSelectPlayCard(playerIdx int) int {
	validIndices := b.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}

	switch b.config.CpuDifficulty {
	case BridgeCpuDifficultyHard:
		return b.cpuPlayHard(playerIdx, validIndices)
	case BridgeCpuDifficultyNormal:
		return b.cpuPlayNormal(playerIdx, validIndices)
	default:
		return validIndices[rand.Intn(len(validIndices))]
	}
}

// cpuPlayNormal 基本的なプレイ戦略
func (b *Bridge) cpuPlayNormal(playerIdx int, validIndices []int) int {
	if len(b.currentTrick) == 0 {
		// リード: 最も強いカード
		return b.selectStrongestCard(playerIdx, validIndices)
	}

	// フォロー: 勝てるなら最小の勝てるカード、勝てないなら最弱
	return b.selectSmartFollow(playerIdx, validIndices)
}

// cpuPlayHard 戦略的プレイ (切り札管理、パートナー連携、ボイド戦略)
func (b *Bridge) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]

	if len(b.currentTrick) == 0 {
		return b.cpuLeadHard(playerIdx, validIndices)
	}

	leadSuit := b.currentTrick[0].Card.GetDesign()
	currentWinner := b.currentTrickWinner()
	isPartnerWinning := b.players[currentWinner].GetTeam() == player.GetTeam()

	// パートナーが勝っている場合は弱いカード
	if isPartnerWinning {
		return b.selectWeakestCard(playerIdx, validIndices)
	}

	// リードスートをフォローできるか確認
	hasLeadSuit := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLeadSuit = true
			break
		}
	}

	if hasLeadSuit {
		return b.selectSmartFollow(playerIdx, validIndices)
	}

	// ボイド: リードスートがない
	if b.trumpSuit >= 0 {
		// 切り札で勝てるか確認
		trumpIndices := []int{}
		for _, idx := range validIndices {
			if player.GetCard(idx).GetDesign() == b.trumpSuit {
				trumpIndices = append(trumpIndices, idx)
			}
		}
		if len(trumpIndices) > 0 {
			// 既にトリックに切り札があるか確認
			highestTrumpRank := 0
			for _, tc := range b.currentTrick {
				if tc.Card.GetDesign() == b.trumpSuit {
					r := b.cardRank(tc.Card)
					if r > highestTrumpRank {
						highestTrumpRank = r
					}
				}
			}
			// 既存の切り札に勝てる最低の切り札を選ぶ
			var winningTrumps []int
			for _, idx := range trumpIndices {
				if b.cardRank(player.GetCard(idx)) > highestTrumpRank {
					winningTrumps = append(winningTrumps, idx)
				}
			}
			if len(winningTrumps) > 0 {
				return b.selectWeakestCard(playerIdx, winningTrumps)
			}
			// 切り札はあるが勝てない: 切り札は温存して他を捨てる
			nonTrumpIndices := []int{}
			for _, idx := range validIndices {
				if player.GetCard(idx).GetDesign() != b.trumpSuit {
					nonTrumpIndices = append(nonTrumpIndices, idx)
				}
			}
			if len(nonTrumpIndices) > 0 {
				return b.findShortestSuitCard(playerIdx, nonTrumpIndices)
			}
			return b.selectWeakestCard(playerIdx, validIndices)
		}
	}

	// 切り札なし / ノートランプ: 最短スートから捨てる
	return b.findShortestSuitCard(playerIdx, validIndices)
}

// cpuLeadHard Hard難易度のリード戦略
func (b *Bridge) cpuLeadHard(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]
	isDeclaringTeam := player.GetTeam() == b.players[b.declarerIdx].GetTeam()

	// 終盤: 最強カードでキャッシュ
	if b.trickNumber >= 10 {
		return b.selectStrongestCard(playerIdx, validIndices)
	}

	// ディクレアラー側: 切り札ドロー
	if isDeclaringTeam && b.trumpSuit >= 0 {
		trumpCount := b.countTrumpsRemaining(playerIdx)
		if trumpCount >= 3 && b.trickNumber <= 4 {
			// 中程度の切り札でリード
			trumpIndices := []int{}
			for _, idx := range validIndices {
				if player.GetCard(idx).GetDesign() == b.trumpSuit {
					trumpIndices = append(trumpIndices, idx)
				}
			}
			if len(trumpIndices) > 0 {
				// 中間ランクの切り札を選ぶ (ソートして中央を取る)
				sort.Slice(trumpIndices, func(i, j int) bool {
					return b.cardRank(player.GetCard(trumpIndices[i])) < b.cardRank(player.GetCard(trumpIndices[j]))
				})
				mid := len(trumpIndices) / 2
				return trumpIndices[mid]
			}
		}
	}

	// ディフェンス側: パートナーのビッドスートをリード
	if !isDeclaringTeam {
		pSuit := b.partnerBidSuit(playerIdx)
		if pSuit > 0 {
			pDesign := b.bidSuitToCardDesign(pSuit)
			var suitIndices []int
			for _, idx := range validIndices {
				if player.GetCard(idx).GetDesign() == pDesign {
					suitIndices = append(suitIndices, idx)
				}
			}
			if len(suitIndices) > 0 {
				return b.selectStrongestCard(playerIdx, suitIndices)
			}
		}
	}

	// デフォルト: 最長スートの最強カード
	longestSuit, _ := b.findLongestSuit(player)
	var longestSuitIndices []int
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == longestSuit {
			longestSuitIndices = append(longestSuitIndices, idx)
		}
	}
	if len(longestSuitIndices) > 0 {
		return b.selectStrongestCard(playerIdx, longestSuitIndices)
	}

	return b.selectStrongestCard(playerIdx, validIndices)
}

// selectStrongestCard 最も強いカードを選択する
func (b *Bridge) selectStrongestCard(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]
	best := validIndices[0]
	bestRank := b.cardRank(player.GetCard(best))
	for _, idx := range validIndices[1:] {
		rank := b.cardRank(player.GetCard(idx))
		if rank > bestRank {
			best = idx
			bestRank = rank
		}
	}
	return best
}

// selectWeakestCard 最も弱いカードを選択する
func (b *Bridge) selectWeakestCard(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]
	best := validIndices[0]
	bestRank := b.cardRank(player.GetCard(best))
	for _, idx := range validIndices[1:] {
		rank := b.cardRank(player.GetCard(idx))
		if rank < bestRank {
			best = idx
			bestRank = rank
		}
	}
	return best
}

// selectSmartFollow フォロー時のスマートな選択
func (b *Bridge) selectSmartFollow(playerIdx int, validIndices []int) int {
	player := b.players[playerIdx]
	currentWinnerRank := b.currentTrickWinnerRank()

	// 勝てるカードの中で最も弱いカード
	var winningIndices []int
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		suit := card.GetDesign()
		rank := b.cardRank(card)

		leadSuit := b.currentTrick[0].Card.GetDesign()

		if b.trumpSuit >= 0 {
			// 切り札で勝てるか
			if suit == b.trumpSuit && b.currentTrickWinnerSuit() != b.trumpSuit {
				winningIndices = append(winningIndices, idx)
				continue
			}
		}
		// 同スートで勝てるか
		if suit == leadSuit && rank > currentWinnerRank {
			winningIndices = append(winningIndices, idx)
		}
	}

	if len(winningIndices) > 0 {
		return b.selectWeakestCard(playerIdx, winningIndices)
	}

	// 勝てないなら最弱
	return b.selectWeakestCard(playerIdx, validIndices)
}

// currentTrickWinner 現在のトリックの暫定勝者を返す
func (b *Bridge) currentTrickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerRank := b.cardRank(b.currentTrick[0].Card)
	winnerSuit := b.currentTrick[0].Card.GetDesign()

	for _, tc := range b.currentTrick[1:] {
		rank := b.cardRank(tc.Card)
		suit := tc.Card.GetDesign()

		if b.trumpSuit >= 0 && suit == b.trumpSuit && winnerSuit != b.trumpSuit {
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = suit
		} else if suit == winnerSuit && rank > winnerRank {
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = suit
		}
	}
	return winnerIdx
}

// currentTrickWinnerRank 現在のトリックの暫定勝者のランクを返す
func (b *Bridge) currentTrickWinnerRank() int {
	winner := b.currentTrickWinner()
	for _, tc := range b.currentTrick {
		if tc.PlayerIdx == winner {
			return b.cardRank(tc.Card)
		}
	}
	return 0
}

// currentTrickWinnerSuit 現在のトリックの暫定勝者のスートを返す
func (b *Bridge) currentTrickWinnerSuit() int {
	winner := b.currentTrickWinner()
	for _, tc := range b.currentTrick {
		if tc.PlayerIdx == winner {
			return tc.Card.GetDesign()
		}
	}
	return 0
}

// --- JSON Serialization ---

// bridgeJSON is the JSON wire format for Bridge.
type bridgeJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*BridgePlayer     `json:"ps"`
	Config           BridgeConfig        `json:"cf"`
	Phase            BridgePhase         `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	DealerIdx        int                 `json:"di"`
	BidPlayerIdx     int                 `json:"bi"`
	BidHistory       []*BridgeBidEntry   `json:"bh"`
	PassCount        int                 `json:"pc"`
	ContractLevel    int                 `json:"cl"`
	ContractSuit     int                 `json:"cs"`
	Doubled          int                 `json:"db"`
	DeclarerIdx      int                 `json:"xi"`
	DummyIdx         int                 `json:"mi"`
	LastBidderIdx    int                 `json:"lb"`
	LastBidTeam      int                 `json:"lt"`
	OpeningLeadDone  bool                `json:"ol"`
	TrumpSuit        int                 `json:"ts"`
	Vulnerability    [BridgeTeamCnt]bool `json:"vu"`
	TeamScores       [BridgeTeamCnt]int  `json:"sc"`
	GamesWon         [BridgeTeamCnt]int  `json:"gw"`
	BelowLine        [BridgeTeamCnt]int  `json:"bl"`
	LeadPlayerIdx    int                 `json:"li"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Bridge) MarshalJSON() ([]byte, error) {
	return json.Marshal(bridgeJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		RoundNumber:      b.roundNumber,
		TrickNumber:      b.trickNumber,
		CurrentPlayerIdx: b.currentPlayerIdx,
		CurrentTrick:     b.currentTrick,
		DealerIdx:        b.dealerIdx,
		BidPlayerIdx:     b.bidPlayerIdx,
		BidHistory:       b.bidHistory,
		PassCount:        b.passCount,
		ContractLevel:    b.contractLevel,
		ContractSuit:     b.contractSuit,
		Doubled:          b.doubled,
		DeclarerIdx:      b.declarerIdx,
		DummyIdx:         b.dummyIdx,
		LastBidderIdx:    b.lastBidderIdx,
		LastBidTeam:      b.lastBidTeam,
		OpeningLeadDone:  b.openingLeadDone,
		TrumpSuit:        b.trumpSuit,
		Vulnerability:    b.vulnerability,
		TeamScores:       b.teamScores,
		GamesWon:         b.gamesWon,
		BelowLine:        b.belowLine,
		LeadPlayerIdx:    b.leadPlayerIdx,
		GameEndFlag:      b.gameEndFlag,
		WinnerTeam:       b.winnerTeam,
		ActionLog:        b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bridge) UnmarshalJSON(data []byte) error {
	var j bridgeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	if j.TrumpCards != nil {
		b.trumpCards = j.TrumpCards
	}
	if len(j.Players) > 0 && len(j.Players) <= 1000 {
		b.players = j.Players
	}
	b.config = j.Config
	b.phase = j.Phase
	b.roundNumber = j.RoundNumber
	b.trickNumber = j.TrickNumber
	b.currentPlayerIdx = j.CurrentPlayerIdx
	if len(j.CurrentTrick) <= 1000 {
		b.currentTrick = j.CurrentTrick
	}
	b.dealerIdx = j.DealerIdx
	b.bidPlayerIdx = j.BidPlayerIdx
	if len(j.BidHistory) <= 1000 {
		b.bidHistory = j.BidHistory
	}
	b.passCount = j.PassCount
	b.contractLevel = j.ContractLevel
	b.contractSuit = j.ContractSuit
	b.doubled = j.Doubled
	b.declarerIdx = j.DeclarerIdx
	b.dummyIdx = j.DummyIdx
	b.lastBidderIdx = j.LastBidderIdx
	b.lastBidTeam = j.LastBidTeam
	b.openingLeadDone = j.OpeningLeadDone
	b.trumpSuit = j.TrumpSuit
	b.vulnerability = j.Vulnerability
	b.teamScores = j.TeamScores
	b.gamesWon = j.GamesWon
	b.belowLine = j.BelowLine
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.gameEndFlag = j.GameEndFlag
	b.winnerTeam = j.WinnerTeam
	if len(j.ActionLog) <= 1000 {
		b.actionLog = j.ActionLog
	}

	return nil
}
