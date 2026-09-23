package domain

// validateIndexList validates that indices is a non-empty list of distinct
// in-range positions for a collection of the given size. It is a generic helper
// shared by several games (e.g. SevenBridge, ContractRummy) across different
// Cloudflare-worker buckets, so it lives in this untagged file to be available
// in every worker binary.
func validateIndexList(indices []int, size int) error {
	if len(indices) == 0 {
		return NewDomainErrorCode(ErrInvalidCard, "shared.errEmptyIndexList", nil)
	}
	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= size {
			return NewDomainErrorCode(ErrInvalidCard, "shared.errCardIndexOutOfRange", nil)
		}
		if seen[idx] {
			return NewDomainErrorCode(ErrInvalidCard, "shared.errDuplicateCardIndex", nil)
		}
		seen[idx] = true
	}
	return nil
}
