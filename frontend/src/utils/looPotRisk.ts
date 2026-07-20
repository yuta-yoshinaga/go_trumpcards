/** Number of tricks fought per Loo deal (Five-card Loo). */
const LOO_TRICK_COUNT = 5;

/** Pot/risk figures for the Loo Play/Pass decision. */
export interface LooPotRisk {
  /** Chips currently in the pot at stake this deal (equals `potStart` at the decide phase). */
  pot: number;
  /**
   * Chips a "looed" participant (played but took no trick) pays into the next
   * deal's pot. The domain penalty is exactly `potStart`.
   */
  looPenalty: number;
  /**
   * Chips won per trick captured — one fifth of `potStart`, floored, matching
   * the domain's `share := potStart / LooTrickCount` integer division.
   */
  perTrick: number;
}

/**
 * Computes the pot reward and loo risk facing the human at the Loo Play/Pass
 * decision.
 *
 * A participant wins `floor(potStart / 5)` chips per trick captured (up to the
 * whole pot across all five tricks), while a participant who plays but takes no
 * trick is "looed" and pays a penalty of `potStart` into the next deal's pot.
 * This is neutral informational display, not advice.
 *
 * @param pot - Chips currently in the pot (equals `potStart` at the decide phase).
 * @param potStart - Pot size at the start of the deal; sizes both the per-trick payout and the loo penalty.
 * @returns The pot/risk figures for display.
 */
export function computeLooPotRisk(pot: number, potStart: number): LooPotRisk {
  const safePot = Math.max(0, pot);
  const safeStart = Math.max(0, potStart);
  return {
    pot: safePot,
    looPenalty: safeStart,
    perTrick: Math.floor(safeStart / LOO_TRICK_COUNT),
  };
}
