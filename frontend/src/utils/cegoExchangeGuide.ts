/**
 * Cego exchange guidance.
 *
 * During the Cego contract the declarer keeps exactly one card, lays the rest of
 * the hand face-down onto their point pile, and takes the hidden blind (Cego).
 * This pure helper derives the current step of that multi-step procedure from the
 * number of currently selected cards, so the page can render a stepper without
 * introducing any new backend state.
 */

/** A step in the Cego exchange procedure, derived purely from selection counts. */
export interface CegoExchangeGuide {
  /** 1-based index of the step the player is on (1 = pick the keep card, 2 = take the blind). */
  currentStep: number;
  /** Total number of steps in the exchange procedure. */
  totalSteps: number;
  /** How many more cards still need to be selected before the exchange can be confirmed. */
  remaining: number;
  /** Whether the required number of keep-cards has been selected. */
  ready: boolean;
  /** How many cards are laid face-down (hand size minus the kept cards). */
  layDownCount: number;
}

/** Total number of steps surfaced in the Cego exchange stepper. */
const CEGO_EXCHANGE_TOTAL_STEPS = 2;

/**
 * Computes the current step of the Cego exchange from the live selection count.
 *
 * @param selectedCount - Number of cards the player has currently selected.
 * @param keepCount - Number of cards that must be kept (the Cego keep count, normally 1).
 * @param handCount - Number of cards in the declarer's hand before the exchange.
 * @returns The derived {@link CegoExchangeGuide}.
 */
export function cegoExchangeGuide(selectedCount: number, keepCount: number, handCount: number): CegoExchangeGuide {
  const remaining = Math.max(0, keepCount - selectedCount);
  const ready = remaining === 0;
  const layDownCount = Math.max(0, handCount - keepCount);
  return {
    currentStep: ready ? 2 : 1,
    totalSteps: CEGO_EXCHANGE_TOTAL_STEPS,
    remaining,
    ready,
    layDownCount,
  };
}
