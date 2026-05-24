/**
 * Compute the human player's win equity in Indian Poker given the visible
 * opponent card ranks. Indian Poker is a single-card showdown — the human's
 * own card is hidden, drawn uniformly from the remaining deck.
 *
 * Win condition: human rank > max(opponent ranks). Ties are losses for the
 * highest opponent (cards are unique within a deck, but a defensive guard
 * still treats `>` strictly).
 *
 * Card ranks 1..13 each have 4 copies in a 52-card deck. After removing the
 * opponents' visible cards, the remaining 52 - N cards are equally likely.
 * Equity = (# remaining cards strictly above the max opponent rank) / (52 - N).
 *
 * @param opponentRanks Visible opponent card ranks (1..13). Empty/invalid
 *   input yields equity = 1 (a no-opponent scenario is a guaranteed win).
 * @returns Equity in [0, 1].
 */
export function computeIndianPokerEquity(opponentRanks: readonly number[]): number {
  const valid = opponentRanks.filter((r) => Number.isInteger(r) && r >= 1 && r <= 13);
  if (valid.length === 0) return 1;
  const maxOpp = Math.max(...valid);

  const rankCounts = new Map<number, number>();
  for (const r of valid) {
    rankCounts.set(r, (rankCounts.get(r) ?? 0) + 1);
  }
  let above = 0;
  for (let r = maxOpp + 1; r <= 13; r += 1) {
    above += 4 - (rankCounts.get(r) ?? 0);
  }
  const remaining = 52 - valid.length;
  if (remaining <= 0) return 0;
  return above / remaining;
}
