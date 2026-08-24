//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HorseWebConfig は H.O.R.S.E. の Web 設定。
type HorseWebConfig struct {
	Seats              *int `json:"seats,omitempty"`
	InitialChips       *int `json:"initialChips,omitempty"`
	HandsPerDiscipline *int `json:"handsPerDiscipline,omitempty"`
}

// ToConfig は HorseWebConfig を domain.HorseConfig に変換する (境界チェック付き)。
//
// **席数は 4/6/9 しか選べない。** 種目側の卓サイズと同じものしか作れないので、
// 範囲で丸めると 5 のような作れない数が通ってしまう ── 一覧に無い数は既定へ落とす。
func (c *HorseWebConfig) ToConfig() domain.HorseConfig {
	cfg := domain.DefaultHorseConfig()
	if c.Seats != nil && domain.HorseValidSeats(*c.Seats) {
		cfg.Seats = *c.Seats
	}
	cfg.InitialChips = webutil.BoundedIntPtr(c.InitialChips,
		domain.HorseMinChips, domain.HorseMaxChips, cfg.InitialChips)
	cfg.HandsPerDiscipline = webutil.BoundedIntPtr(c.HandsPerDiscipline,
		domain.HorseMinHandsPerDiscipline, domain.HorseMaxHandsPerDiscipline, cfg.HandsPerDiscipline)
	return cfg
}

// HorseWebInput は H.O.R.S.E. の Web インプット。
type HorseWebInput struct {
	BaseWebInput
	// Action ベッティングアクション ("fold" / "check" / "call" / "bet" / "raise" / "allin")
	Action *string `json:"action,omitempty"`
	// Amount ベット/レイズ額
	Amount *int `json:"amount,omitempty"`
	// CardIndices 引き直しで交換する札の位置 (0 始まり)。空 = スタンドパット。
	CardIndices []int `json:"cardIndices,omitempty"`
	// Config ゲーム設定
	Config *HorseWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.HorseConfig を構築する。
func (p HorseWebInput) ToConfig() domain.HorseConfig {
	return configOrDefault(p.Config, (*HorseWebConfig).ToConfig, domain.DefaultHorseConfig())
}

// HorseWebOutputSeat は 1 席の出力。
type HorseWebOutputSeat struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	// Cards はその席から見えている札 (CPU は表向きのみ)。
	Cards []*WebOutputCard `json:"cards"`
}

// HorseWebOutputConfig は設定アウトプット。
type HorseWebOutputConfig struct {
	Seats              int `json:"seats"`
	InitialChips       int `json:"initialChips"`
	HandsPerDiscipline int `json:"handsPerDiscipline"`
}

// HorseWebOutput は H.O.R.S.E. の Web アウトプット。
//
// **出すのは打つのに要るものだけ。** どの種目の何ハンド目か、席の残高、そして
// 見えている札 ── 役の判定や勝敗の内訳は種目側の実装が持っているので、ここへ
// 写すと 5 種目ぶんの表示を二重に持つことになる。
type HorseWebOutput struct {
	Seats            []*HorseWebOutputSeat `json:"seats"`
	Phase            int                   `json:"phase"`
	Discipline       int                   `json:"discipline"`
	DisciplineLetter string                `json:"disciplineLetter"`
	DisciplineName   string                `json:"disciplineName"`
	HandInDiscipline int                   `json:"handInDiscipline"`
	HandNumber       int                   `json:"handNumber"`
	CurrentTurn      int                   `json:"currentTurn"`
	HumanSeat        int                   `json:"humanSeat"`
	IsHumanTurn      bool                  `json:"isHumanTurn"`
	CommunityCards   []*WebOutputCard      `json:"communityCards"`
	Pot              int                   `json:"pot"`
	// ToCall はコールに要する額 (0 ならチェック可)。
	ToCall int `json:"toCall"`
	// MinRaise はいまの種目が受け付ける最小のレイズ幅。
	MinRaise    int  `json:"minRaise"`
	TablePhase  int  `json:"tablePhase"`
	GameEndFlag bool `json:"gameEndFlag"`
	WinnerSeat  int  `json:"winnerSeat"`
	// Variant は回しているローテーション (0=H.O.R.S.E. / 1=Eight-Game Mix)。
	//
	// **画面がルート名から推測してはいけない。** 2 つのゲームが 1 つのページを
	// 共有しているので、種目数も席数の選択肢もサーバーの答えで決まる。
	Variant int `json:"variant"`
	// Rotation は回す種目の並び。
	Rotation []int `json:"rotation"`
	// IsDrawPhase は引き直しを待っているか (2-7 トリプルドローのみ)。
	IsDrawPhase bool `json:"isDrawPhase"`
	// DrawIndex は何回目の引き直しか (1..3)。ドロー中でなければ 0。
	DrawIndex int `json:"drawIndex"`
	WebOutputBase
	Config HorseWebOutputConfig `json:"config"`
}

// horseParseAction はアクション文字列を種目のアクション番号に変換する。
//
// 種目側は Holdem の定数をそのまま使う (Stud も同じ並び)。
func horseParseAction(s string) (int, bool) {
	switch s {
	case "fold", "f":
		return domain.HoldemActionFold, true
	case "check", "x":
		return domain.HoldemActionCheck, true
	case "call", "c":
		return domain.HoldemActionCall, true
	case "bet", "b":
		return domain.HoldemActionBet, true
	case "raise", "r":
		return domain.HoldemActionRaise, true
	case "allin", "a":
		return domain.HoldemActionAllIn, true
	default:
		return 0, false
	}
}

// HorseWebController は H.O.R.S.E. の Web コントローラー。
type HorseWebController = GameWebController[usecase.HorseInteractorIF, HorseWebInput, *HorseWebOutput]

// NewHorseWebController, NewHorseWebControllerWithProvider are the standard and
// provider-backed constructors for HorseWebController.
var NewHorseWebController, NewHorseWebControllerWithProvider = webControllerPair[usecase.HorseInteractorIF, HorseWebInput, *HorseWebOutput](
	newHorseDefaultOutput, horseDispatch,
)

func newHorseDefaultOutput(msg string) *HorseWebOutput {
	return &HorseWebOutput{
		Seats:          make([]*HorseWebOutputSeat, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		WinnerSeat:     -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func horseDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HorseInteractorIF, param HorseWebInput, newDefault func(string) *HorseWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, hi.ResetWithConfig(param.ToConfig()))
	case "a", "action":
		if !requireParam(bc, w, newDefault, param.Action == nil, "param error: action is required.") {
			return true
		}
		act, ok := horseParseAction(*param.Action)
		if !requireParam(bc, w, newDefault, !ok,
			"param error: action must be fold, check, call, bet, raise or allin.") {
			return true
		}
		// **ベットとレイズは額が要る。** 未送信を 0 として通すと、金額を
		// 付け忘れた要求がドメインの汎用エラーになって理由が分からなくなる。
		needsAmount := act == domain.HoldemActionBet || act == domain.HoldemActionRaise
		if !requireParam(bc, w, newDefault, needsAmount && param.Amount == nil,
			"param error: amount is required for bet and raise.") {
			return true
		}
		amount := 0
		if param.Amount != nil {
			amount = *param.Amount
		}
		bc.writePresenterResponse(w, hi.Action(act, amount, 0))
	case "d", "draw", "exchange":
		// **札を 1 枚も指定しない要求はスタンドパット。** 「引かない」も手なので
		// 空を弾いてはいけない。
		bc.writePresenterResponse(w, hi.Exchange(param.CardIndices))
	case "n", "next", "nexthand":
		bc.writePresenterResponse(w, hi.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, hi.Hint, hi.ActionLog)
	}
	return true
}
