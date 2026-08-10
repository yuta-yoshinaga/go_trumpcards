/** How many denominations a Vint bid can name: ♠ ♣ ♦ ♥ NT. Mirrors `VintDenomCount`. */
export const VINT_DENOM_COUNT = 5;

/** The standing bid, as sent in `VintResponse.highBid`. */
export interface VintBid {
  denom: number;
  level: number;
}

/**
 * Rank of a (denom, level) pair in Vint's bidding order, mirroring
 * `VintBidRank` in `internal/domain/VintScore.go`:
 *
 * ```go
 * return level*VintDenomCount + denom
 * ```
 *
 * so a higher level always outranks a lower one, and within a level the
 * denominations rise ♠ ♣ ♦ ♥ NT.
 * @param denom - The denomination (0=♠ … 4=NT).
 * @param level - The bid level.
 * @returns The comparable rank.
 */
export function vintBidRank(denom: number, level: number): number {
  return level * VINT_DENOM_COUNT + denom;
}

/**
 * Whether a bid may be made, mirroring the domain's rejection of
 * `VintBidRank(denom, level) <= VintBidRank(high.Denom, high.Level)`.
 *
 * **Equal does not pass**: a bid has to beat the standing one, not match it.
 * With nothing standing, any bid is legal (#4940).
 * @param denom - The denomination being considered.
 * @param level - The level being considered.
 * @param high - The standing bid, or null when nobody has bid.
 * @returns Whether the server would accept it.
 */
export function vintBidBeats(denom: number, level: number, high: VintBid | null | undefined): boolean {
  if (!high) return true;
  return vintBidRank(denom, level) > vintBidRank(high.denom, high.level);
}

/**
 * Whether any denomination at this level still beats the standing bid.
 *
 * A level is not simply "above the standing level": the standing bid's own
 * level stays open through the denominations above it, so ♠4 leaves ♣4 … NT4
 * playable.
 * @param level - The level being considered.
 * @param high - The standing bid, or null when nobody has bid.
 * @returns Whether the level has at least one legal denomination.
 */
export function vintLevelHasLegalBid(level: number, high: VintBid | null | undefined): boolean {
  return vintBidBeats(VINT_DENOM_COUNT - 1, level, high);
}

/**
 * The cheapest bid that still beats the standing one: the very next rank up.
 *
 * Used to move the pickers forward when another player outbids the selection
 * underneath them — leaving the button disabled with no hint of what to pick
 * next is a dead end (#4940).
 * @param high - The standing bid, or null when nobody has bid.
 * @param minLevel - The lowest biddable level.
 * @param maxLevel - The highest biddable level.
 * @returns The next legal pair, or null when the ladder is exhausted.
 */
export function vintNextLegalBid(high: VintBid | null | undefined, minLevel: number, maxLevel: number): VintBid | null {
  if (!high) return { denom: 0, level: minLevel };
  const next = vintBidRank(high.denom, high.level) + 1;
  const level = Math.floor(next / VINT_DENOM_COUNT);
  const denom = next % VINT_DENOM_COUNT;
  if (level > maxLevel) return null;
  return { denom, level: Math.max(level, minLevel) };
}
