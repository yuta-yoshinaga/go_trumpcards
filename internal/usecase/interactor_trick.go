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
	for !g.GetGameEndFlag() {
		if g.GetPhase() != bidPhase {
			break
		}
		if g.IsHumanBidTurn() {
			break
		}
		g.CpuBid()
	}
}

// runCpuTurnsLoop トリックテイキングゲーム共通のCPUターン実行ループ
func runCpuTurnsLoop[P comparable](g trickGame[P], p trickPhases[P]) {
	for !g.GetGameEndFlag() {
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
