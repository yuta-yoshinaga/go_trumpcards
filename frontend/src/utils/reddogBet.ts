/**
 * Minimum bet amount for Red Dog. Mirrors the backend `RedDogMinBet`
 * constant defined in `internal/domain/RedDog.go`. Kept in a single place so
 * the raise-affordability check has one frontend source of truth and cannot
 * silently diverge from the server-side validation.
 */
export const REDDOG_MIN_BET = 10;

/**
 * Reports whether the player can afford the minimum raise given their current
 * chip count. Mirrors the backend guard `chips < RedDogMinBet`.
 */
export function canRedDogRaise(chips: number): boolean {
  return chips >= REDDOG_MIN_BET;
}
