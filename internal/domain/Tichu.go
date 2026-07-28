//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
)

// TichuPlayerCnt ティチュープレイヤー数
const TichuPlayerCnt = 4

// TichuJokerCount 使用する特殊カード枚数 (麻雀/犬/鳳凰/龍)
const TichuJokerCount = 4

// TichuHandSize 各自の初期手札枚数 (56枚 / 4人)
const TichuHandSize = 14

// TichuWinScore マッチ勝利に必要な得点 (参考値)
const TichuWinScore = 1000

// TichuPhase ゲームフェーズ
type TichuPhase int

const (
	// TichuPhaseDeclare 宣言フェーズ (ティチュー/グランド)
	TichuPhaseDeclare TichuPhase = 0
	// TichuPhasePlay プレイフェーズ
	TichuPhasePlay TichuPhase = 1
	// TichuPhaseEnd ディール終了
	TichuPhaseEnd TichuPhase = 2
)

// TichuCpuAction CPUまたは人間の1ターン分の行動記録
type TichuCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
	DeclType    int     // 宣言フェーズ: 宣言種別
	IsPass      bool    // パスかどうか
}

// tichuRoundState ディールごとにリセットされる状態
type tichuRoundState struct {
	phase       TichuPhase
	currentTurn int
	tableCombo  *TichuCombo
	lastPlayIdx int
	trickCards  []*Card
	passCount   int
	declCount   int
	dragonOnTop bool
	gameEndFlag bool
	oneTwo      bool
	startLeader int
	finishOrder []int
	scores      [2]int
	bombCount   int
	cpuActions  []*TichuCpuAction
	humanAction *TichuCpuAction
	actionLog   []*ActionLogEntry
}

// Tichu ティチューゲーム
type Tichu struct {
	trumpCards *TrumpCards
	players    []*TichuPlayer
	config     TichuConfig
	round      tichuRoundState
}

// NewTichu コンストラクタ
func NewTichu(trumpCards *TrumpCards, players []*TichuPlayer, config TichuConfig) *Tichu {
	return &Tichu{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round:      tichuRoundState{lastPlayIdx: -1},
	}
}

// NewDefaultTichu 標準の4人セットアップ (1人間, 3CPU) で作成
func NewDefaultTichu() *Tichu {
	config := DefaultTichuConfig()
	players := []*TichuPlayer{
		NewTichuPlayer(true),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
	}
	return NewTichu(NewTrumpCards(TichuJokerCount), players, config)
}

// TichuTeamOf プレイヤーインデックスのチーム (0 = {0,2}, 1 = {1,3})
func TichuTeamOf(idx int) int { return idx % 2 }

// Reset ディール初期化
func (t *Tichu) Reset() {
	t.round = tichuRoundState{lastPlayIdx: -1}

	t.trumpCards.Shuffle()
	resetPlayers(t.players, func(p *TichuPlayer) {
		p.SetRank(0)
		p.SetDeclType(TichuDeclNone)
		p.collected = make([]*Card, 0)
	})

	for i := 0; i < TichuHandSize*TichuPlayerCnt; i++ {
		card := t.trumpCards.DrawCard()
		if card == nil {
			break
		}
		t.players[i%TichuPlayerCnt].AddCard(card)
	}
	for _, p := range t.players {
		p.SortCardsByStrength()
	}

	t.round.startLeader = t.findMahjongHolder()
	t.round.phase = TichuPhaseDeclare
	t.round.currentTurn = t.round.startLeader
	t.round.trickCards = make([]*Card, 0)
}

// findMahjongHolder 麻雀カードの保持者を返す (見つからなければ0)
func (t *Tichu) findMahjongHolder() int {
	for i, p := range t.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			if tichuSpecialKind(p.GetCard(j)) == TichuMahjong {
				return i
			}
		}
	}
	return 0
}

// PlayerDeclare 人間プレイヤーが宣言する (0=なし, 1=ティチュー, 2=グランド)
func (t *Tichu) PlayerDeclare(declType int) error {
	if t.round.phase != TichuPhaseDeclare {
		return NewDomainError(ErrInvalidPlay, "not in declaration phase")
	}
	if !t.players[t.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if declType < TichuDeclNone || declType > TichuDeclGrand {
		return NewDomainError(ErrInvalidPlay, "invalid declaration type")
	}
	t.executeDeclare(declType)
	return nil
}

// executeDeclare 宣言の共通処理
func (t *Tichu) executeDeclare(declType int) {
	idx := t.round.currentTurn
	t.round.cpuActions = nil
	t.players[idx].SetDeclType(declType)

	action := &TichuCpuAction{PlayerIdx: idx, DeclType: declType}
	if t.players[idx].GetIsHuman() {
		t.round.humanAction = action
	} else {
		t.round.cpuActions = append(t.round.cpuActions, action)
	}

	detail := "no declaration"
	switch declType {
	case TichuDeclTichu:
		detail = "Tichu"
	case TichuDeclGrand:
		detail = "Grand Tichu"
	}
	t.appendLog(idx, "declare", detail, nil)

	t.round.declCount++
	if t.round.declCount >= TichuPlayerCnt {
		t.round.phase = TichuPhasePlay
		t.round.currentTurn = t.round.startLeader
		return
	}
	t.round.currentTurn = (t.round.currentTurn + 1) % TichuPlayerCnt
}

// PlayerPlay 人間プレイヤーがカードを出す (空=パス)
func (t *Tichu) PlayerPlay(indices []int) error {
	if t.round.phase != TichuPhasePlay {
		return NewDomainError(ErrInvalidPlay, "not in play phase")
	}
	if t.round.gameEndFlag {
		return ErrGameEnded
	}
	if !t.players[t.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	t.round.cpuActions = nil

	if len(indices) == 0 {
		if t.round.tableCombo == nil {
			return NewDomainError(ErrInvalidPlay, "must play when leading")
		}
		t.round.humanAction = &TichuCpuAction{PlayerIdx: t.round.currentTurn, IsPass: true}
		t.appendLog(t.round.currentTurn, "pass", "pass", nil)
		t.handlePass()
		return nil
	}

	player := t.players[t.round.currentTurn]
	selected := make([]*Card, 0, len(indices))
	seen := make(map[int]bool)
	for _, idx := range indices {
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "duplicate card indices selected")
		}
		seen[idx] = true
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selected = append(selected, card)
	}

	combo := ClassifyTichu(selected)
	if combo == nil {
		return NewDomainError(ErrInvalidPlay, "invalid card combination")
	}
	if err := t.validatePlay(combo); err != nil {
		return err
	}

	cards := player.RemoveCards(indices)
	combo.Cards = cards
	t.round.humanAction = &TichuCpuAction{PlayerIdx: t.round.currentTurn, PlayedCards: cards}
	t.appendLog(t.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	t.playCombo(t.round.currentTurn, combo)
	return nil
}

// validatePlay 場に対してコンボが有効かを検証する
func (t *Tichu) validatePlay(combo *TichuCombo) error {
	if combo.Type == TichuComboDog {
		if t.round.tableCombo != nil {
			return NewDomainError(ErrInvalidPlay, "dog can only be led")
		}
		return nil
	}
	if t.round.tableCombo == nil {
		return nil // リードは任意の役
	}
	// 鳳凰単体のランクを場に合わせて解決
	if combo.Type == TichuComboSingle && combo.PhoenixSingle && t.round.tableCombo.Type == TichuComboSingle {
		combo.Rank = t.round.tableCombo.Rank
	}
	if !TichuCanBeat(combo, t.round.tableCombo) {
		return NewDomainError(ErrInvalidPlay, "cards cannot beat the current table")
	}
	return nil
}

// playCombo カードプレイ後の共通処理
func (t *Tichu) playCombo(idx int, combo *TichuCombo) {
	player := t.players[idx]

	if combo.Type == TichuComboDog {
		t.appendLog(idx, "dog", "lead passed to partner", nil)
		if player.GetCardsSize() == 0 {
			t.markFinished(idx)
		}
		if t.checkDealEnd() {
			return
		}
		t.passLeadTo((idx + 2) % TichuPlayerCnt)
		return
	}

	t.round.tableCombo = combo
	t.round.lastPlayIdx = idx
	t.round.passCount = 0
	t.round.trickCards = append(t.round.trickCards, combo.Cards...)
	t.round.dragonOnTop = combo.Type == TichuComboSingle && len(combo.Cards) == 1 &&
		tichuSpecialKind(combo.Cards[0]) == TichuDragon
	if tichuIsBomb(combo) {
		t.round.bombCount++
	}

	if player.GetCardsSize() == 0 {
		t.markFinished(idx)
	}
	if t.checkDealEnd() {
		return
	}
	t.advanceTurn()
	t.maybeResolveTrick()
}

// handlePass パス処理
func (t *Tichu) handlePass() {
	t.round.passCount++
	if t.trickShouldEnd() {
		t.resolveTrick()
		return
	}
	t.advanceTurn()
}

// trickShouldEnd 現在のパス数でトリックが確定するか
func (t *Tichu) trickShouldEnd() bool {
	if t.round.tableCombo == nil || t.round.lastPlayIdx < 0 {
		return false
	}
	needed := t.countActive()
	if t.players[t.round.lastPlayIdx].GetCardsSize() > 0 {
		needed--
	}
	return needed > 0 && t.round.passCount >= needed
}

// maybeResolveTrick advanceTurn後にトリック確定をチェック
func (t *Tichu) maybeResolveTrick() {
	if t.trickShouldEnd() {
		t.resolveTrick()
	}
}

// resolveTrick トリックを確定し、勝者がカードを獲得する
func (t *Tichu) resolveTrick() {
	owner := t.round.lastPlayIdx
	if owner < 0 {
		return
	}
	collector := owner
	if t.round.dragonOnTop {
		collector = (owner + 1) % TichuPlayerCnt // 龍のトリックは右隣の相手へ
	}
	t.players[collector].AddCollected(t.round.trickCards)

	t.round.tableCombo = nil
	t.round.trickCards = make([]*Card, 0)
	t.round.passCount = 0
	t.round.dragonOnTop = false

	leader := owner
	if t.players[owner].GetCardsSize() == 0 {
		leader = t.nextActiveAfter(owner)
	}
	t.round.lastPlayIdx = -1
	t.round.currentTurn = leader
}

// passLeadTo 指定プレイヤー (不在ならその次の手番) にリードを渡す
func (t *Tichu) passLeadTo(idx int) {
	t.round.tableCombo = nil
	t.round.trickCards = make([]*Card, 0)
	t.round.passCount = 0
	t.round.dragonOnTop = false
	t.round.lastPlayIdx = -1
	if t.players[idx].GetCardsSize() == 0 {
		idx = t.nextActiveAfter(idx)
	}
	t.round.currentTurn = idx
}

// markFinished プレイヤーを上がり済みにする
func (t *Tichu) markFinished(idx int) {
	if t.players[idx].GetIsFinished() {
		return
	}
	t.players[idx].SetIsFinished(true)
	t.round.finishOrder = append(t.round.finishOrder, idx)
	t.players[idx].SetRank(len(t.round.finishOrder))
}

// checkDealEnd ディール終了条件を判定し、終了なら処理して true を返す
func (t *Tichu) checkDealEnd() bool {
	// ワンツー (同チームが1位2位)
	if len(t.round.finishOrder) == 2 {
		a, b := t.round.finishOrder[0], t.round.finishOrder[1]
		if (a+2)%TichuPlayerCnt == b {
			t.endDeal(true)
			return true
		}
	}
	if t.countActive() <= 1 {
		t.endDeal(false)
		return true
	}
	return false
}

// advanceTurn 次の上がっていないプレイヤーへ手番を進める
func (t *Tichu) advanceTurn() {
	if t.round.gameEndFlag {
		return
	}
	t.round.currentTurn = t.nextActiveAfter(t.round.currentTurn)
}

// nextActiveAfter idx の次の上がっていないプレイヤーを返す
func (t *Tichu) nextActiveAfter(idx int) int {
	for i := 1; i <= TichuPlayerCnt; i++ {
		cand := (idx + i) % TichuPlayerCnt
		if t.players[cand].GetCardsSize() > 0 {
			return cand
		}
	}
	return idx
}

// countActive 手札の残っているプレイヤー数
func (t *Tichu) countActive() int {
	n := 0
	for _, p := range t.players {
		if p.GetCardsSize() > 0 {
			n++
		}
	}
	return n
}

// endDeal ディール終了処理と得点計算
func (t *Tichu) endDeal(oneTwo bool) {
	t.round.gameEndFlag = true
	t.round.oneTwo = oneTwo
	t.round.phase = TichuPhaseEnd

	// 未確定のトリックを勝者へ
	if len(t.round.trickCards) > 0 && t.round.lastPlayIdx >= 0 {
		collector := t.round.lastPlayIdx
		if t.round.dragonOnTop {
			collector = (collector + 1) % TichuPlayerCnt
		}
		t.players[collector].AddCollected(t.round.trickCards)
		t.round.trickCards = make([]*Card, 0)
	}

	var scores [2]int
	if oneTwo {
		winTeam := TichuTeamOf(t.round.finishOrder[0])
		scores[winTeam] += 200
	} else {
		t.settleLastPlayer()
		for i := 0; i < TichuPlayerCnt; i++ {
			scores[TichuTeamOf(i)] += TichuCardsPoints(t.players[i].GetCollected())
		}
	}

	// ティチュー宣言ボーナス/ペナルティ
	for i := 0; i < TichuPlayerCnt; i++ {
		decl := t.players[i].GetDeclType()
		if decl == TichuDeclNone {
			continue
		}
		bonus := 100
		if decl == TichuDeclGrand {
			bonus = 200
		}
		if t.players[i].GetRank() == 1 {
			scores[TichuTeamOf(i)] += bonus
		} else {
			scores[TichuTeamOf(i)] -= bonus
		}
	}

	t.round.scores = scores
	t.appendLog(-1, "end", "deal over", nil)
}

// settleLastPlayer 最後に残ったプレイヤーの手札とトリックを精算する
func (t *Tichu) settleLastPlayer() {
	last := -1
	for i := 0; i < TichuPlayerCnt; i++ {
		if t.players[i].GetCardsSize() > 0 {
			last = i
			break
		}
	}
	if last < 0 {
		return
	}
	t.markFinished(last)
	firstOut := t.round.finishOrder[0]
	// 取ったトリックは最初に上がった人へ
	t.players[firstOut].AddCollected(t.players[last].GetCollected())
	t.players[last].collected = make([]*Card, 0)
	// 残り手札は相手チームへ
	opp := (last + 1) % TichuPlayerCnt
	hand := t.players[last].RemoveCards(t.allIndices(last))
	t.players[opp].AddCollected(hand)
}

// allIndices プレイヤーの全カードインデックス
func (t *Tichu) allIndices(idx int) []int {
	n := t.players[idx].GetCardsSize()
	res := make([]int, n)
	for i := range res {
		res[i] = i
	}
	return res
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行する
func (t *Tichu) CpuPlay() {
	if t.round.gameEndFlag || t.players[t.round.currentTurn].GetIsHuman() {
		return
	}
	if t.round.phase == TichuPhaseDeclare {
		t.executeDeclare(t.cpuDeclare(t.round.currentTurn))
		return
	}

	idx := t.round.currentTurn
	player := t.players[idx]
	indices := t.cpuFindPlay(player)
	if len(indices) == 0 {
		action := &TichuCpuAction{PlayerIdx: idx, IsPass: true}
		t.round.cpuActions = append(t.round.cpuActions, action)
		t.appendLog(idx, "pass", "pass", nil)
		t.handlePass()
		return
	}

	selected := make([]*Card, len(indices))
	for i, ci := range indices {
		selected[i] = player.GetCard(ci)
	}
	combo := ClassifyTichu(selected)
	if combo == nil { // 安全側: 分類できなければパス扱い
		t.handlePass()
		return
	}
	if combo.Type == TichuComboSingle && combo.PhoenixSingle && t.round.tableCombo != nil &&
		t.round.tableCombo.Type == TichuComboSingle {
		combo.Rank = t.round.tableCombo.Rank
	}
	cards := player.RemoveCards(indices)
	combo.Cards = cards
	action := &TichuCpuAction{PlayerIdx: idx, PlayedCards: cards}
	t.round.cpuActions = append(t.round.cpuActions, action)
	t.appendLog(idx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	t.playCombo(idx, combo)
}

// --- Getters ---

// IsHumanTurn 現在の手番が人間か
func (t *Tichu) IsHumanTurn() bool { return t.players[t.round.currentTurn].GetIsHuman() }

// GetPhase 現在のフェーズ取得
func (t *Tichu) GetPhase() TichuPhase { return t.round.phase }

// GetCurrentTurn 現在の手番取得
func (t *Tichu) GetCurrentTurn() int { return t.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (t *Tichu) GetGameEndFlag() bool { return t.round.gameEndFlag }

// GetTableCombo 場の役取得
func (t *Tichu) GetTableCombo() *TichuCombo { return t.round.tableCombo }

// GetLastPlayIdx 最後にカードを出したプレイヤー取得
func (t *Tichu) GetLastPlayIdx() int { return t.round.lastPlayIdx }

// GetStartLeader 先手プレイヤー取得
func (t *Tichu) GetStartLeader() int { return t.round.startLeader }

// GetFinishOrder 上がり順取得
func (t *Tichu) GetFinishOrder() []int { return t.round.finishOrder }

// GetPlayer プレイヤー取得
func (t *Tichu) GetPlayer(idx int) *TichuPlayer {
	if idx < 0 || idx >= len(t.players) {
		return nil
	}
	return t.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (t *Tichu) GetPlayerCnt() int { return len(t.players) }

// GetScores チーム得点取得 (index 0 = {0,2}, 1 = {1,3})
func (t *Tichu) GetScores() [2]int { return t.round.scores }

// GetIsOneTwo ワンツー (ダブルビクトリー) かどうか
func (t *Tichu) GetIsOneTwo() bool { return t.round.oneTwo }

// GetBombCount ボム使用回数取得
func (t *Tichu) GetBombCount() int { return t.round.bombCount }

// GetCpuActions CPU行動履歴取得
func (t *Tichu) GetCpuActions() []*TichuCpuAction { return t.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (t *Tichu) GetHumanAction() *TichuCpuAction { return t.round.humanAction }

// GetConfig 設定取得
func (t *Tichu) GetConfig() TichuConfig { return t.config }

// SetConfig 設定変更
func (t *Tichu) SetConfig(config TichuConfig) { t.config = config }

// HasPendingAction ペンディングアクションがあるか (常にfalse)
func (t *Tichu) HasPendingAction() bool { return false }

// GetActionLog 棋譜取得
func (t *Tichu) GetActionLog() []*ActionLogEntry { return t.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (t *Tichu) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	t.round.actionLog = append(t.round.actionLog, &ActionLogEntry{
		TurnNumber: len(t.round.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- JSON Serialization ---

// tichuCpuActionJSON is the JSON wire format for TichuCpuAction.
type tichuCpuActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
	DeclType    int     `json:"dt"`
	IsPass      bool    `json:"ps"`
}

// MarshalJSON implements json.Marshaler.
func (a *TichuCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(tichuCpuActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
		DeclType:    a.DeclType,
		IsPass:      a.IsPass,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *TichuCpuAction) UnmarshalJSON(data []byte) error {
	var j tichuCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	a.DeclType = j.DeclType
	a.IsPass = j.IsPass
	return nil
}

// tichuComboJSON is the JSON wire format for TichuCombo.
type tichuComboJSON struct {
	Type          TichuComboType `json:"ty"`
	Cards         []*Card        `json:"cs"`
	Rank          int            `json:"rk"`
	Length        int            `json:"ln"`
	PhoenixSingle bool           `json:"px"`
}

// MarshalJSON implements json.Marshaler.
func (c *TichuCombo) MarshalJSON() ([]byte, error) {
	return json.Marshal(tichuComboJSON{
		Type:          c.Type,
		Cards:         c.Cards,
		Rank:          c.Rank,
		Length:        c.Length,
		PhoenixSingle: c.PhoenixSingle,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TichuCombo) UnmarshalJSON(data []byte) error {
	var j tichuComboJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.Type = j.Type
	c.Cards = j.Cards
	c.Rank = j.Rank
	c.Length = j.Length
	c.PhoenixSingle = j.PhoenixSingle
	return nil
}

// tichuJSON is the JSON wire format for Tichu (flattens tichuRoundState).
type tichuJSON struct {
	TrumpCards  *TrumpCards       `json:"tc"`
	Players     []*TichuPlayer    `json:"pl"`
	Config      TichuConfig       `json:"cf"`
	Phase       TichuPhase        `json:"ph"`
	CurrentTurn int               `json:"ct"`
	TableCombo  *TichuCombo       `json:"tb"`
	LastPlayIdx int               `json:"lp"`
	TrickCards  []*Card           `json:"tk"`
	PassCount   int               `json:"pc"`
	DeclCount   int               `json:"dc"`
	DragonOnTop bool              `json:"do"`
	GameEndFlag bool              `json:"ge"`
	OneTwo      bool              `json:"ot"`
	StartLeader int               `json:"sl"`
	FinishOrder []int             `json:"fo"`
	Scores      [2]int            `json:"sc"`
	BombCount   int               `json:"bo"`
	CpuActions  []*TichuCpuAction `json:"ca"`
	HumanAction *TichuCpuAction   `json:"ha"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

const tichuMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (t *Tichu) MarshalJSON() ([]byte, error) {
	return json.Marshal(tichuJSON{
		TrumpCards:  t.trumpCards,
		Players:     t.players,
		Config:      t.config,
		Phase:       t.round.phase,
		CurrentTurn: t.round.currentTurn,
		TableCombo:  t.round.tableCombo,
		LastPlayIdx: t.round.lastPlayIdx,
		TrickCards:  t.round.trickCards,
		PassCount:   t.round.passCount,
		DeclCount:   t.round.declCount,
		DragonOnTop: t.round.dragonOnTop,
		GameEndFlag: t.round.gameEndFlag,
		OneTwo:      t.round.oneTwo,
		StartLeader: t.round.startLeader,
		FinishOrder: t.round.finishOrder,
		Scores:      t.round.scores,
		BombCount:   t.round.bombCount,
		CpuActions:  t.round.cpuActions,
		HumanAction: t.round.humanAction,
		ActionLog:   t.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Tichu) UnmarshalJSON(data []byte) error {
	var j tichuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tichuMaxSliceLen || len(j.TrickCards) > tichuMaxSliceLen ||
		len(j.CpuActions) > tichuMaxSliceLen || len(j.ActionLog) > tichuMaxSliceLen {
		return fmt.Errorf("tichu: input array exceeds maximum allowed size")
	}
	if len(j.Players) != TichuPlayerCnt {
		return fmt.Errorf("tichu: invalid player count %d, want %d", len(j.Players), TichuPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("tichu: nil player in input")
		}
	}
	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		// Tichu requires the four special cards (jokers); a fresh deck must
		// carry them so a post-restore Reset deals a correct 56-card deck.
		t.trumpCards = NewTrumpCards(TichuJokerCount)
	}
	t.players = j.Players
	if t.players == nil {
		t.players = make([]*TichuPlayer, 0)
	}
	t.config = j.Config
	t.round = tichuRoundState{
		phase:       j.Phase,
		currentTurn: j.CurrentTurn,
		tableCombo:  j.TableCombo,
		lastPlayIdx: j.LastPlayIdx,
		trickCards:  j.TrickCards,
		passCount:   j.PassCount,
		declCount:   j.DeclCount,
		dragonOnTop: j.DragonOnTop,
		gameEndFlag: j.GameEndFlag,
		oneTwo:      j.OneTwo,
		startLeader: j.StartLeader,
		finishOrder: j.FinishOrder,
		scores:      j.Scores,
		bombCount:   j.BombCount,
		cpuActions:  j.CpuActions,
		humanAction: j.HumanAction,
		actionLog:   j.ActionLog,
	}
	if t.round.actionLog == nil {
		t.round.actionLog = make([]*ActionLogEntry, 0)
	}
	if t.round.trickCards == nil {
		t.round.trickCards = make([]*Card, 0)
	}
	if t.round.finishOrder == nil {
		t.round.finishOrder = make([]int, 0)
	}
	return nil
}
