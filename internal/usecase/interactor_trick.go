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
