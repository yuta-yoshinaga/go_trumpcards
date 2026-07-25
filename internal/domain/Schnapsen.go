//go:build !js || !wasm || solo

// Package domain シュナプセン / Sixty-Six (Schnapsen / 66) のドメインモデル。
//
// Schnapsen はドイツ・オーストリアで人気の 2 人用トリックテイキングゲーム。
// 20 枚デッキ (A,10,K,Q,J × 4 スート) を使う。各プレイヤーに 5 枚配り、
// 次の 1 枚を表向きの切り札表示カードとして山札の底に置く。
//
// 山札がある間 (第1フェーズ) はマストフォローが無く、何を出してもよい。
// トリックの勝者が山札から 1 枚、敗者も 1 枚を引いて補充する (Briscola 方式)。
// 自分のリード番で同スートの K と Q を両方持っていれば「マリアージュ」を
// 宣言してボーナス点 (20 点、切り札なら 40 点) を獲得し、その K か Q を
// リードする。
//
// 山札が尽きると第2フェーズに移行し、マストフォロー (フォローできる場合は
// 同スート、勝てるなら勝つ、無ければ切り札) が義務付けられる。
// カード点 (A=11,10=10,K=4,Q=3,J=2) とマリアージュボーナスの合計が
// 先に 66 点に達した側がラウンド勝利。誰も 66 点に達せずカードが尽きた
// 場合は最後のトリックを取った側の勝利。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SchnapsenPlayerCnt シュナプセンのプレイヤー数 (2人固定)
const SchnapsenPlayerCnt = 2

// SchnapsenHandSize 各プレイヤーの手札最大枚数 (山札がある間は補充される)
const SchnapsenHandSize = 5

// SchnapsenWinThreshold ラウンド勝利点 (これに達した側が勝ち)
const SchnapsenWinThreshold = 66

// SchnapsenTotalPoints デッキ全体の合計カード点
const SchnapsenTotalPoints = 120

// SchnapsenMarriageBonus 通常スートのマリアージュ (K+Q) ボーナス
const SchnapsenMarriageBonus = 20

// SchnapsenRoyalMarriageBonus 切り札スートのマリアージュ (ロイヤルマリアージュ) ボーナス
const SchnapsenRoyalMarriageBonus = 40

// SchnapsenPhase ゲームフェーズ
type SchnapsenPhase int

// Schnapsenのフェーズ定数
const (
	// SchnapsenPhasePlay トリックプレイフェーズ
	SchnapsenPhasePlay SchnapsenPhase = iota
	// SchnapsenPhaseTrickEnd トリック終了フェーズ
	SchnapsenPhaseTrickEnd
	// SchnapsenPhaseGameEnd ゲーム終了フェーズ
	SchnapsenPhaseGameEnd
)

// SchnapsenTrickCard トリック中の1枚
type SchnapsenTrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// SchnapsenHint ヒント情報
type SchnapsenHint struct {
	CardIndex  *int   // 推奨カードインデックス
	Reason     string // ヒント理由キー
	IsMarriage bool   // 推奨アクションがマリアージュ宣言かどうか
}

// SchnapsenCardPoints カードの得点を返す (A=11,10=10,K=4,Q=3,J=2; その他=0)。
// switch 実装はパッケージ初期化時のグローバルマップを避け、全 Cloudflare
// Worker WASM バイナリ (classic は 1 MB gzip 上限) のサイズを抑える。
func SchnapsenCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // Ace
		return 11
	case 10: // Ten
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

// SchnapsenRankOrder カードのスート内順位を返す (大きいほど強い; A>10>K>Q>J)。
func SchnapsenRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 5
	case 10: // 10
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

// Schnapsen シュナプセンゲームクラス
type Schnapsen struct {
	trumpCards       *TrumpCards
	players          []*SchnapsenPlayer
	config           SchnapsenConfig
	phase            SchnapsenPhase
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*SchnapsenTrickCard
	trumpCard        *Card // 場に表向きで置かれる切り札表示カード (山札の最後)
	trumpSuit        int
	leadPlayerIdx    int
	dealerIdx        int
	playerPoints     []int
	marriageDeclared [CardDesignMax + 1]bool // suit -> 当ラウンドで宣言済か
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定
	actionLog        []*ActionLogEntry
}

// NewSchnapsen コンストラクタ
func NewSchnapsen(trumpCards *TrumpCards, players []*SchnapsenPlayer, config SchnapsenConfig) *Schnapsen {
	return &Schnapsen{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		winnerIdx:    -1,
		playerPoints: make([]int, len(players)),
	}
}

// NewDefaultSchnapsen 標準の 2 人対戦セットアップを返す。
// 人間プレイヤー (idx 0) と CPU (idx 1) の組み合わせ。
func NewDefaultSchnapsen() *Schnapsen {
	players := []*SchnapsenPlayer{
		NewSchnapsenPlayer(true),
		NewSchnapsenPlayer(false),
	}
	return NewSchnapsen(NewTrumpCardsSchnapsen(), players, DefaultSchnapsenConfig())
}

// Reset ゲーム初期化
func (s *Schnapsen) Reset() {
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.trickNumber = 0
	s.currentTrick = nil
	s.leadPlayerIdx = -1
	s.currentPlayerIdx = -1
	s.dealerIdx = 0
	s.playerPoints = make([]int, len(s.players))
	s.marriageDeclared = [CardDesignMax + 1]bool{}
	s.actionLog = nil
	s.trumpCard = nil
	s.trumpSuit = 0

	for _, p := range s.players {
		p.ResetGame()
	}

	s.trumpCards.Shuffle()
	s.dealInitial()
	s.sortAllHands()

	s.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (s *Schnapsen) PlayerPlay(cardIndex int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != SchnapsenPhasePlay {
		return ErrWrongPhase
	}
	if !s.players[s.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := s.players[s.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := s.validatePlay(s.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	s.playCard(s.currentPlayerIdx, played)
	return nil
}

// PlayerDeclareMarriage 人間プレイヤーがマリアージュ (K+Q 同スート) を宣言し、
// 指定したカード (その K か Q) をリードする。リード番でのみ有効。
func (s *Schnapsen) PlayerDeclareMarriage(cardIndex int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != SchnapsenPhasePlay {
		return ErrWrongPhase
	}
	if !s.players[s.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return s.declareMarriage(s.currentPlayerIdx, cardIndex)
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する
func (s *Schnapsen) CpuPlay() {
	if s.gameEndFlag || s.phase != SchnapsenPhasePlay {
		return
	}
	idx := s.currentPlayerIdx
	if s.players[idx].GetIsHuman() {
		return
	}

	// リード番で有益なマリアージュがあれば宣言してリードする
	if len(s.currentTrick) == 0 {
		if cardIdx, ok := s.cpuChooseMarriage(idx); ok {
			_ = s.declareMarriage(idx, cardIdx)
			return
		}
	}

	cardIdx := s.cpuSelectPlayCard(idx)
	played := s.players[idx].RemoveCard(cardIdx)
	s.playCard(idx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (s *Schnapsen) ResolveTrick() {
	if s.phase != SchnapsenPhaseTrickEnd || len(s.currentTrick) != SchnapsenPlayerCnt {
		return
	}

	winnerIdx := s.trickWinner()
	trickCards := make([]*Card, len(s.currentTrick))
	trickPoints := 0
	for i, tc := range s.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += SchnapsenCardPoints(tc.Card)
	}

	s.players[winnerIdx].AddTrick(trickCards)
	s.playerPoints[winnerIdx] += trickPoints

	s.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", s.playerName(winnerIdx), s.trickNumber, trickPoints),
		trickCards)

	s.leadPlayerIdx = winnerIdx

	// 66 点に達したら即ラウンド終了
	if s.playerPoints[winnerIdx] >= SchnapsenWinThreshold {
		s.finishGame()
	}
}

// NextTrick 次のトリックを開始する。山札が残っていれば補充も行う。
// 全カードが尽きたらゲーム終了処理を実行する。
func (s *Schnapsen) NextTrick() {
	if s.phase != SchnapsenPhaseTrickEnd {
		return
	}

	s.drawReplenish()

	if s.allHandsEmpty() {
		s.finishGame()
		return
	}

	s.currentTrick = nil
	s.currentPlayerIdx = s.leadPlayerIdx
	s.trickNumber++
	s.phase = SchnapsenPhasePlay
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (s *Schnapsen) GetPhase() SchnapsenPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Schnapsen) SetPhase(phase SchnapsenPhase) { s.phase = phase }

// GetTrickNumber 現在のトリック番号取得
func (s *Schnapsen) GetTrickNumber() int { return s.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (s *Schnapsen) SetTrickNumber(n int) { s.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (s *Schnapsen) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (s *Schnapsen) SetCurrentPlayerIdx(idx int) { s.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (s *Schnapsen) GetCurrentTrick() []*SchnapsenTrickCard { return s.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (s *Schnapsen) SetCurrentTrick(trick []*SchnapsenTrickCard) { s.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (s *Schnapsen) GetTrumpSuit() int { return s.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (s *Schnapsen) SetTrumpSuit(suit int) { s.trumpSuit = suit }

// GetTrumpCard 場に表向きで置かれている切り札表示カードを取得 (山札に残っていなければ nil)
func (s *Schnapsen) GetTrumpCard() *Card { return s.trumpCard }

// SetTrumpCard 切り札表示カード設定 (テスト用)
func (s *Schnapsen) SetTrumpCard(c *Card) { s.trumpCard = c }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Schnapsen) GetGameEndFlag() bool { return s.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (s *Schnapsen) SetGameEndFlag(flag bool) { s.gameEndFlag = flag }

// GetWinnerIdx 勝者プレイヤーインデックス (-1: 未確定)
func (s *Schnapsen) GetWinnerIdx() int { return s.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (s *Schnapsen) GetPlayerCnt() int { return len(s.players) }

// GetPlayer プレイヤー取得
func (s *Schnapsen) GetPlayer(i int) *SchnapsenPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetPlayerPoints プレイヤーの累積得点取得 (カード点 + マリアージュボーナス)
func (s *Schnapsen) GetPlayerPoints(i int) int {
	if i < 0 || i >= len(s.playerPoints) {
		return 0
	}
	return s.playerPoints[i]
}

// SetPlayerPoints プレイヤー得点設定 (テスト用)
func (s *Schnapsen) SetPlayerPoints(i, points int) {
	if i >= 0 && i < len(s.playerPoints) {
		s.playerPoints[i] = points
	}
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (s *Schnapsen) GetLeadPlayerIdx() int { return s.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (s *Schnapsen) SetLeadPlayerIdx(idx int) { s.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (s *Schnapsen) GetDealerIdx() int { return s.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (s *Schnapsen) SetDealerIdx(idx int) { s.dealerIdx = idx }

// GetStockRemaining 山札の残り枚数 (場に出ている表向き切り札表示カードは含まない;
// それは GetTrumpCard() != nil の間 別カウントとして残る最後の 1 枚)。
func (s *Schnapsen) GetStockRemaining() int {
	return s.trumpCards.GetRemainingCount()
}

// IsEndgame 第2フェーズ (山札と切り札表示カードが尽きてマストフォローになる) かを返す
func (s *Schnapsen) IsEndgame() bool {
	return s.trumpCards.GetRemainingCount() == 0 && s.trumpCard == nil
}

// IsHumanTurn 現在の手番が人間かどうか
func (s *Schnapsen) IsHumanTurn() bool {
	if s.currentPlayerIdx < 0 || s.currentPlayerIdx >= len(s.players) {
		return false
	}
	return s.players[s.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (s *Schnapsen) GetConfig() SchnapsenConfig { return s.config }

// SetConfig 設定変更
func (s *Schnapsen) SetConfig(cfg SchnapsenConfig) { s.config = cfg }

// GetActionLog 棋譜取得
func (s *Schnapsen) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// 第1フェーズは制約なし。第2フェーズはマストフォロー (同スート優先 →
// 勝てるなら勝つ → 無ければ切り札) を適用する。リード番は常に全カード。
func (s *Schnapsen) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	return s.legalPlayIndices(playerIdx)
}

// GetMarriageIndices リード番でマリアージュ宣言を開始できるカード (未宣言スートの
// K または Q で、相方を手札に持つもの) のインデックスを返す。
func (s *Schnapsen) GetMarriageIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	if len(s.currentTrick) != 0 || s.currentPlayerIdx != playerIdx {
		return nil
	}
	p := s.players[playerIdx]
	out := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if s.isMarriageStarter(p, c) {
			out = append(out, i)
		}
	}
	return out
}

// GetHint 人間プレイヤーへのヒントを取得する
func (s *Schnapsen) GetHint() *SchnapsenHint {
	if s.phase != SchnapsenPhasePlay || s.currentPlayerIdx != 0 {
		return nil
	}
	humanIdx := 0
	if s.players[humanIdx].GetCardsSize() == 0 {
		return nil
	}
	if len(s.currentTrick) == 0 {
		if idx, ok := s.cpuChooseMarriage(humanIdx); ok {
			i := idx
			return &SchnapsenHint{CardIndex: &i, Reason: "marriage", IsMarriage: true}
		}
	}
	idx := s.cpuSelectPlayCard(humanIdx)
	return &SchnapsenHint{CardIndex: &idx, Reason: s.playHintReason(humanIdx, idx)}
}

// --- Private methods ---

// dealInitial 各プレイヤーに 5 枚配り、その次の 1 枚を表向きの切り札表示カードとして置く。
func (s *Schnapsen) dealInitial() {
	for range SchnapsenHandSize {
		for i := range SchnapsenPlayerCnt {
			player := s.players[(s.dealerIdx+1+i)%SchnapsenPlayerCnt]
			if c := s.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	s.trumpCard = s.trumpCards.DrawCard()
	if s.trumpCard != nil {
		s.trumpSuit = s.trumpCard.GetDesign()
		s.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(s.trumpCard)), []*Card{s.trumpCard})
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣 (非ディーラー) がリード
func (s *Schnapsen) startPlayPhase() {
	s.trickNumber = 1
	s.currentTrick = nil
	s.leadPlayerIdx = (s.dealerIdx + 1) % SchnapsenPlayerCnt
	s.currentPlayerIdx = s.leadPlayerIdx
	s.phase = SchnapsenPhasePlay
}

// declareMarriage マリアージュを宣言してボーナス加点し、指定の K/Q をリードする共通処理。
func (s *Schnapsen) declareMarriage(playerIdx, cardIndex int) error {
	if len(s.currentTrick) != 0 {
		return NewDomainError(ErrInvalidPlay, "マリアージュはリード時のみ宣言できます")
	}
	player := s.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if !s.isMarriageStarter(player, card) {
		return NewDomainError(ErrInvalidPlay, "そのカードでマリアージュは宣言できません")
	}

	suit := card.GetDesign()
	bonus := SchnapsenMarriageBonus
	if suit == s.trumpSuit {
		bonus = SchnapsenRoyalMarriageBonus
	}
	s.marriageDeclared[suit] = true
	s.playerPoints[playerIdx] += bonus
	s.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares marriage in %s (+%d)", s.playerName(playerIdx), suitStr(suit), bonus), nil)

	// 宣言だけで 66 点に達したら即ラウンド終了 (カードは出さない)
	if s.playerPoints[playerIdx] >= SchnapsenWinThreshold {
		s.finishGame()
		return nil
	}

	played := player.RemoveCard(cardIndex)
	s.playCard(playerIdx, played)
	return nil
}

// isMarriageStarter card が「未宣言スートの K か Q で相方を手札に持つ」かを判定する。
func (s *Schnapsen) isMarriageStarter(player *SchnapsenPlayer, card *Card) bool {
	v := card.GetValue()
	if v != 12 && v != 13 {
		return false
	}
	suit := card.GetDesign()
	if s.marriageDeclared[suit] {
		return false
	}
	partner := 12
	if v == 12 {
		partner = 13
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() == suit && c.GetValue() == partner {
			return true
		}
	}
	return false
}

// playCard カードをプレイする共通処理
func (s *Schnapsen) playCard(playerIdx int, card *Card) {
	s.currentTrick = append(s.currentTrick, &SchnapsenTrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	s.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", s.playerName(playerIdx), cardStr(card)),
		[]*Card{card})

	if len(s.currentTrick) == SchnapsenPlayerCnt {
		s.phase = SchnapsenPhaseTrickEnd
	} else {
		s.currentPlayerIdx = (s.currentPlayerIdx + 1) % SchnapsenPlayerCnt
	}
}

// validatePlay カードのプレイがルール上有効かを検証する。
// 第1フェーズ・リード時は常に有効。第2フェーズの追随時のみマストフォローを課す。
func (s *Schnapsen) validatePlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードが nil です")
	}
	if !s.IsEndgame() || len(s.currentTrick) == 0 {
		return nil
	}
	if !s.cardSatisfiesFollow(playerIdx, card) {
		return NewDomainError(ErrInvalidCard, "第2フェーズではフォロールールに従う必要があります")
	}
	return nil
}

// cardSatisfiesFollow 第2フェーズの追随時に card が合法かを返す。
// 1) リードスートを持つ: そのスートを出す。勝てるカードがあるなら勝つカードのみ可。
// 2) リードスートを持たないが切り札を持つ: 切り札のみ可。
// 3) どちらも持たない: 任意。
func (s *Schnapsen) cardSatisfiesFollow(playerIdx int, card *Card) bool {
	player := s.players[playerIdx]
	leadCard := s.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	if playerHasSuit(player, leadSuit) {
		if card.GetDesign() != leadSuit {
			return false
		}
		// 勝てるカードがあるなら勝たねばならない
		if playerHasSuitWinner(player, leadCard, leadSuit, s.trumpSuit) {
			return schnapsenBeats(card, leadCard, leadSuit, s.trumpSuit)
		}
		return true
	}
	if playerHasSuit(player, s.trumpSuit) {
		return card.GetDesign() == s.trumpSuit
	}
	return true
}

// legalPlayIndices validatePlay を満たすカードのインデックス集合を返す。
func (s *Schnapsen) legalPlayIndices(playerIdx int) []int {
	p := s.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if s.validatePlay(playerIdx, p.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// playerHasSuit プレイヤーが指定スートのカードを持つか
func playerHasSuit(player *SchnapsenPlayer, suit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// playerHasSuitWinner プレイヤーが同スートで leadCard に勝てるカードを持つか
func playerHasSuitWinner(player *SchnapsenPlayer, leadCard *Card, leadSuit, trumpSuit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != leadSuit {
			continue
		}
		if schnapsenBeats(c, leadCard, leadSuit, trumpSuit) {
			return true
		}
	}
	return false
}

// trickWinner 現在のトリックの勝者インデックスを決定する
func (s *Schnapsen) trickWinner() int {
	if len(s.currentTrick) == 0 {
		return 0
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	winnerIdx := s.currentTrick[0].PlayerIdx
	winnerCard := s.currentTrick[0].Card

	for _, tc := range s.currentTrick[1:] {
		if schnapsenBeats(tc.Card, winnerCard, leadSuit, s.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// schnapsenBeats challenger が currentBest に勝つかを判定する。
// ・両者がトランプ: ランクの高い方が勝つ
// ・challenger のみトランプ: challenger が勝つ
// ・両者とも非トランプかつ同じリードスート: ランクの高い方が勝つ
// ・両者とも非トランプで challenger がリードスート以外: challenger は勝てない
func schnapsenBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit

	switch {
	case cIsTrump && bIsTrump:
		return SchnapsenRankOrder(challenger) > SchnapsenRankOrder(currentBest)
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
	return SchnapsenRankOrder(challenger) > SchnapsenRankOrder(currentBest)
}

// drawReplenish トリック勝者が先に 1 枚、次に敗者が 1 枚を山札から引く。
// 山札が空になる過程では、最後の 1 枚は表向き切り札表示カードを引いた扱いになる。
func (s *Schnapsen) drawReplenish() {
	if s.trumpCards.GetRemainingCount() == 0 && s.trumpCard == nil {
		return
	}
	winnerIdx := s.leadPlayerIdx
	loserIdx := (winnerIdx + 1) % SchnapsenPlayerCnt
	for _, idx := range []int{winnerIdx, loserIdx} {
		if c := s.drawOne(); c != nil {
			s.players[idx].AddCard(c)
			s.sortHand(s.players[idx])
		}
	}
}

// drawOne 山札または切り札表示カードから 1 枚引く。優先順位は山札 → 切り札表示カード。
func (s *Schnapsen) drawOne() *Card {
	if c := s.trumpCards.DrawCard(); c != nil {
		return c
	}
	if s.trumpCard != nil {
		c := s.trumpCard
		s.trumpCard = nil
		return c
	}
	return nil
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (s *Schnapsen) allHandsEmpty() bool {
	for _, p := range s.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishGame ゲームを終了させ、勝者を決定する
func (s *Schnapsen) finishGame() {
	if s.gameEndFlag {
		return
	}
	s.gameEndFlag = true
	s.phase = SchnapsenPhaseGameEnd
	s.winnerIdx = s.determineWinner()
	detail := fmt.Sprintf("Game end: %d-%d", s.playerPoints[0], s.playerPoints[1])
	s.appendLog(-1, "game_end", detail, nil)
}

// determineWinner 勝者を決定する。66 点に達した側が勝ち。
// 誰も達していなければ最後のトリックを取った側 (leadPlayerIdx) の勝ち。
func (s *Schnapsen) determineWinner() int {
	switch {
	case s.playerPoints[0] >= SchnapsenWinThreshold:
		return 0
	case s.playerPoints[1] >= SchnapsenWinThreshold:
		return 1
	default:
		return s.leadPlayerIdx
	}
}

// SchnapsenDetermineWinner 66 点ルールでの勝者を返す公開ヘルパー。
// どちらも 66 未満なら lastTrickWinner を勝者とする。
func SchnapsenDetermineWinner(p0, p1, lastTrickWinner int) int {
	switch {
	case p0 >= SchnapsenWinThreshold:
		return 0
	case p1 >= SchnapsenWinThreshold:
		return 1
	default:
		return lastTrickWinner
	}
}

// sortAllHands 全プレイヤーの手札をソートする
func (s *Schnapsen) sortAllHands() {
	for _, p := range s.players {
		s.sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ランク でソートする
func (s *Schnapsen) sortHand(p *SchnapsenPlayer) {
	trumpSuit := s.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return SchnapsenRankOrder(ci) < SchnapsenRankOrder(cj)
	})
}

// playerName プレイヤー名を返す (ログ用)
func (s *Schnapsen) playerName(idx int) string {
	if idx < 0 || idx >= len(s.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if s.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜エントリを追加する
func (s *Schnapsen) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// playHintReason ヒント理由キーを判定する
func (s *Schnapsen) playHintReason(playerIdx, chosenIdx int) string {
	card := s.players[playerIdx].GetCard(chosenIdx)
	pts := SchnapsenCardPoints(card)
	if len(s.currentTrick) == 0 {
		if card.GetDesign() == s.trumpSuit {
			return "lead_trump"
		}
		if pts == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := s.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if schnapsenBeats(card, leadCard, leadSuit, s.trumpSuit) {
		if card.GetDesign() == s.trumpSuit && leadSuit != s.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI (single-difficulty heuristic) ---

// cpuChooseMarriage CPU がリード時に宣言すべきマリアージュのカードインデックスを返す。
// 切り札マリアージュを優先し、宣言時はランクの低い Q をリードする。
func (s *Schnapsen) cpuChooseMarriage(playerIdx int) (int, bool) {
	player := s.players[playerIdx]
	bestIdx := -1
	bestBonus := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetValue() != 12 { // Q を起点に宣言 (低い方をリードして K を温存)
			continue
		}
		if !s.isMarriageStarter(player, c) {
			continue
		}
		bonus := SchnapsenMarriageBonus
		if c.GetDesign() == s.trumpSuit {
			bonus = SchnapsenRoyalMarriageBonus
		}
		if bonus > bestBonus {
			bestBonus = bonus
			bestIdx = i
		}
	}
	return bestIdx, bestIdx >= 0
}

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する (合法手の中から)
func (s *Schnapsen) cpuSelectPlayCard(playerIdx int) int {
	legal := s.legalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return 0
	}
	if len(legal) == 1 {
		return legal[0]
	}

	if len(s.currentTrick) == 0 {
		return s.cpuLead(playerIdx, legal)
	}
	return s.cpuFollow(playerIdx, legal)
}

// cpuLead リード時の選択: 最も低い点数の非トランプを優先する。
func (s *Schnapsen) cpuLead(playerIdx int, legal []int) int {
	player := s.players[playerIdx]
	bestIdx := legal[0]
	bestScore := schnapsenLeadScore(player.GetCard(bestIdx), s.trumpSuit)
	for _, i := range legal[1:] {
		sc := schnapsenLeadScore(player.GetCard(i), s.trumpSuit)
		if sc < bestScore {
			bestScore = sc
			bestIdx = i
		}
	}
	return bestIdx
}

// schnapsenLeadScore 値が小さいほど「リードに適している」(トランプ・高得点札を温存する)
func schnapsenLeadScore(c *Card, trumpSuit int) int {
	score := SchnapsenCardPoints(c)*10 + SchnapsenRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時の選択。合法手のうち勝てる最小コストの札、無ければ最小ダンプ札。
func (s *Schnapsen) cpuFollow(playerIdx int, legal []int) int {
	player := s.players[playerIdx]
	leadCard := s.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	winIdx := -1
	winScore := 0
	dumpIdx := legal[0]
	dumpScoreVal := schnapsenDumpScore(player.GetCard(legal[0]), s.trumpSuit)
	for _, i := range legal {
		c := player.GetCard(i)
		if schnapsenBeats(c, leadCard, leadSuit, s.trumpSuit) {
			sc := schnapsenDumpScore(c, s.trumpSuit)
			if winIdx < 0 || sc < winScore {
				winIdx = i
				winScore = sc
			}
		}
		ds := schnapsenDumpScore(c, s.trumpSuit)
		if ds < dumpScoreVal {
			dumpScoreVal = ds
			dumpIdx = i
		}
	}

	// トリックの得点が高い、または勝てるなら奪取する価値が高い
	if winIdx >= 0 && SchnapsenCardPoints(leadCard) >= 10 {
		return winIdx
	}
	if winIdx >= 0 && leadCard.GetDesign() == s.trumpSuit {
		return winIdx
	}
	// それ以外は最小ダンプ。ただし合法手が勝ち札のみ (第2フェーズの強制勝ち) なら勝つ
	if dumpIdx == winIdx || !s.legalAllowsDump(playerIdx, legal) {
		if winIdx >= 0 {
			return winIdx
		}
	}
	return dumpIdx
}

// legalAllowsDump 合法手の中に「勝たないカード」が含まれるか (= 捨てる自由があるか) を返す。
func (s *Schnapsen) legalAllowsDump(playerIdx int, legal []int) bool {
	player := s.players[playerIdx]
	leadCard := s.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	for _, i := range legal {
		if !schnapsenBeats(player.GetCard(i), leadCard, leadSuit, s.trumpSuit) {
			return true
		}
	}
	return false
}

// schnapsenDumpScore 値が小さいほど「失っても良い」カード
func schnapsenDumpScore(c *Card, trumpSuit int) int {
	score := SchnapsenCardPoints(c)*10 + SchnapsenRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// --- JSON ---

// schnapsenJSON is the JSON wire format for Schnapsen.
type schnapsenJSON struct {
	TrumpCards       *TrumpCards             `json:"tc"`
	Players          []*SchnapsenPlayer      `json:"ps"`
	Config           SchnapsenConfig         `json:"cf"`
	Phase            SchnapsenPhase          `json:"ph"`
	TrickNumber      int                     `json:"tn"`
	CurrentPlayerIdx int                     `json:"ci"`
	CurrentTrick     []*SchnapsenTrickCard   `json:"ct"`
	TrumpCard        *Card                   `json:"tu"`
	TrumpSuit        int                     `json:"ts"`
	LeadPlayerIdx    int                     `json:"li"`
	DealerIdx        int                     `json:"di"`
	PlayerPoints     []int                   `json:"pp"`
	MarriageDeclared [CardDesignMax + 1]bool `json:"md"`
	GameEndFlag      bool                    `json:"ge"`
	WinnerIdx        int                     `json:"wi"`
	ActionLog        []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Schnapsen) MarshalJSON() ([]byte, error) {
	return json.Marshal(schnapsenJSON{
		TrumpCards:       s.trumpCards,
		Players:          s.players,
		Config:           s.config,
		Phase:            s.phase,
		TrickNumber:      s.trickNumber,
		CurrentPlayerIdx: s.currentPlayerIdx,
		CurrentTrick:     s.currentTrick,
		TrumpCard:        s.trumpCard,
		TrumpSuit:        s.trumpSuit,
		LeadPlayerIdx:    s.leadPlayerIdx,
		DealerIdx:        s.dealerIdx,
		PlayerPoints:     s.playerPoints,
		MarriageDeclared: s.marriageDeclared,
		GameEndFlag:      s.gameEndFlag,
		WinnerIdx:        s.winnerIdx,
		ActionLog:        s.actionLog,
	})
}

// schnapsenMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const schnapsenMaxSliceLen = 1000

// errSchnapsenSnapshot is returned for any malformed serialised game state.
// A single shared sentinel (rather than per-field formatted messages) keeps
// UnmarshalJSON's compiled footprint small — the domain package is linked into
// every Cloudflare Worker WASM binary, and the classic worker is at the 1 MB
// gzip free-tier limit.
var errSchnapsenSnapshot = errors.New("schnapsen: invalid serialised game state")

// schnapsenIdxInRange reports whether i is a valid player index.
func schnapsenIdxInRange(i int) bool { return i >= 0 && i < SchnapsenPlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
//
// Validates that the deserialised game state matches Schnapsen's fixed shape
// before adopting it, preventing nil-pointer dereferences and out-of-bounds
// panics (and DoS via oversized slices) from a corrupted or maliciously
// crafted snapshot: exactly SchnapsenPlayerCnt non-nil players, at most
// SchnapsenPlayerCnt non-nil trick cards (each with a non-nil face and an
// in-range PlayerIdx), PlayerPoints aligned to the player count, an ActionLog
// within schnapsenMaxSliceLen with no nil entries, in-range
// current/lead/dealer indices, and a known Phase.
func (s *Schnapsen) UnmarshalJSON(data []byte) error {
	var j schnapsenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != SchnapsenPlayerCnt || len(j.CurrentTrick) > SchnapsenPlayerCnt ||
		(j.PlayerPoints != nil && len(j.PlayerPoints) != SchnapsenPlayerCnt) ||
		len(j.ActionLog) > schnapsenMaxSliceLen ||
		!schnapsenIdxInRange(j.CurrentPlayerIdx) || !schnapsenIdxInRange(j.LeadPlayerIdx) ||
		!schnapsenIdxInRange(j.DealerIdx) ||
		j.Phase < SchnapsenPhasePlay || j.Phase > SchnapsenPhaseGameEnd {
		return errSchnapsenSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errSchnapsenSnapshot
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || !schnapsenIdxInRange(tc.PlayerIdx) {
			return errSchnapsenSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errSchnapsenSnapshot
		}
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCardsSchnapsen()
	}
	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.trickNumber = j.TrickNumber
	s.currentPlayerIdx = j.CurrentPlayerIdx
	s.currentTrick = j.CurrentTrick
	if s.currentTrick == nil {
		s.currentTrick = make([]*SchnapsenTrickCard, 0)
	}
	s.trumpCard = j.TrumpCard
	s.trumpSuit = j.TrumpSuit
	s.leadPlayerIdx = j.LeadPlayerIdx
	s.dealerIdx = j.DealerIdx
	s.playerPoints = j.PlayerPoints
	if s.playerPoints == nil {
		s.playerPoints = make([]int, SchnapsenPlayerCnt)
	}
	s.marriageDeclared = j.MarriageDeclared
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
