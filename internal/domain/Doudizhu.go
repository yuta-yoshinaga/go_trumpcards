//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// DoudizhuPlayerCnt 斗地主プレイヤー数
const DoudizhuPlayerCnt = 3

// DoudizhuJokerCount 斗地主で使用するジョーカー枚数
const DoudizhuJokerCount = 2

// DoudizhuHandSize 初期手札枚数
const DoudizhuHandSize = 17

// DoudizhuKittyCount 底牌枚数
const DoudizhuKittyCount = 3

// DoudizhuMaxBid 最大ビッド値
const DoudizhuMaxBid = 3

// DoudizhuPhase ゲームフェーズ
type DoudizhuPhase int

const (
	// DoudizhuPhaseBid 競り (叫地主) フェーズ
	DoudizhuPhaseBid DoudizhuPhase = 0
	// DoudizhuPhasePlay プレイフェーズ
	DoudizhuPhasePlay DoudizhuPhase = 1
	// DoudizhuPhaseEnd ゲーム終了
	DoudizhuPhaseEnd DoudizhuPhase = 2
)

// DoudizhuCpuAction CPUまたは人間の1ターン分の行動記録
type DoudizhuCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
	BidValue    int     // ビッドフェーズ: 0=パス, 1-3=ビッド値
}

// doudizhuRoundState ラウンドごとにリセットされる状態
type doudizhuRoundState struct {
	phase         DoudizhuPhase
	currentTurn   int
	tableCombo    *DoudizhuCombo
	lastPlayIdx   int
	gameEndFlag   bool
	passCount     int
	kittyCards    []*Card
	bidValues     [DoudizhuPlayerCnt]int
	bidCount      int
	highestBid    int
	highestBidder int
	landlordIdx   int
	baseBid       int
	bombCount     int
	cpuActions    []*DoudizhuCpuAction
	humanAction   *DoudizhuCpuAction
	scores        [DoudizhuPlayerCnt]int
	actionLogBase
}

// Doudizhu 斗地主ゲーム
type Doudizhu struct {
	trumpCards *TrumpCards
	players    []*DoudizhuPlayer
	config     DoudizhuConfig
	round      doudizhuRoundState
}

// NewDoudizhu コンストラクタ
func NewDoudizhu(trumpCards *TrumpCards, players []*DoudizhuPlayer, config DoudizhuConfig) *Doudizhu {
	return &Doudizhu{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: doudizhuRoundState{
			lastPlayIdx:   -1,
			landlordIdx:   -1,
			highestBidder: -1,
		},
	}
}

// NewDefaultDoudizhu 標準の3人セットアップ (1人間, 2CPU) で斗地主を作成
func NewDefaultDoudizhu() *Doudizhu {
	config := DefaultDoudizhuConfig()
	players := []*DoudizhuPlayer{
		NewDoudizhuPlayer(true),
		NewDoudizhuPlayer(false),
		NewDoudizhuPlayer(false),
	}
	return NewDoudizhu(NewTrumpCards(DoudizhuJokerCount), players, config)
}

// Reset ゲーム初期化
func (d *Doudizhu) Reset() {
	d.round = doudizhuRoundState{
		lastPlayIdx:   -1,
		landlordIdx:   -1,
		highestBidder: -1,
	}

	d.trumpCards.Shuffle()

	resetPlayers(d.players, func(p *DoudizhuPlayer) {
		p.SetIsLandlord(false)
	})

	rand.Shuffle(len(d.players), func(i, j int) {
		d.players[i], d.players[j] = d.players[j], d.players[i]
	})

	for i := 0; i < DoudizhuHandSize*DoudizhuPlayerCnt; i++ {
		card := d.trumpCards.DrawCard()
		d.players[i%DoudizhuPlayerCnt].AddCard(card)
	}

	d.round.kittyCards = make([]*Card, 0, DoudizhuKittyCount)
	for i := 0; i < DoudizhuKittyCount; i++ {
		d.round.kittyCards = append(d.round.kittyCards, d.trumpCards.DrawCard())
	}

	for _, p := range d.players {
		p.SortCardsByStrength()
	}

	d.round.phase = DoudizhuPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする (0=パス, 1-3=ビッド値)
func (d *Doudizhu) PlayerBid(value int) error {
	if d.round.phase != DoudizhuPhaseBid {
		return NewDomainError(ErrInvalidPlay, "not in bidding phase")
	}
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	return d.executeBid(value)
}

// executeBid ビッドの共通処理
func (d *Doudizhu) executeBid(value int) error {
	if value < 0 || value > DoudizhuMaxBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("bid value must be 0-%d", DoudizhuMaxBid))
	}
	if value > 0 && value <= d.round.highestBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("bid must exceed current highest bid %d", d.round.highestBid))
	}

	playerIdx := d.round.currentTurn
	d.round.cpuActions = nil

	action := &DoudizhuCpuAction{PlayerIdx: playerIdx, BidValue: value}
	if d.players[playerIdx].GetIsHuman() {
		d.round.humanAction = action
	} else {
		d.round.cpuActions = append(d.round.cpuActions, action)
	}

	if value > 0 {
		d.round.bidValues[playerIdx] = value
		d.round.highestBid = value
		d.round.highestBidder = playerIdx
		d.appendLog(playerIdx, "bid", fmt.Sprintf("bid %d", value), nil)
	} else {
		d.appendLog(playerIdx, "bid", "pass", nil)
	}
	d.round.bidCount++

	if value == DoudizhuMaxBid {
		d.decideLandlord(playerIdx)
		return nil
	}

	if d.round.bidCount >= DoudizhuPlayerCnt {
		if d.round.highestBidder >= 0 {
			d.decideLandlord(d.round.highestBidder)
		} else {
			d.Reset()
		}
		return nil
	}

	d.round.currentTurn = (d.round.currentTurn + 1) % DoudizhuPlayerCnt
	return nil
}

// decideLandlord 地主を決定し、底牌を渡してプレイフェーズへ
func (d *Doudizhu) decideLandlord(idx int) {
	d.round.landlordIdx = idx
	d.round.baseBid = d.round.highestBid
	d.players[idx].SetIsLandlord(true)

	for _, c := range d.round.kittyCards {
		d.players[idx].AddCard(c)
	}
	d.players[idx].SortCardsByStrength()

	d.appendLog(-1, "landlord", fmt.Sprintf("player %d is the landlord", idx), d.round.kittyCards)

	d.round.phase = DoudizhuPhasePlay
	d.round.currentTurn = idx
}

// PlayerPlay 人間プレイヤーがカードを出す (空=パス)
func (d *Doudizhu) PlayerPlay(indices []int) error {
	if d.round.phase != DoudizhuPhasePlay {
		return NewDomainError(ErrInvalidPlay, "not in play phase")
	}
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	d.round.cpuActions = nil

	if len(indices) == 0 {
		if d.round.tableCombo == nil {
			return NewDomainError(ErrInvalidPlay, "must play when leading")
		}
		d.round.passCount++
		d.round.humanAction = &DoudizhuCpuAction{PlayerIdx: d.round.currentTurn}
		d.appendLog(d.round.currentTurn, "pass", "pass", nil)
		d.advanceTurn()
		d.checkPassClear()
		return nil
	}

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

	player := d.players[d.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}

	combo := DoudizhuClassifyCombo(selectedCards)
	if combo == nil {
		return NewDomainError(ErrInvalidPlay, "invalid card combination")
	}

	if d.round.tableCombo != nil {
		if !DoudizhuCanBeat(combo, d.round.tableCombo) {
			return NewDomainError(ErrInvalidPlay, "cards cannot beat the current table")
		}
	}

	cards := player.RemoveCards(indices)
	combo.Cards = cards
	d.round.humanAction = &DoudizhuCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: cards}
	d.appendLog(d.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	d.playCards(d.round.currentTurn, combo)
	return nil
}

// playCards カードプレイ後の共通処理
func (d *Doudizhu) playCards(playerIdx int, combo *DoudizhuCombo) {
	d.round.tableCombo = combo
	d.round.lastPlayIdx = playerIdx
	d.round.passCount = 0

	if combo.Type == DoudizhuComboBomb || combo.Type == DoudizhuComboRocket {
		d.round.bombCount++
	}

	if d.players[playerIdx].GetCardsSize() == 0 {
		d.players[playerIdx].SetIsFinished(true)
		d.endGame()
		return
	}

	d.advanceTurn()
	d.checkPassClear()
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (d *Doudizhu) CpuPlay() {
	if d.round.gameEndFlag || d.players[d.round.currentTurn].GetIsHuman() {
		return
	}

	if d.round.phase == DoudizhuPhaseBid {
		d.cpuBid()
		return
	}

	playerIdx := d.round.currentTurn
	player := d.players[playerIdx]
	playIndices := d.findBestPlay(player)

	if len(playIndices) == 0 {
		d.round.passCount++
		action := &DoudizhuCpuAction{PlayerIdx: playerIdx}
		d.round.cpuActions = append(d.round.cpuActions, action)
		d.appendLog(playerIdx, "pass", "pass", nil)
		d.advanceTurn()
		d.checkPassClear()
	} else {
		selectedCards := make([]*Card, len(playIndices))
		for i, idx := range playIndices {
			selectedCards[i] = player.GetCard(idx)
		}
		combo := DoudizhuClassifyCombo(selectedCards)
		cards := player.RemoveCards(playIndices)
		combo.Cards = cards
		action := &DoudizhuCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
		d.round.cpuActions = append(d.round.cpuActions, action)
		d.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
		d.playCards(playerIdx, combo)
	}
}

// advanceTurn 手番を次のプレイヤーへ進める
func (d *Doudizhu) advanceTurn() {
	if d.round.gameEndFlag {
		return
	}
	d.round.currentTurn = (d.round.currentTurn + 1) % DoudizhuPlayerCnt
}

// checkPassClear 全員パスしたら場をクリアする
func (d *Doudizhu) checkPassClear() {
	if d.round.tableCombo == nil || d.round.lastPlayIdx < 0 {
		return
	}
	if d.round.currentTurn == d.round.lastPlayIdx {
		d.round.tableCombo = nil
		d.round.lastPlayIdx = -1
		d.round.passCount = 0
	}
}

// endGame ゲーム終了処理
func (d *Doudizhu) endGame() {
	d.round.gameEndFlag = true
	d.round.phase = DoudizhuPhaseEnd

	score := d.round.baseBid
	for i := 0; i < d.round.bombCount; i++ {
		score *= 2
	}

	landlordWon := d.players[d.round.landlordIdx].GetIsFinished()
	for i := 0; i < DoudizhuPlayerCnt; i++ {
		if i == d.round.landlordIdx {
			if landlordWon {
				d.round.scores[i] = score * (DoudizhuPlayerCnt - 1)
			} else {
				d.round.scores[i] = -score * (DoudizhuPlayerCnt - 1)
			}
		} else {
			if landlordWon {
				d.round.scores[i] = -score
			} else {
				d.round.scores[i] = score
			}
		}
	}

	d.appendLog(-1, "end", "game over", nil)
}

// --- Getters ---

// IsHumanTurn 現在の手番が人間かどうか
func (d *Doudizhu) IsHumanTurn() bool {
	return d.players[d.round.currentTurn].GetIsHuman()
}

// GetPhase 現在のフェーズ取得
func (d *Doudizhu) GetPhase() DoudizhuPhase { return d.round.phase }

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (d *Doudizhu) GetCurrentTurn() int { return d.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (d *Doudizhu) GetGameEndFlag() bool { return d.round.gameEndFlag }

// GetTableCombo 場の役取得 (nil = クリア)
func (d *Doudizhu) GetTableCombo() *DoudizhuCombo { return d.round.tableCombo }

// GetLastPlayIdx 最後にカードを出したプレイヤーインデックス取得
func (d *Doudizhu) GetLastPlayIdx() int { return d.round.lastPlayIdx }

// GetPlayer プレイヤー取得
func (d *Doudizhu) GetPlayer(idx int) *DoudizhuPlayer {
	return getPlayer(d.players, idx)
}

// GetPlayerCnt プレイヤー数取得
func (d *Doudizhu) GetPlayerCnt() int { return len(d.players) }

// GetKittyCards 底牌取得
func (d *Doudizhu) GetKittyCards() []*Card { return d.round.kittyCards }

// GetLandlordIdx 地主プレイヤーインデックス取得 (-1=未決定)
func (d *Doudizhu) GetLandlordIdx() int { return d.round.landlordIdx }

// GetBaseBid ビッド値取得
func (d *Doudizhu) GetBaseBid() int { return d.round.baseBid }

// GetBombCount ボム/ロケット使用回数取得
func (d *Doudizhu) GetBombCount() int { return d.round.bombCount }

// GetScores スコア取得
func (d *Doudizhu) GetScores() [DoudizhuPlayerCnt]int { return d.round.scores }

// GetBidValues ビッド値取得
func (d *Doudizhu) GetBidValues() [DoudizhuPlayerCnt]int { return d.round.bidValues }

// GetHighestBid 現在の最高ビッド値取得
func (d *Doudizhu) GetHighestBid() int { return d.round.highestBid }

// GetCpuActions CPUターンの行動履歴取得
func (d *Doudizhu) GetCpuActions() []*DoudizhuCpuAction { return d.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (d *Doudizhu) GetHumanAction() *DoudizhuCpuAction { return d.round.humanAction }

// GetConfig 設定取得
func (d *Doudizhu) GetConfig() DoudizhuConfig { return d.config }

// SetConfig 設定変更
func (d *Doudizhu) SetConfig(config DoudizhuConfig) { d.config = config }

// HasPendingAction ペンディングアクションがあるか (常にfalse)
func (d *Doudizhu) HasPendingAction() bool { return false }

// GetActionLog 棋譜を取得する
func (d *Doudizhu) GetActionLog() []*ActionLogEntry { return d.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (d *Doudizhu) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	d.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- JSON Serialization ---

// doudizhuCpuActionJSON is the JSON wire format for DoudizhuCpuAction.
type doudizhuCpuActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
	BidValue    int     `json:"bv"`
}

// MarshalJSON implements json.Marshaler.
func (a *DoudizhuCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(doudizhuCpuActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
		BidValue:    a.BidValue,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *DoudizhuCpuAction) UnmarshalJSON(data []byte) error {
	var j doudizhuCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	a.BidValue = j.BidValue
	return nil
}

// doudizhuComboJSON is the JSON wire format for DoudizhuCombo.
type doudizhuComboJSON struct {
	Type   DoudizhuComboType `json:"ty"`
	Cards  []*Card           `json:"cs"`
	Rank   int               `json:"rk"`
	Length int               `json:"ln"`
}

// MarshalJSON implements json.Marshaler.
func (c *DoudizhuCombo) MarshalJSON() ([]byte, error) {
	return json.Marshal(doudizhuComboJSON{
		Type:   c.Type,
		Cards:  c.Cards,
		Rank:   c.Rank,
		Length: c.Length,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DoudizhuCombo) UnmarshalJSON(data []byte) error {
	var j doudizhuComboJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.Type = j.Type
	c.Cards = j.Cards
	c.Rank = j.Rank
	c.Length = j.Length
	return nil
}

// doudizhuJSON is the JSON wire format for Doudizhu (flattens doudizhuRoundState).
type doudizhuJSON struct {
	TrumpCards    *TrumpCards            `json:"tc"`
	Players       []*DoudizhuPlayer      `json:"pl"`
	Config        DoudizhuConfig         `json:"cf"`
	Phase         DoudizhuPhase          `json:"ph"`
	CurrentTurn   int                    `json:"ct"`
	TableCombo    *DoudizhuCombo         `json:"tb"`
	LastPlayIdx   int                    `json:"lp"`
	GameEndFlag   bool                   `json:"ge"`
	PassCount     int                    `json:"pc"`
	KittyCards    []*Card                `json:"kc"`
	BidValues     [DoudizhuPlayerCnt]int `json:"bv"`
	BidCount      int                    `json:"bc"`
	HighestBid    int                    `json:"hb"`
	HighestBidder int                    `json:"hd"`
	LandlordIdx   int                    `json:"li"`
	BaseBid       int                    `json:"bb"`
	BombCount     int                    `json:"bo"`
	CpuActions    []*DoudizhuCpuAction   `json:"ca"`
	HumanAction   *DoudizhuCpuAction     `json:"ha"`
	Scores        [DoudizhuPlayerCnt]int `json:"sc"`
	ActionLog     []*ActionLogEntry      `json:"al"`
}

const doudizhuMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (d *Doudizhu) MarshalJSON() ([]byte, error) {
	return json.Marshal(doudizhuJSON{
		TrumpCards:    d.trumpCards,
		Players:       d.players,
		Config:        d.config,
		Phase:         d.round.phase,
		CurrentTurn:   d.round.currentTurn,
		TableCombo:    d.round.tableCombo,
		LastPlayIdx:   d.round.lastPlayIdx,
		GameEndFlag:   d.round.gameEndFlag,
		PassCount:     d.round.passCount,
		KittyCards:    d.round.kittyCards,
		BidValues:     d.round.bidValues,
		BidCount:      d.round.bidCount,
		HighestBid:    d.round.highestBid,
		HighestBidder: d.round.highestBidder,
		LandlordIdx:   d.round.landlordIdx,
		BaseBid:       d.round.baseBid,
		BombCount:     d.round.bombCount,
		CpuActions:    d.round.cpuActions,
		HumanAction:   d.round.humanAction,
		Scores:        d.round.scores,
		ActionLog:     d.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Doudizhu) UnmarshalJSON(data []byte) error {
	var j doudizhuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > doudizhuMaxSliceLen || len(j.KittyCards) > doudizhuMaxSliceLen ||
		len(j.CpuActions) > doudizhuMaxSliceLen || len(j.ActionLog) > doudizhuMaxSliceLen {
		return fmt.Errorf("doudizhu: input array exceeds maximum allowed size")
	}
	d.trumpCards = j.TrumpCards
	if d.trumpCards == nil {
		d.trumpCards = NewTrumpCards(0)
	}
	d.players = j.Players
	if d.players == nil {
		d.players = make([]*DoudizhuPlayer, 0)
	}
	d.config = j.Config
	d.round = doudizhuRoundState{
		phase:         j.Phase,
		currentTurn:   j.CurrentTurn,
		tableCombo:    j.TableCombo,
		lastPlayIdx:   j.LastPlayIdx,
		gameEndFlag:   j.GameEndFlag,
		passCount:     j.PassCount,
		kittyCards:    j.KittyCards,
		bidValues:     j.BidValues,
		bidCount:      j.BidCount,
		highestBid:    j.HighestBid,
		highestBidder: j.HighestBidder,
		landlordIdx:   j.LandlordIdx,
		baseBid:       j.BaseBid,
		bombCount:     j.BombCount,
		cpuActions:    j.CpuActions,
		humanAction:   j.HumanAction,
		scores:        j.Scores,
		actionLogBase: actionLogBase{actionLog: j.ActionLog},
	}
	if d.round.actionLog == nil {
		d.round.actionLog = make([]*ActionLogEntry, 0)
	}
	if d.round.kittyCards == nil {
		d.round.kittyCards = make([]*Card, 0)
	}
	return nil
}
