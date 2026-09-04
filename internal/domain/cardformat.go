package domain

// cardformat.go は、元々 Hearts.go に定義されていたドメイン共通ユーティリティを切り出したファイルです。
// cardStr は 115 以上のゲーム実装ファイルから参照されており、ハーツ固有のファイルに同居していると
// ハーツをビルドタグで分離した際に他カテゴリで未定義エラーが発生するため、タグ無しの共通ファイルとして配置しています。

// cardStr カードの文字列表現
func cardStr(card *Card) string {
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	values := map[int]string{
		1: "A", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7",
		8: "8", 9: "9", 10: "10", 11: "J", 12: "Q", 13: "K",
	}
	s, ok := suits[card.GetDesign()]
	if !ok {
		s = "?"
	}
	v, ok := values[card.GetValue()]
	if !ok {
		v = "?"
	}
	return s + v
}
