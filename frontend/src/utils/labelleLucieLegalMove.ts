import type { Card } from '../types/card';

/** Top (movable) card of a fan, or undefined when the fan is empty. */
function fanTop(fan: Card[]): Card | undefined {
  return fan.length > 0 ? fan[fan.length - 1] : undefined;
}

/** Whether the card can move to any foundation: an empty pile takes an Ace,
 * otherwise it builds up in the same suit (value + 1). */
function fitsFoundation(card: Card, foundation: Card[][]): boolean {
  for (const pile of foundation) {
    if (pile.length === 0) {
      if (card.value === 1) return true;
      continue;
    }
    const top = pile[pile.length - 1];
    if (card.design === top.design && card.value === top.value + 1) return true;
  }
  return false;
}

/** Whether the card can stack on a fan top: same suit, one rank lower. */
function canStack(card: Card, dstTop: Card | undefined): boolean {
  if (!dstTop) return false;
  return card.design === dstTop.design && card.value === dstTop.value - 1;
}

/**
 * Reports whether any legal move exists in La Belle Lucie — a fan top that can
 * go to a foundation or stack onto another fan's top. Mirrors the domain's
 * hasAnyLegalMove so the UI can recommend a redeal before the player is stuck.
 */
export function labelleLucieHasLegalMove(fans: Card[][], foundation: Card[][]): boolean {
  return labelleLucieMovableFans(fans, foundation).size > 0;
}

/**
 * Indices of the fans whose top card can move somewhere right now — onto a
 * foundation, or onto another fan's top.
 *
 * `labelleLucieHasLegalMove` is the same question asked of the whole board, so
 * it is answered from this set: one rule, two callers (#5678).
 *
 * @param fans - The fans, each ordered bottom to top.
 * @param foundation - The foundation piles.
 * @returns The set of movable fan indices.
 */
export function labelleLucieMovableFans(fans: Card[][], foundation: Card[][]): Set<number> {
  const movable = new Set<number>();
  for (let i = 0; i < fans.length; i++) {
    const card = fanTop(fans[i]);
    if (!card) continue;
    if (fitsFoundation(card, foundation)) {
      movable.add(i);
      continue;
    }
    for (let j = 0; j < fans.length; j++) {
      if (i !== j && canStack(card, fanTop(fans[j]))) {
        movable.add(i);
        break;
      }
    }
  }
  return movable;
}
