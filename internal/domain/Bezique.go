//go:build !js || !wasm || classic

// Package domain ベジーク (Bezique) のドメインモデル。
//
// Bezique はフランス発祥の 2 人用宣言トリックゲームで、ピノクルの直系祖先。64 枚
// (A,7,8,9,10,J,Q,K × 4 スート × 2 セット) のデッキを使う。各プレイヤーに 8 枚配り、
// 次の 1 枚を表向きにして切り札スートを決める。
//
// 第1フェーズ (山札がある間): マストフォロー無しでトリックを行い、勝者は手札から「役 (メルド)」
// を 1 つ宣言して得点できる (結婚=K+Q 同スート 20 点、ロイヤル結婚=切り札 K+Q 40 点、
// ベジーク=♠Q+♦J 40 点、Aヨンマイ 100 点、K/Q/J 各 80/60/40 点)。宣言後に勝者・敗者の
// 順で山札から 1 枚ずつ補充し、勝者がリードする。
//
// 山札が尽きると第2フェーズに移行し、マストフォロー (同スート→勝てれば勝つ→無ければ切り札)
// が義務付けられ、役宣言はできない。最終トリックの勝者に 10 点ボーナス。手札を使い切ったら
// そのディールを集計し、累積が目標点 (既定 1000) に達した側が試合に勝利する。
//
// トリックの強さ: A>10>K>Q>J>9>8>7。切り札は非切り札より常に強い。カード点は
// A=11,10=10,K=4,Q=3,J=2,9/8/7=0。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BeziqueHandSize 各プレイヤーの手札最大枚数 (山札がある間は補充される)
const BeziqueHandSize = 8

// BeziqueLastTrickBonus 第2フェーズ最終トリックのボーナス点
const BeziqueLastTrickBonus = 10

// メルド得点
const (
	// BeziqueMarriagePoints 結婚 (K+Q 同スート、非切り札)
	BeziqueMarriagePoints = 20
	// BeziqueRoyalMarriagePoints ロイヤル結婚 (切り札の K+Q)
	BeziqueRoyalMarriagePoints = 40
	// BeziqueBeziquePoints ベジーク (♠Q + ♦J)
	BeziqueBeziquePoints = 40
	// BeziqueFourAcesPoints A 4枚
	BeziqueFourAcesPoints = 100
	// BeziqueFourKingsPoints K 4枚
	BeziqueFourKingsPoints = 80
	// BeziqueFourQueensPoints Q 4枚
	BeziqueFourQueensPoints = 60
	// BeziqueFourJacksPoints J 4枚
	BeziqueFourJacksPoints = 40
)

// BeziqueMeldType メルド種別
type BeziqueMeldType int

// メルド種別定数 (UI 表示・i18n キー用)
const (
	// BeziqueMeldMarriage 結婚 (suit==trump ならロイヤル結婚)
	BeziqueMeldMarriage BeziqueMeldType = iota
	// BeziqueMeldBezique ベジーク (♠Q + ♦J)
	BeziqueMeldBezique
	// BeziqueMeldFourAces A 4枚
	BeziqueMeldFourAces
	// BeziqueMeldFourKings K 4枚
	BeziqueMeldFourKings
	// BeziqueMeldFourQueens Q 4枚
	BeziqueMeldFourQueens
	// BeziqueMeldFourJacks J 4枚
	BeziqueMeldFourJacks
)

// メルド宣言済みを記録するビットマスクのビット位置。
// CardDesign は 1..4 (Spade..Diamond) なので結婚は bit 1..4 を使う。
// それ以外の役は衝突を避けて bit 5 以降に置く。
const (
	beziqueBitMarriageBase = 0 // suit (1..4) → bit 1..4
	beziqueBitBezique      = 5
	beziqueBitFourAces     = 6
	beziqueBitFourKings    = 7
	beziqueBitFourQueens   = 8
	beziqueBitFourJacks    = 9
)

// BeziquePhase ゲームフェーズ
type BeziquePhase int

// Bezique のフェーズ定数
const (
	// BeziquePhasePlay トリックプレイフェーズ
	BeziquePhasePlay BeziquePhase = iota
	// BeziquePhaseMeld 役宣言フェーズ (トリック勝者が宣言/パスを選ぶ; 山札がある間のみ)
	BeziquePhaseMeld
	// BeziquePhaseRoundEnd ディール終了フェーズ
	BeziquePhaseRoundEnd
	// BeziquePhaseGameEnd ゲーム終了フェーズ
	BeziquePhaseGameEnd
)

// BeziqueMeld 宣言可能な役 (suit は結婚のスート、それ以外は -1)
type BeziqueMeld struct {
	Type   BeziqueMeldType `json:"ty"`
	Suit   int             `json:"su"`
	Points int             `json:"pt"`
}

// BeziqueHint ヒント情報
type BeziqueHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイフェーズ)
	MeldIndex *int   // 推奨メルドインデックス (役宣言フェーズ; -1 = パス推奨)
	Reason    string // ヒント理由キー
}

// BeziqueCardPoints カードのトリック得点を返す (A=11,10=10,K=4,Q=3,J=2; その他=0)。
func BeziqueCardPoints(c *Card) int {
	if c == nil {
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
	default:
		return 0
	}
}

// BeziqueRankOrder カードのスート内順位を返す (大きいほど強い; A>10>K>Q>J>9>8>7)。
func BeziqueRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 8
	case 10: // 10
		return 7
	case 13: // K
		return 6
	case 12: // Q
		return 5
	case 11: // J
		return 4
	default: // 9,8,7
		return c.GetValue() - 6 // 9→3, 8→2, 7→1
	}
}

// Bezique ベジークゲームクラス
type Bezique struct {
	trumpCards       *TrumpCards
	players          []*BeziquePlayer
	config           BeziqueConfig
	phase            BeziquePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpCard        *Card // 場に表向きで置かれる切り札表示カード
	trumpSuit        int
	leadPlayerIdx    int
	dealerIdx        int
	dealPoints       []int // 当ディールの得点
	dealMeldPoints   []int // 当ディールのうちメルド由来の得点 (内訳表示用; トリック由来 = dealPoints - dealMeldPoints)
	matchScore       []int // 試合累積得点
	meldsDeclared    []int // プレイヤー毎の宣言済みメルドビットマスク
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定
	actionLogBase
}

// NewBezique コンストラクタ
func NewBezique(trumpCards *TrumpCards, players []*BeziquePlayer, config BeziqueConfig) *Bezique {
	return &Bezique{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		dealPoints:     make([]int, len(players)),
		dealMeldPoints: make([]int, len(players)),
		matchScore:     make([]int, len(players)),
		meldsDeclared:  make([]int, len(players)),
	}
}

// NewDefaultBezique 標準の 2 人対戦セットアップを返す (人間 idx 0 + CPU idx 1)。
func NewDefaultBezique() *Bezique {
	players := []*BeziquePlayer{
		NewBeziquePlayer(true),
		NewBeziquePlayer(false),
	}
	return NewBezique(NewTrumpCardsBezique(), players, DefaultBeziqueConfig())
}

// Reset ゲーム初期化
func (b *Bezique) Reset() {
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.roundNumber = 1
	b.dealerIdx = 0
	b.matchScore = make([]int, len(b.players))
	b.actionLog = nil

	b.startDeal()
}

// NextRound 次のディールを開始する
func (b *Bezique) NextRound() {
	if b.phase != BeziquePhaseRoundEnd {
		return
	}
	b.roundNumber++
	b.dealerIdx = (b.dealerIdx + 1) % BeziquePlayerCnt
	b.startDeal()
}

// startDeal 1 ディールを開始する: シャッフル・8枚配り・切り札表示。
func (b *Bezique) startDeal() {
	b.dealPoints = make([]int, len(b.players))
	b.dealMeldPoints = make([]int, len(b.players))
	b.meldsDeclared = make([]int, len(b.players))
	b.currentTrick = nil
	b.trumpCard = nil
	b.trumpSuit = 0

	for _, p := range b.players {
		p.ResetGame()
	}

	b.trumpCards.Shuffle()
	for range BeziqueHandSize {
		for i := range BeziquePlayerCnt {
			player := b.players[(b.dealerIdx+1+i)%BeziquePlayerCnt]
			if c := b.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	b.trumpCard = b.trumpCards.DrawCard()
	if b.trumpCard != nil {
		b.trumpSuit = b.trumpCard.GetDesign()
		b.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(b.trumpCard)), []*Card{b.trumpCard})
	}
	b.sortAllHands()

	b.trickNumber = 1
	b.leadPlayerIdx = (b.dealerIdx + 1) % BeziquePlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.phase = BeziquePhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Bezique) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BeziquePhasePlay {
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

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する (プレイフェーズ)
func (b *Bezique) CpuPlay() {
	if b.gameEndFlag || b.phase != BeziquePhasePlay {
		return
	}
	idx := b.currentPlayerIdx
	if b.players[idx].GetIsHuman() {
		return
	}
	cardIdx := b.cpuSelectPlayCard(idx)
	played := b.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	b.playCard(idx, played)
}

// PlayerDeclareMeld 人間プレイヤー (トリック勝者) が役を宣言する。
func (b *Bezique) PlayerDeclareMeld(meldIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BeziquePhaseMeld {
		return ErrWrongPhase
	}
	if !b.players[b.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	melds := b.availableMelds(b.currentPlayerIdx)
	if meldIndex < 0 || meldIndex >= len(melds) {
		return NewDomainError(ErrInvalidPlay, "宣言できる役がありません")
	}
	b.applyMeld(b.currentPlayerIdx, melds[meldIndex])
	b.afterMeld()
	return nil
}

// PlayerSkipMeld 人間プレイヤーが役宣言をパスする。
func (b *Bezique) PlayerSkipMeld() error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BeziquePhaseMeld {
		return ErrWrongPhase
	}
	if !b.players[b.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	b.appendLog(b.currentPlayerIdx, "meld_skip", fmt.Sprintf("%s declares no meld", playerName(b.players, b.currentPlayerIdx)), nil)
	b.afterMeld()
	return nil
}

// CpuMeld 現在の役宣言手番が CPU の場合に最善メルド (なければパス) を実行する。
func (b *Bezique) CpuMeld() {
	if b.gameEndFlag || b.phase != BeziquePhaseMeld {
		return
	}
	idx := b.currentPlayerIdx
	if b.players[idx].GetIsHuman() {
		return
	}
	melds := b.availableMelds(idx)
	if len(melds) == 0 {
		b.appendLog(idx, "meld_skip", fmt.Sprintf("%s declares no meld", playerName(b.players, idx)), nil)
		b.afterMeld()
		return
	}
	best := 0
	for i, m := range melds {
		if m.Points > melds[best].Points {
			best = i
		}
	}
	b.applyMeld(idx, melds[best])
	b.afterMeld()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (b *Bezique) GetPhase() BeziquePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Bezique) SetPhase(phase BeziquePhase) { b.phase = phase }

// GetRoundNumber 現在のディール番号取得
func (b *Bezique) GetRoundNumber() int { return b.roundNumber }

// SetRoundNumber ディール番号設定 (テスト用)
func (b *Bezique) SetRoundNumber(n int) { b.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (b *Bezique) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Bezique) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Bezique) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Bezique) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Bezique) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Bezique) SetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (b *Bezique) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (b *Bezique) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetTrumpCard 場に表向きの切り札表示カードを取得 (山札に残っていなければ nil)
func (b *Bezique) GetTrumpCard() *Card { return b.trumpCard }

// SetTrumpCard 切り札表示カード設定 (テスト用)
func (b *Bezique) SetTrumpCard(c *Card) { b.trumpCard = c }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Bezique) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerIdx 勝者プレイヤーインデックス (-1: 未確定)
func (b *Bezique) GetWinnerIdx() int { return b.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (b *Bezique) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Bezique) GetPlayer(i int) *BeziquePlayer {
	return getPlayer(b.players, i)
}

// GetDealPoints プレイヤーの当ディール得点取得
func (b *Bezique) GetDealPoints(i int) int {
	if i < 0 || i >= len(b.dealPoints) {
		return 0
	}
	return b.dealPoints[i]
}

// GetDealMeldPoints はプレイヤーの当ディール得点のうちメルド由来分を返す。
// トリック由来の得点は GetDealPoints - GetDealMeldPoints で求められる。
func (b *Bezique) GetDealMeldPoints(i int) int {
	if i < 0 || i >= len(b.dealMeldPoints) {
		return 0
	}
	return b.dealMeldPoints[i]
}

// SetDealPoints プレイヤーの当ディール得点設定 (テスト用)
func (b *Bezique) SetDealPoints(i, points int) {
	if i >= 0 && i < len(b.dealPoints) {
		b.dealPoints[i] = points
	}
}

// GetMatchScore プレイヤーの試合累積得点取得
func (b *Bezique) GetMatchScore(i int) int {
	if i < 0 || i >= len(b.matchScore) {
		return 0
	}
	return b.matchScore[i]
}

// SetMatchScore プレイヤーの試合累積得点設定 (テスト用)
func (b *Bezique) SetMatchScore(i, points int) {
	if i >= 0 && i < len(b.matchScore) {
		b.matchScore[i] = points
	}
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Bezique) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Bezique) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Bezique) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Bezique) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetStockRemaining 山札の残り枚数
func (b *Bezique) GetStockRemaining() int { return b.trumpCards.GetRemainingCount() }

// IsEndgame 第2フェーズ (山札と切り札表示カードが尽きてマストフォロー) かを返す
func (b *Bezique) IsEndgame() bool {
	return b.trumpCards.GetRemainingCount() == 0 && b.trumpCard == nil
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Bezique) IsHumanTurn() bool {
	return isHumanTurn(b.players, b.currentPlayerIdx)
}

// GetConfig 設定取得
func (b *Bezique) GetConfig() BeziqueConfig { return b.config }

// SetConfig 設定変更
func (b *Bezique) SetConfig(cfg BeziqueConfig) { b.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (b *Bezique) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	return b.legalPlayIndices(playerIdx)
}

// GetAvailableMelds トリック勝者が宣言できる役の一覧を返す (役宣言フェーズ用)。
func (b *Bezique) GetAvailableMelds(playerIdx int) []BeziqueMeld {
	if b.phase != BeziquePhaseMeld || playerIdx != b.currentPlayerIdx {
		return nil
	}
	return b.availableMelds(playerIdx)
}

// GetHint 人間プレイヤーへのヒントを取得する
func (b *Bezique) GetHint() *BeziqueHint {
	if b.currentPlayerIdx != 0 {
		return nil
	}
	switch b.phase {
	case BeziquePhaseMeld:
		melds := b.availableMelds(0)
		if len(melds) == 0 {
			skip := -1
			return &BeziqueHint{MeldIndex: &skip, Reason: "meld_skip"}
		}
		best := 0
		for i, m := range melds {
			if m.Points > melds[best].Points {
				best = i
			}
		}
		i := best
		return &BeziqueHint{MeldIndex: &i, Reason: "meld_declare"}
	case BeziquePhasePlay:
		if b.players[0].GetCardsSize() == 0 {
			return nil
		}
		idx := b.cpuSelectPlayCard(0)
		return &BeziqueHint{CardIndex: &idx, Reason: b.playHintReason(0, idx)}
	}
	return nil
}

// --- Private methods ---

// playCard カードをプレイする共通処理。2枚出そろったらトリックを解決する。
func (b *Bezique) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	b.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(b.players, playerIdx), cardStr(card)), []*Card{card})
	if len(b.currentTrick) == BeziquePlayerCnt {
		b.resolveTrick()
		return
	}
	b.currentPlayerIdx = (b.currentPlayerIdx + 1) % BeziquePlayerCnt
}

// resolveTrick トリックを解決し、得点付与とフェーズ遷移を行う。
func (b *Bezique) resolveTrick() {
	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	trickPoints := 0
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += BeziqueCardPoints(tc.Card)
	}
	b.players[winnerIdx].AddTrick(trickCards)
	b.dealPoints[winnerIdx] += trickPoints
	b.leadPlayerIdx = winnerIdx
	b.currentPlayerIdx = winnerIdx
	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", playerName(b.players, winnerIdx), b.trickNumber, trickPoints), trickCards)

	if b.IsEndgame() {
		// 第2フェーズ: 役宣言・補充なし。
		if b.allHandsEmpty() {
			b.dealPoints[winnerIdx] += BeziqueLastTrickBonus
			b.appendLog(winnerIdx, "last_trick",
				fmt.Sprintf("%s takes the last trick (+%d)", playerName(b.players, winnerIdx), BeziqueLastTrickBonus), nil)
			b.scoreDeal()
			return
		}
		b.currentTrick = nil
		b.trickNumber++
		b.phase = BeziquePhasePlay
		return
	}
	// 第1フェーズ: 勝者が役を宣言できる。
	b.phase = BeziquePhaseMeld
}

// afterMeld 役宣言/パス後の補充・次トリック準備。
func (b *Bezique) afterMeld() {
	b.drawReplenish()
	if b.allHandsEmpty() {
		b.scoreDeal()
		return
	}
	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = BeziquePhasePlay
}

// drawReplenish 勝者→敗者の順に山札 (尽きたら切り札表示カード) から 1 枚ずつ引く。
func (b *Bezique) drawReplenish() {
	if b.trumpCards.GetRemainingCount() == 0 && b.trumpCard == nil {
		return
	}
	winnerIdx := b.leadPlayerIdx
	loserIdx := (winnerIdx + 1) % BeziquePlayerCnt
	for _, idx := range []int{winnerIdx, loserIdx} {
		if c := b.drawOne(); c != nil {
			b.players[idx].AddCard(c)
			b.sortHand(b.players[idx])
		}
	}
}

// drawOne 山札 → 切り札表示カード の順に 1 枚引く。
func (b *Bezique) drawOne() *Card {
	return drawOrTakeTrump(b.trumpCards, &b.trumpCard)
}

// scoreDeal ディールを集計して累積し、ゲーム終了を判定する。
func (b *Bezique) scoreDeal() {
	for i := range b.players {
		b.matchScore[i] += b.dealPoints[i]
		b.players[i].SetRoundScore(b.dealPoints[i])
		b.players[i].CommitRoundScore()
	}
	b.appendLog(-1, "deal_score",
		fmt.Sprintf("Deal %d: %d-%d (match %d-%d)", b.roundNumber,
			b.dealPoints[0], b.dealPoints[1], b.matchScore[0], b.matchScore[1]), nil)

	if b.matchScore[0] >= b.config.TargetScore || b.matchScore[1] >= b.config.TargetScore {
		b.finishGame()
		return
	}
	b.phase = BeziquePhaseRoundEnd
}

// finishGame ゲームを終了させ勝者を決定する。
func (b *Bezique) finishGame() {
	b.gameEndFlag = true
	b.phase = BeziquePhaseGameEnd
	switch {
	case b.matchScore[0] > b.matchScore[1]:
		b.winnerIdx = 0
	case b.matchScore[1] > b.matchScore[0]:
		b.winnerIdx = 1
	default:
		b.winnerIdx = b.leadPlayerIdx
	}
	b.appendLog(-1, "game_end", fmt.Sprintf("Game end: %d-%d", b.matchScore[0], b.matchScore[1]), nil)
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (b *Bezique) allHandsEmpty() bool {
	return allHandsEmpty(b.players)
}

// validatePlay カードのプレイがルール上有効かを検証する。
func (b *Bezique) validatePlay(playerIdx int, card *Card) error {
	return validateEndgameFollow(b.currentTrick, b, playerIdx, card)
}

// cardSatisfiesFollow 第2フェーズの追随時に card が合法かを返す。
func (b *Bezique) cardSatisfiesFollow(playerIdx int, card *Card) bool {
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	if beziquePlayerHasSuit(player, leadSuit) {
		if card.GetDesign() != leadSuit {
			return false
		}
		if beziquePlayerHasSuitWinner(player, leadCard, leadSuit, b.trumpSuit) {
			return beziqueBeats(card, leadCard, leadSuit, b.trumpSuit)
		}
		return true
	}
	if beziquePlayerHasSuit(player, b.trumpSuit) {
		return card.GetDesign() == b.trumpSuit
	}
	return true
}

// legalPlayIndices validatePlay を満たすカードのインデックス集合を返す。
func (b *Bezique) legalPlayIndices(playerIdx int) []int {
	return validPlayIndices(b.players[playerIdx], func(c *Card) bool { return b.validatePlay(playerIdx, c) == nil })
}

// beziquePlayerHasSuit プレイヤーが指定スートのカードを持つか
func beziquePlayerHasSuit(player *BeziquePlayer, suit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// beziquePlayerHasSuitWinner プレイヤーが同スートで leadCard に勝てるカードを持つか
func beziquePlayerHasSuitWinner(player *BeziquePlayer, leadCard *Card, leadSuit, trumpSuit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != leadSuit {
			continue
		}
		if beziqueBeats(c, leadCard, leadSuit, trumpSuit) {
			return true
		}
	}
	return false
}

// trickWinner 現在のトリックの勝者インデックスを決定する
func (b *Bezique) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerCard := b.currentTrick[0].Card
	for _, tc := range b.currentTrick[1:] {
		if beziqueBeats(tc.Card, winnerCard, leadSuit, b.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// beziqueBeats challenger が currentBest に勝つかを判定する。
// 同位 (同ランク同スートの 2 枚目) はリードした currentBest が勝つ (challenger は勝てない)。
func beziqueBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit
	switch {
	case cIsTrump && bIsTrump:
		return BeziqueRankOrder(challenger) > BeziqueRankOrder(currentBest)
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
	return BeziqueRankOrder(challenger) > BeziqueRankOrder(currentBest)
}

// --- Melds ---

// availableMelds playerIdx が現在の手札で宣言できる (未宣言の) 役の一覧を返す。
func (b *Bezique) availableMelds(playerIdx int) []BeziqueMeld {
	p := b.players[playerIdx]
	declared := b.meldsDeclared[playerIdx]
	out := make([]BeziqueMeld, 0)

	// 結婚 (K+Q 同スート)
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if declared&(1<<(beziqueBitMarriageBase+suit)) != 0 {
			continue
		}
		if beziqueHandHas(p, suit, 13) && beziqueHandHas(p, suit, 12) {
			pts := BeziqueMarriagePoints
			if suit == b.trumpSuit {
				pts = BeziqueRoyalMarriagePoints
			}
			out = append(out, BeziqueMeld{Type: BeziqueMeldMarriage, Suit: suit, Points: pts})
		}
	}
	// ベジーク (♠Q + ♦J)
	if declared&(1<<beziqueBitBezique) == 0 &&
		beziqueHandHas(p, CardDesignSpade, 12) && beziqueHandHas(p, CardDesignDiamond, 11) {
		out = append(out, BeziqueMeld{Type: BeziqueMeldBezique, Suit: -1, Points: BeziqueBeziquePoints})
	}
	// 4枚役
	type fourSpec struct {
		bit   int
		value int
		mtype BeziqueMeldType
		pts   int
	}
	fours := []fourSpec{
		{beziqueBitFourAces, 1, BeziqueMeldFourAces, BeziqueFourAcesPoints},
		{beziqueBitFourKings, 13, BeziqueMeldFourKings, BeziqueFourKingsPoints},
		{beziqueBitFourQueens, 12, BeziqueMeldFourQueens, BeziqueFourQueensPoints},
		{beziqueBitFourJacks, 11, BeziqueMeldFourJacks, BeziqueFourJacksPoints},
	}
	for _, f := range fours {
		if declared&(1<<f.bit) != 0 {
			continue
		}
		if beziqueCountValue(p, f.value) >= 4 {
			out = append(out, BeziqueMeld{Type: f.mtype, Suit: -1, Points: f.pts})
		}
	}
	return out
}

// applyMeld 役を宣言済みにマークし得点を加算する。
func (b *Bezique) applyMeld(playerIdx int, m BeziqueMeld) {
	bit := beziqueMeldBit(m)
	b.meldsDeclared[playerIdx] |= 1 << bit
	b.dealPoints[playerIdx] += m.Points
	b.dealMeldPoints[playerIdx] += m.Points
	b.appendLog(playerIdx, "meld",
		fmt.Sprintf("%s declares %s (+%d)", playerName(b.players, playerIdx), beziqueMeldName(m), m.Points), nil)
}

// beziqueMeldBit メルドの宣言済みビット位置を返す。
func beziqueMeldBit(m BeziqueMeld) int {
	switch m.Type {
	case BeziqueMeldMarriage:
		return beziqueBitMarriageBase + m.Suit
	case BeziqueMeldBezique:
		return beziqueBitBezique
	case BeziqueMeldFourAces:
		return beziqueBitFourAces
	case BeziqueMeldFourKings:
		return beziqueBitFourKings
	case BeziqueMeldFourQueens:
		return beziqueBitFourQueens
	default: // BeziqueMeldFourJacks
		return beziqueBitFourJacks
	}
}

// beziqueMeldName メルドのログ表示名。
func beziqueMeldName(m BeziqueMeld) string {
	switch m.Type {
	case BeziqueMeldMarriage:
		if m.Points == BeziqueRoyalMarriagePoints {
			return "Royal Marriage"
		}
		return "Marriage (" + suitStr(m.Suit) + ")"
	case BeziqueMeldBezique:
		return "Bezique"
	case BeziqueMeldFourAces:
		return "Four Aces"
	case BeziqueMeldFourKings:
		return "Four Kings"
	case BeziqueMeldFourQueens:
		return "Four Queens"
	default:
		return "Four Jacks"
	}
}

// beziqueHandHas プレイヤーが指定スート・値のカードを持つか
func beziqueHandHas(p *BeziquePlayer, suit, value int) bool {
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == suit && c.GetValue() == value {
			return true
		}
	}
	return false
}

// beziqueCountValue プレイヤーの手札にある指定値のカード枚数
func beziqueCountValue(p *BeziquePlayer, value int) int {
	cnt := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == value {
			cnt++
		}
	}
	return cnt
}

// --- Sorting / helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (b *Bezique) sortAllHands() {
	sortEachHand(b.players, b.sortHand)
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ランク でソートする
func (b *Bezique) sortHand(p *BeziquePlayer) {
	trumpSuit := b.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return BeziqueRankOrder(ci) < BeziqueRankOrder(cj)
	})
}

// playHintReason ヒント理由キーを判定する
func (b *Bezique) playHintReason(playerIdx, chosenIdx int) string {
	card := b.players[playerIdx].GetCard(chosenIdx)
	if len(b.currentTrick) == 0 {
		if card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		if BeziqueCardPoints(card) == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if beziqueBeats(card, leadCard, leadSuit, b.trumpSuit) {
		if card.GetDesign() == b.trumpSuit && leadSuit != b.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI ---

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する (合法手の中から)
func (b *Bezique) cpuSelectPlayCard(playerIdx int) int {
	legal := b.legalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return 0
	}
	if len(legal) == 1 {
		return legal[0]
	}
	if len(b.currentTrick) == 0 {
		return b.cpuLead(playerIdx, legal)
	}
	return b.cpuFollow(playerIdx, legal)
}

// cpuLead リード時: 役に使えるカード (K/Q/A/J) とトランプを温存し、最も安いカードを出す。
func (b *Bezique) cpuLead(playerIdx int, legal []int) int {
	player := b.players[playerIdx]
	bestIdx := legal[0]
	bestScore := beziqueKeepScore(player.GetCard(bestIdx), b.trumpSuit)
	for _, i := range legal[1:] {
		sc := beziqueKeepScore(player.GetCard(i), b.trumpSuit)
		if sc < bestScore {
			bestScore = sc
			bestIdx = i
		}
	}
	return bestIdx
}

// beziqueKeepScore 値が小さいほど「手放してよい」(トランプ・高得点・役札を温存)。
func beziqueKeepScore(c *Card, trumpSuit int) int {
	score := BeziqueCardPoints(c)*10 + BeziqueRankOrder(c)
	// 役に使える札 (K,Q,J,A) を温存
	switch c.GetValue() {
	case 1, 11, 12, 13:
		score += 50
	}
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時: 高得点トリックや切り札リードは勝ちに行き、それ以外は最小ダンプ。
func (b *Bezique) cpuFollow(playerIdx int, legal []int) int {
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	winIdx := -1
	winScore := 0
	dumpIdx := legal[0]
	dumpScore := beziqueKeepScore(player.GetCard(legal[0]), b.trumpSuit)
	for _, i := range legal {
		c := player.GetCard(i)
		if beziqueBeats(c, leadCard, leadSuit, b.trumpSuit) {
			sc := beziqueKeepScore(c, b.trumpSuit)
			if winIdx < 0 || sc < winScore {
				winIdx = i
				winScore = sc
			}
		}
		ds := beziqueKeepScore(c, b.trumpSuit)
		if ds < dumpScore {
			dumpScore = ds
			dumpIdx = i
		}
	}
	if winIdx >= 0 && (BeziqueCardPoints(leadCard) >= 10 || leadCard.GetDesign() == b.trumpSuit) {
		return winIdx
	}
	if !b.legalAllowsDump(playerIdx, legal) && winIdx >= 0 {
		return winIdx
	}
	return dumpIdx
}

// legalAllowsDump 合法手の中に「勝たないカード」が含まれるか (= 捨てる自由があるか)。
func (b *Bezique) legalAllowsDump(playerIdx int, legal []int) bool {
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	for _, i := range legal {
		if !beziqueBeats(player.GetCard(i), leadCard, leadSuit, b.trumpSuit) {
			return true
		}
	}
	return false
}

// --- JSON ---

// beziqueJSON is the JSON wire format for Bezique.
type beziqueJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*BeziquePlayer  `json:"ps"`
	Config           BeziqueConfig     `json:"cf"`
	Phase            BeziquePhase      `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	TrumpCard        *Card             `json:"tu"`
	TrumpSuit        int               `json:"ts"`
	LeadPlayerIdx    int               `json:"li"`
	DealerIdx        int               `json:"di"`
	DealPoints       []int             `json:"dp"`
	DealMeldPoints   []int             `json:"dmp"`
	MatchScore       []int             `json:"ms"`
	MeldsDeclared    []int             `json:"me"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Bezique) MarshalJSON() ([]byte, error) {
	return json.Marshal(beziqueJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		RoundNumber:      b.roundNumber,
		TrickNumber:      b.trickNumber,
		CurrentPlayerIdx: b.currentPlayerIdx,
		CurrentTrick:     b.currentTrick,
		TrumpCard:        b.trumpCard,
		TrumpSuit:        b.trumpSuit,
		LeadPlayerIdx:    b.leadPlayerIdx,
		DealerIdx:        b.dealerIdx,
		DealPoints:       b.dealPoints,
		DealMeldPoints:   b.dealMeldPoints,
		MatchScore:       b.matchScore,
		MeldsDeclared:    b.meldsDeclared,
		GameEndFlag:      b.gameEndFlag,
		WinnerIdx:        b.winnerIdx,
		ActionLog:        b.actionLog,
	})
}

// beziqueMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const beziqueMaxSliceLen = 5000

// errBeziqueSnapshot is the single shared sentinel for any malformed serialised
// game state (kept compact: the domain package links into every Worker WASM).
var errBeziqueSnapshot = errors.New("bezique: invalid serialised game state")

// beziqueIdxInRange reports whether i is a valid player index.
func beziqueIdxInRange(i int) bool { return i >= 0 && i < BeziquePlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bezique) UnmarshalJSON(data []byte) error {
	var j beziqueJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != BeziquePlayerCnt || len(j.CurrentTrick) > BeziquePlayerCnt ||
		(j.DealPoints != nil && len(j.DealPoints) != BeziquePlayerCnt) ||
		(j.MatchScore != nil && len(j.MatchScore) != BeziquePlayerCnt) ||
		(j.MeldsDeclared != nil && len(j.MeldsDeclared) != BeziquePlayerCnt) ||
		len(j.ActionLog) > beziqueMaxSliceLen ||
		!beziqueIdxInRange(j.CurrentPlayerIdx) || !beziqueIdxInRange(j.LeadPlayerIdx) ||
		!beziqueIdxInRange(j.DealerIdx) ||
		j.TrumpSuit < 0 || j.TrumpSuit > CardDesignDiamond ||
		j.RoundNumber < 1 ||
		j.Phase < BeziquePhasePlay || j.Phase > BeziquePhaseGameEnd {
		return errBeziqueSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errBeziqueSnapshot
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || !beziqueIdxInRange(tc.PlayerIdx) {
			return errBeziqueSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errBeziqueSnapshot
		}
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCardsBezique()
	}
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.roundNumber = j.RoundNumber
	b.trickNumber = j.TrickNumber
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.currentTrick = j.CurrentTrick
	if b.currentTrick == nil {
		b.currentTrick = make([]*TrickCard, 0)
	}
	b.trumpCard = j.TrumpCard
	b.trumpSuit = j.TrumpSuit
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.dealerIdx = j.DealerIdx
	b.dealPoints = beziqueEnsureLen(j.DealPoints)
	b.dealMeldPoints = beziqueEnsureLen(j.DealMeldPoints)
	b.matchScore = beziqueEnsureLen(j.MatchScore)
	b.meldsDeclared = beziqueEnsureLen(j.MeldsDeclared)
	b.gameEndFlag = j.GameEndFlag
	b.winnerIdx = j.WinnerIdx
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// beziqueEnsureLen returns s if it already has BeziquePlayerCnt entries, else a
// fresh zeroed slice of the right length.
func beziqueEnsureLen(s []int) []int {
	if len(s) == BeziquePlayerCnt {
		return s
	}
	return make([]int, BeziquePlayerCnt)
}
