package usecase

// trickGame トリックテイキングゲーム共通インタフェース
type trickGame[P comparable] interface {
	GetGameEndFlag() bool
	GetPhase() P
	IsHumanTurn() bool
	CpuPlay()
	ResolveTrick()
	NextTrick()
}

// trickPhases トリックテイキングゲームのフェーズ定数
type trickPhases[P comparable] struct {
	play     P
	trickEnd P
	roundEnd P
	gameEnd  P
}

// bidGame ビッドフェーズを持つゲーム共通インタフェース
type bidGame[P comparable] interface {
	GetGameEndFlag() bool
	GetPhase() P
	IsHumanBidTurn() bool
	CpuBid()
}

// runCpuBidsLoop ビッドフェーズでCPUのビッドを自動実行するループ。
// bidPhase に該当するフェーズの間、人間の手番になるまでCPUにビッドさせる。
func runCpuBidsLoop[P comparable](g bidGame[P], bidPhase P) {
	// runCpuTurnsLoop と同じ理由で上限を持つ。ビッドが進まないまま回ると
	// goroutine が戻らない (#5416)。
	for turns := 0; !g.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if g.GetPhase() != bidPhase {
			break
		}
		if g.IsHumanBidTurn() {
			break
		}
		g.CpuBid()
	}
}

// maxCpuTurnsPerCall は 1 回の呼び出しで回す CPU ターンの上限。
//
// **進まない CpuPlay からループを守るため。**#4606 で、手札の尽きた席に手番が
// 回ると playCard が nil を触ってサーバごと落ちることが分かった。各ゲームの
// CpuPlay に「札が無ければ何もしない」ガードを入れたが、このループは phase と
// IsHumanTurn しか見ないので、**何もしないまま同じ状態で呼ばれ続ける**——
// パニックがハングに変わる (#4607 のレビュー指摘)。
//
// 1 ラウンドの CPU プレイは、最も多いゲームでも 52 枚 ÷ プレイヤー数 × トリック
// 数の範囲に収まる。1000 は正常な進行では決して届かない値で、届いたときは
// 盤面が進んでいないことを意味する。
const maxCpuTurnsPerCall = 1000

// runCpuTurnsLoop トリックテイキングゲーム共通のCPUターン実行ループ
func runCpuTurnsLoop[P comparable](g trickGame[P], p trickPhases[P]) {
	for turns := 0; !g.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			// 盤面が進んでいない。ここで抜けないと goroutine が回り続ける。
			return
		}

		phase := g.GetPhase()
		if phase == p.trickEnd || phase == p.roundEnd || phase == p.gameEnd {
			break
		}
		if phase != p.play {
			break
		}
		if g.IsHumanTurn() {
			break
		}
		g.CpuPlay()
		if g.GetPhase() == p.trickEnd {
			g.ResolveTrick()
			if g.GetPhase() == p.roundEnd {
				break
			}
			g.NextTrick()
		}
	}
}
