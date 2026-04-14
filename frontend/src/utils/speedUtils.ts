/** Maximum card value (King = 13). */
const CARD_VALUE_MAX = 13;

/** Check if two card ranks are adjacent (with K↔A wrap-around). */
export function isAdjacentRank(a: number, b: number): boolean {
  const diff = Math.abs(a - b);
  return diff === 1 || diff === CARD_VALUE_MAX - 1;
}
