package domain

// HeartsShootTheMoonAlertThreshold は「ムーンを狙っている」と見なす
// ラウンド得点の下限。1 ラウンドの罰点は ♥×13 + ♠Q=13 で最大 26 なので、
// その半分を超えて 1 人が独占した時点で警戒に値する。
const HeartsShootTheMoonAlertThreshold = 13

// HeartsShootTheMoonAlertIdx はムーンを狙っていると見られるプレイヤーの
// インデックスを返す。該当が無ければ ok=false。
//
// 判定は「配られた罰点の合計がしきい値以上」かつ「その全部を 1 人が持って
// いる」こと。Web 側 (frontend/src/utils/heartsShootMoonAlert.ts) と同じ規則で、
// 両者がずれないように domain 側にも同じ判定を置いている。
//
// **負のラウンド得点があるときは判定しない。** Omnibus の J♦ ボーナス (-10) は
// roundScore に畳み込まれるため、負の数からは「罰点を取っていない」とは言えず、
// 誤検知するくらいなら黙る。
func HeartsShootTheMoonAlertIdx(roundScores []int) (int, bool) {
	total := 0
	leaderIdx := -1
	leaderScore := 0
	for i, s := range roundScores {
		if s < 0 {
			return 0, false
		}
		if s == 0 {
			continue
		}
		total += s
		if s > leaderScore {
			leaderScore, leaderIdx = s, i
		}
	}
	if leaderIdx < 0 || total < HeartsShootTheMoonAlertThreshold || leaderScore < total {
		return 0, false
	}
	return leaderIdx, true
}
