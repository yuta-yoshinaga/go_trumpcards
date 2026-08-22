//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenTwentySevenWebConfig はセブン・トゥエンティセブン (SevenTwentySeven) の Web 設定。
type SevenTwentySevenWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は SevenTwentySevenWebConfig を domain.SevenTwentySevenConfig に変換する (境界チェック付き)。
func (c *SevenTwentySevenWebConfig) ToConfig() domain.SevenTwentySevenConfig {
	cfg := domain.DefaultSevenTwentySevenConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.SevenTwentySevenMinPlayerCount, domain.SevenTwentySevenMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.SevenTwentySevenMinAnte, domain.SevenTwentySevenMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.SevenTwentySevenMinStartingChips, domain.SevenTwentySevenMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.SevenTwentySevenMinTargetRounds, domain.SevenTwentySevenMaxTargetRounds)
	return cfg
}

// SevenTwentySevenWebInput はセブン・トゥエンティセブン Web インプット。
type SevenTwentySevenWebInput struct {
	BaseWebInput
	// **引く / 止まるはコマンドで表す。** `card` と `stand` の 2 つで、
	// 追加の入力パラメータは要らない。
	Config *SevenTwentySevenWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.SevenTwentySevenConfig を構築する。
func (p SevenTwentySevenWebInput) ToConfig() domain.SevenTwentySevenConfig {
	return configOrDefault(p.Config, (*SevenTwentySevenWebConfig).ToConfig, domain.DefaultSevenTwentySevenConfig())
}

// SevenTwentySevenWebOutputPlayer は 1 プレイヤーの出力。
type SevenTwentySevenWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	Chips   int  `json:"chips"`
	// Standing は「もう引かない」と宣言済みか。
	Standing bool `json:"standing"`
	// LowScore / HighScore は 7 側 / 27 側の得点を表示用の文字列で返す
	// （"6.5" / "21"）。**超過した側は "-"。** 数値で返さないのは、
	// 0.5 刻みを JSON の数値で往復させるとクライアント側で丸めの判断が要るため。
	LowScore  string           `json:"lowScore"`
	HighScore string           `json:"highScore"`
	Out       bool             `json:"out"`
	RoundBet  int              `json:"roundBet"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`

	// WonLow / WonHigh はこのラウンドでその側を取ったか。**両方 true なら総取り。**
	WonLow  bool `json:"wonLow"`
	WonHigh bool `json:"wonHigh"`
}

// SevenTwentySevenWebOutputHint はヒント出力。
type SevenTwentySevenWebOutputHint struct {
	// Draw は 1 枚引くことを勧めるか。false は「止まれ」。
	Draw   bool   `json:"draw"`
	Reason string `json:"reason"`
}

// SevenTwentySevenWebOutputConfig は設定アウトプット。
type SevenTwentySevenWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// SevenTwentySevenWebOutput はセブン・トゥエンティセブン Web アウトプット。
type SevenTwentySevenWebOutput struct {
	Players     []*SevenTwentySevenWebOutputPlayer `json:"players"`
	Phase       int                                `json:"phase"`
	RoundNumber int                                `json:"roundNumber"`
	Pot         int                                `json:"pot"`
	CarryPot    int                                `json:"carryPot"`
	// CarryCount 連続で持ち越された回数 (CUI の sevenTwentySeven.result.carry と同じ値)。
	CarryCount int `json:"carryCount"`
	Ante       int `json:"ante"`
	Chips      int `json:"chips"`
	// **勝者は 2 人いる。** 7 側と 27 側それぞれ (-1 = 該当なし)。
	// 同一人物なら総取り。
	LowWinner      int                             `json:"lowWinner"`
	HighWinner     int                             `json:"highWinner"`
	DrawRound      int                             `json:"drawRound"`
	MatchWinnerIdx int                             `json:"matchWinnerIdx"`
	Result         int                             `json:"result"`
	GameEndFlag    bool                            `json:"gameEndFlag"`
	Hint           *SevenTwentySevenWebOutputHint  `json:"hint,omitempty"`
	Config         SevenTwentySevenWebOutputConfig `json:"config"`
	WebOutputBase
}

// SevenTwentySevenWebController はセブン・トゥエンティセブン Web コントローラークラス。
type SevenTwentySevenWebController = GameWebController[usecase.SevenTwentySevenInteractorIF, SevenTwentySevenWebInput, *SevenTwentySevenWebOutput]

// NewSevenTwentySevenWebController, NewSevenTwentySevenWebControllerWithProvider are the standard and
// provider-backed constructors for SevenTwentySevenWebController.
var NewSevenTwentySevenWebController, NewSevenTwentySevenWebControllerWithProvider = webControllerPair[usecase.SevenTwentySevenInteractorIF, SevenTwentySevenWebInput, *SevenTwentySevenWebOutput](
	newSevenTwentySevenDefaultOutput, sevenTwentySevenDispatch,
)

func newSevenTwentySevenDefaultOutput(msg string) *SevenTwentySevenWebOutput {
	return &SevenTwentySevenWebOutput{
		Players:        make([]*SevenTwentySevenWebOutputPlayer, 0),
		LowWinner:      -1,
		HighWinner:     -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func sevenTwentySevenDispatch(bc *baseController, w http.ResponseWriter, ti usecase.SevenTwentySevenInteractorIF, param SevenTwentySevenWebInput, newDefault func(string) *SevenTwentySevenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "card", "c":
		bc.writePresenterResponse(w, ti.TakeCard(true))
	case "stand", "s":
		bc.writePresenterResponse(w, ti.TakeCard(false))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
