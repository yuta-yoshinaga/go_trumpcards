import type { Card } from '../types/common';
import type { BiribaMeldData } from '../types/games/biriba';
import { canastaFamilyCardValue } from './canastaFamilyScore';

export type BiribaMeldReason =
  | 'selectCards'
  | 'tooFewCards'
  | 'blackThree'
  | 'sameSuit'
  | 'duplicateRank'
  | 'naturalCards'
  | 'tooManyWilds'
  | 'notConsecutive'
  | 'wildRange'
  | 'initialMinimum'
  | 'discardTopRequired';

export interface BiribaMeldStatus {
  ok: boolean;
  reason?: BiribaMeldReason;
}

/** Mirrors validateBiribaSequence, validateNewMeld, validateMeldAddition, findExistingMeldForCards, minimumMeldValue and CanastaFamilyCardValue in internal/domain/Canasta.go. For the page flow, one group is submitted, so the selected-card total matches Go’s total across all submitted groups. */
export function evaluateBiribaMeld(
  selected: readonly Card[],
  melds: readonly BiribaMeldData[],
  options: { hasInitMeld: boolean; minMeld: number; drewFromDiscard?: boolean; includesDrawnCard?: boolean },
): BiribaMeldStatus {
  if (selected.length === 0) return { ok: false, reason: 'selectCards' };
  if (options.drewFromDiscard && !options.includesDrawnCard) {
    return { ok: false, reason: 'discardTopRequired' };
  }
  const existing = melds.find((meld) => isSequence([...meld.cards, ...selected]).ok);
  const result = isSequence(existing ? [...existing.cards, ...selected] : selected);
  if (!result.ok) return result;
  if (!options.hasInitMeld) {
    const total = selected.reduce((sum, card) => sum + canastaFamilyCardValue(card), 0);
    if (total < options.minMeld) return { ok: false, reason: 'initialMinimum' };
  }
  return { ok: true };
}

const CARD_VALUE_MAX = 13;

function isSequence(cards: readonly Card[]): BiribaMeldStatus {
  if (cards.length < 3) return { ok: false, reason: 'tooFewCards' };
  let naturalCount = 0;
  let wildCount = 0;
  let suit: Card['design'] | undefined;
  let minRank = 0;
  let maxRank = 0;
  const seen = new Set<number>();

  for (const card of cards) {
    if (isWild(card)) {
      wildCount++;
      continue;
    }
    if (isBlackThree(card)) return { ok: false, reason: 'blackThree' };
    naturalCount++;
    if (suit === undefined) suit = card.design;
    else if (card.design !== suit) return { ok: false, reason: 'sameSuit' };
    if (seen.has(card.value)) return { ok: false, reason: 'duplicateRank' };
    seen.add(card.value);
    if (minRank === 0 || card.value < minRank) minRank = card.value;
    if (card.value > maxRank) maxRank = card.value;
  }

  if (naturalCount < 2) return { ok: false, reason: 'naturalCards' };
  if (wildCount > 3) return { ok: false, reason: 'tooManyWilds' };
  const missing = maxRank - minRank + 1 - naturalCount;
  if (missing > wildCount) return { ok: false, reason: 'notConsecutive' };
  const remainingWildCount = wildCount - missing;
  const availableExtension = minRank - 1 + (CARD_VALUE_MAX - maxRank);
  if (remainingWildCount > availableExtension) return { ok: false, reason: 'wildRange' };
  return { ok: true };
}

function isWild(card: Card): boolean {
  return card.design === 'JOKER' || card.value === 2;
}
function isBlackThree(card: Card): boolean {
  return card.value === 3 && (card.design === 'SPADE' || card.design === 'CLOVER');
}
