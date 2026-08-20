//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// ZhengPlayerCnt 争上游プレイヤー数
const ZhengPlayerCnt = 4

// ZhengJokerCount 争上游で使用するジョーカー枚数 (小+大)
const ZhengJokerCount = 2

// ZhengMaxCardsPerPlayer 1人あたりの最大配布枚数 (54枚を4人に配ると先頭2人が14枚)
const ZhengMaxCardsPerPlayer = 14

// ZhengAction プレイヤーの1ターン分の行動記録
type ZhengAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// zhengRoundState ラウンドごとにリセットされる状態
type zhengRoundState struct {
	currentTurn       int            // 現在の手番プレイヤーインデックス
	tableCards        []*Card        // 場に出されているカード (nil = 場はクリア)
	tablePlayType     ZhengPlayType  // 場のプレイタイプ
	lastPlayPlayerIdx int            // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag       bool           // ゲーム終了フラグ
	passCount         int            // 最後の出し以降の連続パス数
	cpuActions        []*ZhengAction // 人間ターン後のCPUの行動履歴
	humanAction       *ZhengAction   // 人間の最後の行動
	actionLogBase
}

// Zheng 争上游ゲームクラス
type Zheng struct {
	trumpCards *TrumpCards
	players    []*ZhengPlayer
	config     ZhengConfig
	round      zhengRoundState
}

// NewZheng コンストラクタ
func NewZheng(trumpCards *TrumpCards, players []*ZhengPlayer, config ZhengConfig) *Zheng {
	return &Zheng{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: zhengRoundState{
			lastPlayPlayerIdx: -1,
			actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultZheng returns Zheng with the standard 4-player setup (1 human, 3 CPU)
// on a 54-card deck (52 + 2 jokers).
func NewDefaultZheng() *Zheng {
	config := DefaultZhengConfig()
	players := []*ZhengPlayer{
		NewZhengPlayer(true),
		NewZhengPlayer(false),
		NewZhengPlayer(false),
		NewZhengPlayer(false),
	}
	return NewZheng(NewTrumpCards(ZhengJokerCount), players, config)
}

// Reset ゲーム初期化
func (z *Zheng) Reset() {
	z.round = zhengRoundState{
		lastPlayPlayerIdx: -1,
		actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}

	z.trumpCards.Shuffle()

	resetPlayers(z.players, func(p *ZhengPlayer) {
		p.SetRank(-1)
	})

	dealAllCards(z.trumpCards, z.players)

	for _, p := range z.players {
		p.SortCardsByZhengStrength()
	}

	z.round.currentTurn = z.findSpade3Holder()
}

// findSpade3Holder ♠3を持っているプレイヤーを探す (最初のリード役)
func (z *Zheng) findSpade3Holder() int {
	for i, p := range z.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c == nil {
				continue
			}
			if c.GetValue() == 3 && c.GetDesign() == CardDesignSpade {
				return i
			}
		}
	}
	return 0
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
func (z *Zheng) PlayerPlay(indices []int) error {
	if z.round.gameEndFlag {
		return ErrGameEnded
	}
	if !z.players[z.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	z.round.cpuActions = nil

	if len(indices) == 0 {
		if len(z.round.tableCards) == 0 {
			return NewDomainError(ErrInvalidPlay, "must play cards when table is empty")
		}
		z.round.passCount++
		z.round.humanAction = &ZhengAction{PlayerIdx: z.round.currentTurn, PlayedCards: nil}
		z.appendLog(z.round.currentTurn, "pass", "pass", nil)
		z.advanceTurn()
		z.checkPassClear()
		return nil
	}

	// 重複インデックスを除去
	{
		cp := make([]int, len(indices))
		copy(cp, indices)
		sort.Ints(cp)
		unique := make([]int, 0, len(cp))
		for i, idx := range cp {
			if i == 0 || idx != cp[i-1] {
				unique = append(unique, idx)
			}
		}
		indices = unique
	}

	player := z.players[z.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}

	// **何が悪いのかまで返す。**一律 "cannot be played" だと、枚数・役の種類・強さの
	// どれで弾かれたのかが CUI からは分からない (Web は理由を出し分けている)。
	if reason := ZhengInvalidReason(selectedCards, z.round.tableCards, z.round.tablePlayType); reason != "" {
		return NewDomainError(ErrInvalidPlay, zhengInvalidReasonMessage(reason))
	}

	cards := player.RemoveCards(indices)
	playType := zhengClassifyPlay(cards)
	z.round.humanAction = &ZhengAction{PlayerIdx: z.round.currentTurn, PlayedCards: cards}
	z.appendLog(z.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	z.playCards(z.round.currentTurn, cards, playType)
	return nil
}

// playCards カードプレイ後の共通処理
func (z *Zheng) playCards(playerIdx int, cards []*Card, playType ZhengPlayType) {
	z.round.tableCards = cards
	z.round.tablePlayType = playType
	z.round.lastPlayPlayerIdx = playerIdx
	z.round.passCount = 0

	if z.players[playerIdx].GetCardsSize() == 0 {
		z.finishPlayer(playerIdx)
	}

	if !z.checkGameEnd() {
		z.advanceTurn()
		z.checkPassClear()
	}
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (z *Zheng) CpuPlay() {
	if z.round.gameEndFlag || z.players[z.round.currentTurn].GetIsHuman() {
		return
	}

	playerIdx := z.round.currentTurn
	player := z.players[playerIdx]

	playIndices := z.findBestPlay(player)

	if len(playIndices) == 0 {
		z.round.passCount++
		action := &ZhengAction{PlayerIdx: playerIdx, PlayedCards: nil}
		z.round.cpuActions = append(z.round.cpuActions, action)
		z.appendLog(playerIdx, "pass", "pass", nil)
		z.advanceTurn()
		z.checkPassClear()
	} else {
		cards := player.RemoveCards(playIndices)
		playType := zhengClassifyPlay(cards)
		action := &ZhengAction{PlayerIdx: playerIdx, PlayedCards: cards}
		z.round.cpuActions = append(z.round.cpuActions, action)
		z.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
		z.playCards(playerIdx, cards, playType)
	}
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (z *Zheng) advanceTurn() {
	if z.round.gameEndFlag {
		return
	}
	next := nextActivePlayer(z.players, z.round.currentTurn, 1)
	if next >= 0 {
		z.round.currentTurn = next
	}
}

// clearTable 場をリセットして次のリードを自由にする
func (z *Zheng) clearTable() {
	z.round.tableCards = nil
	z.round.tablePlayType = ZhengPlayInvalid
	z.round.lastPlayPlayerIdx = -1
	z.round.passCount = 0
}

// checkPassClear 全員パスしたら場をクリアする。
//
// 通常は、場にカードを出したプレイヤー (lastPlayPlayerIdx) まで手番が一周
// すれば（＝他全員がパスすれば）クリアし、そのプレイヤーが次のリードを取る。
// ただしそのプレイヤーが上がって場を離れている場合は手番が戻ってこないため、
// 残りのアクティブプレイヤー全員がパスし切った時点でクリアし、上がった
// プレイヤーの次のアクティブプレイヤーがリードを取る。
func (z *Zheng) checkPassClear() {
	if len(z.round.tableCards) == 0 || z.round.lastPlayPlayerIdx < 0 {
		return
	}
	if z.players[z.round.lastPlayPlayerIdx].GetIsFinished() {
		active := countPlayers(z.players, func(p *ZhengPlayer) bool { return !p.GetIsFinished() })
		if z.round.passCount >= active {
			z.clearTable()
		}
		return
	}
	if z.round.currentTurn == z.round.lastPlayPlayerIdx {
		z.clearTable()
	}
}

// finishPlayer プレイヤーを上がりにしてランクを付与。
// 場のクリアはここでは行わない（checkPassClear が、上がったプレイヤーの一手に
// 他プレイヤーが応じる機会を確保したうえでクリアする）。
func (z *Zheng) finishPlayer(idx int) {
	rank := z.countFinished() + 1
	z.players[idx].SetIsFinished(true)
	z.players[idx].SetRank(rank)
	z.appendLog(idx, "finish", fmt.Sprintf("player %d finished (rank %d)", idx, rank), nil)
}

// checkGameEnd ゲーム終了チェック
func (z *Zheng) checkGameEnd() bool {
	active := ZhengPlayerCnt - z.countFinished()
	if active <= 1 {
		for i, p := range z.players {
			if !p.GetIsFinished() {
				z.finishPlayer(i)
				break
			}
		}
		z.round.gameEndFlag = true
		return true
	}
	return false
}

// countFinished 上がり済みプレイヤー数を返す
func (z *Zheng) countFinished() int {
	return countPlayers(z.players, func(p *ZhengPlayer) bool { return p.GetIsFinished() })
}

// HasPendingAction ペンディングアクションがあるか (争上游にはない)
func (z *Zheng) HasPendingAction() bool { return false }

// IsHumanTurn 現在の手番が人間かどうか
func (z *Zheng) IsHumanTurn() bool {
	return z.players[z.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (z *Zheng) GetCurrentTurn() int { return z.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (z *Zheng) GetGameEndFlag() bool { return z.round.gameEndFlag }

// GetTableCards 場のカード取得
func (z *Zheng) GetTableCards() []*Card { return z.round.tableCards }

// GetTablePlayType 場のプレイタイプ取得
func (z *Zheng) GetTablePlayType() ZhengPlayType { return z.round.tablePlayType }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得
func (z *Zheng) GetLastPlayPlayerIdx() int { return z.round.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (z *Zheng) GetPlayer(idx int) *ZhengPlayer {
	return getPlayer(z.players, idx)
}

// GetPlayerCnt プレイヤー数取得
func (z *Zheng) GetPlayerCnt() int { return len(z.players) }

// GetCpuActions CPUターンの行動履歴取得
func (z *Zheng) GetCpuActions() []*ZhengAction { return z.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (z *Zheng) GetHumanAction() *ZhengAction { return z.round.humanAction }

// GetPassCount 現在のパスカウント取得
func (z *Zheng) GetPassCount() int { return z.round.passCount }

// GetConfig 設定取得
func (z *Zheng) GetConfig() ZhengConfig { return z.config }

// SetConfig 設定変更
func (z *Zheng) SetConfig(config ZhengConfig) { z.config = config }

// GetActionLog 棋譜を取得する
func (z *Zheng) GetActionLog() []*ActionLogEntry { return z.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (z *Zheng) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	z.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- JSON Serialization ---

// zhengActionJSON is the JSON wire format for ZhengAction.
type zhengActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (a *ZhengAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(zhengActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *ZhengAction) UnmarshalJSON(data []byte) error {
	var j zhengActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	return nil
}

// zhengJSON is the JSON wire format for Zheng.
type zhengJSON struct {
	TrumpCards        *TrumpCards       `json:"tc"`
	Players           []*ZhengPlayer    `json:"pl"`
	Config            ZhengConfig       `json:"cf"`
	CurrentTurn       int               `json:"ct"`
	TableCards        []*Card           `json:"tb"`
	TablePlayType     ZhengPlayType     `json:"tt"`
	LastPlayPlayerIdx int               `json:"lp"`
	GameEndFlag       bool              `json:"ge"`
	PassCount         int               `json:"pc"`
	CpuActions        []*ZhengAction    `json:"ca"`
	HumanAction       *ZhengAction      `json:"ha"`
	ActionLog         []*ActionLogEntry `json:"al"`
}

// zhengMaxSliceLen caps slice sizes during deserialisation.
const zhengMaxSliceLen = 1000

// zheng deserialisation validation errors.
var (
	errZhengOversized       = errors.New("zheng: input array exceeds maximum allowed size")
	errZhengInvalidPlayers  = errors.New("zheng: invalid players")
	errZhengInvalidCard     = errors.New("zheng: invalid card")
	errZhengInvalidIndex    = errors.New("zheng: player index out of range")
	errZhengInvalidRank     = errors.New("zheng: invalid rank")
	errZhengInvalidPlayType = errors.New("zheng: invalid table play type")
	errZhengInvalidAction   = errors.New("zheng: invalid action")
)

// zhengValidCard は復元されたカードの整合性 (デザイン・値の範囲) を検証する。
// ジョーカー (design 0) は値1 (小) / 2 (大) のみ許容する。
func zhengValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignJoker {
		return v == 1 || v == 2
	}
	return d >= CardDesignSpade && d <= CardDesignMax && v >= 1 && v <= CardValueMax
}

// zhengCheckCards はカードスライス内の全カードを検証する。
func zhengCheckCards(cards []*Card) error {
	for _, c := range cards {
		if !zhengValidCard(c) {
			return errZhengInvalidCard
		}
	}
	return nil
}

// zhengInRange はプレイヤーインデックスが [0, ZhengPlayerCnt) かを返す。
func zhengInRange(i int) bool { return i >= 0 && i < ZhengPlayerCnt }

// zhengInRangeOrUnset はプレイヤーインデックスが -1 または範囲内かを返す。
func zhengInRangeOrUnset(i int) bool { return i == -1 || zhengInRange(i) }

// zhengCheckAction は復元された行動記録を検証する (nil行動は許容)。
func zhengCheckAction(a *ZhengAction) error {
	if a == nil {
		return nil
	}
	if !zhengInRange(a.PlayerIdx) {
		return errZhengInvalidAction
	}
	return zhengCheckCards(a.PlayedCards)
}

// MarshalJSON implements json.Marshaler.
func (z *Zheng) MarshalJSON() ([]byte, error) {
	return json.Marshal(zhengJSON{
		TrumpCards:        z.trumpCards,
		Players:           z.players,
		Config:            z.config,
		CurrentTurn:       z.round.currentTurn,
		TableCards:        z.round.tableCards,
		TablePlayType:     z.round.tablePlayType,
		LastPlayPlayerIdx: z.round.lastPlayPlayerIdx,
		GameEndFlag:       z.round.gameEndFlag,
		PassCount:         z.round.passCount,
		CpuActions:        z.round.cpuActions,
		HumanAction:       z.round.humanAction,
		ActionLog:         z.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (z *Zheng) UnmarshalJSON(data []byte) error {
	var j zhengJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > zhengMaxSliceLen || len(j.TableCards) > zhengMaxSliceLen ||
		len(j.CpuActions) > zhengMaxSliceLen || len(j.ActionLog) > zhengMaxSliceLen {
		return errZhengOversized
	}
	if len(j.Players) != ZhengPlayerCnt {
		return errZhengInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil || p.RankedGamePlayer == nil {
			return errZhengInvalidPlayers
		}
		if r := p.GetRank(); r < -1 || r > ZhengPlayerCnt {
			return errZhengInvalidRank
		}
		for i := 0; i < p.GetCardsSize(); i++ {
			if !zhengValidCard(p.GetCard(i)) {
				return errZhengInvalidCard
			}
		}
	}
	if !zhengInRangeOrUnset(j.CurrentTurn) || !zhengInRangeOrUnset(j.LastPlayPlayerIdx) {
		return errZhengInvalidIndex
	}
	if j.TablePlayType < ZhengPlayInvalid || j.TablePlayType > ZhengPlayJokerBomb {
		return errZhengInvalidPlayType
	}
	if err := zhengCheckCards(j.TableCards); err != nil {
		return err
	}
	for _, a := range j.CpuActions {
		if a == nil {
			return errZhengInvalidAction
		}
		if err := zhengCheckAction(a); err != nil {
			return err
		}
	}
	if err := zhengCheckAction(j.HumanAction); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	z.trumpCards = j.TrumpCards
	if z.trumpCards == nil {
		z.trumpCards = NewTrumpCards(ZhengJokerCount)
	}
	z.players = j.Players
	z.config = j.Config
	z.round = zhengRoundState{
		currentTurn:       max(j.CurrentTurn, 0),
		tableCards:        j.TableCards,
		tablePlayType:     j.TablePlayType,
		lastPlayPlayerIdx: j.LastPlayPlayerIdx,
		gameEndFlag:       j.GameEndFlag,
		passCount:         max(j.PassCount, 0),
		cpuActions:        j.CpuActions,
		humanAction:       j.HumanAction,
		actionLogBase:     actionLogBase{actionLog: j.ActionLog},
	}
	if z.round.actionLog == nil {
		z.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// zhengInvalidReasonMessage は出せない理由の識別子を人間向けの文にする。
func zhengInvalidReasonMessage(reason string) string {
	switch reason {
	case ZhengInvalidType:
		return "選んだ札はどの役にもなりません"
	case ZhengInvalidWrongType:
		return "場と同じ役の種類で出してください"
	case ZhengInvalidWrongCount:
		return "場と同じ枚数で出してください"
	case ZhengInvalidTooWeak:
		return "場より強い役で出してください"
	case ZhengInvalidNeedBomb:
		return "場が爆弾です。爆弾でしか切れません"
	case ZhengInvalidUnbeatable:
		return "場がジョーカーボムです。勝てる役はありません"
	default:
		return "selected cards cannot be played"
	}
}
