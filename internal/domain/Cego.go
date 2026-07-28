//go:build !js || !wasm || extra3

// Package domain チェゴ (Cego) のドメインモデル。
//
// Cego はドイツ・バーデン地方の 54 枚タロック (Tarock) デッキを用いる 4 人用コントラクト・
// トリックテイキング。本実装は 1 人のデクレアラー (declarer) が残り 3 人と対戦する 1 対 3 の
// MVP で、Cego の代名詞である「場札 (Cego / blind) との交換」を中心に据える。パートナーシップ
// (王呼び) は存在しない。
//
// # デッキ (54 枚) — Königrufen と同一
//
//   - スート札 32 枚: design = 1..4 (4 スート)、value = 1..8。1..4 はピップ札、
//     5 = ジャック (Jack)、6 = カヴァリエ (Cavalier)、7 = クイーン (Queen)、8 = キング (King)。
//   - 切り札 (Tarock) 21 枚: design = CegoTrumpDesign (5)、value = 1..21。
//   - スキュース (Sküs / Fool) 1 枚: design = CegoSkusDesign (6)、value = 0。**最強の切り札**
//     として振る舞う。トゥルル (Trull) 名誉札は切り札 I (Pagat)・切り札 XXI・スキュースの 3 枚。
//
// # 簡略化ルール (本実装が採用する縮小版) — すべて MVP のための省略
//
//   - 配札: 各プレイヤーに 11 枚 (44 枚) を 1 枚ずつ 11 巡で配り、残り 10 枚を伏せた場札
//     (Cego / blind) として中央に置く。11 枚 × 4 人 + 10 枚 = 54 枚。1 ディール 11 トリック。
//   - 入札 (Bidding): 本来のバーデン・コントラクト梯子 (Solo / Räuber / Bettel / …) は **省略**し、
//     単一の「プレイ宣言」のみを扱う。1 巡でパスまたはプレイを宣言し、最初にプレイを宣言した者が
//     デクレアラー (梯子は 1 段のみなので後続は上回れない)。全員パスならディーラーの左隣が強制
//     デクレアラー (再配札なし — 簡略化)。倍率は常に 1。
//   - コントラクト選択 (Cego の代名詞): デクレアラーは Cego か Handspiel (Solo) を選ぶ。
//   - **Cego**: デクレアラーは自分の 11 枚から 1 枚だけ残し (CegoKeepCount=1)、残り 10 枚を
//     伏せて自分の得点山 (stash, stashOwner=0=デクレアラー側) に置き、10 枚の場札をすべて手に
//     取って新しい 11 枚の手札を作る (フレンチタロットのエクサンジュ écart を踏襲)。伏せた 10 枚は
//     デクレアラー側の得点に計上される (キング等を伏せても自分の得点になる)。
//   - **Handspiel (Solo)**: デクレアラーは配られた 11 枚をそのまま使い、場札 10 枚は対戦側の
//     得点山 (stash, stashOwner=1=対戦側) に渡る。伏せたまま公開されない。
//     ※本来 Handspiel は得点が高いが、倍率差は **省略** (常に 1)。
//   - トリックプレイ (11 トリック): リードスートに従う義務。ボイド時は切り札 (Tarock) を出す
//     義務。切り札が場に出ていれば可能な限り上位切り札を出す義務 (オーバートランプ)。スキュースは
//     最強の切り札。最強切り札が勝ち、なければリードスートの最高札が勝つ。
//   - カードポイント (簡略化した 1 枚ごとの整数配点、合計 CegoTotalPoints=106):
//     キング=5、クイーン=4、カヴァリエ=3、ジャック=2、トゥルル (Pagat I・XXI・Sküs)=各 5、
//     その他 (ピップ 1-4・素の切り札 2-20)=1。本来の「3 枚ずつ数えて 2 を引く」正式計算は
//     **省略**。全札の合計は必ず 106。
//   - 得点: デクレアラー単独の獲得点が過半 (2×declarerPoints > 106、すなわち 54 点以上) なら
//     成功。ゼロサムで精算する (1 対 3): デクレアラー ±3×base、対戦側各人 ∓base。base =
//     (CegoBaseGameValue + |declarerPoints - 53|) × mult (mult は常に 1)。ボーナス
//     (Trull / König / Ultimo) は **省略**。
//   - 累積得点: TargetDeals ディール後、累積最上位が勝者。CegoResult は人間視点。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// CegoPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const CegoPlayerCnt = 4

// CegoHandSize 各プレイヤーの配り札枚数
const CegoHandSize = 11

// CegoBlindSize 場札 (Cego / blind) の枚数
const CegoBlindSize = 10

// CegoKeepCount Cego コントラクト時にデクレアラーが手元に残す枚数
const CegoKeepCount = 1

// CegoLayDownCount Cego コントラクト時にデクレアラーが伏せる枚数
const CegoLayDownCount = CegoHandSize - CegoKeepCount

// CegoDeckSize デッキ枚数 (54 枚タロックデッキ)
const CegoDeckSize = 54

// CegoTrickCount 1 ディールのトリック数
const CegoTrickCount = 11

// CegoDefaultDeals マッチを構成するディール数 (既定)
const CegoDefaultDeals = 5

// CegoSuitCnt スート数
const CegoSuitCnt = 4

// CegoTrumpDesign 切り札 (Tarock) の仮想デザイン値。1..4 はスート、5 が切り札。
const CegoTrumpDesign = 5

// CegoSkusDesign スキュース (Sküs / Fool) の仮想デザイン値。
const CegoSkusDesign = 6

// CegoSkusValue スキュースのカード値。
const CegoSkusValue = 0

// CegoSkusRank スキュースをトリック比較で扱う際の切り札ランク (最強、21 より上)。
const CegoSkusRank = 22

// CegoMaxTrump 切り札の最大値 (21)。
const CegoMaxTrump = 21

// CegoPagatValue パガト (最小の切り札トゥルル, 切り札 I) の値。
const CegoPagatValue = 1

// CegoKingValue スート札のキング (King) の値。
const CegoKingValue = 8

// CegoSuitMaxValue スート札の最大値 (キング)。
const CegoSuitMaxValue = 8

// CegoTotalPoints デッキ総カードポイント (簡略化した配点の合計)。
const CegoTotalPoints = 106

// CegoBaseGameValue 精算の基礎ゲーム価値。
const CegoBaseGameValue = 10

// CegoBid 入札 (コントラクト) 種別
type CegoBid int

// Cego の入札定数 (値が大きいほど高い入札)
const (
	// CegoBidPass パス / 未入札
	CegoBidPass CegoBid = 0
	// CegoBidPlay プレイ宣言 — 本実装唯一の入札段
	CegoBidPlay CegoBid = 1
)

// CegoContract コントラクト (交換方式) 種別
type CegoContract int

// Cego のコントラクト定数
const (
	// CegoContractNone 未選択
	CegoContractNone CegoContract = 0
	// CegoContractCego Cego (場札交換) — 1 枚残し 10 枚を伏せ、場札を取る
	CegoContractCego CegoContract = 1
	// CegoContractHandspiel Handspiel / Solo — 配られた手札のまま、場札は対戦側へ
	CegoContractHandspiel CegoContract = 2
)

// CegoPhase ゲームフェーズ
type CegoPhase int

// Cego のフェーズ定数
const (
	// CegoPhaseBid 入札フェーズ
	CegoPhaseBid CegoPhase = 0
	// CegoPhaseContract コントラクト選択 (Cego / Handspiel) フェーズ
	CegoPhaseContract CegoPhase = 1
	// CegoPhaseExchange 場札交換 (Cego コントラクト時のみ) フェーズ
	CegoPhaseExchange CegoPhase = 2
	// CegoPhasePlay トリックプレイフェーズ
	CegoPhasePlay CegoPhase = 3
	// CegoPhaseTrickEnd トリック終了フェーズ
	CegoPhaseTrickEnd CegoPhase = 4
	// CegoPhaseRoundEnd ディール終了フェーズ
	CegoPhaseRoundEnd CegoPhase = 5
	// CegoPhaseGameEnd ゲーム終了フェーズ
	CegoPhaseGameEnd CegoPhase = 6
)

// CegoPhaseMin フェーズ下限 (検証用)
const CegoPhaseMin = int(CegoPhaseBid)

// CegoPhaseMax フェーズ上限 (検証用)
const CegoPhaseMax = int(CegoPhaseGameEnd)

// CegoOutcome ディール結果 (デクレアラー視点)
type CegoOutcome int

// Cego のディール結果定数
const (
	// CegoOutcomeNone 未確定
	CegoOutcomeNone CegoOutcome = 0
	// CegoOutcomeWin デクレアラーがコントラクトを達成
	CegoOutcomeWin CegoOutcome = 1
	// CegoOutcomeLoss デクレアラーがコントラクトを失敗
	CegoOutcomeLoss CegoOutcome = 2
)

// CegoResult 人間視点のマッチ結果。casino タグの GameResult は solo ワーカーから到達不能なため、
// ゲームローカルの結果型を定義する。
type CegoResult int

// Cego のマッチ結果定数
const (
	// CegoResultLose 敗北
	CegoResultLose CegoResult = -1
	// CegoResultNone 未確定 / 引き分け
	CegoResultNone CegoResult = 0
	// CegoResultWin 勝利
	CegoResultWin CegoResult = 1
)

// CegoHint ヒント情報
type CegoHint struct {
	Bid         *int   // 推奨入札 (入札フェーズ)。nil の場合はパス推奨
	Contract    *int   // 推奨コントラクト (コントラクトフェーズ)
	CardIndices []int  // 推奨カードインデックス (交換/プレイ)
	Reason      string // ヒント理由キー
}

// CegoBreakdown 得点計算の内訳 (純粋関数 cegoScoreDeal の出力)。
type CegoBreakdown struct {
	// DeclarerPoints デクレアラーが獲得したカードポイント合計。
	DeclarerPoints int
	// Threshold 「過半」の閾値 (この値を超える = 成功)。
	Threshold int
	// Won コントラクト成否。
	Won bool
	// Diff 閾値との差 (絶対値、整数点)。
	Diff int
	// Base (CegoBaseGameValue + Diff) × Mult。
	Base int
	// Mult 入札倍率 (常に 1)。
	Mult int
	// DeclarerScore デクレアラーの得点変動。
	DeclarerScore int
	// OpponentScore 対戦側 1 人の得点変動。
	OpponentScore int
}

// Cego チェゴのゲームクラス
type Cego struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*CegoPlayer
	config           CegoConfig
	phase            CegoPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	// --- bidding state ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    CegoBid
	highestBidder int
	passed        [CegoPlayerCnt]bool
	// --- contract state ---
	declarerIdx  int
	contract     CegoBid
	contractType CegoContract
	// --- blind / stash state (blind はサーバー側のみ、Web 出力に漏らさない) ---
	blind      []*Card // 場札 (Cego / blind, 10 枚)。伏せられ公開されない。
	stash      []*Card // 得点計上用に脇へ置いた 10 枚
	stashOwner int     // 0 = デクレアラー側 (Cego), 1 = 対戦側 (Handspiel)
	// --- scoring ---
	playerScores    [CegoPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         CegoOutcome
	result          CegoResult
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int
	actionLog       []*ActionLogEntry
}

// NewCego コンストラクタ
func NewCego(players []*CegoPlayer, config CegoConfig) *Cego {
	return &Cego{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        CegoBidPass,
		contractType:    CegoContractNone,
		stashOwner:      0,
	}
}

// NewDefaultCego 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultCego() *Cego {
	players := make([]*CegoPlayer, CegoPlayerCnt)
	players[0] = NewCegoPlayer(true)
	for i := 1; i < CegoPlayerCnt; i++ {
		players[i] = NewCegoPlayer(false)
	}
	return NewCego(players, DefaultCegoConfig())
}

// buildCegoDeck 54 枚タロックデッキを直接構築する。スート札 (design 1..4, value 1..8) 32 枚 +
// 切り札 (design 5, value 1..21) 21 枚 + スキュース (design 6, value 0)。
func buildCegoDeck() []*Card {
	deck := make([]*Card, 0, CegoDeckSize)
	for suit := 1; suit <= CegoSuitCnt; suit++ {
		for val := 1; val <= CegoSuitMaxValue; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= CegoMaxTrump; val++ {
		deck = append(deck, NewCard(CegoTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(CegoSkusDesign, CegoSkusValue, false))
	return deck
}

// Reset ゲーム初期化
func (g *Cego) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [CegoPlayerCnt]int{}
	g.result = CegoResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Cego) NextRound() {
	if g.phase != CegoPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % CegoPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Cego) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = CegoBidPass
	g.contractType = CegoContractNone
	g.blind = nil
	g.stash = nil
	g.stashOwner = 0
	g.outcome = CegoOutcomeNone
	g.scored = false
	g.passed = [CegoPlayerCnt]bool{}
	g.highestBid = CegoBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % CegoPlayerCnt
	g.phase = CegoPhaseBid
}

// deal 各プレイヤーへ 11 枚を 1 枚ずつ配り、場札 (Cego / blind) 10 枚を伏せて脇に置く。
func (g *Cego) deal() {
	g.deck = buildCegoDeck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	g.blind = make([]*Card, 0, CegoBlindSize)
	for r := 0; r < CegoHandSize; r++ {
		for j := 0; j < CegoPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % CegoPlayerCnt
			if c := g.drawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	for k := 0; k < CegoBlindSize; k++ {
		if c := g.drawCard(); c != nil {
			g.blind = append(g.blind, c)
		}
	}
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Cego) drawCard() *Card {
	if g.deckDrawCnt >= len(g.deck) {
		return nil
	}
	card := g.deck[g.deckDrawCnt]
	card.SetDraw(true)
	g.deckDrawCnt++
	return card
}

// --- Bidding ---

// PlayerBid 人間プレイヤーが入札する。
func (g *Cego) PlayerBid(bid CegoBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CegoPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	if !cegoValidBid(bid) {
		return NewDomainError(ErrInvalidPlay, "無効な入札です (play)")
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay, "現在の入札より高い入札が必要です")
	}
	g.applyBid(g.bidPlayerIdx, bid)
	return nil
}

// PlayerPass 人間プレイヤーがパスする。
func (g *Cego) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CegoPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU プレイヤーが 1 回入札する (入札 or パス)。
func (g *Cego) CpuBid() {
	if g.gameEndFlag || g.phase != CegoPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= CegoPlayerCnt {
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

// applyBid 入札を適用する。
func (g *Cego) applyBid(idx int, bid CegoBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), cegoBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを適用する。
func (g *Cego) applyPass(idx int) {
	g.passed[idx] = true
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid 入札を次のプレイヤーへ進め、1 巡終了でコントラクトを確定する。
func (g *Cego) advanceBid() {
	g.bidActedCnt++
	if g.bidActedCnt >= CegoPlayerCnt {
		g.finalizeBid()
		return
	}
	g.bidPlayerIdx = (g.bidPlayerIdx + 1) % CegoPlayerCnt
}

// finalizeBid 入札を確定し、デクレアラーを決定してコントラクト選択へ進む。全員パスならディーラーの
// 左隣を強制デクレアラーとする (再配札なし)。
func (g *Cego) finalizeBid() {
	if g.highestBidder < 0 {
		g.declarerIdx = (g.dealerIdx + 1) % CegoPlayerCnt
		g.contract = CegoBidPlay
		g.appendLog(g.declarerIdx, "forced",
			fmt.Sprintf("all passed — %s is forced to declare", g.playerName(g.declarerIdx)), nil)
	} else {
		g.declarerIdx = g.highestBidder
		g.contract = g.highestBid
		g.appendLog(g.declarerIdx, "win_bid",
			fmt.Sprintf("%s takes the declaration", g.playerName(g.declarerIdx)), nil)
	}
	g.currentPlayerIdx = g.declarerIdx
	g.phase = CegoPhaseContract
}

// --- Contract choice ---

// PlayerChooseContract 人間デクレアラーがコントラクト (Cego / Handspiel) を選ぶ。
func (g *Cego) PlayerChooseContract(ct CegoContract) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CegoPhaseContract {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if ct != CegoContractCego && ct != CegoContractHandspiel {
		return NewDomainError(ErrInvalidPlay, "コントラクトは Cego か Handspiel を選んでください")
	}
	g.applyContract(ct)
	return nil
}

// CpuChooseContract CPU デクレアラーがコントラクトを自動選択する。
func (g *Cego) CpuChooseContract() {
	if g.gameEndFlag || g.phase != CegoPhaseContract {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	g.applyContract(g.cpuSelectContract())
}

// applyContract コントラクトを確定する。Cego なら交換フェーズへ、Handspiel なら場札を対戦側へ
// 渡して即プレイへ進む。
func (g *Cego) applyContract(ct CegoContract) {
	g.contractType = ct
	g.appendLog(g.declarerIdx, "contract",
		fmt.Sprintf("%s chooses %s", g.playerName(g.declarerIdx), cegoContractName(ct)), nil)
	if ct == CegoContractHandspiel {
		// 場札は対戦側の得点山に渡る (伏せたまま公開しない)。
		g.stash = g.blind
		g.stashOwner = 1
		g.blind = make([]*Card, 0)
		g.startPlay()
		return
	}
	g.currentPlayerIdx = g.declarerIdx
	g.phase = CegoPhaseExchange
}

// cpuSelectContract CPU デクレアラーのコントラクト選択。手札が強ければ Handspiel (手札温存)、
// 弱ければ Cego (場札と交換)。
func (g *Cego) cpuSelectContract() CegoContract {
	strength := g.evalHand(g.declarerIdx)
	threshold := 24
	switch g.config.CpuDifficulty {
	case CegoCpuDifficultyEasy:
		threshold = 20
	case CegoCpuDifficultyHard:
		threshold = 28
	}
	if strength >= threshold {
		return CegoContractHandspiel
	}
	return CegoContractCego
}

// --- Exchange (Cego) ---

// PlayerDiscard 人間デクレアラーが Cego 交換で残す 1 枚を選ぶ。keepIndices は残す札のインデックス
// (ちょうど CegoKeepCount 枚)。残り 10 枚は伏せてデクレアラー側の得点山へ移し、場札 10 枚を取る。
func (g *Cego) PlayerDiscard(keepIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CegoPhaseExchange {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doExchange(keepIndices)
}

// CpuDiscard CPU デクレアラーが Cego 交換を自動で行う。
func (g *Cego) CpuDiscard() {
	if g.gameEndFlag || g.phase != CegoPhaseExchange {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	_ = g.doExchange(g.cpuSelectKeep(g.declarerIdx))
}

// doExchange Cego 交換の共通処理。残す 1 枚を除いた札をデクレアラー側の得点山 (stash) に伏せ、
// 場札 10 枚をデクレアラーの手札に加える。
func (g *Cego) doExchange(keepIndices []int) error {
	player := g.players[g.declarerIdx]
	if len(keepIndices) != CegoKeepCount {
		return NewDomainError(ErrInvalidCard, "残す札をちょうど 1 枚選んでください")
	}
	seen := make(map[int]bool, CegoKeepCount)
	for _, idx := range keepIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードを 2 回選べません")
		}
		seen[idx] = true
	}
	// 残す札以外を伏せる。
	layDown := make([]int, 0, CegoLayDownCount)
	for i := 0; i < player.GetCardsSize(); i++ {
		if !seen[i] {
			layDown = append(layDown, i)
		}
	}
	discarded := player.RemoveCards(layDown)
	g.stash = discarded
	g.stashOwner = 0
	// 場札を手札に加える。
	for _, c := range g.blind {
		player.AddCard(c)
	}
	g.blind = make([]*Card, 0)
	g.appendLog(g.declarerIdx, "exchange",
		fmt.Sprintf("%s lays down %d cards and takes the Cego", g.playerName(g.declarerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlay()
	return nil
}

// cpuSelectKeep CPU デクレアラーが残す 1 枚 (最も価値の高い札) のインデックスを選ぶ。
func (g *Cego) cpuSelectKeep(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	if n == 0 {
		return []int{0}
	}
	best := 0
	bestVal := cegoKeepValue(p.GetCard(0))
	for i := 1; i < n; i++ {
		if v := cegoKeepValue(p.GetCard(i)); v > bestVal {
			bestVal = v
			best = i
		}
	}
	return []int{best}
}

// cegoKeepValue 温存価値を返す (トゥルル > キング > 高位切り札 > 一般札)。
func cegoKeepValue(c *Card) int {
	if c == nil {
		return 0
	}
	if cegoIsTrull(c) {
		return 100000
	}
	if cegoIsKing(c) {
		return 90000
	}
	if cegoIsTrumpLike(c) {
		return 10000 + c.GetValue()
	}
	return c.GetValue()*10 + cegoCardPoints(c)
}

// --- Play ---

// startPlay プレイフェーズを開始する。エルダー (ディーラーの左隣) がリードする。
func (g *Cego) startPlay() {
	g.sortAllHands()
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % CegoPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = CegoPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Cego) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CegoPhasePlay {
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

// CpuPlay CPU プレイヤーが 1 ターン実行する。
func (g *Cego) CpuPlay() {
	if g.gameEndFlag || g.phase != CegoPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Cego) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cegoCardStr(card)), []*Card{card})
	if len(g.currentTrick) == CegoPlayerCnt {
		g.phase = CegoPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % CegoPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。全札を勝者のトリック山へ。最終トリックなら
// RoundEnd に入り得点計算を発火する。
func (g *Cego) ResolveTrick() {
	if g.phase != CegoPhaseTrickEnd || len(g.currentTrick) != CegoPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	allCards := make([]*Card, 0, CegoPlayerCnt)
	for _, tc := range g.currentTrick {
		allCards = append(allCards, tc.Card)
	}
	g.players[winnerIdx].AddTrick(allCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", g.playerName(winnerIdx), g.trickNumber), allCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= CegoTrickCount {
		g.lastTrickWinner = winnerIdx
		g.lastTrickCards = allCards
		g.phase = CegoPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = CegoPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Cego) NextTrick() {
	if g.phase != CegoPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = CegoPhasePlay
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ)。
func (g *Cego) ScoreRound() {
	if g.phase != CegoPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd RoundEnd 突入時に一度だけ得点計算と精算を行う (scored フラグでガード)。
func (g *Cego) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	bd := g.computeBreakdown()
	if bd.Won {
		g.outcome = CegoOutcomeWin
	} else {
		g.outcome = CegoOutcomeLoss
	}
	for i := 0; i < CegoPlayerCnt; i++ {
		if i == g.declarerIdx {
			g.playerScores[i] += bd.DeclarerScore
		} else {
			g.playerScores[i] += bd.OpponentScore
		}
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: declarer(%s) %s declPts=%d/%d won=%t base=%d",
			g.roundNumber, g.playerName(g.declarerIdx), cegoContractName(g.contractType),
			bd.DeclarerPoints, CegoTotalPoints, bd.Won, bd.Base), nil)
	g.checkGameEnd()
}

// computeBreakdown 現在のディールの得点内訳を計算する。
func (g *Cego) computeBreakdown() CegoBreakdown {
	return cegoScoreDeal(g.declarerCaptured(), cegoBidMult(g.contract))
}

// declarerCaptured デクレアラーが獲得したカードポイント (トリック山 + Cego 交換で伏せた stash) を
// 返す。
func (g *Cego) declarerCaptured() int {
	pts := g.playerTrickPoints(g.declarerIdx)
	if g.stashOwner == 0 {
		for _, c := range g.stash {
			pts += cegoCardPoints(c)
		}
	}
	return pts
}

// playerTrickPoints プレイヤー idx が獲得したトリック山のカードポイント合計を返す。
func (g *Cego) playerTrickPoints(idx int) int {
	if idx < 0 || idx >= len(g.players) {
		return 0
	}
	pts := 0
	for _, trick := range g.players[idx].GetTricksTaken() {
		for _, c := range trick {
			pts += cegoCardPoints(c)
		}
	}
	return pts
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積得点最上位を勝者とする。
func (g *Cego) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < CegoPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.phase = CegoPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	if tie {
		g.winnerPlayer = -1
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
	} else {
		g.winnerPlayer = leader
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// humanResult 人間 (seat 0) 視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None。
func (g *Cego) humanResult(leader int, tie bool) CegoResult {
	human := g.findHumanIdx()
	if human < 0 {
		return CegoResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return CegoResultNone
		}
		return CegoResultWin
	}
	return CegoResultLose
}

// --- Scoring helper (pure) ---

// cegoBidMult 入札倍率を返す (Play=1)。
func cegoBidMult(_ CegoBid) int { return 1 }

// cegoScoreDeal ディール得点を計算する純粋関数。declarerPoints はデクレアラーの獲得カードポイント、
// mult は倍率。デクレアラーが過半 (2×declarerPoints > CegoTotalPoints) を取れば成功。精算はゼロサム
// (1 対 3): デクレアラー ±3×base、対戦側各人 ∓base。
func cegoScoreDeal(declarerPoints int, mult int) CegoBreakdown {
	threshold := CegoTotalPoints / 2 // 53 — これを超えれば成功
	won := 2*declarerPoints > CegoTotalPoints
	diff := declarerPoints - threshold
	if diff < 0 {
		diff = -diff
	}
	base := (CegoBaseGameValue + diff) * mult
	winSign := 1
	if !won {
		winSign = -1
	}
	return CegoBreakdown{
		DeclarerPoints: declarerPoints,
		Threshold:      threshold,
		Won:            won,
		Diff:           diff,
		Base:           base,
		Mult:           mult,
		DeclarerScore:  winSign * 3 * base,
		OpponentScore:  -winSign * base,
	}
}

// --- Card classification / points ---

// cegoIsTrump 切り札 (Tarock, design 5) か。
func cegoIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == CegoTrumpDesign
}

// cegoIsSkus スキュース (design 6) か。
func cegoIsSkus(c *Card) bool {
	return c != nil && c.GetDesign() == CegoSkusDesign
}

// cegoIsTrumpLike 切り札またはスキュース (トリックで切り札として振る舞う札) か。
func cegoIsTrumpLike(c *Card) bool {
	return cegoIsTrump(c) || cegoIsSkus(c)
}

// cegoIsKing スート札のキングか。
func cegoIsKing(c *Card) bool {
	return c != nil && !cegoIsTrumpLike(c) && c.GetValue() == CegoKingValue
}

// cegoIsTrull トゥルル (Pagat I / 切り札 XXI / Sküs) か。
func cegoIsTrull(c *Card) bool {
	if c == nil {
		return false
	}
	if cegoIsSkus(c) {
		return true
	}
	return cegoIsTrump(c) && (c.GetValue() == CegoPagatValue || c.GetValue() == CegoMaxTrump)
}

// cegoTrumpValue 切り札札のトリック比較用の値を返す (スキュース=22 で最強、非切り札=0)。
func cegoTrumpValue(c *Card) int {
	if cegoIsSkus(c) {
		return CegoSkusRank
	}
	if cegoIsTrump(c) {
		return c.GetValue()
	}
	return 0
}

// cegoCardPoints カードの (簡略化した) カードポイントを返す。
// キング=5、クイーン=4、カヴァリエ=3、ジャック=2、トゥルル=5、その他=1。
func cegoCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if cegoIsTrull(c) {
		return 5
	}
	if cegoIsTrumpLike(c) {
		return 1
	}
	switch c.GetValue() {
	case CegoKingValue: // King
		return 5
	case 7: // Queen
		return 4
	case 6: // Cavalier
		return 3
	case 5: // Jack
		return 2
	default:
		return 1
	}
}

// --- Trick logic ---

// ledSuit 現在のトリックのリードスート (実効スート) を返す。最初の札が切り札系なら CegoTrumpDesign、
// それ以外はその design。トリックが空なら -1。
func (g *Cego) ledSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0].Card
	if cegoIsTrumpLike(first) {
		return CegoTrumpDesign
	}
	if first == nil {
		return -1
	}
	return first.GetDesign()
}

// highestTrumpInTrick 現在のトリック中の最強切り札の値を返す (0=切り札なし、スキュース=22)。
func (g *Cego) highestTrumpInTrick() int {
	best := 0
	for _, tc := range g.currentTrick {
		if v := cegoTrumpValue(tc.Card); v > best {
			best = v
		}
	}
	return best
}

// validatePlay マストフォロー + 切り札義務 + オーバートランプ義務を検証する。
func (g *Cego) validatePlay(playerIdx int, card *Card) error {
	valid := g.getValidPlayIndices(playerIdx)
	player := g.players[playerIdx]
	for _, idx := range valid {
		if player.GetCard(idx) == card {
			return nil
		}
	}
	return NewDomainError(ErrInvalidPlay, "フォロー義務・切り札義務・オーバートランプ義務に反しています")
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Cego) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	n := player.GetCardsSize()
	all := make([]int, 0, n)
	for i := 0; i < n; i++ {
		all = append(all, i)
	}
	if len(g.currentTrick) == 0 {
		return all
	}
	led := g.ledSuit()
	highestTrump := g.highestTrumpInTrick()
	var base []int
	if led == CegoTrumpDesign {
		base = g.trumpFollowIndices(player, highestTrump)
	} else {
		base = g.suitFollowIndices(player, led, highestTrump)
	}
	if len(base) == 0 {
		return all
	}
	return base
}

// trumpFollowIndices 切り札がリードされた場合の合法札を返す。
func (g *Cego) trumpFollowIndices(player *CegoPlayer, highestTrump int) []int {
	trumps := g.trumpLikeIndices(player)
	if len(trumps) == 0 {
		return g.allIndices(player)
	}
	higher := cegoFilter(trumps, func(idx int) bool {
		return cegoTrumpValue(player.GetCard(idx)) > highestTrump
	})
	if len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitFollowIndices スートがリードされた場合の合法札を返す。
func (g *Cego) suitFollowIndices(player *CegoPlayer, led, highestTrump int) []int {
	ledCards := g.suitOf(player, led)
	if len(ledCards) > 0 {
		return ledCards
	}
	trumps := g.trumpLikeIndices(player)
	if len(trumps) == 0 {
		return g.allIndices(player)
	}
	higher := cegoFilter(trumps, func(idx int) bool {
		return cegoTrumpValue(player.GetCard(idx)) > highestTrump
	})
	if highestTrump > 0 && len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitOf 指定スート design (非切り札系) の手札インデックスを返す。
func (g *Cego) suitOf(player *CegoPlayer, design int) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		if cegoIsTrumpLike(c) {
			continue
		}
		if c.GetDesign() == design {
			out = append(out, i)
		}
	}
	return out
}

// trumpLikeIndices 切り札系 (切り札 + スキュース) の手札インデックスを返す。
func (g *Cego) trumpLikeIndices(player *CegoPlayer) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if cegoIsTrumpLike(player.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// allIndices 全手札インデックスを返す。
func (g *Cego) allIndices(player *CegoPlayer) []int {
	out := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// trickWinner トリックの勝者を決定する。切り札系があれば最強切り札、無ければリードスートの最強札。
func (g *Cego) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.ledSuit()
	winIdx := g.currentTrick[0].PlayerIdx
	winRank := -1
	for _, tc := range g.currentTrick {
		r := cegoWinRank(tc.Card, led)
		if r > winRank {
			winRank = r
			winIdx = tc.PlayerIdx
		}
	}
	return winIdx
}

// cegoWinRank トリック勝敗比較用のランクを返す (高いほど強い)。切り札系 = 1000+切り札値
// (スキュース=1022 で最強)、リードスート = 値、それ以外 = -1。
func cegoWinRank(c *Card, led int) int {
	if c == nil {
		return -1
	}
	if cegoIsTrumpLike(c) {
		return 1000 + cegoTrumpValue(c)
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return -1
}

// --- CPU AI ---

// cpuSelectBid CPU の入札選択 (ok=false でパス)。手札評価が閾値以上で未確定なら Play。
func (g *Cego) cpuSelectBid(playerIdx int) (CegoBid, bool) {
	if g.highestBid >= CegoBidPlay {
		return CegoBidPass, false
	}
	strength := g.evalHand(playerIdx)
	base := 20
	switch g.config.CpuDifficulty {
	case CegoCpuDifficultyEasy:
		base = 26
	case CegoCpuDifficultyHard:
		base = 16
	}
	if strength >= base {
		return CegoBidPlay, true
	}
	return CegoBidPass, false
}

// evalHand 手札の強さを大まかに見積もる (トゥルル・切り札枚数・キング・高位切り札から算出)。
func (g *Cego) evalHand(playerIdx int) int {
	p := g.players[playerIdx]
	score := 0
	trumps := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		switch {
		case cegoIsTrull(c):
			score += 6
		case cegoIsTrumpLike(c):
			trumps++
			if c.GetValue() >= 15 {
				score += 2
			} else {
				score++
			}
		case cegoIsKing(c):
			score += 3
		case c.GetValue() == 7: // Queen
			score += 2
		}
	}
	score += trumps
	return score
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Cego) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == CegoCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 味方 (デクレアラーかどうか) を意識した戦略プレイ。
func (g *Cego) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		if g.isDeclarerSide(playerIdx) {
			return g.maxByRank(playerIdx, valid)
		}
		return g.minByPoints(playerIdx, valid)
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexInTrick(winnerIdx)].Card
	winnerSide := g.isDeclarerSide(winnerIdx)
	mySide := g.isDeclarerSide(playerIdx)
	winners := cegoFilter(valid, func(idx int) bool {
		return cegoWinRank(p.GetCard(idx), led) > cegoWinRank(winCard, led)
	})
	if winnerSide == mySide {
		return g.maxByPoints(playerIdx, valid) // 味方が勝っている → 点札を渡す
	}
	if len(winners) > 0 {
		return g.minByRank(playerIdx, winners) // 勝てる最弱札で取りに行く
	}
	return g.minByPoints(playerIdx, valid)
}

// isDeclarerSide playerIdx がデクレアラー側 (Cego ではデクレアラー本人のみ) か。
func (g *Cego) isDeclarerSide(playerIdx int) bool {
	return playerIdx == g.declarerIdx
}

// indexInTrick currentTrick 内で playerIdx の位置を返す (-1=なし)。
func (g *Cego) indexInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// maxByRank 勝敗ランク最大の札を返す。
func (g *Cego) maxByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := cegoPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := cegoPlayRank(p.GetCard(idx), led); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByRank 勝敗ランク最小の札を返す。
func (g *Cego) minByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := cegoPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := cegoPlayRank(p.GetCard(idx), led); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByPoints カードポイント最大の札を返す。
func (g *Cego) maxByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := cegoCardPoints(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := cegoCardPoints(p.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByPoints カードポイント最小の札を返す。
func (g *Cego) minByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := cegoCardPoints(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := cegoCardPoints(p.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// cegoPlayRank プレイ順比較用のランク (切り札系 = 1000+切り札値)。
func cegoPlayRank(c *Card, led int) int {
	if c == nil {
		return -1
	}
	if cegoIsTrumpLike(c) {
		return 1000 + cegoTrumpValue(c)
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return c.GetValue()
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Cego) GetHint() *CegoHint {
	human := g.findHumanIdx()
	if human < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case CegoPhaseBid:
		if g.bidPlayerIdx != human {
			return nil
		}
		if bid, ok := g.cpuSelectBid(human); ok {
			b := int(bid)
			return &CegoHint{Bid: &b, Reason: "bid_take"}
		}
		return &CegoHint{Reason: "bid_pass"}
	case CegoPhaseContract:
		if g.declarerIdx != human {
			return nil
		}
		ct := int(g.cpuSelectContract())
		reason := "contract_cego"
		if CegoContract(ct) == CegoContractHandspiel {
			reason = "contract_handspiel"
		}
		return &CegoHint{Contract: &ct, Reason: reason}
	case CegoPhaseExchange:
		if g.declarerIdx != human {
			return nil
		}
		return &CegoHint{CardIndices: g.cpuSelectKeep(human), Reason: "keep_best"}
	case CegoPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &CegoHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Cego) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if g.isDeclarerSide(playerIdx) {
			return "lead_high"
		}
		return "lead_low"
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexInTrick(winnerIdx)].Card
	if cegoWinRank(card, led) > cegoWinRank(winCard, led) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Cego) sortAllHands() {
	for _, p := range g.players {
		cegoSortHand(p)
	}
}

// cegoSortHand 手札をスート→値でソートする (切り札・スキュースは末尾)。
func cegoSortHand(p *CegoPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i] == nil || cards[j] == nil {
			return cards[i] != nil
		}
		di, dj := cards[i].GetDesign(), cards[j].GetDesign()
		if di != dj {
			return di < dj
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Cego) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Cego) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// isHumanBidTurn 現在の入札手番が人間か。
func (g *Cego) isHumanBidTurn() bool {
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.bidPlayerIdx].GetIsHuman()
}

// appendLog 棋譜にエントリを追加する。
func (g *Cego) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// cegoBidName 入札の表示名を返す。
func cegoBidName(bid CegoBid) string {
	if bid == CegoBidPlay {
		return "play"
	}
	return "pass"
}

// cegoContractName コントラクトの表示名を返す。
func cegoContractName(ct CegoContract) string {
	switch ct {
	case CegoContractCego:
		return "cego"
	case CegoContractHandspiel:
		return "handspiel"
	default:
		return "none"
	}
}

// cegoCardStr カードのログ表示文字列 (切り札・スキュース対応)。
func cegoCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	if cegoIsSkus(c) {
		return "Sküs"
	}
	if cegoIsTrump(c) {
		return fmt.Sprintf("T%d", c.GetValue())
	}
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	s, ok := suits[c.GetDesign()]
	if !ok {
		s = "?"
	}
	return fmt.Sprintf("%s%d", s, c.GetValue())
}

// cegoValidBid bid が有効な入札 (Play) か。
func cegoValidBid(bid CegoBid) bool {
	return bid == CegoBidPlay
}

// cegoValidBidVal bid が定義済みの入札値 (Pass 含む) か。
func cegoValidBidVal(bid CegoBid) bool {
	return bid >= CegoBidPass && bid <= CegoBidPlay
}

// cegoFilter 述語を満たすインデックスを抽出する。
func cegoFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Cego) GetPhase() CegoPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Cego) SetPhase(phase CegoPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Cego) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Cego) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Cego) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Cego) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Cego) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Cego) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Cego) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Cego) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Cego) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Cego) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Cego) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Cego) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetBidPlayerIdx 入札手番インデックス取得
func (g *Cego) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx 入札手番インデックス設定 (テスト用)
func (g *Cego) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetHighestBid 現在の最高入札取得
func (g *Cego) GetHighestBid() CegoBid { return g.highestBid }

// SetHighestBid 最高入札設定 (テスト用)
func (g *Cego) SetHighestBid(b CegoBid) { g.highestBid = b }

// GetHighestBidder 最高入札者取得 (-1=なし)
func (g *Cego) GetHighestBidder() int { return g.highestBidder }

// GetDeclarerIdx デクレアラーインデックス取得 (-1=未確定)
func (g *Cego) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx デクレアラーインデックス設定 (テスト用)
func (g *Cego) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract コントラクト (確定入札) 取得
func (g *Cego) GetContract() CegoBid { return g.contract }

// SetContract コントラクト設定 (テスト用)
func (g *Cego) SetContract(b CegoBid) { g.contract = b }

// GetContractType コントラクト種別 (Cego / Handspiel) 取得
func (g *Cego) GetContractType() CegoContract { return g.contractType }

// SetContractType コントラクト種別設定 (テスト用)
func (g *Cego) SetContractType(ct CegoContract) { g.contractType = ct }

// GetBlindCount 場札 (Cego / blind) の枚数取得。中身は公開しない。
func (g *Cego) GetBlindCount() int { return len(g.blind) }

// GetBlind 場札取得 (テスト/内部用)。Web プレゼンターは中身を出力してはならない。
func (g *Cego) GetBlind() []*Card { return g.blind }

// SetBlind 場札設定 (テスト用)
func (g *Cego) SetBlind(blind []*Card) { g.blind = blind }

// GetStashOwner stash の所有側取得 (0=デクレアラー側, 1=対戦側)
func (g *Cego) GetStashOwner() int { return g.stashOwner }

// GetStash stash 取得 (テスト用)
func (g *Cego) GetStash() []*Card { return g.stash }

// SetStash stash 設定 (テスト用)
func (g *Cego) SetStash(stash []*Card) { g.stash = stash }

// SetStashOwner stash 所有側設定 (テスト用)
func (g *Cego) SetStashOwner(o int) { g.stashOwner = o }

// GetPlayerScores プレイヤー別累積得点取得
func (g *Cego) GetPlayerScores() [CegoPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積得点設定 (テスト用)
func (g *Cego) SetPlayerScores(s [CegoPlayerCnt]int) { g.playerScores = s }

// GetCardPoints プレイヤー i が獲得したカードポイント合計を返す (表示用)。
func (g *Cego) GetCardPoints(i int) int { return g.playerTrickPoints(i) }

// GetOutcome 直近ディールの結果取得
func (g *Cego) GetOutcome() CegoOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Cego) GetResult() CegoResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Cego) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Cego) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Cego) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Cego) GetPlayer(i int) *CegoPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番 (Play) が人間か。
func (g *Cego) IsHumanTurn() bool {
	if g.phase != CegoPhasePlay {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在の入札手番が人間か。
func (g *Cego) IsHumanBidTurn() bool {
	if g.phase != CegoPhaseBid {
		return false
	}
	return g.isHumanBidTurn()
}

// IsHumanContractTurn 現在のコントラクト選択手番が人間 (=人間デクレアラー) か。
func (g *Cego) IsHumanContractTurn() bool {
	if g.phase != CegoPhaseContract || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// IsHumanExchangeTurn 現在の場札交換手番が人間 (=人間デクレアラー) か。
func (g *Cego) IsHumanExchangeTurn() bool {
	if g.phase != CegoPhaseExchange || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Cego) GetConfig() CegoConfig { return g.config }

// SetConfig 設定変更
func (g *Cego) SetConfig(cfg CegoConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Cego) GetActionLog() []*ActionLogEntry {
	if g.actionLog == nil {
		return []*ActionLogEntry{}
	}
	return g.actionLog
}

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Cego) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != CegoPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *Cego) ComputeBreakdownPublic() CegoBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Cego) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Cego) LedSuitPublic() int { return g.ledSuit() }

// CegoScoreDeal はディール得点計算の純粋関数の公開ラッパー (テスト用)。
func CegoScoreDeal(declarerPoints int, mult int) CegoBreakdown {
	return cegoScoreDeal(declarerPoints, mult)
}

// CegoBidMultPublic は入札倍率を返す (テスト用)。
func CegoBidMultPublic(bid CegoBid) int { return cegoBidMult(bid) }

// CegoCardPointsPublic はカードのカードポイントを返す (テスト用)。
func CegoCardPointsPublic(c *Card) int { return cegoCardPoints(c) }

// CegoIsTrullPublic はカードがトゥルルか返す (テスト用)。
func CegoIsTrullPublic(c *Card) bool { return cegoIsTrull(c) }

// CegoIsTrumpPublic はカードが切り札か返す (テスト用)。
func CegoIsTrumpPublic(c *Card) bool { return cegoIsTrump(c) }

// CegoIsSkusPublic はカードがスキュースか返す (テスト用)。
func CegoIsSkusPublic(c *Card) bool { return cegoIsSkus(c) }

// CegoIsKingPublic はカードがスートのキングか返す (テスト用)。
func CegoIsKingPublic(c *Card) bool { return cegoIsKing(c) }

// BuildCegoDeckPublic は 54 枚デッキを構築する (テスト用)。
func BuildCegoDeckPublic() []*Card { return buildCegoDeck() }

// --- JSON ---

// cegoJSON is the JSON wire format for Cego.
type cegoJSON struct {
	Deck             []*Card             `json:"dk"`
	DeckDrawCnt      int                 `json:"dw"`
	Players          []*CegoPlayer       `json:"ps"`
	Config           CegoConfig          `json:"cf"`
	Phase            CegoPhase           `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LeadPlayerIdx    int                 `json:"li"`
	DealerIdx        int                 `json:"di"`
	BidPlayerIdx     int                 `json:"bi"`
	BidActedCnt      int                 `json:"ba"`
	HighestBid       CegoBid             `json:"hb"`
	HighestBidder    int                 `json:"hr"`
	Passed           [CegoPlayerCnt]bool `json:"pd"`
	DeclarerIdx      int                 `json:"dc"`
	Contract         CegoBid             `json:"co"`
	ContractType     CegoContract        `json:"cn"`
	Blind            []*Card             `json:"bl"`
	Stash            []*Card             `json:"st"`
	StashOwner       int                 `json:"so"`
	PlayerScores     [CegoPlayerCnt]int  `json:"sc"`
	LastTrickWinner  int                 `json:"lt"`
	LastTrickCards   []*Card             `json:"lc"`
	Outcome          CegoOutcome         `json:"oc"`
	Result           CegoResult          `json:"rs"`
	Scored           bool                `json:"sd"`
	GameEndFlag      bool                `json:"ge"`
	WinnerPlayer     int                 `json:"wp"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Cego) MarshalJSON() ([]byte, error) {
	return json.Marshal(cegoJSON{
		Deck:             g.deck,
		DeckDrawCnt:      g.deckDrawCnt,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		BidPlayerIdx:     g.bidPlayerIdx,
		BidActedCnt:      g.bidActedCnt,
		HighestBid:       g.highestBid,
		HighestBidder:    g.highestBidder,
		Passed:           g.passed,
		DeclarerIdx:      g.declarerIdx,
		Contract:         g.contract,
		ContractType:     g.contractType,
		Blind:            g.blind,
		Stash:            g.stash,
		StashOwner:       g.stashOwner,
		PlayerScores:     g.playerScores,
		LastTrickWinner:  g.lastTrickWinner,
		LastTrickCards:   g.lastTrickCards,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// cegoMaxSliceLen caps slice sizes during deserialisation.
const cegoMaxSliceLen = 5000

// 各種デシリアライズ検証エラー。
var (
	errCegoOversized      = errors.New("cego: input array exceeds maximum allowed size")
	errCegoInvalidPlayers = errors.New("cego: invalid player count")
	errCegoInvalidTrick   = errors.New("cego: invalid trick card")
	errCegoInvalidCard    = errors.New("cego: invalid card element")
	errCegoInvalidIndex   = errors.New("cego: index field out of range")
	errCegoInvalidPhase   = errors.New("cego: phase out of range")
	errCegoInvalidBid     = errors.New("cego: bid value out of range")
	errCegoInvalidOutcome = errors.New("cego: outcome/result value out of range")
)

// cegoValidCard デシリアライズ時のカード妥当性を検証する (nil 拒否, 値域チェック)。
func cegoValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d, v := c.GetDesign(), c.GetValue()
	switch d {
	case CegoSkusDesign:
		return v == CegoSkusValue
	case CegoTrumpDesign:
		return v >= 1 && v <= CegoMaxTrump
	default:
		return d >= 1 && d <= CegoSuitCnt && v >= 1 && v <= CegoSuitMaxValue
	}
}

// cegoCheckCards スライスの各要素のカード妥当性を検証する。
func cegoCheckCards(cards []*Card) error {
	for _, c := range cards {
		if !cegoValidCard(c) {
			return errCegoInvalidCard
		}
	}
	return nil
}

// cegoInRange v が [0, PlayerCnt) か。
func cegoInRange(v int) bool { return v >= 0 && v < CegoPlayerCnt }

// cegoInRangeOrUnset v が -1 (未設定) または [0, PlayerCnt) か。
func cegoInRangeOrUnset(v int) bool { return v == -1 || cegoInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Cego) UnmarshalJSON(data []byte) error {
	var j cegoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cegoMaxSliceLen || len(j.CurrentTrick) > cegoMaxSliceLen ||
		len(j.ActionLog) > cegoMaxSliceLen || len(j.Blind) > cegoMaxSliceLen ||
		len(j.Stash) > cegoMaxSliceLen || len(j.Deck) > cegoMaxSliceLen ||
		len(j.LastTrickCards) > cegoMaxSliceLen {
		return errCegoOversized
	}
	if len(j.Players) != CegoPlayerCnt {
		return errCegoInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errCegoInvalidPlayers
		}
	}
	for _, c := range j.Deck {
		if !cegoValidCard(c) {
			return errCegoInvalidCard
		}
	}
	if err := cegoCheckCards(j.Blind); err != nil {
		return err
	}
	if err := cegoCheckCards(j.Stash); err != nil {
		return err
	}
	if err := cegoCheckCards(j.LastTrickCards); err != nil {
		return err
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || !cegoValidCard(tc.Card) {
			return errCegoInvalidTrick
		}
		if !cegoInRange(tc.PlayerIdx) {
			return errCegoInvalidTrick
		}
	}
	if !cegoInRange(j.CurrentPlayerIdx) || !cegoInRange(j.DealerIdx) ||
		!cegoInRange(j.BidPlayerIdx) {
		return errCegoInvalidIndex
	}
	if !cegoInRangeOrUnset(j.LeadPlayerIdx) || !cegoInRangeOrUnset(j.DeclarerIdx) ||
		!cegoInRangeOrUnset(j.HighestBidder) || !cegoInRangeOrUnset(j.LastTrickWinner) ||
		!cegoInRangeOrUnset(j.WinnerPlayer) {
		return errCegoInvalidIndex
	}
	if j.StashOwner < 0 || j.StashOwner > 1 {
		return errCegoInvalidIndex
	}
	if j.ContractType < CegoContractNone || j.ContractType > CegoContractHandspiel {
		return errCegoInvalidIndex
	}
	if int(j.Phase) < CegoPhaseMin || int(j.Phase) > CegoPhaseMax {
		return errCegoInvalidPhase
	}
	if !cegoValidBidVal(j.HighestBid) || !cegoValidBidVal(j.Contract) {
		return errCegoInvalidBid
	}
	// プレイ以降はデクレアラー・コントラクトが確定していなければならない。
	if j.Phase >= CegoPhasePlay && j.Phase <= CegoPhaseRoundEnd {
		if !cegoInRange(j.DeclarerIdx) || !cegoValidBid(j.Contract) ||
			!cegoInRange(j.LeadPlayerIdx) {
			return errCegoInvalidIndex
		}
	}
	if j.Outcome < CegoOutcomeNone || j.Outcome > CegoOutcomeLoss {
		return errCegoInvalidOutcome
	}
	if j.Result < CegoResultLose || j.Result > CegoResultWin {
		return errCegoInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.deck = j.Deck
	if g.deck == nil {
		g.deck = make([]*Card, 0)
	}
	g.deckDrawCnt = j.DeckDrawCnt
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.bidActedCnt = j.BidActedCnt
	g.highestBid = j.HighestBid
	g.highestBidder = j.HighestBidder
	g.passed = j.Passed
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.contractType = j.ContractType
	g.blind = j.Blind
	if g.blind == nil {
		g.blind = make([]*Card, 0)
	}
	g.stash = j.Stash
	if g.stash == nil {
		g.stash = make([]*Card, 0)
	}
	g.stashOwner = j.StashOwner
	g.playerScores = j.PlayerScores
	g.lastTrickWinner = j.LastTrickWinner
	g.lastTrickCards = j.LastTrickCards
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
