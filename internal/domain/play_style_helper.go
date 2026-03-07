package domain

// playStyleName プレイスタイル名を返す共通ヘルパー
func playStyleName(index int, names []string) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "Unknown"
}
