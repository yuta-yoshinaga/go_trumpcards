package controller

import (
	"strconv"
	"strings"
)

// parseIntList converts CLI args (each possibly comma-separated) into an int
// slice, skipping non-numeric tokens; empty input returns nil. It is a generic
// CUI helper shared by several games (GinRummy, Cribbage, SevenBridge, ...)
// across different Cloudflare-worker buckets, so it lives in this untagged file
// to be available in every worker binary.
func parseIntList(args []string) []int {
	if len(args) == 0 {
		return nil
	}
	var result []int
	for _, a := range args {
		for _, s := range strings.Split(a, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			v, err := strconv.Atoi(s)
			if err == nil {
				result = append(result, v)
			}
		}
	}
	return result
}
