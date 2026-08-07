/**
 * Compute the human player's win equity in Indian Poker given the visible
 * opponent card ranks. Indian Poker is a single-card showdown — the human's
 * own card is hidden, drawn uniformly from the remaining deck.
 *
 * Win condition: human rank > max(opponent ranks). Ties are losses for the
 * highest opponent (cards are unique within a deck, but a defensive guard
 * still treats `>` strictly).
 *
 * **ランクは 2..14。**ドメインの `indianPokerCardRank` がエース(1) を 14 に
 * リマップして返すので、1 は決して来ず 14 が最強になる。ここを 1..13 で
 * 弾いていたため、相手のエースが無効値として計算から丸ごと外れ、最も危険な
 * 場面ほど勝率を高く見せていた (#4690)。
 *
 * Card ranks 2..14 each have 4 copies in a 52-card deck. After removing the
 * opponents' visible cards, the remaining 52 - N cards are equally likely.
 * Equity = (# remaining cards strictly above the max opponent rank) / (52 - N).
 *
 * @param opponentRanks Visible opponent card ranks (2..14; ace is 14). Empty/invalid
 *   input yields equity = 1 (a no-opponent scenario is a guaranteed win).
 * @returns Equity in [0, 1].
 */
export function computeIndianPokerEquity(opponentRanks: readonly number[]): number {
  const valid = opponentRanks.filter((r) => Number.isInteger(r) && r >= 2 && r <= 14);
  if (valid.length === 0) return 1;
  const maxOpp = Math.max(...valid);

  const rankCounts = new Map<number, number>();
  for (const r of valid) {
    rankCounts.set(r, (rankCounts.get(r) ?? 0) + 1);
  }
  let above = 0;
  for (let r = maxOpp + 1; r <= 14; r += 1) {
    above += 4 - (rankCounts.get(r) ?? 0);
  }
  const remaining = 52 - valid.length;
  if (remaining <= 0) return 0;
  return above / remaining;
}
