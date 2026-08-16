package usecase

type playableGame interface {
	gameEndChecker
	IsHumanTurn() bool
}

type gameEndChecker interface {
	GetGameEndFlag() bool
}

type outputPresenter[G any] interface {
	Output(game G, lastErr error) string
}

// guardNotPlayable returns early output if the game has ended or it is not the human's turn.
func guardNotPlayable[G playableGame](game G, p outputPresenter[G]) (string, bool) {
	if game.GetGameEndFlag() || !game.IsHumanTurn() {
		return p.Output(game, nil), true
	}
	return "", false
}

// guardGameEnd returns early output if the game has ended.
func guardGameEnd[G gameEndChecker](game G, p outputPresenter[G]) (string, bool) {
	if game.GetGameEndFlag() {
		return p.Output(game, nil), true
	}
	return "", false
}

// roundAdvancer is a game that can start its next round.
type roundAdvancer interface {
	gameEndChecker
	NextRound()
}

// advanceRound runs the shared "start the next round" step: bail out if the
// game is over, advance, let the CPUs play, then present.
//
// Consolidates 22 byte-identical NextRound implementations across the rummy and
// trick-taking interactors (Burraco, Canasta, Carioca, Chinchon, Conquian,
// ContractRummy, …). They differed only in receiver name and concrete type,
// which is why a name-based search never grouped them — see issue #5368.
//
// runCpu is a callback rather than a method on the interface: each interactor's
// runCpuTurns is unexported, so no interface can name it.
//
// The guard is the reason all 22 bodies were identical: without it a finished
// game deals another round. Order matters too — the CPUs must act on the new
// round, not the one that just ended. Both are pinned by tests.
func advanceRound[G roundAdvancer](game G, p outputPresenter[G], runCpu func()) string {
	if out, blocked := guardGameEnd(game, p); blocked {
		return out
	}
	game.NextRound()
	runCpu()
	return p.Output(game, nil)
}

// MaxCpuIterations は CPU ターンを回すループの反復上限。
//
// **ドメインが局を終わらせないと、上限の無いループは戻らない。** CLI ならプロンプトが
// 返らず、Web ならリクエストが返らない。durak では実際に「山札切れ + 防御が成立しない」
// 配置の循環が 20 万局に 14 局あり (#5414)、その配りに当たった局は永久に終わらなかった。
//
// 通常 1 局で CPU が動くのは数十回なので、1000 に達したらドメイン側のバグ。
const MaxCpuIterations = 1000

// runCpuTurnsUntil は「ゲームが終わる」以外にも抜ける条件があるループ用。stop が
// true を返したら抜ける。stop は play の**前**に読むので、既に抜ける局面で始まった
// 場合は一度も play しない。
//
// 同じ 5 行を 31 箇所に書き写さずヘルパ 1 つにしたのは、上限に当たる分岐が
// **ゲームごとには到達不能**だから。実物の局は 1000 手のはるか手前で phase ガードに
// 当たるので、31 箇所に書き写すとどのテストも実行できない `return` が 31 個できる。
func runCpuTurnsUntil(g gameEndChecker, stop func() bool, play func()) bool {
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() || stop() {
			return true
		}
		play()
	}
	return false
}

// runCpuTurnsCapped はゲームが終わるか人間の手番になるまで play を回す。
//
// 戻り値は上限に当たらずに抜けたかどうか。**false はドメインのバグを意味する**ので、
// 呼び出し側が握り潰さずに扱えるよう返している。
func runCpuTurnsCapped(g playableGame, play func()) bool {
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() || g.IsHumanTurn() {
			return true
		}
		play()
	}
	return false
}
