// Package domain トゥルコ (Truco) のドメインモデル。
//
// Truco は南米・スペインで人気のトリックテイキングとブラフ (ベッティング) を
// 融合したゲーム。本実装は 2 人対戦 (人間 1 + CPU 1) のみを扱う。
//
// 40 枚デッキ (8,9,10 を除く / NewTrumpCardsBriscola を流用) を使い、各マノ
// (hand) で 3 枚ずつ配って 3 バサ (baza / trick) を行い、2 バサ先取した側が
// そのマノの勝者となる (マストフォローなし)。同強カード同士のバサは「パルダ
// (parda / 引き分け)」となり、標準ルールでマノの勝者を決める。
//
// 最大の特徴は「Truco」のベッティングで、プレイ中に Truco→Retruco→Vale Cuatro
// と賭け点を引き上げられる。相手は受諾 (Quiero) / 拒否 (No Quiero = 直前点で
// 即敗北) / 再引き上げ を選択する。マッチは設定点 (既定 15) 先取で勝利する。
//
// カードの強さ (matadores) は Truco 固有で、1-Espadas(♠) > 1-Bastos(♣) >
// 7-Espadas(♠) > 7-Oros(♦) > 3 > 2 > 1(偽) > K > Q > J > 7(偽) > 6 > 5 > 4。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// TrucoPlayerCnt トゥルコのプレイヤー数 (v1は2人固定)
const TrucoPlayerCnt = 2

// TrucoHandSize 各プレイヤーに配る手札枚数
const TrucoHandSize = 3

// TrucoTricksPerHand 1マノあたりのバサ (トリック) 数
const TrucoTricksPerHand = 3

// TrucoPhase ゲームフェーズ
type TrucoPhase int

// Trucoのフェーズ定数
const (
	// TrucoPhasePlay カードプレイ (または Truco 宣言) フェーズ
	TrucoPhasePlay TrucoPhase = iota
	// TrucoPhaseRespond Truco 宣言への応答待ちフェーズ
	TrucoPhaseRespond
	// TrucoPhaseTrickEnd バサ (トリック) 終了フェーズ
	TrucoPhaseTrickEnd
	// TrucoPhaseHandEnd マノ (hand) 終了フェーズ
	TrucoPhaseHandEnd
	// TrucoPhaseGameEnd マッチ終了フェーズ
	TrucoPhaseGameEnd
)

// Truco のベッティングレベル定数。賭け点 = レベル + 1。
const (
	// TrucoLevelNone 宣言なし (1点)
	TrucoLevelNone = 0
	// TrucoLevelTruco Truco 宣言 (2点)
	TrucoLevelTruco = 1
	// TrucoLevelRetruco Retruco 宣言 (3点)
	TrucoLevelRetruco = 2
	// TrucoLevelValeCuatro Vale Cuatro 宣言 (4点)
	TrucoLevelValeCuatro = 3
	// TrucoMaxLevel 引き上げ可能な最大レベル
	TrucoMaxLevel = TrucoLevelValeCuatro
)

// TrucoHint ヒント情報。Action は "play" / "call" / "accept" / "decline"。
type TrucoHint struct {
	Action    string // 推奨アクション種別
	CardIndex *int   // 推奨カードインデックス (Action=="play" のとき)
	Reason    string // ヒント理由キー
}

// TrucoLevelValue ベッティングレベルに対応する賭け点を返す。
func TrucoLevelValue(level int) int {
	if level < TrucoLevelNone {
		return 1
	}
	if level > TrucoMaxLevel {
		return TrucoMaxLevel + 1
	}
	return level + 1
}

// TrucoCardStrength カードの Truco 強度を返す (大きいほど強い)。
// 上位4枚 (matadores) はスート固有。それ以外は数値の階層で決まる。
func TrucoCardStrength(c *Card) int {
	if c == nil {
		return 0
	}
	v, d := c.GetValue(), c.GetDesign()
	switch {
	case v == 1 && d == CardDesignSpade: // 1 de Espadas
		return 14
	case v == 1 && d == CardDesignClover: // 1 de Bastos
		return 13
	case v == 7 && d == CardDesignSpade: // 7 de Espadas
		return 12
	case v == 7 && d == CardDesignDiamond: // 7 de Oros
		return 11
	}
	switch v {
	case 3:
		return 10
	case 2:
		return 9
	case 1: // 偽のエース (Copas/Oros)
		return 8
	case 13: // Rey
		return 7
	case 12: // Caballo
		return 6
	case 11: // Sota
		return 5
	case 7: // 偽の7 (Copas/Bastos)
		return 4
	case 6:
		return 3
	case 5:
		return 2
	case 4:
		return 1
	}
	return 0
}

// Truco トゥルコゲームクラス
type Truco struct {
	trumpCards        *TrumpCards
	players           []*TrucoPlayer
	config            TrucoConfig
	phase             TrucoPhase
	handNumber        int
	trickNumber       int
	currentPlayerIdx  int
	responderIdx      int // Respond フェーズで応答すべきプレイヤー、それ以外は -1
	currentTrick      []*TrickCard
	trickResults      []int // 完了したバサの勝者 (0/1) または -1 (パルダ)
	leadPlayerIdx     int
	manoIdx           int // 親 (elder hand = 非ディーラー)。全パルダ時のタイブレーク
	dealerIdx         int
	handStake         int // 現在の確定賭け点 (1..4)
	acceptedLevel     int // 受諾済みベッティングレベル (0..3)
	pendingLevel      int // 応答待ちで提示中のレベル (Respond 中のみ > 0)
	trucoCallerIdx    int // 応答待ちの宣言者、それ以外は -1
	matchTarget       int
	playerMatchPoints []int
	handWinnerIdx     int // 直近マノの勝者 (-1: 未確定)
	gameEndFlag       bool
	winnerIdx         int // マッチ勝者 (-1: 未確定)
	actionLogBase
}

// NewTruco コンストラクタ
func NewTruco(trumpCards *TrumpCards, players []*TrucoPlayer, config TrucoConfig) *Truco {
	return &Truco{
		trumpCards:        trumpCards,
		players:           players,
		config:            config,
		responderIdx:      -1,
		trucoCallerIdx:    -1,
		winnerIdx:         -1,
		handWinnerIdx:     -1,
		matchTarget:       config.normalized().MatchTarget,
		playerMatchPoints: make([]int, len(players)),
	}
}

// NewDefaultTruco 標準の 2 人対戦セットアップを返す (人間 idx 0 + CPU idx 1)。
func NewDefaultTruco() *Truco {
	players := []*TrucoPlayer{
		NewTrucoPlayer(true),
		NewTrucoPlayer(false),
	}
	return NewTruco(NewTrumpCardsBriscola(), players, DefaultTrucoConfig())
}

// Reset マッチを初期化する
func (t *Truco) Reset() {
	cfg := t.config.normalized()
	t.config = cfg
	t.matchTarget = cfg.MatchTarget
	t.gameEndFlag = false
	t.winnerIdx = -1
	t.handNumber = 1
	t.dealerIdx = 0
	t.manoIdx = 1 // 非ディーラーが親 (先手)
	t.playerMatchPoints = make([]int, len(t.players))
	t.actionLog = nil
	t.dealHand()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (t *Truco) PlayerPlay(cardIndex int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TrucoPhasePlay {
		return ErrWrongPhase
	}
	if !t.players[t.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := t.players[t.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	played := player.RemoveCard(cardIndex)
	t.playCard(t.currentPlayerIdx, played)
	return nil
}

// DeclareTruco 人間プレイヤーが Truco を宣言 (または Retruco/Vale Cuatro へ再引き上げ) する。
func (t *Truco) DeclareTruco() error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	actor, ok := t.declareActor()
	if !ok || actor != 0 {
		return ErrNotHumanTurn
	}
	if !t.canDeclare(actor) {
		return NewDomainError(ErrWrongPhase, "これ以上引き上げできません")
	}
	t.callTruco(actor)
	return nil
}

// RespondTruco 人間プレイヤーが Truco 宣言に応答する (true=受諾 / false=拒否)。
func (t *Truco) RespondTruco(accept bool) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TrucoPhaseRespond {
		return ErrWrongPhase
	}
	if t.responderIdx != 0 {
		return ErrNotHumanTurn
	}
	t.respond(0, accept)
	return nil
}

// Next バサ終了 / マノ終了フェーズから次の状態へ進める。
func (t *Truco) Next() {
	switch t.phase {
	case TrucoPhaseTrickEnd:
		t.advanceTrick()
	case TrucoPhaseHandEnd:
		t.advanceHand()
	default:
		// 進行不要なフェーズ
	}
}

// CpuStep 現在の手番が CPU の場合に 1 アクションを実行する。
func (t *Truco) CpuStep() {
	if t.gameEndFlag {
		return
	}
	switch t.phase {
	case TrucoPhasePlay:
		if t.players[t.currentPlayerIdx].GetIsHuman() {
			return
		}
		t.cpuActPlay(t.currentPlayerIdx)
	case TrucoPhaseRespond:
		if t.responderIdx < 0 || t.players[t.responderIdx].GetIsHuman() {
			return
		}
		t.cpuActRespond(t.responderIdx)
	default:
		// CPU の行動なし
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (t *Truco) GetPhase() TrucoPhase { return t.phase }

// SetPhase フェーズ設定 (テスト用)
func (t *Truco) SetPhase(p TrucoPhase) { t.phase = p }

// GetHandNumber 現在のマノ番号取得
func (t *Truco) GetHandNumber() int { return t.handNumber }

// GetTrickNumber 現在のバサ番号取得
func (t *Truco) GetTrickNumber() int { return t.trickNumber }

// SetTrickNumber バサ番号設定 (テスト用)
func (t *Truco) SetTrickNumber(n int) { t.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (t *Truco) GetCurrentPlayerIdx() int { return t.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (t *Truco) SetCurrentPlayerIdx(idx int) { t.currentPlayerIdx = idx }

// GetResponderIdx 応答すべきプレイヤーインデックス取得 (-1: 応答待ちでない)
func (t *Truco) GetResponderIdx() int { return t.responderIdx }

// SetResponderIdx 応答プレイヤー設定 (テスト用)
func (t *Truco) SetResponderIdx(idx int) { t.responderIdx = idx }

// GetCurrentTrick 現在のバサ取得
func (t *Truco) GetCurrentTrick() []*TrickCard { return t.currentTrick }

// SetCurrentTrick バサ設定 (テスト用)
func (t *Truco) SetCurrentTrick(trick []*TrickCard) { t.currentTrick = trick }

// GetTrickResults 当該マノで完了したバサの勝者リスト取得 (0/1 または -1=パルダ)
func (t *Truco) GetTrickResults() []int { return t.trickResults }

// SetTrickResults バサ結果設定 (テスト用)
func (t *Truco) SetTrickResults(r []int) { t.trickResults = r }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (t *Truco) GetLeadPlayerIdx() int { return t.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤー設定 (テスト用)
func (t *Truco) SetLeadPlayerIdx(idx int) { t.leadPlayerIdx = idx }

// GetManoIdx 親 (elder hand) インデックス取得
func (t *Truco) GetManoIdx() int { return t.manoIdx }

// SetManoIdx 親インデックス設定 (テスト用)
func (t *Truco) SetManoIdx(idx int) { t.manoIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (t *Truco) GetDealerIdx() int { return t.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (t *Truco) SetDealerIdx(idx int) { t.dealerIdx = idx }

// GetHandStake 現在の確定賭け点取得
func (t *Truco) GetHandStake() int { return t.handStake }

// SetHandStake 確定賭け点設定 (テスト用)
func (t *Truco) SetHandStake(v int) { t.handStake = v }

// GetAcceptedLevel 受諾済みベッティングレベル取得
func (t *Truco) GetAcceptedLevel() int { return t.acceptedLevel }

// SetAcceptedLevel 受諾済みレベル設定 (テスト用)
func (t *Truco) SetAcceptedLevel(l int) { t.acceptedLevel = l }

// GetPendingLevel 応答待ち提示レベル取得 (0: 応答待ちでない)
func (t *Truco) GetPendingLevel() int { return t.pendingLevel }

// SetPendingLevel 応答待ちレベル設定 (テスト用)
func (t *Truco) SetPendingLevel(l int) { t.pendingLevel = l }

// GetTrucoCallerIdx 応答待ちの宣言者インデックス取得 (-1: なし)
func (t *Truco) GetTrucoCallerIdx() int { return t.trucoCallerIdx }

// SetTrucoCallerIdx 宣言者インデックス設定 (テスト用)
func (t *Truco) SetTrucoCallerIdx(idx int) { t.trucoCallerIdx = idx }

// GetMatchTarget マッチ目標点取得
func (t *Truco) GetMatchTarget() int { return t.matchTarget }

// GetPlayerMatchPoints プレイヤーのマッチ累積点取得
func (t *Truco) GetPlayerMatchPoints(i int) int {
	if i < 0 || i >= len(t.playerMatchPoints) {
		return 0
	}
	return t.playerMatchPoints[i]
}

// SetPlayerMatchPoints プレイヤーのマッチ累積点設定 (テスト用)
func (t *Truco) SetPlayerMatchPoints(i, points int) {
	if i >= 0 && i < len(t.playerMatchPoints) {
		t.playerMatchPoints[i] = points
	}
}

// GetHandWinnerIdx 直近マノの勝者取得 (-1: 未確定)
func (t *Truco) GetHandWinnerIdx() int { return t.handWinnerIdx }

// SetHandWinnerIdx マノ勝者設定 (テスト用)
func (t *Truco) SetHandWinnerIdx(idx int) { t.handWinnerIdx = idx }

// GetGameEndFlag マッチ終了フラグ取得
func (t *Truco) GetGameEndFlag() bool { return t.gameEndFlag }

// SetGameEndFlag マッチ終了フラグ設定 (テスト用)
func (t *Truco) SetGameEndFlag(flag bool) { t.gameEndFlag = flag }

// GetWinnerIdx マッチ勝者インデックス取得 (-1: 未確定)
func (t *Truco) GetWinnerIdx() int { return t.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (t *Truco) GetPlayerCnt() int { return len(t.players) }

// GetPlayer プレイヤー取得
func (t *Truco) GetPlayer(i int) *TrucoPlayer {
	if i < 0 || i >= len(t.players) {
		return nil
	}
	return t.players[i]
}

// GetConfig 設定取得
func (t *Truco) GetConfig() TrucoConfig { return t.config }

// SetConfig 設定変更
func (t *Truco) SetConfig(cfg TrucoConfig) { t.config = cfg }

// IsHumanTurn 現在の手番 (プレイまたは応答) が人間かどうか
func (t *Truco) IsHumanTurn() bool {
	switch t.phase {
	case TrucoPhasePlay:
		return t.idxIsHuman(t.currentPlayerIdx)
	case TrucoPhaseRespond:
		return t.idxIsHuman(t.responderIdx)
	default:
		return false
	}
}

// CanDeclareTruco 現在の手番の人間が Truco 宣言 (または再引き上げ) 可能かを返す。
func (t *Truco) CanDeclareTruco() bool {
	actor, ok := t.declareActor()
	if !ok || actor != 0 {
		return false
	}
	return t.canDeclare(actor)
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// Truco には must-follow 制約がないため、現在の手札全てが対象。
func (t *Truco) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(t.players) {
		return nil
	}
	p := t.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// GetHint 人間プレイヤー (idx 0) へのヒントを取得する。
func (t *Truco) GetHint() *TrucoHint {
	switch t.phase {
	case TrucoPhasePlay:
		if t.currentPlayerIdx != 0 || t.players[0].GetCardsSize() == 0 {
			return nil
		}
		idx := t.cpuSelectPlayCard(0)
		if t.canDeclare(0) && t.handTopStrength(0) >= 11 {
			return &TrucoHint{Action: "call", CardIndex: &idx, Reason: "hintCall"}
		}
		return &TrucoHint{Action: "play", CardIndex: &idx, Reason: t.playHintReason(0, idx)}
	case TrucoPhaseRespond:
		if t.responderIdx != 0 {
			return nil
		}
		if t.handTopStrength(0) >= 10 {
			return &TrucoHint{Action: "accept", Reason: "hintAccept"}
		}
		return &TrucoHint{Action: "decline", Reason: "hintDecline"}
	default:
		return nil
	}
}

// --- Private: turn / declaration helpers ---

func (t *Truco) idxIsHuman(idx int) bool {
	if idx < 0 || idx >= len(t.players) {
		return false
	}
	return t.players[idx].GetIsHuman()
}

// declareActor 現在 Truco 宣言できる手番のプレイヤーを返す。
func (t *Truco) declareActor() (int, bool) {
	switch t.phase {
	case TrucoPhasePlay:
		return t.currentPlayerIdx, true
	case TrucoPhaseRespond:
		return t.responderIdx, true
	default:
		return -1, false
	}
}

// canDeclare actor が (さらに) 引き上げ可能かを判定する。
func (t *Truco) canDeclare(actor int) bool {
	if actor < 0 {
		return false
	}
	switch t.phase {
	case TrucoPhasePlay:
		return t.pendingLevel == 0 && t.acceptedLevel < TrucoMaxLevel
	case TrucoPhaseRespond:
		return t.pendingLevel < TrucoMaxLevel
	default:
		return false
	}
}

// callTruco caller が宣言/再引き上げを行う。
func (t *Truco) callTruco(caller int) {
	if t.phase == TrucoPhaseRespond {
		t.pendingLevel++
	} else {
		t.pendingLevel = t.acceptedLevel + 1
	}
	t.trucoCallerIdx = caller
	t.responderIdx = 1 - caller
	t.phase = TrucoPhaseRespond
	t.appendLog(caller, "truco",
		fmt.Sprintf("%s calls %s", t.playerName(caller), trucoLevelName(t.pendingLevel)), nil)
}

// respond responder が宣言に応答する。
func (t *Truco) respond(responder int, accept bool) {
	if accept {
		t.acceptedLevel = t.pendingLevel
		t.handStake = TrucoLevelValue(t.acceptedLevel)
		t.pendingLevel = 0
		t.trucoCallerIdx = -1
		t.responderIdx = -1
		t.phase = TrucoPhasePlay
		t.appendLog(responder, "accept",
			fmt.Sprintf("%s accepts (stake %d)", t.playerName(responder), t.handStake), nil)
		return
	}
	// 拒否: 宣言者が直前の確定点でマノを取る
	caller := t.trucoCallerIdx
	t.pendingLevel = 0
	t.responderIdx = -1
	t.trucoCallerIdx = -1
	t.handWinnerIdx = caller
	t.phase = TrucoPhaseHandEnd
	t.appendLog(responder, "decline",
		fmt.Sprintf("%s declines; %s wins hand (%d pt)",
			t.playerName(responder), t.playerName(caller), t.handStake), nil)
}

// --- Private: trick / hand progression ---

// playCard カードをプレイする共通処理。
func (t *Truco) playCard(playerIdx int, card *Card) {
	t.currentTrick = append(t.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	t.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", t.playerName(playerIdx), cardStr(card)), []*Card{card})
	if len(t.currentTrick) == TrucoPlayerCnt {
		t.finishBaza()
	} else {
		t.currentPlayerIdx = 1 - t.currentPlayerIdx
	}
}

// finishBaza バサ完了処理。勝者を記録し TrickEnd フェーズへ移す。
func (t *Truco) finishBaza() {
	result := trucoBazaWinner(t.currentTrick)
	t.trickResults = append(t.trickResults, result)
	cards := []*Card{t.currentTrick[0].Card, t.currentTrick[1].Card}
	if result < 0 {
		t.appendLog(-1, "baza_tie", fmt.Sprintf("Baza %d: parda", t.trickNumber), cards)
		// パルダはリードプレイヤーが維持
	} else {
		t.leadPlayerIdx = result
		t.appendLog(result, "baza_win",
			fmt.Sprintf("%s wins baza %d", t.playerName(result), t.trickNumber), cards)
	}
	t.phase = TrucoPhaseTrickEnd
}

// advanceTrick TrickEnd から次のバサ、またはマノ終了へ進める。
func (t *Truco) advanceTrick() {
	decided, winner := t.manoResult()
	if decided {
		t.handWinnerIdx = winner
		t.phase = TrucoPhaseHandEnd
		return
	}
	t.trickNumber++
	t.currentTrick = nil
	t.currentPlayerIdx = t.leadPlayerIdx
	t.phase = TrucoPhasePlay
}

// advanceHand HandEnd からマノの点を加算し、次マノ配り直しまたはマッチ終了へ進める。
func (t *Truco) advanceHand() {
	if t.handWinnerIdx >= 0 && t.handWinnerIdx < len(t.playerMatchPoints) {
		t.playerMatchPoints[t.handWinnerIdx] += t.handStake
	}
	t.appendLog(t.handWinnerIdx, "hand_end",
		fmt.Sprintf("Hand %d: %s +%d (match %d-%d)", t.handNumber,
			t.playerName(t.handWinnerIdx), t.handStake,
			t.playerMatchPoints[0], t.playerMatchPoints[1]), nil)

	if t.playerMatchPoints[t.handWinnerIdx] >= t.matchTarget {
		t.gameEndFlag = true
		t.winnerIdx = t.handWinnerIdx
		t.phase = TrucoPhaseGameEnd
		t.appendLog(-1, "game_end",
			fmt.Sprintf("Match end: %d-%d", t.playerMatchPoints[0], t.playerMatchPoints[1]), nil)
		return
	}
	t.dealerIdx = 1 - t.dealerIdx
	t.manoIdx = 1 - t.manoIdx
	t.handNumber++
	t.dealHand()
}

// dealHand 現在の dealerIdx/manoIdx で新しいマノを配る。
func (t *Truco) dealHand() {
	t.trickNumber = 1
	t.currentTrick = nil
	t.trickResults = nil
	t.handStake = 1
	t.acceptedLevel = TrucoLevelNone
	t.pendingLevel = 0
	t.trucoCallerIdx = -1
	t.responderIdx = -1
	t.handWinnerIdx = -1

	for _, p := range t.players {
		p.ResetGame()
	}
	t.trumpCards.Shuffle()
	for range TrucoHandSize {
		for i := range TrucoPlayerCnt {
			player := t.players[(t.manoIdx+i)%TrucoPlayerCnt]
			if c := t.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	t.sortAllHands()

	t.leadPlayerIdx = t.manoIdx
	t.currentPlayerIdx = t.manoIdx
	t.phase = TrucoPhasePlay
	t.appendLog(-1, "deal", fmt.Sprintf("Hand %d dealt (dealer=%s)",
		t.handNumber, t.playerName(t.dealerIdx)), nil)
}

// --- Pure rule helpers ---

// trucoBazaWinner バサの勝者プレイヤーインデックスを返す。同強なら -1 (パルダ)。
func trucoBazaWinner(trick []*TrickCard) int {
	if len(trick) < TrucoPlayerCnt {
		return -1
	}
	s0 := TrucoCardStrength(trick[0].Card)
	s1 := TrucoCardStrength(trick[1].Card)
	switch {
	case s0 > s1:
		return trick[0].PlayerIdx
	case s1 > s0:
		return trick[1].PlayerIdx
	default:
		return -1
	}
}

// manoResult 現在の trickResults からマノの勝者を判定する。
// decided=false の場合は次バサが必要。標準のアルゼンチン式タイブレークに従う。
func (t *Truco) manoResult() (bool, int) {
	return resolveMano(t.trickResults, t.manoIdx)
}

// resolveMano バサ結果列からマノの勝者を決定する純関数。
// results の各要素は 0/1 (勝者) または -1 (パルダ)。manoIdx は親 (全パルダ時の勝者)。
func resolveMano(results []int, manoIdx int) (bool, int) {
	w0, w1 := 0, 0
	for _, r := range results {
		switch r {
		case 0:
			w0++
		case 1:
			w1++
		}
	}
	if w0 >= 2 {
		return true, 0
	}
	if w1 >= 2 {
		return true, 1
	}
	switch len(results) {
	case 0, 1:
		return false, -1
	case 2:
		return decideAfterTwo(results)
	default:
		return true, decideAfterThree(results, manoIdx)
	}
}

// decideAfterTwo 2バサ終了時点でマノが確定するかを判定する。
func decideAfterTwo(r []int) (bool, int) {
	a, b := r[0], r[1]
	switch {
	case a != -1 && b == -1: // 1勝目を取り2バサ目パルダ → 1勝目の勝者
		return true, a
	case a == -1 && b != -1: // 1バサ目パルダで2バサ目を取った側
		return true, b
	default: // 1-1 で割れた or 両パルダ → 3バサ目へ
		return false, -1
	}
}

// decideAfterThree 3バサ終了時点でマノの勝者を決定する。
func decideAfterThree(r []int, manoIdx int) int {
	a, b, c := r[0], r[1], r[2]
	if a != -1 && b != -1 && a != b {
		// 1-1 で割れた場合、3バサ目が決める。3バサ目パルダなら1勝目の勝者。
		if c != -1 {
			return c
		}
		return a
	}
	// 1・2バサ目とも (実質) パルダ。3バサ目が決め、全パルダなら親。
	if c != -1 {
		return c
	}
	return manoIdx
}

// --- Sorting / naming / logging ---

// sortAllHands 全プレイヤーの手札を強度昇順でソートする。
func (t *Truco) sortAllHands() {
	for _, p := range t.players {
		cards := make([]*Card, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards[i] = p.GetCard(i)
		}
		sort.Slice(cards, func(i, j int) bool {
			return TrucoCardStrength(cards[i]) < TrucoCardStrength(cards[j])
		})
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

// playerName プレイヤー名を返す (ログ用)。
func (t *Truco) playerName(idx int) string {
	if idx < 0 || idx >= len(t.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if t.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// trucoLevelName ベッティングレベルの表示名を返す。
func trucoLevelName(level int) string {
	switch level {
	case TrucoLevelTruco:
		return "Truco"
	case TrucoLevelRetruco:
		return "Retruco"
	case TrucoLevelValeCuatro:
		return "Vale Cuatro"
	default:
		return "Truco"
	}
}

// playHintReason プレイ推奨の理由キーを判定する。
func (t *Truco) playHintReason(playerIdx, chosenIdx int) string {
	card := t.players[playerIdx].GetCard(chosenIdx)
	if len(t.currentTrick) == 0 {
		if TrucoCardStrength(card) >= 11 {
			return "leadStrong"
		}
		return "leadLow"
	}
	opp := t.currentTrick[0].Card
	if TrucoCardStrength(card) > TrucoCardStrength(opp) {
		return "followWin"
	}
	return "followDump"
}

// --- CPU AI (single-difficulty heuristic) ---

// cpuActPlay CPU のプレイフェーズ行動: 宣言するか、カードを出す。
func (t *Truco) cpuActPlay(idx int) {
	if t.canDeclare(idx) && t.cpuWantsToCall(idx) {
		t.callTruco(idx)
		return
	}
	cardIdx := t.cpuSelectPlayCard(idx)
	played := t.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	t.playCard(idx, played)
}

// cpuActRespond CPU の応答フェーズ行動: 再引き上げ / 受諾 / 拒否。
func (t *Truco) cpuActRespond(idx int) {
	switch t.cpuRespondDecision(idx) {
	case "raise":
		t.callTruco(idx)
	case "accept":
		t.respond(idx, true)
	default:
		t.respond(idx, false)
	}
}

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する。
func (t *Truco) cpuSelectPlayCard(idx int) int {
	player := t.players[idx]
	n := player.GetCardsSize()
	if n <= 1 {
		return 0
	}
	if len(t.currentTrick) == 0 {
		return t.cpuWeakestIdx(idx)
	}
	oppStrength := TrucoCardStrength(t.currentTrick[0].Card)
	bestIdx, bestStrength := -1, 0
	for i := 0; i < n; i++ {
		s := TrucoCardStrength(player.GetCard(i))
		if s > oppStrength && (bestIdx < 0 || s < bestStrength) {
			bestIdx = i
			bestStrength = s
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return t.cpuWeakestIdx(idx)
}

// cpuWeakestIdx 最も弱いカードのインデックスを返す。
func (t *Truco) cpuWeakestIdx(idx int) int {
	player := t.players[idx]
	bestIdx, bestStrength := 0, TrucoCardStrength(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		s := TrucoCardStrength(player.GetCard(i))
		if s < bestStrength {
			bestStrength = s
			bestIdx = i
		}
	}
	return bestIdx
}

// handTopStrength 手札中の最強カードの強度を返す。
func (t *Truco) handTopStrength(idx int) int {
	player := t.players[idx]
	top := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if s := TrucoCardStrength(player.GetCard(i)); s > top {
			top = s
		}
	}
	return top
}

// cpuWantsToCall CPU が Truco を宣言したいかを判定する (強い手は積極的、稀にブラフ)。
func (t *Truco) cpuWantsToCall(idx int) bool {
	top := t.handTopStrength(idx)
	r := rand.Float64()
	switch {
	case top >= 11:
		return r < 0.7
	case top >= 9:
		return r < 0.3
	default:
		return r < 0.07 // ブラフ
	}
}

// cpuRespondDecision CPU の応答を決定する ("raise" / "accept" / "decline")。
func (t *Truco) cpuRespondDecision(idx int) string {
	top := t.handTopStrength(idx)
	r := rand.Float64()
	if t.pendingLevel < TrucoMaxLevel && top >= 12 && r < 0.4 {
		return "raise"
	}
	if top >= 10 {
		return "accept"
	}
	if top >= 7 {
		if r < 0.75 {
			return "accept"
		}
		return "decline"
	}
	if r < 0.2 { // 弱い手でも稀に受諾 (ブラフ)
		return "accept"
	}
	return "decline"
}

// --- JSON ---

// trucoJSON is the JSON wire format for Truco.
type trucoJSON struct {
	TrumpCards        *TrumpCards       `json:"tc"`
	Players           []*TrucoPlayer    `json:"ps"`
	Config            TrucoConfig       `json:"cf"`
	Phase             TrucoPhase        `json:"ph"`
	HandNumber        int               `json:"hn"`
	TrickNumber       int               `json:"tn"`
	CurrentPlayerIdx  int               `json:"ci"`
	ResponderIdx      int               `json:"ri"`
	CurrentTrick      []*TrickCard      `json:"ct"`
	TrickResults      []int             `json:"tr"`
	LeadPlayerIdx     int               `json:"li"`
	ManoIdx           int               `json:"mi"`
	DealerIdx         int               `json:"di"`
	HandStake         int               `json:"hs"`
	AcceptedLevel     int               `json:"al"`
	PendingLevel      int               `json:"pl"`
	TrucoCallerIdx    int               `json:"tk"`
	MatchTarget       int               `json:"mt"`
	PlayerMatchPoints []int             `json:"pp"`
	HandWinnerIdx     int               `json:"hw"`
	GameEndFlag       bool              `json:"ge"`
	WinnerIdx         int               `json:"wi"`
	ActionLog         []*ActionLogEntry `json:"lg"`
}

// MarshalJSON implements json.Marshaler.
func (t *Truco) MarshalJSON() ([]byte, error) {
	return json.Marshal(trucoJSON{
		TrumpCards:        t.trumpCards,
		Players:           t.players,
		Config:            t.config,
		Phase:             t.phase,
		HandNumber:        t.handNumber,
		TrickNumber:       t.trickNumber,
		CurrentPlayerIdx:  t.currentPlayerIdx,
		ResponderIdx:      t.responderIdx,
		CurrentTrick:      t.currentTrick,
		TrickResults:      t.trickResults,
		LeadPlayerIdx:     t.leadPlayerIdx,
		ManoIdx:           t.manoIdx,
		DealerIdx:         t.dealerIdx,
		HandStake:         t.handStake,
		AcceptedLevel:     t.acceptedLevel,
		PendingLevel:      t.pendingLevel,
		TrucoCallerIdx:    t.trucoCallerIdx,
		MatchTarget:       t.matchTarget,
		PlayerMatchPoints: t.playerMatchPoints,
		HandWinnerIdx:     t.handWinnerIdx,
		GameEndFlag:       t.gameEndFlag,
		WinnerIdx:         t.winnerIdx,
		ActionLog:         t.actionLog,
	})
}

// trucoMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const trucoMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler.
//
// Validates that the deserialised game state matches Truco's fixed shape
// (TrucoPlayerCnt players, at most TrucoPlayerCnt cards on the current baza,
// PlayerMatchPoints aligned to the player count, bounded TrickResults/ActionLog)
// to prevent DoS via crafted payloads and out-of-bounds access during play.
func (t *Truco) UnmarshalJSON(data []byte) error {
	var j trucoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != TrucoPlayerCnt {
		return fmt.Errorf("truco: expected %d players, got %d", TrucoPlayerCnt, len(j.Players))
	}
	if len(j.CurrentTrick) > TrucoPlayerCnt {
		return fmt.Errorf("truco: current trick has %d cards (max %d)", len(j.CurrentTrick), TrucoPlayerCnt)
	}
	if len(j.TrickResults) > TrucoTricksPerHand {
		return fmt.Errorf("truco: trick results has %d entries (max %d)", len(j.TrickResults), TrucoTricksPerHand)
	}
	if j.PlayerMatchPoints != nil && len(j.PlayerMatchPoints) != TrucoPlayerCnt {
		return fmt.Errorf("truco: expected %d player points entries, got %d", TrucoPlayerCnt, len(j.PlayerMatchPoints))
	}
	if len(j.ActionLog) > trucoMaxSliceLen {
		return fmt.Errorf("truco: action log exceeds maximum allowed size")
	}
	// Nil entries restored from untrusted JSON would panic on first
	// dereference (sortAllHands, trucoBazaWinner, finishBaza, ...).
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("truco: player at index %d is nil", i)
		}
	}
	for i, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("truco: trick card at index %d is nil", i)
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= TrucoPlayerCnt {
			return fmt.Errorf("truco: trick card player index %d out of bounds", tc.PlayerIdx)
		}
	}
	for i, e := range j.ActionLog {
		if e == nil {
			return fmt.Errorf("truco: action log entry at index %d is nil", i)
		}
	}
	// Index fields drive slice accesses (t.players[...]); out-of-range values
	// from a crafted payload would panic during play. -1 is the valid
	// "unset" sentinel for the responder/caller/winner indices.
	indexBounds := []struct {
		name string
		val  int
		min  int
	}{
		{"currentPlayerIdx", j.CurrentPlayerIdx, 0},
		{"leadPlayerIdx", j.LeadPlayerIdx, 0},
		{"manoIdx", j.ManoIdx, 0},
		{"dealerIdx", j.DealerIdx, 0},
		{"responderIdx", j.ResponderIdx, -1},
		{"trucoCallerIdx", j.TrucoCallerIdx, -1},
		{"handWinnerIdx", j.HandWinnerIdx, -1},
		{"winnerIdx", j.WinnerIdx, -1},
	}
	for _, b := range indexBounds {
		if b.val < b.min || b.val >= TrucoPlayerCnt {
			return fmt.Errorf("truco: %s %d out of bounds", b.name, b.val)
		}
	}
	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCardsBriscola()
	}
	t.players = j.Players
	t.config = j.Config
	t.phase = j.Phase
	t.handNumber = j.HandNumber
	t.trickNumber = j.TrickNumber
	t.currentPlayerIdx = j.CurrentPlayerIdx
	t.responderIdx = j.ResponderIdx
	t.currentTrick = j.CurrentTrick
	if t.currentTrick == nil {
		t.currentTrick = make([]*TrickCard, 0)
	}
	t.trickResults = j.TrickResults
	t.leadPlayerIdx = j.LeadPlayerIdx
	t.manoIdx = j.ManoIdx
	t.dealerIdx = j.DealerIdx
	// Stake/level fields are clamped (not rejected) to the same invariants as
	// TrucoConfig.normalized(): a zero-value handStake from a minimal payload
	// is normalised to the base stake rather than treated as corrupt.
	t.handStake = j.HandStake
	if t.handStake < 1 || t.handStake > TrucoMaxLevel+1 {
		t.handStake = 1
	}
	t.acceptedLevel = j.AcceptedLevel
	if t.acceptedLevel < 0 || t.acceptedLevel > TrucoMaxLevel {
		t.acceptedLevel = TrucoLevelNone
	}
	t.pendingLevel = j.PendingLevel
	if t.pendingLevel < 0 || t.pendingLevel > TrucoMaxLevel {
		t.pendingLevel = TrucoLevelNone
	}
	t.trucoCallerIdx = j.TrucoCallerIdx
	t.matchTarget = j.MatchTarget
	if t.matchTarget < TrucoMinMatchTarget || t.matchTarget > TrucoMaxMatchTarget {
		t.matchTarget = TrucoDefaultMatchTarget
	}
	t.playerMatchPoints = j.PlayerMatchPoints
	if t.playerMatchPoints == nil {
		t.playerMatchPoints = make([]int, TrucoPlayerCnt)
	}
	t.handWinnerIdx = j.HandWinnerIdx
	t.gameEndFlag = j.GameEndFlag
	t.winnerIdx = j.WinnerIdx
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
