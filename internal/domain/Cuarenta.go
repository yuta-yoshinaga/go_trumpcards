//go:build !js || !wasm || extra2

// Package domain クアレンタ (Cuarenta) のドメインモデル。
//
// Cuarenta はエクアドルの国民的カードゲーム。標準 52 枚から 8,9,10 を除いた
// 40 枚デッキ (A,2,3,4,5,6,7,J,Q,K × 4 スート) を使い、4 人 2 チーム
// (席 {0,2} vs {1,3}、席 0 = 人間) で対戦する。
//
// 各プレイヤーは手札 1 枚を場に出す。場に同じ「ランク」のカードがあれば
// それらを全て捕獲する (合計ではなく純粋なランク一致)。捕獲なしなら出した
// カードはそのまま場に残る。
//
// 得点 (本実装の確定ルール):
//   - caída (カイーダ): 直前のプレイヤーが場に「置いた」(捕獲できなかった)
//     カードと同じランクを次の手番で即捕獲すると +2。直前が捕獲手だった
//     場合や、置かれてから他のプレイが挟まった場合は対象外。
//   - ronda (ロンダ): 1 手で同ランクのカードを 3 枚以上捕獲すると、
//     3 枚目以降 1 枚につき +1 (場に同ランクが複数並んでいるときに発生)。
//   - limpia (リンピア / 場の掃き): 捕獲によって場札が全て無くなると +1
//     (ラウンド最後の 1 手は掃きに数えない)。
//
// ラウンド終了 (山札・全手札を出し切る) 時に、捕獲枚数が 20 枚 (40 枚の過半)
// を超えたチームに最多取りボーナス +6。残った場札は最後に捕獲したチームの
// ものになる。各得点はチームの累計点へ加算し、いずれかのチームが
// CuarentaTargetScore (デフォルト 40) に達したら勝利。未達なら新ラウンド。
package domain

import (
	"encoding/json"
	"fmt"
)

// CuarentaPlayerCnt クアレンタのプレイヤー数 (4 人固定)。
const CuarentaPlayerCnt = 4

// CuarentaTeamCnt チーム数 (2 チーム固定)。
const CuarentaTeamCnt = 2

// CuarentaHandSize 1 回の配札で各プレイヤーに配る枚数。
const CuarentaHandSize = 5

// CuarentaInitialTableSize 各ディール開始時に場へ置く表向きカード枚数。
const CuarentaInitialTableSize = 4

// クアレンタのフェーズ (数値 enum)。
type CuarentaPhase int

// CuarentaPhase 定数。
const (
	// CuarentaPhasePlay プレイ中 (人間または CPU の手番)。
	CuarentaPhasePlay CuarentaPhase = 0
	// CuarentaPhaseRoundEnd ラウンド終了 (得点計算済み、次ラウンド待ち)。
	CuarentaPhaseRoundEnd CuarentaPhase = 1
	// CuarentaPhaseGameEnd ゲーム終了 (どちらかのチームが目標点に到達)。
	CuarentaPhaseGameEnd CuarentaPhase = 2
)

// CuarentaTeamOf 席インデックスからチーム番号を返す (i % 2)。
func CuarentaTeamOf(i int) int { return i % 2 }

// GetTeamCapturedCount はチームの捕獲枚数合計を返す。
//
// **合算は 1 箇所に置く。**精算・CUI・Web で別々に足すと、席とチームの対応を
// 変えたときに片方だけ古いままになる (#4893)。
func (g *Cuarenta) GetTeamCapturedCount(team int) int {
	total := 0
	for i, p := range g.players {
		if p != nil && CuarentaTeamOf(i) == team {
			total += p.CapturedCount()
		}
	}
	return total
}

// CuarentaAction はプレイヤー 1 ターン分の行動記録。
type CuarentaAction struct {
	PlayerIdx     int     // 行動したプレイヤーインデックス
	PlayedCard    *Card   // 出した手札 1 枚
	CapturedCards []*Card // 捕獲した場札 (捕獲なしの場合は空)
	IsCaida       bool    // caída が発生したか
	IsLimpia      bool    // limpia (場の掃き) が発生したか
	RondaBonus    int     // ronda による追加点 (0 なら無し)
}

// CuarentaRoundDetail は 1 ラウンドの得点内訳 (チーム別)。
type CuarentaRoundDetail struct {
	CapturedCount map[int]int // チーム別捕獲枚数
	Caida         map[int]int // チーム別 caída 加点合計
	Ronda         map[int]int // チーム別 ronda 加点合計
	Limpia        map[int]int // チーム別 limpia 加点合計
	MostCards     int         // 最多取りボーナスを得たチーム (-1 = なし)
	Gained        map[int]int // チーム別にこのラウンドで得た点数
}

// cuarentaRoundState はラウンドごとにリセットされる状態。
type cuarentaRoundState struct {
	phase          CuarentaPhase
	currentTurn    int
	tableCards     []*Card
	lastCaptureIdx int             // 最後に捕獲したプレイヤー (-1 = なし)
	lastLaidCard   *Card           // 直前の手番が「置いた」カード (caída 判定用、捕獲時は nil)
	humanAction    *CuarentaAction // 人間の最後の行動
	cpuActions     []*CuarentaAction
	actionLogBase
	gameEndFlag  bool
	roundWinners []int // ゲーム終了時の勝者チーム
	lastDetail   *CuarentaRoundDetail
}

// Cuarenta はクアレンタゲームの状態を保持する集約ルート。
type Cuarenta struct {
	trumpCards *TrumpCards
	players    []*CuarentaPlayer
	teamScore  [CuarentaTeamCnt]int // チーム累計点
	config     CuarentaConfig
	round      cuarentaRoundState
}

// NewCuarenta コンストラクタ。
func NewCuarenta(trumpCards *TrumpCards, players []*CuarentaPlayer, config CuarentaConfig) *Cuarenta {
	return &Cuarenta{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: cuarentaRoundState{
			phase:          CuarentaPhasePlay,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultCuarenta returns a Cuarenta with the standard 4-player setup
// (1 human, 3 CPU) and DefaultCuarentaConfig.
func NewDefaultCuarenta() *Cuarenta {
	config := DefaultCuarentaConfig()
	players := make([]*CuarentaPlayer, CuarentaPlayerCnt)
	players[0] = NewCuarentaPlayer(true)
	for i := 1; i < CuarentaPlayerCnt; i++ {
		players[i] = NewCuarentaPlayer(false)
	}
	return NewCuarenta(NewTrumpCardsScopa(), players, config)
}

// Reset は新しい「ゲーム」を開始する。累計得点もクリアする。
func (g *Cuarenta) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
	}
	g.teamScore = [CuarentaTeamCnt]int{}
	g.trumpCards = NewTrumpCardsScopa()
	g.trumpCards.Shuffle()
	g.round = cuarentaRoundState{
		phase:          CuarentaPhasePlay,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound は次のラウンドを開始する (累計得点は維持)。
func (g *Cuarenta) NextRound() {
	if g.round.gameEndFlag {
		return
	}
	for _, p := range g.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
	}
	g.trumpCards = NewTrumpCardsScopa()
	g.trumpCards.Shuffle()
	g.round.phase = CuarentaPhasePlay
	g.round.currentTurn = 0
	g.round.tableCards = nil
	g.round.lastCaptureIdx = -1
	g.round.lastLaidCard = nil
	g.round.humanAction = nil
	g.round.cpuActions = nil
	g.startRound()
}

// startRound は最初のパックを配り、場へ 4 枚を表向きで置く。
func (g *Cuarenta) startRound() {
	g.dealNextPack()
	for i := 0; i < CuarentaInitialTableSize; i++ {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.round.tableCards = append(g.round.tableCards, card)
	}
	g.appendLog(-1, "deal", fmt.Sprintf("dealt %d table cards", len(g.round.tableCards)), g.round.tableCards)
	g.round.phase = CuarentaPhasePlay
}

// dealNextPack は各プレイヤーに CuarentaHandSize 枚配る。
// 山札が尽きたら部分的に配って終わる。
func (g *Cuarenta) dealNextPack() {
	for k := 0; k < CuarentaHandSize; k++ {
		for i := 0; i < len(g.players); i++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				return
			}
			g.players[i].AddCard(card)
		}
	}
}

// allHandsEmpty は全員の手札が空か。
func (g *Cuarenta) allHandsEmpty() bool {
	return allHandsEmpty(g.players)
}

// PlayerPlay は人間プレイヤーが手札 handIdx を出す。
func (g *Cuarenta) PlayerPlay(handIdx int) error {
	if err := g.guardHumanTurn(); err != nil {
		return err
	}
	g.round.cpuActions = nil
	return g.applyPlay(g.round.currentTurn, handIdx, g.setHumanAction)
}

// CpuPlay は CPU のターンを 1 回進める。
func (g *Cuarenta) CpuPlay() {
	if g.round.gameEndFlag || g.round.phase != CuarentaPhasePlay {
		return
	}
	if g.players[g.round.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := g.round.currentTurn
	handIdx := g.chooseCpuPlay(playerIdx)
	_ = g.applyPlay(playerIdx, handIdx, g.appendCpuAction)
}

// guardHumanTurn は人間ターンかつゲーム進行中か確認。
func (g *Cuarenta) guardHumanTurn() error {
	if g.round.gameEndFlag {
		return ErrGameEnded
	}
	if g.round.phase != CuarentaPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

func (g *Cuarenta) setHumanAction(a *CuarentaAction) { g.round.humanAction = a }

func (g *Cuarenta) appendCpuAction(a *CuarentaAction) {
	g.round.cpuActions = append(g.round.cpuActions, a)
}

// applyPlay は 1 手を実行する共通処理。
// 出したカードと同ランクの場札が 1 枚以上あれば全て捕獲し、なければ場に置く。
func (g *Cuarenta) applyPlay(playerIdx, handIdx int, record func(*CuarentaAction)) error {
	player := g.players[playerIdx]
	handCard := player.GetCard(handIdx)
	if handCard == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}

	matchIdxs := cuarentaRankMatchIndexes(handCard, g.round.tableCards)
	_ = player.RemoveCard(handIdx)

	if len(matchIdxs) == 0 {
		// 捕獲なし: 場に置く。caída 判定用に「置いたカード」を記録。
		g.round.tableCards = append(g.round.tableCards, handCard)
		g.round.lastLaidCard = handCard
		action := &CuarentaAction{PlayerIdx: playerIdx, PlayedCard: handCard}
		record(action)
		g.appendLog(playerIdx, "play", "placed on table", []*Card{handCard})
		g.postActionAdvance()
		return nil
	}

	// 捕獲。
	captured := make([]*Card, 0, len(matchIdxs))
	for _, idx := range matchIdxs {
		captured = append(captured, g.round.tableCards[idx])
	}
	g.removeTableCardsByIndex(matchIdxs)

	pile := make([]*Card, 0, len(captured)+1)
	pile = append(pile, handCard)
	pile = append(pile, captured...)
	player.AddCaptured(pile)
	g.round.lastCaptureIdx = playerIdx

	team := CuarentaTeamOf(playerIdx)

	// caída: 直前の手番が「置いた」カードと同ランクを即捕獲したか。
	isCaida := g.round.lastLaidCard != nil &&
		g.round.lastLaidCard.GetValue() == handCard.GetValue()
	if isCaida {
		g.teamScore[team] += CuarentaScoreCaida
	}

	// ronda: 同ランク 3 枚以上 (出したカード + 捕獲枚数) を一度に取ると 3 枚目以降 +1/枚。
	rondaBonus := 0
	totalSameRank := len(captured) + 1
	if totalSameRank >= 3 {
		rondaBonus = (totalSameRank - 2) * CuarentaScoreRondaPerExtra
		g.teamScore[team] += rondaBonus
	}

	// limpia: 場札を全て掃いた (ラウンド最後の 1 手を除く)。
	isLimpia := len(g.round.tableCards) == 0 && !g.isLastPlayOfRound()
	if isLimpia {
		g.teamScore[team] += CuarentaScoreLimpia
	}

	// この手は捕獲だったので caída 連鎖の起点をリセット。
	g.round.lastLaidCard = nil

	action := &CuarentaAction{
		PlayerIdx:     playerIdx,
		PlayedCard:    handCard,
		CapturedCards: captured,
		IsCaida:       isCaida,
		IsLimpia:      isLimpia,
		RondaBonus:    rondaBonus,
	}
	record(action)
	g.appendLog(playerIdx, "capture", fmt.Sprintf("captured %d card(s)", len(captured)), pile)
	g.postActionAdvance()
	return nil
}

// postActionAdvance はアクション後の共通進行処理。
func (g *Cuarenta) postActionAdvance() {
	if g.isRoundOver() {
		g.finishRound()
		return
	}
	g.round.currentTurn = (g.round.currentTurn + 1) % len(g.players)
	if g.allHandsEmpty() && g.trumpCards.GetRemainingCount() > 0 {
		g.dealNextPack()
	}
}

// isRoundOver は現在のラウンドが終了しているか (手札 0 + 山札 0)。
func (g *Cuarenta) isRoundOver() bool {
	return g.allHandsEmpty() && g.trumpCards.GetRemainingCount() == 0
}

// isLastPlayOfRound は今の手がラウンド最後の 1 手か (掃きボーナス除外用)。
func (g *Cuarenta) isLastPlayOfRound() bool {
	return g.allHandsEmpty() && g.trumpCards.GetRemainingCount() == 0
}

// finishRound はラウンド終了処理: 残り場札を最後の捕獲者に渡し、得点計算。
func (g *Cuarenta) finishRound() {
	g.round.phase = CuarentaPhaseRoundEnd
	leftover := append([]*Card(nil), g.round.tableCards...)
	g.round.tableCards = nil
	if g.round.lastCaptureIdx >= 0 && len(leftover) > 0 {
		g.players[g.round.lastCaptureIdx].AddCaptured(leftover)
		g.appendLog(g.round.lastCaptureIdx, "lastTake", fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}

	detail := g.scoreRound()
	g.round.lastDetail = detail
	for t := 0; t < CuarentaTeamCnt; t++ {
		g.teamScore[t] += detail.Gained[t]
	}

	maxScore := 0
	for _, s := range g.teamScore {
		if s > maxScore {
			maxScore = s
		}
	}
	if maxScore >= g.config.TargetScore {
		g.round.gameEndFlag = true
		g.round.phase = CuarentaPhaseGameEnd
		winners := make([]int, 0)
		for t := 0; t < CuarentaTeamCnt; t++ {
			if g.teamScore[t] == maxScore {
				winners = append(winners, t)
			}
		}
		g.round.roundWinners = winners
		g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended at %d points", maxScore), nil)
	} else {
		g.appendLog(-1, "roundEnd", "round ended", nil)
	}
}

// scoreRound はラウンド終了時の最多取りボーナスのみを集計する。
// caída/ronda/limpia は applyPlay 内で即時加点済みのため、ここでは内訳の
// 表示用に再計上しつつ、Gained には最多取りボーナスだけを入れる。
func (g *Cuarenta) scoreRound() *CuarentaRoundDetail {
	det := &CuarentaRoundDetail{
		CapturedCount: make(map[int]int),
		Caida:         make(map[int]int),
		Ronda:         make(map[int]int),
		Limpia:        make(map[int]int),
		Gained:        make(map[int]int),
		MostCards:     -1,
	}
	for t := 0; t < CuarentaTeamCnt; t++ {
		det.CapturedCount[t] = g.GetTeamCapturedCount(t)
	}
	// 即時加点の内訳をログ用に再構成 (humanAction + cpuActions では網羅できないため
	// 表示は概算。Gained は最多取りボーナスのみを担当する)。
	mostTeam := -1
	for t := 0; t < CuarentaTeamCnt; t++ {
		if det.CapturedCount[t] > CuarentaMostCardsThreshold {
			mostTeam = t
		}
	}
	det.MostCards = mostTeam
	if mostTeam >= 0 {
		det.Gained[mostTeam] += CuarentaScoreMostCards
	}
	return det
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (g *Cuarenta) removeTableCardsByIndex(idxs []int) {
	g.round.tableCards = removeIndices(g.round.tableCards, idxs)
}

// appendLog 棋譜にエントリを追加する。
func (g *Cuarenta) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn 現在の手番が人間かどうか。
func (g *Cuarenta) IsHumanTurn() bool {
	if g.round.gameEndFlag {
		return false
	}
	return g.players[g.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得。
func (g *Cuarenta) GetCurrentTurn() int { return g.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得。
func (g *Cuarenta) GetGameEndFlag() bool { return g.round.gameEndFlag }

// GetTableCards 場札取得。
func (g *Cuarenta) GetTableCards() []*Card { return g.round.tableCards }

// GetPlayer プレイヤー取得。
func (g *Cuarenta) GetPlayer(idx int) *CuarentaPlayer {
	return getPlayer(g.players, idx)
}

// GetPlayerCnt プレイヤー数取得。
func (g *Cuarenta) GetPlayerCnt() int { return len(g.players) }

// GetTeamScore 指定チームの累計点取得。
func (g *Cuarenta) GetTeamScore(team int) int {
	if team < 0 || team >= CuarentaTeamCnt {
		return 0
	}
	return g.teamScore[team]
}

// GetCpuActions CPU ターンの行動履歴取得。
func (g *Cuarenta) GetCpuActions() []*CuarentaAction { return g.round.cpuActions }

// GetHumanAction 人間の最後の行動取得。
func (g *Cuarenta) GetHumanAction() *CuarentaAction { return g.round.humanAction }

// GetConfig ローカルルール設定取得。
func (g *Cuarenta) GetConfig() CuarentaConfig { return g.config }

// SetConfig ローカルルール設定を変更。
func (g *Cuarenta) SetConfig(config CuarentaConfig) { g.config = config }

// GetActionLog 棋譜取得。
func (g *Cuarenta) GetActionLog() []*ActionLogEntry { return g.round.actionLog }

// GetPhase 現在のフェーズ取得 (数値)。
func (g *Cuarenta) GetPhase() int { return int(g.round.phase) }

// GetLastRoundDetail 直前ラウンドの得点詳細取得 (nil の場合もあり得る)。
func (g *Cuarenta) GetLastRoundDetail() *CuarentaRoundDetail { return g.round.lastDetail }

// GetLastCaptureIdx 最後に捕獲したプレイヤー (-1 = なし)。
func (g *Cuarenta) GetLastCaptureIdx() int { return g.round.lastCaptureIdx }

// GetRoundWinners ゲーム終了時の勝者チームリスト。
func (g *Cuarenta) GetRoundWinners() []int { return g.round.roundWinners }

// GetRemainingDeck 山札の残り枚数。
func (g *Cuarenta) GetRemainingDeck() int { return g.trumpCards.GetRemainingCount() }

// --- JSON Serialization ---

// cuarentaActionJSON is the JSON wire format for CuarentaAction.
type cuarentaActionJSON struct {
	PlayerIdx     int     `json:"pi"`
	PlayedCard    *Card   `json:"pc"`
	CapturedCards []*Card `json:"cc"`
	IsCaida       bool    `json:"cd"`
	IsLimpia      bool    `json:"lp"`
	RondaBonus    int     `json:"rb"`
}

// MarshalJSON implements json.Marshaler.
func (a *CuarentaAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuarentaActionJSON{
		PlayerIdx:     a.PlayerIdx,
		PlayedCard:    a.PlayedCard,
		CapturedCards: a.CapturedCards,
		IsCaida:       a.IsCaida,
		IsLimpia:      a.IsLimpia,
		RondaBonus:    a.RondaBonus,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *CuarentaAction) UnmarshalJSON(data []byte) error {
	var j cuarentaActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCard = j.PlayedCard
	a.CapturedCards = j.CapturedCards
	a.IsCaida = j.IsCaida
	a.IsLimpia = j.IsLimpia
	a.RondaBonus = j.RondaBonus
	return nil
}

// cuarentaRoundDetailJSON is the JSON wire format for CuarentaRoundDetail.
type cuarentaRoundDetailJSON struct {
	CapturedCount map[int]int `json:"cc"`
	Caida         map[int]int `json:"cd"`
	Ronda         map[int]int `json:"rd"`
	Limpia        map[int]int `json:"lp"`
	MostCards     int         `json:"mc"`
	Gained        map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *CuarentaRoundDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuarentaRoundDetailJSON{
		CapturedCount: d.CapturedCount,
		Caida:         d.Caida,
		Ronda:         d.Ronda,
		Limpia:        d.Limpia,
		MostCards:     d.MostCards,
		Gained:        d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *CuarentaRoundDetail) UnmarshalJSON(data []byte) error {
	var j cuarentaRoundDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.CapturedCount = j.CapturedCount
	d.Caida = j.Caida
	d.Ronda = j.Ronda
	d.Limpia = j.Limpia
	d.MostCards = j.MostCards
	d.Gained = j.Gained
	return nil
}

// cuarentaJSON is the JSON wire format for Cuarenta.
type cuarentaJSON struct {
	TrumpCards     *TrumpCards          `json:"tc"`
	Players        []*CuarentaPlayer    `json:"pl"`
	TeamScore      []int                `json:"ts"`
	Config         CuarentaConfig       `json:"cf"`
	Phase          int                  `json:"ph"`
	CurrentTurn    int                  `json:"ct"`
	TableCards     []*Card              `json:"tb"`
	LastCaptureIdx int                  `json:"lc"`
	LastLaidCard   *Card                `json:"ll"`
	HumanAction    *CuarentaAction      `json:"ha"`
	CpuActions     []*CuarentaAction    `json:"ca"`
	ActionLog      []*ActionLogEntry    `json:"al"`
	GameEndFlag    bool                 `json:"ge"`
	RoundWinners   []int                `json:"rw"`
	LastDetail     *CuarentaRoundDetail `json:"ld"`
}

// cuarentaMaxSliceLen caps slice sizes during deserialisation.
const cuarentaMaxSliceLen = 1000

// errCuarentaBadState は不正な永続化データを拒否する共通センチネル。
var errCuarentaBadState = fmt.Errorf("cuarenta: invalid serialized state")

// MarshalJSON implements json.Marshaler.
func (g *Cuarenta) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuarentaJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		TeamScore:      g.teamScore[:],
		Config:         g.config,
		Phase:          int(g.round.phase),
		CurrentTurn:    g.round.currentTurn,
		TableCards:     g.round.tableCards,
		LastCaptureIdx: g.round.lastCaptureIdx,
		LastLaidCard:   g.round.lastLaidCard,
		HumanAction:    g.round.humanAction,
		CpuActions:     g.round.cpuActions,
		ActionLog:      g.round.actionLog,
		GameEndFlag:    g.round.gameEndFlag,
		RoundWinners:   g.round.roundWinners,
		LastDetail:     g.round.lastDetail,
	})
}

// UnmarshalJSON implements json.Unmarshaler with defensive validation.
func (g *Cuarenta) UnmarshalJSON(data []byte) error {
	var j cuarentaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.TrumpCards == nil {
		return errCuarentaBadState
	}
	if len(j.Players) != CuarentaPlayerCnt {
		return errCuarentaBadState
	}
	for _, p := range j.Players {
		if p == nil {
			return errCuarentaBadState
		}
	}
	if j.Phase < int(CuarentaPhasePlay) || j.Phase > int(CuarentaPhaseGameEnd) {
		return errCuarentaBadState
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= CuarentaPlayerCnt {
		return errCuarentaBadState
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= CuarentaPlayerCnt {
		return errCuarentaBadState
	}
	if len(j.TableCards) > cuarentaMaxSliceLen ||
		len(j.CpuActions) > cuarentaMaxSliceLen ||
		len(j.ActionLog) > cuarentaMaxSliceLen {
		return errCuarentaBadState
	}

	var teamScore [CuarentaTeamCnt]int
	for t := 0; t < CuarentaTeamCnt && t < len(j.TeamScore); t++ {
		if j.TeamScore[t] < 0 || j.TeamScore[t] > 100000 {
			return errCuarentaBadState
		}
		teamScore[t] = j.TeamScore[t]
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.teamScore = teamScore
	g.config = j.Config
	g.round = cuarentaRoundState{
		phase:          CuarentaPhase(j.Phase),
		currentTurn:    j.CurrentTurn,
		tableCards:     j.TableCards,
		lastCaptureIdx: j.LastCaptureIdx,
		lastLaidCard:   j.LastLaidCard,
		humanAction:    j.HumanAction,
		cpuActions:     j.CpuActions,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
		gameEndFlag:    j.GameEndFlag,
		roundWinners:   j.RoundWinners,
		lastDetail:     j.LastDetail,
	}
	if g.round.actionLog == nil {
		g.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
