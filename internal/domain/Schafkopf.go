//go:build !js || !wasm || extra4

// Package domain シャーフコップ (Schafkopf) のドメインモデル。
//
// Schafkopf はドイツの Schafkopf を起源とするアメリカ中西部のトリックテイキング
// ゲーム。全クイーン・全ジャック・ダイヤ全札が固定切り札となる独自のランク体系を
// 持ち、5 人中 1 人の「ピッカー」が秘密の相棒（呼びカードの保持者）と組んで 120 点
// のカードポイントを争う。
//
// 切り札 (強い順): Q♣ > Q♠ > Q♥ > Q♦ > J♣ > J♠ > J♥ > J♦ > A♦ > 10♦ > K♦ > 9♦ > 8♦ > 7♦
// フェイル札 (各スート, 強い順): A > 10 > K > 9 > 8 > 7
// カードポイント: A=11, 10=10, K=4, Q=3, J=2, 9/8/7=0 (合計 120)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// SchafkopfPlayerCnt プレイヤー数 (人間 1 + CPU 3)
//
// **バイエルンの原典は 4 人。** クローン元の米国版シープスヘッドは 5 人卓で、
// そのぶんブラインド (伏せ札) を置いていた。
const SchafkopfPlayerCnt = 4

// SchafkopfHandSize 各プレイヤーへ配る手札枚数。
//
// **32 枚を 4 人で配り切る。** ブラインドは無い ── 5 人卓の
// シープスヘッドは 30 枚配って 2 枚を伏せていたが、4 人なら割り切れる。
const SchafkopfHandSize = 8

// SchafkopfTrickCount 1 ラウンドのトリック数
const SchafkopfTrickCount = 6

// SchafkopfTotalPoints 1 ラウンドのカードポイント合計
const SchafkopfTotalPoints = 120

// SchafkopfPhase ゲームフェーズ
type SchafkopfPhase int

// Schafkopf のフェーズ定数
const (
	// SchafkopfPhasePick 契約の宣言 (引き受けるか降りるか) フェーズ
	SchafkopfPhasePick SchafkopfPhase = 0
	// SchafkopfPhaseCall 宣言者が相棒となる呼びカード (フェイル A) を指定するフェーズ
	//
	// **埋めフェーズは無い。** クローン元の米国版シープスヘッドは 5 人卓で
	// 2 枚を伏せ、ピッカーが拾って 2 枚埋める。32 枚を 4 人で配り切る
	// バイエルンの原典には伏せ札も埋め札も存在しないので、宣言の次は呼び。
	SchafkopfPhaseCall SchafkopfPhase = 1
	// SchafkopfPhasePlay トリックプレイフェーズ
	SchafkopfPhasePlay SchafkopfPhase = 2
	// SchafkopfPhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	SchafkopfPhaseTrickEnd SchafkopfPhase = 3
	// SchafkopfPhaseRoundEnd ラウンド終了フェーズ
	SchafkopfPhaseRoundEnd SchafkopfPhase = 4
	// SchafkopfPhaseGameEnd ゲーム終了フェーズ
	SchafkopfPhaseGameEnd SchafkopfPhase = 5
)

// SchafkopfHint ヒント情報
type SchafkopfHint struct {
	CardIndices []int  // 推奨カードインデックス (プレイ・埋めフェーズ)
	Suit        int    // 推奨呼びスート (呼びフェーズ, それ以外は 0)
	Pick        bool   // 推奨ピック判断 (ピックフェーズ)
	Reason      string // ヒント理由キー
}

// Schafkopf シャーフコップのゲームクラス
type Schafkopf struct {
	trumpCards       *TrumpCards
	players          []*SchafkopfPlayer
	config           SchafkopfConfig
	phase            SchafkopfPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	passCount        int // 現ピックフェーズでパスした人数
	pickerIdx        int
	// contract は採用された契約、soloSuit は Solo で選ばれた切り札スート。
	// **切り札の構成が契約で変わる**ので、盤面の一部として持つ。
	contract        SchafkopfContract
	soloSuit        int  // ピッカー (-1 = 未確定)
	partnerIdx      int  // 相棒 (-1 = 単独 or 未確定)
	calledSuit      int  // 呼びスート (0 = 未確定/単独)
	partnerRevealed bool // 呼びカードがプレイされ相棒が判明したか
	roundPickerPts  int  // 直近ラウンドのピッカー組得点
	roundMultiplier int  // 直近ラウンドの倍率 (1/2/3)
	roundPickerWon  bool // 直近ラウンドでピッカー組が勝ったか
	gameEndFlag     bool
	winnerIdx       int // ゲーム勝者 (-1 = 未確定)
	actionLogBase
}

// NewSchafkopf コンストラクタ
func NewSchafkopf(trumpCards *TrumpCards, players []*SchafkopfPlayer, config SchafkopfConfig) *Schafkopf {
	return &Schafkopf{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		pickerIdx:  -1,
		partnerIdx: -1,
		winnerIdx:  -1,
	}
}

// NewDefaultSchafkopf 標準の 5 人構成 (人間 1, CPU 4) と既定設定で生成する。
// CUI / Web / Worker 構築の単一の真実源。
func NewDefaultSchafkopf() *Schafkopf {
	cfg := DefaultSchafkopfConfig()
	players := make([]*SchafkopfPlayer, SchafkopfPlayerCnt)
	players[0] = NewSchafkopfPlayer(true, cfg.StartChips)
	for i := 1; i < SchafkopfPlayerCnt; i++ {
		players[i] = NewSchafkopfPlayer(false, cfg.StartChips)
	}
	return NewSchafkopf(NewTrumpCardsBelote(), players, cfg)
}

// Reset ゲーム初期化: チップを開始値へ戻し最初のラウンドを開始する。
func (g *Schafkopf) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	for _, p := range g.players {
		p.SetChips(g.config.StartChips)
	}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Schafkopf) NextRound() {
	if g.phase != SchafkopfPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SchafkopfPlayerCnt
	g.startRound()
}

// startRound 手札・ブラインドを配り、ピックフェーズを開始する。
func (g *Schafkopf) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.pickerIdx = -1
	g.partnerIdx = -1
	g.calledSuit = 0
	g.partnerRevealed = false
	g.passCount = 0
	// **契約はラウンドごとにやり直す。** 持ち越すと前ラウンドの Wenz が
	// 次の配りでも切り札構成を支配し、盤面が静かに壊れる。
	g.contract = SchafkopfContractRufspiel
	g.soloSuit = 0
	g.roundPickerPts = 0
	g.roundMultiplier = 0
	g.roundPickerWon = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	// ピックはディーラーの左隣から始まる。
	g.leadPlayerIdx = (g.dealerIdx + 1) % SchafkopfPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SchafkopfPhasePick
}

// deal 各プレイヤーへ 6 枚、ブラインドへ 2 枚を配る。
func (g *Schafkopf) deal() {
	for i := 0; i < SchafkopfHandSize; i++ {
		for _, p := range g.players {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	// **ブラインドは無い。** 32 / 4 = 8 で配り切るので、伏せる札が残らない。
}

// PlayerDeclare は人間が契約を宣言する (または降りる)。
//
// **契約は 3 種類。** クローン元の米国版シープスヘッドは契約が 1 つしかなく
// 「ブラインドを取るか降りるか」の 2 値で足りたが、こちらは Rufspiel
// (A 呼び) / Wenz (Unter だけが切り札) / Solo (切り札スートを選ぶ) から選ぶ。
// soloSuit は Solo のときだけ意味を持つ。
func (g *Schafkopf) PlayerDeclare(pick bool, contract SchafkopfContract, soloSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SchafkopfPhasePick {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !pick && g.passCount >= SchafkopfPlayerCnt-1 {
		return NewDomainError(ErrInvalidPlay, "最後のプレイヤーはパスできません")
	}
	if pick {
		if contract < SchafkopfContractRufspiel || contract > SchafkopfContractSolo {
			return NewDomainError(ErrInvalidPlay, "その契約は宣言できません")
		}
		// **Solo は切り札スートを要る。** 0 はスートではないので、
		// 受け取ると「どの札とも一致しない切り札」の盤面ができる。
		if contract == SchafkopfContractSolo &&
			(soloSuit < CardDesignSpade || soloSuit > CardDesignMax) {
			return NewDomainError(ErrInvalidCard, "切り札スートを指定してください")
		}
		g.contract = contract
		g.soloSuit = soloSuit
	}
	g.resolvePick(g.currentPlayerIdx, pick)
	return nil
}

// resolvePick ピック/パスを反映し、フェーズを進める。
func (g *Schafkopf) resolvePick(playerIdx int, pick bool) {
	// 4 人がパスした場合、最後のプレイヤーは強制的にピックする。
	if !pick && g.passCount >= SchafkopfPlayerCnt-1 {
		pick = true
	}
	if pick {
		if !g.players[playerIdx].GetIsHuman() {
			g.contract, g.soloSuit = g.cpuDecideContract(playerIdx)
		}
		g.becomePicker(playerIdx)
		return
	}
	g.passCount++
	g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes", playerName(g.players, playerIdx)), nil)
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SchafkopfPlayerCnt
}

// becomePicker 宣言者を確定する。
//
// **ブラインドも埋めも無い。** クローン元の米国版シープスヘッドは 5 人卓で
// 2 枚を伏せ、ピッカーがそれを拾って 2 枚埋める。バイエルンの原典は 32 枚を
// 4 人で配り切るので伏せ札が無く、拾う札も埋める札も存在しない。
// 宣言者が決まったら、そのまま呼びフェーズ (Rufspiel の A 呼び) へ進む。
func (g *Schafkopf) becomePicker(playerIdx int) {
	g.pickerIdx = playerIdx
	g.appendLog(playerIdx, "declare",
		fmt.Sprintf("%s takes the contract", playerName(g.players, playerIdx)), nil)
	g.currentPlayerIdx = playerIdx

	// **Wenz と Solo は単独プレイ。** 相棒を呼ぶのは Rufspiel だけ。
	if g.contract != SchafkopfContractRufspiel {
		g.partnerIdx = -1
		g.appendLog(playerIdx, "alone",
			fmt.Sprintf("%s plays %s alone", playerName(g.players, playerIdx),
				schafkopfContractName(g.contract)), nil)
		g.beginPlay()
		return
	}

	// 呼べるフェイル A が無ければ相棒を作れないので単独で戦う。
	if len(g.callableSuits()) == 0 {
		g.partnerIdx = -1
		g.appendLog(playerIdx, "alone",
			fmt.Sprintf("%s plays alone", playerName(g.players, playerIdx)), nil)
		g.beginPlay()
		return
	}
	g.phase = SchafkopfPhaseCall
}

// PlayerCall 人間ピッカーが呼びスートを指定する。
func (g *Schafkopf) PlayerCall(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SchafkopfPhaseCall {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != g.pickerIdx || !g.players[g.pickerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isCallableSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "そのスートは呼べません")
	}
	g.applyCall(suit)
	return nil
}

// applyCall 呼びスートを確定し、相棒を決定してプレイフェーズへ進む。
func (g *Schafkopf) applyCall(suit int) {
	g.calledSuit = suit
	g.partnerIdx = g.holderOfCalledAce(suit)
	g.appendLog(g.pickerIdx, "call",
		fmt.Sprintf("%s calls the %s Ace", playerName(g.players, g.pickerIdx), suitStr(suit)), nil)
	g.beginPlay()
}

// beginPlay プレイフェーズを開始する。リードはピックフェーズの先頭プレイヤー。
func (g *Schafkopf) beginPlay() {
	g.leadPlayerIdx = (g.dealerIdx + 1) % SchafkopfPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber = 1
	g.currentTrick = nil
	g.phase = SchafkopfPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Schafkopf) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SchafkopfPhasePlay {
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

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する。フェーズに応じて
// ピック / 埋め / 呼び / プレイ を処理する。
func (g *Schafkopf) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case SchafkopfPhasePick:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		g.resolvePick(g.currentPlayerIdx, g.cpuDecidePick(g.currentPlayerIdx))
	case SchafkopfPhaseCall:
		if g.pickerIdx < 0 || g.players[g.pickerIdx].GetIsHuman() {
			return
		}
		g.applyCall(g.cpuSelectCall(g.pickerIdx))
	case SchafkopfPhasePlay:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		player := g.players[g.currentPlayerIdx]
		idx := g.cpuSelectPlayCard(g.currentPlayerIdx)
		played := player.RemoveCard(idx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	default:
		// 解決待ちのフェーズでは何もしない。
	}
}

// playCard カードをプレイする共通処理。
func (g *Schafkopf) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// 呼びカードがプレイされたら相棒が判明する。
	if g.calledSuit != 0 && !g.partnerRevealed &&
		card.GetValue() == 1 && card.GetDesign() == g.calledSuit && !g.isTrump(card) {
		g.partnerRevealed = true
		g.appendLog(playerIdx, "partner_reveal",
			fmt.Sprintf("%s is the picker's partner", playerName(g.players, playerIdx)), nil)
	}

	if len(g.currentTrick) == SchafkopfPlayerCnt {
		g.phase = SchafkopfPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SchafkopfPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Schafkopf) ResolveTrick() {
	if g.phase != SchafkopfPhaseTrickEnd || len(g.currentTrick) != SchafkopfPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += schafkopfCardPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", playerName(g.players, winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	// Clear the resolved trick so a spurious second ResolveTrick call cannot
	// double-count its points (defensive — NextTrick also clears it).
	g.currentTrick = nil
	if g.trickNumber >= SchafkopfTrickCount {
		g.phase = SchafkopfPhaseRoundEnd
	} else {
		g.phase = SchafkopfPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Schafkopf) NextTrick() {
	if g.phase != SchafkopfPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SchafkopfPhasePlay
}

// ScoreRound ラウンドの得点を確定し、チップ精算とゲーム終了判定を行う。
func (g *Schafkopf) ScoreRound() {
	if g.phase != SchafkopfPhaseRoundEnd {
		return
	}

	pickerPts := g.pickerTeamPoints()
	defenderPts := SchafkopfTotalPoints - pickerPts
	pickerWon := pickerPts >= 61
	loserPts := defenderPts
	if !pickerWon {
		loserPts = pickerPts
	}
	mult := schafkopfMultiplier(loserPts, g.loserTookNoTrick(pickerWon))

	g.roundPickerPts = pickerPts
	g.roundMultiplier = mult
	g.roundPickerWon = pickerWon
	g.settleChips(pickerWon, mult)

	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: picker team %d pts (%s, x%d)",
			g.roundNumber, pickerPts, schafkopfOutcomeStr(pickerWon), mult), nil)

	if w := g.chipLeaderAtTarget(); w >= 0 {
		g.gameEndFlag = true
		g.winnerIdx = w
		g.phase = SchafkopfPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, w)), nil)
	}
}

// settleChips チップ精算 (ゼロサム)。
//
// ピッカー組勝利時: 各ディフェンダーが unit*mult を支払い、相棒ありなら
// ピッカー 2 : 相棒 1 の比で受け取る。単独なら全額ピッカーへ。敗北時は符号反転。
func (g *Schafkopf) settleChips(pickerWon bool, mult int) {
	unit := g.config.BaseChips * mult
	sign := 1
	if !pickerWon {
		sign = -1
	}
	// **席数から導く。** クローン元は 5 人卓なので「守備 3 人が 1 ずつ払い、
	// ピッカー 2 + 相棒 1 が受け取る」と数を直接書いていた。4 人卓でその式を
	// 使うと守備 2 人が払った 2 に対して宣言側が 3 受け取り、**チップが 1 増える**
	// (実測: 80 → 82)。守備が払った総額をそのまま宣言側で分ける形にする。
	defenders := 0
	for i := range g.players {
		if g.isPickerTeam(i) {
			continue
		}
		defenders++
		// ディフェンダー: 敗北側なら支払い、勝利側なら受け取り。
		g.players[i].AddChips(-sign * unit)
	}
	pot := defenders * unit
	if g.partnerIdx >= 0 {
		// 宣言側 2 人で山分け。奇数なら端数はピッカーへ寄せる (総量は保つ)。
		partnerShare := pot / 2
		g.players[g.partnerIdx].AddChips(sign * partnerShare)
		g.players[g.pickerIdx].AddChips(sign * (pot - partnerShare))
	} else {
		g.players[g.pickerIdx].AddChips(sign * pot)
	}
}

// --- Scoring helpers ---

// pickerTeamPoints ピッカー組 (ピッカー + 相棒 + 埋め札) の獲得カードポイント。
func (g *Schafkopf) pickerTeamPoints() int {
	pts := 0
	for i := range g.players {
		if g.isPickerTeam(i) {
			pts += schafkopfTrickPoints(g.players[i].GetTricksTaken())
		}
	}
	// **埋め札は無い。** クローン元の 5 人卓は伏せた 2 枚の点をピッカー組へ
	// 足していたが、32 枚を 4 人で配り切るこのゲームには埋め札が存在しない。
	// 120 点はすべてトリックから出る。
	return pts
}

// loserTookNoTrick 敗北側が 1 トリックも取れなかったか (シュバルツ判定用)。
func (g *Schafkopf) loserTookNoTrick(pickerWon bool) bool {
	for i := range g.players {
		onPickerTeam := g.isPickerTeam(i)
		loserSide := onPickerTeam != pickerWon // 敗北側のプレイヤー
		if loserSide && g.players[i].GetTrickCount() > 0 {
			return false
		}
	}
	return true
}

// chipLeaderAtTarget 目標チップに到達した最上位プレイヤーを返す (-1 = なし)。
func (g *Schafkopf) chipLeaderAtTarget() int {
	best, bestChips := -1, g.config.TargetChips-1
	for i, p := range g.players {
		if p.GetChips() >= g.config.TargetChips && p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// isPickerTeam playerIdx がピッカー組 (ピッカー or 相棒) か。
func (g *Schafkopf) isPickerTeam(playerIdx int) bool {
	return playerIdx == g.pickerIdx || (g.partnerIdx >= 0 && playerIdx == g.partnerIdx)
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (リードスートに従う) を検証する。
func (g *Schafkopf) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.suitID(g.currentTrick[0].Card)
	if g.suitID(card) != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーがスート ID (切り札含む) のカードを持っているか。
func (g *Schafkopf) playerHasSuit(playerIdx, suitID int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if g.suitID(p.GetCard(i)) == suitID {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Schafkopf) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.suitID(g.currentTrick[0].Card)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := schafkopfStrength(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 勝負に絡むのは「切り札」または「リードスートに従った札」のみ。
		// フェイルがリードされても切り札を出せば勝てる (切り札 > フェイル)。
		if !g.isTrump(tc.Card) && g.suitID(tc.Card) != leadSuit {
			continue
		}
		if s := g.strength(tc.Card); s > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Schafkopf) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Call helpers ---

// callableSuits ピッカーが呼べるフェイル A のスート一覧を返す。ピッカーが
// 既に持つフェイル A は呼べない。
func (g *Schafkopf) callableSuits() []int {
	if g.pickerIdx < 0 {
		return nil
	}
	var suits []int
	for _, suit := range schafkopfFailSuits() {
		if !g.pickerHoldsAce(suit) && g.holderOfCalledAce(suit) >= 0 {
			suits = append(suits, suit)
		}
	}
	return suits
}

// isCallableSuit 指定スートが呼び可能か。
func (g *Schafkopf) isCallableSuit(suit int) bool {
	for _, s := range g.callableSuits() {
		if s == suit {
			return true
		}
	}
	return false
}

// pickerHoldsAce ピッカーが指定フェイルスートの A を持っているか。
func (g *Schafkopf) pickerHoldsAce(suit int) bool {
	picker := g.players[g.pickerIdx]
	for i := 0; i < picker.GetCardsSize(); i++ {
		c := picker.GetCard(i)
		if c.GetValue() == 1 && c.GetDesign() == suit && !g.isTrump(c) {
			return true
		}
	}
	return false
}

// holderOfCalledAce 指定フェイルスートの A を手札に持つプレイヤーを返す (-1=なし)。
func (g *Schafkopf) holderOfCalledAce(suit int) int {
	for i, p := range g.players {
		if i == g.pickerIdx {
			continue
		}
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetValue() == 1 && c.GetDesign() == suit && !g.isTrump(c) {
				return i
			}
		}
	}
	return -1
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Schafkopf) sortAllHands() {
	for _, p := range g.players {
		g.sortHand(p)
	}
}

// sortHand 手札を切り札を先頭に、スートごと強い順にソートする。
//
// **契約で切り札が変わる**ので、並び順も契約を知っている必要がある
// (Wenz ではダイヤが平札に落ちて位置が変わる)。
func (g *Schafkopf) sortHand(p *SchafkopfPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si, sj := g.suitID(ci), g.suitID(cj)
		if si != sj {
			return si < sj
		}
		return g.strength(ci) > g.strength(cj)
	})
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Schafkopf) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopStrength 現在のトリック勝者 winnerIdx の札の強さを返す。防御的に、
// 勝者がトリック内に見つからない場合は極小値を返す (パニック回避)。
func (g *Schafkopf) trickTopStrength(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return schafkopfStrength(g.currentTrick[idx].Card)
}

// --- Card classification ---

// schafkopfIsTrump 切り札 (全 Q, 全 J, ダイヤ全札) か。
// schafkopfIsTrump は Rufspiel / Solo の既定 (ダイヤ + Ober + Unter) を返す。
//
// **契約によって切り札が変わる**ので、盤面の判定には g.isTrump を使うこと。
// この関数は契約を持たない文脈 (札の分類テストなど) のための既定値。
func schafkopfIsTrump(card *Card) bool {
	return card.GetDesign() == CardDesignDiamond || card.GetValue() == schafkopfUnter || card.GetValue() == schafkopfOber
}

// schafkopfUnter / schafkopfOber は切り札の中核をなす 2 つの札位。
const (
	// schafkopfUnter ウンター (J)。**Wenz ではこの 4 枚だけが切り札。**
	schafkopfUnter = 11
	// schafkopfOber オーバー (Q)。
	schafkopfOber = 12
)

// SchafkopfContract 契約の種類。
type SchafkopfContract int

const (
	// SchafkopfContractRufspiel 呼びゲーム。ダイヤ + Ober + Unter が切り札で、
	// 宣言者は自分が持たないフェイル A を呼んで一時的な相棒を作る。
	SchafkopfContractRufspiel SchafkopfContract = iota
	// SchafkopfContractWenz ヴェンツ。**Unter (J) の 4 枚だけが切り札**で、
	// ダイヤも Ober も平札になる。宣言者は単独で戦う。
	SchafkopfContractWenz
	// SchafkopfContractSolo ソロ。宣言者が切り札スートを選び、
	// そのスート + Ober + Unter が切り札。単独で戦う。
	SchafkopfContractSolo
)

// isTrump は**現在の契約における**切り札かを返す。
//
// クローン元の米国版シープスヘッドは契約が 1 つしかないので、切り札は
// ダイヤ + Ober + Unter の固定だった。バイエルンの原典は契約で切り替わる。
func (g *Schafkopf) isTrump(card *Card) bool {
	if card == nil {
		return false
	}
	switch g.contract {
	case SchafkopfContractWenz:
		// **Wenz は Unter だけ。** Ober もダイヤも平札に落ちる。
		return card.GetValue() == schafkopfUnter
	case SchafkopfContractSolo:
		return card.GetDesign() == g.soloSuit ||
			card.GetValue() == schafkopfUnter || card.GetValue() == schafkopfOber
	default:
		// Rufspiel: ダイヤ + Ober + Unter。
		return schafkopfIsTrump(card)
	}
}

// schafkopfContractName は契約名を返す (棋譜用)。
func schafkopfContractName(c SchafkopfContract) string {
	switch c {
	case SchafkopfContractWenz:
		return "Wenz"
	case SchafkopfContractSolo:
		return "Solo"
	default:
		return "Rufspiel"
	}
}

// GetContract は採用された契約を返す。
func (g *Schafkopf) GetContract() SchafkopfContract { return g.contract }

// GetSoloSuit は Solo で選ばれた切り札スートを返す。
func (g *Schafkopf) GetSoloSuit() int { return g.soloSuit }

// SetContractForTest は契約と Solo の切り札スートを設定する (テスト用)。
func (g *Schafkopf) SetContractForTest(c SchafkopfContract, soloSuit int) {
	g.contract = c
	g.soloSuit = soloSuit
}

// suitID は**現在の契約における**トリック上のスート ID を返す。
func (g *Schafkopf) suitID(card *Card) int {
	if g.isTrump(card) {
		return schafkopfTrumpSuit
	}
	return card.GetDesign()
}

// schafkopfTrumpSuit 切り札を表すスート ID。
const schafkopfTrumpSuit = 0

// schafkopfFailSuits フェイル (非切り札) スートの一覧。
func schafkopfFailSuits() []int {
	return []int{CardDesignClover, CardDesignSpade, CardDesignHeart}
}

// schafkopfStrength トリックでの強さ。切り札はすべてフェイル札より強い。
//
//	Q♣>Q♠>Q♥>Q♦ > J♣>J♠>J♥>J♦ > A♦>10♦>K♦>9♦>8♦>7♦ > (フェイル) A>10>K>9>8>7
//
// strength は**現在の契約における**札の強さを返す。
//
// 切り札は 100 番台に載せて平札より必ず強くする。**どの札が切り札かは
// 契約で変わる**ので、固定の判定 (ダイヤ + Ober + Unter) を使うと、
// Wenz でダイヤが平札に落ちても強さだけ切り札のまま残り、トリックの勝敗が
// 変わらない ── 切り札集合を切り替えただけでは効かない。
func (g *Schafkopf) strength(card *Card) int {
	const trumpBase = 100
	v := card.GetValue()
	if !g.isTrump(card) {
		return schafkopfFailRank(v)
	}
	if v == schafkopfOber {
		return trumpBase + 30 + schafkopfTrumpSuitOrder(card.GetDesign())
	}
	if v == schafkopfUnter {
		return trumpBase + 20 + schafkopfTrumpSuitOrder(card.GetDesign())
	}
	// 切り札スートの平札 (Rufspiel のダイヤ / Solo の選んだスート)。
	return trumpBase + schafkopfFailRank(v)
}

// schafkopfStrength は Rufspiel の既定の強さを返す (契約を持たない文脈用)。
func schafkopfStrength(card *Card) int {
	const trumpBase = 100
	v := card.GetValue()
	if v == schafkopfOber {
		return trumpBase + 30 + schafkopfTrumpSuitOrder(card.GetDesign())
	}
	if v == schafkopfUnter {
		return trumpBase + 20 + schafkopfTrumpSuitOrder(card.GetDesign())
	}
	if card.GetDesign() == CardDesignDiamond {
		return trumpBase + schafkopfFailRank(v)
	}
	return schafkopfFailRank(v)
}

// schafkopfTrumpSuitOrder Q/J の切り札内スート順位 (♣>♠>♥>♦)。
func schafkopfTrumpSuitOrder(design int) int {
	switch design {
	case CardDesignClover:
		return 3
	case CardDesignSpade:
		return 2
	case CardDesignHeart:
		return 1
	default: // Diamond
		return 0
	}
}

// schafkopfFailRank フェイル/ダイヤ非絵札の強さ順位。A>10>K>9>8>7。
func schafkopfFailRank(value int) int {
	switch value {
	case 1: // Ace
		return 5
	case 10:
		return 4
	case 13: // King
		return 3
	case 9:
		return 2
	case 8:
		return 1
	default: // 7
		return 0
	}
}

// schafkopfCardPoints カードポイント。A=11,10=10,K=4,Q=3,J=2,その他=0。
func schafkopfCardPoints(value int) int {
	switch value {
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

// schafkopfTrickPoints 取得トリック群の合計カードポイント。
func schafkopfTrickPoints(tricks [][]*Card) int {
	pts := 0
	for _, t := range tricks {
		for _, c := range t {
			pts += schafkopfCardPoints(c.GetValue())
		}
	}
	return pts
}

// schafkopfMultiplier チップ倍率。敗北側 0 トリック(シュバルツ)=3、
// 30 点以下(シュナイダー)=2、それ以外=1。
func schafkopfMultiplier(loserPoints int, loserNoTrick bool) int {
	if loserNoTrick {
		return 3
	}
	if loserPoints <= 30 {
		return 2
	}
	return 1
}

// schafkopfOutcomeStr 勝敗の表示文字列。
func schafkopfOutcomeStr(pickerWon bool) string {
	if pickerWon {
		return "picker team wins"
	}
	return "defenders win"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Schafkopf) GetPhase() SchafkopfPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Schafkopf) SetPhase(phase SchafkopfPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Schafkopf) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Schafkopf) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Schafkopf) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Schafkopf) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Schafkopf) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Schafkopf) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Schafkopf) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Schafkopf) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Schafkopf) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Schafkopf) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Schafkopf) GetDealerIdx() int { return g.dealerIdx }

// GetPickerIdx ピッカーのインデックス取得 (-1=未確定)
func (g *Schafkopf) GetPickerIdx() int { return g.pickerIdx }

// SetPickerIdx ピッカー設定 (テスト用)
func (g *Schafkopf) SetPickerIdx(idx int) { g.pickerIdx = idx }

// GetPartnerIdx 相棒のインデックス取得 (-1=単独/未確定)
func (g *Schafkopf) GetPartnerIdx() int { return g.partnerIdx }

// SetPartnerIdx 相棒設定 (テスト用)
func (g *Schafkopf) SetPartnerIdx(idx int) { g.partnerIdx = idx }

// GetCalledSuit 呼びスート取得 (0=未確定/単独)
func (g *Schafkopf) GetCalledSuit() int { return g.calledSuit }

// IsPartnerRevealed 相棒が判明済みか
func (g *Schafkopf) IsPartnerRevealed() bool { return g.partnerRevealed }

// GetPassCount 現ピックフェーズのパス人数取得
func (g *Schafkopf) GetPassCount() int { return g.passCount }

// GetRoundPickerPoints 直近ラウンドのピッカー組得点取得
func (g *Schafkopf) GetRoundPickerPoints() int { return g.roundPickerPts }

// GetRoundMultiplier 直近ラウンドの倍率取得
func (g *Schafkopf) GetRoundMultiplier() int { return g.roundMultiplier }

// GetRoundPickerWon 直近ラウンドでピッカー組が勝ったか取得
func (g *Schafkopf) GetRoundPickerWon() bool { return g.roundPickerWon }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Schafkopf) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定)
func (g *Schafkopf) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Schafkopf) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Schafkopf) GetPlayer(i int) *SchafkopfPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか。
func (g *Schafkopf) IsHumanTurn() bool {
	if g.phase == SchafkopfPhaseCall {
		return g.pickerIdx >= 0 && g.players[g.pickerIdx].GetIsHuman()
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Schafkopf) GetConfig() SchafkopfConfig { return g.config }

// SetConfig 設定変更
func (g *Schafkopf) SetConfig(cfg SchafkopfConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Schafkopf) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if g.phase != SchafkopfPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// GetCallableSuits ピッカーが呼べるフェイルスート一覧を返す。
func (g *Schafkopf) GetCallableSuits() []int { return g.callableSuits() }

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Schafkopf) GetHint() *SchafkopfHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case SchafkopfPhasePick:
		if g.currentPlayerIdx != human {
			return nil
		}
		pick := g.cpuDecidePick(human)
		reason := "pick_pass"
		if pick {
			reason = "pick_take"
		}
		return &SchafkopfHint{Pick: pick, Reason: reason}
	case SchafkopfPhaseCall:
		if g.pickerIdx != human {
			return nil
		}
		return &SchafkopfHint{Suit: g.cpuSelectCall(human), Reason: "call_suit"}
	case SchafkopfPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &SchafkopfHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	default:
		return nil
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Schafkopf) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.suitID(g.currentTrick[0].Card)
	if g.suitID(card) != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	if g.strength(card) > topStrength {
		return "follow_win"
	}
	return "follow_duck"
}

// --- CPU AI ---

// cpuDecidePick CPU がピックするか判断する。切り札の枚数とクイーンで評価。
func (g *Schafkopf) cpuDecidePick(playerIdx int) bool {
	player := g.players[playerIdx]
	trumpCnt, queenCnt := 0, 0
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if g.isTrump(c) {
			trumpCnt++
		}
		if c.GetValue() == 12 {
			queenCnt++
		}
	}
	if g.config.CpuDifficulty == SchafkopfCpuDifficultyEasy {
		return trumpCnt >= 4
	}
	return trumpCnt >= 3 || (trumpCnt >= 2 && queenCnt >= 1)
}

// cpuDecideContract CPU が宣言する契約を選ぶ。
//
// **CPU も 3 契約から選ぶ。** ここを Rufspiel 固定にすると、Wenz と Solo は
// 人間が宣言したときにしか盤面に出ず、切り札構成の 2/3 が CPU 相手に
// 一度も現れない。
func (g *Schafkopf) cpuDecideContract(playerIdx int) (SchafkopfContract, int) {
	player := g.players[playerIdx]
	unterCnt := 0
	suitCnt := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetValue() == schafkopfUnter {
			unterCnt++
		}
		suitCnt[c.GetDesign()]++
	}
	// Unter だけが切り札の Wenz は、Unter を握っているほど強い。
	if unterCnt >= 3 {
		return SchafkopfContractWenz, 0
	}
	// 1 スートに偏っていれば、そのスートを切り札にした Solo が強い。
	for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
		if suitCnt[suit] >= 5 {
			return SchafkopfContractSolo, suit
		}
	}
	return SchafkopfContractRufspiel, 0
}

// cpuSelectCall CPU が呼ぶスートを選ぶ。手札に当該スートのフェイル札を持つ
// スートを優先する。
func (g *Schafkopf) cpuSelectCall(playerIdx int) int {
	callable := g.callableSuits()
	if len(callable) == 0 {
		return 0
	}
	player := g.players[playerIdx]
	for _, suit := range callable {
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if c.GetDesign() == suit && !g.isTrump(c) {
				return suit
			}
		}
	}
	return callable[0]
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Schafkopf) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SchafkopfCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Schafkopf) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低い札を出して温存する。
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return schafkopfCardPoints(c.GetValue())*100 + g.strength(c)
		})
	}

	leadSuit := g.suitID(g.currentTrick[0].Card)
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	partnerWinning := g.cpuSameTeam(playerIdx, winnerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += schafkopfCardPoints(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if g.suitID(player.GetCard(idx)) == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 味方が勝っていれば得点札を渡し、そうでなければ低得点札を捨てる。
		if partnerWinning {
			return pickHighest(player, valid, func(c *Card) int {
				if g.isTrump(c) {
					return -g.strength(c) // 切り札は温存
				}
				return schafkopfCardPoints(c.GetValue())*100 - g.strength(c)
			})
		}
		return pickLowest(player, valid, func(c *Card) int {
			return schafkopfCardPoints(c.GetValue())*100 + g.strength(c)
		})
	}

	winners := filterIndices(follows, func(idx int) bool {
		return schafkopfStrength(player.GetCard(idx)) > topStrength
	})

	if partnerWinning {
		nonWinners := filterIndices(follows, func(idx int) bool {
			return schafkopfStrength(player.GetCard(idx)) < topStrength
		})
		if len(nonWinners) > 0 {
			return pickHighest(player, nonWinners, func(c *Card) int {
				return schafkopfCardPoints(c.GetValue())*100 - g.strength(c)
			})
		}
		return pickLowest(player, follows, func(c *Card) int { return g.strength(c) })
	}

	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.strength(c) })
	}
	return pickLowest(player, follows, func(c *Card) int {
		return schafkopfCardPoints(c.GetValue())*100 + g.strength(c)
	})
}

// cpuSameTeam CPU 視点で 2 プレイヤーが同じ組か。相棒が未判明の場合でも、
// CPU はチーム構成を把握しているものとして扱う (簡易 AI)。
func (g *Schafkopf) cpuSameTeam(a, b int) bool {
	return g.isPickerTeam(a) == g.isPickerTeam(b)
}

// --- JSON ---

// schafkopfJSON is the JSON wire format for Schafkopf.
type schafkopfJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*SchafkopfPlayer `json:"ps"`
	Config           SchafkopfConfig    `json:"cf"`
	Phase            SchafkopfPhase     `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	CurrentPlayerIdx int                `json:"ci"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	LeadPlayerIdx    int                `json:"li"`
	DealerIdx        int                `json:"di"`
	PassCount        int                `json:"pc"`
	PickerIdx        int                `json:"pk"`
	// 契約は切り札の構成そのもの。落とすと復元後に Rufspiel に化け、
	// Wenz/Solo の盤面で切り札が総入れ替えになる。
	Contract        SchafkopfContract `json:"co"`
	SoloSuit        int               `json:"ss"`
	PartnerIdx      int               `json:"pt"`
	CalledSuit      int               `json:"cs"`
	PartnerRevealed bool              `json:"pr"`
	RoundPickerPts  int               `json:"rp"`
	RoundMultiplier int               `json:"rm"`
	RoundPickerWon  bool              `json:"rw"`
	GameEndFlag     bool              `json:"ge"`
	WinnerIdx       int               `json:"wi"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Schafkopf) MarshalJSON() ([]byte, error) {
	return json.Marshal(schafkopfJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		PassCount:        g.passCount,
		PickerIdx:        g.pickerIdx,
		Contract:         g.contract,
		SoloSuit:         g.soloSuit,
		PartnerIdx:       g.partnerIdx,
		CalledSuit:       g.calledSuit,
		PartnerRevealed:  g.partnerRevealed,
		RoundPickerPts:   g.roundPickerPts,
		RoundMultiplier:  g.roundMultiplier,
		RoundPickerWon:   g.roundPickerWon,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// schafkopfMaxSliceLen caps slice sizes during deserialisation.
const schafkopfMaxSliceLen = 1000

// errSchafkopfOversized is the single sentinel error for oversized input arrays.
var errSchafkopfOversized = errors.New("schafkopf: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Schafkopf) UnmarshalJSON(data []byte) error {
	var j schafkopfJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > schafkopfMaxSliceLen || len(j.CurrentTrick) > schafkopfMaxSliceLen ||
		len(j.ActionLog) > schafkopfMaxSliceLen {
		return errSchafkopfOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBelote()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*SchafkopfPlayer, 0)
	}
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
	g.passCount = j.PassCount
	g.pickerIdx = j.PickerIdx
	g.contract = j.Contract
	g.soloSuit = j.SoloSuit
	g.partnerIdx = j.PartnerIdx
	g.calledSuit = j.CalledSuit
	g.partnerRevealed = j.PartnerRevealed
	g.roundPickerPts = j.RoundPickerPts
	g.roundMultiplier = j.RoundMultiplier
	g.roundPickerWon = j.RoundPickerWon
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
