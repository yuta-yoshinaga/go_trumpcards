//go:build !js || !wasm || extra4

// Package domain セブン・トゥエンティセブン (SevenTwentySeven) のドメインモデル。
//
// アメリカの「dealer's choice」ホームゲーム。標準 52 枚デッキを使い、2〜7 人が
// アンティを入れてから 2 枚ずつ配られる。**ポットは 2 つに割れる。**
//
// # 1 ラウンドの流れ
//
//  1. 全員がアンティをポットに払い、2 枚ずつ配られる。
//  2. 各巡、各自が「もう 1 枚引く」か「止まる」かを選ぶ。**止まったらそのラウンドは
//     もう配られない。** 引くと言った人にだけ 1 枚ずつ配る。
//  3. 全員が止まった時点でショーダウン。
//  4. **7 に最も近い (超えていない) 手**と **27 に最も近い (超えていない) 手**が
//     それぞれポットの半分を取る。同じ人が両方を制したら総取り (スクープ)。
//  5. 片側に生存者がいなければ、もう片方が総取りする。両側とも全滅ならポットは
//     次ラウンドへ持ち越す。同点は分け合う。
//
// # カードの点数
//
// ここがこのゲーム固有の値:
//
//   - 絵札 (J/Q/K) = **0.5 点**
//   - エース = **1 点 または 11 点** (1 枚ごとに選べる)
//   - それ以外は数字どおり
//
// **点は ×2 の整数で保持する** (`seventwentyseven_score.go`)。絵札が 1、数字が 2n、
// 目標が 14 / 54 になり、0.5 刻みを float64 の比較なしに正確に扱える。表示のときだけ
// 2 で割る。Call Break が int×10 でスコアを持っているのと同じ判断。
//
// # 超過
//
// **超えた側だけが失格になる。** 19 点の手は 7 側では死んでいるが 27 側では生きている。
// この非対称がゲームの読み合いそのもので、片側の判定を共有すると消える。
//
// # 停止条件
//
// 規定ラウンド数 (TargetRounds) を消化するか、アンティを払えるプレイヤーが 2 人未満
// (人間の脱落を含む) になると終了し、チップ最多のプレイヤーが勝者となる。
//
// # デッキ
//
// 標準 52 枚 (ジョーカーなし)。NewTrumpCards(0) は extra4 ワーカーから到達可能。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SevenTwentySevenPhase はゲームフェーズ。
type SevenTwentySevenPhase int

// SevenTwentySeven のフェーズ定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// SevenTwentySevenPhaseDraw 引くか止まるかの受付中 (人間の宣言待ち)。ワイヤー値 0。
	//
	// **Guts と違って 1 回で終わらない。** 全員が「止まる」を選ぶまで何巡でも
	// 続き、そのたびに希望者へ 1 枚ずつ配る。ここがこのゲームの駆け引きで、
	// 1 回の宣言で解決してしまうと 7 と 27 のどちらに寄せるかの選択が消える。
	SevenTwentySevenPhaseDraw SevenTwentySevenPhase = 0
	// SevenTwentySevenPhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 1。
	SevenTwentySevenPhaseResult SevenTwentySevenPhase = 1
)

// SevenTwentySevenSide は 2 方向のうちどちらを指すか。
const (
	// SevenTwentySevenSideLow は 7 に近い側。
	SevenTwentySevenSideLow = 0
	// SevenTwentySevenSideHigh は 27 に近い側。
	SevenTwentySevenSideHigh = 1
)

// SevenTwentySevenResult は人間プレイヤーから見たラウンド結果。
// 値は GameResult と同一だが、この型名は JSON ペイロードに出るため統合していない（#4462）。
type SevenTwentySevenResult int

const (
	// SevenTwentySevenResultLose 負け (どちらの側も取れなかった)
	SevenTwentySevenResultLose SevenTwentySevenResult = -1
	// SevenTwentySevenResultNone 結果なし (未参加 / 未解決)
	SevenTwentySevenResultNone SevenTwentySevenResult = 0
	// SevenTwentySevenResultWin 勝ち (ポットの一部または全部を獲得)
	SevenTwentySevenResultWin SevenTwentySevenResult = 1
)

// sevenTwentySevenMaxSliceLen はデシリアライズ時のスライス長の上限。
const sevenTwentySevenMaxSliceLen = 1000

// sevenTwentySevenCpuLowSlack / HighSlack は CPU が「まだ引く」と判断する余裕幅
// （内部値）。0 にすると目標ちょうどでない限り引き続けて必ず超過するので、
// 手前で止まる幅が要る。低い側は 1 枚で壊れるので狭く、高い側は広く取る。
const (
	sevenTwentySevenCpuLowSlack  = 2 * SevenTwentySevenScoreScale
	sevenTwentySevenCpuHighSlack = 4 * SevenTwentySevenScoreScale
)

// デシリアライズ検証用のセンチネルエラー。
var (
	errSevenTwentySevenSnapshot      = errors.New("sevenTwentySeven: invalid serialised game state")
	errSevenTwentySevenInvalidPlayer = errors.New("sevenTwentySeven: invalid player state")
)

// SevenTwentySevenHint はヒント情報 (人間への in/out 助言)。
type SevenTwentySevenHint struct {
	// Draw は 1 枚引くことを勧めるか。false は「止まれ」。
	Draw bool
	// Reason はヒント理由キー。**どちらの側を狙っているかを含む** ——
	// 「引け」だけでは 7 狙いか 27 狙いかが読めず、助言として成立しない。
	Reason string
}

// sevenTwentySevenState はゲーム進行状態。
type sevenTwentySevenState struct {
	phase       SevenTwentySevenPhase
	roundNumber int
	pot         int // 現在のラウンドのポット
	carryPot    int // 次ラウンドへ持ち越す種銭 (全員が両側とも超えたラウンドのポット)
	carryCount  int // 両側とも全滅してポットが連続繰り越しになった回数
	// lowWinner / highWinner は 7 側 / 27 側それぞれの勝者 (-1 = 該当なし)。
	// 同一人物なら総取り。どちらも -1 ならポットは次ラウンドへ持ち越す。
	lowWinner      int
	highWinner     int
	drawRound      int // 何巡目の「引く / 止まる」か (1 始まり)
	matchWinnerIdx int // ゲーム全体の勝者 (-1 = 未確定)
	result         SevenTwentySevenResult
	gameEndFlag    bool
	scored         bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// SevenTwentySeven はセブン・トゥエンティセブンの状態を保持する集約ルート。
type SevenTwentySeven struct {
	trumpCards *TrumpCards
	players    []*SevenTwentySevenPlayer
	config     SevenTwentySevenConfig
	state      sevenTwentySevenState
}

// NewSevenTwentySeven はコンストラクタ。
func NewSevenTwentySeven(trumpCards *TrumpCards, players []*SevenTwentySevenPlayer, config SevenTwentySevenConfig) *SevenTwentySeven {
	return &SevenTwentySeven{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: sevenTwentySevenState{
			phase:          SevenTwentySevenPhaseDraw,
			lowWinner:      -1,
			highWinner:     -1,
			matchWinnerIdx: -1,
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultSevenTwentySeven は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultSevenTwentySeven() *SevenTwentySeven {
	cfg := DefaultSevenTwentySevenConfig()
	g := NewSevenTwentySeven(NewTrumpCards(0), sevenTwentySevenNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// sevenTwentySevenNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func sevenTwentySevenNewPlayers(cfg SevenTwentySevenConfig) []*SevenTwentySevenPlayer {
	players := make([]*SevenTwentySevenPlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewSevenTwentySevenPlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数を設定から作り直し、第 1 ラウンドを配る。
func (g *SevenTwentySeven) Reset() {
	g.players = sevenTwentySevenNewPlayers(g.config)
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	g.state = sevenTwentySevenState{
		phase:          SevenTwentySevenPhaseDraw,
		roundNumber:    1,
		lowWinner:      -1,
		highWinner:     -1,
		matchWinnerIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound は同じチップを保持したまま次のラウンドを配る。Result フェーズかつゲーム
// 継続中のときのみ有効。
func (g *SevenTwentySeven) NextRound() {
	if g.state.phase != SevenTwentySevenPhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 脱落判定・アンティ徴収・配札・宣言フェーズへ遷移。
func (g *SevenTwentySeven) startRound() {
	// ラウンド単位の状態をクリア。
	g.state.lowWinner = -1
	g.state.highWinner = -1
	g.state.result = SevenTwentySevenResultNone
	g.state.scored = false
	for _, p := range g.players {
		p.ResetForRound()
	}
	// アンティを払えないプレイヤーは脱落。
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() < g.config.Ante {
			p.SetOut(true)
		}
	}
	// 参加可能なプレイヤーが 2 人未満、または人間が脱落 → ゲーム終了。
	if g.solventCount() < SevenTwentySevenMinPlayerCount || g.players[0].GetOut() {
		g.endGame()
		return
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	// 持ち越しを種銭にし、アンティ徴収 + 配札。
	g.state.pot = g.state.carryPot
	g.state.carryPot = 0
	for _, p := range g.players {
		if p.GetOut() {
			continue
		}
		p.SubtractChips(g.config.Ante)
		p.AddRoundBet(g.config.Ante)
		g.state.pot += g.config.Ante
		for i := 0; i < SevenTwentySevenHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.state.drawRound = 1
	g.state.phase = SevenTwentySevenPhaseDraw
	g.appendLog(-1, "deal",
		fmt.Sprintf("Round %d: ante %d, pot %d", g.state.roundNumber, g.config.Ante, g.state.pot), nil)
}

// TakeCard は人間 (seat 0) の「引く / 止まる」を受け付ける。
//
// **1 回では終わらない。** 止まると宣言した人はその後配られず、まだ引く人が
// いるあいだラウンドは続く。全員が止まった時点でショーダウン。
func (g *SevenTwentySeven) TakeCard(draw bool) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != SevenTwentySevenPhaseDraw {
		return NewDomainError(ErrWrongPhase, "you can only draw or stand during the draw phase")
	}
	human := g.players[0]
	if human.GetOut() {
		return NewDomainError(ErrInvalidPlay, "you are out of the game")
	}
	if human.GetStanding() {
		return NewDomainError(ErrInvalidPlay, "you have already stood pat this round")
	}
	human.SetStanding(!draw)
	g.appendLog(0, "declare", sevenTwentySevenDeclareText(0, draw), nil)

	// **人間が止まったあとも CPU は引き続ける。** そこで待ってしまうと、
	// 人間には打つ手が無いのに全員が止まるまで盤が進まず、ゲームが固まる。
	// 人間がまだ引くつもりなら 1 巡で戻り、止まったなら決着まで自分で回す。
	// 上限は「山札 52 枚 ÷ 1 巡 1 枚」を超えない安全弁。
	for guard := 0; guard < CardCnt; guard++ {
		g.cpuDeclare()
		g.dealToDrawers()

		if g.everyoneStanding() {
			g.settle()
			return nil
		}
		g.state.drawRound++
		if !g.players[0].GetStanding() {
			return nil // 人間の次の判断を待つ
		}
	}
	// ここに来るのは配りが詰まったとき。決着させて先へ進める。
	g.settle()
	return nil
}

// dealToDrawers は「まだ止まっていない」全員に 1 枚ずつ配る。
//
// **止まった人には二度と配らない。** ここを取り違えると、止まった判断が
// 無意味になり、超過が事故でしか起きなくなる。
func (g *SevenTwentySeven) dealToDrawers() {
	for _, p := range g.players {
		if p.GetOut() || p.GetStanding() {
			continue
		}
		c := g.trumpCards.DrawCard()
		if c == nil {
			// 山札が尽きた: これ以上引けないので全員を止める。
			// 52 枚あるので通常は起きないが、起きたときに無限ループへ
			// 落ちないほうが大事。
			for _, q := range g.players {
				q.SetStanding(true)
			}
			g.appendLog(-1, "deck_empty", "the deck ran out; everyone stands pat", nil)
			return
		}
		p.AddCard(c)
		idx := sevenTwentySevenIndexOf(g.players, p)
		g.appendLog(idx, "draw",
			fmt.Sprintf("%s takes a card", playerName(g.players, idx)), []*Card{c})
	}
}

// everyoneStanding は参加者全員が止まったかを返す。
func (g *SevenTwentySeven) everyoneStanding() bool {
	for _, p := range g.players {
		if !p.GetOut() && !p.GetStanding() {
			return false
		}
	}
	return true
}

// cpuDeclare は CPU の「引く / 止まる」を決める。
func (g *SevenTwentySeven) cpuDeclare() {
	for i, p := range g.players {
		if i == 0 || p.GetOut() || p.GetStanding() {
			continue
		}
		draw := g.cpuDraws(p)
		p.SetStanding(!draw)
		g.appendLog(i, "declare", sevenTwentySevenDeclareText(i, draw), nil)
	}
}

// cpuDraws は CPU が 1 枚引くかを決める。
//
// **まずどちらの側を狙うかを決める。** 7 側と 27 側では判断が正反対で、
// 常に 27 を追うと**低い側の勝負が消える**（実測: 4 人 × 3 ラウンドで 7 側の
// 生存者が 0 になった）。7 を狙うなら 1 枚で壊れるので早く止まり、27 を狙うなら
// 届くまで引く。
//
// 選び方は「その側の目標にどれだけ近いか」の比較。低い側は絵札 (0.5) を
// 引ければ伸ばせるので、目標に近ければ守りに入る価値がある。
func (g *SevenTwentySeven) cpuDraws(p *SevenTwentySevenPlayer) bool {
	cards := sevenTwentySevenHandOf(p)
	low, lowOK := SevenTwentySevenBestFor(cards, SevenTwentySevenLowTarget)
	high, highOK := SevenTwentySevenBestFor(cards, SevenTwentySevenHighTarget)

	switch {
	case !lowOK && !highOK:
		return false // どちらも失格。引いても失うものは無いが、得るものも無い
	case lowOK && low == SevenTwentySevenLowTarget:
		return false // 7 ちょうど。これ以上引けば壊すだけ
	case highOK && high == SevenTwentySevenHighTarget:
		return false // 27 ちょうど
	case !highOK:
		// 27 側は失格済み。7 側だけが残っている。
		return low < SevenTwentySevenLowTarget-sevenTwentySevenCpuLowSlack
	case !lowOK:
		// 7 側は失格済み。27 側だけ。
		return high < SevenTwentySevenHighTarget-sevenTwentySevenCpuHighSlack
	default:
		// 両側が生きている。**残りが少ないほうを狙う。** 7 側は残り 7 点以内で
		// 必ず近く、そこから 1 枚引くと壊れやすいので、守れるなら守る。
		lowGap := SevenTwentySevenLowTarget - low
		highGap := SevenTwentySevenHighTarget - high
		if lowGap <= highGap {
			return lowGap > sevenTwentySevenCpuLowSlack
		}
		return highGap > sevenTwentySevenCpuHighSlack
	}
}

// settle は 2 方向それぞれの勝者を決め、ポットを分ける。
//
//   - **7 に最も近い（超えていない）人**と**27 に最も近い（超えていない）人**が
//     それぞれポットの半分を取る。
//   - **同一人物が両方を制したら総取り。** これがこのゲームの狙いどころ。
//   - 片側に生存者がいなければ、もう片方が総取り。両側とも全滅ならポットは
//     次ラウンドへ持ち越す（carryPot）。
//   - 同点は分け合う。
func (g *SevenTwentySeven) settle() {
	g.state.lowWinner = -1
	g.state.highWinner = -1
	g.state.result = SevenTwentySevenResultNone

	lowWinners := g.closestTo(SevenTwentySevenLowTarget)
	highWinners := g.closestTo(SevenTwentySevenHighTarget)

	if len(lowWinners) > 0 {
		g.state.lowWinner = lowWinners[0]
	}
	if len(highWinners) > 0 {
		g.state.highWinner = highWinners[0]
	}

	if len(lowWinners) == 0 && len(highWinners) == 0 {
		g.state.carryPot = g.state.pot
		g.state.carryCount++
		g.appendLog(-1, "result",
			fmt.Sprintf("everyone busted both ways; pot %d carries over", g.state.pot), nil)
		g.finishRound()
		return
	}

	// 片側が全滅していれば、もう片方が総取り。
	shares := map[int]int{}
	switch {
	case len(lowWinners) == 0:
		g.payOut(shares, highWinners, g.state.pot)
	case len(highWinners) == 0:
		g.payOut(shares, lowWinners, g.state.pot)
	default:
		half := g.state.pot / 2
		g.payOut(shares, lowWinners, half)
		// 端数は 27 側へ。半分に割れない額をどちらかに寄せる必要があり、
		// 到達の難しい側に付けるほうが筋が通る。
		g.payOut(shares, highWinners, g.state.pot-half)
	}

	for idx, amount := range shares {
		g.players[idx].AddChips(amount)
		g.appendLog(idx, "win",
			fmt.Sprintf("%s takes %d", playerName(g.players, idx), amount), nil)
	}
	g.state.carryCount = 0
	g.setHumanResult(shares)
	g.finishRound()
}

// payOut は winners に amount を等分して shares に積む（端数は先頭から 1 ずつ）。
func (g *SevenTwentySeven) payOut(shares map[int]int, winners []int, amount int) {
	if len(winners) == 0 || amount <= 0 {
		return
	}
	each := amount / len(winners)
	rem := amount % len(winners)
	for i, idx := range winners {
		pay := each
		if i < rem {
			pay++
		}
		shares[idx] += pay
	}
}

// closestTo は target を超えない範囲で最も近いプレイヤーを返す（同点は全員）。
// 誰も残っていなければ空。
func (g *SevenTwentySeven) closestTo(target int) []int {
	best, found := 0, false
	var winners []int
	for i, p := range g.players {
		if p.GetOut() {
			continue
		}
		score, ok := SevenTwentySevenBestFor(sevenTwentySevenHandOf(p), target)
		if !ok {
			continue // その側は失格
		}
		switch {
		case !found || score > best:
			best, found = score, true
			winners = []int{i}
		case score == best:
			winners = append(winners, i)
		}
	}
	return winners
}

// finishRound はラウンドを閉じ、結果フェーズへ移す。
func (g *SevenTwentySeven) finishRound() {
	g.state.pot = 0
	g.state.scored = true
	g.state.phase = SevenTwentySevenPhaseResult
	g.checkGameEnd()
}

// setHumanResult は人間 (seat 0) の勝敗結果を設定する。
func (g *SevenTwentySeven) setHumanResult(shares map[int]int) {
	human := g.players[0]
	if human.GetOut() {
		g.state.result = SevenTwentySevenResultNone
		return
	}
	if shares[0] > 0 {
		g.state.result = SevenTwentySevenResultWin
		return
	}
	g.state.result = SevenTwentySevenResultLose
}

// checkGameEnd は停止条件 (規定ラウンド到達 or 参加可能者 2 人未満) を判定し、
// 満たせばゲームを終了させる。
func (g *SevenTwentySeven) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds || g.solventCount() < SevenTwentySevenMinPlayerCount {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *SevenTwentySeven) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = SevenTwentySevenPhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", playerName(g.players, g.state.matchWinnerIdx)), nil)
}

// --- ヘルパー ---

// solventCount はアンティを払える (非脱落かつチップ >= アンティ) プレイヤー数を返す。
func (g *SevenTwentySeven) solventCount() int {
	return countPlayers(g.players, func(p *SevenTwentySevenPlayer) bool { return !p.GetOut() && p.GetChips() >= g.config.Ante })
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *SevenTwentySeven) richestIdx() int {
	return maxIndexBy(g.players, func(p *SevenTwentySevenPlayer) int { return p.GetChips() })
}

// sevenTwentySevenDeclareText は宣言の棋譜テキストを返す。
func sevenTwentySevenDeclareText(idx int, draw bool) string {
	verb := "stands pat"
	if draw {
		verb = "asks for a card"
	}
	return fmt.Sprintf("player %d %s", idx, verb)
}

func (g *SevenTwentySeven) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- Hint ---

// GetHint は引くか止まるかの助言を返す。
//
// **どちらの側を狙っているかで理由が変わる。** 「引け」とだけ言われても、
// 7 を狙って引くのか 27 を狙って引くのかで意味が正反対なので、理由キーで示す。
func (g *SevenTwentySeven) GetHint() *SevenTwentySevenHint {
	if g.state.phase != SevenTwentySevenPhaseDraw || g.state.gameEndFlag {
		return nil
	}
	human := g.players[0]
	if human.GetOut() || human.GetStanding() || human.GetCardsSize() == 0 {
		return nil
	}
	cards := sevenTwentySevenHandOf(human)
	low, lowOK := SevenTwentySevenBestFor(cards, SevenTwentySevenLowTarget)
	high, highOK := SevenTwentySevenBestFor(cards, SevenTwentySevenHighTarget)

	switch {
	case !lowOK && !highOK:
		return &SevenTwentySevenHint{Draw: false, Reason: "bust_both"}
	case lowOK && low == SevenTwentySevenLowTarget:
		return &SevenTwentySevenHint{Draw: false, Reason: "exactly_seven"}
	case highOK && high == SevenTwentySevenHighTarget:
		return &SevenTwentySevenHint{Draw: false, Reason: "exactly_twentyseven"}
	case g.cpuDraws(human):
		if !highOK {
			return &SevenTwentySevenHint{Draw: true, Reason: "chase_seven"}
		}
		return &SevenTwentySevenHint{Draw: true, Reason: "chase_twentyseven"}
	default:
		return &SevenTwentySevenHint{Draw: false, Reason: "stand_pat"}
	}
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *SevenTwentySeven) GetPhase() SevenTwentySevenPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *SevenTwentySeven) SetPhase(p SevenTwentySevenPhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *SevenTwentySeven) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *SevenTwentySeven) GetRoundNumber() int { return g.state.roundNumber }

// GetPot は現在のポットを返す。
func (g *SevenTwentySeven) GetPot() int { return g.state.pot }

// SetPot はポットを設定する (テスト用)。
func (g *SevenTwentySeven) SetPot(v int) { g.state.pot = v }

// GetCarryPot は次ラウンドへの持ち越し種銭を返す。
func (g *SevenTwentySeven) GetCarryPot() int { return g.state.carryPot }

// GetCarryCount 全員アウトでポットが連続して繰り越された回数を返す。
func (g *SevenTwentySeven) GetCarryCount() int { return g.state.carryCount }

// GetAnte はアンティ額を返す。
func (g *SevenTwentySeven) GetAnte() int { return g.config.Ante }

// GetLowWinner は 7 側の勝者を返す (-1 = 該当なし)。
func (g *SevenTwentySeven) GetLowWinner() int { return g.state.lowWinner }

// GetHighWinner は 27 側の勝者を返す (-1 = 該当なし)。
func (g *SevenTwentySeven) GetHighWinner() int { return g.state.highWinner }

// GetDrawRound は何巡目の「引く / 止まる」かを返す (1 始まり)。
func (g *SevenTwentySeven) GetDrawRound() int { return g.state.drawRound }

// GetScore は playerIdx の side 側の得点（内部値）と、その側で生きているかを返す。
func (g *SevenTwentySeven) GetScore(playerIdx, side int) (int, bool) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return 0, false
	}
	target := SevenTwentySevenLowTarget
	if side == SevenTwentySevenSideHigh {
		target = SevenTwentySevenHighTarget
	}
	return SevenTwentySevenBestFor(sevenTwentySevenHandOf(g.players[playerIdx]), target)
}

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *SevenTwentySeven) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *SevenTwentySeven) GetResult() SevenTwentySevenResult { return g.state.result }

// GetPlayerCnt はプレイヤー数を返す。
func (g *SevenTwentySeven) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *SevenTwentySeven) GetPlayer(i int) *SevenTwentySevenPlayer {
	return getPlayer(g.players, i)
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *SevenTwentySeven) GetChips() int {
	return chipsOfFirst(g.players)
}

// GetConfig はローカルルール設定を返す。
func (g *SevenTwentySeven) GetConfig() SevenTwentySevenConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *SevenTwentySeven) SetConfig(cfg SevenTwentySevenConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *SevenTwentySeven) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// --- JSON Serialization ---

// sevenTwentySevenJSON is the JSON wire format for SevenTwentySeven.
type sevenTwentySevenJSON struct {
	TrumpCards  *TrumpCards               `json:"tc"`
	Players     []*SevenTwentySevenPlayer `json:"ps"`
	Config      SevenTwentySevenConfig    `json:"cf"`
	Phase       SevenTwentySevenPhase     `json:"ph"`
	RoundNumber int                       `json:"rn"`
	Pot         int                       `json:"pt"`
	CarryPot    int                       `json:"cp"`
	CarryCount  int                       `json:"cc"`
	// **勝者は 2 人いる。** 7 側と 27 側それぞれ。同一人物なら総取り。
	// 片方を落とすと、復元した盤で「なぜ半分しか貰えていないのか」が消える。
	LowWinner      int                    `json:"lw"`
	HighWinner     int                    `json:"hw"`
	DrawRound      int                    `json:"dr"`
	MatchWinnerIdx int                    `json:"mw"`
	Result         SevenTwentySevenResult `json:"re"`
	GameEndFlag    bool                   `json:"ge"`
	Scored         bool                   `json:"sc"`
	ActionLog      []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *SevenTwentySeven) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevenTwentySevenJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.state.phase,
		RoundNumber:    g.state.roundNumber,
		Pot:            g.state.pot,
		CarryPot:       g.state.carryPot,
		CarryCount:     g.state.carryCount,
		LowWinner:      g.state.lowWinner,
		HighWinner:     g.state.highWinner,
		DrawRound:      g.state.drawRound,
		MatchWinnerIdx: g.state.matchWinnerIdx,
		Result:         g.state.result,
		GameEndFlag:    g.state.gameEndFlag,
		Scored:         g.state.scored,
		ActionLog:      g.state.actionLog,
	})
}

// sevenTwentySevenValidPhase は有効なフェーズかどうか。
func sevenTwentySevenValidPhase(p SevenTwentySevenPhase) bool {
	return p == SevenTwentySevenPhaseDraw || p == SevenTwentySevenPhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *SevenTwentySeven) UnmarshalJSON(data []byte) error {
	var j sevenTwentySevenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("sevenTwentySeven: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < SevenTwentySevenMinPlayerCount || n > SevenTwentySevenMaxPlayerCount || n != j.Config.PlayerCount {
		return errSevenTwentySevenSnapshot
	}
	if len(j.ActionLog) > sevenTwentySevenMaxSliceLen {
		return errSevenTwentySevenSnapshot
	}
	if !sevenTwentySevenValidPhase(j.Phase) {
		return errSevenTwentySevenSnapshot
	}
	if j.RoundNumber < 1 || j.Pot < 0 || j.CarryPot < 0 {
		return errSevenTwentySevenSnapshot
	}
	if j.LowWinner < -1 || j.LowWinner >= n || j.HighWinner < -1 || j.HighWinner >= n {
		return errSevenTwentySevenSnapshot
	}
	if j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n || j.DrawRound < 0 {
		return errSevenTwentySevenSnapshot
	}
	if j.Result < SevenTwentySevenResultLose || j.Result > SevenTwentySevenResultWin {
		return errSevenTwentySevenSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errSevenTwentySevenSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errSevenTwentySevenSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.state = sevenTwentySevenState{
		phase:          j.Phase,
		roundNumber:    j.RoundNumber,
		pot:            j.Pot,
		carryPot:       j.CarryPot,
		carryCount:     j.CarryCount,
		lowWinner:      j.LowWinner,
		highWinner:     j.HighWinner,
		drawRound:      j.DrawRound,
		matchWinnerIdx: j.MatchWinnerIdx,
		result:         j.Result,
		gameEndFlag:    j.GameEndFlag,
		scored:         j.Scored,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// sevenTwentySevenIndexOf は players の中での p の席番号を返す (見つからなければ -1)。
func sevenTwentySevenIndexOf(players []*SevenTwentySevenPlayer, p *SevenTwentySevenPlayer) int {
	for i, q := range players {
		if q == p {
			return i
		}
	}
	return -1
}

// sevenTwentySevenHandOf は p の手札を切り出す。枚数は引いた回数だけ増える。
func sevenTwentySevenHandOf(p *SevenTwentySevenPlayer) []*Card {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return cards
}
