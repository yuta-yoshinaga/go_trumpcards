//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// FiveHundredPlayerCnt 500 のプレイヤー数
const FiveHundredPlayerCnt = 4

// FiveHundredHandSize 各プレイヤーの手札枚数
const FiveHundredHandSize = 10

// FiveHundredKittySize キティ(場札)枚数
const FiveHundredKittySize = 3

// FiveHundredTeamCnt チーム数
const FiveHundredTeamCnt = 2

// FiveHundredTrickCnt 1ラウンドのトリック数
const FiveHundredTrickCnt = 10

// fiveHundredJokerSuit ノートランプ契約時のジョーカーの仮想スート (どの実スートにも一致しない)
const fiveHundredJokerSuit = 99

// FiveHundredPhase ゲームフェーズ
type FiveHundredPhase int

// FiveHundred のフェーズ定数
const (
	// FiveHundredPhaseBid オークション(ビッド)フェーズ
	FiveHundredPhaseBid FiveHundredPhase = 0
	// FiveHundredPhaseKittyExchange キティ交換フェーズ (落札者が3枚受け取り3枚捨てる)
	FiveHundredPhaseKittyExchange FiveHundredPhase = 1
	// FiveHundredPhasePlay トリックプレイフェーズ
	FiveHundredPhasePlay FiveHundredPhase = 2
	// FiveHundredPhaseTrickEnd トリック終了フェーズ
	FiveHundredPhaseTrickEnd FiveHundredPhase = 3
	// FiveHundredPhaseRoundEnd ラウンド終了フェーズ
	FiveHundredPhaseRoundEnd FiveHundredPhase = 4
	// FiveHundredPhaseGameEnd ゲーム終了フェーズ
	FiveHundredPhaseGameEnd FiveHundredPhase = 5
)

// FiveHundredContractKind 契約種別
type FiveHundredContractKind int

// FiveHundred の契約種別定数
const (
	// FiveHundredContractNone 契約未確定
	FiveHundredContractNone FiveHundredContractKind = 0
	// FiveHundredContractSuit 切り札契約 (6-10トリック)
	FiveHundredContractSuit FiveHundredContractKind = 1
	// FiveHundredContractNoTrump ノートランプ契約 (6-10トリック)
	FiveHundredContractNoTrump FiveHundredContractKind = 2
	// FiveHundredContractMisere ミゼール (落札者が単独で0トリックを目指す, 250点)
	FiveHundredContractMisere FiveHundredContractKind = 3
	// FiveHundredContractOpenMisere オープンミゼール (手札公開で0トリック, 520点)
	FiveHundredContractOpenMisere FiveHundredContractKind = 4
)

// FiveHundredBid ビッド(契約)を表す値
type FiveHundredBid struct {
	Kind   FiveHundredContractKind `json:"k"`
	Tricks int                     `json:"t"` // 切り札/NT契約は6-10, それ以外は0
	Suit   int                     `json:"s"` // 切り札契約は CardDesign*, それ以外は -1
}

// fiveHundredSuitBase 切り札スートの基礎点 (6トリック時の点数)。
// 500のスート序列 ♠<♣<♦<♥ に従う (数値スート定数とは順序が異なるため明示マップで対応)。
func fiveHundredSuitBase(suit int) int {
	switch suit {
	case CardDesignSpade:
		return 40
	case CardDesignClover:
		return 60
	case CardDesignDiamond:
		return 80
	case CardDesignHeart:
		return 100
	}
	return 0
}

// Value はビッドの得点(アヴォンデール表)を返す。
func (b FiveHundredBid) Value() int {
	switch b.Kind {
	case FiveHundredContractSuit:
		return fiveHundredSuitBase(b.Suit) + (b.Tricks-6)*100
	case FiveHundredContractNoTrump:
		return 120 + (b.Tricks-6)*100
	case FiveHundredContractMisere:
		return 250
	case FiveHundredContractOpenMisere:
		return 520
	}
	return 0
}

// Order はビッドの比較順位を返す (高いほど強いビッド)。
// オープンミゼールは10NT(520)を上回る最上位ビッドのため530として扱う。
func (b FiveHundredBid) Order() int {
	if b.Kind == FiveHundredContractOpenMisere {
		return 530
	}
	return b.Value()
}

// valid はビッドが文法的に妥当かを返す。
func (b FiveHundredBid) valid() bool {
	switch b.Kind {
	case FiveHundredContractSuit:
		return b.Tricks >= 6 && b.Tricks <= 10 && fiveHundredSuitBase(b.Suit) > 0
	case FiveHundredContractNoTrump:
		return b.Tricks >= 6 && b.Tricks <= 10
	case FiveHundredContractMisere, FiveHundredContractOpenMisere:
		return true
	}
	return false
}

// FiveHundredHint ヒント情報
type FiveHundredHint struct {
	BidKind        *int   // 推奨ビッド種別 (ビッドフェーズ)
	BidTricks      *int   // 推奨トリック数
	BidSuit        *int   // 推奨スート
	Pass           *bool  // パス推奨か
	DiscardIndices []int  // 推奨ディスカード3枚 (キティ交換フェーズ)
	CardIndex      *int   // 推奨カードインデックス (プレイフェーズ)
	JokerSuit      *int   // ジョーカーリード時の推奨スート
	Reason         string // ヒント理由キー
}

// FiveHundred 500 (Five Hundred) ゲームクラス
type FiveHundred struct {
	trumpCards       *TrumpCards
	players          []*FiveHundredPlayer
	config           FiveHundredConfig
	phase            FiveHundredPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	kitty            []*Card
	leadPlayerIdx    int
	jokerLeadSuit    int // ジョーカーがリードされた際の指名スート (-1 = なし)
	// --- bidding state ---
	bidPlayerIdx  int
	passed        [FiveHundredPlayerCnt]bool
	highestBid    *FiveHundredBid
	highestBidder int
	// --- contract state ---
	contract    FiveHundredBid
	declarerIdx int
	trumpSuit   int // 切り札スート (切り札契約のみ, それ以外は -1)
	// --- scoring ---
	teamScores  [FiveHundredTeamCnt]int
	gameEndFlag bool
	winnerTeam  int // 勝利チーム (-1 = 未確定)
	actionLogBase
}

// NewFiveHundred コンストラクタ
func NewFiveHundred(trumpCards *TrumpCards, players []*FiveHundredPlayer, config FiveHundredConfig) *FiveHundred {
	return &FiveHundred{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerTeam:    -1,
		roundNumber:   0,
		dealerIdx:     0,
		trumpSuit:     -1,
		highestBidder: -1,
		declarerIdx:   -1,
		jokerLeadSuit: -1,
	}
}

// NewDefaultFiveHundred returns FiveHundred with the standard 4-player team
// setup (human team 0, alternating CPU teams) and DefaultFiveHundredConfig.
// Single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFiveHundred() *FiveHundred {
	players := []*FiveHundredPlayer{
		NewFiveHundredPlayer(true, 0),
		NewFiveHundredPlayer(false, 1),
		NewFiveHundredPlayer(false, 0),
		NewFiveHundredPlayer(false, 1),
	}
	return NewFiveHundred(NewTrumpCardsFiveHundred(), players, DefaultFiveHundredConfig())
}

// Reset ゲーム初期化
func (g *FiveHundred) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [FiveHundredTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *FiveHundred) NextRound() {
	if g.phase != FiveHundredPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % FiveHundredPlayerCnt
	g.startRound()
}

// startRound ラウンドの状態を初期化して配り直す
func (g *FiveHundred) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.jokerLeadSuit = -1
	g.passed = [FiveHundredPlayerCnt]bool{}
	g.highestBid = nil
	g.highestBidder = -1
	g.contract = FiveHundredBid{}
	g.declarerIdx = -1
	g.trumpSuit = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.dealRound()
	g.phase = FiveHundredPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % FiveHundredPlayerCnt
}

// dealRound 10枚ずつ配り、残り3枚をキティにする
func (g *FiveHundred) dealRound() {
	g.trumpCards.Shuffle()
	g.kitty = nil
	for range FiveHundredHandSize {
		for j := range FiveHundredPlayerCnt {
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
func (g *FiveHundred) PlayerBid(kind FiveHundredContractKind, tricks, suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FiveHundredPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	bid := FiveHundredBid{Kind: kind, Tricks: tricks, Suit: suit}
	if !bid.valid() {
		return NewDomainError(ErrInvalidPlay, "無効なビッドです")
	}
	if g.highestBid != nil && bid.Order() <= g.highestBid.Order() {
		return NewDomainError(ErrInvalidPlay, "現在のビッドより高い必要があります")
	}
	g.applyBid(humanIdx, bid)
	return nil
}

// PlayerPass 人間プレイヤーがパスする
func (g *FiveHundred) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FiveHundredPhaseBid {
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
func (g *FiveHundred) CpuBid() {
	if g.gameEndFlag || g.phase != FiveHundredPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= FiveHundredPlayerCnt {
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
func (g *FiveHundred) applyBid(idx int, bid FiveHundredBid) {
	b := bid
	g.players[idx].SetBid(&b)
	g.highestBid = &b
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), fiveHundredBidLabel(b)), nil)
	g.advanceBid()
}

// applyPass パスを適用する
func (g *FiveHundred) applyPass(idx int) {
	g.passed[idx] = true
	g.players[idx].SetPassed(true)
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid ビッドを次へ進め、終了判定を行う
func (g *FiveHundred) advanceBid() {
	active := 0
	for i := range FiveHundredPlayerCnt {
		if !g.passed[i] {
			active++
		}
	}
	if g.highestBid != nil && active <= 1 {
		g.finalizeBid()
		return
	}
	if g.highestBid == nil && active == 0 {
		g.redeal()
		return
	}
	next := (g.bidPlayerIdx + 1) % FiveHundredPlayerCnt
	for g.passed[next] {
		next = (next + 1) % FiveHundredPlayerCnt
	}
	g.bidPlayerIdx = next
}

// redeal 全員パスした場合、同じラウンドを配り直す
func (g *FiveHundred) redeal() {
	g.appendLog(-1, "redeal", "All players passed. Redealing.", nil)
	g.passed = [FiveHundredPlayerCnt]bool{}
	g.highestBid = nil
	g.highestBidder = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.dealRound()
	g.phase = FiveHundredPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % FiveHundredPlayerCnt
}

// finalizeBid 落札を確定し、落札者にキティを渡す
func (g *FiveHundred) finalizeBid() {
	g.contract = *g.highestBid
	g.declarerIdx = g.highestBidder
	if g.contract.Kind == FiveHundredContractSuit {
		g.trumpSuit = g.contract.Suit
	} else {
		g.trumpSuit = -1
	}
	g.players[g.declarerIdx].SetIsDeclarer(true)
	for _, c := range g.kitty {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.kitty = nil
	g.appendLog(g.declarerIdx, "win_bid",
		fmt.Sprintf("%s wins the contract: %s", g.playerName(g.declarerIdx), fiveHundredBidLabel(g.contract)), nil)
	g.sortAllHands()
	g.phase = FiveHundredPhaseKittyExchange
	g.currentPlayerIdx = g.declarerIdx
}

// --- Kitty exchange ---

// PlayerExchangeKitty 人間(落札者)が3枚捨てる
func (g *FiveHundred) PlayerExchangeKitty(discardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FiveHundredPhaseKittyExchange {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doExchange(discardIndices)
}

// CpuExchange CPU(落札者)が3枚捨てる
func (g *FiveHundred) CpuExchange() {
	if g.gameEndFlag || g.phase != FiveHundredPhaseKittyExchange {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	_ = g.doExchange(g.cpuSelectDiscards(g.declarerIdx))
}

// doExchange キティ交換の共通処理
func (g *FiveHundred) doExchange(discardIndices []int) error {
	player := g.players[g.declarerIdx]
	if len(discardIndices) != FiveHundredKittySize {
		return NewDomainError(ErrInvalidCard, "3枚捨ててください")
	}
	seen := make(map[int]bool, FiveHundredKittySize)
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
func (g *FiveHundred) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.jokerLeadSuit = -1
	lead := g.declarerIdx
	if g.isSkipped(lead) {
		lead = g.nextActivePlayer(lead)
	}
	g.leadPlayerIdx = lead
	g.currentPlayerIdx = lead
	g.phase = FiveHundredPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
// jokerSuit はノートランプ契約でジョーカーをリードする際の指名スート (それ以外では無視)。
func (g *FiveHundred) PlayerPlay(cardIndex, jokerSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FiveHundredPhasePlay {
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
	g.playCard(g.currentPlayerIdx, played, jokerSuit)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行する
func (g *FiveHundred) CpuPlay() {
	if g.gameEndFlag || g.phase != FiveHundredPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	idx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	card := g.players[g.currentPlayerIdx].GetCard(idx)
	jokerSuit := -1
	if len(g.currentTrick) == 0 && g.isJoker(card) && !g.isSuitContract() {
		jokerSuit = g.cpuNominateSuit(g.currentPlayerIdx)
	}
	played := g.players[g.currentPlayerIdx].RemoveCard(idx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(g.currentPlayerIdx, played, jokerSuit)
}

// playCard カードをプレイする共通処理
func (g *FiveHundred) playCard(playerIdx int, card *Card, jokerSuit int) {
	if len(g.currentTrick) == 0 && g.isJoker(card) && !g.isSuitContract() {
		if jokerSuit >= CardDesignSpade && jokerSuit <= CardDesignDiamond {
			g.jokerLeadSuit = jokerSuit
		} else {
			g.jokerLeadSuit = g.cpuNominateSuit(playerIdx)
		}
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", g.playerName(playerIdx), fiveHundredCardLabel(card)), []*Card{card})
	if len(g.currentTrick) == g.activePlayerCount() {
		g.phase = FiveHundredPhaseTrickEnd
	} else {
		g.currentPlayerIdx = g.nextActivePlayer(g.currentPlayerIdx)
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *FiveHundred) ResolveTrick() {
	if g.phase != FiveHundredPhaseTrickEnd || len(g.currentTrick) != g.activePlayerCount() {
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
	if g.trickNumber >= FiveHundredTrickCnt {
		g.phase = FiveHundredPhaseRoundEnd
	} else {
		g.phase = FiveHundredPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *FiveHundred) NextTrick() {
	if g.phase != FiveHundredPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.jokerLeadSuit = -1
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = FiveHundredPhasePlay
}

// FiveHundredRoundResult はラウンド終了時の内訳。
//
// **「何点動いたか」を画面が言えるようにする。**ラウンド終了バナーは定型文
// しか出しておらず、成否も増減もヘッダーの数字を前後で見比べるしかなかった
// (#4809)。ScoreRound はこの結果をそのまま適用するので、表示と実際の加算が
// ずれない。
type FiveHundredRoundResult struct {
	DeclarerTeam  int
	DefenderTeam  int
	ContractValue int
	// NeedTricks は成立に必要なトリック数 (ミゼールでは 0)。
	NeedTricks int
	// DeclarerTricks はミゼールなら宣言者本人、それ以外は宣言側チームの獲得数。
	DeclarerTricks int
	DefenderTricks int
	Misere         bool
	Made           bool
	// Slam は全トリック獲得によるボーナス (250 点) が適用されたか。
	Slam bool
	// DeclarerDelta / DefenderDelta は各チームの得点増減。
	DeclarerDelta int
	DefenderDelta int
}

// GetRoundResult はラウンド終了フェーズでの内訳を返す (それ以外は nil)。
// 計算のみで状態は変えないので、ScoreRound の前に何度呼んでも安全。
func (g *FiveHundred) GetRoundResult() *FiveHundredRoundResult {
	if g.phase != FiveHundredPhaseRoundEnd || g.declarerIdx < 0 {
		return nil
	}
	return g.computeRoundResult()
}

// computeRoundResult はラウンドの得点内訳を計算する (状態は変えない)。
func (g *FiveHundred) computeRoundResult() *FiveHundredRoundResult {
	declTeam := g.players[g.declarerIdx].GetTeam()
	defTeam := 1 - declTeam
	teamTricks := [FiveHundredTeamCnt]int{}
	for _, p := range g.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}
	bidVal := g.contract.Value()
	r := &FiveHundredRoundResult{
		DeclarerTeam:   declTeam,
		DefenderTeam:   defTeam,
		ContractValue:  bidVal,
		DefenderTricks: teamTricks[defTeam],
		Misere:         g.isMisere(),
	}
	if r.Misere {
		r.DeclarerTricks = g.players[g.declarerIdx].GetTrickCount()
		r.Made = r.DeclarerTricks == 0
		if r.Made {
			r.DeclarerDelta = bidVal
		} else {
			r.DeclarerDelta = -bidVal
		}
		return r
	}
	r.DeclarerTricks = teamTricks[declTeam]
	r.NeedTricks = g.contract.Tricks
	r.Made = r.DeclarerTricks >= r.NeedTricks
	switch {
	case r.Made && r.DeclarerTricks == FiveHundredTrickCnt && bidVal < 250:
		r.Slam = true
		r.DeclarerDelta = 250 // 全トリック獲得のスラムボーナス
	case r.Made:
		r.DeclarerDelta = bidVal
	default:
		r.DeclarerDelta = -bidVal
	}
	// 守備側は獲得トリック1つにつき10点
	r.DefenderDelta = 10 * teamTricks[defTeam]
	return r
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *FiveHundred) ScoreRound() {
	if g.phase != FiveHundredPhaseRoundEnd {
		return
	}
	r := g.computeRoundResult()
	g.teamScores[r.DeclarerTeam] += r.DeclarerDelta
	g.teamScores[r.DefenderTeam] += r.DefenderDelta

	switch {
	case r.Misere && r.Made:
		g.appendLog(-1, "misere_made",
			fmt.Sprintf("Team %d makes misere! +%d", r.DeclarerTeam, r.ContractValue), nil)
	case r.Misere:
		g.appendLog(-1, "misere_failed",
			fmt.Sprintf("Team %d fails misere (%d tricks). -%d", r.DeclarerTeam, r.DeclarerTricks, r.ContractValue), nil)
	case r.Made:
		g.appendLog(-1, "contract_made",
			fmt.Sprintf("Team %d makes the contract (%d tricks). +%d", r.DeclarerTeam, r.DeclarerTricks, r.DeclarerDelta), nil)
	default:
		g.appendLog(-1, "contract_failed",
			fmt.Sprintf("Team %d is set (%d/%d tricks). -%d", r.DeclarerTeam, r.DeclarerTricks, r.NeedTricks, r.ContractValue), nil)
	}

	for ti := range FiveHundredTeamCnt {
		g.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points", ti, g.teamScores[ti]), nil)
	}
	g.checkGameEnd(r.DeclarerTeam)
}

// checkGameEnd ゲーム終了判定。先に TargetScore に到達したチームが勝利、
// -TargetScore 以下になったチームは敗北 (相手の勝利)。落札チームを優先判定する。
func (g *FiveHundred) checkGameEnd(declTeam int) {
	target := g.config.TargetScore
	other := 1 - declTeam
	switch {
	case g.teamScores[declTeam] >= target:
		g.endGame(declTeam)
	case g.teamScores[other] >= target:
		g.endGame(other)
	case g.teamScores[declTeam] <= -target:
		g.endGame(other)
	case g.teamScores[other] <= -target:
		g.endGame(declTeam)
	}
}

// endGame ゲームを終了させる
func (g *FiveHundred) endGame(team int) {
	g.gameEndFlag = true
	g.phase = FiveHundredPhaseGameEnd
	g.winnerTeam = team
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", team), nil)
}

// --- Card ranking (bowers + joker) ---

// isJoker カードがジョーカーかどうか
func (g *FiveHundred) isJoker(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// isSuitContract 切り札契約かどうか
func (g *FiveHundred) isSuitContract() bool {
	return g.contract.Kind == FiveHundredContractSuit
}

// isMisere ミゼール系契約かどうか
func (g *FiveHundred) isMisere() bool {
	return g.contract.Kind == FiveHundredContractMisere || g.contract.Kind == FiveHundredContractOpenMisere
}

// isRightBower 右バウアー (切り札スートのJ) かどうか
func (g *FiveHundred) isRightBower(c *Card) bool {
	return g.isSuitContract() && c != nil && c.GetValue() == 11 && c.GetDesign() == g.trumpSuit
}

// isLeftBower 左バウアー (同色スートのJ) かどうか
func (g *FiveHundred) isLeftBower(c *Card) bool {
	return g.isSuitContract() && c != nil && c.GetValue() == 11 && c.GetDesign() == sameColorSuit(g.trumpSuit)
}

// effectiveSuit カードの実効スートを返す。
// ジョーカー: 切り札契約では切り札スート、NT/ミゼールでは仮想スート。
// 左バウアー: 切り札スート。
func (g *FiveHundred) effectiveSuit(c *Card) int {
	if c == nil {
		return -1
	}
	if g.isJoker(c) {
		if g.isSuitContract() {
			return g.trumpSuit
		}
		return fiveHundredJokerSuit
	}
	if g.isLeftBower(c) {
		return g.trumpSuit
	}
	return c.GetDesign()
}

// cardRank トリック比較用のカードランクを返す (高い=強い)。
// ジョーカー(700) > 右バウアー(600) > 左バウアー(500) > 切り札A..9(400+) > 平A..(100+)。
func (g *FiveHundred) cardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if g.isJoker(c) {
		return 700
	}
	if g.isRightBower(c) {
		return 600
	}
	if g.isLeftBower(c) {
		return 500
	}
	base := c.GetValue()
	if base == 1 {
		base = 14 // Ace high
	}
	if g.isSuitContract() && g.effectiveSuit(c) == g.trumpSuit {
		return 400 + base
	}
	return 100 + base
}

// leadSuit 現在のトリックのリードスートを返す
func (g *FiveHundred) leadSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0].Card
	if g.isJoker(first) {
		if g.isSuitContract() {
			return g.trumpSuit
		}
		if g.jokerLeadSuit >= CardDesignSpade && g.jokerLeadSuit <= CardDesignDiamond {
			return g.jokerLeadSuit
		}
		return fiveHundredJokerSuit
	}
	return g.effectiveSuit(first)
}

// trickScore リードスートを踏まえたトリック比較値を返す (0=勝てない)
func (g *FiveHundred) trickScore(c *Card, leadSuit int) int {
	if g.isJoker(c) {
		return 700
	}
	es := g.effectiveSuit(c)
	if g.isSuitContract() && es == g.trumpSuit {
		return g.cardRank(c)
	}
	if es == leadSuit {
		return g.cardRank(c)
	}
	return 0
}

// trickWinner トリックの勝者を決定する
func (g *FiveHundred) trickWinner() int {
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
func (g *FiveHundred) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil // リードは自由
	}
	if g.isJoker(card) {
		return nil // ジョーカーはいつでも出せる
	}
	ls := g.leadSuit()
	if g.effectiveSuit(card) != ls && g.playerHasSuit(playerIdx, ls) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが実効スートのカードを持っているか (ジョーカーは除く)
func (g *FiveHundred) playerHasSuit(playerIdx, suit int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if g.isJoker(c) {
			continue
		}
		if g.effectiveSuit(c) == suit {
			return true
		}
	}
	return false
}

// --- Active player helpers (misere = declarer plays alone) ---

// activePlayerCount アクティブなプレイヤー数 (ミゼールは3)
func (g *FiveHundred) activePlayerCount() int {
	if g.isMisere() {
		return FiveHundredPlayerCnt - 1
	}
	return FiveHundredPlayerCnt
}

// skippedIdx ミゼールで除外される落札者のパートナーのインデックス (-1 = なし)
func (g *FiveHundred) skippedIdx() int {
	if !g.isMisere() || g.declarerIdx < 0 {
		return -1
	}
	return (g.declarerIdx + 2) % FiveHundredPlayerCnt
}

// isSkipped ミゼールで除外されるプレイヤーか
func (g *FiveHundred) isSkipped(idx int) bool {
	return g.isMisere() && idx == g.skippedIdx()
}

// nextActivePlayer 次のアクティブプレイヤーを返す (除外対象を飛ばす)
func (g *FiveHundred) nextActivePlayer(idx int) int {
	next := (idx + 1) % FiveHundredPlayerCnt
	if g.isSkipped(next) {
		next = (next + 1) % FiveHundredPlayerCnt
	}
	return next
}

// --- CPU AI ---

// cpuSelectBid CPUのビッド選択 (ok=false でパス)。CPUはミゼールをビッドしない。
func (g *FiveHundred) cpuSelectBid(playerIdx int) (FiveHundredBid, bool) {
	bestEst := 0
	best := FiveHundredBid{}
	found := false
	candidates := []struct {
		kind FiveHundredContractKind
		suit int
	}{
		{FiveHundredContractSuit, CardDesignSpade},
		{FiveHundredContractSuit, CardDesignClover},
		{FiveHundredContractSuit, CardDesignDiamond},
		{FiveHundredContractSuit, CardDesignHeart},
		{FiveHundredContractNoTrump, -1},
	}
	for _, c := range candidates {
		est := g.evalHand(playerIdx, c.kind, c.suit)
		if est > bestEst || (est == bestEst && found && c.kind == FiveHundredContractNoTrump) {
			bestEst = est
			best = FiveHundredBid{Kind: c.kind, Tricks: clampInt(est, 6, 10), Suit: c.suit}
			found = true
		}
	}
	threshold := 7
	switch g.config.CpuDifficulty {
	case FiveHundredCpuDifficultyEasy:
		threshold = 8
	case FiveHundredCpuDifficultyHard:
		threshold = 6
	}
	if bestEst < threshold {
		return FiveHundredBid{}, false
	}
	if g.highestBid != nil && best.Order() <= g.highestBid.Order() {
		return FiveHundredBid{}, false
	}
	return best, true
}

// evalHand 指定契約におけるおおよその獲得トリック数を見積もる
func (g *FiveHundred) evalHand(playerIdx int, kind FiveHundredContractKind, suit int) int {
	p := g.players[playerIdx]
	est := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == CardDesignJoker {
			est++
			continue
		}
		if kind == FiveHundredContractSuit {
			if c.GetValue() == 11 && c.GetDesign() == suit {
				est++ // right bower
				continue
			}
			if c.GetValue() == 11 && c.GetDesign() == sameColorSuit(suit) {
				est++ // left bower
				continue
			}
			if c.GetDesign() == suit {
				if c.GetValue() == 1 || c.GetValue() == 13 || c.GetValue() == 12 {
					est++
				}
				continue
			}
			if c.GetValue() == 1 {
				est++ // off-suit ace
			}
			continue
		}
		// no-trump
		if c.GetValue() == 1 {
			est++ // ace
		}
	}
	return est
}

// cpuSelectDiscards CPU(落札者)が捨てる3枚のインデックスを選ぶ (最弱3枚)
func (g *FiveHundred) cpuSelectDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return g.cardRank(p.GetCard(idxs[a])) < g.cardRank(p.GetCard(idxs[b]))
	})
	count := FiveHundredKittySize
	if count > n {
		count = n
	}
	return append([]int(nil), idxs[:count]...)
}

// cpuNominateSuit ジョーカーリード時にCPUが指名するスート (手札で最も多いスート)
func (g *FiveHundred) cpuNominateSuit(playerIdx int) int {
	p := g.players[playerIdx]
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if g.isJoker(c) {
			continue
		}
		counts[c.GetDesign()]++
	}
	best := CardDesignSpade
	bestCnt := -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			best = suit
		}
	}
	return best
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選ぶ
func (g *FiveHundred) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	p := g.players[playerIdx]
	// リード: 最も強いカード
	if len(g.currentTrick) == 0 {
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
	winnerTeam := g.players[g.trickWinner()].GetTeam()
	myTeam := p.GetTeam()
	if winnerTeam == myTeam {
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
func (g *FiveHundred) weakestValid(playerIdx int, valid []int) int {
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
func (g *FiveHundred) currentWinnerScore(ls int) int {
	best := 0
	for _, tc := range g.currentTrick {
		if s := g.trickScore(tc.Card, ls); s > best {
			best = s
		}
	}
	return best
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *FiveHundred) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *FiveHundred) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Hint ---

// GetHint 現在の人間の手番に対するヒントを返す
func (g *FiveHundred) GetHint() *FiveHundredHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case FiveHundredPhaseBid:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		if bid, ok := g.cpuSelectBid(humanIdx); ok {
			kind := int(bid.Kind)
			tricks := bid.Tricks
			suit := bid.Suit
			return &FiveHundredHint{BidKind: &kind, BidTricks: &tricks, BidSuit: &suit, Reason: "strategic_bid"}
		}
		pass := true
		return &FiveHundredHint{Pass: &pass, Reason: "pass_recommended"}
	case FiveHundredPhaseKittyExchange:
		if g.declarerIdx != humanIdx {
			return nil
		}
		return &FiveHundredHint{DiscardIndices: g.cpuSelectDiscards(humanIdx), Reason: "discard_weakest"}
	case FiveHundredPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuSelectPlayCard(humanIdx)
		hint := &FiveHundredHint{CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
		if len(g.currentTrick) == 0 && g.isJoker(g.players[humanIdx].GetCard(idx)) && !g.isSuitContract() {
			js := g.cpuNominateSuit(humanIdx)
			hint.JokerSuit = &js
		}
		return hint
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する
func (g *FiveHundred) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if g.isSuitContract() && g.effectiveSuit(card) == g.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}
	if g.effectiveSuit(card) == g.leadSuit() {
		return "follow_suit"
	}
	if g.isSuitContract() && g.effectiveSuit(card) == g.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *FiveHundred) GetPhase() FiveHundredPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *FiveHundred) SetPhase(phase FiveHundredPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *FiveHundred) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号取得
func (g *FiveHundred) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *FiveHundred) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *FiveHundred) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *FiveHundred) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *FiveHundred) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *FiveHundred) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *FiveHundred) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *FiveHundred) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *FiveHundred) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *FiveHundred) GetPlayer(i int) *FiveHundredPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *FiveHundred) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *FiveHundred) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッド手番インデックス取得
func (g *FiveHundred) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx ビッド手番インデックス設定 (テスト用)
func (g *FiveHundred) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *FiveHundred) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *FiveHundred) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得 (-1 = なし)
func (g *FiveHundred) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *FiveHundred) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetContractKind 契約種別取得
func (g *FiveHundred) GetContractKind() int { return int(g.contract.Kind) }

// GetContractTricks 契約トリック数取得
func (g *FiveHundred) GetContractTricks() int { return g.contract.Tricks }

// GetContractValue 契約の得点取得
func (g *FiveHundred) GetContractValue() int { return g.contract.Value() }

// SetContract 契約設定 (テスト用)
func (g *FiveHundred) SetContract(kind FiveHundredContractKind, tricks, suit int) {
	g.contract = FiveHundredBid{Kind: kind, Tricks: tricks, Suit: suit}
	if kind == FiveHundredContractSuit {
		g.trumpSuit = suit
	} else {
		g.trumpSuit = -1
	}
}

// GetDeclarerIdx 落札者インデックス取得 (-1 = 未確定)
func (g *FiveHundred) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 落札者インデックス設定 (テスト用)
func (g *FiveHundred) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetHighestBid 現在の最高ビッド取得 (nil = なし)
func (g *FiveHundred) GetHighestBid() *FiveHundredBid { return g.highestBid }

// GetHighestBidder 最高ビッダーのインデックス取得 (-1 = なし)
func (g *FiveHundred) GetHighestBidder() int { return g.highestBidder }

// GetJokerLeadSuit ジョーカーリードの指名スート取得 (-1 = なし)
func (g *FiveHundred) GetJokerLeadSuit() int { return g.jokerLeadSuit }

// GetKitty キティ取得
func (g *FiveHundred) GetKitty() []*Card { return g.kitty }

// GetTeamScore チームスコア取得
func (g *FiveHundred) GetTeamScore(team int) int {
	if team < 0 || team >= FiveHundredTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *FiveHundred) SetTeamScore(team, score int) {
	if team >= 0 && team < FiveHundredTeamCnt {
		g.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *FiveHundred) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (g *FiveHundred) IsHumanBidTurn() bool {
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *FiveHundred) GetConfig() FiveHundredConfig { return g.config }

// SetConfig 設定変更
func (g *FiveHundred) SetConfig(cfg FiveHundredConfig) { g.config = cfg }

// CardRankPublic カードランク取得 (テスト用)
func (g *FiveHundred) CardRankPublic(card *Card) int { return g.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用)
func (g *FiveHundred) EffectiveSuitPublic(card *Card) int { return g.effectiveSuit(card) }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *FiveHundred) sortAllHands() {
	for _, p := range g.players {
		g.sortHand(p)
	}
}

// sortHand プレイヤーの手札を実効スート→ランク順にソートする
func (g *FiveHundred) sortHand(p *FiveHundredPlayer) {
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

// playerName プレイヤー名を返す
func (g *FiveHundred) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// fiveHundredCardLabel カードのログ表示文字列 (ジョーカー対応)
func fiveHundredCardLabel(c *Card) string {
	if c == nil {
		return "??"
	}
	if c.GetDesign() == CardDesignJoker {
		return "JK"
	}
	return cardStr(c)
}

// fiveHundredBidLabel ビッドのログ表示文字列
func fiveHundredBidLabel(b FiveHundredBid) string {
	switch b.Kind {
	case FiveHundredContractSuit:
		return fmt.Sprintf("%d%s (%d)", b.Tricks, suitStr(b.Suit), b.Value())
	case FiveHundredContractNoTrump:
		return fmt.Sprintf("%dNT (%d)", b.Tricks, b.Value())
	case FiveHundredContractMisere:
		return fmt.Sprintf("Misere (%d)", b.Value())
	case FiveHundredContractOpenMisere:
		return fmt.Sprintf("Open Misere (%d)", b.Value())
	}
	return "Pass"
}

// clampInt 整数を [min,max] に収める
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// --- JSON ---

// fiveHundredJSON is the JSON wire format for FiveHundred.
type fiveHundredJSON struct {
	TrumpCards       *TrumpCards                `json:"tc"`
	Players          []*FiveHundredPlayer       `json:"ps"`
	Config           FiveHundredConfig          `json:"cf"`
	Phase            FiveHundredPhase           `json:"ph"`
	RoundNumber      int                        `json:"rn"`
	TrickNumber      int                        `json:"tn"`
	CurrentPlayerIdx int                        `json:"ci"`
	CurrentTrick     []*TrickCard               `json:"ct"`
	DealerIdx        int                        `json:"di"`
	Kitty            []*Card                    `json:"kt"`
	LeadPlayerIdx    int                        `json:"li"`
	JokerLeadSuit    int                        `json:"jl"`
	BidPlayerIdx     int                        `json:"bi"`
	Passed           [FiveHundredPlayerCnt]bool `json:"pd"`
	HighestBid       *FiveHundredBid            `json:"hb"`
	HighestBidder    int                        `json:"hr"`
	Contract         FiveHundredBid             `json:"cn"`
	DeclarerIdx      int                        `json:"dc"`
	TrumpSuit        int                        `json:"ts"`
	TeamScores       [FiveHundredTeamCnt]int    `json:"sc"`
	GameEndFlag      bool                       `json:"ge"`
	WinnerTeam       int                        `json:"wt"`
	ActionLog        []*ActionLogEntry          `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *FiveHundred) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiveHundredJSON{
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
		LeadPlayerIdx:    g.leadPlayerIdx,
		JokerLeadSuit:    g.jokerLeadSuit,
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

// fiveHundredMaxSliceLen caps slice sizes during deserialisation.
const fiveHundredMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *FiveHundred) UnmarshalJSON(data []byte) error {
	var j fiveHundredJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > fiveHundredMaxSliceLen || len(j.CurrentTrick) > fiveHundredMaxSliceLen ||
		len(j.Kitty) > fiveHundredMaxSliceLen || len(j.ActionLog) > fiveHundredMaxSliceLen {
		return fmt.Errorf("fivehundred: input array exceeds maximum allowed size")
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsFiveHundred()
	}
	g.players = j.Players
	if len(g.players) != FiveHundredPlayerCnt {
		return fmt.Errorf("fivehundred: invalid player count: %d", len(g.players))
	}
	for _, p := range g.players {
		if p == nil {
			return fmt.Errorf("fivehundred: player is nil")
		}
	}
	g.config = j.Config
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("fivehundred: invalid config: %w", err)
	}
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.dealerIdx = j.DealerIdx
	g.kitty = j.Kitty
	if g.kitty == nil {
		g.kitty = make([]*Card, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.jokerLeadSuit = j.JokerLeadSuit
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
	return nil
}
