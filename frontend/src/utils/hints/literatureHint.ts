import type { Card, LiteratureResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (0 = Play)。 */
const PLAY_PHASE = 0;

/** ハーフスートの状態。0 = 未確定。 */
const HALF_SUIT_OPEN = 0;

/**
 * Returns a frontend {@link HintResult} for Literature, or null when no
 * suggestion is available.
 *
 * Literature is a deduction game and the type says so — even a teammate's hand
 * is hidden — so this does not try to guess where a card is. It covers the two
 * rules that decide whether a move is *possible* at all:
 *
 *  - you may only ask for a card in a half-suit you already hold one of;
 *  - a claim needs every one of the six, and a wrong one **cancels** the
 *    half-suit for both sides rather than handing it over.
 */
export function getLiteratureHint(state: LiteratureResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  // 自分の札が属する、まだ決まっていないハーフスート。
  const askable = state.halfSuitCards.some(
    (cards, i) => state.halfSuits[i] === HALF_SUIT_OPEN && cards.some((c) => holds(human.cards, c)),
  );
  if (!askable) {
    // **持っていないハーフスートは聞けない。**残るのは宣言だけ。
    return { targetAction: 'claim', reason: 'frontendHint.literatureMustClaim', confidence: 'moderate' };
  }

  // 6 枚すべてを自陣が持っている確信がなければ、外した宣言は**取り消し**になる。
  return { targetAction: 'ask', reason: 'frontendHint.literatureAskHeldSuit', confidence: 'moderate' };
}

/** その札を手札に持っているか。 */
function holds(hand: Card[], c: Card): boolean {
  return hand.some((h) => h.design === c.design && h.value === c.value);
}
