import type { SkatResponse } from '../types/games/skat';

/** The score breakdown a finished Skat round carries. */
type SkatScoreBreakdown = NonNullable<SkatResponse['scoreBreakdown']>;

/**
 * Which bonuses raised the round's multiplier, in the order they are announced.
 *
 * The order is fixed rather than derived from the object so the same contract
 * always reads the same way; a caller renders the labels it gets back.
 *
 * @param bd - The breakdown from the server.
 * @returns The translation keys of the bonuses that apply, possibly empty.
 */
export function skatScoreBonusKeys(bd: SkatScoreBreakdown): string[] {
  const keys: string[] = [];
  if (bd.hand) keys.push('bonus.hand');
  if (bd.schneider) keys.push('bonus.schneider');
  if (bd.schwarz) keys.push('bonus.schwarz');
  if (bd.doubled) keys.push('bonus.doubled');
  if (bd.overbid) keys.push('bonus.overbid');
  return keys;
}
