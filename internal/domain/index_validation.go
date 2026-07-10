package domain

// validateIndexList validates that indices is a non-empty list of distinct
// in-range positions for a collection of the given size. It is a generic helper
// shared by several games (e.g. SevenBridge, ContractRummy) across different
// Cloudflare-worker buckets, so it lives in this untagged file to be available
// in every worker binary.
func validateIndexList(indices []int, size int) error {
	if len(indices) == 0 {
		return NewDomainError(ErrInvalidCard, "インデックスが空です")
	}
	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= size {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
	}
	return nil
}
