//go:build !js || !wasm || extra

// Package domain ケーニッヒルーフェン (Königrufen / Tarock) のドメインモデル。
//
// Königrufen はオーストリアの 54 枚タロック (Tarock) デッキを用いる 4 人用コントラクト・
// トリックテイキング。本実装は「王呼び (Rufer / call-a-king)」コントラクトのみを扱う MVP で、
// 1 人のデクレアラー (declarer) が呼んだ王を持つ秘密のパートナーと組み、残り 2 人 (王が場札
// またはデクレアラー自身にある場合は残り 3 人) と対戦する 2 対 2 (または 1 対 3) の力学を持つ。
//
// # デッキ (54 枚)
//
//   - スート札 32 枚: design = 1..4 (4 スート)、value = 1..8。1..4 はピップ札、
//     5 = ジャック (Jack)、6 = カヴァリエ (Cavalier)、7 = クイーン (Queen)、8 = キング (King)。
//     各スートの上位 4 枚 (value 5..8) がコート札 J/C/Q/K。
//   - 切り札 (Tarock) 21 枚: design = KoenigrufenTrumpDesign (5)、value = 1..21。
//   - スキュース (Sküs / Fool) 1 枚: design = KoenigrufenSkusDesign (6)、value = 0。
//     フレンチタロットのエクスキューズと異なり **最強の切り札** として振る舞い、トリックを
//     勝ち取る (フォロー免除やカード返却は行わない)。トゥルル (Trull) 名誉札は切り札 I (Pagat)・
//     切り札 XXI・スキュースの 3 枚。
//
// # 簡略化ルール (本実装が採用する縮小版) — すべて MVP のための省略
//
//   - 配札: 各プレイヤーに 12 枚 (48 枚) + 6 枚の場札 (talon)。3 枚パケットで配る。
//   - 入札 (Bidding): 本来のオーストリアのコントラクト梯子 (Dreier / Solo / Bettel / Piccolo …)
//     は **省略**し、単一の Rufer コントラクトのみを扱う。1 巡でパスまたは Rufer を宣言し、
//     最初に Rufer を宣言した者がデクレアラー (Rufer は 1 段のみなので後続は上回れない)。全員
//     パスならディーラーの左隣が強制デクレアラー (再配札は行わない — 簡略化)。倍率は常に 1。
//   - 王呼び (Call a king): デクレアラーが自分の持たないスートのキングを 1 枚指名し、その
//     キングを持つプレイヤーが秘密のパートナーになる。**パートナーの正体はキングが場に出る
//     まで (traditionally) 秘匿**され、Web 出力に漏れてはならない (partnerIdx / isPartner は
//     partnerRevealed=false の間は隠す)。デクレアラーが 4 枚のキングをすべて持つ場合は自動的に
//     単独 (solo, partnerIdx=-1)。呼んだキングが場札にある場合はパートナー不在 (partnerIdx=-1、
//     単独扱い) とする — 簡略化。
//   - 場札交換 (Talon exchange): デクレアラーが 6 枚の場札をすべて手に取り (18 枚)、6 枚を
//     伏せて捨てる (フレンチタロットのシアン écart を踏襲)。キング・トゥルル (Pagat/XXI/Sküs)
//     は捨てられない。切り札は他に捨てられる札が足りない場合のみ許可。捨て札はデクレアラー側の
//     得点札に計上。
//   - トリックプレイ (12 トリック): リードスートに従う義務。ボイド時は切り札 (Tarock) を出す
//     義務。切り札が場に出ていれば可能な限り上位切り札を出す義務 (オーバートランプ)。スキュースは
//     最強の切り札。最強切り札が勝ち、なければリードスートの最高札が勝つ。
//   - カードポイント (簡略化した 1 枚ごとの整数配点、合計 KoenigrufenTotalPoints=106):
//     キング=5、クイーン=4、カヴァリエ=3、ジャック=2、トゥルル (Pagat I・XXI・Sküs)=各 5、
//     その他 (ピップ 1-4・素の切り札 2-20)=1。本来の「3 枚ずつ数えて 2 を引く」正式計算は
//     **省略**し、相対的な名誉札の価値のみ保持する。両チームの合計は必ずデッキ総計に一致する。
//   - 得点: デクレアラー側 (デクレアラー + 秘密のパートナー) の獲得点が過半 (2×teamPoints > 106、
//     すなわち 54 点以上) なら成功。ゼロサムで ± 精算する。ボーナス (Trull / König / Pagat
//     ultimo) は **省略**。
//   - 累積得点: TargetDeals ディール後、累積最上位が勝者。KoenigrufenResult は人間視点。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// KoenigrufenPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const KoenigrufenPlayerCnt = 4

// KoenigrufenHandSize 各プレイヤーの配り札枚数
const KoenigrufenHandSize = 12

// KoenigrufenTalonSize 場札 (talon) の枚数
const KoenigrufenTalonSize = 6

// KoenigrufenDeckSize デッキ枚数 (54 枚タロックデッキ)
const KoenigrufenDeckSize = 54

// KoenigrufenTrickCount 1 ディールのトリック数
const KoenigrufenTrickCount = 12

// KoenigrufenDefaultDeals マッチを構成するディール数 (既定)
const KoenigrufenDefaultDeals = 4

// KoenigrufenSuitCnt スート数
const KoenigrufenSuitCnt = 4

// KoenigrufenTrumpDesign 切り札 (Tarock) の仮想デザイン値。1..4 はスート、5 が切り札。
const KoenigrufenTrumpDesign = 5

// KoenigrufenSkusDesign スキュース (Sküs / Fool) の仮想デザイン値。
const KoenigrufenSkusDesign = 6

// KoenigrufenSkusValue スキュースのカード値。
const KoenigrufenSkusValue = 0

// KoenigrufenSkusRank スキュースをトリック比較で扱う際の切り札ランク (最強、21 より上)。
const KoenigrufenSkusRank = 22

// KoenigrufenMaxTrump 切り札の最大値 (21)。
const KoenigrufenMaxTrump = 21

// KoenigrufenPagatValue パガト (最小の切り札トゥルル, 切り札 I) の値。
const KoenigrufenPagatValue = 1

// KoenigrufenKingValue スート札のキング (King) の値。
const KoenigrufenKingValue = 8

// KoenigrufenSuitMaxValue スート札の最大値 (キング)。
const KoenigrufenSuitMaxValue = 8

// KoenigrufenTotalPoints デッキ総カードポイント (簡略化した配点の合計)。
const KoenigrufenTotalPoints = 106

// KoenigrufenBaseGameValue 精算の基礎ゲーム価値。
const KoenigrufenBaseGameValue = 10

// KoenigrufenBid 入札 (コントラクト) 種別
type KoenigrufenBid int

// Königrufen の入札定数 (値が大きいほど高い入札)
const (
	// KoenigrufenBidPass パス / 未入札
	KoenigrufenBidPass KoenigrufenBid = 0
	// KoenigrufenBidRufer ルーファー (王呼びコントラクト) — 本実装唯一のコントラクト
	KoenigrufenBidRufer KoenigrufenBid = 1
)

// KoenigrufenPhase ゲームフェーズ
type KoenigrufenPhase int

// Königrufen のフェーズ定数
const (
	// KoenigrufenPhaseBid 入札フェーズ
	KoenigrufenPhaseBid KoenigrufenPhase = 0
	// KoenigrufenPhaseCall 王呼びフェーズ
	KoenigrufenPhaseCall KoenigrufenPhase = 1
	// KoenigrufenPhaseTalon 場札交換 (discard) フェーズ
	KoenigrufenPhaseTalon KoenigrufenPhase = 2
	// KoenigrufenPhasePlay トリックプレイフェーズ
	KoenigrufenPhasePlay KoenigrufenPhase = 3
	// KoenigrufenPhaseTrickEnd トリック終了フェーズ
	KoenigrufenPhaseTrickEnd KoenigrufenPhase = 4
	// KoenigrufenPhaseRoundEnd ディール終了フェーズ
	KoenigrufenPhaseRoundEnd KoenigrufenPhase = 5
	// KoenigrufenPhaseGameEnd ゲーム終了フェーズ
	KoenigrufenPhaseGameEnd KoenigrufenPhase = 6
)

// KoenigrufenPhaseMin フェーズ下限 (検証用)
const KoenigrufenPhaseMin = int(KoenigrufenPhaseBid)

// KoenigrufenPhaseMax フェーズ上限 (検証用)
const KoenigrufenPhaseMax = int(KoenigrufenPhaseGameEnd)

// KoenigrufenOutcome ディール結果 (デクレアラー側視点)
type KoenigrufenOutcome int

// Königrufen のディール結果定数
const (
	// KoenigrufenOutcomeNone 未確定
	KoenigrufenOutcomeNone KoenigrufenOutcome = 0
	// KoenigrufenOutcomeWin デクレアラー側がコントラクトを達成
	KoenigrufenOutcomeWin KoenigrufenOutcome = 1
	// KoenigrufenOutcomeLoss デクレアラー側がコントラクトを失敗
	KoenigrufenOutcomeLoss KoenigrufenOutcome = 2
)

// KoenigrufenResult 人間視点のマッチ結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type KoenigrufenResult int

// Königrufen のマッチ結果定数
const (
	// KoenigrufenResultLose 敗北
	KoenigrufenResultLose KoenigrufenResult = -1
	// KoenigrufenResultNone 未確定 / 引き分け
	KoenigrufenResultNone KoenigrufenResult = 0
	// KoenigrufenResultWin 勝利
	KoenigrufenResultWin KoenigrufenResult = 1
)

// KoenigrufenHint ヒント情報
type KoenigrufenHint struct {
	Bid         *int   // 推奨入札 (入札フェーズ)。nil の場合はパス推奨
	CallSuit    *int   // 推奨呼び王スート (王呼びフェーズ)
	CardIndices []int  // 推奨カードインデックス (交換/プレイ)
	Reason      string // ヒント理由キー
}

// KoenigrufenBreakdown 得点計算の内訳 (純粋関数 koenigrufenScoreDeal の出力)。
type KoenigrufenBreakdown struct {
	// TeamPoints デクレアラー側 (デクレアラー + パートナー) が獲得したカードポイント合計。
	TeamPoints int
	// Threshold 「過半」の閾値 (この値を超える = 成功)。
	Threshold int
	// Won コントラクト成否。
	Won bool
	// Solo デクレアラーが単独 (パートナー不在) か。
	Solo bool
	// Diff 閾値との差 (絶対値、整数点)。
	Diff int
	// Base (KoenigrufenBaseGameValue + Diff) × Mult。
	Base int
	// Mult 入札倍率 (常に 1)。
	Mult int
	// DeclarerScore デクレアラーの得点変動。
	DeclarerScore int
	// PartnerScore パートナーの得点変動 (Solo なら 0)。
	PartnerScore int
	// OpponentScore 対戦側 1 人の得点変動。
	OpponentScore int
}

// Koenigrufen ケーニッヒルーフェンのゲームクラス
type Koenigrufen struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*KoenigrufenPlayer
	config           KoenigrufenConfig
	phase            KoenigrufenPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	// --- bidding state ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    KoenigrufenBid
	highestBidder int
	passed        [KoenigrufenPlayerCnt]bool
	// --- contract state ---
	declarerIdx int
	contract    KoenigrufenBid
	// --- call-a-king state (partnerIdx はサーバー側のみ、Web 出力に漏らさない) ---
	calledKing      int // 呼んだキングのスート (1..4)、未呼び/-1
	partnerIdx      int // 秘密のパートナー (-1=単独 solo)
	partnerRevealed bool
	// --- talon state ---
	talon      []*Card // 場札 (6 枚)
	stash      []*Card // 得点計上用に脇へ置いた 6 枚 (デクレアラー側)
	stashOwner int     // 0 = デクレアラー側 (Königrufen では常に 0)
	// --- scoring ---
	playerScores    [KoenigrufenPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         KoenigrufenOutcome
	result          KoenigrufenResult
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int
	actionLogBase
}

// NewKoenigrufen コンストラクタ
func NewKoenigrufen(players []*KoenigrufenPlayer, config KoenigrufenConfig) *Koenigrufen {
	return &Koenigrufen{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        KoenigrufenBidPass,
		calledKing:      -1,
		partnerIdx:      -1,
		stashOwner:      0,
	}
}

// NewDefaultKoenigrufen 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultKoenigrufen() *Koenigrufen {
	players := make([]*KoenigrufenPlayer, KoenigrufenPlayerCnt)
	players[0] = NewKoenigrufenPlayer(true)
	for i := 1; i < KoenigrufenPlayerCnt; i++ {
		players[i] = NewKoenigrufenPlayer(false)
	}
	return NewKoenigrufen(players, DefaultKoenigrufenConfig())
}

// buildKoenigrufenDeck 54 枚タロックデッキを直接構築する。スート札 (design 1..4, value 1..8)
// 32 枚 + 切り札 (design 5, value 1..21) 21 枚 + スキュース (design 6, value 0)。
func buildKoenigrufenDeck() []*Card {
	deck := make([]*Card, 0, KoenigrufenDeckSize)
	for suit := 1; suit <= KoenigrufenSuitCnt; suit++ {
		for val := 1; val <= KoenigrufenSuitMaxValue; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= KoenigrufenMaxTrump; val++ {
		deck = append(deck, NewCard(KoenigrufenTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(KoenigrufenSkusDesign, KoenigrufenSkusValue, false))
	return deck
}

// Reset ゲーム初期化
func (g *Koenigrufen) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [KoenigrufenPlayerCnt]int{}
	g.result = KoenigrufenResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Koenigrufen) NextRound() {
	if g.phase != KoenigrufenPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % KoenigrufenPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Koenigrufen) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = KoenigrufenBidPass
	g.calledKing = -1
	g.partnerIdx = -1
	g.partnerRevealed = false
	g.talon = nil
	g.stash = nil
	g.stashOwner = 0
	g.outcome = KoenigrufenOutcomeNone
	g.scored = false
	g.passed = [KoenigrufenPlayerCnt]bool{}
	g.highestBid = KoenigrufenBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % KoenigrufenPlayerCnt
	g.phase = KoenigrufenPhaseBid
}

// deal 3 枚パケットで各プレイヤーへ 12 枚を配り、場札 6 枚を脇に置く。
func (g *Koenigrufen) deal() {
	g.deck = buildKoenigrufenDeck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	g.talon = make([]*Card, 0, KoenigrufenTalonSize)
	packets := KoenigrufenHandSize / 3 // 4 ラウンド
	for r := 0; r < packets; r++ {
		for j := 0; j < KoenigrufenPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % KoenigrufenPlayerCnt
			for k := 0; k < 3; k++ {
				if c := g.drawCard(); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	// 残り 6 枚を場札 (talon) とする。
	for k := 0; k < KoenigrufenTalonSize; k++ {
		if c := g.drawCard(); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Koenigrufen) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// --- Bidding ---

// PlayerBid 人間プレイヤーが入札する。
func (g *Koenigrufen) PlayerBid(bid KoenigrufenBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KoenigrufenPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	if !koenigrufenValidBid(bid) {
		return NewDomainError(ErrInvalidPlay, "無効な入札です (rufer)")
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay, "現在の入札より高い入札が必要です")
	}
	g.applyBid(g.bidPlayerIdx, bid)
	return nil
}

// PlayerPass 人間プレイヤーがパスする。
func (g *Koenigrufen) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KoenigrufenPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU プレイヤーが 1 回入札する (入札 or パス)。
func (g *Koenigrufen) CpuBid() {
	if g.gameEndFlag || g.phase != KoenigrufenPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= KoenigrufenPlayerCnt {
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
func (g *Koenigrufen) applyBid(idx int, bid KoenigrufenBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", playerName(g.players, idx), koenigrufenBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを適用する。
func (g *Koenigrufen) applyPass(idx int) {
	g.passed[idx] = true
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	g.advanceBid()
}

// advanceBid 入札を次のプレイヤーへ進め、1 巡終了でコントラクトを確定する。
func (g *Koenigrufen) advanceBid() {
	g.bidActedCnt++
	if g.bidActedCnt >= KoenigrufenPlayerCnt {
		g.finalizeBid()
		return
	}
	g.bidPlayerIdx = (g.bidPlayerIdx + 1) % KoenigrufenPlayerCnt
}

// finalizeBid 入札を確定し、デクレアラーを決定して王呼びへ進む。全員パスならディーラーの左隣を
// 強制デクレアラーとする (再配札なし)。
func (g *Koenigrufen) finalizeBid() {
	if g.highestBidder < 0 {
		g.declarerIdx = (g.dealerIdx + 1) % KoenigrufenPlayerCnt
		g.contract = KoenigrufenBidRufer
		g.appendLog(g.declarerIdx, "forced",
			fmt.Sprintf("all passed — %s is forced to take Rufer", playerName(g.players, g.declarerIdx)), nil)
	} else {
		g.declarerIdx = g.highestBidder
		g.contract = g.highestBid
		g.appendLog(g.declarerIdx, "win_bid",
			fmt.Sprintf("%s takes the contract %s", playerName(g.players, g.declarerIdx), koenigrufenBidName(g.contract)), nil)
	}
	g.enterCallOrSolo()
}

// enterCallOrSolo デクレアラーが 4 枚のキングをすべて持つ場合は自動的に単独 (solo) として
// 王呼びを飛ばし、そうでなければ王呼びフェーズへ入る。
func (g *Koenigrufen) enterCallOrSolo() {
	if g.declarerKingsHeld() == KoenigrufenSuitCnt {
		g.calledKing = -1
		g.partnerIdx = -1
		g.appendLog(g.declarerIdx, "solo",
			fmt.Sprintf("%s holds all four kings and plays solo", playerName(g.players, g.declarerIdx)), nil)
		g.finalizeCall()
		return
	}
	g.currentPlayerIdx = g.declarerIdx
	g.phase = KoenigrufenPhaseCall
}

// --- Call a king ---

// PlayerCallKing 人間デクレアラーが呼ぶキングのスート (1..4) を指名する。
func (g *Koenigrufen) PlayerCallKing(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KoenigrufenPhaseCall {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if suit < 1 || suit > KoenigrufenSuitCnt {
		return NewDomainError(ErrInvalidPlay, "スートは 1..4 で指定してください")
	}
	if g.playerHoldsKing(g.declarerIdx, suit) {
		return NewDomainError(ErrInvalidPlay, "自分が持つキングは呼べません")
	}
	g.applyCallKing(suit)
	return nil
}

// CpuCallKing CPU デクレアラーが呼ぶキングを自動選択する。
func (g *Koenigrufen) CpuCallKing() {
	if g.gameEndFlag || g.phase != KoenigrufenPhaseCall {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	g.applyCallKing(g.cpuSelectCallKing())
}

// applyCallKing 呼び王を確定し、秘密のパートナーを決定して場札交換へ進む。パートナーの正体は
// partnerRevealed が true になるまで秘匿される (Web 出力へは漏らさない)。
func (g *Koenigrufen) applyCallKing(suit int) {
	g.calledKing = suit
	g.partnerIdx = g.findKingHolder(suit)
	// ログにはパートナーの正体を書かず、呼んだキングのみを記録する。
	g.appendLog(g.declarerIdx, "call_king",
		fmt.Sprintf("%s calls the King of suit %d", playerName(g.players, g.declarerIdx), suit), nil)
	g.finalizeCall()
}

// finalizeCall 場札をデクレアラーの手札に加え、場札交換 (discard) フェーズへ進む。
func (g *Koenigrufen) finalizeCall() {
	for _, c := range g.talon {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.talon = make([]*Card, 0)
	g.sortAllHands()
	g.currentPlayerIdx = g.declarerIdx
	g.phase = KoenigrufenPhaseTalon
}

// declarerKingsHeld デクレアラーが持つキングのスート数を返す。
func (g *Koenigrufen) declarerKingsHeld() int {
	if g.declarerIdx < 0 {
		return 0
	}
	cnt := 0
	for suit := 1; suit <= KoenigrufenSuitCnt; suit++ {
		if g.playerHoldsKing(g.declarerIdx, suit) {
			cnt++
		}
	}
	return cnt
}

// playerHoldsKing playerIdx が指定スートのキングを持つか。
func (g *Koenigrufen) playerHoldsKing(playerIdx, suit int) bool {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return false
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if !koenigrufenIsTrumpLike(c) && c.GetDesign() == suit && c.GetValue() == KoenigrufenKingValue {
			return true
		}
	}
	return false
}

// findKingHolder 指定スートのキングを持つ非デクレアラーを返す (-1=場札にあり単独扱い)。
func (g *Koenigrufen) findKingHolder(suit int) int {
	for i := 0; i < len(g.players); i++ {
		if i == g.declarerIdx {
			continue
		}
		if g.playerHoldsKing(i, suit) {
			return i
		}
	}
	return -1
}

// cpuSelectCallKing CPU デクレアラーが呼ぶキングのスートを選ぶ (自分が持たない最初のスート)。
func (g *Koenigrufen) cpuSelectCallKing() int {
	for suit := 1; suit <= KoenigrufenSuitCnt; suit++ {
		if !g.playerHoldsKing(g.declarerIdx, suit) {
			return suit
		}
	}
	return 1
}

// --- Talon exchange (discard) ---

// PlayerDiscard 人間デクレアラーが場札交換で 6 枚を伏せて捨てる。
func (g *Koenigrufen) PlayerDiscard(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KoenigrufenPhaseTalon {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doDiscard(cardIndices)
}

// CpuDiscard CPU デクレアラーが場札交換で 6 枚を自動で捨てる。
func (g *Koenigrufen) CpuDiscard() {
	if g.gameEndFlag || g.phase != KoenigrufenPhaseTalon {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	_ = g.doDiscard(g.cpuSelectDiscards(g.declarerIdx))
}

// doDiscard 場札交換の共通処理。捨てた 6 枚をデクレアラー側の得点札 (stash) とする。
func (g *Koenigrufen) doDiscard(cardIndices []int) error {
	player := g.players[g.declarerIdx]
	if len(cardIndices) != KoenigrufenTalonSize {
		return NewDomainError(ErrInvalidCard, "ちょうど 6 枚を捨ててください")
	}
	seen := make(map[int]bool, KoenigrufenTalonSize)
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードを 2 回選べません")
		}
		seen[idx] = true
	}
	if err := g.validateDiscards(player, cardIndices); err != nil {
		return err
	}
	discarded := player.RemoveCards(cardIndices)
	g.stash = discarded
	g.stashOwner = 0
	g.appendLog(g.declarerIdx, "discard",
		fmt.Sprintf("%s discards %d cards to the talon", playerName(g.players, g.declarerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlay()
	return nil
}

// validateDiscards 場札交換の合法性を検証する。キング・トゥルル (Pagat/XXI/Sküs) は不可。切り札は
// 捨てられる非切り札・非キング札が 6 枚に満たない場合のみ許可。
func (g *Koenigrufen) validateDiscards(player *KoenigrufenPlayer, cardIndices []int) error {
	discardable := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if koenigrufenDiscardable(player.GetCard(i)) {
			discardable++
		}
	}
	allowTrump := discardable < KoenigrufenTalonSize
	for _, idx := range cardIndices {
		c := player.GetCard(idx)
		if koenigrufenIsKing(c) {
			return NewDomainError(ErrInvalidPlay, "キングは捨てられません")
		}
		if koenigrufenIsTrull(c) {
			return NewDomainError(ErrInvalidPlay, "トゥルル (Pagat/XXI/Sküs) は捨てられません")
		}
		if koenigrufenIsTrumpLike(c) && !allowTrump {
			return NewDomainError(ErrInvalidPlay, "切り札は (やむを得ない場合を除き) 捨てられません")
		}
	}
	return nil
}

// koenigrufenDiscardable 通常の場札交換に出せる札か (非切り札・非スキュース・非キング)。
func koenigrufenDiscardable(c *Card) bool {
	if c == nil || koenigrufenIsTrumpLike(c) {
		return false
	}
	return c.GetValue() != KoenigrufenKingValue
}

// --- Play ---

// startPlay プレイフェーズを開始する。エルダー (ディーラーの左隣) がリードする。
func (g *Koenigrufen) startPlay() {
	g.sortAllHands()
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % KoenigrufenPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = KoenigrufenPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Koenigrufen) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KoenigrufenPhasePlay {
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
func (g *Koenigrufen) CpuPlay() {
	if g.gameEndFlag || g.phase != KoenigrufenPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
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

// playCard カードをプレイする共通処理。呼ばれたキングが出たらパートナーを公開する。
func (g *Koenigrufen) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	if koenigrufenIsCalledKing(card, g.calledKing) {
		g.partnerRevealed = true
	}
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), koenigrufenCardStr(card)), []*Card{card})
	if len(g.currentTrick) == KoenigrufenPlayerCnt {
		g.phase = KoenigrufenPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % KoenigrufenPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。全札 (スキュースを含む) を勝者のトリック山へ。
// 最終トリックなら RoundEnd に入り得点計算を発火する。
func (g *Koenigrufen) ResolveTrick() {
	if g.phase != KoenigrufenPhaseTrickEnd || len(g.currentTrick) != KoenigrufenPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	allCards := make([]*Card, 0, KoenigrufenPlayerCnt)
	for _, tc := range g.currentTrick {
		allCards = append(allCards, tc.Card)
	}
	g.players[winnerIdx].AddTrick(allCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), allCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= KoenigrufenTrickCount {
		g.lastTrickWinner = winnerIdx
		g.lastTrickCards = allCards
		g.phase = KoenigrufenPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = KoenigrufenPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Koenigrufen) NextTrick() {
	if g.phase != KoenigrufenPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = KoenigrufenPhasePlay
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ)。
func (g *Koenigrufen) ScoreRound() {
	if g.phase != KoenigrufenPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd RoundEnd 突入時に一度だけ得点計算と精算を行う (scored フラグでガード)。ここで
// パートナーを公開する。
func (g *Koenigrufen) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.partnerRevealed = true
	bd := g.computeBreakdown()
	if bd.Won {
		g.outcome = KoenigrufenOutcomeWin
	} else {
		g.outcome = KoenigrufenOutcomeLoss
	}
	for i := 0; i < KoenigrufenPlayerCnt; i++ {
		switch {
		case i == g.declarerIdx:
			g.playerScores[i] += bd.DeclarerScore
		case !bd.Solo && i == g.partnerIdx:
			g.playerScores[i] += bd.PartnerScore
		default:
			g.playerScores[i] += bd.OpponentScore
		}
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: declarer(%s) %s teamPts=%d/%d won=%t base=%d",
			g.roundNumber, playerName(g.players, g.declarerIdx), koenigrufenBidName(g.contract),
			bd.TeamPoints, KoenigrufenTotalPoints, bd.Won, bd.Base), nil)
	g.checkGameEnd()
}

// computeBreakdown 現在のディールの得点内訳を計算する。
func (g *Koenigrufen) computeBreakdown() KoenigrufenBreakdown {
	teamPts := g.teamCaptured()
	solo := g.partnerIdx < 0
	return koenigrufenScoreDeal(teamPts, solo, koenigrufenBidMult(g.contract))
}

// teamCaptured デクレアラー側 (デクレアラー + パートナー + stash) が獲得したカードポイントを返す。
func (g *Koenigrufen) teamCaptured() int {
	pts := 0
	pts += g.playerTrickPoints(g.declarerIdx)
	if g.partnerIdx >= 0 && g.partnerIdx != g.declarerIdx {
		pts += g.playerTrickPoints(g.partnerIdx)
	}
	for _, c := range g.stash {
		pts += koenigrufenCardPoints(c)
	}
	return pts
}

// playerTrickPoints プレイヤー idx が獲得したトリック山のカードポイント合計を返す。
func (g *Koenigrufen) playerTrickPoints(idx int) int {
	if idx < 0 || idx >= len(g.players) {
		return 0
	}
	pts := 0
	for _, trick := range g.players[idx].GetTricksTaken() {
		for _, c := range trick {
			pts += koenigrufenCardPoints(c)
		}
	}
	return pts
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積得点最上位を勝者とする。
func (g *Koenigrufen) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < KoenigrufenPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.phase = KoenigrufenPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	if tie {
		g.winnerPlayer = -1
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
	} else {
		g.winnerPlayer = leader
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
	}
}

// humanResult 人間 (seat 0) 視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None。
func (g *Koenigrufen) humanResult(leader int, tie bool) KoenigrufenResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return KoenigrufenResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return KoenigrufenResultNone
		}
		return KoenigrufenResultWin
	}
	return KoenigrufenResultLose
}

// --- Scoring helper (pure) ---

// koenigrufenBidMult 入札倍率を返す (Rufer=1)。
func koenigrufenBidMult(_ KoenigrufenBid) int { return 1 }

// koenigrufenScoreDeal ディール得点を計算する純粋関数。teamPoints はデクレアラー側の獲得
// カードポイント、solo は単独か、mult は倍率。デクレアラー側が過半 (2×teamPoints >
// KoenigrufenTotalPoints) を取れば成功。精算はゼロサム。
//
//   - 単独 (1 対 3): デクレアラー ±3×base、対戦側各人 ∓base。
//   - 組 (2 対 2): デクレアラー ±base、パートナー ±base、対戦側各人 ∓base。
func koenigrufenScoreDeal(teamPoints int, solo bool, mult int) KoenigrufenBreakdown {
	threshold := KoenigrufenTotalPoints / 2 // 53 — これを超えれば成功
	won := 2*teamPoints > KoenigrufenTotalPoints
	diff := teamPoints - threshold
	if diff < 0 {
		diff = -diff
	}
	base := (KoenigrufenBaseGameValue + diff) * mult
	winSign := 1
	if !won {
		winSign = -1
	}
	bd := KoenigrufenBreakdown{
		TeamPoints: teamPoints,
		Threshold:  threshold,
		Won:        won,
		Solo:       solo,
		Diff:       diff,
		Base:       base,
		Mult:       mult,
	}
	if solo {
		bd.DeclarerScore = winSign * 3 * base
		bd.PartnerScore = 0
		bd.OpponentScore = -winSign * base
	} else {
		bd.DeclarerScore = winSign * base
		bd.PartnerScore = winSign * base
		bd.OpponentScore = -winSign * base
	}
	return bd
}

// --- Card classification / points ---

// koenigrufenIsTrump 切り札 (Tarock, design 5) か。
func koenigrufenIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == KoenigrufenTrumpDesign
}

// koenigrufenIsSkus スキュース (design 6) か。
func koenigrufenIsSkus(c *Card) bool {
	return c != nil && c.GetDesign() == KoenigrufenSkusDesign
}

// koenigrufenIsTrumpLike 切り札またはスキュース (トリックで切り札として振る舞う札) か。
func koenigrufenIsTrumpLike(c *Card) bool {
	return koenigrufenIsTrump(c) || koenigrufenIsSkus(c)
}

// koenigrufenIsKing スート札のキングか。
func koenigrufenIsKing(c *Card) bool {
	return c != nil && !koenigrufenIsTrumpLike(c) && c.GetValue() == KoenigrufenKingValue
}

// koenigrufenIsTrull トゥルル (Pagat I / 切り札 XXI / Sküs) か。
func koenigrufenIsTrull(c *Card) bool {
	if c == nil {
		return false
	}
	if koenigrufenIsSkus(c) {
		return true
	}
	return koenigrufenIsTrump(c) && (c.GetValue() == KoenigrufenPagatValue || c.GetValue() == KoenigrufenMaxTrump)
}

// koenigrufenIsCalledKing card が呼ばれたスートのキングか。
func koenigrufenIsCalledKing(c *Card, calledSuit int) bool {
	return calledSuit >= 1 && koenigrufenIsKing(c) && c.GetDesign() == calledSuit
}

// koenigrufenTrumpValue 切り札札のトリック比較用の値を返す (スキュース=22 で最強、非切り札=0)。
func koenigrufenTrumpValue(c *Card) int {
	if koenigrufenIsSkus(c) {
		return KoenigrufenSkusRank
	}
	if koenigrufenIsTrump(c) {
		return c.GetValue()
	}
	return 0
}

// koenigrufenCardPoints カードの (簡略化した) カードポイントを返す。
// キング=5、クイーン=4、カヴァリエ=3、ジャック=2、トゥルル=5、その他=1。
func koenigrufenCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if koenigrufenIsTrull(c) {
		return 5
	}
	if koenigrufenIsTrumpLike(c) {
		return 1
	}
	switch c.GetValue() {
	case KoenigrufenKingValue: // King
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

// ledSuit 現在のトリックのリードスート (実効スート) を返す。最初の札が切り札系なら
// KoenigrufenTrumpDesign、それ以外はその design。トリックが空なら -1。
func (g *Koenigrufen) ledSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0].Card
	if koenigrufenIsTrumpLike(first) {
		return KoenigrufenTrumpDesign
	}
	return first.GetDesign()
}

// highestTrumpInTrick 現在のトリック中の最強切り札の値を返す (0=切り札なし、スキュース=22)。
func (g *Koenigrufen) highestTrumpInTrick() int {
	best := 0
	for _, tc := range g.currentTrick {
		if v := koenigrufenTrumpValue(tc.Card); v > best {
			best = v
		}
	}
	return best
}

// validatePlay マストフォロー + 切り札義務 + オーバートランプ義務を検証する。
func (g *Koenigrufen) validatePlay(playerIdx int, card *Card) error {
	return validateCardIsPlayable(g.getValidPlayIndices(playerIdx), g.players[playerIdx], card)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Koenigrufen) getValidPlayIndices(playerIdx int) []int {
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
	if led == KoenigrufenTrumpDesign {
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
func (g *Koenigrufen) trumpFollowIndices(player *KoenigrufenPlayer, highestTrump int) []int {
	trumps := g.trumpLikeIndices(player)
	if len(trumps) == 0 {
		return g.allIndices(player) // 切り札なし → 任意の札
	}
	higher := koenigrufenFilter(trumps, func(idx int) bool {
		return koenigrufenTrumpValue(player.GetCard(idx)) > highestTrump
	})
	if len(higher) > 0 {
		return higher // オーバートランプ義務
	}
	return trumps
}

// suitFollowIndices スートがリードされた場合の合法札を返す。
func (g *Koenigrufen) suitFollowIndices(player *KoenigrufenPlayer, led, highestTrump int) []int {
	ledCards := g.suitOf(player, led)
	if len(ledCards) > 0 {
		return ledCards // フォロー義務
	}
	trumps := g.trumpLikeIndices(player)
	if len(trumps) == 0 {
		return g.allIndices(player) // ボイド + 切り札なし → 任意
	}
	higher := koenigrufenFilter(trumps, func(idx int) bool {
		return koenigrufenTrumpValue(player.GetCard(idx)) > highestTrump
	})
	if highestTrump > 0 && len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitOf 指定スート design (非切り札系) の手札インデックスを返す。
func (g *Koenigrufen) suitOf(player *KoenigrufenPlayer, design int) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		if koenigrufenIsTrumpLike(c) {
			continue
		}
		if c.GetDesign() == design {
			out = append(out, i)
		}
	}
	return out
}

// trumpLikeIndices 切り札系 (切り札 + スキュース) の手札インデックスを返す。
func (g *Koenigrufen) trumpLikeIndices(player *KoenigrufenPlayer) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if koenigrufenIsTrumpLike(player.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// allIndices 全手札インデックスを返す。
func (g *Koenigrufen) allIndices(player *KoenigrufenPlayer) []int {
	out := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// trickWinner トリックの勝者を決定する。切り札系があれば最強切り札、無ければリードスートの最強札。
func (g *Koenigrufen) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.ledSuit()
	winIdx := g.currentTrick[0].PlayerIdx
	winRank := -1
	for _, tc := range g.currentTrick {
		r := koenigrufenWinRank(tc.Card, led)
		if r > winRank {
			winRank = r
			winIdx = tc.PlayerIdx
		}
	}
	return winIdx
}

// koenigrufenWinRank トリック勝敗比較用のランクを返す (高いほど強い)。切り札系 = 1000+切り札値
// (スキュース=1022 で最強)、リードスート = 値、それ以外 = -1。
func koenigrufenWinRank(c *Card, led int) int {
	if c == nil {
		return -1
	}
	if koenigrufenIsTrumpLike(c) {
		return 1000 + koenigrufenTrumpValue(c)
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return -1
}

// --- CPU AI ---

// cpuSelectBid CPU の入札選択 (ok=false でパス)。手札評価が閾値以上で Rufer 未確定なら Rufer。
func (g *Koenigrufen) cpuSelectBid(playerIdx int) (KoenigrufenBid, bool) {
	if g.highestBid >= KoenigrufenBidRufer {
		return KoenigrufenBidPass, false
	}
	strength := g.evalHand(playerIdx)
	base := 20
	switch g.config.CpuDifficulty {
	case KoenigrufenCpuDifficultyEasy:
		base = 26
	case KoenigrufenCpuDifficultyHard:
		base = 16
	}
	if strength >= base {
		return KoenigrufenBidRufer, true
	}
	return KoenigrufenBidPass, false
}

// evalHand 手札の強さを大まかに見積もる (トゥルル・切り札枚数・キング・高位切り札から算出)。
func (g *Koenigrufen) evalHand(playerIdx int) int {
	p := g.players[playerIdx]
	score := 0
	trumps := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		switch {
		case koenigrufenIsTrull(c):
			score += 6
		case koenigrufenIsTrumpLike(c):
			trumps++
			if c.GetValue() >= 15 {
				score += 2
			} else {
				score++
			}
		case koenigrufenIsKing(c):
			score += 3
		case c.GetValue() == 7: // Queen
			score += 2
		}
	}
	score += trumps
	return score
}

// cpuSelectDiscards CPU デクレアラーが捨てる 6 枚のインデックスを選ぶ。価値の低い札から捨て、
// キング・トゥルル・切り札は温存する。
func (g *Koenigrufen) cpuSelectDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	keepValue := func(c *Card) int {
		if c == nil {
			return 0
		}
		if koenigrufenIsTrull(c) {
			return 100000
		}
		if koenigrufenIsKing(c) {
			return 90000
		}
		if koenigrufenIsTrumpLike(c) {
			return 10000 + c.GetValue()
		}
		return c.GetValue()*10 + koenigrufenCardPoints(c)
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return keepValue(p.GetCard(idxs[a])) < keepValue(p.GetCard(idxs[b]))
	})
	discardable := make([]int, 0, n)
	trumpFallback := make([]int, 0, n)
	for _, idx := range idxs {
		c := p.GetCard(idx)
		if koenigrufenDiscardable(c) {
			discardable = append(discardable, idx)
		} else if koenigrufenIsTrumpLike(c) && !koenigrufenIsTrull(c) {
			trumpFallback = append(trumpFallback, idx)
		}
	}
	chosen := make([]int, 0, KoenigrufenTalonSize)
	for _, idx := range discardable {
		if len(chosen) >= KoenigrufenTalonSize {
			break
		}
		chosen = append(chosen, idx)
	}
	for _, idx := range trumpFallback {
		if len(chosen) >= KoenigrufenTalonSize {
			break
		}
		chosen = append(chosen, idx)
	}
	return chosen
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Koenigrufen) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == KoenigrufenCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 味方 (デクレアラー側かどうか) を意識した戦略プレイ。パートナーが未公開でも CPU は
// サーバー側で partnerIdx を知っている前提で味方判定を行う (人間には漏れない)。
func (g *Koenigrufen) cpuPlaySmart(playerIdx int, valid []int) int {
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
	winners := koenigrufenFilter(valid, func(idx int) bool {
		return koenigrufenWinRank(p.GetCard(idx), led) > koenigrufenWinRank(winCard, led)
	})
	if winnerSide == mySide {
		return g.maxByPoints(playerIdx, valid) // 味方が勝っている → 点札を渡す
	}
	if len(winners) > 0 {
		return g.minByRank(playerIdx, winners) // 勝てる最弱札で取りに行く
	}
	return g.minByPoints(playerIdx, valid)
}

// isDeclarerSide playerIdx がデクレアラー側 (デクレアラー or パートナー) か。
func (g *Koenigrufen) isDeclarerSide(playerIdx int) bool {
	if playerIdx == g.declarerIdx {
		return true
	}
	return g.partnerIdx >= 0 && playerIdx == g.partnerIdx
}

// indexInTrick currentTrick 内で playerIdx の位置を返す (-1=なし)。
func (g *Koenigrufen) indexInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// maxByRank 勝敗ランク最大の札を返す。
func (g *Koenigrufen) maxByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := koenigrufenPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := koenigrufenPlayRank(p.GetCard(idx), led); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByRank 勝敗ランク最小の札を返す。
func (g *Koenigrufen) minByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := koenigrufenPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := koenigrufenPlayRank(p.GetCard(idx), led); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByPoints カードポイント最大の札を返す。
func (g *Koenigrufen) maxByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := koenigrufenCardPoints(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := koenigrufenCardPoints(p.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByPoints カードポイント最小の札を返す。
func (g *Koenigrufen) minByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := koenigrufenCardPoints(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := koenigrufenCardPoints(p.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// koenigrufenPlayRank プレイ順比較用のランク (切り札系 = 1000+切り札値)。
func koenigrufenPlayRank(c *Card, led int) int {
	if koenigrufenIsTrumpLike(c) {
		return 1000 + koenigrufenTrumpValue(c)
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return c.GetValue()
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Koenigrufen) GetHint() *KoenigrufenHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case KoenigrufenPhaseBid:
		if g.bidPlayerIdx != human {
			return nil
		}
		if bid, ok := g.cpuSelectBid(human); ok {
			b := int(bid)
			return &KoenigrufenHint{Bid: &b, Reason: "bid_take"}
		}
		return &KoenigrufenHint{Reason: "bid_pass"}
	case KoenigrufenPhaseCall:
		if g.declarerIdx != human {
			return nil
		}
		suit := g.cpuSelectCallKing()
		return &KoenigrufenHint{CallSuit: &suit, Reason: "call_king"}
	case KoenigrufenPhaseTalon:
		if g.declarerIdx != human {
			return nil
		}
		return &KoenigrufenHint{CardIndices: g.cpuSelectDiscards(human), Reason: "discard_weak"}
	case KoenigrufenPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &KoenigrufenHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Koenigrufen) playHintReason(playerIdx, chosenIdx int) string {
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
	if koenigrufenWinRank(card, led) > koenigrufenWinRank(winCard, led) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Koenigrufen) sortAllHands() {
	for _, p := range g.players {
		koenigrufenSortHand(p)
	}
}

// koenigrufenSortHand 手札をスート→値でソートする (切り札・スキュースは末尾)。
func koenigrufenSortHand(p *KoenigrufenPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
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

// isHumanBidTurn 現在の入札手番が人間か。
func (g *Koenigrufen) isHumanBidTurn() bool {
	return isHumanTurn(g.players, g.bidPlayerIdx)
}

// appendLog 棋譜にエントリを追加する。
func (g *Koenigrufen) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLogAt(len(g.actionLog)+1, playerIdx, actionType, detail, cards)
}

// koenigrufenBidName 入札の表示名を返す。
func koenigrufenBidName(bid KoenigrufenBid) string {
	if bid == KoenigrufenBidRufer {
		return "rufer"
	}
	return "pass"
}

// koenigrufenCardStr カードのログ表示文字列 (切り札・スキュース対応)。
func koenigrufenCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	if koenigrufenIsSkus(c) {
		return "Sküs"
	}
	if koenigrufenIsTrump(c) {
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

// koenigrufenValidBid bid が有効な入札 (Rufer) か。
func koenigrufenValidBid(bid KoenigrufenBid) bool {
	return bid == KoenigrufenBidRufer
}

// koenigrufenValidBidVal bid が定義済みの入札値 (Pass 含む) か。
func koenigrufenValidBidVal(bid KoenigrufenBid) bool {
	return bid >= KoenigrufenBidPass && bid <= KoenigrufenBidRufer
}

// koenigrufenFilter 述語を満たすインデックスを抽出する。
func koenigrufenFilter(indices []int, pred func(int) bool) []int {
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
func (g *Koenigrufen) GetPhase() KoenigrufenPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Koenigrufen) SetPhase(phase KoenigrufenPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Koenigrufen) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Koenigrufen) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Koenigrufen) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Koenigrufen) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Koenigrufen) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Koenigrufen) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Koenigrufen) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Koenigrufen) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Koenigrufen) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Koenigrufen) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Koenigrufen) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Koenigrufen) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetBidPlayerIdx 入札手番インデックス取得
func (g *Koenigrufen) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx 入札手番インデックス設定 (テスト用)
func (g *Koenigrufen) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetHighestBid 現在の最高入札取得
func (g *Koenigrufen) GetHighestBid() KoenigrufenBid { return g.highestBid }

// SetHighestBid 最高入札設定 (テスト用)
func (g *Koenigrufen) SetHighestBid(b KoenigrufenBid) { g.highestBid = b }

// GetHighestBidder 最高入札者取得 (-1=なし)
func (g *Koenigrufen) GetHighestBidder() int { return g.highestBidder }

// GetDeclarerIdx デクレアラーインデックス取得 (-1=未確定)
func (g *Koenigrufen) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx デクレアラーインデックス設定 (テスト用)
func (g *Koenigrufen) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract コントラクト (確定入札) 取得
func (g *Koenigrufen) GetContract() KoenigrufenBid { return g.contract }

// SetContract コントラクト設定 (テスト用)
func (g *Koenigrufen) SetContract(b KoenigrufenBid) { g.contract = b }

// GetCalledKing 呼ばれたキングのスート取得 (-1=未呼び/単独)
func (g *Koenigrufen) GetCalledKing() int { return g.calledKing }

// SetCalledKing 呼ばれたキング設定 (テスト用)
func (g *Koenigrufen) SetCalledKing(suit int) { g.calledKing = suit }

// GetPartnerIdx 秘密のパートナーインデックス取得 (-1=単独)。サーバー内部専用。Web プレゼンターは
// partnerRevealed=false の間これを出力してはならない。
func (g *Koenigrufen) GetPartnerIdx() int { return g.partnerIdx }

// SetPartnerIdx パートナー設定 (テスト用)
func (g *Koenigrufen) SetPartnerIdx(idx int) { g.partnerIdx = idx }

// GetPartnerRevealed パートナーが公開済みか取得
func (g *Koenigrufen) GetPartnerRevealed() bool { return g.partnerRevealed }

// SetPartnerRevealed パートナー公開フラグ設定 (テスト用)
func (g *Koenigrufen) SetPartnerRevealed(v bool) { g.partnerRevealed = v }

// GetTalonCount 場札の枚数取得
func (g *Koenigrufen) GetTalonCount() int { return len(g.talon) }

// GetTalon 場札取得 (テスト用)
func (g *Koenigrufen) GetTalon() []*Card { return g.talon }

// SetTalon 場札設定 (テスト用)
func (g *Koenigrufen) SetTalon(talon []*Card) { g.talon = talon }

// GetStashOwner stash の所有側取得 (Königrufen では常に 0=デクレアラー側)
func (g *Koenigrufen) GetStashOwner() int { return g.stashOwner }

// GetPlayerScores プレイヤー別累積得点取得
func (g *Koenigrufen) GetPlayerScores() [KoenigrufenPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積得点設定 (テスト用)
func (g *Koenigrufen) SetPlayerScores(s [KoenigrufenPlayerCnt]int) { g.playerScores = s }

// GetCardPoints プレイヤー i が獲得したカードポイント合計を返す (表示用)。
func (g *Koenigrufen) GetCardPoints(i int) int { return g.playerTrickPoints(i) }

// GetOutcome 直近ディールの結果取得
func (g *Koenigrufen) GetOutcome() KoenigrufenOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Koenigrufen) GetResult() KoenigrufenResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Koenigrufen) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Koenigrufen) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Koenigrufen) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Koenigrufen) GetPlayer(i int) *KoenigrufenPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (Play) が人間か。
func (g *Koenigrufen) IsHumanTurn() bool {
	if g.phase != KoenigrufenPhasePlay {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在の入札手番が人間か。
func (g *Koenigrufen) IsHumanBidTurn() bool {
	if g.phase != KoenigrufenPhaseBid {
		return false
	}
	return g.isHumanBidTurn()
}

// IsHumanCallTurn 現在の王呼び手番が人間 (=人間デクレアラー) か。
func (g *Koenigrufen) IsHumanCallTurn() bool {
	if g.phase != KoenigrufenPhaseCall || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// IsHumanDiscardTurn 現在の場札交換手番が人間 (=人間デクレアラー) か。
func (g *Koenigrufen) IsHumanDiscardTurn() bool {
	if g.phase != KoenigrufenPhaseTalon || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Koenigrufen) GetConfig() KoenigrufenConfig { return g.config }

// SetConfig 設定変更
func (g *Koenigrufen) SetConfig(cfg KoenigrufenConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Koenigrufen) GetActionLog() []*ActionLogEntry {
	return sliceOrEmpty(g.actionLog)
}

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Koenigrufen) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != KoenigrufenPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *Koenigrufen) ComputeBreakdownPublic() KoenigrufenBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Koenigrufen) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Koenigrufen) LedSuitPublic() int { return g.ledSuit() }

// KoenigrufenScoreDeal はディール得点計算の純粋関数の公開ラッパー (テスト用)。
func KoenigrufenScoreDeal(teamPoints int, solo bool, mult int) KoenigrufenBreakdown {
	return koenigrufenScoreDeal(teamPoints, solo, mult)
}

// KoenigrufenBidMultPublic は入札倍率を返す (テスト用)。
func KoenigrufenBidMultPublic(bid KoenigrufenBid) int { return koenigrufenBidMult(bid) }

// KoenigrufenCardPointsPublic はカードのカードポイントを返す (テスト用)。
func KoenigrufenCardPointsPublic(c *Card) int { return koenigrufenCardPoints(c) }

// KoenigrufenIsTrullPublic はカードがトゥルルか返す (テスト用)。
func KoenigrufenIsTrullPublic(c *Card) bool { return koenigrufenIsTrull(c) }

// KoenigrufenIsTrumpPublic はカードが切り札か返す (テスト用)。
func KoenigrufenIsTrumpPublic(c *Card) bool { return koenigrufenIsTrump(c) }

// KoenigrufenIsSkusPublic はカードがスキュースか返す (テスト用)。
func KoenigrufenIsSkusPublic(c *Card) bool { return koenigrufenIsSkus(c) }

// KoenigrufenIsKingPublic はカードがスートのキングか返す (テスト用)。
func KoenigrufenIsKingPublic(c *Card) bool { return koenigrufenIsKing(c) }

// BuildKoenigrufenDeckPublic は 54 枚デッキを構築する (テスト用)。
func BuildKoenigrufenDeckPublic() []*Card { return buildKoenigrufenDeck() }

// --- JSON ---

// koenigrufenJSON is the JSON wire format for Koenigrufen.
type koenigrufenJSON struct {
	Deck             []*Card                    `json:"dk"`
	DeckDrawCnt      int                        `json:"dw"`
	Players          []*KoenigrufenPlayer       `json:"ps"`
	Config           KoenigrufenConfig          `json:"cf"`
	Phase            KoenigrufenPhase           `json:"ph"`
	RoundNumber      int                        `json:"rn"`
	TrickNumber      int                        `json:"tn"`
	CurrentPlayerIdx int                        `json:"ci"`
	CurrentTrick     []*TrickCard               `json:"ct"`
	LeadPlayerIdx    int                        `json:"li"`
	DealerIdx        int                        `json:"di"`
	BidPlayerIdx     int                        `json:"bi"`
	BidActedCnt      int                        `json:"ba"`
	HighestBid       KoenigrufenBid             `json:"hb"`
	HighestBidder    int                        `json:"hr"`
	Passed           [KoenigrufenPlayerCnt]bool `json:"pd"`
	DeclarerIdx      int                        `json:"dc"`
	Contract         KoenigrufenBid             `json:"co"`
	CalledKing       int                        `json:"ck"`
	PartnerIdx       int                        `json:"pn"`
	PartnerRevealed  bool                       `json:"pr"`
	Talon            []*Card                    `json:"tl"`
	Stash            []*Card                    `json:"st"`
	StashOwner       int                        `json:"so"`
	PlayerScores     [KoenigrufenPlayerCnt]int  `json:"sc"`
	LastTrickWinner  int                        `json:"lt"`
	LastTrickCards   []*Card                    `json:"lc"`
	Outcome          KoenigrufenOutcome         `json:"oc"`
	Result           KoenigrufenResult          `json:"rs"`
	Scored           bool                       `json:"sd"`
	GameEndFlag      bool                       `json:"ge"`
	WinnerPlayer     int                        `json:"wp"`
	ActionLog        []*ActionLogEntry          `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Koenigrufen) MarshalJSON() ([]byte, error) {
	return json.Marshal(koenigrufenJSON{
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
		CalledKing:       g.calledKing,
		PartnerIdx:       g.partnerIdx,
		PartnerRevealed:  g.partnerRevealed,
		Talon:            g.talon,
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

// koenigrufenMaxSliceLen caps slice sizes during deserialisation.
const koenigrufenMaxSliceLen = 5000

// 各種デシリアライズ検証エラー。
var (
	errKoenigrufenOversized      = errors.New("koenigrufen: input array exceeds maximum allowed size")
	errKoenigrufenInvalidPlayers = errors.New("koenigrufen: invalid player count")
	errKoenigrufenInvalidTrick   = errors.New("koenigrufen: invalid trick card")
	errKoenigrufenInvalidCard    = errors.New("koenigrufen: invalid card element")
	errKoenigrufenInvalidIndex   = errors.New("koenigrufen: index field out of range")
	errKoenigrufenInvalidPhase   = errors.New("koenigrufen: phase out of range")
	errKoenigrufenInvalidBid     = errors.New("koenigrufen: bid value out of range")
	errKoenigrufenInvalidOutcome = errors.New("koenigrufen: outcome/result value out of range")
)

// koenigrufenValidCard デシリアライズ時のカード妥当性を検証する (nil 拒否, 値域チェック)。
func koenigrufenValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d, v := c.GetDesign(), c.GetValue()
	switch d {
	case KoenigrufenSkusDesign:
		return v == KoenigrufenSkusValue
	case KoenigrufenTrumpDesign:
		return v >= 1 && v <= KoenigrufenMaxTrump
	default:
		return d >= 1 && d <= KoenigrufenSuitCnt && v >= 1 && v <= KoenigrufenSuitMaxValue
	}
}

// koenigrufenCheckCards スライスの各要素のカード妥当性を検証する。
func koenigrufenCheckCards(cards []*Card) error {
	for _, c := range cards {
		if !koenigrufenValidCard(c) {
			return errKoenigrufenInvalidCard
		}
	}
	return nil
}

// koenigrufenInRange v が [0, PlayerCnt) か。
func koenigrufenInRange(v int) bool { return v >= 0 && v < KoenigrufenPlayerCnt }

// koenigrufenInRangeOrUnset v が -1 (未設定) または [0, PlayerCnt) か。
func koenigrufenInRangeOrUnset(v int) bool { return v == -1 || koenigrufenInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Koenigrufen) UnmarshalJSON(data []byte) error {
	var j koenigrufenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > koenigrufenMaxSliceLen || len(j.CurrentTrick) > koenigrufenMaxSliceLen ||
		len(j.ActionLog) > koenigrufenMaxSliceLen || len(j.Talon) > koenigrufenMaxSliceLen ||
		len(j.Stash) > koenigrufenMaxSliceLen || len(j.Deck) > koenigrufenMaxSliceLen ||
		len(j.LastTrickCards) > koenigrufenMaxSliceLen {
		return errKoenigrufenOversized
	}
	if len(j.Players) != KoenigrufenPlayerCnt {
		return errKoenigrufenInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errKoenigrufenInvalidPlayers
		}
	}
	for _, c := range j.Deck {
		if !koenigrufenValidCard(c) {
			return errKoenigrufenInvalidCard
		}
	}
	if err := koenigrufenCheckCards(j.Talon); err != nil {
		return err
	}
	if err := koenigrufenCheckCards(j.Stash); err != nil {
		return err
	}
	if err := koenigrufenCheckCards(j.LastTrickCards); err != nil {
		return err
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || !koenigrufenValidCard(tc.Card) {
			return errKoenigrufenInvalidTrick
		}
		if !koenigrufenInRange(tc.PlayerIdx) {
			return errKoenigrufenInvalidTrick
		}
	}
	if !koenigrufenInRange(j.CurrentPlayerIdx) || !koenigrufenInRange(j.DealerIdx) ||
		!koenigrufenInRange(j.BidPlayerIdx) {
		return errKoenigrufenInvalidIndex
	}
	if !koenigrufenInRangeOrUnset(j.LeadPlayerIdx) || !koenigrufenInRangeOrUnset(j.DeclarerIdx) ||
		!koenigrufenInRangeOrUnset(j.HighestBidder) || !koenigrufenInRangeOrUnset(j.LastTrickWinner) ||
		!koenigrufenInRangeOrUnset(j.WinnerPlayer) || !koenigrufenInRangeOrUnset(j.PartnerIdx) {
		return errKoenigrufenInvalidIndex
	}
	if j.CalledKing != -1 && (j.CalledKing < 1 || j.CalledKing > KoenigrufenSuitCnt) {
		return errKoenigrufenInvalidIndex
	}
	if j.StashOwner < 0 || j.StashOwner > 1 {
		return errKoenigrufenInvalidIndex
	}
	if int(j.Phase) < KoenigrufenPhaseMin || int(j.Phase) > KoenigrufenPhaseMax {
		return errKoenigrufenInvalidPhase
	}
	if !koenigrufenValidBidVal(j.HighestBid) || !koenigrufenValidBidVal(j.Contract) {
		return errKoenigrufenInvalidBid
	}
	// プレイ以降はデクレアラー・コントラクトが確定していなければならない。
	if j.Phase >= KoenigrufenPhasePlay && j.Phase <= KoenigrufenPhaseRoundEnd {
		if !koenigrufenInRange(j.DeclarerIdx) || !koenigrufenValidBid(j.Contract) ||
			!koenigrufenInRange(j.LeadPlayerIdx) {
			return errKoenigrufenInvalidIndex
		}
	}
	if j.Outcome < KoenigrufenOutcomeNone || j.Outcome > KoenigrufenOutcomeLoss {
		return errKoenigrufenInvalidOutcome
	}
	if j.Result < KoenigrufenResultLose || j.Result > KoenigrufenResultWin {
		return errKoenigrufenInvalidOutcome
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
	g.calledKing = j.CalledKing
	g.partnerIdx = j.PartnerIdx
	g.partnerRevealed = j.PartnerRevealed
	g.talon = j.Talon
	if g.talon == nil {
		g.talon = make([]*Card, 0)
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
