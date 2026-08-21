import type { Card } from '../types/card';

/** A fan holds at most three cards (`ShamrocksFanSize`). */
const FAN_CAP = 3;

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

/**
 * Whether the card can move onto fan `dst`.
 *
 * **This duplicates the server rule** (`Shamrocks.canMoveFanToFan`) -- change it
 * whenever the domain changes. The page used to borrow
 * `labelleLucieMovableFans`, which encodes La Belle Lucie's rule instead: same
 * suit, exactly one rank lower, no fan cap, and an empty fan never a target.
 * All three differ here, so the movable ring missed real moves, could ring a
 * move into a full fan, and the stuck banner fired with legal moves available.
 */
function canMoveOnto(card: Card, dst: Card[]): boolean {
  if (dst.length >= FAN_CAP) return false;
  if (dst.length === 0) return true;
  const top = fanTop(dst);
  if (!top) return false;
  const diff = card.value - top.value;
  return diff === 1 || diff === -1;
}

/**
 * Reports whether any legal move exists in Shamrocks -- a fan top that can go to
 * a foundation or onto another fan. Mirrors the domain's hasAnyLegalMove so the
 * UI can warn before the player is stuck.
 */
export function shamrocksHasLegalMove(fans: Card[][], foundation: Card[][]): boolean {
  return shamrocksMovableFans(fans, foundation).size > 0;
}

/**
 * Indices of the fans whose top card can move somewhere right now.
 *
 * `shamrocksHasLegalMove` is the same question asked of the whole board, so it
 * is answered from this set: one rule, two callers (#5678).
 *
 * @param fans - The fans, each ordered bottom to top.
 * @param foundation - The foundation piles.
 * @returns The set of movable fan indices.
 */
export function shamrocksMovableFans(fans: Card[][], foundation: Card[][]): Set<number> {
  const movable = new Set<number>();
  for (let i = 0; i < fans.length; i++) {
    const card = fanTop(fans[i]);
    if (!card) continue;
    if (fitsFoundation(card, foundation)) {
      movable.add(i);
      continue;
    }
    for (let j = 0; j < fans.length; j++) {
      if (i !== j && canMoveOnto(card, fans[j])) {
        movable.add(i);
        break;
      }
    }
  }
  return movable;
}
