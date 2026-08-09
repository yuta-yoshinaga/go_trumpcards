//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// BidWhistPlayerCnt Bid Whist のプレイヤー数
const BidWhistPlayerCnt = 4

// BidWhistHandSize 各プレイヤーの手札枚数 (54枚 − 6枚キティ = 48枚 / 4人)
const BidWhistHandSize = 12

// BidWhistKittySize キティ(場札)枚数
const BidWhistKittySize = 6

// BidWhistTeamCnt チーム数
const BidWhistTeamCnt = 2

// BidWhistTrickCnt 1ラウンドのトリック数
const BidWhistTrickCnt = 12

// BidWhistBook ブック (この枚数を超えたトリックがビッドの達成にカウントされる)
const BidWhistBook = 6

// BidWhistMinBid 最小ビッド (ブックを超えて宣言する目標トリック数の下限)
const BidWhistMinBid = 1

// BidWhistMaxBid 最大ビッド (全12トリックのうちブック6を超える最大)
const BidWhistMaxBid = 7

// bidWhistJokerSuit ノートランプ時のジョーカーの仮想スート (どの実スートにも一致しない)
const bidWhistJokerSuit = 99

// BidWhistPhase ゲームフェーズ
type BidWhistPhase int

// Bid Whist のフェーズ定数
const (
	// BidWhistPhaseBid オークション(ビッド)フェーズ
	BidWhistPhaseBid BidWhistPhase = 0
	// BidWhistPhaseTrumpDeclaration 切り札宣言フェーズ (Uptown/Downtown のみ。落札者がスートを選ぶ)
	BidWhistPhaseTrumpDeclaration BidWhistPhase = 1
	// BidWhistPhaseKittyExchange キティ交換フェーズ (落札者が6枚受け取り6枚捨てる)
	BidWhistPhaseKittyExchange BidWhistPhase = 2
	// BidWhistPhasePlay トリックプレイフェーズ
	BidWhistPhasePlay BidWhistPhase = 3
	// BidWhistPhaseTrickEnd トリック終了フェーズ
	BidWhistPhaseTrickEnd BidWhistPhase = 4
	// BidWhistPhaseRoundEnd ラウンド終了フェーズ
	BidWhistPhaseRoundEnd BidWhistPhase = 5
	// BidWhistPhaseGameEnd ゲーム終了フェーズ
	BidWhistPhaseGameEnd BidWhistPhase = 6
)

// Bid Whist のビッド方向 (カードの強さの向き / 切り札の有無)
const (
	// BidWhistDirectionUptown アップタウン (切り札あり、A が最強の通常序列)
	BidWhistDirectionUptown = 0
	// BidWhistDirectionDowntown ダウンタウン (切り札あり、2 が最強の逆序列)
	BidWhistDirectionDowntown = 1
	// BidWhistDirectionNoTrump ノートランプ (切り札なし、A 最強。ジョーカーは死札)
	BidWhistDirectionNoTrump = 2
)

// BidWhistBid ビッド(契約)を表す値。Tricks はブックを超える目標トリック数 (1-7)。
type BidWhistBid struct {
	Tricks    int `json:"t"`
	Direction int `json:"d"`
}

// valid はビッドが文法的に妥当かを返す。
func (b BidWhistBid) valid() bool {
	if b.Tricks < BidWhistMinBid || b.Tricks > BidWhistMaxBid {
		return false
	}
	return b.Direction >= BidWhistDirectionUptown && b.Direction <= BidWhistDirectionNoTrump
}

// Order はビッドの比較順位を返す (高いほど強い)。同じトリック数では
// Uptown < Downtown < NoTrump の順で No Trump が最も強いビッドとなる。
func (b BidWhistBid) Order() int {
	return b.Tricks*10 + b.Direction
}

// BidWhistHint ヒント情報
type BidWhistHint struct {
	BidTricks      *int   // 推奨トリック数 (ビッドフェーズ)
	BidDirection   *int   // 推奨方向 (ビッドフェーズ)
	Pass           *bool  // パス推奨か
	TrumpSuit      *int   // 推奨切り札スート (切り札宣言フェーズ)
	DiscardIndices []int  // 推奨ディスカード6枚 (キティ交換フェーズ)
	CardIndex      *int   // 推奨カードインデックス (プレイフェーズ)
	Reason         string // ヒント理由キー
}

// BidWhist Bid Whist ゲームクラス
//
// ルール概要:
//   - 4人2チーム制 (idx 0+2 vs 1+3)、54枚デッキ (52枚 + ジョーカー2枚)、各12枚配布、6枚キティ。
//   - ビッドはディーラーの左隣から1人1回。目標トリック数(1-7)+方向(Uptown/Downtown/No Trump)またはパス。
//     2人目以降は現在の最高ビッドより厳密に強いビッドのみ。全員パスなら再配布。
//   - 落札者は (切り札ありなら) 切り札スートを宣言し、キティ6枚を受け取って6枚捨てる。
//   - ジョーカーは切り札ゲームでは最強の2枚 (ビッグ>リトル)。ノートランプでは死札。
//   - Uptown は A 最強、Downtown は 2 最強の逆序列。
//   - スコア: ブック(6)を超えたトリック数。落札チームが 6+ビッド 以上取れば +(獲得−6)、
//     未達ならセットで −ビッド。守備チームは 6 を超えた分だけ加点。先取 TargetScore で勝利。
type BidWhist struct {
	trumpCards       *TrumpCards
	players          []*BidWhistPlayer
	config           BidWhistConfig
	phase            BidWhistPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	kitty            []*Card
	// declarerKitty retains a copy of the six kitty cards handed to the declarer
	// at bid finalisation, kept only through the kitty-exchange phase so the UI
	// can highlight which cards in the declarer's merged hand came from the
	// kitty. Cleared once play begins. Never used for game logic.
	declarerKitty []*Card
	leadPlayerIdx int
	// --- bidding state ---
	bidPlayerIdx  int
	passed        [BidWhistPlayerCnt]bool
	highestBid    *BidWhistBid
	highestBidder int
	// --- contract state ---
	contract    BidWhistBid
	declarerIdx int
	trumpSuit   int // 切り札スート (Uptown/Downtown のみ。NT/未宣言は -1)
	// --- scoring ---
	teamScores  [BidWhistTeamCnt]int
	gameEndFlag bool
	winnerTeam  int
	actionLogBase
}

// NewBidWhist コンストラクタ
func NewBidWhist(trumpCards *TrumpCards, players []*BidWhistPlayer, config BidWhistConfig) *BidWhist {
	return &BidWhist{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerTeam:    -1,
		roundNumber:   0,
		dealerIdx:     0,
		trumpSuit:     -1,
		highestBidder: -1,
		declarerIdx:   -1,
	}
}

// NewDefaultBidWhist は4人 (1人間 + 3 CPU) の標準セットアップを返す。
// CUI / Web / Worker 共通の構築 SSoT。
func NewDefaultBidWhist() *BidWhist {
	players := []*BidWhistPlayer{
		NewBidWhistPlayer(true, 0),
		NewBidWhistPlayer(false, 1),
		NewBidWhistPlayer(false, 0),
		NewBidWhistPlayer(false, 1),
	}
	return NewBidWhist(NewTrumpCards(2), players, DefaultBidWhistConfig())
}

// Reset ゲーム初期化
func (g *BidWhist) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [BidWhistTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *BidWhist) NextRound() {
	if g.phase != BidWhistPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % BidWhistPlayerCnt
	g.startRound()
}

// startRound ラウンドの状態を初期化して配り直す
func (g *BidWhist) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.passed = [BidWhistPlayerCnt]bool{}
	g.highestBid = nil
	g.highestBidder = -1
	g.contract = BidWhistBid{}
	g.declarerIdx = -1
	g.trumpSuit = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.dealRound()
	g.phase = BidWhistPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % BidWhistPlayerCnt
}

// dealRound 12枚ずつ配り、残り6枚をキティにする
func (g *BidWhist) dealRound() {
	g.trumpCards.Shuffle()
	g.kitty = nil
	g.declarerKitty = nil
	for range BidWhistHandSize {
		for j := range BidWhistPlayerCnt {
			if card := g.trumpCards.DrawCard(); card != nil {
				g.players[j].AddCard(card)
			}
		}
	}
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.kitty = append(g.kitty, card)
	}
	g.sortAllHands()
}

// --- Bidding ---

// PlayerBid 人間プレイヤーがビッドする
func (g *BidWhist) PlayerBid(tricks, direction int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BidWhistPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	bid := BidWhistBid{Tricks: tricks, Direction: direction}
	if !bid.valid() {
		return NewDomainError(ErrInvalidPlay, "無効なビッドです")
	}
	if g.highestBid != nil && bid.Order() <= g.highestBid.Order() {
		return NewDomainError(ErrInvalidPlay, "現在のビッドより強い必要があります")
	}
	g.applyBid(humanIdx, bid)
	return nil
}

// PlayerPass 人間プレイヤーがパスする
func (g *BidWhist) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BidWhistPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	g.applyPass(humanIdx)
	return nil
}

// CpuBid CPUプレイヤーが1ビッド実行する
func (g *BidWhist) CpuBid() {
	if g.gameEndFlag || g.phase != BidWhistPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= BidWhistPlayerCnt {
		return
	}
	if g.players[g.bidPlayerIdx].GetIsHuman() {
		return
	}
	if bid, ok := g.cpuSelectBid(g.bidPlayerIdx); ok {
		g.applyBid(g.bidPlayerIdx, bid)
	} else {
		g.applyPass(g.bidPlayerIdx)
	}
}

// applyBid ビッドを適用する
func (g *BidWhist) applyBid(idx int, bid BidWhistBid) {
	b := bid
	g.players[idx].SetBid(&b)
	g.highestBid = &b
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), bidWhistBidLabel(b)), nil)
	g.advanceBid()
}

// applyPass パスを適用する
func (g *BidWhist) applyPass(idx int) {
	g.passed[idx] = true
	g.players[idx].SetPassed(true)
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid ビッドを次へ進める。1巡 (4人) で終了し、落札判定を行う。
func (g *BidWhist) advanceBid() {
	g.bidPlayerIdx = (g.bidPlayerIdx + 1) % BidWhistPlayerCnt
	if g.bidPlayerIdx == (g.dealerIdx+1)%BidWhistPlayerCnt {
		g.finishBid()
	}
}

// finishBid 1巡が終わったら落札を確定するか、全員パスなら再配布する
func (g *BidWhist) finishBid() {
	if g.highestBidder < 0 {
		g.redeal()
		return
	}
	g.finalizeBid()
}

// redeal 全員パスした場合、同じディーラーで配り直す
func (g *BidWhist) redeal() {
	g.appendLog(-1, "redeal", "All players passed. Redealing.", nil)
	for _, p := range g.players {
		p.ResetRound()
	}
	g.highestBid = nil
	g.highestBidder = -1
	g.passed = [BidWhistPlayerCnt]bool{}
	g.dealRound()
	g.phase = BidWhistPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % BidWhistPlayerCnt
}

// finalizeBid 落札を確定し、落札者にキティを渡す
func (g *BidWhist) finalizeBid() {
	g.contract = *g.highestBid
	g.declarerIdx = g.highestBidder
	g.players[g.declarerIdx].SetIsDeclarer(true)
	g.declarerKitty = append([]*Card(nil), g.kitty...)
	for _, c := range g.kitty {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.kitty = nil
	g.appendLog(g.declarerIdx, "win_bid",
		fmt.Sprintf("%s wins the bid: %s", g.playerName(g.declarerIdx), bidWhistBidLabel(g.contract)), nil)
	g.sortAllHands()
	g.currentPlayerIdx = g.declarerIdx
	if g.isNoTrump() {
		g.trumpSuit = -1
		g.phase = BidWhistPhaseKittyExchange
	} else {
		g.phase = BidWhistPhaseTrumpDeclaration
	}
}

// --- Trump declaration ---

// PlayerDeclareTrump 人間(落札者)が切り札スートを宣言する
func (g *BidWhist) PlayerDeclareTrump(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BidWhistPhaseTrumpDeclaration {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !bidWhistValidSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "切り札スートは ♠/♣/♥/♦ から選んでください")
	}
	g.applyTrumpDeclaration(suit)
	return nil
}

// CpuDeclareTrump CPU(落札者)が切り札スートを宣言する
func (g *BidWhist) CpuDeclareTrump() {
	if g.gameEndFlag || g.phase != BidWhistPhaseTrumpDeclaration {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	g.applyTrumpDeclaration(g.cpuSelectTrump(g.declarerIdx))
}

// applyTrumpDeclaration 切り札スートを設定し、キティ交換フェーズへ進む
func (g *BidWhist) applyTrumpDeclaration(suit int) {
	g.trumpSuit = suit
	g.appendLog(g.declarerIdx, "trump",
		fmt.Sprintf("%s declares %s as trump", g.playerName(g.declarerIdx), suitName(suit)), nil)
	g.sortAllHands()
	g.phase = BidWhistPhaseKittyExchange
}

// --- Kitty exchange ---

// PlayerExchangeKitty 人間(落札者)が6枚捨てる
func (g *BidWhist) PlayerExchangeKitty(discardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BidWhistPhaseKittyExchange {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doExchange(discardIndices)
}

// CpuExchange CPU(落札者)が6枚捨てる
func (g *BidWhist) CpuExchange() {
	if g.gameEndFlag || g.phase != BidWhistPhaseKittyExchange {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	_ = g.doExchange(g.cpuSelectDiscards(g.declarerIdx))
}

// doExchange キティ交換の共通処理
func (g *BidWhist) doExchange(discardIndices []int) error {
	player := g.players[g.declarerIdx]
	if len(discardIndices) != BidWhistKittySize {
		return NewDomainError(ErrInvalidCard, "6枚捨ててください")
	}
	seen := make(map[int]bool, BidWhistKittySize)
	for _, idx := range discardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードは選べません")
		}
		seen[idx] = true
	}
	discarded := player.RemoveCards(discardIndices)
	g.kitty = discarded
	g.appendLog(g.declarerIdx, "exchange",
		fmt.Sprintf("%s discards %d cards", g.playerName(g.declarerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlayPhase()
	return nil
}

// --- Play ---

// startPlayPhase プレイフェーズを開始する (落札者がリード)
func (g *BidWhist) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.declarerKitty = nil
	g.leadPlayerIdx = g.declarerIdx
	g.currentPlayerIdx = g.declarerIdx
	g.phase = BidWhistPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *BidWhist) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BidWhistPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行する
func (g *BidWhist) CpuPlay() {
	if g.gameEndFlag || g.phase != BidWhistPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	idx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	played := g.players[g.currentPlayerIdx].RemoveCard(idx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(g.currentPlayerIdx, played)
}

// playCard カードをプレイする共通処理
func (g *BidWhist) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", g.playerName(playerIdx), bidWhistCardLabel(card)), []*Card{card})
	if len(g.currentTrick) == BidWhistPlayerCnt {
		g.phase = BidWhistPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % BidWhistPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *BidWhist) ResolveTrick() {
	if g.phase != BidWhistPhaseTrickEnd || len(g.currentTrick) != BidWhistPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	cards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		cards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(cards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", g.playerName(winnerIdx), g.trickNumber), cards)
	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= BidWhistTrickCnt {
		g.phase = BidWhistPhaseRoundEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *BidWhist) NextTrick() {
	if g.phase != BidWhistPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = BidWhistPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *BidWhist) ScoreRound() {
	if g.phase != BidWhistPhaseRoundEnd {
		return
	}
	declTeam := g.players[g.declarerIdx].GetTeam()
	defTeam := 1 - declTeam
	teamTricks := [BidWhistTeamCnt]int{}
	for _, p := range g.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}
	bid := g.contract.Tricks
	need := BidWhistBook + bid
	declTricks := teamTricks[declTeam]

	if declTricks >= need {
		gain := declTricks - BidWhistBook
		g.teamScores[declTeam] += gain
		g.appendLog(-1, "contract_made",
			fmt.Sprintf("Team %d makes the bid (%d/%d tricks). +%d", declTeam, declTricks, need, gain), nil)
	} else {
		g.teamScores[declTeam] -= bid
		g.appendLog(-1, "contract_failed",
			fmt.Sprintf("Team %d is set (%d/%d tricks). -%d", declTeam, declTricks, need, bid), nil)
	}
	if defTricks := teamTricks[defTeam]; defTricks > BidWhistBook {
		g.teamScores[defTeam] += defTricks - BidWhistBook
		g.appendLog(-1, "defender_score",
			fmt.Sprintf("Team %d (defenders) take %d over the book. +%d", defTeam, defTricks-BidWhistBook, defTricks-BidWhistBook), nil)
	}

	for ti := range BidWhistTeamCnt {
		g.appendLog(-1, "team_score", fmt.Sprintf("Team %d: %d points", ti, g.teamScores[ti]), nil)
	}
	g.checkGameEnd(declTeam)
}

// checkGameEnd ゲーム終了判定。先に TargetScore に到達したチームが勝利。
// 同点でタイブレークが必要な場合は落札チームを優先する。
func (g *BidWhist) checkGameEnd(declTeam int) {
	reached := false
	for ti := 0; ti < BidWhistTeamCnt; ti++ {
		if g.teamScores[ti] >= g.config.TargetScore {
			reached = true
			break
		}
	}
	if !reached {
		return
	}
	g.gameEndFlag = true
	g.phase = BidWhistPhaseGameEnd
	switch {
	case g.teamScores[0] > g.teamScores[1]:
		g.winnerTeam = 0
	case g.teamScores[1] > g.teamScores[0]:
		g.winnerTeam = 1
	default:
		g.winnerTeam = declTeam
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", g.winnerTeam), nil)
}

// --- Card ranking ---

// isNoTrump ノートランプ契約かどうか
func (g *BidWhist) isNoTrump() bool { return g.contract.Direction == BidWhistDirectionNoTrump }

// isDowntown ダウンタウン契約 (2 が最強の逆序列) かどうか
func (g *BidWhist) isDowntown() bool { return g.contract.Direction == BidWhistDirectionDowntown }

// isJoker カードがジョーカーかどうか
func (g *BidWhist) isJoker(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// isBigJoker ビッグジョーカー (value=2) かどうか
func (g *BidWhist) isBigJoker(c *Card) bool {
	return g.isJoker(c) && c.GetValue() == 2
}

// directionalRank はカードの値を方向に応じた強さ順位 (大=強) に変換する。
//   - Uptown / No Trump: A(=14) > K(13) > … > 2(2)
//   - Downtown: 2(=14) > 3(13) > … > K(3) > A(2)
func (g *BidWhist) directionalRank(value int) int {
	if g.isDowntown() {
		if value == 1 {
			return 2 // Ace is lowest downtown
		}
		return 16 - value // 2→14 … 13(K)→3
	}
	if value == 1 {
		return 14 // Ace high uptown / no-trump
	}
	return value
}

// effectiveSuit カードの実効スートを返す。ジョーカーは切り札ゲームでは切り札スート、
// ノートランプでは仮想スート (どの実スートにも一致しない)。
func (g *BidWhist) effectiveSuit(c *Card) int {
	if c == nil {
		return -1
	}
	if g.isJoker(c) {
		if g.isNoTrump() {
			return bidWhistJokerSuit
		}
		return g.trumpSuit
	}
	return c.GetDesign()
}

// cardRank トリック比較・ソート用のカードランクを返す (大=強)。
//
//	切り札ゲーム: ビッグジョーカー(902) > リトルジョーカー(901) > 切り札(400+) > 平札(100+)
//	ノートランプ: ジョーカー(0, 死札) < 平札(100+)
func (g *BidWhist) cardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if g.isJoker(c) {
		if g.isNoTrump() {
			return 0
		}
		if g.isBigJoker(c) {
			return 902
		}
		return 901
	}
	r := g.directionalRank(c.GetValue())
	if !g.isNoTrump() && c.GetDesign() == g.trumpSuit {
		return 400 + r
	}
	return 100 + r
}

// leadSuit 現在のトリックのリードスートを返す
func (g *BidWhist) leadSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	return g.effectiveSuit(g.currentTrick[0].Card)
}

// trickScore リードスートを踏まえたトリック比較値を返す (0以下=勝てない)
func (g *BidWhist) trickScore(c *Card, ls int) int {
	if g.isNoTrump() {
		if g.isJoker(c) {
			return -1 // 死札はどのトリックにも勝てない
		}
		// ジョーカーが (唯一の手札として) リードされた場合 ls=仮想スート。全員が競合する。
		if ls == bidWhistJokerSuit || g.effectiveSuit(c) == ls {
			return g.cardRank(c)
		}
		return 0
	}
	if g.isJoker(c) {
		return g.cardRank(c) // ジョーカーは常に最強の切り札
	}
	es := g.effectiveSuit(c)
	if es == g.trumpSuit || es == ls {
		return g.cardRank(c)
	}
	return 0
}

// trickWinner トリックの勝者を決定する
func (g *BidWhist) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	ls := g.leadSuit()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerScore := g.trickScore(g.currentTrick[0].Card, ls)
	for _, tc := range g.currentTrick[1:] {
		if s := g.trickScore(tc.Card, ls); s > winnerScore {
			winnerScore = s
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// validatePlay カードのプレイが有効か検証する
func (g *BidWhist) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		// ノートランプでは死札ジョーカーをリードできない (唯一の手札の場合を除く)
		if g.isNoTrump() && g.isJoker(card) && g.players[playerIdx].GetCardsSize() > 1 {
			return NewDomainError(ErrInvalidPlay, "ノートランプではジョーカーをリードできません")
		}
		return nil
	}
	ls := g.leadSuit()
	if g.effectiveSuit(card) != ls && g.playerHasSuit(playerIdx, ls) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが実効スートのカードを持っているか
func (g *BidWhist) playerHasSuit(playerIdx, suit int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if g.effectiveSuit(p.GetCard(i)) == suit {
			return true
		}
	}
	return false
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *BidWhist) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *BidWhist) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- CPU AI ---

// cpuSelectBid CPUのビッド選択 (ok=false でパス)
func (g *BidWhist) cpuSelectBid(playerIdx int) (BidWhistBid, bool) {
	bestEst := 0
	bestDir := BidWhistDirectionUptown
	for _, dir := range []int{BidWhistDirectionUptown, BidWhistDirectionDowntown, BidWhistDirectionNoTrump} {
		if est := g.estimateDirection(playerIdx, dir); est > bestEst {
			bestEst = est
			bestDir = dir
		}
	}
	threshold := BidWhistBook + 2
	switch g.config.CpuDifficulty {
	case BidWhistCpuDifficultyEasy:
		threshold = BidWhistBook + 3
	case BidWhistCpuDifficultyHard:
		threshold = BidWhistBook + 1
	}
	if bestEst < threshold {
		return BidWhistBid{}, false
	}
	tricks := bidWhistClamp(bestEst-BidWhistBook, BidWhistMinBid, BidWhistMaxBid)
	bid := BidWhistBid{Tricks: tricks, Direction: bestDir}
	if g.highestBid != nil && bid.Order() <= g.highestBid.Order() {
		return BidWhistBid{}, false
	}
	return bid, true
}

// estimateDirection 指定方向でおおよそ取れるトリック数を見積もる (切り札ゲームは最良スートを採用)
func (g *BidWhist) estimateDirection(playerIdx, dir int) int {
	if dir == BidWhistDirectionNoTrump {
		return g.estimateTricks(playerIdx, dir, -1)
	}
	best := 0
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if est := g.estimateTricks(playerIdx, dir, s); est > best {
			best = est
		}
	}
	return best
}

// estimateTricks 指定方向・スートでのおおよその獲得トリック数
func (g *BidWhist) estimateTricks(playerIdx, dir, suit int) int {
	p := g.players[playerIdx]
	est, jokers, trumpLen := 0, 0, 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == CardDesignJoker {
			jokers++
			continue
		}
		high := bidWhistIsHighCard(c.GetValue(), dir)
		if dir != BidWhistDirectionNoTrump && c.GetDesign() == suit {
			trumpLen++
			if high {
				est++
			}
		} else if high {
			est++
		}
	}
	if dir != BidWhistDirectionNoTrump {
		est += jokers
		if trumpLen >= 5 {
			est += trumpLen - 4
		}
	}
	return est
}

// cpuSelectTrump CPU(落札者)が切り札スートを選ぶ (最長 + 高札スート)
func (g *BidWhist) cpuSelectTrump(playerIdx int) int {
	p := g.players[playerIdx]
	score := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if g.isJoker(c) {
			continue
		}
		score[c.GetDesign()]++
		if bidWhistIsHighCard(c.GetValue(), g.contract.Direction) {
			score[c.GetDesign()] += 2
		}
	}
	best, bestScore := CardDesignSpade, -1
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if score[s] > bestScore {
			bestScore = score[s]
			best = s
		}
	}
	return best
}

// cpuSelectDiscards CPU(落札者)が捨てる6枚のインデックスを選ぶ (最弱6枚)
func (g *BidWhist) cpuSelectDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return g.cardRank(p.GetCard(idxs[a])) < g.cardRank(p.GetCard(idxs[b]))
	})
	count := BidWhistKittySize
	if count > n {
		count = n
	}
	return append([]int(nil), idxs[:count]...)
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選ぶ
func (g *BidWhist) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	p := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// リード: 最も強いカード
		bestIdx := valid[0]
		for _, idx := range valid[1:] {
			if g.cardRank(p.GetCard(idx)) > g.cardRank(p.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}
	// フォロー: パートナーが勝っていれば最弱、それ以外は勝てる最弱、無理なら最弱
	ls := g.leadSuit()
	winScore := g.currentWinnerScore(ls)
	if g.players[g.trickWinner()].GetTeam() == p.GetTeam() {
		return g.weakestValid(playerIdx, valid)
	}
	over := []int{}
	for _, idx := range valid {
		if g.trickScore(p.GetCard(idx), ls) > winScore {
			over = append(over, idx)
		}
	}
	if len(over) > 0 {
		bestIdx := over[0]
		for _, idx := range over[1:] {
			if g.cardRank(p.GetCard(idx)) < g.cardRank(p.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}
	return g.weakestValid(playerIdx, valid)
}

// weakestValid 有効札のうち最弱のインデックスを返す
func (g *BidWhist) weakestValid(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	bestIdx := valid[0]
	for _, idx := range valid[1:] {
		if g.cardRank(p.GetCard(idx)) < g.cardRank(p.GetCard(bestIdx)) {
			bestIdx = idx
		}
	}
	return bestIdx
}

// currentWinnerScore 現在のトリックでの暫定最強トリック値を返す
func (g *BidWhist) currentWinnerScore(ls int) int {
	best := 0
	for _, tc := range g.currentTrick {
		if s := g.trickScore(tc.Card, ls); s > best {
			best = s
		}
	}
	return best
}

// --- Hint ---

// GetHint 現在の人間の手番に対するヒントを返す
func (g *BidWhist) GetHint() *BidWhistHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case BidWhistPhaseBid:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		if bid, ok := g.cpuSelectBid(humanIdx); ok {
			tricks := bid.Tricks
			dir := bid.Direction
			return &BidWhistHint{BidTricks: &tricks, BidDirection: &dir, Reason: "strategic_bid"}
		}
		pass := true
		return &BidWhistHint{Pass: &pass, Reason: "pass_recommended"}
	case BidWhistPhaseTrumpDeclaration:
		if g.declarerIdx != humanIdx {
			return nil
		}
		suit := g.cpuSelectTrump(humanIdx)
		return &BidWhistHint{TrumpSuit: &suit, Reason: "trump_longest"}
	case BidWhistPhaseKittyExchange:
		if g.declarerIdx != humanIdx {
			return nil
		}
		return &BidWhistHint{DiscardIndices: g.cpuSelectDiscards(humanIdx), Reason: "discard_weakest"}
	case BidWhistPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuSelectPlayCard(humanIdx)
		return &BidWhistHint{CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する
func (g *BidWhist) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if !g.isNoTrump() && g.effectiveSuit(card) == g.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}
	if g.effectiveSuit(card) == g.leadSuit() {
		return "follow_suit"
	}
	if !g.isNoTrump() && g.effectiveSuit(card) == g.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *BidWhist) GetPhase() BidWhistPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *BidWhist) SetPhase(phase BidWhistPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *BidWhist) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号取得
func (g *BidWhist) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *BidWhist) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *BidWhist) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *BidWhist) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *BidWhist) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *BidWhist) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *BidWhist) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *BidWhist) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *BidWhist) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *BidWhist) GetPlayer(i int) *BidWhistPlayer {
	return getPlayer(g.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *BidWhist) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *BidWhist) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッド手番インデックス取得
func (g *BidWhist) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx ビッド手番インデックス設定 (テスト用)
func (g *BidWhist) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *BidWhist) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *BidWhist) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得 (-1 = NT/未宣言)
func (g *BidWhist) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *BidWhist) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetContractTricks 契約トリック数取得 (ブックを超える目標数)
func (g *BidWhist) GetContractTricks() int { return g.contract.Tricks }

// GetContractDirection 契約方向取得
func (g *BidWhist) GetContractDirection() int { return g.contract.Direction }

// SetContract 契約設定 (テスト用)
func (g *BidWhist) SetContract(tricks, direction, suit int) {
	g.contract = BidWhistBid{Tricks: tricks, Direction: direction}
	if direction == BidWhistDirectionNoTrump {
		g.trumpSuit = -1
	} else {
		g.trumpSuit = suit
	}
}

// GetDeclarerIdx 落札者インデックス取得 (-1 = 未確定)
func (g *BidWhist) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 落札者インデックス設定 (テスト用)
func (g *BidWhist) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetHighestBid 現在の最高ビッド取得 (nil = なし)
func (g *BidWhist) GetHighestBid() *BidWhistBid { return g.highestBid }

// GetHighestBidder 最高ビッダーのインデックス取得 (-1 = なし)
func (g *BidWhist) GetHighestBidder() int { return g.highestBidder }

// GetKitty キティ取得
func (g *BidWhist) GetKitty() []*Card { return g.kitty }

// SetDeclarerKitty 落札者へ渡したキティを設定 (テスト用)
func (g *BidWhist) SetDeclarerKitty(cards []*Card) { g.declarerKitty = cards }

// bidWhistCardKey は Card を (design, value) で一意に識別するキーを返す。
// 標準52枚 + ジョーカー2枚 (value 1/2 で区別) を通じて衝突しない。
func bidWhistCardKey(c *Card) int { return c.GetDesign()*100 + c.GetValue() }

// GetKittyIndices はキティ交換フェーズ中、落札者の手札のうちキティ由来の6枚の
// インデックスを返す。それ以外のフェーズ・落札者未確定・キティ非保持時は空スライスを返す。
// カード値で照合するためシリアライズを跨いでも安定する。読み取り専用でゲーム状態を変更しない。
func (g *BidWhist) GetKittyIndices() []int {
	if g.phase != BidWhistPhaseKittyExchange || g.declarerIdx < 0 ||
		g.declarerIdx >= len(g.players) || len(g.declarerKitty) == 0 {
		return []int{}
	}
	want := make(map[int]int, len(g.declarerKitty))
	for _, c := range g.declarerKitty {
		if c != nil {
			want[bidWhistCardKey(c)]++
		}
	}
	player := g.players[g.declarerIdx]
	indices := make([]int, 0, len(g.declarerKitty))
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		key := bidWhistCardKey(c)
		if want[key] > 0 {
			want[key]--
			indices = append(indices, i)
		}
	}
	return indices
}

// GetTeamScore チームスコア取得
func (g *BidWhist) GetTeamScore(team int) int {
	if team < 0 || team >= BidWhistTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *BidWhist) SetTeamScore(team, score int) {
	if team >= 0 && team < BidWhistTeamCnt {
		g.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *BidWhist) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (g *BidWhist) IsHumanBidTurn() bool {
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.bidPlayerIdx].GetIsHuman()
}

// IsHumanDeclarerTurn 現在の落札者(切り札宣言/キティ交換手番)が人間かどうか
func (g *BidWhist) IsHumanDeclarerTurn() bool {
	if g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *BidWhist) GetConfig() BidWhistConfig { return g.config }

// SetConfig 設定変更
func (g *BidWhist) SetConfig(cfg BidWhistConfig) { g.config = cfg }

// CardRankPublic カードランク取得 (テスト用)
func (g *BidWhist) CardRankPublic(card *Card) int { return g.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用)
func (g *BidWhist) EffectiveSuitPublic(card *Card) int { return g.effectiveSuit(card) }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *BidWhist) sortAllHands() {
	for _, p := range g.players {
		g.sortHand(p)
	}
}

// sortHand プレイヤーの手札を実効スート→ランク順にソートする
func (g *BidWhist) sortHand(p *BidWhistPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		si := g.effectiveSuit(cards[i])
		sj := g.effectiveSuit(cards[j])
		if si != sj {
			return si < sj
		}
		return g.cardRank(cards[i]) < g.cardRank(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。idx<0 はチーム単位イベント (集計・終局) の番兵。
func (g *BidWhist) playerName(idx int) string {
	if idx < 0 {
		return "System"
	}
	if idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// bidWhistValidSuit 4スートのうちいずれかか
func bidWhistValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// bidWhistIsHighCard 指定方向で「強い札」と見なせるか (見積もり用ヒューリスティック)
func bidWhistIsHighCard(value, dir int) bool {
	if dir == BidWhistDirectionDowntown {
		return value == 2 || value == 3 || value == 4
	}
	return value == 1 || value >= 12 // A, K, Q
}

// bidWhistClamp 整数を [min,max] に収める
func bidWhistClamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// bidWhistCardLabel カードのログ表示文字列 (ジョーカー対応)
func bidWhistCardLabel(c *Card) string {
	if c == nil {
		return "??"
	}
	if c.GetDesign() == CardDesignJoker {
		if c.GetValue() == 2 {
			return "BJ" // Big Joker
		}
		return "LJ" // Little Joker
	}
	return cardStr(c)
}

// bidWhistDirectionLabel 方向のログ表示文字列
func bidWhistDirectionLabel(dir int) string {
	switch dir {
	case BidWhistDirectionUptown:
		return "Uptown"
	case BidWhistDirectionDowntown:
		return "Downtown"
	case BidWhistDirectionNoTrump:
		return "No Trump"
	}
	return "?"
}

// bidWhistBidLabel ビッドのログ表示文字列
func bidWhistBidLabel(b BidWhistBid) string {
	return fmt.Sprintf("%d %s", b.Tricks, bidWhistDirectionLabel(b.Direction))
}

// --- JSON ---

// bidWhistJSON is the JSON wire format for BidWhist.
type bidWhistJSON struct {
	TrumpCards       *TrumpCards             `json:"tc"`
	Players          []*BidWhistPlayer       `json:"ps"`
	Config           BidWhistConfig          `json:"cf"`
	Phase            BidWhistPhase           `json:"ph"`
	RoundNumber      int                     `json:"rn"`
	TrickNumber      int                     `json:"tn"`
	CurrentPlayerIdx int                     `json:"ci"`
	CurrentTrick     []*TrickCard            `json:"ct"`
	DealerIdx        int                     `json:"di"`
	Kitty            []*Card                 `json:"kt"`
	DeclarerKitty    []*Card                 `json:"dk"`
	LeadPlayerIdx    int                     `json:"li"`
	BidPlayerIdx     int                     `json:"bi"`
	Passed           [BidWhistPlayerCnt]bool `json:"pd"`
	HighestBid       *BidWhistBid            `json:"hb"`
	HighestBidder    int                     `json:"hr"`
	Contract         BidWhistBid             `json:"cn"`
	DeclarerIdx      int                     `json:"dc"`
	TrumpSuit        int                     `json:"ts"`
	TeamScores       [BidWhistTeamCnt]int    `json:"sc"`
	GameEndFlag      bool                    `json:"ge"`
	WinnerTeam       int                     `json:"wt"`
	ActionLog        []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *BidWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(bidWhistJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		DealerIdx:        g.dealerIdx,
		Kitty:            g.kitty,
		DeclarerKitty:    g.declarerKitty,
		LeadPlayerIdx:    g.leadPlayerIdx,
		BidPlayerIdx:     g.bidPlayerIdx,
		Passed:           g.passed,
		HighestBid:       g.highestBid,
		HighestBidder:    g.highestBidder,
		Contract:         g.contract,
		DeclarerIdx:      g.declarerIdx,
		TrumpSuit:        g.trumpSuit,
		TeamScores:       g.teamScores,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// bidWhistMaxActionLogLen caps the action-log size during deserialisation
// (DoS guard). Other slices have exact structural bounds enforced below.
const bidWhistMaxActionLogLen = 5000

// UnmarshalJSON implements json.Unmarshaler.
func (g *BidWhist) UnmarshalJSON(data []byte) error {
	var j bidWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// Enforce tight, structural bounds on every restored slice (DoS guard).
	if len(j.Players) != BidWhistPlayerCnt {
		return fmt.Errorf("bidwhist: invalid player count: %d", len(j.Players))
	}
	if len(j.CurrentTrick) > BidWhistPlayerCnt || len(j.Kitty) > BidWhistKittySize ||
		len(j.DeclarerKitty) > BidWhistKittySize || len(j.ActionLog) > bidWhistMaxActionLogLen {
		return fmt.Errorf("bidwhist: input array exceeds maximum allowed size")
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(2)
	}
	g.players = j.Players
	for _, p := range g.players {
		if p == nil {
			return fmt.Errorf("bidwhist: player is nil")
		}
	}
	g.config = j.Config
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("bidwhist: invalid config: %w", err)
	}
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	for _, tc := range g.currentTrick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("bidwhist: trick card is nil")
		}
	}
	g.dealerIdx = j.DealerIdx
	g.kitty = j.Kitty
	if g.kitty == nil {
		g.kitty = make([]*Card, 0)
	}
	for _, c := range g.kitty {
		if c == nil {
			return fmt.Errorf("bidwhist: kitty card is nil")
		}
	}
	g.declarerKitty = j.DeclarerKitty
	for _, c := range g.declarerKitty {
		if c == nil {
			return fmt.Errorf("bidwhist: declarer kitty card is nil")
		}
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.passed = j.Passed
	g.highestBid = j.HighestBid
	g.highestBidder = j.HighestBidder
	g.contract = j.Contract
	g.declarerIdx = j.DeclarerIdx
	g.trumpSuit = j.TrumpSuit
	g.teamScores = j.TeamScores
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	for _, al := range g.actionLog {
		if al == nil {
			return fmt.Errorf("bidwhist: action log entry is nil")
		}
	}
	return nil
}
