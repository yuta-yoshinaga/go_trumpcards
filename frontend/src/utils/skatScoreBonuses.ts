import type { SkatResponse } from '../types/games/skat';

/** The score breakdown a finished Skat round carries. */
type SkatScoreBreakdown = NonNullable<SkatResponse['scoreBreakdown']>;

/**
 * Which bonuses raised the round's multiplier, in the order they are announced.
 *
 * The order is fixed rather than derived from the object so the same contract
 * always reads the same way; a caller renders the labels it gets back.
 *
 * The doubling on a loss and the overbid replacement are deliberately absent:
 * neither is a multiplier bonus, and both change the equation itself, which
 * {@link skatScoreFormulaKey} states instead.
 *
 * @param bd - The breakdown from the server.
 * @returns The translation keys of the bonuses that apply, possibly empty.
 */
export function skatScoreBonusKeys(bd: SkatScoreBreakdown): string[] {
  const keys: string[] = [];
  if (bd.hand) keys.push('bonus.hand');
  if (bd.schneider) keys.push('bonus.schneider');
  if (bd.schwarz) keys.push('bonus.schwarz');
  return keys;
}

/**
 * Which sentence states this round's arithmetic.
 *
 * **There are three, and picking one is not cosmetic.** A win is
 * `base x multiplier`; a loss is that doubled; an overbid throws the product
 * away and pays `bid x 2`. One fixed sentence would print "11 x 3 = 66" for
 * every lost round, which is false and is exactly the confusion the breakdown
 * exists to remove.
 *
 * @param bd - The breakdown from the server.
 * @returns The i18n key of the line to render.
 */
export function skatScoreFormulaKey(bd: SkatScoreBreakdown): string {
  if (bd.overbid) return 'scoreBreakdownOverbid';
  if (bd.doubled) return 'scoreBreakdownDoubled';
  return 'scoreBreakdown';
}
