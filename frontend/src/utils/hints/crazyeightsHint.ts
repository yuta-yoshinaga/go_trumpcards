import type { Card, CrazyEightsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CrazyEightsPhase } from '../../types/phases';

/** Value of 8 — the wild card in Crazy Eights. */
const EIGHT = 8;

/** Map from suit number to Card design string. */
const SUIT_NUM_TO_DESIGN: Record<number, Card['design']> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Returns a frontend HintResult for Crazy Eights, or null if no suggestion. */
export function getCrazyEightsHint(state: CrazyEightsResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;

  if (state.phase === CrazyEightsPhase.CHOOSE_SUIT) {
    return getChooseSuitHint(human.cards);
  }

  if (state.phase === CrazyEightsPhase.PLAY) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for choose suit phase: pick the suit with most cards in hand. */
function getChooseSuitHint(cards: Card[]): HintResult {
  const suitCounts = new Map<Card['design'], number>();
  for (const c of cards) {
    if (c.value !== EIGHT) {
      suitCounts.set(c.design, (suitCounts.get(c.design) ?? 0) + 1);
    }
  }

  if (suitCounts.size === 0) {
    return { targetAction: 'chooseSuit', reason: 'hint.chooseMostSuit', confidence: 'moderate' };
  }

  return { targetAction: 'chooseSuit', reason: 'hint.chooseMostSuit', confidence: 'strong' };
}

/** Hint for play phase: play matching card or draw. */
function getPlayHint(cards: Card[], state: CrazyEightsResponse): HintResult {
  const top = state.discardTop;
  if (!top) {
    return { targetAction: 'play', reason: 'hint.playMatchingSuit', confidence: 'moderate' };
  }

  // **8 の後は指定スートだけが通る。**ドメインは chosenSuit > 0 のとき
  // `card.GetDesign() == g.chosenSuit` だけを見て、場札のランクは見ない。
  // ランク一致をここで門番しないと、出せない札を strong で勧める (#4598)。
  const called = state.chosenSuit > 0 ? SUIT_NUM_TO_DESIGN[state.chosenSuit] : undefined;
  const effectiveSuit = called ?? top.design;

  const eights = cards.filter((c) => c.value === EIGHT);
  const nonEights = cards.filter((c) => c.value !== EIGHT);

  const suitMatch = nonEights.some((c) => c.design === effectiveSuit);
  const valueMatch = called === undefined && nonEights.some((c) => c.value === top.value);
  const hasNonEightPlay = suitMatch || valueMatch;

  if (hasNonEightPlay && eights.length > 0) {
    return { targetAction: 'play', reason: 'hint.saveEight', confidence: 'moderate' };
  }

  if (suitMatch) {
    return { targetAction: 'play', reason: 'hint.playMatchingSuit', confidence: 'strong' };
  }

  if (valueMatch) {
    return { targetAction: 'play', reason: 'hint.playMatchingValue', confidence: 'strong' };
  }

  if (eights.length > 0) {
    return { targetAction: 'play', reason: 'hint.playEight', confidence: 'strong' };
  }

  return { targetAction: 'draw', reason: 'hint.drawCard', confidence: 'moderate' };
}
