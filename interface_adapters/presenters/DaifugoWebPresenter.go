package presenters

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// DaifugoWebPresenter 大富豪Webプレゼンタークラス
type DaifugoWebPresenter struct{}

// NewDaifugoWebPresenter コンストラクタ
func NewDaifugoWebPresenter() *DaifugoWebPresenter {
	return &DaifugoWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (dwp *DaifugoWebPresenter) Output(d *entities.Daifugo) interface{} {
	resObj := new(controllers.DaifugoWebOutput)
	resObj.Players = make([]*controllers.DaifugoWebOutputPlayer, 0)
	resObj.CurrentTurn = d.GetCurrentTurn()
	resObj.IsRevolution = d.GetIsRevolution()
	resObj.GameEndFlag = d.GetGameEndFlag()

	lastPlay := d.GetLastPlay()
	resObj.LastPlay = make([]*controllers.DaifugoWebOutputCard, 0)
	for _, card := range lastPlay {
		resObj.LastPlay = append(resObj.LastPlay, dwp.getCardObj(card))
	}

	for i, player := range d.GetPlayers() {
		pObj := new(controllers.DaifugoWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.CardCount = player.GetCardsSize()
		pObj.Rank = player.GetRank()
		pObj.Cards = make([]*controllers.DaifugoWebOutputCard, 0)
		
		// 大富豪の場合、手札は全プレイヤー分見えてもよいかもしれないが
		// ババ抜き等と同様に自分（人間）の分だけ送る。
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, dwp.getCardObj(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	if d.GetGameEndFlag() {
		// 1位のプレイヤーIDを探す
		winnerIdx := -1
		for i, p := range d.GetPlayers() {
			if p.GetRank() == 0 {
				winnerIdx = i
				break
			}
		}

		if winnerIdx == 0 {
			resObj.Message = "ゲーム終了！ あなたの勝ち！"
		} else {
			resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx)
		}
	}

	return resObj
}

// getCardObj カード情報オブジェクト取得
func (dwp *DaifugoWebPresenter) getCardObj(card *entities.Card) *controllers.DaifugoWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controllers.DaifugoWebOutputCard)
	switch card.GetDesign() {
	case entities.CardDesignSpade:
		res.Design = "SPADE"
	case entities.CardDesignClover:
		res.Design = "CLOVER"
	case entities.CardDesignHeart:
		res.Design = "HEART"
	case entities.CardDesignDiamond:
		res.Design = "DIAMOND"
	default:
		res.Design = "JOKER"
	}
	res.Value = card.GetValue()
	return res
}
