//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// BelotePlayerCnt ベロートプレイヤー数
const BelotePlayerCnt = 4

// BeloteHandSize 各プレイヤーの手札枚数
const BeloteHandSize = 8

// BeloteTeamCnt チーム数
const BeloteTeamCnt = 2

// BeloteFirstDealSize 初回配布で配るカード枚数 (ターンアップ前)
const BeloteFirstDealSize = 5

// BeloteDixDeDerDefault Dix de Der (最終トリック) ボーナス
const BeloteDixDeDerDefault = 10

// BeloteRebeloteBonus K+Q トランプの宣言ボーナス
const BeloteRebeloteBonus = 20

// BeloteRoundCardPointsTotal 1ラウンドのカード合計点数 (Dix de Der 含まず)
const BeloteRoundCardPointsTotal = 152

// BelotePhase ゲームフェーズ
type BelotePhase int

// Beloteのフェーズ定数
const (
	// BelotePhaseBidPickUp ピックアップフェーズ (ラウンド1: ターンアップのスートを切り札に指名するか)
	BelotePhaseBidPickUp BelotePhase = 0
	// BelotePhaseBidCallTrump コールトランプフェーズ (ラウンド2: 別のスートを切り札に指名するか)
	BelotePhaseBidCallTrump BelotePhase = 1
	// BelotePhasePlay トリックプレイフェーズ
	BelotePhasePlay BelotePhase = 2
	// BelotePhaseTrickEnd トリック終了フェーズ
	BelotePhaseTrickEnd BelotePhase = 3
	// BelotePhaseRoundEnd ラウンド終了フェーズ
	BelotePhaseRoundEnd BelotePhase = 4
	// BelotePhaseGameEnd ゲーム終了フェーズ
	BelotePhaseGameEnd BelotePhase = 5
)

// BeloteHint ヒント情報
type BeloteHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	OrderUp   *bool  // ピックアップすべきか
	Suit      *int   // 推奨スート (コールトランプ時)
	Reason    string // ヒント理由キー
}

// BeloteTrickCard トリック中の1枚
type BeloteTrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// Belote ベロートゲームクラス
type Belote struct {
	trumpCards        *TrumpCards
	players           []*BelotePlayer
	config            BeloteConfig
	phase             BelotePhase
	roundNumber       int
	trickNumber       int
	currentPlayerIdx  int
	currentTrick      []*BeloteTrickCard
	dealerIdx         int
	trumpSuit         int // 切り札スート (0 = 未確定, CardDesignSpade等)
	faceUpCard        *Card
	makerTeam         int // 切り札を指名したチーム (0 or 1)
	makerPlayerIdx    int // 切り札を指名したプレイヤー
	teamScores        [BeloteTeamCnt]int
	roundPoints       [BeloteTeamCnt]int // 当ラウンドの累計カード点数
	roundBeloteBonus  [BeloteTeamCnt]int // 当ラウンドの Belote+Rebelote ボーナス
	beloteHolderIdx   int                // K+Q を両方持つプレイヤー (-1 = なし)
	beloteKingPlayed  bool               // 当ラウンドで K of trumps が出されたか
	beloteQueenPlayed bool               // 当ラウンドで Q of trumps が出されたか
	beloteDeclared    bool               // 当ラウンドで Belote/Rebelote が宣言済か
	lastTrickWinner   int                // 直近の Dix de Der 用 (-1 = 未確定)
	leadPlayerIdx     int
	bidPlayerIdx      int // 現在のビッド手番
	bidPassCount      int // 連続パス数 (両ラウンド合計; 8で再配布)
	gameEndFlag       bool
	winnerTeam        int // 勝利チーム (-1 = 未確定)
	actionLog         []*ActionLogEntry
}

// NewBelote コンストラクタ
func NewBelote(trumpCards *TrumpCards, players []*BelotePlayer, config BeloteConfig) *Belote {
	return &Belote{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerTeam:      -1,
		roundNumber:     0,
		dealerIdx:       0,
		beloteHolderIdx: -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultBelote 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)
// と DefaultBeloteConfig を組み合わせたデフォルト構築。CUI/Web/Worker 共通の SSoT。
func NewDefaultBelote() *Belote {
	players := []*BelotePlayer{
		NewBelotePlayer(true, 0),
		NewBelotePlayer(false, 1),
		NewBelotePlayer(false, 0),
		NewBelotePlayer(false, 1),
	}
	return NewBelote(NewTrumpCardsBelote(), players, DefaultBeloteConfig())
}

// Reset ゲーム初期化
func (b *Belote) Reset() {
	b.gameEndFlag = false
	b.winnerTeam = -1
	b.roundNumber = 1
	b.trickNumber = 0
	b.dealerIdx = 0
	b.teamScores = [BeloteTeamCnt]int{}
	b.actionLog = nil
	b.trumpSuit = 0
	b.makerTeam = 0
	b.makerPlayerIdx = -1

	for _, p := range b.players {
		p.ResetRound()
	}

	b.beginRound()
}

// NextRound 次のラウンドを開始する
func (b *Belote) NextRound() {
	if b.phase != BelotePhaseRoundEnd {
		return
	}

	b.roundNumber++
	b.dealerIdx = (b.dealerIdx + 1) % BelotePlayerCnt
	b.trickNumber = 0
	b.trumpSuit = 0
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.makerPlayerIdx = -1

	for _, p := range b.players {
		p.ResetRound()
	}

	b.beginRound()
}

// beginRound ラウンドの初期処理 (配布 + ビッドフェーズ突入)
func (b *Belote) beginRound() {
	// Reset() doesn't clear these explicitly; clearing them here keeps a
	// mid-game reset from leaking ghost trick cards into the new bid phase.
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.roundPoints = [BeloteTeamCnt]int{}
	b.roundBeloteBonus = [BeloteTeamCnt]int{}
	b.beloteHolderIdx = -1
	b.beloteKingPlayed = false
	b.beloteQueenPlayed = false
	b.beloteDeclared = false
	b.lastTrickWinner = -1
	b.bidPassCount = 0
	b.faceUpCard = nil

	b.dealFirstFive()
	b.flipTurnUp()
	b.phase = BelotePhaseBidPickUp
	b.bidPlayerIdx = (b.dealerIdx + 1) % BelotePlayerCnt
}

// dealFirstFive 初回配布: 各プレイヤーに 5 枚配る (20 枚消費)
func (b *Belote) dealFirstFive() {
	b.trumpCards.Shuffle()
	for range BeloteFirstDealSize {
		for j := range BelotePlayerCnt {
			card := b.trumpCards.DrawCard()
			if card != nil {
				b.players[j].AddCard(card)
			}
		}
	}
}

// flipTurnUp 山札のトップを表向きにする
func (b *Belote) flipTurnUp() {
	b.faceUpCard = b.trumpCards.DrawCard()
}

// dealRemainder 残り3枚ずつを配る (トランプ確定後)
// トランプ取得者にターンアップカードを 1 枚渡し、残り 2 枚で 3 枚目を補う。
// 他プレイヤーには 3 枚配る。全員 8 枚になる。
func (b *Belote) dealRemainder(makerIdx int) {
	if b.faceUpCard != nil {
		b.players[makerIdx].AddCard(b.faceUpCard)
		b.faceUpCard = nil
	}
	for i := range BelotePlayerCnt {
		need := 3
		if i == makerIdx {
			need = 2
		}
		for range need {
			card := b.trumpCards.DrawCard()
			if card != nil {
				b.players[i].AddCard(card)
			}
		}
	}
	b.sortAllHands()
	b.detectBeloteHolder()
}

// detectBeloteHolder K+Q of trumps を両方持つプレイヤーを記録する
func (b *Belote) detectBeloteHolder() {
	if !b.config.EnableBeloteRebelote {
		return
	}
	for i, p := range b.players {
		hasK := false
		hasQ := false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c.GetDesign() != b.trumpSuit {
				continue
			}
			if c.GetValue() == 13 {
				hasK = true
			} else if c.GetValue() == 12 {
				hasQ = true
			}
		}
		if hasK && hasQ {
			b.beloteHolderIdx = i
			return
		}
	}
}

// --- Bid: PickUp (ラウンド1) ---

// PlayerPickUp 人間プレイヤーがピックアップ判断する
func (b *Belote) PlayerPickUp(orderUp bool) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BelotePhaseBidPickUp {
		return ErrWrongPhase
	}
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	if orderUp {
		b.doOrderUp(humanIdx)
	} else {
		b.appendLog(humanIdx, "pass", fmt.Sprintf("%s passes", b.playerName(humanIdx)), nil)
		b.advanceBidPickUp()
	}
	return nil
}

// CpuPickUp CPUプレイヤーがピックアップ判断する
func (b *Belote) CpuPickUp() {
	if b.gameEndFlag || b.phase != BelotePhaseBidPickUp {
		return
	}
	if b.bidPlayerIdx < 0 || b.bidPlayerIdx >= BelotePlayerCnt {
		return
	}
	if b.players[b.bidPlayerIdx].GetIsHuman() {
		return
	}

	if b.cpuSelectPickUp(b.bidPlayerIdx) {
		b.doOrderUp(b.bidPlayerIdx)
	} else {
		b.appendLog(b.bidPlayerIdx, "pass", fmt.Sprintf("%s passes", b.playerName(b.bidPlayerIdx)), nil)
		b.advanceBidPickUp()
	}
}

// doOrderUp ターンアップのスートを切り札に確定する
func (b *Belote) doOrderUp(playerIdx int) {
	b.trumpSuit = b.faceUpCard.GetDesign()
	b.makerTeam = b.players[playerIdx].GetTeam()
	b.makerPlayerIdx = playerIdx
	b.appendLog(playerIdx, "order_up",
		fmt.Sprintf("%s takes %s as trump", b.playerName(playerIdx), cardStr(b.faceUpCard)),
		[]*Card{b.faceUpCard})

	b.dealRemainder(playerIdx)
	b.startPlayPhase()
}

// advanceBidPickUp ピックアップフェーズのビッドを進める
func (b *Belote) advanceBidPickUp() {
	b.bidPassCount++
	b.bidPlayerIdx = (b.bidPlayerIdx + 1) % BelotePlayerCnt

	startIdx := (b.dealerIdx + 1) % BelotePlayerCnt
	if b.bidPlayerIdx == startIdx {
		b.phase = BelotePhaseBidCallTrump
	}
}

// --- Bid: CallTrump (ラウンド2) ---

// PlayerCallTrump 人間プレイヤーがスートを指名する
func (b *Belote) PlayerCallTrump(suit int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BelotePhaseBidCallTrump {
		return ErrWrongPhase
	}
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if b.faceUpCard != nil && suit == b.faceUpCard.GetDesign() {
		return NewDomainError(ErrInvalidPlay, "ラウンド1の表向きスートは選べません")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}
	b.doCallTrump(humanIdx, suit)
	return nil
}

// PlayerPassCall 人間プレイヤーがコールフェーズでパスする
func (b *Belote) PlayerPassCall() error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BelotePhaseBidCallTrump {
		return ErrWrongPhase
	}
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	b.appendLog(humanIdx, "pass", fmt.Sprintf("%s passes", b.playerName(humanIdx)), nil)
	b.advanceBidCallTrump()
	return nil
}

// CpuCallTrump CPUプレイヤーがコールトランプ判断する
func (b *Belote) CpuCallTrump() {
	if b.gameEndFlag || b.phase != BelotePhaseBidCallTrump {
		return
	}
	if b.players[b.bidPlayerIdx].GetIsHuman() {
		return
	}

	suit := b.cpuSelectCallTrump(b.bidPlayerIdx)
	if suit > 0 {
		b.doCallTrump(b.bidPlayerIdx, suit)
	} else {
		b.appendLog(b.bidPlayerIdx, "pass", fmt.Sprintf("%s passes", b.playerName(b.bidPlayerIdx)), nil)
		b.advanceBidCallTrump()
	}
}

// doCallTrump スートを指名する
func (b *Belote) doCallTrump(playerIdx int, suit int) {
	b.trumpSuit = suit
	b.makerTeam = b.players[playerIdx].GetTeam()
	b.makerPlayerIdx = playerIdx
	suitName := suitStr(suit)
	b.appendLog(playerIdx, "call_trump",
		fmt.Sprintf("%s calls %s as trump", b.playerName(playerIdx), suitName), nil)

	b.dealRemainder(playerIdx)
	b.startPlayPhase()
}

// advanceBidCallTrump コールトランプフェーズのビッドを進める
// 全員パスでパスアウト時はラウンドを再配布する。
func (b *Belote) advanceBidCallTrump() {
	b.bidPassCount++
	b.bidPlayerIdx = (b.bidPlayerIdx + 1) % BelotePlayerCnt

	startIdx := (b.dealerIdx + 1) % BelotePlayerCnt
	if b.bidPlayerIdx == startIdx {
		// パスアウト: ディーラーを次に進めて再配布
		b.appendLog(-1, "pass_out", "All players passed; redealing", nil)
		b.dealerIdx = (b.dealerIdx + 1) % BelotePlayerCnt
		for _, p := range b.players {
			p.ResetRound()
		}
		b.beginRound()
	}
}

// --- Play Phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Belote) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BelotePhasePlay {
		return ErrWrongPhase
	}
	if !b.players[b.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := b.players[b.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := b.validatePlay(b.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	b.playCard(b.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行
func (b *Belote) CpuPlay() {
	if b.gameEndFlag || b.phase != BelotePhasePlay {
		return
	}
	if b.players[b.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := b.players[b.currentPlayerIdx]
	cardIdx := b.cpuSelectPlayCard(b.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	b.playCard(b.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (b *Belote) ResolveTrick() {
	if b.phase != BelotePhaseTrickEnd || len(b.currentTrick) != BelotePlayerCnt {
		return
	}

	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	trickPoints := 0
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += beloteCardPoints(tc.Card, b.trumpSuit)
	}

	b.players[winnerIdx].AddTrick(trickCards)
	b.roundPoints[b.players[winnerIdx].GetTeam()] += trickPoints

	winnerName := b.playerName(winnerIdx)
	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", winnerName, b.trickNumber, trickPoints),
		trickCards)

	b.leadPlayerIdx = winnerIdx
	b.lastTrickWinner = winnerIdx

	if b.trickNumber >= BeloteHandSize {
		// Dix de Der
		b.roundPoints[b.players[winnerIdx].GetTeam()] += b.config.DixDeDer
		b.appendLog(winnerIdx, "dix_de_der",
			fmt.Sprintf("%s wins last trick +%d (Dix de Der)", winnerName, b.config.DixDeDer), nil)
		b.phase = BelotePhaseRoundEnd
	} else {
		b.phase = BelotePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (b *Belote) NextTrick() {
	if b.phase != BelotePhaseTrickEnd {
		return
	}
	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = BelotePhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (b *Belote) ScoreRound() {
	if b.phase != BelotePhaseRoundEnd {
		return
	}

	maker := b.makerTeam
	defender := 1 - b.makerTeam
	makerPts := b.roundPoints[maker] + b.roundBeloteBonus[maker]
	defenderPts := b.roundPoints[defender] + b.roundBeloteBonus[defender]

	switch {
	case makerPts > defenderPts:
		// メイカー勝利
		b.teamScores[maker] += makerPts
		b.teamScores[defender] += defenderPts
		b.appendLog(-1, "maker_win",
			fmt.Sprintf("Team %d (maker) wins the round: %d vs %d",
				maker, makerPts, defenderPts), nil)
	case makerPts == defenderPts:
		// 同点 = メイカー側 dedans (Litige); 防衛側がカード点を総取り
		b.teamScores[defender] += defenderPts + b.roundPoints[maker]
		// Belote ボーナスはどちらの宣言でも自チームに残る
		b.teamScores[maker] += b.roundBeloteBonus[maker]
		b.appendLog(-1, "litige",
			fmt.Sprintf("Litige (tie %d-%d): defenders take maker's card points",
				makerPts, defenderPts), nil)
	default:
		// メイカー dedans: 防衛側がカード点を総取り
		b.teamScores[defender] += defenderPts + b.roundPoints[maker]
		b.teamScores[maker] += b.roundBeloteBonus[maker]
		b.appendLog(-1, "dedans",
			fmt.Sprintf("Team %d (maker) is dedans: %d < %d", maker, makerPts, defenderPts), nil)
	}

	// Capot ボーナス (全 8 トリック獲得)
	makerTricks := 0
	for _, p := range b.players {
		if p.GetTeam() == maker {
			makerTricks += p.GetTrickCount()
		}
	}
	switch makerTricks {
	case BeloteHandSize:
		b.teamScores[maker] += 90
		b.appendLog(-1, "capot",
			fmt.Sprintf("Team %d capot (+90)", maker), nil)
	case 0:
		b.teamScores[defender] += 90
		b.appendLog(-1, "capot_defender",
			fmt.Sprintf("Team %d capot against maker (+90)", defender), nil)
	}

	for ti := range BeloteTeamCnt {
		b.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points (total %d)",
				ti, b.roundPoints[ti]+b.roundBeloteBonus[ti], b.teamScores[ti]), nil)
	}

	b.checkGameEnd()
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (b *Belote) GetPhase() BelotePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Belote) SetPhase(p BelotePhase) { b.phase = p }

// GetRoundNumber 現在のラウンド番号取得
func (b *Belote) GetRoundNumber() int { return b.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (b *Belote) SetRoundNumber(n int) { b.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (b *Belote) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Belote) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Belote) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Belote) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Belote) GetCurrentTrick() []*BeloteTrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Belote) SetCurrentTrick(trick []*BeloteTrickCard) { b.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Belote) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (b *Belote) GetWinnerTeam() int { return b.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (b *Belote) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Belote) GetPlayer(i int) *BelotePlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Belote) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Belote) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (b *Belote) GetBidPlayerIdx() int { return b.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (b *Belote) SetBidPlayerIdx(idx int) { b.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Belote) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Belote) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得
func (b *Belote) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (b *Belote) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetFaceUpCard 表向きカード取得 (nil = ピックアップ済み)
func (b *Belote) GetFaceUpCard() *Card { return b.faceUpCard }

// SetFaceUpCard 表向きカード設定 (テスト用)
func (b *Belote) SetFaceUpCard(card *Card) { b.faceUpCard = card }

// GetMakerTeam メイカーチーム取得
func (b *Belote) GetMakerTeam() int { return b.makerTeam }

// SetMakerTeam メイカーチーム設定 (テスト用)
func (b *Belote) SetMakerTeam(team int) { b.makerTeam = team }

// GetMakerPlayerIdx メイカープレイヤー取得
func (b *Belote) GetMakerPlayerIdx() int { return b.makerPlayerIdx }

// SetBeloteHolderIdx Belote/Rebelote (K+Q トランプ) 所持者を設定する (テスト用)。
// 通常は dealRemainder 経由で detectBeloteHolder が呼ばれて埋まる。
func (b *Belote) SetBeloteHolderIdx(idx int) { b.beloteHolderIdx = idx }

// GetTeamScore チームスコア取得
func (b *Belote) GetTeamScore(team int) int {
	if team < 0 || team >= BeloteTeamCnt {
		return 0
	}
	return b.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (b *Belote) SetTeamScore(team, score int) {
	if team >= 0 && team < BeloteTeamCnt {
		b.teamScores[team] = score
	}
}

// GetRoundPoints 当ラウンドのチーム別カード点数取得
func (b *Belote) GetRoundPoints(team int) int {
	if team < 0 || team >= BeloteTeamCnt {
		return 0
	}
	return b.roundPoints[team]
}

// GetRoundBeloteBonus 当ラウンドの Belote/Rebelote ボーナス取得
func (b *Belote) GetRoundBeloteBonus(team int) int {
	if team < 0 || team >= BeloteTeamCnt {
		return 0
	}
	return b.roundBeloteBonus[team]
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Belote) IsHumanTurn() bool {
	if b.currentPlayerIdx < 0 || b.currentPlayerIdx >= len(b.players) {
		return false
	}
	return b.players[b.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (b *Belote) IsHumanBidTurn() bool {
	if b.bidPlayerIdx < 0 || b.bidPlayerIdx >= len(b.players) {
		return false
	}
	return b.players[b.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (b *Belote) GetConfig() BeloteConfig { return b.config }

// SetConfig 設定変更
func (b *Belote) SetConfig(cfg BeloteConfig) { b.config = cfg }

// GetActionLog 棋譜取得
func (b *Belote) GetActionLog() []*ActionLogEntry { return b.actionLog }

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (b *Belote) CardRankPublic(card *Card) int { return b.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (b *Belote) CardPointsPublic(card *Card) int { return beloteCardPoints(card, b.trumpSuit) }

// --- Ranking + scoring helpers ---

// beloteTrumpRank トランプスートのカードランク (高 = 強)
// J=8, 9=7, A=6, 10=5, K=4, Q=3, 8=2, 7=1
func beloteTrumpRank(value int) int {
	switch value {
	case 11:
		return 8
	case 9:
		return 7
	case 1:
		return 6
	case 10:
		return 5
	case 13:
		return 4
	case 12:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// beloteNonTrumpRank 非トランプスートのカードランク (高 = 強)
// A=8, 10=7, K=6, Q=5, J=4, 9=3, 8=2, 7=1
func beloteNonTrumpRank(value int) int {
	switch value {
	case 1:
		return 8
	case 10:
		return 7
	case 13:
		return 6
	case 12:
		return 5
	case 11:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// beloteCardPoints トランプスートを踏まえたカード点数を返す
// 切り札: J=20, 9=14, A=11, 10=10, K=4, Q=3, 8=0, 7=0
// 非切り札: A=11, 10=10, K=4, Q=3, J=2, 9=0, 8=0, 7=0
func beloteCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11:
			return 20
		case 9:
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}

// cardRank トリック比較用ランクを返す (高い = 強い)
// 切り札スート: 200 + trumpRank, 非切り札: 100 + nonTrumpRank
func (b *Belote) cardRank(card *Card) int {
	if card.GetDesign() == b.trumpSuit {
		return 200 + beloteTrumpRank(card.GetValue())
	}
	return 100 + beloteNonTrumpRank(card.GetValue())
}

// --- Trick play helpers ---

// startPlayPhase プレイフェーズを開始する
func (b *Belote) startPlayPhase() {
	b.trickNumber = 1
	b.currentTrick = nil
	b.leadPlayerIdx = (b.dealerIdx + 1) % BelotePlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.phase = BelotePhasePlay
}

// playCard カードをプレイする共通処理
func (b *Belote) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &BeloteTrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	b.maybeDeclareBeloteRebelote(playerIdx, card)
	b.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", b.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(b.currentTrick) == BelotePlayerCnt {
		b.phase = BelotePhaseTrickEnd
	} else {
		b.currentPlayerIdx = (b.currentPlayerIdx + 1) % BelotePlayerCnt
	}
}

// maybeDeclareBeloteRebelote K+Q トランプの宣言を自動処理する
func (b *Belote) maybeDeclareBeloteRebelote(playerIdx int, card *Card) {
	if !b.config.EnableBeloteRebelote {
		return
	}
	if playerIdx != b.beloteHolderIdx {
		return
	}
	if card.GetDesign() != b.trumpSuit {
		return
	}
	switch card.GetValue() {
	case 13:
		b.beloteKingPlayed = true
	case 12:
		b.beloteQueenPlayed = true
	default:
		return
	}
	if b.beloteKingPlayed && b.beloteQueenPlayed && !b.beloteDeclared {
		team := b.players[playerIdx].GetTeam()
		b.roundBeloteBonus[team] += BeloteRebeloteBonus
		b.beloteDeclared = true
		b.appendLog(playerIdx, "belote_rebelote",
			fmt.Sprintf("%s declares Belote/Rebelote (+%d)",
				b.playerName(playerIdx), BeloteRebeloteBonus), nil)
	}
}

// validatePlay カードのプレイが Belote のルールに従っているか検証する
// ベロートの義務:
//
//  1. フォロースート可能ならフォロースート (リードスートのカードがある限り)
//  2. フォロースート不可かつトランプがある場合は必ずトランプを出す (obligation à couper)
//  3. トランプを出す場合、既出のトランプより強いトランプがあるなら必ずオーバートランプする (obligation à monter)
//  4. リードがトランプの場合、必ずフォロー (＋オーバートランプ可能ならする)
func (b *Belote) validatePlay(playerIdx int, card *Card) error {
	if len(b.currentTrick) == 0 {
		return nil
	}
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	hasLead := b.playerHasSuit(player, leadSuit)
	cardSuit := card.GetDesign()

	if leadSuit == b.trumpSuit {
		// リードがトランプ: トランプを必ず出す。出すなら可能な限りオーバートランプ。
		if hasLead {
			if cardSuit != b.trumpSuit {
				return NewDomainError(ErrInvalidPlay, "リードスート (切り札) に従ってください")
			}
			highest := b.highestTrumpInTrick()
			canOverTrump := b.playerCanBeatTrump(player, highest)
			if canOverTrump && beloteTrumpRank(card.GetValue()) <= highest {
				return NewDomainError(ErrInvalidPlay, "オーバートランプしてください (obligation à monter)")
			}
			return nil
		}
		// リードがトランプ かつ自分はトランプを持っていない: 任意のカード可
		return nil
	}

	// リードが非トランプ
	if hasLead {
		if cardSuit != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		return nil
	}

	// フォロースート不可
	hasTrump := b.playerHasSuit(player, b.trumpSuit)
	trickHasTrump := b.trickContainsTrump()
	partnerIdx := (playerIdx + 2) % BelotePlayerCnt
	partnerWinning := b.partnerIsCurrentlyWinning(playerIdx, partnerIdx)

	if hasTrump && !partnerWinning {
		// トランプ義務
		if cardSuit != b.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "切り札を出してください (obligation à couper)")
		}
		// オーバートランプ義務 (トリックに既に切り札が出ている場合)
		if trickHasTrump {
			highest := b.highestTrumpInTrick()
			canOverTrump := b.playerCanBeatTrump(player, highest)
			if canOverTrump && beloteTrumpRank(card.GetValue()) <= highest {
				return NewDomainError(ErrInvalidPlay, "オーバートランプしてください (obligation à monter)")
			}
		}
		return nil
	}
	// 切り札なし or パートナーが現勝者: 任意のカード可
	return nil
}

// playerHasSuit プレイヤーが特定スートを持っているか
func (b *Belote) playerHasSuit(p *BelotePlayer, suit int) bool {
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// playerCanBeatTrump プレイヤーが指定ランク超えのトランプを持っているか
func (b *Belote) playerCanBeatTrump(p *BelotePlayer, rank int) bool {
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() == b.trumpSuit && beloteTrumpRank(c.GetValue()) > rank {
			return true
		}
	}
	return false
}

// highestTrumpInTrick 現トリック内の最強トランプランク (なければ 0)
func (b *Belote) highestTrumpInTrick() int {
	best := 0
	for _, tc := range b.currentTrick {
		if tc.Card.GetDesign() != b.trumpSuit {
			continue
		}
		r := beloteTrumpRank(tc.Card.GetValue())
		if r > best {
			best = r
		}
	}
	return best
}

// trickContainsTrump 現トリックにトランプが含まれているか
func (b *Belote) trickContainsTrump() bool {
	for _, tc := range b.currentTrick {
		if tc.Card.GetDesign() == b.trumpSuit {
			return true
		}
	}
	return false
}

// partnerIsCurrentlyWinning 現トリックでパートナーが現勝者か
func (b *Belote) partnerIsCurrentlyWinning(playerIdx, partnerIdx int) bool {
	if len(b.currentTrick) == 0 {
		return false
	}
	winnerIdx := b.currentLeader()
	return winnerIdx == partnerIdx
}

// currentLeader 現在のトリック先頭時点での仮勝者を返す
func (b *Belote) currentLeader() int {
	if len(b.currentTrick) == 0 {
		return -1
	}
	winner := b.currentTrick[0].PlayerIdx
	winnerRank := b.cardRank(b.currentTrick[0].Card)
	winnerSuit := b.currentTrick[0].Card.GetDesign()
	for _, tc := range b.currentTrick[1:] {
		suit := tc.Card.GetDesign()
		rank := b.cardRank(tc.Card)
		if suit == b.trumpSuit && winnerSuit != b.trumpSuit {
			winner = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = suit
			continue
		}
		if suit == winnerSuit && rank > winnerRank {
			winner = tc.PlayerIdx
			winnerRank = rank
		}
	}
	return winner
}

// trickWinner トリックの勝者を決定する
func (b *Belote) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	return b.currentLeader()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (b *Belote) GetValidPlayIndices(playerIdx int) []int {
	return b.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (b *Belote) getValidPlayIndices(playerIdx int) []int {
	player := b.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return b.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Game end + bookkeeping ---

func (b *Belote) checkGameEnd() {
	for ti := range BeloteTeamCnt {
		if b.teamScores[ti] >= b.config.TargetScore {
			b.gameEndFlag = true
			b.phase = BelotePhaseGameEnd
			if b.teamScores[0] >= b.teamScores[1] {
				b.winnerTeam = 0
			} else {
				b.winnerTeam = 1
			}
			b.appendLog(-1, "game_end",
				fmt.Sprintf("Team %d wins the game!", b.winnerTeam), nil)
			return
		}
	}
}

// sortAllHands 全プレイヤーの手札をソートする (スート → ランク)
func (b *Belote) sortAllHands() {
	for _, p := range b.players {
		beloteSortHand(p, b)
	}
}

func beloteSortHand(p *BelotePlayer, b *Belote) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si := ci.GetDesign()
		sj := cj.GetDesign()
		if si != sj {
			return si < sj
		}
		// Strongest card first within each suit — matches the manual's display
		// example and lets the CPU's "play valid[0]" easy-lead actually pick a
		// strong card instead of the weakest.
		return b.cardRank(ci) > b.cardRank(cj)
	})
}

func (b *Belote) findHumanIdx() int {
	for i, p := range b.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

func (b *Belote) playerName(idx int) string {
	if idx < 0 || idx >= len(b.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if b.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

func (b *Belote) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Hints ---

// GetHint 現フェーズのヒントを返す (人間プレイヤー視点)
func (b *Belote) GetHint() *BeloteHint {
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}
	switch b.phase {
	case BelotePhaseBidPickUp:
		if b.bidPlayerIdx != humanIdx {
			return nil
		}
		ok := b.cpuEvalPickUp(humanIdx)
		return &BeloteHint{OrderUp: &ok, Reason: "strategic_pickup"}
	case BelotePhaseBidCallTrump:
		if b.bidPlayerIdx != humanIdx {
			return nil
		}
		suit := b.cpuEvalCallTrump(humanIdx)
		if suit > 0 {
			return &BeloteHint{Suit: &suit, Reason: "strategic_call"}
		}
		return &BeloteHint{Reason: "pass_recommended"}
	case BelotePhasePlay:
		if b.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := b.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := b.cpuPlayChoose(humanIdx, valid)
		return &BeloteHint{CardIndex: &idx, Reason: b.playHintReason(idx)}
	}
	return nil
}

func (b *Belote) playHintReason(chosenIdx int) string {
	humanIdx := b.findHumanIdx()
	if humanIdx < 0 {
		return ""
	}
	card := b.players[humanIdx].GetCard(chosenIdx)
	if len(b.currentTrick) == 0 {
		if card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == b.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- CPU AI ---

// cpuSelectPickUp CPUがピックアップ判断する (true = 取得)
func (b *Belote) cpuSelectPickUp(playerIdx int) bool {
	switch b.config.CpuDifficulty {
	case BeloteCpuDifficultyHard:
		return b.cpuEvalPickUp(playerIdx)
	case BeloteCpuDifficultyNormal:
		return b.cpuPickUpNormal(playerIdx)
	default:
		return b.cpuPickUpEasy(playerIdx)
	}
}

func (b *Belote) cpuPickUpEasy(playerIdx int) bool {
	if rand.Intn(100) < 25 {
		return true
	}
	if playerIdx == b.dealerIdx && rand.Intn(100) < 40 {
		return true
	}
	return false
}

func (b *Belote) cpuPickUpNormal(playerIdx int) bool {
	score := b.evalHandForTrump(playerIdx, b.faceUpCard.GetDesign())
	if score >= 25 {
		return true
	}
	if playerIdx == b.dealerIdx && score >= 18 {
		return true
	}
	return false
}

func (b *Belote) cpuEvalPickUp(playerIdx int) bool {
	score := b.evalHandForTrump(playerIdx, b.faceUpCard.GetDesign())
	if playerIdx == b.dealerIdx {
		score += 6
	}
	return score >= 28
}

// cpuSelectCallTrump CPUがコールトランプで指名するスート (0 = パス)
func (b *Belote) cpuSelectCallTrump(playerIdx int) int {
	bestSuit := 0
	bestScore := 0
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if b.faceUpCard != nil && suit == b.faceUpCard.GetDesign() {
			continue
		}
		score := b.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	threshold := 30
	switch b.config.CpuDifficulty {
	case BeloteCpuDifficultyEasy:
		threshold = 22
	case BeloteCpuDifficultyHard:
		threshold = 34
	}
	if bestScore >= threshold {
		return bestSuit
	}
	return 0
}

func (b *Belote) cpuEvalCallTrump(playerIdx int) int {
	return b.cpuSelectCallTrump(playerIdx)
}

// evalHandForTrump 仮定したトランプスートに対する手札評価値を返す
// (高い = 強い: トランプ J/9/A、長いトランプ列、外スートの A をボーナス)
func (b *Belote) evalHandForTrump(playerIdx, trumpSuit int) int {
	p := b.players[playerIdx]
	score := 0
	trumpCount := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() == trumpSuit {
			trumpCount++
			switch c.GetValue() {
			case 11:
				score += 14 // J
			case 9:
				score += 10
			case 1:
				score += 7 // A
			case 10:
				score += 5
			case 13:
				score += 3
			case 12:
				score += 2
			}
			continue
		}
		switch c.GetValue() {
		case 1:
			score += 4 // 外スート A
		case 10:
			score += 2
		}
	}
	if trumpCount >= 4 {
		score += 6
	} else if trumpCount == 3 {
		score += 3
	}
	return score
}

// cpuSelectPlayCard CPUがプレイするカードを選ぶ
func (b *Belote) cpuSelectPlayCard(playerIdx int) int {
	valid := b.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	switch b.config.CpuDifficulty {
	case BeloteCpuDifficultyEasy:
		return valid[rand.Intn(len(valid))]
	default:
		return b.cpuPlayChoose(playerIdx, valid)
	}
}

// cpuPlayChoose 標準ヒューリスティック:
//   - リード時: 強いトランプ (J/9) または高得点の非トランプ A/10 を優先
//   - フォロー時: 勝てるなら最弱の勝てるカード、勝てないなら最低点のカードを捨てる
func (b *Belote) cpuPlayChoose(playerIdx int, valid []int) int {
	player := b.players[playerIdx]
	if len(b.currentTrick) == 0 {
		// リード: 強い切り札 (J/9) で相手の切り札を引き出すか、外スートの A/10 を切り出す。
		best := valid[0]
		bestScore := -1
		for _, idx := range valid {
			c := player.GetCard(idx)
			s := 0
			if c.GetDesign() == b.trumpSuit {
				switch c.GetValue() {
				case 11: // Trump J — strongest; lead it to flush opponents' trumps.
					s = 25
				case 9: // Trump 9 — second-strongest; also useful as a lead.
					s = 15
				}
			} else {
				switch c.GetValue() {
				case 1:
					s = 30
				case 10:
					s = 20
				case 13:
					s = 5
				}
			}
			if s > bestScore {
				bestScore = s
				best = idx
			}
		}
		return best
	}

	// フォロー時
	winnerIdx := b.currentLeader()
	partnerIdx := (playerIdx + 2) % BelotePlayerCnt
	partnerWinning := winnerIdx == partnerIdx

	if partnerWinning {
		// パートナーが勝者: 最も価値が高い (= 点数高い) カードを出す
		best := valid[0]
		bestPts := -1
		for _, idx := range valid {
			pts := beloteCardPoints(player.GetCard(idx), b.trumpSuit)
			if pts > bestPts {
				bestPts = pts
				best = idx
			}
		}
		return best
	}

	// 勝てるカードがあれば最弱の勝てるカードを出す
	winnable := -1
	winnableRank := 9999
	for _, idx := range valid {
		c := player.GetCard(idx)
		if b.cardWouldWinTrick(c) {
			r := b.cardRank(c)
			if r < winnableRank {
				winnableRank = r
				winnable = idx
			}
		}
	}
	if winnable >= 0 {
		return winnable
	}

	// 勝てない: 最低点のカードを捨てる
	worst := valid[0]
	worstPts := 9999
	for _, idx := range valid {
		pts := beloteCardPoints(player.GetCard(idx), b.trumpSuit)
		if pts < worstPts {
			worstPts = pts
			worst = idx
		}
	}
	return worst
}

// cardWouldWinTrick 指定カードを今出した場合に現状の勝者を上回るか
func (b *Belote) cardWouldWinTrick(c *Card) bool {
	if len(b.currentTrick) == 0 {
		return true
	}
	winIdx := b.currentLeader()
	var winCard *Card
	for _, tc := range b.currentTrick {
		if tc.PlayerIdx == winIdx {
			winCard = tc.Card
			break
		}
	}
	if winCard == nil {
		return true
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	cSuit := c.GetDesign()
	wSuit := winCard.GetDesign()

	if cSuit == b.trumpSuit && wSuit != b.trumpSuit {
		return true
	}
	if cSuit == wSuit {
		return b.cardRank(c) > b.cardRank(winCard)
	}
	// 現勝者がトランプの場合: 非トランプは勝てない
	if wSuit == b.trumpSuit {
		return false
	}
	// 同じ非トランプ・リードスート同士は cSuit==wSuit で扱い済み
	// c がリードスートで wSuit が非リードスートの非トランプ (起こり得ない)
	if cSuit == leadSuit {
		return b.cardRank(c) > b.cardRank(winCard)
	}
	return false
}

// --- JSON ---

// beloteJSON Belote の JSON 表現
type beloteJSON struct {
	TrumpCards        *TrumpCards        `json:"tc"`
	Players           []*BelotePlayer    `json:"pl"`
	Config            BeloteConfig       `json:"cfg"`
	Phase             BelotePhase        `json:"ph"`
	RoundNumber       int                `json:"rn"`
	TrickNumber       int                `json:"tn"`
	CurrentPlayerIdx  int                `json:"cp"`
	CurrentTrick      []*BeloteTrickCard `json:"ct"`
	DealerIdx         int                `json:"di"`
	TrumpSuit         int                `json:"ts"`
	FaceUpCard        *Card              `json:"fu,omitempty"`
	MakerTeam         int                `json:"mt"`
	MakerPlayerIdx    int                `json:"mp"`
	TeamScores        [BeloteTeamCnt]int `json:"sc"`
	RoundPoints       [BeloteTeamCnt]int `json:"rp"`
	RoundBeloteBonus  [BeloteTeamCnt]int `json:"rb"`
	BeloteHolderIdx   int                `json:"bh"`
	BeloteKingPlayed  bool               `json:"bk"`
	BeloteQueenPlayed bool               `json:"bq"`
	BeloteDeclared    bool               `json:"bd"`
	LastTrickWinner   int                `json:"lw"`
	LeadPlayerIdx     int                `json:"li"`
	BidPlayerIdx      int                `json:"bi"`
	BidPassCount      int                `json:"bp"`
	GameEndFlag       bool               `json:"ge"`
	WinnerTeam        int                `json:"wt"`
	ActionLog         []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Belote) MarshalJSON() ([]byte, error) {
	return json.Marshal(beloteJSON{
		TrumpCards:        b.trumpCards,
		Players:           b.players,
		Config:            b.config,
		Phase:             b.phase,
		RoundNumber:       b.roundNumber,
		TrickNumber:       b.trickNumber,
		CurrentPlayerIdx:  b.currentPlayerIdx,
		CurrentTrick:      b.currentTrick,
		DealerIdx:         b.dealerIdx,
		TrumpSuit:         b.trumpSuit,
		FaceUpCard:        b.faceUpCard,
		MakerTeam:         b.makerTeam,
		MakerPlayerIdx:    b.makerPlayerIdx,
		TeamScores:        b.teamScores,
		RoundPoints:       b.roundPoints,
		RoundBeloteBonus:  b.roundBeloteBonus,
		BeloteHolderIdx:   b.beloteHolderIdx,
		BeloteKingPlayed:  b.beloteKingPlayed,
		BeloteQueenPlayed: b.beloteQueenPlayed,
		BeloteDeclared:    b.beloteDeclared,
		LastTrickWinner:   b.lastTrickWinner,
		LeadPlayerIdx:     b.leadPlayerIdx,
		BidPlayerIdx:      b.bidPlayerIdx,
		BidPassCount:      b.bidPassCount,
		GameEndFlag:       b.gameEndFlag,
		WinnerTeam:        b.winnerTeam,
		ActionLog:         b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Belote) UnmarshalJSON(data []byte) error {
	var j beloteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.trumpCards = j.TrumpCards
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.roundNumber = j.RoundNumber
	b.trickNumber = j.TrickNumber
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.currentTrick = j.CurrentTrick
	b.dealerIdx = j.DealerIdx
	b.trumpSuit = j.TrumpSuit
	b.faceUpCard = j.FaceUpCard
	b.makerTeam = j.MakerTeam
	b.makerPlayerIdx = j.MakerPlayerIdx
	b.teamScores = j.TeamScores
	b.roundPoints = j.RoundPoints
	b.roundBeloteBonus = j.RoundBeloteBonus
	b.beloteHolderIdx = j.BeloteHolderIdx
	b.beloteKingPlayed = j.BeloteKingPlayed
	b.beloteQueenPlayed = j.BeloteQueenPlayed
	b.beloteDeclared = j.BeloteDeclared
	b.lastTrickWinner = j.LastTrickWinner
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.bidPlayerIdx = j.BidPlayerIdx
	b.bidPassCount = j.BidPassCount
	b.gameEndFlag = j.GameEndFlag
	b.winnerTeam = j.WinnerTeam
	b.actionLog = j.ActionLog
	return nil
}
