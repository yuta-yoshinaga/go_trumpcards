//go:build !js || !wasm || classic

// Package domain ブリュスカンビーユ (Brusquembille) のドメインモデル。
//
// Brusquembille はイタリアの古典的なトリックテイキングゲーム。40枚デッキ
// (8,9,10 を除く) を使い、本実装では 2 人対戦のみを扱う。最大の特徴は
// 「リードスートに従う義務 (must-follow) がない」ことで、validatePlay は
// 常に成功する。トリックの勝者はトランプ (brusquembille) > リードスートの
// ブリュスカンビーユ順位 で決まり、ブリュスカンビーユ順位は A>3>K>Q>J>7>6>5>4>2 となる。
// 各カードには独自の点数 (A=11, 3=10, K=4, Q=3, J=2, それ以外=0) が
// あり、合計 120 点を 2 人で取り合う。60 点を超えた側が勝者で、
// 60-60 は引き分け。
package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ブリュスカンビーユの席数。**2〜5 人**で遊べる。
//
// クローン元のブリスコラは 2 人固定 (`BriscolaPlayerCnt = 2`) で、トリックの
// 長さも補充の順番もその前提で書かれていた。ここは席数を可変にしてある。
const (
	// BrusquembilleMinPlayerCnt 最小プレイヤー数
	BrusquembilleMinPlayerCnt = 2
	// BrusquembilleMaxPlayerCnt 最大プレイヤー数
	BrusquembilleMaxPlayerCnt = 5
	// BrusquembilleDefaultPlayerCnt 既定のプレイヤー数
	BrusquembilleDefaultPlayerCnt = 2
)

// BrusquembilleHandSize 各プレイヤーの手札最大枚数 (山札がある間は補充される)
const BrusquembilleHandSize = 3

// BrusquembilleWinThreshold 勝利点 (これを超える点数で勝ち、超えない側は負け)
const BrusquembilleWinThreshold = 60

// BrusquembilleTotalPoints デッキ全体の合計点
const BrusquembilleTotalPoints = 120

// BrusquembillePhase ゲームフェーズ
type BrusquembillePhase int

// Brusquembilleのフェーズ定数
const (
	// BrusquembillePhasePlay トリックプレイフェーズ
	BrusquembillePhasePlay BrusquembillePhase = iota
	// BrusquembillePhaseTrickEnd トリック終了フェーズ
	BrusquembillePhaseTrickEnd
	// BrusquembillePhaseGameEnd ゲーム終了フェーズ
	BrusquembillePhaseGameEnd
)

// BrusquembilleHint ヒント情報
type BrusquembilleHint struct {
	CardIndex *int   // 推奨カードインデックス
	Reason    string // ヒント理由キー
}

// brusquembilleCardPoints ブリュスカンビーユのカード点数。
//
// **3 ではなく 10 が 10 点。** クローン元のブリスコラは 40 枚のイタリア式
// デッキで「3」が 10 点だが、ブリュスカンビーユは 32 枚のフランス式
// (7-8-9-10-J-Q-K-A) で **3 という札が存在しない**。表をそのまま持ってくると
// 10 点ぶんが盤から消え、合計点が 120 ではなく 80 になる。
var brusquembilleCardPoints = map[int]int{
	1:  11, // As
	10: 10, // Dix
	13: 4,  // Roi
	12: 3,  // Dame
	11: 2,  // Valet
}

// brusquembilleRankOrder スート内のカード強さ。値が大きいほど強い。
// A>10>K>Q>J>9>8>7 を 1-base で表現する。
//
// **10 を必ず入れる。** クローン元の表は 2/4/5/6 という 32 枚デッキに無い札を
// 並べる一方で **10 を一度も挙げていない**。そのまま使うと、10 は既定値 0 に
// 落ちて**盤で一番弱い札**になる —— このゲームでは A に次ぐ 2 番目に強い札
// なので、勝ち負けが逆さまになる。しかも点数表では 10 点の札なので、
// 「一番大事な札が一番弱い」という壊れ方をする。
var brusquembilleRankOrder = map[int]int{
	7:  1,
	8:  2,
	9:  3,
	11: 4, // Valet
	12: 5, // Dame
	13: 6, // Roi
	10: 7, // Dix
	1:  8, // As
}

// BrusquembilleCardPoints カードの得点を返す (公開ヘルパー)。
func BrusquembilleCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	return brusquembilleCardPoints[c.GetValue()]
}

// BrusquembilleRankOrder カードのスート内順位を返す (大きいほど強い)。
func BrusquembilleRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	return brusquembilleRankOrder[c.GetValue()]
}

// Brusquembille ブリュスカンビーユゲームクラス
type Brusquembille struct {
	trumpCards       *TrumpCards
	players          []*BrusquembillePlayer
	config           BrusquembilleConfig
	phase            BrusquembillePhase
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpCard        *Card // 場に表向きで置かれるトランプ (山札の最後)
	trumpSuit        int
	leadPlayerIdx    int
	dealerIdx        int
	playerPoints     []int
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定または引き分け
	actionLogBase
}

// NewBrusquembille コンストラクタ
func NewBrusquembille(trumpCards *TrumpCards, players []*BrusquembillePlayer, config BrusquembilleConfig) *Brusquembille {
	return &Brusquembille{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		winnerIdx:    -1,
		playerPoints: make([]int, len(players)),
	}
}

// NewDefaultBrusquembille 標準の 2 人対戦セットアップを返す。
// 人間プレイヤー (idx 0) と CPU (idx 1) の組み合わせ。
func NewDefaultBrusquembille() *Brusquembille {
	cfg := DefaultBrusquembilleConfig()
	return NewBrusquembille(NewTrumpCards32(), NewBrusquembillePlayersForTable(cfg.PlayerCnt), cfg)
}

// NewBrusquembillePlayersForTable は席数ぶんのプレイヤーを作る (席 0 が人間)。
// 範囲外は既定値に丸める。
func NewBrusquembillePlayersForTable(cnt int) []*BrusquembillePlayer {
	if cnt < BrusquembilleMinPlayerCnt || cnt > BrusquembilleMaxPlayerCnt {
		cnt = BrusquembilleDefaultPlayerCnt
	}
	players := make([]*BrusquembillePlayer, 0, cnt)
	players = append(players, NewBrusquembillePlayer(true))
	for i := 1; i < cnt; i++ {
		players = append(players, NewBrusquembillePlayer(false))
	}
	return players
}

// Reset ゲーム初期化
func (b *Brusquembille) Reset() {
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.trickNumber = 0
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.currentPlayerIdx = -1
	b.dealerIdx = 0
	b.playerPoints = make([]int, len(b.players))
	b.actionLog = nil
	b.trumpCard = nil
	b.trumpSuit = 0

	for _, p := range b.players {
		p.ResetGame()
	}

	b.trumpCards.Shuffle()
	b.stripForTableSize()
	b.dealInitial()
	b.sortAllHands()

	b.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Brusquembille) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BrusquembillePhasePlay {
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

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する
func (b *Brusquembille) CpuPlay() {
	if b.gameEndFlag || b.phase != BrusquembillePhasePlay {
		return
	}
	if b.players[b.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := b.players[b.currentPlayerIdx]
	cardIdx := b.cpuSelectPlayCard(b.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	b.playCard(b.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (b *Brusquembille) ResolveTrick() {
	if b.phase != BrusquembillePhaseTrickEnd || len(b.currentTrick) != len(b.players) {
		return
	}

	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	trickPoints := 0
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += BrusquembilleCardPoints(tc.Card)
	}

	b.players[winnerIdx].AddTrick(trickCards)
	b.playerPoints[winnerIdx] += trickPoints

	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", playerName(b.players, winnerIdx), b.trickNumber, trickPoints),
		trickCards)

	b.leadPlayerIdx = winnerIdx
	// Phase is already BrusquembillePhaseTrickEnd (guarded at function entry); leave it.
}

// NextTrick 次のトリックを開始する。山札が残っていれば補充も行う。
// 全カードが尽きたらゲーム終了処理を実行する。
func (b *Brusquembille) NextTrick() {
	if b.phase != BrusquembillePhaseTrickEnd {
		return
	}

	b.drawReplenish()

	if b.allHandsEmpty() {
		b.finishGame()
		return
	}

	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = BrusquembillePhasePlay
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (b *Brusquembille) GetPhase() BrusquembillePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Brusquembille) SetPhase(phase BrusquembillePhase) { b.phase = phase }

// GetTrickNumber 現在のトリック番号取得
func (b *Brusquembille) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Brusquembille) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Brusquembille) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Brusquembille) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Brusquembille) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Brusquembille) SetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (b *Brusquembille) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (b *Brusquembille) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetTrumpCard 場に表向きで置かれているトランプカードを取得 (山札に残っていなければ nil)
func (b *Brusquembille) GetTrumpCard() *Card { return b.trumpCard }

// SetTrumpCard トランプカード設定 (テスト用)
func (b *Brusquembille) SetTrumpCard(c *Card) { b.trumpCard = c }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Brusquembille) GetGameEndFlag() bool { return b.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (b *Brusquembille) SetGameEndFlag(flag bool) { b.gameEndFlag = flag }

// GetWinnerIdx 勝者プレイヤーインデックス (-1: 未確定または引き分け)
func (b *Brusquembille) GetWinnerIdx() int { return b.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (b *Brusquembille) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Brusquembille) GetPlayer(i int) *BrusquembillePlayer {
	return getPlayer(b.players, i)
}

// GetPlayerPoints プレイヤーの累積得点取得
func (b *Brusquembille) GetPlayerPoints(i int) int {
	if i < 0 || i >= len(b.playerPoints) {
		return 0
	}
	return b.playerPoints[i]
}

// SetPlayerPoints プレイヤー得点設定 (テスト用)
func (b *Brusquembille) SetPlayerPoints(i, points int) {
	if i >= 0 && i < len(b.playerPoints) {
		b.playerPoints[i] = points
	}
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Brusquembille) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Brusquembille) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Brusquembille) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Brusquembille) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetStockRemaining 山札の残り枚数 (場に出ている表向きトランプは含まない;
// それは GetTrumpCard() != nil の間 別カウントとして残る最後の 1 枚)。
func (b *Brusquembille) GetStockRemaining() int {
	return b.trumpCards.GetRemainingCount()
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Brusquembille) IsHumanTurn() bool {
	return isHumanTurn(b.players, b.currentPlayerIdx)
}

// GetConfig 設定取得
func (b *Brusquembille) GetConfig() BrusquembilleConfig { return b.config }

// SetConfig 設定変更
// SetConfig は設定を差し替える。
//
// **席数が変わったら卓を組み直す。** 代入するだけだと、config は 4 人卓と
// 言っているのに b.players は 2 人のままで、Reset しても 2 人卓が始まる ——
// 設定が効いていないのに効いたように見える一番たちの悪い形。
func (b *Brusquembille) SetConfig(cfg BrusquembilleConfig) {
	b.config = cfg
	if cfg.PlayerCnt > 0 && cfg.PlayerCnt != len(b.players) {
		b.players = NewBrusquembillePlayersForTable(cfg.PlayerCnt)
		b.playerPoints = make([]int, len(b.players))
	}
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// 前半 (山札が残っている) は手札全てが対象。**山札が尽きた後半は
// リードスートに追従できるならその札だけ**が対象になる。
//
// validatePlay と同じ判断をここでも書くのではなく、validatePlay に問い合わせる。
// 二重に持つと、片方だけ直したときに「UI では選べるのにエラーになる」
// (あるいはその逆) が起きる。
func (b *Brusquembille) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	p := b.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if b.validatePlay(playerIdx, p.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// GetHint 人間プレイヤーへのヒントを取得する
func (b *Brusquembille) GetHint() *BrusquembilleHint {
	if b.phase != BrusquembillePhasePlay || b.currentPlayerIdx != 0 {
		return nil
	}
	humanIdx := 0
	if b.players[humanIdx].GetCardsSize() == 0 {
		return nil
	}
	idx := b.cpuSelectPlayCard(humanIdx)
	return &BrusquembilleHint{CardIndex: &idx, Reason: b.playHintReason(humanIdx, idx)}
}

// --- Private methods ---

// stripForTableSize は、席数で割り切れるまで山札から最も弱い札を抜く。
//
// **32 は 3 でも 5 でも割り切れない。** 3 人卓なら 32-9(配り)-1(切札)=22 で
// 補充は 23 回、5 人卓なら 17 回。どちらも席数で割れないので、手札の減り方が
// 揃わず、最後に誰かの手札だけ残って**打ち切れない**。実測で 3 人卓と 5 人卓が
// そのまま止まった。
//
// 歴史的にもこの理由で低い札を抜いて調整する。抜くのは最弱の 7 から順に、
// `32 % 席数` 枚 —— 2 人・4 人卓では 1 枚も抜かない。
func (b *Brusquembille) stripForTableSize() {
	surplus := b.trumpCards.GetRemainingCount() % len(b.players)
	// 最弱の 7 から順に抜く。7 を使い切ったら 8 …と上げていく
	// (32 枚を 2〜5 人で割るのに 4 枚を超えて抜くことは無いので、実際には 7 だけ)。
	for _, v := range []int{7, 8, 9} {
		if surplus <= 0 {
			break
		}
		surplus -= b.trumpCards.RemoveCardsByValue(v, surplus)
	}
	surplus = b.trumpCards.GetRemainingCount() % len(b.players)
	if surplus == 0 {
		b.appendLog(-1, "strip", fmt.Sprintf("deck trimmed to %d so it divides by %d seats",
			b.trumpCards.GetRemainingCount(), len(b.players)), nil)
	}
}

// dealInitial 各プレイヤーに 3 枚配り、その次の 1 枚を表向きトランプとして山札の底に置く。
func (b *Brusquembille) dealInitial() {
	for range BrusquembilleHandSize {
		for i := range b.players {
			player := b.players[(b.dealerIdx+1+i)%len(b.players)]
			if c := b.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	// 次の 1 枚をトランプとして表向きに置く (デッキの底相当: 最後に引かれる)
	b.trumpCard = b.trumpCards.DrawCard()
	if b.trumpCard != nil {
		b.trumpSuit = b.trumpCard.GetDesign()
		b.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(b.trumpCard)), []*Card{b.trumpCard})
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (b *Brusquembille) startPlayPhase() {
	b.trickNumber = 1
	b.currentTrick = nil
	b.leadPlayerIdx = (b.dealerIdx + 1) % len(b.players)
	b.currentPlayerIdx = b.leadPlayerIdx
	b.phase = BrusquembillePhasePlay
}

// playCard カードをプレイする共通処理
func (b *Brusquembille) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	b.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(b.players, playerIdx), cardStr(card)),
		[]*Card{card})

	if len(b.currentTrick) == len(b.players) {
		b.phase = BrusquembillePhaseTrickEnd
	} else {
		b.currentPlayerIdx = (b.currentPlayerIdx + 1) % len(b.players)
	}
}

// validatePlay カードのプレイがルール上有効かを検証する。
// Brusquembille には must-follow がないため、プレイヤーが手札に持つカードであれば常に有効。
// IsFollowRequired は「いまリードスートに追従する義務があるか」を返す。
//
// **ブリュスカンビーユの肝はこの二相構造。** 山札が残っているあいだは
// クローン元のブリスコラと同じく自由に出せるが、**山札を使い切った時点で
// 追従必須に切り替わる**。前半は手札を補充できるので自由に捨てられ、
// 後半は補充が無いので持っている札で受けざるを得ない、という設計。
func (b *Brusquembille) IsFollowRequired() bool {
	return b.GetStockRemaining() == 0 && b.trumpCard == nil
}

// hasSuit は playerIdx がそのスートの札を持っているかを返す。
func (b *Brusquembille) hasSuit(playerIdx, suit int) bool {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return false
	}
	p := b.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetDesign() == suit {
			return true
		}
	}
	return false
}

// leadSuit は現在のトリックのリードスートを返す (まだ誰も出していなければ -1)。
func (b *Brusquembille) leadSuit() int {
	if len(b.currentTrick) == 0 {
		return -1
	}
	if c := b.currentTrick[0].Card; c != nil {
		return c.GetDesign()
	}
	return -1
}

func (b *Brusquembille) validatePlay(playerIdx int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードが nil です")
	}
	// 前半 (山札あり) は自由出し。クローン元と同じ。
	if !b.IsFollowRequired() {
		return nil
	}
	lead := b.leadSuit()
	// リードなら何を出してもよい。
	if lead < 0 {
		return nil
	}
	// **持っているのに違うスートを出すのは反則。** 持っていなければ自由。
	if card.GetDesign() != lead && b.hasSuit(playerIdx, lead) {
		return NewDomainError(ErrInvalidCard, "山札が尽きた後はリードスートに追従してください")
	}
	return nil
}

// trickWinner 現在のトリックの勝者インデックスを決定する
func (b *Brusquembille) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerCard := b.currentTrick[0].Card

	for _, tc := range b.currentTrick[1:] {
		if brusquembilleBeats(tc.Card, winnerCard, leadSuit, b.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// brusquembilleBeats challenger が currentBest に勝つかを判定する。
// ・両者がトランプ: ブリュスカンビーユ順位の高い方が勝つ
// ・challenger のみトランプ: challenger が勝つ
// ・両者とも非トランプかつ同じリードスート: ブリュスカンビーユ順位の高い方が勝つ
// ・両者とも非トランプで challenger がリードスート以外: challenger は勝てない
func brusquembilleBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit

	switch {
	case cIsTrump && bIsTrump:
		return BrusquembilleRankOrder(challenger) > BrusquembilleRankOrder(currentBest)
	case cIsTrump:
		return true
	case bIsTrump:
		return false
	}
	// ともに非トランプ
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return BrusquembilleRankOrder(challenger) > BrusquembilleRankOrder(currentBest)
}

// drawReplenish トリック勝者が先に 1 枚、次に敗者が 1 枚を山札から引く。
// 山札が空になっていく過程では、最後の 1 枚は表向きトランプ (trumpCard) を引いた扱いになる。
func (b *Brusquembille) drawReplenish() {
	if b.trumpCards.GetRemainingCount() == 0 && b.trumpCard == nil {
		return
	}
	winnerIdx := b.leadPlayerIdx
	// **勝者から順に、卓を一周して補充する。** クローン元は 2 人固定なので
	// 「勝者ともう一人」で済んでいた。
	for i := range b.players {
		idx := (winnerIdx + i) % len(b.players)
		if c := b.drawOne(); c != nil {
			b.players[idx].AddCard(c)
			b.sortHand(b.players[idx])
		}
	}
}

// drawOne 山札またはトランプカードから 1 枚引く。優先順位は山札 → トランプカード。
func (b *Brusquembille) drawOne() *Card {
	return drawOrTakeTrump(b.trumpCards, &b.trumpCard)
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (b *Brusquembille) allHandsEmpty() bool {
	return allHandsEmpty(b.players)
}

// finishGame ゲームを終了させ、勝者を決定する
func (b *Brusquembille) finishGame() {
	b.gameEndFlag = true
	b.phase = BrusquembillePhaseGameEnd
	b.winnerIdx = BrusquembilleDetermineWinner(b.playerPoints)
	parts := make([]string, 0, len(b.playerPoints))
	for _, pt := range b.playerPoints {
		parts = append(parts, strconv.Itoa(pt))
	}
	detail := "Game end: " + strings.Join(parts, "-")
	b.appendLog(-1, "game_end", detail, nil)
}

// BrusquembilleDetermineWinner は最多得点の席を返す。同点が並べば -1 (引き分け)。
//
// **全席を見る。** クローン元のブリスコラは 2 人固定なので「席 0 が 60 超なら
// 席 0、そうでなく席 1 が 60 超なら席 1」で足りたが、この卓は 2〜5 席ある。
// その形のまま席数だけ増やすと、**席 2 以降がどれだけ取っても勝者にならない**。
//
// 2 人卓では従来どおり「60 点超で勝ち、60-60 は引き分け」になる: 合計 120 点を
// 二人で分けるので、単独最多は必ず 60 点超だから。3 席以上では 60 を超えなくても
// 単独最多なら勝ちで、これが素直な一般化。
func BrusquembilleDetermineWinner(points []int) int {
	best, bestIdx, tied := -1, -1, false
	for i, pt := range points {
		switch {
		case pt > best:
			best, bestIdx, tied = pt, i, false
		case pt == best:
			tied = true
		}
	}
	if tied || bestIdx < 0 {
		return -1
	}
	return bestIdx
}

// sortAllHands 全プレイヤーの手札をソートする
func (b *Brusquembille) sortAllHands() {
	sortEachHand(b.players, b.sortHand)
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ブリュスカンビーユ順位 でソートする
func (b *Brusquembille) sortHand(p *BrusquembillePlayer) {
	trumpSuit := b.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump // 非トランプを先に
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return BrusquembilleRankOrder(ci) < BrusquembilleRankOrder(cj)
	})
}

// playHintReason ヒント理由キーを判定する
func (b *Brusquembille) playHintReason(playerIdx, chosenIdx int) string {
	card := b.players[playerIdx].GetCard(chosenIdx)
	pts := BrusquembilleCardPoints(card)
	if len(b.currentTrick) == 0 {
		if card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		if pts == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if brusquembilleBeats(card, leadCard, leadSuit, b.trumpSuit) {
		if card.GetDesign() == b.trumpSuit && leadSuit != b.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI (single-difficulty heuristic) ---

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する
func (b *Brusquembille) cpuSelectPlayCard(playerIdx int) int {
	player := b.players[playerIdx]
	// **手札が空なら選びようがない。**cpuLead / cpuFollow は最初の札を無条件に
	// 読むので、ここで抜けないとセレクタの中でパニックする (#4606)。
	if player.GetCardsSize() <= 1 {
		return 0
	}

	if len(b.currentTrick) == 0 {
		return b.cpuLead(playerIdx)
	}
	return b.cpuFollow(playerIdx)
}

// cpuLead リード時の選択: 最も低い点数の非トランプを優先する。
// 全カードがトランプ・点数札しか無い場合は最も弱い順位のカードを選ぶ。
func (b *Brusquembille) cpuLead(playerIdx int) int {
	player := b.players[playerIdx]
	bestIdx := 0
	bestScore := brusqCpuLeadScore(player.GetCard(0), b.trumpSuit)
	for i := 1; i < player.GetCardsSize(); i++ {
		s := brusqCpuLeadScore(player.GetCard(i), b.trumpSuit)
		if s < bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

// brusqCpuLeadScore 値が小さいほど「リードに適している」(トランプを温存し、点数の高い札を温存する)
func brusqCpuLeadScore(c *Card, trumpSuit int) int {
	score := BrusquembilleCardPoints(c)*10 + BrusquembilleRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時の選択。
// 1) リードがトランプ: 勝てる最小トランプ、無ければ最小点数札を捨てる
// 2) リードが点数札 (A/3): 勝てる最小トランプ、無ければ最小点数札を捨てる
// 3) リードが低点数: 同スートの最小勝ち札 (非トランプ) があればそれ、無ければ最小点数札を捨てる
//
// いずれの分岐でも、トリック既出点数が高い (>= 11) 場合は積極的にトランプで奪取する。
func (b *Brusquembille) cpuFollow(playerIdx int) int {
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	trickPoints := BrusquembilleCardPoints(leadCard)

	// 既存の同スート勝ちカード (非トランプ前提) を探す
	if leadSuit != b.trumpSuit {
		if idx := brusqSmallestSameSuitWinner(player, leadCard, leadSuit); idx >= 0 && trickPoints == 0 {
			return idx
		}
	}

	// 高点数または相手がトランプ → トランプで取りに行く価値あり
	if trickPoints >= 10 || leadCard.GetDesign() == b.trumpSuit {
		if idx := brusqSmallestWinningTrump(player, leadCard, leadSuit, b.trumpSuit); idx >= 0 {
			return idx
		}
	}

	// 同スート勝ちで点数があるなら奪取
	if leadSuit != b.trumpSuit {
		if idx := brusqSmallestSameSuitWinner(player, leadCard, leadSuit); idx >= 0 {
			return idx
		}
	}

	return brusqSmallestDump(player, b.trumpSuit)
}

// brusqSmallestSameSuitWinner リードスートに従って勝てる最小ランクのカード (非トランプ) を返す
func brusqSmallestSameSuitWinner(player *BrusquembillePlayer, leadCard *Card, leadSuit int) int {
	bestIdx := -1
	bestRank := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != leadSuit {
			continue
		}
		if BrusquembilleRankOrder(c) <= BrusquembilleRankOrder(leadCard) {
			continue
		}
		r := BrusquembilleRankOrder(c)
		if bestIdx < 0 || r < bestRank {
			bestIdx = i
			bestRank = r
		}
	}
	return bestIdx
}

// brusqSmallestWinningTrump 勝てる最小ランクのトランプを返す。
// リード自体がトランプの場合はそれより強いトランプを探す。
func brusqSmallestWinningTrump(player *BrusquembillePlayer, leadCard *Card, leadSuit, trumpSuit int) int {
	bestIdx := -1
	bestRank := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != trumpSuit {
			continue
		}
		// リードがトランプならランクで上回る必要がある
		if leadSuit == trumpSuit && BrusquembilleRankOrder(c) <= BrusquembilleRankOrder(leadCard) {
			continue
		}
		r := BrusquembilleRankOrder(c)
		if bestIdx < 0 || r < bestRank {
			bestIdx = i
			bestRank = r
		}
	}
	return bestIdx
}

// brusqSmallestDump 取られても痛くないカードを 1 枚捨てる。
// 優先順: 非トランプの 0 点札 → 非トランプの低点数札 → 低ランクのトランプ。
func brusqSmallestDump(player *BrusquembillePlayer, trumpSuit int) int {
	bestIdx := 0
	bestScore := brusqDumpScore(player.GetCard(0), trumpSuit)
	for i := 1; i < player.GetCardsSize(); i++ {
		s := brusqDumpScore(player.GetCard(i), trumpSuit)
		if s < bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

// brusqDumpScore 値が小さいほど「失っても良い」カード
func brusqDumpScore(c *Card, trumpSuit int) int {
	score := BrusquembilleCardPoints(c)*10 + BrusquembilleRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// --- JSON ---

// brusquembilleJSON is the JSON wire format for Brusquembille.
type brusquembilleJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*BrusquembillePlayer `json:"ps"`
	Config           BrusquembilleConfig    `json:"cf"`
	Phase            BrusquembillePhase     `json:"ph"`
	TrickNumber      int                    `json:"tn"`
	CurrentPlayerIdx int                    `json:"ci"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	TrumpCard        *Card                  `json:"tu"`
	TrumpSuit        int                    `json:"ts"`
	LeadPlayerIdx    int                    `json:"li"`
	DealerIdx        int                    `json:"di"`
	PlayerPoints     []int                  `json:"pp"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Brusquembille) MarshalJSON() ([]byte, error) {
	return json.Marshal(brusquembilleJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		TrickNumber:      b.trickNumber,
		CurrentPlayerIdx: b.currentPlayerIdx,
		CurrentTrick:     b.currentTrick,
		TrumpCard:        b.trumpCard,
		TrumpSuit:        b.trumpSuit,
		LeadPlayerIdx:    b.leadPlayerIdx,
		DealerIdx:        b.dealerIdx,
		PlayerPoints:     b.playerPoints,
		GameEndFlag:      b.gameEndFlag,
		WinnerIdx:        b.winnerIdx,
		ActionLog:        b.actionLog,
	})
}

// brusquembilleMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const brusquembilleMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
//
// Validates that the deserialised game state matches Brusquembille's fixed shape
// (2-5 players, at most one card per seat on the current
// trick, PlayerPoints aligned to the player count) and that the variable-length
// ActionLog does not exceed brusquembilleMaxSliceLen, preventing DoS via crafted
// payloads and out-of-bounds access during play.
func (b *Brusquembille) UnmarshalJSON(data []byte) error {
	var j brusquembilleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) < BrusquembilleMinPlayerCnt || len(j.Players) > BrusquembilleMaxPlayerCnt {
		return fmt.Errorf("brusquembille: expected %d-%d players, got %d",
			BrusquembilleMinPlayerCnt, BrusquembilleMaxPlayerCnt, len(j.Players))
	}
	if len(j.CurrentTrick) > len(j.Players) {
		return fmt.Errorf("brusquembille: current trick has %d cards (max %d)", len(j.CurrentTrick), len(j.Players))
	}
	if j.PlayerPoints != nil && len(j.PlayerPoints) != len(j.Players) {
		return fmt.Errorf("brusquembille: expected %d player points entries, got %d", len(j.Players), len(j.PlayerPoints))
	}
	if len(j.ActionLog) > brusquembilleMaxSliceLen {
		return fmt.Errorf("brusquembille: action log exceeds maximum allowed size")
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards32()
	}
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
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
	b.playerPoints = j.PlayerPoints
	if b.playerPoints == nil {
		b.playerPoints = make([]int, len(b.players))
	}
	b.gameEndFlag = j.GameEndFlag
	b.winnerIdx = j.WinnerIdx
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
