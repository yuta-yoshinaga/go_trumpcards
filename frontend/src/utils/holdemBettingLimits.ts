/**
 * Betting-limit keys in the same order as `BettingLimitNames` in
 * `internal/domain/betting.go`. Keep this order synchronized with the Go
 * source; do not change only one side.
 */
export const HOLDEM_BETTING_LIMIT_KEYS = ['fixed', 'potLimit', 'noLimit'] as const;

const HOLDEM_BETTING_LIMIT_NAMES: Record<(typeof HOLDEM_BETTING_LIMIT_KEYS)[number], string> = {
  fixed: 'Fixed',
  potLimit: 'Pot Limit',
  noLimit: 'No Limit',
};

/** Return the English name for a Hold'em betting-limit index. */
export function getHoldemBettingLimitName(index: number): string {
  const key = HOLDEM_BETTING_LIMIT_KEYS[index];
  return key === undefined ? 'Unknown' : HOLDEM_BETTING_LIMIT_NAMES[key];
}
