//go:build !js || !wasm || extra

// Package domain バウエルンシュナプセン (Bauernschnapsen) のドメインモデル。
//
// Bauernschnapsen はドイツ・シュヴァーベン地方の 4 人 2 チーム制トリックテイキング
// ゲームで、Schnapsen / 66 ファミリーに属する point-trick ゲーム。
//
// デッキは 48 枚 = A,10,K,Q,J,7 の 24 枚 (6 ランク × 4 スート) を 2 組重ねたもの。
// 各プレイヤーに 5 枚ずつ配り (計 20 枚)、次の 1 枚を表向きの切り札表示カードと
// して場に置く (Schnapsen 方式、山札の最後の引き札として残る)。残り 27 枚が山札。
//
// 山札がある間 (第1フェーズ) はマストフォローが無く、何を出してもよい。
// トリック後、勝者から順に各プレイヤーが山札から 1 枚ずつ引いて 5 枚に補充する。
// 自分のリード番で同スートの K と Q を両方持っていれば「マリアージュ」を宣言して
// チームへボーナス点 (20 点、切り札なら 40 点) を獲得し、その K か Q をリードする。
//
// 山札が尽きると第2フェーズに移行し、マストフォロー (フォローできる場合は同スート、
// 無ければ切り札) が義務付けられる。スート内ランクは A>10>K>Q>J>7。同一カード
// (同スート・同ランク) が同一トリックに出た場合は先に出した側が勝つ。
//
// 全カードを出し終えたら各チームのカード点 (A=11,10=10,K=4,Q=3,J=2,7=0) と
// マリアージュ点を集計し、累計に加算する。先に 101 点に達したチームが勝利。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// BauernschnapsenNoTrump は「切り札スート未定 / 切り札なし」を表す。
//
// 切り札は配った札ではなく**契約の宣言**で決まるので、配り終えた時点では
// まだ決まっていない。Bettel (トリックを 1 つも取らない契約) では最後まで
// 切り札なしのまま進む。
const BauernschnapsenNoTrump = -1

// BauernschnapsenPlayerCnt バウエルンシュナプセンのプレイヤー数 (4人固定)
const BauernschnapsenPlayerCnt = 4

// BauernschnapsenTeamCnt チーム数
const BauernschnapsenTeamCnt = 2

// BauernschnapsenHandSize 各プレイヤーの手札枚数 (山札がある間は補充される)
const BauernschnapsenHandSize = 5

// BauernschnapsenMarriageBonus 通常スートのマリアージュ (K+Q) ボーナス
const BauernschnapsenMarriageBonus = 20

// BauernschnapsenRoyalMarriageBonus 切り札スートのマリアージュ (ロイヤルマリアージュ) ボーナス
const BauernschnapsenRoyalMarriageBonus = 40

// BauernschnapsenRoundCardPointsTotal 1ラウンドのカード合計点数。
//
// (A11 + 10:10 + K4 + Q3 + J2) × 4 スート = **120**。シュナプセン / 66 と同じ。
//
// クローン元のガイゲルは 7 を含む 24 枚を 2 組重ねるので 240 点だった。
// 過半は 61 点 (120 の半分を超える最小値)。
const BauernschnapsenRoundCardPointsTotal = 120

// BauernschnapsenPhase ゲームフェーズ
type BauernschnapsenPhase int

// Bauernschnapsenのフェーズ定数
const (
	// BauernschnapsenPhaseContract 契約宣言フェーズ。
	//
	// **クローン元のガイゲルには無い。** ガイゲルは表向きの切り札表示カードで
	// 切り札が決まるが、こちらは 20 枚を配り切るので表向きの札が無く、
	// 切り札も遊び方も**宣言で決める**。
	BauernschnapsenPhaseContract BauernschnapsenPhase = iota
	// BauernschnapsenPhasePlay トリックプレイフェーズ
	BauernschnapsenPhasePlay
	// BauernschnapsenPhaseTrickEnd トリック終了フェーズ
	BauernschnapsenPhaseTrickEnd
	// BauernschnapsenPhaseRoundEnd ラウンド終了フェーズ
	BauernschnapsenPhaseRoundEnd
	// BauernschnapsenPhaseGameEnd ゲーム終了フェーズ
	BauernschnapsenPhaseGameEnd
)

// BauernschnapsenContract は宣言できる契約の種類。
type BauernschnapsenContract int

// 契約の定数。値が大きいほど強い宣言 (高い契約が採用される)。
const (
	// BauernschnapsenContractNone 未宣言
	BauernschnapsenContractNone BauernschnapsenContract = iota
	// BauernschnapsenContractRufer 通常契約。切り札スートを指定し、
	// 自チームがカード点の過半 (61 点以上、120 の半分超) を取れば成功。
	BauernschnapsenContractRufer
	// BauernschnapsenContractFarbenzwang 同スート縛り。切り札を指定し、
	// **相手にトリックを 1 つも渡さない**ことを目指す高い契約。
	BauernschnapsenContractFarbenzwang
	// BauernschnapsenContractBettel トリックを 1 つも取らない契約。
	// **切り札なし**で、1 つでも取ったら失敗。
	BauernschnapsenContractBettel
)

// BauernschnapsenContractValue は契約の得点倍率を返す。
// 難しい契約ほど成功時の見返りが大きい。
func BauernschnapsenContractValue(c BauernschnapsenContract) int {
	switch c {
	case BauernschnapsenContractRufer:
		return 1
	case BauernschnapsenContractFarbenzwang:
		return 3
	case BauernschnapsenContractBettel:
		return 2
	default:
		return 0
	}
}

// BauernschnapsenHint ヒント情報
type BauernschnapsenHint struct {
	CardIndex  *int   // 推奨カードインデックス
	Reason     string // ヒント理由キー
	IsMarriage bool   // 推奨アクションがマリアージュ宣言かどうか
}

// Bauernschnapsen バウエルンシュナプセンゲームクラス
type Bauernschnapsen struct {
	trumpCards       *TrumpCards
	players          []*BauernschnapsenPlayer
	config           BauernschnapsenConfig
	phase            BauernschnapsenPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpSuit        int // 切り札スート。BauernschnapsenNoTrump = 未定 / 切り札なし

	// contract は採用された契約、declarerIdx はそれを宣言した席。
	// contractBids は各席の宣言 (契約フェーズの入力)。
	contract      BauernschnapsenContract
	declarerIdx   int
	contractBids  [BauernschnapsenPlayerCnt]BauernschnapsenContract
	leadPlayerIdx int
	dealerIdx     int
	teamScores    [BauernschnapsenTeamCnt]int // 累計チームスコア
	roundPoints   [BauernschnapsenTeamCnt]int // 当ラウンドのカード点 (トリック獲得分)
	roundMarriage [BauernschnapsenTeamCnt]int // 当ラウンドのマリアージュ点
	// roundTricks は当ラウンドで各チームが取ったトリック数。
	// **Bettel と Farbenzwang は点ではなくトリック数で成否が決まる**ので、
	// カード点とは別に数える。
	roundTricks [BauernschnapsenTeamCnt]int
	// seatTricks は席別の獲得トリック数。ベテルは**宣言者ひとり**が
	// 1 トリックも取らない契約なので、チーム合計では判定できない。
	seatTricks       [BauernschnapsenPlayerCnt]int
	marriageDeclared [CardDesignMax + 1]bool // suit -> 当ラウンドで宣言済か
	lastTrickWinner  int
	gameEndFlag      bool
	winnerTeam       int // -1: 未確定
	actionLogBase
}

// NewBauernschnapsen コンストラクタ
func NewBauernschnapsen(trumpCards *TrumpCards, players []*BauernschnapsenPlayer, config BauernschnapsenConfig) *Bauernschnapsen {
	return &Bauernschnapsen{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerTeam:      -1,
		roundNumber:     0,
		dealerIdx:       0,
		lastTrickWinner: -1,
	}
}

// NewDefaultBauernschnapsen 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)
// と DefaultBauernschnapsenConfig を組み合わせたデフォルト構築。CUI/Web/Worker 共通の SSoT。
func NewDefaultBauernschnapsen() *Bauernschnapsen {
	players := []*BauernschnapsenPlayer{
		NewBauernschnapsenPlayer(true, 0),
		NewBauernschnapsenPlayer(false, 1),
		NewBauernschnapsenPlayer(false, 0),
		NewBauernschnapsenPlayer(false, 1),
	}
	return NewBauernschnapsen(newBauernschnapsenDeck(), players, DefaultBauernschnapsenConfig())
}

// newBauernschnapsenDeck バウエルンシュナプセン用 20 枚デッキを生成する。
//
// A,10,K,Q,J (値: 1,10,13,12,11) × 4 スート = 20 枚。**シュナプセンと同じ**。
//
// クローン元のガイゲルは 7 を含む 24 枚を 2 組重ねた 48 枚デッキだが、
// バウエルンシュナプセンは 20 枚 1 組。**4 人に 5 枚ずつ配ると 0 枚残る**ので、
// ガイゲルにあった山札 (talon) からの補充がそもそも成立しない。
func newBauernschnapsenDeck() *TrumpCards {
	bauernschnapsenValues := []int{1, 10, 13, 12, 11} // A,10,K,Q,J
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(bauernschnapsenValues) * len(suits) // 20

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range bauernschnapsenValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// BauernschnapsenCardPoints カードの得点を返す (A=11,10=10,K=4,Q=3,J=2,7=0)。
// SchnapsenCardPoints と同一の値だが、schnapsen は別ワーカー (solo) のビルドタグ
// 配下にあり extra ワーカーからは参照できないため、ここで同じ switch を再実装する。
// グローバルマップを避けて全 Cloudflare Worker WASM バイナリのサイズを抑える。
func BauernschnapsenCardPoints(c *Card) int {
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
	default: // 7
		return 0
	}
}

// BauernschnapsenRankOrder カードのスート内順位を返す (大きいほど強い; A>10>K>Q>J>7)。
func BauernschnapsenRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 6
	case 10: // 10
		return 5
	case 13: // K
		return 4
	case 12: // Q
		return 3
	case 11: // J
		return 2
	case 7: // 7
		return 1
	default:
		return 0
	}
}

// Reset ゲーム初期化
func (g *Bauernschnapsen) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.dealerIdx = 0
	g.teamScores = [BauernschnapsenTeamCnt]int{}
	g.actionLog = nil
	g.trumpSuit = BauernschnapsenNoTrump
	g.contract = BauernschnapsenContractNone
	g.declarerIdx = -1
	g.contractBids = [BauernschnapsenPlayerCnt]BauernschnapsenContract{}

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// NextRound 次のラウンドを開始する
func (g *Bauernschnapsen) NextRound() {
	if g.phase != BauernschnapsenPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % BauernschnapsenPlayerCnt
	g.trickNumber = 0
	g.trumpSuit = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// beginRound ラウンドの初期処理 (配布 + プレイフェーズ突入)
func (g *Bauernschnapsen) beginRound() {
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.roundPoints = [BauernschnapsenTeamCnt]int{}
	g.roundMarriage = [BauernschnapsenTeamCnt]int{}
	g.marriageDeclared = [CardDesignMax + 1]bool{}
	g.lastTrickWinner = -1
	g.roundTricks = [BauernschnapsenTeamCnt]int{}
	g.seatTricks = [BauernschnapsenPlayerCnt]int{}

	g.trumpCards.Shuffle()
	g.dealInitial()
	g.sortAllHands()
	g.startContractPhase()
}

// dealInitial 各プレイヤーに 5 枚配る。**20 枚を配り切るので山札は残らない。**
//
// クローン元のガイゲルは 48 枚デッキで、配った残りを山札にし、その最後の 1 枚を
// 表向きの切り札表示カードとして置いていた。バウエルンシュナプセンは 20 枚 ×
// 4 人 × 5 枚でちょうど尽きるので、**山札も切り札表示カードも存在しない**。
// 切り札は表向きの札ではなく**契約の宣言**で決まる (declareContract を参照)。
//
// issue の仕様 1 は「残りをストックとして中央に置く」と書いているが、
// 20 - 4×5 = 0 なので置く札が無い。
func (g *Bauernschnapsen) dealInitial() {
	for range BauernschnapsenHandSize {
		for i := range BauernschnapsenPlayerCnt {
			player := g.players[(g.dealerIdx+1+i)%BauernschnapsenPlayerCnt]
			if c := g.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	// 切り札はまだ決まらない。契約の宣言で決める。
	g.trumpSuit = BauernschnapsenNoTrump
	g.appendLog(-1, "deal", "dealt 5 cards each; no talon", nil)
}

// startContractPhase 契約フェーズ開始: ディーラーの左隣から宣言する。
//
// **切り札は配った札では決まらない。** 20 枚を配り切るので表向きの札が無く、
// 誰かが契約を宣言して初めて切り札 (または「切り札なし」) が決まる。
func (g *Bauernschnapsen) startContractPhase() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.contract = BauernschnapsenContractNone
	g.declarerIdx = -1
	g.contractBids = [BauernschnapsenPlayerCnt]BauernschnapsenContract{}
	g.leadPlayerIdx = (g.dealerIdx + 1) % BauernschnapsenPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = BauernschnapsenPhaseContract
}

// DeclareContract は席 playerIdx の宣言を受け取る。
//
// 全員が宣言し終えたら、**一番高い契約**を採用してプレイフェーズへ進む。
// 同じ高さが並んだらディーラーの左隣に近い席が優先 (宣言順が早い方)。
// 誰も宣言しなければ既定の通常契約でディーラーの左隣が declarer になる。
func (g *Bauernschnapsen) DeclareContract(playerIdx int, c BauernschnapsenContract, trumpSuit int) error {
	if g.phase != BauernschnapsenPhaseContract {
		return NewDomainError(ErrWrongPhase, "契約フェーズではありません")
	}
	if playerIdx != g.currentPlayerIdx {
		return NewDomainError(ErrInvalidPlay, "あなたの手番ではありません")
	}
	if c < BauernschnapsenContractNone || c > BauernschnapsenContractBettel {
		return NewDomainError(ErrInvalidCard, "その契約は宣言できません")
	}
	// **切り札を要る契約はスートも要る。** Bettel は切り札なしなので受け取らない。
	if c == BauernschnapsenContractRufer || c == BauernschnapsenContractFarbenzwang {
		if trumpSuit < CardDesignSpade || trumpSuit > CardDesignMax {
			return NewDomainError(ErrInvalidCard, "切り札スートを指定してください")
		}
	}

	g.contractBids[playerIdx] = c
	if c > g.contract {
		g.contract = c
		g.declarerIdx = playerIdx
		if c == BauernschnapsenContractBettel {
			g.trumpSuit = BauernschnapsenNoTrump
		} else {
			g.trumpSuit = trumpSuit
		}
	}
	g.appendLog(playerIdx, "contract", fmt.Sprintf("declared %d", c), nil)

	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % BauernschnapsenPlayerCnt
	if g.currentPlayerIdx == g.leadPlayerIdx {
		g.finishContractPhase()
	}
	return nil
}

// finishContractPhase は採用された契約を確定してプレイへ進む。
func (g *Bauernschnapsen) finishContractPhase() {
	if g.contract == BauernschnapsenContractNone {
		// **誰も宣言しなければ既定の通常契約。** 宣言なしで進むと切り札が
		// 未定のままトリックが始まり、比較が壊れる。
		g.contract = BauernschnapsenContractRufer
		g.declarerIdx = g.leadPlayerIdx
		g.trumpSuit = g.pickDefaultTrump(g.declarerIdx)
	}
	g.appendLog(g.declarerIdx, "contractFinal",
		fmt.Sprintf("contract %d, trump %d", g.contract, g.trumpSuit), nil)
	g.startPlayPhase()
}

// pickDefaultTrump は手札で一番枚数の多いスートを返す。
func (g *Bauernschnapsen) pickDefaultTrump(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			counts[c.GetDesign()]++
		}
	}
	best, bestSuit := -1, CardDesignSpade
	for suit, n := range counts {
		if n > best {
			best, bestSuit = n, suit
		}
	}
	return bestSuit
}

// CpuDeclareContract は CPU 席の宣言を進める。
//
// 判断はごく単純にしてある: 切り札候補のスートが 3 枚以上あれば通常契約、
// 点札 (A/10) を 1 枚も持たなければ Bettel (トリックを取らない契約) を狙い、
// それ以外は宣言しない。**Farbenzwang は宣言しない** —— 相手に 1 トリックも
// 渡さない契約なので、CPU の手筋では達成できず、宣言すると失点にしかならない。
func (g *Bauernschnapsen) CpuDeclareContract() {
	if g.phase != BauernschnapsenPhaseContract {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	suit := g.pickDefaultTrump(idx)
	best, pointCards := 0, 0
	p := g.players[idx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if c.GetDesign() == suit {
			best++
		}
		if v := c.GetValue(); v == 1 || v == 10 {
			pointCards++
		}
	}
	switch {
	case pointCards == 0:
		_ = g.DeclareContract(idx, BauernschnapsenContractBettel, BauernschnapsenNoTrump)
	case best >= 3:
		_ = g.DeclareContract(idx, BauernschnapsenContractRufer, suit)
	default:
		_ = g.DeclareContract(idx, BauernschnapsenContractNone, BauernschnapsenNoTrump)
	}
}

// IsHumanContractTurn は人間 (席 0) の宣言待ちかを返す。
func (g *Bauernschnapsen) IsHumanContractTurn() bool {
	return g.phase == BauernschnapsenPhaseContract && g.currentPlayerIdx == 0
}

// GetContract は採用された契約を返す。
func (g *Bauernschnapsen) GetContract() BauernschnapsenContract { return g.contract }

// GetDeclarerIdx は契約を宣言した席を返す (-1 = 未確定)。
func (g *Bauernschnapsen) GetDeclarerIdx() int { return g.declarerIdx }

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (g *Bauernschnapsen) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % BauernschnapsenPlayerCnt
	if g.contract == BauernschnapsenContractBettel && g.declarerIdx >= 0 {
		g.leadPlayerIdx = g.declarerIdx
	}
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = BauernschnapsenPhasePlay
}

// --- Play actions ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Bauernschnapsen) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BauernschnapsenPhasePlay {
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

// PlayerDeclareMarriage 人間プレイヤーがマリアージュ (K+Q 同スート) を宣言し、
// 指定したカード (その K か Q) をリードする。リード番でのみ有効。
func (g *Bauernschnapsen) PlayerDeclareMarriage(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BauernschnapsenPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.declareMarriage(g.currentPlayerIdx, cardIndex)
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する
func (g *Bauernschnapsen) CpuPlay() {
	if g.gameEndFlag || g.phase != BauernschnapsenPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}

	// リード番で有益なマリアージュがあれば宣言してリードする
	if len(g.currentTrick) == 0 {
		if cardIdx, ok := g.cpuChooseMarriage(idx); ok {
			_ = g.declareMarriage(idx, cardIdx)
			return
		}
	}

	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// declareMarriage マリアージュを宣言してチームへボーナス加点し、指定の K/Q をリードする共通処理。
// 同一スートのマリアージュは 1 ラウンドにつき 1 回のみ宣言できる (先着優先)。
func (g *Bauernschnapsen) declareMarriage(playerIdx, cardIndex int) error {
	if len(g.currentTrick) != 0 {
		return NewDomainError(ErrInvalidPlay, "マリアージュはリード時のみ宣言できます")
	}
	player := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if !g.isMarriageStarter(player, card) {
		return NewDomainError(ErrInvalidPlay, "そのカードでマリアージュは宣言できません")
	}

	suit := card.GetDesign()
	bonus := BauernschnapsenMarriageBonus
	if suit == g.trumpSuit {
		bonus = BauernschnapsenRoyalMarriageBonus
	}
	team := player.GetTeam()
	g.marriageDeclared[suit] = true
	g.roundMarriage[team] += bonus
	g.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares marriage in %s (+%d for team %d)",
			playerName(g.players, playerIdx), suitStr(suit), bonus, team), nil)

	played := player.RemoveCard(cardIndex)
	g.playCard(playerIdx, played)
	return nil
}

// isMarriageStarter card が「未宣言スートの K か Q で相方を手札に持つ」かを判定する。
func (g *Bauernschnapsen) isMarriageStarter(player *BauernschnapsenPlayer, card *Card) bool {
	// **Bettel にマリアージュは無い。** 切り札を持たず、1 トリックも取らない
	// ことだけを競う契約なので、キング+クイーンの宣言は意味を持たない。
	if g.contract == BauernschnapsenContractBettel {
		return false
	}
	v := card.GetValue()
	if v != 12 && v != 13 {
		return false
	}
	suit := card.GetDesign()
	if g.marriageDeclared[suit] {
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
func (g *Bauernschnapsen) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)),
		[]*Card{card})

	if len(g.currentTrick) == BauernschnapsenPlayerCnt {
		g.phase = BauernschnapsenPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % BauernschnapsenPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *Bauernschnapsen) ResolveTrick() {
	if g.phase != BauernschnapsenPhaseTrickEnd || len(g.currentTrick) != BauernschnapsenPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	trickPoints := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += BauernschnapsenCardPoints(tc.Card)
	}

	g.players[winnerIdx].AddTrick(trickCards)
	g.roundPoints[g.players[winnerIdx].GetTeam()] += trickPoints
	g.roundTricks[g.players[winnerIdx].GetTeam()]++
	g.seatTricks[winnerIdx]++

	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", playerName(g.players, winnerIdx), g.trickNumber, trickPoints),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	g.lastTrickWinner = winnerIdx
	g.phase = BauernschnapsenPhaseTrickEnd
}

// NextTrick 次のトリックを開始する。第1フェーズ (山札あり) では勝者から順に補充する。
// 全カードが尽きたらラウンド終了処理を実行する。
func (g *Bauernschnapsen) NextTrick() {
	if g.phase != BauernschnapsenPhaseTrickEnd {
		return
	}

	if g.allHandsEmpty() {
		g.phase = BauernschnapsenPhaseRoundEnd
		return
	}

	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = BauernschnapsenPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *Bauernschnapsen) ScoreRound() {
	if g.phase != BauernschnapsenPhaseRoundEnd {
		return
	}

	// **成否は契約ごとに条件が違う。** クローン元のガイゲルはカード点を
	// そのまま足すだけだったが、こちらは宣言した契約を達成できたかで決まる。
	declarerTeam := 0
	if g.declarerIdx >= 0 {
		declarerTeam = g.players[g.declarerIdx].GetTeam()
	}
	made := g.contractMade(declarerTeam)
	value := BauernschnapsenContractValue(g.contract)

	for ti := range BauernschnapsenTeamCnt {
		var gained int
		switch {
		case ti == declarerTeam && made:
			gained = value
		case ti != declarerTeam && !made:
			// 宣言側が落とせば相手チームが取る。
			gained = value
		}
		g.teamScores[ti] += gained
		g.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d card pts, %d marriage, %d tricks -> +%d (total %d)",
				ti, g.roundPoints[ti], g.roundMarriage[ti], g.roundTricks[ti],
				gained, g.teamScores[ti]), nil)
	}

	g.checkGameEnd()
}

// contractMade は宣言側チームが契約を達成したかを返す。
//
// 契約ごとに条件がまったく違う:
//
//   - Rufer: カード点 (マリアージュ込み) の過半を取る
//   - Farbenzwang: **相手に 1 トリックも渡さない**
//   - Bettel: **自分が 1 トリックも取らない** (切り札なし)
func (g *Bauernschnapsen) contractMade(declarerTeam int) bool {
	other := 1 - declarerTeam
	switch g.contract {
	case BauernschnapsenContractBettel:
		// **チーム合計ではなく宣言者ひとりの席。** ベテルは本来ひとりで
		// 戦う契約で、追従必須の 4 人卓ではパートナーの分まで 0 に抑える
		// のは構造的にほぼ不可能になる。
		if g.declarerIdx < 0 {
			return false
		}
		return g.seatTricks[g.declarerIdx] == 0
	case BauernschnapsenContractFarbenzwang:
		return g.roundTricks[other] == 0
	default:
		mine := g.roundPoints[declarerTeam] + g.roundMarriage[declarerTeam]
		return mine*2 > BauernschnapsenRoundCardPointsTotal
	}
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (g *Bauernschnapsen) allHandsEmpty() bool {
	return allHandsEmpty(g.players)
}

// IsEndgame は常に true。**このゲームは最初から「終盤」**。
//
// クローン元のガイゲルは山札がある間は自由出しで、尽きてから追従必須の
// 第 2 フェーズに入る二相構造だった。バウエルンシュナプセンは 20 枚を配り切って
// 山札が無いので、**1 トリック目から追従必須**。共有の validateEndgameFollow を
// そのまま使うために、この述語で「常に終盤」と答える。
func (g *Bauernschnapsen) IsEndgame() bool { return true }

// --- Validation / follow rules ---

// validatePlay カードのプレイがルール上有効かを検証する。
// リード時は常に有効。**追随時は必ずマストフォロー** (山札が無いので
// クローン元のような「自由に出せる前半」は存在しない)。
func (g *Bauernschnapsen) validatePlay(playerIdx int, card *Card) error {
	return validateEndgameFollow(g.currentTrick, g, playerIdx, card)
}

// cardSatisfiesFollow 追随時に card が合法かを返す。
// 1) リードスートを持つ: そのスートを出す。
// 2) リードスートを持たないが切り札を持つ: 切り札のみ可。
// 3) どちらも持たない: 任意。
func (g *Bauernschnapsen) cardSatisfiesFollow(playerIdx int, card *Card) bool {
	player := g.players[playerIdx]
	leadSuit := g.currentTrick[0].Card.GetDesign()

	if bauernschnapsenPlayerHasSuit(player, leadSuit) {
		return card.GetDesign() == leadSuit
	}
	if bauernschnapsenPlayerHasSuit(player, g.trumpSuit) {
		return card.GetDesign() == g.trumpSuit
	}
	return true
}

// legalPlayIndices validatePlay を満たすカードのインデックス集合を返す。
func (g *Bauernschnapsen) legalPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// bauernschnapsenPlayerHasSuit プレイヤーが指定スートのカードを持つか
func bauernschnapsenPlayerHasSuit(player *BauernschnapsenPlayer, suit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// trickWinner 現在のトリックの勝者インデックスを決定する。
// 同一カード (同スート・同ランク) が出た場合は先に出した側が勝つ (bauernschnapsenBeats は厳密 >)。
func (g *Bauernschnapsen) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card

	for _, tc := range g.currentTrick[1:] {
		if bauernschnapsenBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// bauernschnapsenBeats challenger が currentBest に厳密に勝つかを判定する。
// ・両者がトランプ: ランクの高い方が勝つ
// ・challenger のみトランプ: challenger が勝つ
// ・両者とも非トランプかつ同じリードスート: ランクの高い方が勝つ
// ・両者とも非トランプで challenger がリードスート以外: challenger は勝てない
// ランク同値 (ダブルデッキの同一カード) では false を返すため、先に出した方が勝つ。
func bauernschnapsenBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit

	switch {
	case cIsTrump && bIsTrump:
		return BauernschnapsenRankOrder(challenger) > BauernschnapsenRankOrder(currentBest)
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
	return BauernschnapsenRankOrder(challenger) > BauernschnapsenRankOrder(currentBest)
}

// --- Game end ---

func (g *Bauernschnapsen) checkGameEnd() {
	for ti := range BauernschnapsenTeamCnt {
		if g.teamScores[ti] >= g.config.TargetScore {
			g.gameEndFlag = true
			g.phase = BauernschnapsenPhaseGameEnd
			if g.teamScores[0] >= g.teamScores[1] {
				g.winnerTeam = 0
			} else {
				g.winnerTeam = 1
			}
			g.appendLog(-1, "game_end",
				fmt.Sprintf("Team %d wins the game!", g.winnerTeam), nil)
			return
		}
	}
}

// --- State getters / setters ---

// GetSeatTricks 当ラウンドで席 idx が取ったトリック数。
// ベテルの成否はこの数で決まる。
func (g *Bauernschnapsen) GetSeatTricks(idx int) int {
	if idx < 0 || idx >= BauernschnapsenPlayerCnt {
		return 0
	}
	return g.seatTricks[idx]
}

// GetRoundTricks 当ラウンドでチーム team が取ったトリック数。
// ベテル / 同スート縛りの成否はこの数で決まる。
func (g *Bauernschnapsen) GetRoundTricks(team int) int {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return 0
	}
	return g.roundTricks[team]
}

// GetPhase 現在のフェーズ取得
func (g *Bauernschnapsen) GetPhase() BauernschnapsenPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Bauernschnapsen) SetPhase(p BauernschnapsenPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号取得
func (g *Bauernschnapsen) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Bauernschnapsen) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Bauernschnapsen) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Bauernschnapsen) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Bauernschnapsen) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Bauernschnapsen) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Bauernschnapsen) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Bauernschnapsen) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetTrumpSuit 切り札スート取得
func (g *Bauernschnapsen) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Bauernschnapsen) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Bauernschnapsen) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Bauernschnapsen) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Bauernschnapsen) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Bauernschnapsen) GetPlayer(i int) *BauernschnapsenPlayer {
	return getPlayer(g.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Bauernschnapsen) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Bauernschnapsen) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Bauernschnapsen) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Bauernschnapsen) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTeamScore チームスコア取得
func (g *Bauernschnapsen) GetTeamScore(team int) int {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *Bauernschnapsen) SetTeamScore(team, score int) {
	if team >= 0 && team < BauernschnapsenTeamCnt {
		g.teamScores[team] = score
	}
}

// GetRoundPoints 当ラウンドのチーム別カード点数取得
func (g *Bauernschnapsen) GetRoundPoints(team int) int {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return 0
	}
	return g.roundPoints[team]
}

// GetRoundMarriagePoints 当ラウンドのマリアージュ得点取得
func (g *Bauernschnapsen) GetRoundMarriagePoints(team int) int {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return 0
	}
	return g.roundMarriage[team]
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Bauernschnapsen) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Bauernschnapsen) GetConfig() BauernschnapsenConfig { return g.config }

// SetConfig 設定変更
func (g *Bauernschnapsen) SetConfig(cfg BauernschnapsenConfig) { g.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// 第1フェーズは制約なし。第2フェーズはマストフォローを適用する。
func (g *Bauernschnapsen) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.legalPlayIndices(playerIdx)
}

// GetMarriageIndices リード番でマリアージュ宣言を開始できるカード (未宣言スートの
// K または Q で、相方を手札に持つもの) のインデックスを返す。
func (g *Bauernschnapsen) GetMarriageIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if len(g.currentTrick) != 0 || g.currentPlayerIdx != playerIdx {
		return nil
	}
	p := g.players[playerIdx]
	out := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if g.isMarriageStarter(p, p.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Bauernschnapsen) CardRankPublic(card *Card) int { return BauernschnapsenRankOrder(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Bauernschnapsen) CardPointsPublic(card *Card) int { return BauernschnapsenCardPoints(card) }

// --- Hints ---

// GetHint 人間プレイヤー (idx 0) へのヒントを取得する
func (g *Bauernschnapsen) GetHint() *BauernschnapsenHint {
	if g.phase != BauernschnapsenPhasePlay || g.currentPlayerIdx != 0 {
		return nil
	}
	humanIdx := 0
	if g.players[humanIdx].GetCardsSize() == 0 {
		return nil
	}
	if len(g.currentTrick) == 0 {
		if idx, ok := g.cpuChooseMarriage(humanIdx); ok {
			i := idx
			return &BauernschnapsenHint{CardIndex: &i, Reason: "marriage", IsMarriage: true}
		}
	}
	idx := g.cpuSelectPlayCard(humanIdx)
	return &BauernschnapsenHint{CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
}

// playHintReason ヒント理由キーを判定する
func (g *Bauernschnapsen) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	pts := BauernschnapsenCardPoints(card)
	// トリックを取ってはいけない契約では「取りに行け」という理由は嘘になる。
	if g.avoidsTricks(playerIdx) {
		return "duck"
	}
	if len(g.currentTrick) == 0 {
		if card.GetDesign() == g.trumpSuit {
			return "lead_trump"
		}
		if pts == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := g.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if bauernschnapsenBeats(card, leadCard, leadSuit, g.trumpSuit) {
		if card.GetDesign() == g.trumpSuit && leadSuit != g.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI ---

// cpuChooseMarriage CPU がリード時に宣言すべきマリアージュのカードインデックスを返す。
// 切り札マリアージュを優先し、宣言時はランクの低い Q をリードする。
func (g *Bauernschnapsen) cpuChooseMarriage(playerIdx int) (int, bool) {
	player := g.players[playerIdx]
	bestIdx := -1
	bestBonus := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetValue() != 12 { // Q を起点に宣言 (低い方をリードして K を温存)
			continue
		}
		if !g.isMarriageStarter(player, c) {
			continue
		}
		bonus := BauernschnapsenMarriageBonus
		if c.GetDesign() == g.trumpSuit {
			bonus = BauernschnapsenRoyalMarriageBonus
		}
		if bonus > bestBonus {
			bestBonus = bonus
			bestIdx = i
		}
	}
	return bestIdx, bestIdx >= 0
}

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する (合法手の中から)
// avoidsTricks は席 playerIdx が**トリックを取りたくない**かを返す。
//
// ベテルは宣言者が 1 トリックも取らない契約なので、宣言者にとって
// 「トリックを取る」は敗着そのもの。クローン元のガイゲルは常に取りに行く
// 戦略しか持っておらず、そのまま使うと**CPU が自分で宣言したベテルを
// 自分で落とす**。宣言者以外は逆に取りに行けばよい (落とさせれば得点する)。
func (g *Bauernschnapsen) avoidsTricks(playerIdx int) bool {
	if g.contract != BauernschnapsenContractBettel || g.declarerIdx < 0 {
		return false
	}
	// contractMade はベテルを seatTricks[declarerIdx] で判定するので、
	// 回避するのも宣言者ひとり。パートナーは普通に打ってよい。
	return playerIdx == g.declarerIdx
}

func (g *Bauernschnapsen) cpuSelectPlayCard(playerIdx int) int {
	legal := g.legalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return 0
	}
	if len(legal) == 1 {
		return legal[0]
	}
	if g.config.CpuDifficulty == BauernschnapsenCpuDifficultyEasy {
		return legal[rand.Intn(len(legal))]
	}
	if g.avoidsTricks(playerIdx) {
		return g.cpuDuck(playerIdx, legal)
	}
	if len(g.currentTrick) == 0 {
		return g.cpuLead(playerIdx, legal)
	}
	return g.cpuFollow(playerIdx, legal)
}

// cpuDuck はトリックを取らないための札を選ぶ。
//
// 場に負ける札があればその中で一番強い札 (手札から高い札を安全に処分できる)、
// 全部勝ってしまうなら一番弱い札を出す。リード時は単純に一番弱い札。
func (g *Bauernschnapsen) cpuDuck(playerIdx int, legal []int) int {
	player := g.players[playerIdx]
	weakest := legal[0]
	weakestScore := bauernschnapsenLeadScore(player.GetCard(weakest), g.trumpSuit)
	for _, i := range legal[1:] {
		if sc := bauernschnapsenLeadScore(player.GetCard(i), g.trumpSuit); sc < weakestScore {
			weakestScore, weakest = sc, i
		}
	}
	if len(g.currentTrick) == 0 {
		return weakest
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	loseIdx, loseScore := -1, 0
	for _, i := range legal {
		c := player.GetCard(i)
		if bauernschnapsenBeats(c, g.currentBestCard(), leadSuit, g.trumpSuit) {
			continue
		}
		if sc := bauernschnapsenLeadScore(c, g.trumpSuit); loseIdx < 0 || sc > loseScore {
			loseIdx, loseScore = i, sc
		}
	}
	if loseIdx >= 0 {
		return loseIdx
	}
	return weakest
}

// cpuLead リード時の選択: 最も低い点数の非トランプを優先する。
func (g *Bauernschnapsen) cpuLead(playerIdx int, legal []int) int {
	player := g.players[playerIdx]
	bestIdx := legal[0]
	bestScore := bauernschnapsenLeadScore(player.GetCard(bestIdx), g.trumpSuit)
	for _, i := range legal[1:] {
		sc := bauernschnapsenLeadScore(player.GetCard(i), g.trumpSuit)
		if sc < bestScore {
			bestScore = sc
			bestIdx = i
		}
	}
	return bestIdx
}

// bauernschnapsenLeadScore 値が小さいほど「リードに適している」(トランプ・高得点札を温存する)
func bauernschnapsenLeadScore(c *Card, trumpSuit int) int {
	score := BauernschnapsenCardPoints(c)*10 + BauernschnapsenRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時の選択。パートナーが勝っていれば高点を載せ、そうでなければ
// 勝てる最小コストの札、無ければ最小ダンプ札を出す。
func (g *Bauernschnapsen) cpuFollow(playerIdx int, legal []int) int {
	player := g.players[playerIdx]
	leadCard := g.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	partnerIdx := (playerIdx + 2) % BauernschnapsenPlayerCnt
	partnerWinning := g.currentLeaderIdx() == partnerIdx

	if partnerWinning {
		best := legal[0]
		bestPts := -1
		for _, i := range legal {
			pts := BauernschnapsenCardPoints(player.GetCard(i))
			if pts > bestPts {
				bestPts = pts
				best = i
			}
		}
		return best
	}

	winIdx := -1
	winScore := 0
	dumpIdx := legal[0]
	dumpScoreVal := bauernschnapsenLeadScore(player.GetCard(legal[0]), g.trumpSuit)
	for _, i := range legal {
		c := player.GetCard(i)
		if bauernschnapsenBeats(c, g.currentBestCard(), leadSuit, g.trumpSuit) {
			sc := bauernschnapsenLeadScore(c, g.trumpSuit)
			if winIdx < 0 || sc < winScore {
				winIdx = i
				winScore = sc
			}
		}
		ds := bauernschnapsenLeadScore(c, g.trumpSuit)
		if ds < dumpScoreVal {
			dumpScoreVal = ds
			dumpIdx = i
		}
	}
	if winIdx >= 0 && BauernschnapsenCardPoints(leadCard) >= 10 {
		return winIdx
	}
	if !g.legalAllowsDump(playerIdx, legal) && winIdx >= 0 {
		return winIdx
	}
	return dumpIdx
}

// currentLeaderIdx 現時点のトリックの暫定勝者のプレイヤーインデックスを返す
func (g *Bauernschnapsen) currentLeaderIdx() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if bauernschnapsenBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// currentBestCard 現時点のトリックの暫定勝者カードを返す
func (g *Bauernschnapsen) currentBestCard() *Card {
	idx := g.currentLeaderIdx()
	for _, tc := range g.currentTrick {
		if tc.PlayerIdx == idx {
			return tc.Card
		}
	}
	if len(g.currentTrick) > 0 {
		return g.currentTrick[0].Card
	}
	return nil
}

// legalAllowsDump 合法手の中に「勝たないカード」が含まれるか (= 捨てる自由があるか) を返す。
func (g *Bauernschnapsen) legalAllowsDump(playerIdx int, legal []int) bool {
	player := g.players[playerIdx]
	best := g.currentBestCard()
	if best == nil {
		return true
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	for _, i := range legal {
		if !bauernschnapsenBeats(player.GetCard(i), best, leadSuit, g.trumpSuit) {
			return true
		}
	}
	return false
}

// --- Sorting / bookkeeping ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Bauernschnapsen) sortAllHands() {
	sortEachHand(g.players, g.sortHand)
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ランク でソートする
func (g *Bauernschnapsen) sortHand(p *BauernschnapsenPlayer) {
	trumpSuit := g.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return BauernschnapsenRankOrder(ci) < BauernschnapsenRankOrder(cj)
	})
}

// --- Test-only helpers ---

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Bauernschnapsen) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < BauernschnapsenTeamCnt {
		g.roundPoints[team] += pts
	}
}

// GetConfigDeckHelper returns a fresh 20-card Bauernschnapsen deck (テスト用コンストラクタ補助)。
func (g *Bauernschnapsen) GetConfigDeckHelper() *TrumpCards { return newBauernschnapsenDeck() }

// --- JSON ---

// bauernschnapsenJSON is the JSON wire format for Bauernschnapsen.
type bauernschnapsenJSON struct {
	TrumpCards       *TrumpCards                 `json:"tc"`
	Players          []*BauernschnapsenPlayer    `json:"pl"`
	Config           BauernschnapsenConfig       `json:"cfg"`
	Phase            BauernschnapsenPhase        `json:"ph"`
	RoundNumber      int                         `json:"rn"`
	TrickNumber      int                         `json:"tn"`
	CurrentPlayerIdx int                         `json:"cp"`
	CurrentTrick     []*TrickCard                `json:"ct"`
	TrumpSuit        int                         `json:"ts"`
	LeadPlayerIdx    int                         `json:"li"`
	DealerIdx        int                         `json:"di"`
	TeamScores       [BauernschnapsenTeamCnt]int `json:"sc"`
	RoundPoints      [BauernschnapsenTeamCnt]int `json:"rp"`
	RoundMarriage    [BauernschnapsenTeamCnt]int `json:"rm"`
	// トリック数は契約の成否そのもの。落とすと復元後に
	// ベテル / 同スート縛りが 0 トリックとして「達成」に化ける。
	RoundTricks      [BauernschnapsenTeamCnt]int   `json:"rt"`
	SeatTricks       [BauernschnapsenPlayerCnt]int `json:"st"`
	MarriageDeclared [CardDesignMax + 1]bool       `json:"md"`
	LastTrickWinner  int                           `json:"lw"`
	// 契約は宣言で決まるので、盤面の一部として保存する。落とすと復元後に
	// 切り札も遊び方も分からなくなる。
	Contract     BauernschnapsenContract                           `json:"co"`
	DeclarerIdx  int                                               `json:"de"`
	ContractBids [BauernschnapsenPlayerCnt]BauernschnapsenContract `json:"cb"`
	GameEndFlag  bool                                              `json:"ge"`
	WinnerTeam   int                                               `json:"wt"`
	ActionLog    []*ActionLogEntry                                 `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Bauernschnapsen) MarshalJSON() ([]byte, error) {
	return json.Marshal(bauernschnapsenJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		TrumpSuit:        g.trumpSuit,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		TeamScores:       g.teamScores,
		RoundPoints:      g.roundPoints,
		RoundMarriage:    g.roundMarriage,
		RoundTricks:      g.roundTricks,
		SeatTricks:       g.seatTricks,
		MarriageDeclared: g.marriageDeclared,
		LastTrickWinner:  g.lastTrickWinner,
		Contract:         g.contract,
		DeclarerIdx:      g.declarerIdx,
		ContractBids:     g.contractBids,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// bauernschnapsenMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const bauernschnapsenMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. 列挙値・インデックス範囲・スライス要素を
// 検証し、不正な場合はエラーを返す (パニックさせない)。
func (g *Bauernschnapsen) UnmarshalJSON(data []byte) error {
	var j bauernschnapsenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// **下限は契約フェーズ。** Play より前に契約フェーズを足したので、
	// Play を下限にすると配り直後の盤 (契約待ち) を復元できない。
	if j.Phase < BauernschnapsenPhaseContract || j.Phase > BauernschnapsenPhaseGameEnd {
		return NewDomainError(ErrInvalidPlay, "無効なフェーズです")
	}
	// **未確定は BauernschnapsenNoTrump (-1)。** 切り札は宣言で決まるので、
	// 配り直後や Bettel の局面ではスートが無い。0 だけを未確定として扱うと
	// その盤を復元できない。
	if j.TrumpSuit != BauernschnapsenNoTrump && j.TrumpSuit != 0 &&
		(j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return NewDomainError(ErrInvalidPlay, "無効な切り札スートです")
	}
	if len(j.Players) != BauernschnapsenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤー数が不正です")
	}
	// 直接インデックス参照される値の範囲検証 (不正な KV 状態でのパニック防止)。
	// currentPlayerIdx / dealerIdx は常に有効なプレイヤー。leadPlayerIdx /
	// lastTrickWinner / winnerTeam は未確定を表す -1 を許可する。
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= BauernschnapsenPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= BauernschnapsenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤーインデックスが範囲外です")
	}
	if j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= BauernschnapsenPlayerCnt ||
		j.LastTrickWinner < -1 || j.LastTrickWinner >= BauernschnapsenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "リード/勝者インデックスが範囲外です")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= BauernschnapsenTeamCnt {
		return NewDomainError(ErrInvalidPlay, "勝者チームが範囲外です")
	}
	if len(j.CurrentTrick) > BauernschnapsenPlayerCnt || len(j.ActionLog) > bauernschnapsenMaxSliceLen {
		return NewDomainError(ErrInvalidPlay, "状態スライスが不正です")
	}
	for _, p := range j.Players {
		if p == nil {
			return NewDomainError(ErrInvalidPlay, "プレイヤーが nil です")
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return NewDomainError(ErrInvalidPlay, "トリックカードが nil です")
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return NewDomainError(ErrInvalidPlay, "棋譜エントリが nil です")
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newBauernschnapsenDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	g.trumpSuit = j.TrumpSuit
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.teamScores = j.TeamScores
	g.roundPoints = j.RoundPoints
	g.roundMarriage = j.RoundMarriage
	g.roundTricks = j.RoundTricks
	g.seatTricks = j.SeatTricks
	g.marriageDeclared = j.MarriageDeclared
	g.lastTrickWinner = j.LastTrickWinner
	g.contract = j.Contract
	g.declarerIdx = j.DeclarerIdx
	g.contractBids = j.ContractBids
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}

// SetContractForTest はテスト用に契約と宣言者を設定する。
func (g *Bauernschnapsen) SetContractForTest(c BauernschnapsenContract, declarerIdx int) {
	g.contract = c
	g.declarerIdx = declarerIdx
}

// SetSeatTricksForTest はテスト用に席別の獲得トリック数を設定する。
func (g *Bauernschnapsen) SetSeatTricksForTest(idx, tricks int) {
	if idx < 0 || idx >= BauernschnapsenPlayerCnt {
		return
	}
	g.seatTricks[idx] = tricks
}

// SetRoundResultForTest はテスト用にチームのラウンド成績を設定する。
func (g *Bauernschnapsen) SetRoundResultForTest(team, points, tricks int) {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return
	}
	g.roundPoints[team] = points
	g.roundTricks[team] = tricks
}

// ContractMadeForTest はテスト用に契約の成否を返す。
func (g *Bauernschnapsen) ContractMadeForTest(declarerTeam int) bool {
	return g.contractMade(declarerTeam)
}
