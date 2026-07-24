/** Pot-odds figures for the Bouillotte Call/Raise/Fold decision. */
export interface BouillottePotOdds {
  /** Chips the human must add to call (`currentBet - roundBet`, floored at 0). */
  callAmount: number;
  /** True when nothing is owed to stay in (a free check); no odds are shown. */
  isFree: boolean;
  /** Call cost as a share of the resulting pot, as a percentage rounded to 1 decimal. */
  percentage: number;
  /** Simplified pot side of the `pot : call` odds ratio. */
  ratioPot: number;
  /** Simplified call side of the `pot : call` odds ratio. */
  ratioCall: number;
}

/** Greatest common divisor (non-negative integers) for simplifying the odds ratio. */
function gcd(a: number, b: number): number {
  return b === 0 ? a : gcd(b, a % b);
}

/**
 * Computes the pot odds facing the human at the Bouillotte betting decision.
 *
 * The amount to call is `currentBet - roundBet` (chips already wagered this
 * round). Pot odds are `call / (pot + call)`, i.e. the fraction of the resulting
 * pot the call represents, expressed as a percentage and as a simplified
 * `pot : call` ratio. When nothing is owed the result is a free check
 * (`isFree`), for which no odds are meaningful.
 *
 * @param pot - Chips already in the pot (includes every player's wagers so far).
 * @param currentBet - Total per-player contribution required to stay in.
 * @param roundBet - Chips the human has already wagered this round.
 * @returns The pot-odds figures for display.
 */
export function computeBouillottePotOdds(pot: number, currentBet: number, roundBet: number): BouillottePotOdds {
  const callAmount = Math.max(0, currentBet - roundBet);
  if (callAmount <= 0) {
    return { callAmount: 0, isFree: true, percentage: 0, ratioPot: 0, ratioCall: 0 };
  }
  const safePot = Math.max(0, pot);
  const total = safePot + callAmount;
  const percentage = Math.round((callAmount / total) * 1000) / 10;
  const divisor = gcd(safePot, callAmount) || 1;
  return {
    callAmount,
    isFree: false,
    percentage,
    ratioPot: safePot / divisor,
    ratioCall: callAmount / divisor,
  };
}
