import type { Card } from '../types/card';

/**
 * Corner label the backend attaches to a Wizard special card. The Go presenter
 * (`WizardWebPresenter.wizardFace`) sends `label: "Wizard"` for the highest
 * trump card, which is always legal to play.
 */
const WIZARD_LABEL = 'Wizard';
/**
 * Corner label the backend attaches to a Jester special card
 * (`label: "Jester"`). Like the Wizard card it is exempt from the follow-suit
 * obligation, and it is skipped when deriving the trick's led suit.
 */
const JESTER_LABEL = 'Jester';

/**
 * Whether `card` is a Wizard or Jester special card. Both serialize with
 * `design: 'JOKER'` from the backend and are distinguished only by their
 * `label`; either one is always legal to play (mirrors the Wizard/Jester
 * exemption in `Wizard.validatePlay`).
 *
 * @param card - The card to test.
 * @returns `true` for a Wizard or Jester card.
 */
export function isWizardSpecial(card: Card): boolean {
  return card.label === WIZARD_LABEL || card.label === JESTER_LABEL;
}

/**
 * The led suit of the current trick, as a `CardDesign`, or `null` when no suit
 * is established. Mirrors `Wizard.leadSuitOfTrick` in
 * `internal/domain/Wizard.go`: scanning from the first played card, a Wizard
 * means "no led suit" (any card is legal), a Jester is skipped, and the first
 * normal suit card fixes the led suit.
 *
 * @param trick - The cards already played into the current trick, in order.
 * @returns The led suit design, or `null` when the trick is empty, a Wizard led,
 *   or only Jesters have been played so far.
 */
export function wizardLeadSuit(trick: Card[]): Card['design'] | null {
  for (const card of trick) {
    if (card.label === WIZARD_LABEL) return null; // a Wizard led → no led suit
    if (card.label === JESTER_LABEL) continue; // a Jester never sets the suit
    return card.design; // first normal suit card fixes the led suit
  }
  return null;
}

/**
 * Whether `card` from the human hand may be legally played into the current
 * trick. Mirrors `Wizard.validatePlay` in `internal/domain/Wizard.go`:
 *
 * - Leading (empty trick): any card is legal.
 * - A Wizard or Jester card is always legal.
 * - Otherwise, if a led suit is established and the player still holds it, the
 *   card must match that suit; a player void in the led suit may play anything.
 *
 * @param card - The candidate card from the player's hand.
 * @param trick - The cards already played into the current trick, in order.
 * @param hand - The player's full hand, used to test whether they still hold the
 *   led suit.
 * @returns `true` if the card can be legally played.
 */
export function isWizardLegalPlay(card: Card, trick: Card[], hand: Card[]): boolean {
  if (trick.length === 0) return true; // leading: anything is legal
  if (isWizardSpecial(card)) return true; // Wizard/Jester always legal
  const lead = wizardLeadSuit(trick);
  if (lead === null) return true; // no led suit (Wizard led / all Jesters)
  if (card.design === lead) return true; // following the led suit
  return !hand.some((c) => c.design === lead); // void in led suit → any card legal
}
