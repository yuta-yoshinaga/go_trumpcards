//go:build !js || !wasm || extra

// Package domain フレンチタロット (French Tarot) のドメインモデル。
//
// French Tarot は 78 枚のタロットデッキを用いる 4 人用コントラクト・トリックテイキング
// ゲーム。1 人のデクレアラー (preneur, taker) が残り 3 人の防御側 (defenders) と対戦する。
// 本実装は 4 人プレイの標準形 (人間 1 + CPU 3) を扱い、5 人用の「王呼び (called king)」
// バリアントは採用しない (デクレアラーは常に単独)。
//
// # デッキ (78 枚)
//
//   - スート札 56 枚: design = 1..4 (4 スート)、value = 1..14
//     (1-10, Valet=11, Cavalier=12, Dame=13, Roi=14)。
//   - 切り札 (atouts) 21 枚: design = FrenchTarotTrumpDesign (5)、value = 1..21。
//   - エクスキューズ (Excuse / Fool) 1 枚: design = FrenchTarotExcuseDesign (6)、value = 0。
//
// ブー (bouts / oudlers) は 切り札 1 (Petit)・切り札 21・エクスキューズ の 3 枚。
//
// # 簡略化ルール (本実装が採用する縮小版)
//
//   - 配札: 各プレイヤーに 18 枚 (72 枚) + シアン (chien, 犬) 6 枚。3 枚パケットで配る。
//     配り順の簡略化: 6 ラウンド繰り返し、各ラウンドで「4 人それぞれに 3 枚 → シアンに 1 枚」。
//     本来の「シアンは最初/最後のパケットに置かない・連続で置かない」制約は省略 (シャッフル
//     済みのため実害なし)。
//   - 入札 (1 巡・昇順): ディーラーの左隣から 1 巡。各プレイヤーはパスするか、現在の最高入札
//     より高い入札を行う。入札種: Petite < Garde < Garde Sans < Garde Contre。全員パスなら
//     再配札。最高入札者がデクレアラー (人間・CPU いずれもなり得る)。倍率: Petite=1, Garde=2,
//     Garde Sans=4, Garde Contre=6。
//   - シアン交換:
//     Petite / Garde — デクレアラーがシアンを公開して手札に加え (24 枚)、6 枚を伏せて捨てる
//     (エカルト écart)。エカルトはキング・エクスキューズを含められない。切り札は「捨てなければ
//     6 枚に満たない場合のみ」許可 (捨て札公開の nuance は省略)。エカルトはデクレアラー側の
//     得点に計上。
//     Garde Sans — シアンは非公開のままデクレアラーの得点札に加わる (交換なし)。
//     Garde Contre — シアンは防御側の得点札に加わる。
//   - トリックプレイ (切り札優先・18 トリック): リードスートに従う義務。ボイド時は切り札を出す
//     義務。切り札が場に出ていれば (またはオーバートランプ時) 可能なら上位切り札を出す義務
//     (monter)。切り札が出ていて上回れない/リードスートもボイドで切り札も無い場合は任意の札。
//     エクスキューズはいつでも出せ (フォロー免除)、トリックを取られず、出したプレイヤー自身の
//     トリック山に戻る。**簡略化**: 本来の「エクスキューズと引き換えに低点札 (1/2 点札) を勝者に
//     返す」交換処理は省略し、単に所有者が保持する。
//   - カードポイント (ハーフポイント): Roi / 各ブー (Petit・21・Excuse) = 4.5、Dame = 3.5、
//     Cavalier = 2.5、Valet = 1.5、その他 = 0.5。合計 91 点。内部では丸め誤差を避けるため
//     2 倍した整数 (ハーフポイント) で保持する (合計 182)。
//   - ブー数による目標点: 0 → 56、1 → 51、2 → 41、3 → 36。
//   - 得点: diff = デクレアラー点 − 目標点。base = (25 + |diff|) × 倍率。デクレアラーは 3 人の
//     防御側それぞれと base をやり取り (成功時 +3×base / 失敗時 −3×base、各防御側は逆符号)。
//     ハーフポイント空間で勝敗を厳密判定し、|diff| は 1 点未満を切り上げて整数点にする
//     (デクレアラー有利方向)。**プティ・オ・ブー (petit au bout)**: 最終トリックで Petit を
//     獲得した側に +10×倍率 (実装済み)。**簡略化**: プワニェ (poignée) とシュレム (chelem) の
//     宣言ボーナスは省略。
//   - 累積得点: TargetDeals ディール後、累積得点最上位が勝者。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// FrenchTarotPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const FrenchTarotPlayerCnt = 4

// FrenchTarotHandSize 各プレイヤーの配り札枚数
const FrenchTarotHandSize = 18

// FrenchTarotChienSize シアン (犬) の枚数
const FrenchTarotChienSize = 6

// FrenchTarotDeckSize デッキ枚数 (78 枚タロットデッキ)
const FrenchTarotDeckSize = 78

// FrenchTarotTrickCount 1 ディールのトリック数
const FrenchTarotTrickCount = 18

// FrenchTarotDefaultDeals マッチを構成するディール数 (既定)
const FrenchTarotDefaultDeals = 4

// FrenchTarotSuitCnt スート数
const FrenchTarotSuitCnt = 4

// FrenchTarotTrumpDesign 切り札 (atout) の仮想デザイン値。1..4 はスート、5 が切り札。
const FrenchTarotTrumpDesign = 5

// FrenchTarotExcuseDesign エクスキューズ (Excuse / Fool) の仮想デザイン値。
const FrenchTarotExcuseDesign = 6

// FrenchTarotExcuseValue エクスキューズのカード値。
const FrenchTarotExcuseValue = 0

// FrenchTarotMaxTrump 切り札の最大値 (21)。
const FrenchTarotMaxTrump = 21

// FrenchTarotPetitValue プティ (最小の切り札, ブー) の値。
const FrenchTarotPetitValue = 1

// FrenchTarotKingValue スート札のキング (Roi) の値。
const FrenchTarotKingValue = 14

// FrenchTarotPetitAuBoutBonus プティ・オ・ブーのボーナス (倍率適用前)。
const FrenchTarotPetitAuBoutBonus = 10

// FrenchTarotBaseGamePoints 得点計算の基礎点 (25)。
const FrenchTarotBaseGamePoints = 25

// FrenchTarotBid 入札 (コントラクト) 種別
type FrenchTarotBid int

// French Tarot の入札定数 (値が大きいほど高い入札)
const (
	// FrenchTarotBidPass パス / 未入札
	FrenchTarotBidPass FrenchTarotBid = 0
	// FrenchTarotBidPetite プティット (シアン公開・交換あり)
	FrenchTarotBidPetite FrenchTarotBid = 1
	// FrenchTarotBidGarde ガルド (シアン公開・交換あり、倍率 2)
	FrenchTarotBidGarde FrenchTarotBid = 2
	// FrenchTarotBidGardeSans ガルド・サン (シアン非公開でデクレアラー側、倍率 4)
	FrenchTarotBidGardeSans FrenchTarotBid = 3
	// FrenchTarotBidGardeContre ガルド・コントル (シアンは防御側、倍率 6)
	FrenchTarotBidGardeContre FrenchTarotBid = 4
)

// FrenchTarotPhase ゲームフェーズ
type FrenchTarotPhase int

// French Tarot のフェーズ定数
const (
	// FrenchTarotPhaseBid 入札フェーズ
	FrenchTarotPhaseBid FrenchTarotPhase = 0
	// FrenchTarotPhaseChien シアン交換 (エカルト) フェーズ
	FrenchTarotPhaseChien FrenchTarotPhase = 1
	// FrenchTarotPhasePlay トリックプレイフェーズ
	FrenchTarotPhasePlay FrenchTarotPhase = 2
	// FrenchTarotPhaseTrickEnd トリック終了フェーズ
	FrenchTarotPhaseTrickEnd FrenchTarotPhase = 3
	// FrenchTarotPhaseRoundEnd ディール終了フェーズ
	FrenchTarotPhaseRoundEnd FrenchTarotPhase = 4
	// FrenchTarotPhaseGameEnd ゲーム終了フェーズ
	FrenchTarotPhaseGameEnd FrenchTarotPhase = 5
)

// FrenchTarotPhaseMin フェーズ下限 (検証用)
const FrenchTarotPhaseMin = int(FrenchTarotPhaseBid)

// FrenchTarotPhaseMax フェーズ上限 (検証用)
const FrenchTarotPhaseMax = int(FrenchTarotPhaseGameEnd)

// FrenchTarotOutcome ディール結果 (デクレアラー視点)
type FrenchTarotOutcome int

// French Tarot のディール結果定数
const (
	// FrenchTarotOutcomeNone 未確定
	FrenchTarotOutcomeNone FrenchTarotOutcome = 0
	// FrenchTarotOutcomeWin デクレアラーがコントラクトを達成
	FrenchTarotOutcomeWin FrenchTarotOutcome = 1
	// FrenchTarotOutcomeLoss デクレアラーがコントラクトを失敗
	FrenchTarotOutcomeLoss FrenchTarotOutcome = 2
)

// FrenchTarotResult 人間視点のマッチ結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type FrenchTarotResult int

// French Tarot のマッチ結果定数
const (
	// FrenchTarotResultLose 敗北
	FrenchTarotResultLose FrenchTarotResult = -1
	// FrenchTarotResultNone 未確定 / 引き分け
	FrenchTarotResultNone FrenchTarotResult = 0
	// FrenchTarotResultWin 勝利
	FrenchTarotResultWin FrenchTarotResult = 1
)

// FrenchTarotHint ヒント情報
type FrenchTarotHint struct {
	Bid         *int   // 推奨入札 (入札フェーズ)。nil の場合はパス推奨
	CardIndices []int  // 推奨カードインデックス (エカルト/プレイ)
	Reason      string // ヒント理由キー
}

// FrenchTarotBreakdown 得点計算の内訳 (純粋関数 frenchTarotScoreDeal の出力)。
type FrenchTarotBreakdown struct {
	// DeclarerHalfPoints デクレアラーが獲得したカードのハーフポイント合計。
	DeclarerHalfPoints int
	// Bouts デクレアラーが獲得したブー数。
	Bouts int
	// Target 目標点 (整数点)。
	Target int
	// Won コントラクト成否。
	Won bool
	// DiffPoints |デクレアラー点 − 目標点| を整数点に切り上げた値。
	DiffPoints int
	// Mult 入札倍率。
	Mult int
	// Base (25 + DiffPoints) × Mult (プティ・オ・ブー除く、防御側 1 人あたりの基礎額)。
	Base int
	// PetitDelta プティ・オ・ブーによる調整 (±10×Mult、防御側 1 人あたり)。
	PetitDelta int
	// PerDefender 防御側 1 人がデクレアラーへ支払う額 (符号付き)。
	PerDefender int
	// DeclarerScore デクレアラーの得点変動 (= 3 × PerDefender)。
	DeclarerScore int
	// DefenderScore 防御側 1 人の得点変動 (= −PerDefender)。
	DefenderScore int
}

// FrenchTarot フレンチタロットのゲームクラス
type FrenchTarot struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*FrenchTarotPlayer
	config           FrenchTarotConfig
	phase            FrenchTarotPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	// --- bidding state ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    FrenchTarotBid
	highestBidder int
	passed        [FrenchTarotPlayerCnt]bool
	// --- contract state ---
	declarerIdx int
	contract    FrenchTarotBid
	// --- chien state ---
	chien         []*Card // シアン (6 枚)
	chienRevealed bool
	stash         []*Card // 得点計上用に脇へ置いた 6 枚 (エカルト or シアン)
	stashOwner    int     // 0 = デクレアラー, 1 = 防御側
	// --- scoring ---
	playerScores    [FrenchTarotPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         FrenchTarotOutcome
	result          FrenchTarotResult
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int
	actionLogBase
}

// NewFrenchTarot コンストラクタ
func NewFrenchTarot(players []*FrenchTarotPlayer, config FrenchTarotConfig) *FrenchTarot {
	return &FrenchTarot{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        FrenchTarotBidPass,
		stashOwner:      0,
	}
}

// NewDefaultFrenchTarot 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultFrenchTarot() *FrenchTarot {
	players := make([]*FrenchTarotPlayer, FrenchTarotPlayerCnt)
	players[0] = NewFrenchTarotPlayer(true)
	for i := 1; i < FrenchTarotPlayerCnt; i++ {
		players[i] = NewFrenchTarotPlayer(false)
	}
	return NewFrenchTarot(players, DefaultFrenchTarotConfig())
}

// buildFrenchTarotDeck 78 枚タロットデッキを直接構築する。スート札 (design 1..4, value
// 1..14) 56 枚 + 切り札 (design 5, value 1..21) 21 枚 + エクスキューズ (design 6, value 0)。
func buildFrenchTarotDeck() []*Card {
	deck := make([]*Card, 0, FrenchTarotDeckSize)
	for suit := 1; suit <= FrenchTarotSuitCnt; suit++ {
		for val := 1; val <= FrenchTarotKingValue; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= FrenchTarotMaxTrump; val++ {
		deck = append(deck, NewCard(FrenchTarotTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(FrenchTarotExcuseDesign, FrenchTarotExcuseValue, false))
	return deck
}

// Reset ゲーム初期化
func (g *FrenchTarot) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [FrenchTarotPlayerCnt]int{}
	g.result = FrenchTarotResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *FrenchTarot) NextRound() {
	if g.phase != FrenchTarotPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % FrenchTarotPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *FrenchTarot) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = FrenchTarotBidPass
	g.chien = nil
	g.chienRevealed = false
	g.stash = nil
	g.stashOwner = 0
	g.outcome = FrenchTarotOutcomeNone
	g.scored = false
	g.passed = [FrenchTarotPlayerCnt]bool{}
	g.highestBid = FrenchTarotBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % FrenchTarotPlayerCnt
	g.phase = FrenchTarotPhaseBid
}

// deal 3 枚パケットで各プレイヤーへ 18 枚を配り、シアン 6 枚を脇に置く。
func (g *FrenchTarot) deal() {
	g.deck = buildFrenchTarotDeck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	g.chien = make([]*Card, 0, FrenchTarotChienSize)
	packets := FrenchTarotHandSize / 3 // 6 ラウンド
	for r := 0; r < packets; r++ {
		for j := 0; j < FrenchTarotPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % FrenchTarotPlayerCnt
			for k := 0; k < 3; k++ {
				if c := g.drawCard(); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
		if c := g.drawCard(); c != nil {
			g.chien = append(g.chien, c)
		}
	}
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *FrenchTarot) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// --- Bidding ---

// PlayerBid 人間プレイヤーが入札する。
func (g *FrenchTarot) PlayerBid(bid FrenchTarotBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FrenchTarotPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	if !frenchTarotValidBid(bid) {
		return NewDomainError(ErrInvalidPlay, "無効な入札です (petite/garde/gardesans/gardecontre)")
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay, "現在の入札より高い入札が必要です")
	}
	g.applyBid(g.bidPlayerIdx, bid)
	return nil
}

// PlayerPass 人間プレイヤーがパスする。
func (g *FrenchTarot) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FrenchTarotPhaseBid {
		return ErrWrongPhase
	}
	if !g.isHumanBidTurn() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU プレイヤーが 1 回入札する (入札 or パス)。
func (g *FrenchTarot) CpuBid() {
	if g.gameEndFlag || g.phase != FrenchTarotPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= FrenchTarotPlayerCnt {
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
func (g *FrenchTarot) applyBid(idx int, bid FrenchTarotBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", playerName(g.players, idx), frenchTarotBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを適用する。
func (g *FrenchTarot) applyPass(idx int) {
	g.passed[idx] = true
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	g.advanceBid()
}

// advanceBid 入札を次のプレイヤーへ進め、1 巡終了で確定/再配札を判定する。
func (g *FrenchTarot) advanceBid() {
	g.bidActedCnt++
	if g.bidActedCnt >= FrenchTarotPlayerCnt {
		if g.highestBidder < 0 {
			g.redeal()
			return
		}
		g.finalizeBid()
		return
	}
	g.bidPlayerIdx = (g.bidPlayerIdx + 1) % FrenchTarotPlayerCnt
}

// redeal 全員パスした場合、次のディーラーで配り直す。
func (g *FrenchTarot) redeal() {
	g.appendLog(-1, "redeal", "All players passed. Redealing.", nil)
	g.dealerIdx = (g.dealerIdx + 1) % FrenchTarotPlayerCnt
	g.startRound()
}

// finalizeBid 入札を確定し、デクレアラーを決定してシアン交換 or プレイへ進む。
func (g *FrenchTarot) finalizeBid() {
	g.declarerIdx = g.highestBidder
	g.contract = g.highestBid
	g.appendLog(g.declarerIdx, "win_bid",
		fmt.Sprintf("%s takes the contract %s", playerName(g.players, g.declarerIdx), frenchTarotBidName(g.contract)), nil)
	switch g.contract {
	case FrenchTarotBidPetite, FrenchTarotBidGarde:
		// シアンを公開してデクレアラーの手札に加え、エカルトを待つ。
		g.chienRevealed = true
		for _, c := range g.chien {
			g.players[g.declarerIdx].AddCard(c)
		}
		g.chien = make([]*Card, 0)
		g.sortAllHands()
		g.currentPlayerIdx = g.declarerIdx
		g.phase = FrenchTarotPhaseChien
	case FrenchTarotBidGardeSans:
		// シアンは非公開でデクレアラー側の得点札。
		g.stash = append([]*Card(nil), g.chien...)
		g.stashOwner = 0
		g.chien = make([]*Card, 0)
		g.startPlay()
	default: // GardeContre
		// シアンは防御側の得点札。
		g.stash = append([]*Card(nil), g.chien...)
		g.stashOwner = 1
		g.chien = make([]*Card, 0)
		g.startPlay()
	}
}

// --- Chien exchange (écart) ---

// PlayerDiscard 人間デクレアラーがシアン交換で 6 枚を伏せて捨てる。
func (g *FrenchTarot) PlayerDiscard(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FrenchTarotPhaseChien {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doDiscard(cardIndices)
}

// CpuDiscard CPU デクレアラーがシアン交換で 6 枚を自動で捨てる。
func (g *FrenchTarot) CpuDiscard() {
	if g.gameEndFlag || g.phase != FrenchTarotPhaseChien {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	_ = g.doDiscard(g.cpuSelectDiscards(g.declarerIdx))
}

// doDiscard エカルトの共通処理。捨てた 6 枚をデクレアラー側の得点札 (stash) とする。
func (g *FrenchTarot) doDiscard(cardIndices []int) error {
	player := g.players[g.declarerIdx]
	if len(cardIndices) != FrenchTarotChienSize {
		return NewDomainError(ErrInvalidCard, "ちょうど 6 枚を捨ててください")
	}
	seen := make(map[int]bool, FrenchTarotChienSize)
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
		fmt.Sprintf("%s discards %d cards to the chien", playerName(g.players, g.declarerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlay()
	return nil
}

// validateDiscards エカルトの合法性を検証する。キング・エクスキューズは不可。切り札は捨て札
// 可能な非切り札・非キング・非エクスキューズ札が 6 枚に満たない場合のみ許可。
func (g *FrenchTarot) validateDiscards(player *FrenchTarotPlayer, cardIndices []int) error {
	discardable := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if frenchTarotDiscardable(player.GetCard(i)) {
			discardable++
		}
	}
	allowTrump := discardable < FrenchTarotChienSize
	// 判定は FrenchTarotUnburiableReason ただ一つ。CUI/Web が出す案内と、ここで
	// 実際に弾く条件が別々に書かれていると、片方だけ直ったときに黙ってずれる。
	for _, idx := range cardIndices {
		switch FrenchTarotUnburiableReason(player.GetCard(idx)) {
		case FrenchTarotUnburiableExcuse:
			return NewDomainError(ErrInvalidPlay, "エクスキューズは捨てられません")
		case FrenchTarotUnburiableBout:
			// プティ (切り札1) と 21 は bout であり、公式ルール上いかなる場合も écart に
			// 出せない (手札24枚中 bout は最大3枚なので、除外しても捨て札6枚は必ず確保できる)。
			return NewDomainError(ErrInvalidPlay, "プティ・21 は捨てられません")
		case FrenchTarotUnburiableKing:
			return NewDomainError(ErrInvalidPlay, "キングは捨てられません")
		case FrenchTarotUnburiableTrump:
			if !allowTrump {
				return NewDomainError(ErrInvalidPlay, "切り札は (やむを得ない場合を除き) 捨てられません")
			}
		}
	}
	return nil
}

// エカルトに出せない理由の識別子。Web (frontend/src/utils/frenchtarotEcart.ts の
// FrenchTarotUnburiableReason) と同じ語を使う。
const (
	// FrenchTarotUnburiableKing キング (Roi) は常に捨てられない。
	FrenchTarotUnburiableKing = "king"
	// FrenchTarotUnburiableExcuse エクスキューズは常に捨てられない。
	FrenchTarotUnburiableExcuse = "excuse"
	// FrenchTarotUnburiableBout プティ (切り札1) と 21 は bout なので常に捨てられない。
	FrenchTarotUnburiableBout = "bout"
	// FrenchTarotUnburiableTrump 通常の切り札。捨て札が足りないときだけ出せる。
	FrenchTarotUnburiableTrump = "trump"
)

// FrenchTarotUnburiableReason はエカルトに出せない理由を返す (出せる札なら "")。
// validateDiscards と同じ順序で判定する。
func FrenchTarotUnburiableReason(c *Card) string {
	switch {
	case c == nil:
		return ""
	case frenchTarotIsExcuse(c):
		return FrenchTarotUnburiableExcuse
	case frenchTarotIsTrump(c) && frenchTarotIsBout(c):
		return FrenchTarotUnburiableBout
	case !frenchTarotIsTrump(c) && c.GetValue() == FrenchTarotKingValue:
		return FrenchTarotUnburiableKing
	case frenchTarotIsTrump(c):
		return FrenchTarotUnburiableTrump
	default:
		return ""
	}
}

// FrenchTarotBuriableIndices はいまエカルトに出せる手札の添字を返す。
// 切り札は「自由に捨てられる札が 6 枚に足りないとき」だけ出せる (validateDiscards と同条件)。
func FrenchTarotBuriableIndices(player *FrenchTarotPlayer) []int {
	if player == nil {
		return nil
	}
	discardable := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if frenchTarotDiscardable(player.GetCard(i)) {
			discardable++
		}
	}
	allowTrump := discardable < FrenchTarotChienSize
	out := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		switch FrenchTarotUnburiableReason(player.GetCard(i)) {
		case "":
			out = append(out, i)
		case FrenchTarotUnburiableTrump:
			if allowTrump {
				out = append(out, i)
			}
		}
	}
	return out
}

// frenchTarotDiscardable 通常エカルトに出せる札か (非切り札・非キング・非エクスキューズ)。
func frenchTarotDiscardable(c *Card) bool {
	if c == nil || frenchTarotIsTrump(c) || frenchTarotIsExcuse(c) {
		return false
	}
	return c.GetValue() != FrenchTarotKingValue
}

// --- Play ---

// startPlay プレイフェーズを開始する。エルダー (ディーラーの左隣) がリードする。
func (g *FrenchTarot) startPlay() {
	g.sortAllHands()
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % FrenchTarotPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = FrenchTarotPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *FrenchTarot) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != FrenchTarotPhasePlay {
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
func (g *FrenchTarot) CpuPlay() {
	if g.gameEndFlag || g.phase != FrenchTarotPhasePlay {
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

// playCard カードをプレイする共通処理。
func (g *FrenchTarot) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), frenchTarotCardStr(card)), []*Card{card})
	if len(g.currentTrick) == FrenchTarotPlayerCnt {
		g.phase = FrenchTarotPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % FrenchTarotPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。エクスキューズは出した本人が保持し、残りを
// 勝者のトリック山へ。最終トリックなら RoundEnd に入り得点計算を発火する。
func (g *FrenchTarot) ResolveTrick() {
	if g.phase != FrenchTarotPhaseTrickEnd || len(g.currentTrick) != FrenchTarotPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	var excuseOwner = -1
	var excuseCard *Card
	won := make([]*Card, 0, FrenchTarotPlayerCnt)
	allCards := make([]*Card, 0, FrenchTarotPlayerCnt)
	for _, tc := range g.currentTrick {
		allCards = append(allCards, tc.Card)
		if frenchTarotIsExcuse(tc.Card) {
			excuseOwner = tc.PlayerIdx
			excuseCard = tc.Card
			continue
		}
		won = append(won, tc.Card)
	}
	g.players[winnerIdx].AddTrick(won)
	if excuseOwner >= 0 && excuseCard != nil {
		// エクスキューズは所有者が自分のトリック山に保持する (低点札の返却は省略)。
		g.players[excuseOwner].AddTrick([]*Card{excuseCard})
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), allCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= FrenchTarotTrickCount {
		g.lastTrickWinner = winnerIdx
		g.lastTrickCards = allCards
		g.phase = FrenchTarotPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = FrenchTarotPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *FrenchTarot) NextTrick() {
	if g.phase != FrenchTarotPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = FrenchTarotPhasePlay
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ)。
func (g *FrenchTarot) ScoreRound() {
	if g.phase != FrenchTarotPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd RoundEnd 突入時に一度だけ得点計算と精算を行う (scored フラグでガード)。
func (g *FrenchTarot) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	bd := g.computeBreakdown()
	if bd.Won {
		g.outcome = FrenchTarotOutcomeWin
	} else {
		g.outcome = FrenchTarotOutcomeLoss
	}
	for i := 0; i < FrenchTarotPlayerCnt; i++ {
		if i == g.declarerIdx {
			g.playerScores[i] += bd.DeclarerScore
		} else {
			g.playerScores[i] += bd.DefenderScore
		}
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: declarer(%s) %s target=%d pts=%d/2 bouts=%d base=%d",
			g.roundNumber, playerName(g.players, g.declarerIdx), frenchTarotBidName(g.contract),
			bd.Target, bd.DeclarerHalfPoints, bd.Bouts, bd.Base), nil)
	g.checkGameEnd()
}

// computeBreakdown 現在のディールの得点内訳を計算する。
func (g *FrenchTarot) computeBreakdown() FrenchTarotBreakdown {
	declHalf, bouts := g.declarerCaptured()
	petitSign := g.petitAuBoutSign()
	return frenchTarotScoreDeal(declHalf, bouts, petitSign, frenchTarotBidMult(g.contract))
}

// declarerCaptured デクレアラーが獲得したカードのハーフポイント合計とブー数を返す。
func (g *FrenchTarot) declarerCaptured() (int, int) {
	half, bouts := 0, 0
	if g.declarerIdx >= 0 && g.declarerIdx < len(g.players) {
		for _, trick := range g.players[g.declarerIdx].GetTricksTaken() {
			for _, c := range trick {
				half += frenchTarotCardHalfPoints(c)
				if frenchTarotIsBout(c) {
					bouts++
				}
			}
		}
	}
	if g.stashOwner == 0 {
		for _, c := range g.stash {
			half += frenchTarotCardHalfPoints(c)
			if frenchTarotIsBout(c) {
				bouts++
			}
		}
	}
	return half, bouts
}

// petitAuBoutSign 最終トリックで Petit を獲得した側を返す (+1=デクレアラー, -1=防御側, 0=なし)。
func (g *FrenchTarot) petitAuBoutSign() int {
	hasPetit := false
	for _, c := range g.lastTrickCards {
		if frenchTarotIsTrump(c) && c.GetValue() == FrenchTarotPetitValue {
			hasPetit = true
			break
		}
	}
	if !hasPetit || g.lastTrickWinner < 0 {
		return 0
	}
	if g.lastTrickWinner == g.declarerIdx {
		return 1
	}
	return -1
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積得点最上位を勝者とする。
func (g *FrenchTarot) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < FrenchTarotPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.phase = FrenchTarotPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	if tie {
		// 同点トップは引き分け: winnerPlayer を -1 にして勝者演出/メッセージを抑制する
		// (GoStop/HachiHachi と同様。GetResult も None を返す)。
		g.winnerPlayer = -1
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
	} else {
		g.winnerPlayer = leader
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
	}
}

// humanResult 人間 (seat 0) 視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None。
func (g *FrenchTarot) humanResult(leader int, tie bool) FrenchTarotResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return FrenchTarotResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return FrenchTarotResultNone
		}
		return FrenchTarotResultWin
	}
	return FrenchTarotResultLose
}

// --- Scoring helper (pure) ---

// frenchTarotTarget ブー数に対応する目標点を返す (0→56, 1→51, 2→41, 3→36)。
func frenchTarotTarget(bouts int) int {
	switch {
	case bouts <= 0:
		return 56
	case bouts == 1:
		return 51
	case bouts == 2:
		return 41
	default:
		return 36
	}
}

// frenchTarotBidMult 入札倍率を返す (Petite=1, Garde=2, GardeSans=4, GardeContre=6)。
func frenchTarotBidMult(bid FrenchTarotBid) int {
	switch bid {
	case FrenchTarotBidGarde:
		return 2
	case FrenchTarotBidGardeSans:
		return 4
	case FrenchTarotBidGardeContre:
		return 6
	default:
		return 1
	}
}

// frenchTarotScoreDeal ディール得点を計算する純粋関数。declHalf はデクレアラー獲得札の
// ハーフポイント合計、bouts はブー数、petitSign はプティ・オ・ブーの符号 (+1/-1/0)、
// mult は入札倍率。防御側精算はゼロサム (DeclarerScore = 3 × PerDefender, DefenderScore =
// −PerDefender)。
func frenchTarotScoreDeal(declHalf, bouts, petitSign, mult int) FrenchTarotBreakdown {
	target := frenchTarotTarget(bouts)
	targetHalf := target * 2
	won := declHalf >= targetHalf
	diffHalf := declHalf - targetHalf
	if diffHalf < 0 {
		diffHalf = -diffHalf
	}
	// 1 点未満を切り上げて整数点にする (デクレアラー有利方向)。
	diffPoints := (diffHalf + 1) / 2
	base := (FrenchTarotBaseGamePoints + diffPoints) * mult
	petitDelta := petitSign * FrenchTarotPetitAuBoutBonus * mult
	contractSign := 1
	if !won {
		contractSign = -1
	}
	perDefender := contractSign*base + petitDelta
	return FrenchTarotBreakdown{
		DeclarerHalfPoints: declHalf,
		Bouts:              bouts,
		Target:             target,
		Won:                won,
		DiffPoints:         diffPoints,
		Mult:               mult,
		Base:               base,
		PetitDelta:         petitDelta,
		PerDefender:        perDefender,
		DeclarerScore:      3 * perDefender,
		DefenderScore:      -perDefender,
	}
}

// --- Card classification / points ---

// frenchTarotIsTrump 切り札か。
func frenchTarotIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == FrenchTarotTrumpDesign
}

// frenchTarotIsExcuse エクスキューズか。
func frenchTarotIsExcuse(c *Card) bool {
	return c != nil && c.GetDesign() == FrenchTarotExcuseDesign
}

// frenchTarotIsBout ブー (Petit / 21 / Excuse) か。
func frenchTarotIsBout(c *Card) bool {
	if c == nil {
		return false
	}
	if frenchTarotIsExcuse(c) {
		return true
	}
	return frenchTarotIsTrump(c) && (c.GetValue() == FrenchTarotPetitValue || c.GetValue() == FrenchTarotMaxTrump)
}

// frenchTarotCardHalfPoints カードのハーフポイント (点数×2) を返す。
// Roi/各ブー = 9, Dame = 7, Cavalier = 5, Valet = 3, その他 = 1。
func frenchTarotCardHalfPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if frenchTarotIsBout(c) {
		return 9
	}
	if frenchTarotIsTrump(c) || frenchTarotIsExcuse(c) {
		return 1
	}
	switch c.GetValue() {
	case FrenchTarotKingValue: // Roi
		return 9
	case 13: // Dame
		return 7
	case 12: // Cavalier
		return 5
	case 11: // Valet
		return 3
	default:
		return 1
	}
}

// --- Trick logic ---

// ledSuit 現在のトリックのリードスートを返す。最初の非エクスキューズ札の design。
// エクスキューズのみでスートが未確定なら -1。
func (g *FrenchTarot) ledSuit() int {
	for _, tc := range g.currentTrick {
		if !frenchTarotIsExcuse(tc.Card) {
			return tc.Card.GetDesign()
		}
	}
	return -1
}

// highestTrumpInTrick 現在のトリック中の最強切り札の値を返す (0=切り札なし)。
func (g *FrenchTarot) highestTrumpInTrick() int {
	best := 0
	for _, tc := range g.currentTrick {
		if frenchTarotIsTrump(tc.Card) && tc.Card.GetValue() > best {
			best = tc.Card.GetValue()
		}
	}
	return best
}

// validatePlay マストフォロー + 切り札義務 + オーバートランプ義務を検証する。
func (g *FrenchTarot) validatePlay(playerIdx int, card *Card) error {
	return validateCardIsPlayable(g.getValidPlayIndices(playerIdx), g.players[playerIdx], card)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *FrenchTarot) getValidPlayIndices(playerIdx int) []int {
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
	if led == -1 {
		// エクスキューズがリードされスートが未確定 → 任意の札。
		return all
	}
	excuseIdx := -1
	for i := 0; i < n; i++ {
		if frenchTarotIsExcuse(player.GetCard(i)) {
			excuseIdx = i
		}
	}
	highestTrump := g.highestTrumpInTrick()
	var base []int
	if led == FrenchTarotTrumpDesign {
		base = g.trumpFollowIndices(player, highestTrump)
	} else {
		base = g.suitFollowIndices(player, led, highestTrump)
	}
	// エクスキューズはいつでも合法。
	if excuseIdx >= 0 {
		base = frenchTarotAppendUnique(base, excuseIdx)
	}
	if len(base) == 0 {
		return all
	}
	return base
}

// trumpFollowIndices 切り札がリードされた場合の合法な非エクスキューズ札を返す。
func (g *FrenchTarot) trumpFollowIndices(player *FrenchTarotPlayer, highestTrump int) []int {
	trumps := g.suitOf(player, FrenchTarotTrumpDesign)
	if len(trumps) == 0 {
		return g.nonExcuseIndices(player) // 切り札なし → 任意の非エクスキューズ札
	}
	higher := filterIndices(trumps, func(idx int) bool {
		return player.GetCard(idx).GetValue() > highestTrump
	})
	if len(higher) > 0 {
		return higher // オーバートランプ義務
	}
	return trumps
}

// suitFollowIndices スートがリードされた場合の合法な非エクスキューズ札を返す。
func (g *FrenchTarot) suitFollowIndices(player *FrenchTarotPlayer, led, highestTrump int) []int {
	ledCards := g.suitOf(player, led)
	if len(ledCards) > 0 {
		return ledCards // フォロー義務
	}
	trumps := g.suitOf(player, FrenchTarotTrumpDesign)
	if len(trumps) == 0 {
		return g.nonExcuseIndices(player) // ボイド + 切り札なし → 任意
	}
	// ボイド → 切り札義務 (+ 場に切り札があればオーバートランプ義務)。
	higher := filterIndices(trumps, func(idx int) bool {
		return player.GetCard(idx).GetValue() > highestTrump
	})
	if highestTrump > 0 && len(higher) > 0 {
		return higher
	}
	return trumps
}

// suitOf 指定 design の (非エクスキューズ) 手札インデックスを返す。
func (g *FrenchTarot) suitOf(player *FrenchTarotPlayer, design int) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if frenchTarotIsExcuse(c) {
			continue
		}
		if c.GetDesign() == design {
			out = append(out, i)
		}
	}
	return out
}

// nonExcuseIndices エクスキューズを除く全手札インデックスを返す。
func (g *FrenchTarot) nonExcuseIndices(player *FrenchTarotPlayer) []int {
	var out []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if !frenchTarotIsExcuse(player.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札。
// エクスキューズは決して勝たない。
func (g *FrenchTarot) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.ledSuit()
	winIdx := g.currentTrick[0].PlayerIdx
	winRank := -1
	for _, tc := range g.currentTrick {
		r := frenchTarotWinRank(tc.Card, led)
		if r > winRank {
			winRank = r
			winIdx = tc.PlayerIdx
		}
	}
	return winIdx
}

// frenchTarotWinRank トリック勝敗比較用のランクを返す (高いほど強い)。エクスキューズ = -1、
// 切り札 = 1000+値、リードスート = 値、それ以外 = -1。
func frenchTarotWinRank(c *Card, led int) int {
	if frenchTarotIsExcuse(c) {
		return -1
	}
	if frenchTarotIsTrump(c) {
		return 1000 + c.GetValue()
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return -1
}

// --- CPU AI ---

// cpuSelectBid CPU の入札選択 (ok=false でパス)。手札評価が閾値以上なら最高入札を 1 段上回る。
func (g *FrenchTarot) cpuSelectBid(playerIdx int) (FrenchTarotBid, bool) {
	strength := g.evalHand(playerIdx)
	// 難易度で閾値を調整。
	base := 22
	switch g.config.CpuDifficulty {
	case FrenchTarotCpuDifficultyEasy:
		base = 28
	case FrenchTarotCpuDifficultyHard:
		base = 18
	}
	// strength から希望入札上限を決める。
	var want FrenchTarotBid
	switch {
	case strength >= base+14:
		want = FrenchTarotBidGardeContre
	case strength >= base+9:
		want = FrenchTarotBidGardeSans
	case strength >= base+4:
		want = FrenchTarotBidGarde
	case strength >= base:
		want = FrenchTarotBidPetite
	default:
		return FrenchTarotBidPass, false
	}
	next := g.highestBid + 1
	if next < FrenchTarotBidPetite {
		next = FrenchTarotBidPetite
	}
	if want < next {
		return FrenchTarotBidPass, false
	}
	return next, true
}

// evalHand 手札の強さを大まかに見積もる (ブー・切り札枚数・キング・高位切り札から算出)。
func (g *FrenchTarot) evalHand(playerIdx int) int {
	p := g.players[playerIdx]
	score := 0
	trumps := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		switch {
		case frenchTarotIsBout(c):
			score += 6
		case frenchTarotIsTrump(c):
			trumps++
			if c.GetValue() >= 15 {
				score += 2
			} else {
				score++
			}
		case c.GetValue() == FrenchTarotKingValue:
			score += 3
		case c.GetValue() == 13:
			score += 2
		}
	}
	score += trumps // 切り札枚数ボーナス
	return score
}

// cpuSelectDiscards CPU デクレアラーが捨てる 6 枚のインデックスを選ぶ。価値の低い札から捨て、
// キング・エクスキューズ・ブー・切り札は温存する。
func (g *FrenchTarot) cpuSelectDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	keepValue := func(c *Card) int {
		if frenchTarotIsExcuse(c) {
			return 100000
		}
		if !frenchTarotIsTrump(c) && c.GetValue() == FrenchTarotKingValue {
			return 90000
		}
		if frenchTarotIsBout(c) {
			return 80000
		}
		if frenchTarotIsTrump(c) {
			return 10000 + c.GetValue()
		}
		return c.GetValue()*10 + frenchTarotCardHalfPoints(c)
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return keepValue(p.GetCard(idxs[a])) < keepValue(p.GetCard(idxs[b]))
	})
	// 上位 (捨てるべき) 6 枚を選ぶが、キング・エクスキューズは避ける。切り札は
	// 非切り札で埋められない場合のみ含める。
	discardable := make([]int, 0, n)
	trumpFallback := make([]int, 0, n)
	for _, idx := range idxs {
		c := p.GetCard(idx)
		if frenchTarotDiscardable(c) {
			discardable = append(discardable, idx)
		} else if frenchTarotIsTrump(c) && !frenchTarotIsBout(c) {
			trumpFallback = append(trumpFallback, idx)
		}
	}
	chosen := make([]int, 0, FrenchTarotChienSize)
	for _, idx := range discardable {
		if len(chosen) >= FrenchTarotChienSize {
			break
		}
		chosen = append(chosen, idx)
	}
	for _, idx := range trumpFallback {
		if len(chosen) >= FrenchTarotChienSize {
			break
		}
		chosen = append(chosen, idx)
	}
	return chosen
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *FrenchTarot) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == FrenchTarotCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart デクレアラー vs 防御側を意識した戦略プレイ。
func (g *FrenchTarot) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	// リード: 強い札 (デクレアラーは主導権を取りに、防御側は無難に安い札)。
	if len(g.currentTrick) == 0 {
		if playerIdx == g.declarerIdx {
			return g.maxByRank(playerIdx, valid)
		}
		return g.minByPoints(playerIdx, valid)
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexInTrick(winnerIdx)].Card
	winnerIsDecl := winnerIdx == g.declarerIdx
	iAmDecl := playerIdx == g.declarerIdx
	// 勝てる札。
	winners := filterIndices(valid, func(idx int) bool {
		return frenchTarotWinRank(p.GetCard(idx), led) > frenchTarotWinRank(winCard, led)
	})
	sameSideWinning := winnerIsDecl == iAmDecl
	if sameSideWinning {
		// 味方が勝っている → 点札を渡す (最高点)。
		return g.maxByPoints(playerIdx, valid)
	}
	// 相手が勝っている → 勝てる最弱札で取りに行く、無理なら最安札。
	if len(winners) > 0 {
		return g.minByRank(playerIdx, winners)
	}
	return g.minByPoints(playerIdx, valid)
}

// indexInTrick currentTrick 内で playerIdx の位置を返す (-1=なし)。
func (g *FrenchTarot) indexInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// maxByRank 勝敗ランク最大の札を返す。
func (g *FrenchTarot) maxByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := frenchTarotPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := frenchTarotPlayRank(p.GetCard(idx), led); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByRank 勝敗ランク最小の札を返す。
func (g *FrenchTarot) minByRank(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	best := indices[0]
	bestScore := frenchTarotPlayRank(p.GetCard(best), led)
	for _, idx := range indices[1:] {
		if s := frenchTarotPlayRank(p.GetCard(idx), led); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByPoints ハーフポイント最大の札を返す (エクスキューズは避ける)。
func (g *FrenchTarot) maxByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := frenchTarotPointKey(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := frenchTarotPointKey(p.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// minByPoints ハーフポイント最小の札を返す。
func (g *FrenchTarot) minByPoints(playerIdx int, indices []int) int {
	p := g.players[playerIdx]
	best := indices[0]
	bestScore := frenchTarotPointKey(p.GetCard(best))
	for _, idx := range indices[1:] {
		if s := frenchTarotPointKey(p.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// frenchTarotPointKey 点札温存判断用のキー。ブー/エクスキューズは高く扱う。
func frenchTarotPointKey(c *Card) int {
	if frenchTarotIsBout(c) {
		return 100 + frenchTarotCardHalfPoints(c)
	}
	return frenchTarotCardHalfPoints(c)
}

// frenchTarotPlayRank プレイ順比較用のランク (エクスキューズは最弱扱い)。
func frenchTarotPlayRank(c *Card, led int) int {
	if frenchTarotIsExcuse(c) {
		return -100
	}
	if frenchTarotIsTrump(c) {
		return 1000 + c.GetValue()
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return c.GetValue()
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *FrenchTarot) GetHint() *FrenchTarotHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case FrenchTarotPhaseBid:
		if g.bidPlayerIdx != human {
			return nil
		}
		if bid, ok := g.cpuSelectBid(human); ok {
			b := int(bid)
			return &FrenchTarotHint{Bid: &b, Reason: "bid_take"}
		}
		return &FrenchTarotHint{Reason: "bid_pass"}
	case FrenchTarotPhaseChien:
		if g.declarerIdx != human {
			return nil
		}
		return &FrenchTarotHint{CardIndices: g.cpuSelectDiscards(human), Reason: "discard_weak"}
	case FrenchTarotPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &FrenchTarotHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *FrenchTarot) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if playerIdx == g.declarerIdx {
			return "lead_high"
		}
		return "lead_low"
	}
	led := g.ledSuit()
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexInTrick(winnerIdx)].Card
	if frenchTarotWinRank(card, led) > frenchTarotWinRank(winCard, led) {
		return "follow_win"
	}
	if frenchTarotIsExcuse(card) {
		return "play_excuse"
	}
	return "follow_duck"
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *FrenchTarot) sortAllHands() {
	for _, p := range g.players {
		frenchTarotSortHand(p)
	}
}

// frenchTarotSortHand 手札をスート→値でソートする (切り札・エクスキューズは末尾)。
func frenchTarotSortHand(p *FrenchTarotPlayer) {
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
func (g *FrenchTarot) isHumanBidTurn() bool {
	return isHumanTurn(g.players, g.bidPlayerIdx)
}

// appendLog 棋譜にエントリを追加する。
func (g *FrenchTarot) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLogAt(len(g.actionLog)+1, playerIdx, actionType, detail, cards)
}

// frenchTarotBidName 入札の表示名を返す。
func frenchTarotBidName(bid FrenchTarotBid) string {
	switch bid {
	case FrenchTarotBidPetite:
		return "petite"
	case FrenchTarotBidGarde:
		return "garde"
	case FrenchTarotBidGardeSans:
		return "garde-sans"
	case FrenchTarotBidGardeContre:
		return "garde-contre"
	default:
		return "pass"
	}
}

// frenchTarotCardStr カードのログ表示文字列 (切り札・エクスキューズ対応)。
func frenchTarotCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	if frenchTarotIsExcuse(c) {
		return "Excuse"
	}
	if frenchTarotIsTrump(c) {
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

// frenchTarotValidBid bid が有効な入札 (Petite..GardeContre) か。
func frenchTarotValidBid(bid FrenchTarotBid) bool {
	return bid >= FrenchTarotBidPetite && bid <= FrenchTarotBidGardeContre
}

// frenchTarotValidBidVal bid が定義済みの入札値 (Pass 含む) か。
func frenchTarotValidBidVal(bid FrenchTarotBid) bool {
	return bid >= FrenchTarotBidPass && bid <= FrenchTarotBidGardeContre
}

// frenchTarotAppendUnique スライスに未含有のインデックスを追加する。
func frenchTarotAppendUnique(indices []int, idx int) []int {
	for _, v := range indices {
		if v == idx {
			return indices
		}
	}
	return append(indices, idx)
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *FrenchTarot) GetPhase() FrenchTarotPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *FrenchTarot) SetPhase(phase FrenchTarotPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *FrenchTarot) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *FrenchTarot) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *FrenchTarot) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *FrenchTarot) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *FrenchTarot) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *FrenchTarot) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *FrenchTarot) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *FrenchTarot) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *FrenchTarot) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *FrenchTarot) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *FrenchTarot) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *FrenchTarot) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetBidPlayerIdx 入札手番インデックス取得
func (g *FrenchTarot) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx 入札手番インデックス設定 (テスト用)
func (g *FrenchTarot) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetHighestBid 現在の最高入札取得
func (g *FrenchTarot) GetHighestBid() FrenchTarotBid { return g.highestBid }

// SetHighestBid 最高入札設定 (テスト用)
func (g *FrenchTarot) SetHighestBid(b FrenchTarotBid) { g.highestBid = b }

// GetHighestBidder 最高入札者取得 (-1=なし)
func (g *FrenchTarot) GetHighestBidder() int { return g.highestBidder }

// GetDeclarerIdx デクレアラーインデックス取得 (-1=未確定)
func (g *FrenchTarot) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx デクレアラーインデックス設定 (テスト用)
func (g *FrenchTarot) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract コントラクト (確定入札) 取得
func (g *FrenchTarot) GetContract() FrenchTarotBid { return g.contract }

// SetContract コントラクト設定 (テスト用)
func (g *FrenchTarot) SetContract(b FrenchTarotBid) { g.contract = b }

// GetChienCount シアンの枚数取得
func (g *FrenchTarot) GetChienCount() int { return len(g.chien) }

// GetChien シアン取得 (テスト用)
func (g *FrenchTarot) GetChien() []*Card { return g.chien }

// SetChien シアン設定 (テスト用)
func (g *FrenchTarot) SetChien(chien []*Card) { g.chien = chien }

// GetChienRevealed シアン公開済みか取得
func (g *FrenchTarot) GetChienRevealed() bool { return g.chienRevealed }

// GetStashOwner stash (脇に置いた 6 枚) の所有側取得 (0=デクレアラー, 1=防御側)
func (g *FrenchTarot) GetStashOwner() int { return g.stashOwner }

// GetPlayerScores プレイヤー別累積得点取得
func (g *FrenchTarot) GetPlayerScores() [FrenchTarotPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積得点設定 (テスト用)
func (g *FrenchTarot) SetPlayerScores(s [FrenchTarotPlayerCnt]int) { g.playerScores = s }

// GetCardPoints プレイヤー i が獲得したハーフポイント合計を返す (表示用)。
func (g *FrenchTarot) GetCardPoints(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	sum := 0
	for _, trick := range g.players[i].GetTricksTaken() {
		for _, c := range trick {
			sum += frenchTarotCardHalfPoints(c)
		}
	}
	return sum
}

// GetOutcome 直近ディールの結果取得
func (g *FrenchTarot) GetOutcome() FrenchTarotOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *FrenchTarot) GetResult() FrenchTarotResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *FrenchTarot) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *FrenchTarot) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *FrenchTarot) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *FrenchTarot) GetPlayer(i int) *FrenchTarotPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (Play) が人間か。
func (g *FrenchTarot) IsHumanTurn() bool {
	if g.phase != FrenchTarotPhasePlay {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在の入札手番が人間か。
func (g *FrenchTarot) IsHumanBidTurn() bool {
	if g.phase != FrenchTarotPhaseBid {
		return false
	}
	return g.isHumanBidTurn()
}

// IsHumanDiscardTurn 現在のシアン交換手番が人間 (=人間デクレアラー) か。
func (g *FrenchTarot) IsHumanDiscardTurn() bool {
	if g.phase != FrenchTarotPhaseChien || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *FrenchTarot) GetConfig() FrenchTarotConfig { return g.config }

// SetConfig 設定変更
func (g *FrenchTarot) SetConfig(cfg FrenchTarotConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *FrenchTarot) GetActionLog() []*ActionLogEntry {
	return sliceOrEmpty(g.actionLog)
}

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *FrenchTarot) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != FrenchTarotPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *FrenchTarot) ComputeBreakdownPublic() FrenchTarotBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *FrenchTarot) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *FrenchTarot) LedSuitPublic() int { return g.ledSuit() }

// FrenchTarotScoreDeal はディール得点計算の純粋関数の公開ラッパー (テスト用)。
func FrenchTarotScoreDeal(declHalf, bouts, petitSign, mult int) FrenchTarotBreakdown {
	return frenchTarotScoreDeal(declHalf, bouts, petitSign, mult)
}

// FrenchTarotTargetForBouts はブー数に対応する目標点を返す (テスト用)。
func FrenchTarotTargetForBouts(bouts int) int { return frenchTarotTarget(bouts) }

// FrenchTarotBidMultPublic は入札倍率を返す (テスト用)。
func FrenchTarotBidMultPublic(bid FrenchTarotBid) int { return frenchTarotBidMult(bid) }

// FrenchTarotCardHalfPointsPublic はカードのハーフポイントを返す (テスト用)。
func FrenchTarotCardHalfPointsPublic(c *Card) int { return frenchTarotCardHalfPoints(c) }

// FrenchTarotIsBoutPublic はカードがブーか返す (テスト用)。
func FrenchTarotIsBoutPublic(c *Card) bool { return frenchTarotIsBout(c) }

// FrenchTarotIsTrumpPublic はカードが切り札か返す (テスト用)。
func FrenchTarotIsTrumpPublic(c *Card) bool { return frenchTarotIsTrump(c) }

// FrenchTarotIsExcusePublic はカードがエクスキューズか返す (テスト用)。
func FrenchTarotIsExcusePublic(c *Card) bool { return frenchTarotIsExcuse(c) }

// BuildFrenchTarotDeckPublic は 78 枚デッキを構築する (テスト用)。
func BuildFrenchTarotDeckPublic() []*Card { return buildFrenchTarotDeck() }

// --- JSON ---

// frenchTarotJSON is the JSON wire format for FrenchTarot.
type frenchTarotJSON struct {
	Deck             []*Card                    `json:"dk"`
	DeckDrawCnt      int                        `json:"dw"`
	Players          []*FrenchTarotPlayer       `json:"ps"`
	Config           FrenchTarotConfig          `json:"cf"`
	Phase            FrenchTarotPhase           `json:"ph"`
	RoundNumber      int                        `json:"rn"`
	TrickNumber      int                        `json:"tn"`
	CurrentPlayerIdx int                        `json:"ci"`
	CurrentTrick     []*TrickCard               `json:"ct"`
	LeadPlayerIdx    int                        `json:"li"`
	DealerIdx        int                        `json:"di"`
	BidPlayerIdx     int                        `json:"bi"`
	BidActedCnt      int                        `json:"ba"`
	HighestBid       FrenchTarotBid             `json:"hb"`
	HighestBidder    int                        `json:"hr"`
	Passed           [FrenchTarotPlayerCnt]bool `json:"pd"`
	DeclarerIdx      int                        `json:"dc"`
	Contract         FrenchTarotBid             `json:"co"`
	Chien            []*Card                    `json:"ch"`
	ChienRevealed    bool                       `json:"cr"`
	Stash            []*Card                    `json:"st"`
	StashOwner       int                        `json:"so"`
	PlayerScores     [FrenchTarotPlayerCnt]int  `json:"sc"`
	LastTrickWinner  int                        `json:"lt"`
	LastTrickCards   []*Card                    `json:"lc"`
	Outcome          FrenchTarotOutcome         `json:"oc"`
	Result           FrenchTarotResult          `json:"rs"`
	Scored           bool                       `json:"sd"`
	GameEndFlag      bool                       `json:"ge"`
	WinnerPlayer     int                        `json:"wp"`
	ActionLog        []*ActionLogEntry          `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *FrenchTarot) MarshalJSON() ([]byte, error) {
	return json.Marshal(frenchTarotJSON{
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
		Chien:            g.chien,
		ChienRevealed:    g.chienRevealed,
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

// frenchTarotMaxSliceLen caps slice sizes during deserialisation.
const frenchTarotMaxSliceLen = 5000

// 各種デシリアライズ検証エラー。
var (
	errFrenchTarotOversized      = errors.New("frenchtarot: input array exceeds maximum allowed size")
	errFrenchTarotInvalidPlayers = errors.New("frenchtarot: invalid player count")
	errFrenchTarotInvalidTrick   = errors.New("frenchtarot: invalid trick card")
	errFrenchTarotInvalidCard    = errors.New("frenchtarot: invalid card element")
	errFrenchTarotInvalidIndex   = errors.New("frenchtarot: index field out of range")
	errFrenchTarotInvalidPhase   = errors.New("frenchtarot: phase out of range")
	errFrenchTarotInvalidBid     = errors.New("frenchtarot: bid value out of range")
	errFrenchTarotInvalidOutcome = errors.New("frenchtarot: outcome/result value out of range")
)

// frenchTarotValidCard デシリアライズ時のカード妥当性を検証する (nil 拒否, 値域チェック)。
func frenchTarotValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d, v := c.GetDesign(), c.GetValue()
	switch d {
	case FrenchTarotExcuseDesign:
		return v == FrenchTarotExcuseValue
	case FrenchTarotTrumpDesign:
		return v >= 1 && v <= FrenchTarotMaxTrump
	default:
		return d >= 1 && d <= FrenchTarotSuitCnt && v >= 1 && v <= FrenchTarotKingValue
	}
}

// frenchTarotCheckCards スライスの各要素のカード妥当性を検証する。
func frenchTarotCheckCards(cards []*Card) error {
	for _, c := range cards {
		if !frenchTarotValidCard(c) {
			return errFrenchTarotInvalidCard
		}
	}
	return nil
}

// frenchTarotInRange v が [0, PlayerCnt) か。
func frenchTarotInRange(v int) bool { return v >= 0 && v < FrenchTarotPlayerCnt }

// frenchTarotInRangeOrUnset v が -1 (未設定) または [0, PlayerCnt) か。
func frenchTarotInRangeOrUnset(v int) bool { return v == -1 || frenchTarotInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *FrenchTarot) UnmarshalJSON(data []byte) error {
	var j frenchTarotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > frenchTarotMaxSliceLen || len(j.CurrentTrick) > frenchTarotMaxSliceLen ||
		len(j.ActionLog) > frenchTarotMaxSliceLen || len(j.Chien) > frenchTarotMaxSliceLen ||
		len(j.Stash) > frenchTarotMaxSliceLen || len(j.Deck) > frenchTarotMaxSliceLen ||
		len(j.LastTrickCards) > frenchTarotMaxSliceLen {
		return errFrenchTarotOversized
	}
	if len(j.Players) != FrenchTarotPlayerCnt {
		return errFrenchTarotInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errFrenchTarotInvalidPlayers
		}
	}
	for _, c := range j.Deck {
		if !frenchTarotValidCard(c) {
			return errFrenchTarotInvalidCard
		}
	}
	if err := frenchTarotCheckCards(j.Chien); err != nil {
		return err
	}
	if err := frenchTarotCheckCards(j.Stash); err != nil {
		return err
	}
	if err := frenchTarotCheckCards(j.LastTrickCards); err != nil {
		return err
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || !frenchTarotValidCard(tc.Card) {
			return errFrenchTarotInvalidTrick
		}
		if !frenchTarotInRange(tc.PlayerIdx) {
			return errFrenchTarotInvalidTrick
		}
	}
	if !frenchTarotInRange(j.CurrentPlayerIdx) || !frenchTarotInRange(j.DealerIdx) ||
		!frenchTarotInRange(j.BidPlayerIdx) {
		return errFrenchTarotInvalidIndex
	}
	if !frenchTarotInRangeOrUnset(j.LeadPlayerIdx) || !frenchTarotInRangeOrUnset(j.DeclarerIdx) ||
		!frenchTarotInRangeOrUnset(j.HighestBidder) || !frenchTarotInRangeOrUnset(j.LastTrickWinner) ||
		!frenchTarotInRangeOrUnset(j.WinnerPlayer) {
		return errFrenchTarotInvalidIndex
	}
	if j.StashOwner < 0 || j.StashOwner > 1 {
		return errFrenchTarotInvalidIndex
	}
	if int(j.Phase) < FrenchTarotPhaseMin || int(j.Phase) > FrenchTarotPhaseMax {
		return errFrenchTarotInvalidPhase
	}
	if !frenchTarotValidBidVal(j.HighestBid) || !frenchTarotValidBidVal(j.Contract) {
		return errFrenchTarotInvalidBid
	}
	// プレイ以降はデクレアラー・コントラクトが確定していなければならない。
	if j.Phase >= FrenchTarotPhasePlay && j.Phase <= FrenchTarotPhaseRoundEnd {
		if !frenchTarotInRange(j.DeclarerIdx) || !frenchTarotValidBid(j.Contract) ||
			!frenchTarotInRange(j.LeadPlayerIdx) {
			return errFrenchTarotInvalidIndex
		}
	}
	if j.Outcome < FrenchTarotOutcomeNone || j.Outcome > FrenchTarotOutcomeLoss {
		return errFrenchTarotInvalidOutcome
	}
	if j.Result < FrenchTarotResultLose || j.Result > FrenchTarotResultWin {
		return errFrenchTarotInvalidOutcome
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
	g.chien = j.Chien
	if g.chien == nil {
		g.chien = make([]*Card, 0)
	}
	g.chienRevealed = j.ChienRevealed
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
