package domain

// This file carries no build tag on purpose, so every Cloudflare Worker bucket
// can reach GameResult.
//
// It used to live in BlackJack.go, which is tagged `!js || !wasm || casino`.
// Win/lose/draw is not a blackjack concept, and the tag made it unreachable from
// the other buckets, so twelve games each declared their own three-value copy --
// AnacondaResult, BouillotteResult, CegoResult, FrenchTarotResult, GoStopResult,
// GutsResult, KoenigrufenResult, OichoKabuResult, PrimeroResult, ScartoResult,
// TrenteEtQuaranteResult, WattenResult -- nine of them with a comment saying
// outright that the casino tag was the reason, and one noting its values are
// identical to GameResult's.
//
// Nothing is gained by tagging this: TinyGo's dead-code elimination drops the
// type from any worker that does not reference it, so an untagged home costs no
// binary size while removing the reason to duplicate. Consolidating the twelve
// existing copies is separate work -- their names appear in JSON payloads.

// GameResult ゲーム勝敗結果
type GameResult int

// GameResult定数
const (
	// GameResultWin 勝利
	GameResultWin GameResult = 1
	// GameResultDraw 引き分け
	GameResultDraw GameResult = 0
	// GameResultLose 敗北
	GameResultLose GameResult = -1
)
