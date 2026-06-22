/** Number of seats in Hearts. */
const HEARTS_PLAYER_COUNT = 4;

/**
 * Computes the seat index a player passes to, mirroring the backend
 * `passTarget` logic: left = next seat, right = previous, across = opposite,
 * and the no-pass round returns the same seat.
 *
 * @param fromIdx - The passing player's seat index (0-3).
 * @param direction - Pass direction: 0=left, 1=right, 2=across, 3=none.
 * @returns The recipient's seat index.
 */
export function heartsPassTarget(fromIdx: number, direction: number): number {
  switch (direction) {
    case 0: // left
      return (fromIdx + 1) % HEARTS_PLAYER_COUNT;
    case 1: // right
      return (fromIdx + HEARTS_PLAYER_COUNT - 1) % HEARTS_PLAYER_COUNT;
    case 2: // across
      return (fromIdx + 2) % HEARTS_PLAYER_COUNT;
    default: // none
      return fromIdx;
  }
}
