import type { Card, TienLenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 値の強さ: 3 が最弱、2 が最強 (sync: `tienLenValueStrength`, internal/domain/TienLenEval.go:12)。 */
const TWO_STRENGTH = 12;
const ACE_STRENGTH = 11;

/** スート順 ♠ < ♣ < ♦ < ♥ (sync: `tienLenSuitStrength`)。 */
const SUIT_STRENGTH: Record<string, number> = { SPADE: 0, CLOVER: 1, DIAMOND: 2, HEART: 3 };

/** 合成の強さは `値 * 4 + スート` (sync: TienLenEval.go:9)。 */
const SUIT_SPAN = 4;

/**
 * Returns a frontend {@link HintResult} for Tiến Lên, or null when no suggestion
 * is available.
 *
 * There is no server-side GetHint, so the ordering is ported here.
 *
 * What makes this game different from the other shedding games is that **the
 * suit is part of the rank**. Strength is `value * 4 + suit` with
 * ♠ < ♣ < ♦ < ♥ (`TienLenEval.go:9`), so two cards of the same face are not
 * interchangeable — the three of spades is the single weakest card in the deck
 * and the two of hearts the single strongest. A hint that compared only values
 * would call them equal and point at the wrong one half the time.
 *
 * The advice itself is the ordinary shedding one: lead the weakest card, and
 * when following play the cheapest thing that still beats the table.
 */
export function getTienLenHint(state: TienLenResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  if (state.tableCards.length === 0) {
    return {
      targetAction: `card-${weakestIndex(human.cards)}`,
      reason: 'frontendHint.tienlenLead',
      confidence: 'moderate',
    };
  }

  const beat = weakestBeating(human.cards, state.tableCards);
  if (beat < 0) {
    return { targetAction: 'pass', reason: 'frontendHint.tienlenPass', confidence: 'moderate' };
  }
  return { targetAction: `card-${beat}`, reason: 'frontendHint.tienlenFollow', confidence: 'moderate' };
}

/** 値とスートを合わせた強さ。 */
function strength(c: Card): number {
  const v = c.value === 2 ? TWO_STRENGTH : c.value === 1 ? ACE_STRENGTH : c.value - 3;
  return v * SUIT_SPAN + (SUIT_STRENGTH[c.design] ?? 0);
}

/** 一番弱い札の位置。 */
function weakestIndex(hand: Card[]): number {
  let best = 0;
  for (let i = 1; i < hand.length; i += 1) {
    if (strength(hand[i]) < strength(hand[best])) best = i;
  }
  return best;
}

/**
 * 場札より強い札のうち一番弱いものの位置。無ければ -1。
 *
 * 場札の枚数（ペアや連番）は見ていない。どの組を出すかはプレイヤーが選ぶので、
 * ここでは「その強さを超える札がそもそもあるか」だけを答える。無ければパスしか
 * ない、という判断は枚数に依らず正しい。
 */
function weakestBeating(hand: Card[], table: Card[]): number {
  // 場の代表は一番強い札。組の強さはそれで決まる。
  const target = Math.max(...table.map(strength));
  let best = -1;
  for (let i = 0; i < hand.length; i += 1) {
    if (strength(hand[i]) <= target) continue;
    if (best < 0 || strength(hand[i]) < strength(hand[best])) best = i;
  }
  return best;
}
