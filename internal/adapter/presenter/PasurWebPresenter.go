//go:build !js || !wasm || extra

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PasurWebPresenter パスールWebプレゼンタークラス
type PasurWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PasurWebPresenter) Output(s interfaces.PasurGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = pasurHintOutput(s)
	return marshalOrError(resObj)
}

// pasurHintOutput はヒントを出力形に変換する。
func pasurHintOutput(s interfaces.PasurGame) *controller.PasurWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.PasurWebOutputHint{
		CardIndex: hint.CardIndex, Reason: hint.Reason,
		Table: intSliceOrEmpty(hint.TableIndices),
	}
}

// buildBase 共通フィールドを構築
func (p *PasurWebPresenter) buildBase(s interfaces.PasurGame) *controller.PasurWebOutput {
	resObj := new(controller.PasurWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.Table = cardsToOutputOrEmpty(s.GetTableCards())
	resObj.CaptureOptions = pasurCaptureOptions(s)
	resObj.DeckRemaining = s.GetDeckRemaining()
	resObj.PacksDealt = s.GetPacksDealt()
	resObj.LastCaptureIdx = s.GetLastCaptureIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.Winners = intSliceOrEmpty(s.GetWinners())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.PasurWebOutputConfig{PlayerCnt: s.GetConfig().PlayerCnt}
	return resObj
}

// pasurCaptureOptions は人間の手札ごとの捕獲候補を返す。
//
// **11 の部分集合をページ側で作り直さない。** 作り直せば必ずズレて、サーバが
// 拒否する組み合わせを送ることになります。
func pasurCaptureOptions(s interfaces.PasurGame) [][][]int {
	out := make([][][]int, 0)
	human := s.GetPlayer(0)
	if human == nil {
		return out
	}
	for i := 0; i < human.GetCardsSize(); i++ {
		opts := s.GetCaptureOptions(0, i)
		normalised := make([][]int, 0, len(opts))
		for _, o := range opts {
			normalised = append(normalised, intSliceOrEmpty(o))
		}
		out = append(out, normalised)
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PasurWebPresenter) buildPlayersOutput(s interfaces.PasurGame) []*controller.PasurWebOutputPlayer {
	out := make([]*controller.PasurWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.PasurWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.GetCapturedCount(),
			Soors:         player.GetSoors(),
			Score:         s.GetScore(i),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PasurWebPresenter) buildMessage(s interfaces.PasurGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winners := s.GetWinners()
		switch {
		case len(winners) > 1:
			return "", "pasur.result.tie", map[string]string{"n": strconv.Itoa(len(winners))}
		case len(winners) == 1 && winners[0] == 0:
			return "", "pasur.result.you", nil
		case len(winners) == 1:
			return "", "pasur.result.cpu", map[string]string{"idx": strconv.Itoa(winners[0])}
		default:
			return "", "pasur.result.tie", map[string]string{"n": "0"}
		}
	}
	// **場に何が残っているかがこのゲームの情報のすべて。**
	return "", "pasur.play", map[string]string{
		"table": strconv.Itoa(len(s.GetTableCards())),
		"deck":  strconv.Itoa(s.GetDeckRemaining()),
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *PasurWebPresenter) HintOutput(s interfaces.PasurGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = pasurHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PasurWebPresenter) ActionLogOutput(s interfaces.PasurGame) string {
	return actionLogOutputJSON(s)
}
