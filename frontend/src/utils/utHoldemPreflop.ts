import type { Card } from '../types/card';

export type UTHPreflopStrength = 'strong' | 'moderate' | 'weak';

function effectiveRank(value: number): number {
  return value === 1 ? 14 : value;
}

/**
 * Coarse pre-flop strength bucket for Ultimate Texas Hold'em: "strong" recommends a 4x bet,
 * "moderate" recommends a 3x, "weak" recommends Check. Based on the published basic strategy
 * (any pair, any Ace, suited Broadway → 4x; suited connectors / K-anything offsuit → 3x).
 */
export function utHoldemPreflopStrength(hand: Card[]): UTHPreflopStrength {
  if (hand.length < 2) return 'weak';
  const [a, b] = hand;
  const ra = effectiveRank(a.value);
  const rb = effectiveRank(b.value);
  const hi = Math.max(ra, rb);
  const lo = Math.min(ra, rb);
  const suited = a.design === b.design;
  if (ra === rb) return 'strong'; // any pair
  if (hi === 14) return 'strong'; // any ace
  if (hi === 13 && suited) return 'strong'; // suited king
  if (hi === 13 && lo >= 11) return 'strong'; // K-Q/J offsuit
  if (hi === 12 && lo >= 11) return 'strong'; // Q-J either way
  if (hi === 13) return 'moderate'; // K-anything offsuit
  if (suited && hi - lo <= 2 && lo >= 6) return 'moderate'; // suited connector mid+
  if (suited && hi >= 12) return 'moderate'; // Q-x suited
  return 'weak';
}
