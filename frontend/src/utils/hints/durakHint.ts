import type { Card, DurakResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Durak phase: attack (current attacker picks an attack card). */
const PHASE_ATTACK = 0;

/** Durak phase: defend (current defender plays or takes). */
const PHASE_DEFEND = 1;

/** Map Card.design to suit key used for trump matching. */
const DESIGN_TO_SUIT: Record<Card['design'], string> = {
  SPADE: 'S',
  CLOVER: 'C',
  HEART: 'H',
  DIAMOND: 'D',
  JOKER: '',
};

/** Card value comparable rank (Ace highest). */
function cardRank(value: number): number {
  return value === 1 ? 14 : value;
}

function isTrump(card: Card, trumpSuit: string): boolean {
  return DESIGN_TO_SUIT[card.design] === trumpSuit;
}

function sortAscending(cards: Card[]): Card[] {
  return [...cards].sort((a, b) => cardRank(a.value) - cardRank(b.value));
}

/** Returns a Durak hint or null. */
export function getDurakHint(state: DurakResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0) return null;
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;

  const isAttacker = state.attackerIdx === humanIdx;
  const isDefender = state.defenderIdx === humanIdx;

  if (state.phase === PHASE_ATTACK && isAttacker) {
    return getAttackHint(human.cards, state);
  }

  if (state.phase === PHASE_DEFEND && isDefender) {
    return getDefendHint(human.cards, state);
  }

  return null;
}

function getAttackHint(hand: Card[], state: DurakResponse): HintResult {
  const nonTrumps = hand.filter((c) => !isTrump(c, state.trumpSuit));

  // Continuation attack: must match a value already on the table.
  if (state.tablePairs.length > 0) {
    const onTable = new Set<number>();
    for (const pair of state.tablePairs) {
      onTable.add(pair.attack.value);
      if (pair.defense) onTable.add(pair.defense.value);
    }
    const matching = hand.filter((c) => onTable.has(c.value));
    if (matching.length === 0) {
      return { targetAction: 'pass', reason: 'hint.passBout', confidence: 'moderate' };
    }
    const matchingNonTrumps = matching.filter((c) => !isTrump(c, state.trumpSuit));
    if (matchingNonTrumps.length > 0) {
      return { targetAction: 'attack', reason: 'hint.attackLowNonTrump', confidence: 'strong' };
    }
    return { targetAction: 'pass', reason: 'hint.passBout', confidence: 'moderate' };
  }

  // Fresh attack: lead with the lowest non-trump; trumps are a last resort.
  if (nonTrumps.length > 0) {
    return { targetAction: 'attack', reason: 'hint.attackLowNonTrump', confidence: 'strong' };
  }
  return { targetAction: 'attack', reason: 'hint.attackOnlyTrumps', confidence: 'moderate' };
}

function getDefendHint(hand: Card[], state: DurakResponse): HintResult {
  const undefended = state.tablePairs.find((p) => p.defense === null);
  if (!undefended) {
    return { targetAction: 'take', reason: 'hint.takeNothingToDefend', confidence: 'moderate' };
  }

  const attack = undefended.attack;
  const attackIsTrump = isTrump(attack, state.trumpSuit);
  const sorted = sortAscending(hand);

  // Prefer lowest same-suit card that beats the attack.
  for (const c of sorted) {
    if (!isTrump(c, state.trumpSuit) && c.design === attack.design && cardRank(c.value) > cardRank(attack.value)) {
      return { targetAction: 'defend', reason: 'hint.defendSameSuit', confidence: 'strong' };
    }
  }

  // Otherwise the lowest trump — unless the attack itself is trump, in which case we need a higher trump.
  for (const c of sorted) {
    if (!isTrump(c, state.trumpSuit)) continue;
    if (attackIsTrump && cardRank(c.value) <= cardRank(attack.value)) continue;
    return { targetAction: 'defend', reason: 'hint.defendWithTrump', confidence: 'moderate' };
  }

  return { targetAction: 'take', reason: 'hint.takeBout', confidence: 'moderate' };
}
