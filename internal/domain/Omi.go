//go:build !js || !wasm || extra5

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// OmiPlayerCnt オミプレイヤー数
const OmiPlayerCnt = 4

// OmiHandSize 各プレイヤーの手札枚数 (32枚デッキを4人に配り切るため8枚)
const OmiHandSize = 8

// OmiInitialDealSize 切り札決定前の初期手札枚数 (各自4枚)
const OmiInitialDealSize = 4

// OmiTeamCnt チーム数
const OmiTeamCnt = 2

// OmiPhase ゲームフェーズ
type OmiPhase int

// Omiのフェーズ定数
// フェーズ遷移: CallTrump(0) → Play(1) → TrickEnd(2) → RoundEnd(3) → GameEnd(4)
const (
	// OmiPhaseCallTrump コールトランプフェーズ (指名者が切り札スートを宣言する)
	OmiPhaseCallTrump OmiPhase = 0
	// OmiPhasePlay トリックプレイフェーズ
	OmiPhasePlay OmiPhase = 1
	// OmiPhaseTrickEnd トリック終了フェーズ
	OmiPhaseTrickEnd OmiPhase = 2
	// OmiPhaseRoundEnd ラウンド終了フェーズ
	OmiPhaseRoundEnd OmiPhase = 3
	// OmiPhaseGameEnd ゲーム終了フェーズ
	OmiPhaseGameEnd OmiPhase = 4

	// OmiPhaseMin 最小フェーズ値
	OmiPhaseMin = OmiPhaseCallTrump
	// OmiPhaseMax 最大フェーズ値
	OmiPhaseMax = OmiPhaseGameEnd
)

// OmiHint ヒント情報
type OmiHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	Suit      *int   // 推奨スート (コールトランプ時)
	Reason    string // ヒント理由キー
}

// Omi オミゲームクラス
// スリランカの国民的トリックテイキングゲーム。
// 32枚デッキ (7-A) を用い、4人 (2対2のチーム戦) でプレイする。
type Omi struct {
	trumpCards       *TrumpCards
	players          []*OmiPlayer
	config           OmiConfig
	phase            OmiPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	trumpSuit        int // 切り札スート (CardDesignSpade等)
	trumpCallerIdx   int // 切り札を宣言した席番号
	makerTeam        int // 切り札を指名したチーム (0 or 1)
	teamScores       [OmiTeamCnt]int
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerTeam       int // 勝利チーム (-1 = 未確定)
	actionLogBase
}

// NewOmi コンストラクタ
func NewOmi(trumpCards *TrumpCards, players []*OmiPlayer, config OmiConfig) *Omi {
	return &Omi{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerTeam:     -1,
		roundNumber:    0,
		dealerIdx:      0,
		trumpCallerIdx: -1,
		makerTeam:      -1,
	}
}

// NewDefaultOmi は4人標準設定 (席0+2 vs 席1+3、席0が人間) の Omi を生成する。
// デッキには NewTrumpCards32() (7-A の32枚) を使用する。
func NewDefaultOmi() *Omi {
	players := []*OmiPlayer{
		NewOmiPlayer(true, 0),
		NewOmiPlayer(false, 1),
		NewOmiPlayer(false, 0),
		NewOmiPlayer(false, 1),
	}
	return NewOmi(NewTrumpCards32(), players, DefaultOmiConfig())
}

// Reset ゲーム初期化
func (e *Omi) Reset() {
	e.gameEndFlag = false
	e.winnerTeam = -1
	e.roundNumber = 1
	e.trickNumber = 0
	e.dealerIdx = 0
	e.teamScores = [OmiTeamCnt]int{}
	e.actionLog = nil
	e.trumpSuit = 0
	e.trumpCallerIdx = -1
	e.makerTeam = -1

	e.startRound()
}

// NextRound 次のラウンドを開始する
func (e *Omi) NextRound() {
	if e.phase != OmiPhaseRoundEnd || e.gameEndFlag {
		return
	}

	e.roundNumber++
	e.dealerIdx = (e.dealerIdx + 1) % OmiPlayerCnt
	e.trickNumber = 0
	e.trumpSuit = 0
	e.trumpCallerIdx = -1
	e.makerTeam = -1
	e.currentTrick = nil
	e.leadPlayerIdx = -1

	e.startRound()
}

// startRound はラウンドを開始し、第1段階の配り (各自4枚) を行う。
//
// 裁定事項:
// 指名者はディーラーの右隣 (反時計回りの最初の席: (dealerIdx+1)%4) とする。
// ラウンドごとにディーラーが回るので指名者も回る。
func (e *Omi) startRound() {
	for _, p := range e.players {
		p.ResetRound()
	}

	e.trumpCards = NewTrumpCards32()
	e.trumpCards.Shuffle()

	// 第1段階: 全員に4枚ずつ配る (親の右隣から反時計回り)
	for i := 0; i < OmiInitialDealSize; i++ {
		for j := 0; j < OmiPlayerCnt; j++ {
			seat := (e.dealerIdx + 1 + j) % OmiPlayerCnt
			if card := e.trumpCards.DrawCard(); card != nil {
				e.players[seat].AddCard(card)
			}
		}
	}

	e.sortAllHands()

	// 切り札指名フェーズへ
	e.trumpCallerIdx = (e.dealerIdx + 1) % OmiPlayerCnt
	e.currentPlayerIdx = e.trumpCallerIdx
	e.phase = OmiPhaseCallTrump

	e.appendLog(-1, "deal_first_batch",
		fmt.Sprintf("round %d: dealer=%d, caller=%d, dealt %d cards each",
			e.roundNumber, e.dealerIdx, e.trumpCallerIdx, OmiInitialDealSize), nil)
}

// dealSecondBatch は切り札決定後、第2段階の配り (残り4枚) を行う。
// これにより全員の手札が8枚 (合計32枚) となり、山札は空になる。
func (e *Omi) dealSecondBatch() {
	for i := 0; i < OmiInitialDealSize; i++ {
		for j := 0; j < OmiPlayerCnt; j++ {
			seat := (e.dealerIdx + 1 + j) % OmiPlayerCnt
			if card := e.trumpCards.DrawCard(); card != nil {
				e.players[seat].AddCard(card)
			}
		}
	}
	e.sortAllHands()
}

// --- Trump Selection: CallTrump Phase ---

// PlayerCallTrump 人間プレイヤーが切り札スートを指名する。
//
// 裁定事項:
// 指名者は 4 スートのいずれかを宣言する。パスは無い (必ず誰かが宣言する形にして、
// 流局を作らない裁定を採用)。
func (e *Omi) PlayerCallTrump(suit int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != OmiPhaseCallTrump {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(e.players)
	if humanIdx < 0 || e.trumpCallerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}

	e.doCallTrump(humanIdx, suit)
	return nil
}

// CpuCallTrump CPUプレイヤーがコールトランプ判断する。
func (e *Omi) CpuCallTrump() {
	if e.gameEndFlag || e.phase != OmiPhaseCallTrump {
		return
	}
	if e.players[e.trumpCallerIdx].GetIsHuman() {
		return
	}

	suit := e.cpuSelectTrumpSuit(e.trumpCallerIdx)
	e.doCallTrump(e.trumpCallerIdx, suit)
}

// doCallTrump は切り札を確定し、残り4枚を配ってプレイフェーズを開始する。
func (e *Omi) doCallTrump(callerIdx int, suit int) {
	e.trumpSuit = suit
	e.trumpCallerIdx = callerIdx
	e.makerTeam = e.players[callerIdx].GetTeam()

	suitName := suitStr(suit)
	e.appendLog(callerIdx, "call_trump",
		fmt.Sprintf("%s calls %s as trump", playerName(e.players, callerIdx), suitName), nil)

	// 切り札確定後、残り4枚を配る
	e.dealSecondBatch()

	// プレイフェーズ開始
	e.startPlayPhase()
}

// IsHumanCallTrumpTurn は現在コールトランプ手番が人間かを返す。
func (e *Omi) IsHumanCallTrumpTurn() bool {
	if e.gameEndFlag || e.phase != OmiPhaseCallTrump {
		return false
	}
	return e.players[e.trumpCallerIdx].GetIsHuman()
}

// IsHumanTurn 現在の手番が人間かを返す。
func (e *Omi) IsHumanTurn() bool {
	if e.gameEndFlag {
		return false
	}
	switch e.phase {
	case OmiPhaseCallTrump:
		return e.IsHumanCallTrumpTurn()
	case OmiPhasePlay:
		return e.players[e.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// IsHumanBidTurn ビッド/宣言手番が人間かを返す (互換性用)
func (e *Omi) IsHumanBidTurn() bool {
	return e.IsHumanCallTrumpTurn()
}

// --- Play Phase ---

// startPlayPhase プレイフェーズを開始する
func (e *Omi) startPlayPhase() {
	e.trickNumber = 1
	e.currentTrick = nil
	// 第1トリックのリードは切り札指名者 (親の右隣) から始まる
	e.leadPlayerIdx = (e.dealerIdx + 1) % OmiPlayerCnt
	e.currentPlayerIdx = e.leadPlayerIdx
	e.phase = OmiPhasePlay
}

// PlayerPlay 人間プレイヤーがカードを出す
func (e *Omi) PlayerPlay(cardIndex int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != OmiPhasePlay {
		return ErrWrongPhase
	}
	if !e.players[e.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := e.players[e.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return ErrInvalidCard
	}

	card := player.GetCard(cardIndex)
	if err := e.validatePlay(e.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	e.playCard(e.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行する
func (e *Omi) CpuPlay() {
	if e.gameEndFlag || e.phase != OmiPhasePlay {
		return
	}
	if e.players[e.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := e.players[e.currentPlayerIdx]
	cardIdx := e.cpuSelectPlayCard(e.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	if played == nil {
		return
	}
	e.playCard(e.currentPlayerIdx, played)
}

// playCard カードを場に出して手番を進める
func (e *Omi) playCard(playerIdx int, card *Card) {
	e.currentTrick = append(e.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	e.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(e.players, playerIdx), cardStr(card)), []*Card{card})

	if len(e.currentTrick) == OmiPlayerCnt {
		e.phase = OmiPhaseTrickEnd
	} else {
		e.currentPlayerIdx = (e.currentPlayerIdx + 1) % OmiPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (e *Omi) ResolveTrick() {
	if e.phase != OmiPhaseTrickEnd || len(e.currentTrick) != OmiPlayerCnt {
		return
	}

	winnerIdx := e.trickWinner()
	trickCards := make([]*Card, len(e.currentTrick))
	for i, tc := range e.currentTrick {
		trickCards[i] = tc.Card
	}

	e.players[winnerIdx].AddTrick(trickCards)

	winnerName := playerName(e.players, winnerIdx)
	e.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, e.trickNumber), trickCards)

	e.leadPlayerIdx = winnerIdx

	if e.trickNumber >= OmiHandSize {
		e.phase = OmiPhaseRoundEnd
	} else {
		e.phase = OmiPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (e *Omi) NextTrick() {
	if e.phase != OmiPhaseTrickEnd {
		return
	}
	e.currentTrick = nil
	e.currentPlayerIdx = e.leadPlayerIdx
	e.trickNumber++
	e.phase = OmiPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う。
//
// 裁定事項:
//  1. 5 トリック以上取ったチームが 1 点。
//  2. 8 トリック全取り (Omi) は追加点として合計 2 点とする
//     (1点+ボーナス1点ではなく、全取り達成時の合計得点として2点を与える解釈を採用)。
//  3. 4 対 4 の引き分け時は両チーム 0 点とする (8トリック制のため4-4が発生しうるが、
//     勝ち越し条件を満たさないため両者0点とする裁定を採用)。
//  4. 宣言側/非宣言側による得点変動は設けない。
//  5. 終局は 10 点先取 (先に到達したチームの勝利)。
func (e *Omi) ScoreRound() {
	if e.phase != OmiPhaseRoundEnd {
		return
	}

	// チームごとのトリック数を集計
	teamTricks := [OmiTeamCnt]int{}
	for _, p := range e.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}

	for t := 0; t < OmiTeamCnt; t++ {
		tricks := teamTricks[t]
		if tricks == OmiHandSize {
			// 8 トリック全取り (Omi): 合計 2 点
			e.teamScores[t] += 2
			e.appendLog(-1, "omi_win",
				fmt.Sprintf("Team %d wins all %d tricks (Omi)! +2 points", t, OmiHandSize), nil)
		} else if tricks >= 5 {
			// 5-7 トリック獲得: 1 点
			e.teamScores[t] += 1
			e.appendLog(-1, "round_win",
				fmt.Sprintf("Team %d wins the round with %d tricks! +1 point", t, tricks), nil)
		}
	}

	if teamTricks[0] == 4 && teamTricks[1] == 4 {
		// 4-4 引き分け: 両チーム 0 点
		e.appendLog(-1, "round_draw", "Round ends in a 4-4 draw. 0 points for both teams.", nil)
	}

	// スコアログ
	for ti := range OmiTeamCnt {
		e.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points (tricks: %d)", ti, e.teamScores[ti], teamTricks[ti]), nil)
	}

	// ゲーム終了判定 (PointLimit 到達)
	maxScore := max(e.teamScores[0], e.teamScores[1])
	if maxScore >= e.config.PointLimit {
		e.gameEndFlag = true
		e.phase = OmiPhaseGameEnd
		if e.teamScores[0] > e.teamScores[1] {
			e.winnerTeam = 0
		} else if e.teamScores[1] > e.teamScores[0] {
			e.winnerTeam = 1
		}
		e.appendLog(-1, "game_end",
			fmt.Sprintf("Game over! Winner: Team %d (Team 0: %d, Team 1: %d)",
				e.winnerTeam, e.teamScores[0], e.teamScores[1]), nil)
	}
}

// --- Validation & Rules ---

// validatePlay カードのプレイが有効か検証する。
// フォロー規則:
// - リードカードと同じスートを持っていれば必ず出さなければならない (マストフォロー)。
// - フォローできなければ任意のカードを出せる (切り札強制ではない)。
func (e *Omi) validatePlay(playerIdx int, card *Card) error {
	if len(e.currentTrick) == 0 {
		return nil // リードは自由
	}

	leadSuit := e.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if e.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか
func (e *Omi) playerHasSuit(playerIdx int, suit int) bool {
	p := e.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。
// 切り札が出ていれば最強の切り札が勝ち、なければリードスートの最強が勝つ。
func (e *Omi) trickWinner() int {
	if len(e.currentTrick) == 0 {
		return 0
	}

	winnerIdx := e.currentTrick[0].PlayerIdx
	winnerRank := e.cardRank(e.currentTrick[0].Card)
	leadSuit := e.currentTrick[0].Card.GetDesign()
	winnerIsTrump := (leadSuit == e.trumpSuit)

	for _, tc := range e.currentTrick[1:] {
		rank := e.cardRank(tc.Card)
		suit := tc.Card.GetDesign()
		isTrump := (suit == e.trumpSuit)

		if isTrump && !winnerIsTrump {
			// 切り札が非切り札の勝者に勝つ
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
			winnerIsTrump = true
		} else if isTrump && winnerIsTrump {
			// 切り札同士の比較
			if rank > winnerRank {
				winnerIdx = tc.PlayerIdx
				winnerRank = rank
			}
		} else if !winnerIsTrump && suit == leadSuit {
			// 切り札が出ていない場合のリードスート同士の比較
			if rank > winnerRank {
				winnerIdx = tc.PlayerIdx
				winnerRank = rank
			}
		}
	}
	return winnerIdx
}

// cardRank トリック比較用のカードランクを返す (高い = 強い)。
// Omi にはジャック昇格 (Bower) は存在せず、純粋なランク順となる:
// A(14) > K(13) > Q(12) > J(11) > 10(10) > 9(9) > 8(8) > 7(7)
// 切り札スートは 200+rank、非切り札スートは 100+rank。
func (e *Omi) cardRank(card *Card) int {
	base := card.GetValue()
	if base == 1 { // Ace は最強 (14)
		base = 14
	}
	if card.GetDesign() == e.trumpSuit {
		return 200 + base
	}
	return 100 + base
}

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (e *Omi) CardRankPublic(card *Card) int { return e.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト・互換用公開メソッド)
func (e *Omi) EffectiveSuitPublic(card *Card) int { return card.GetDesign() }

// sortAllHands 全プレイヤーの手札をソート
func (e *Omi) sortAllHands() {
	for _, p := range e.players {
		e.sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート・ランク順にソート
func (e *Omi) sortHand(player *OmiPlayer) {
	sortPlayerHand(player, func(ci, cj *Card) bool {
		si := ci.GetDesign()
		sj := cj.GetDesign()
		if si != sj {
			return si < sj
		}
		return e.cardRank(ci) < e.cardRank(cj)
	})
}

// --- Getters & Setters ---

// GetConfig ゲーム設定を取得する
func (e *Omi) GetConfig() OmiConfig { return e.config }

// SetConfig 設定変更
func (e *Omi) SetConfig(cfg OmiConfig) { e.config = cfg }

// GetGameEndFlag ゲーム終了フラグ取得
func (e *Omi) GetGameEndFlag() bool { return e.gameEndFlag }

// GetPhase 現在のフェーズ取得
func (e *Omi) GetPhase() OmiPhase { return e.phase }

// GetRoundNumber ラウンド番号取得
func (e *Omi) GetRoundNumber() int { return e.roundNumber }

// GetTrickNumber トリック番号取得
func (e *Omi) GetTrickNumber() int { return e.trickNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (e *Omi) GetCurrentPlayerIdx() int { return e.currentPlayerIdx }

// GetCurrentTrick 現在のトリックを取得
func (e *Omi) GetCurrentTrick() []*TrickCard { return e.currentTrick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (e *Omi) GetLeadPlayerIdx() int { return e.leadPlayerIdx }

// GetDealerIdx ディーラーインデックス取得
func (e *Omi) GetDealerIdx() int { return e.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (e *Omi) GetTrumpSuit() int { return e.trumpSuit }

// GetTrumpCallerIdx 切り札指名者インデックス取得
func (e *Omi) GetTrumpCallerIdx() int { return e.trumpCallerIdx }

// GetCallerIdx 切り札指名者インデックス取得 (エイリアス)
func (e *Omi) GetCallerIdx() int { return e.trumpCallerIdx }

// GetMakerTeam 切り札を指名したチーム取得
func (e *Omi) GetMakerTeam() int { return e.makerTeam }

// GetTeamScore チームスコア取得
func (e *Omi) GetTeamScore(team int) int {
	if team < 0 || team >= OmiTeamCnt {
		return 0
	}
	return e.teamScores[team]
}

// SetPhase フェーズ設定 (テスト用)
func (e *Omi) SetPhase(phase OmiPhase) { e.phase = phase }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (e *Omi) SetRoundNumber(n int) { e.roundNumber = n }

// SetTrickNumber トリック番号設定 (テスト用)
func (e *Omi) SetTrickNumber(n int) { e.trickNumber = n }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (e *Omi) SetCurrentPlayerIdx(idx int) { e.currentPlayerIdx = idx }

// SetCurrentTrick トリック設定 (テスト用)
func (e *Omi) SetCurrentTrick(trick []*TrickCard) { e.currentTrick = trick }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (e *Omi) SetLeadPlayerIdx(idx int) { e.leadPlayerIdx = idx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (e *Omi) SetDealerIdx(idx int) { e.dealerIdx = idx }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (e *Omi) SetTrumpSuit(suit int) { e.trumpSuit = suit }

// SetTrumpCallerIdx 切り札指名者設定 (テスト用)
func (e *Omi) SetTrumpCallerIdx(idx int) { e.trumpCallerIdx = idx }

// SetMakerTeam メイカーチーム設定 (テスト用)
func (e *Omi) SetMakerTeam(team int) { e.makerTeam = team }

// SetTeamScore チームスコア設定 (テスト用)
func (e *Omi) SetTeamScore(team, score int) {
	if team >= 0 && team < OmiTeamCnt {
		e.teamScores[team] = score
	}
}

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (e *Omi) SetGameEndFlag(flag bool) { e.gameEndFlag = flag }

// SetWinnerTeam 勝利チーム設定 (テスト用)
func (e *Omi) SetWinnerTeam(team int) { e.winnerTeam = team }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (e *Omi) GetWinnerTeam() int { return e.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (e *Omi) GetPlayerCnt() int { return len(e.players) }

// GetPlayer 指定インデックスのプレイヤー取得
func (e *Omi) GetPlayer(i int) *OmiPlayer {
	if i < 0 || i >= len(e.players) {
		return nil
	}
	return e.players[i]
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (e *Omi) GetValidPlayIndices(playerIdx int) []int {
	return e.getValidPlayIndices(playerIdx)
}

func (e *Omi) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(e.players[playerIdx], func(c *Card) bool { return e.validatePlay(playerIdx, c) == nil })
}

// GetHint ヒントを取得する
func (e *Omi) GetHint() *OmiHint {
	humanIdx := findHumanIdx(e.players)
	if humanIdx < 0 {
		return nil
	}

	switch e.phase {
	case OmiPhaseCallTrump:
		if e.trumpCallerIdx != humanIdx {
			return nil
		}
		suit := e.cpuSelectTrumpSuit(humanIdx)
		return &OmiHint{Suit: &suit, Reason: "strategic_call"}

	case OmiPhasePlay:
		if e.currentPlayerIdx != humanIdx {
			return nil
		}
		validIndices := e.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := e.cpuPlayHard(humanIdx, validIndices)
		return &OmiHint{CardIndex: &idx, Reason: e.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (e *Omi) playHintReason(chosenIdx int) string {
	humanIdx := findHumanIdx(e.players)
	if humanIdx < 0 {
		return "normal_play"
	}
	player := e.players[humanIdx]
	if chosenIdx < 0 || chosenIdx >= player.GetCardsSize() {
		return "normal_play"
	}
	card := player.GetCard(chosenIdx)

	if len(e.currentTrick) == 0 {
		if card.GetDesign() == e.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}

	leadSuit := e.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == e.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- CPU AI ---

// cpuSelectTrumpSuit 4枚の手札を評価し、最も強いスートを切り札として選択する
func (e *Omi) cpuSelectTrumpSuit(playerIdx int) int {
	bestSuit := CardDesignSpade
	bestScore := -1

	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for _, suit := range suits {
		score := e.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	return bestSuit
}

// evalHandForTrump は指定スートを切り札とした場合の手札強度を評価する
func (e *Omi) evalHandForTrump(playerIdx int, suit int) int {
	player := e.players[playerIdx]
	score := 0

	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		val := c.GetValue()
		if val == 1 {
			val = 14
		}

		if c.GetDesign() == suit {
			// 切り札カード: 枚数と強さに大きな重み
			score += 10 + val
		} else {
			// サイドスートのハイカード
			switch val {
			case 14:
				score += 4
			case 13:
				score += 2
			}
		}
	}
	return score
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (e *Omi) cpuSelectPlayCard(playerIdx int) int {
	validIndices := e.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch e.config.CpuDifficulty {
	case OmiCpuDifficultyHard:
		return e.cpuPlayHard(playerIdx, validIndices)
	case OmiCpuDifficultyNormal:
		return e.cpuPlayNormal(playerIdx, validIndices)
	default:
		return e.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (e *Omi) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 基本戦略プレイ
func (e *Omi) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := e.players[playerIdx]

	if len(e.currentTrick) == 0 {
		// リード: 最も強いカード
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank > bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー: 勝てるなら最小の勝てるカード、勝てないなら最弱
	winnerRank := e.currentWinnerRank()
	overCards := []int{}
	for _, idx := range validIndices {
		if e.cardRank(player.GetCard(idx)) > winnerRank {
			overCards = append(overCards, idx)
		}
	}
	if len(overCards) > 0 {
		bestIdx := overCards[0]
		for _, idx := range overCards[1:] {
			if e.cardRank(player.GetCard(idx)) < e.cardRank(player.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝てない: 最弱カード
	bestIdx := validIndices[0]
	bestRank := e.cardRank(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		rank := e.cardRank(player.GetCard(idx))
		if rank < bestRank {
			bestRank = rank
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 高度な戦略プレイ (パートナーの勝利状況を考慮)
func (e *Omi) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := e.players[playerIdx]

	if len(e.currentTrick) == 0 {
		// リード: 最強カード
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank > bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	isLastPlayer := len(e.currentTrick) == OmiPlayerCnt-1
	winnerRank := e.currentWinnerRank()
	winnerTeam := e.players[e.currentTrickWinnerIdx()].GetTeam()
	myTeam := e.players[playerIdx].GetTeam()

	// パートナーが勝っていて自分が最後番なら、最弱カードで流す
	if winnerTeam == myTeam && isLastPlayer {
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank < bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝ちに行く
	overCards := []int{}
	for _, idx := range validIndices {
		if e.cardRank(player.GetCard(idx)) > winnerRank {
			overCards = append(overCards, idx)
		}
	}
	if len(overCards) > 0 {
		bestIdx := overCards[0]
		for _, idx := range overCards[1:] {
			if e.cardRank(player.GetCard(idx)) < e.cardRank(player.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝てない: 最弱カードを捨てる
	bestIdx := validIndices[0]
	bestRank := e.cardRank(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		rank := e.cardRank(player.GetCard(idx))
		if rank < bestRank {
			bestRank = rank
			bestIdx = idx
		}
	}
	return bestIdx
}

// currentWinnerRank 現在のトリックで最強カードのランクを返す
func (e *Omi) currentWinnerRank() int {
	if len(e.currentTrick) == 0 {
		return 0
	}
	winnerIdx := e.currentTrickWinnerIdx()
	for _, tc := range e.currentTrick {
		if tc.PlayerIdx == winnerIdx {
			return e.cardRank(tc.Card)
		}
	}
	return e.cardRank(e.currentTrick[0].Card)
}

// currentTrickWinnerIdx 現在のトリック内で暫定勝者のインデックスを返す
func (e *Omi) currentTrickWinnerIdx() int {
	if len(e.currentTrick) == 0 {
		return 0
	}
	return e.trickWinner()
}

// --- JSON Serialization ---

// omiJSON is the JSON wire format for Omi.
type omiJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*OmiPlayer      `json:"ps"`
	Config           OmiConfig         `json:"cf"`
	Phase            OmiPhase          `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	DealerIdx        int               `json:"di"`
	TrumpSuit        int               `json:"ts"`
	TrumpCallerIdx   int               `json:"tci"`
	MakerTeam        int               `json:"mt"`
	TeamScores       [OmiTeamCnt]int   `json:"sc"`
	LeadPlayerIdx    int               `json:"li"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (e *Omi) MarshalJSON() ([]byte, error) {
	return json.Marshal(omiJSON{
		TrumpCards:       e.trumpCards,
		Players:          e.players,
		Config:           e.config,
		Phase:            e.phase,
		RoundNumber:      e.roundNumber,
		TrickNumber:      e.trickNumber,
		CurrentPlayerIdx: e.currentPlayerIdx,
		CurrentTrick:     e.currentTrick,
		DealerIdx:        e.dealerIdx,
		TrumpSuit:        e.trumpSuit,
		TrumpCallerIdx:   e.trumpCallerIdx,
		MakerTeam:        e.makerTeam,
		TeamScores:       e.teamScores,
		LeadPlayerIdx:    e.leadPlayerIdx,
		GameEndFlag:      e.gameEndFlag,
		WinnerTeam:       e.winnerTeam,
		ActionLog:        e.actionLog,
	})
}

const omiMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (e *Omi) UnmarshalJSON(data []byte) error {
	var j omiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > omiMaxSliceLen || len(j.CurrentTrick) > omiMaxSliceLen ||
		len(j.ActionLog) > omiMaxSliceLen {
		return fmt.Errorf("omi: input array exceeds maximum allowed size")
	}
	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCards32()
	}
	e.players = j.Players
	if e.players == nil {
		e.players = make([]*OmiPlayer, 0)
	}
	e.config = j.Config
	e.phase = j.Phase
	e.roundNumber = j.RoundNumber
	e.trickNumber = j.TrickNumber
	e.currentPlayerIdx = j.CurrentPlayerIdx
	e.currentTrick = j.CurrentTrick
	if e.currentTrick == nil {
		e.currentTrick = make([]*TrickCard, 0)
	}
	e.dealerIdx = j.DealerIdx
	e.trumpSuit = j.TrumpSuit
	e.trumpCallerIdx = j.TrumpCallerIdx
	e.makerTeam = j.MakerTeam
	e.teamScores = j.TeamScores
	e.leadPlayerIdx = j.LeadPlayerIdx
	e.gameEndFlag = j.GameEndFlag
	e.winnerTeam = j.WinnerTeam
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
