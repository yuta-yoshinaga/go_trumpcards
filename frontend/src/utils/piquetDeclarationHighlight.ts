import type { Card, PiquetClaim, PiquetDeclaration } from '../types/card';
import { PiquetDeclarationKind } from '../types/phases';

/** Indices + label describing which cards in the human's hand back the latest declaration. */
export interface DeclarationHighlight {
  cardIndices: number[];
  labelKey: 'meldPoint' | 'meldSequence' | 'meldSet';
  count: number;
  won: boolean;
}

/**
 * Returns the indices of cards in `hand` that constitute the human player's
 * relevant meld for `decl`. Returns `null` when the declaration was a tie
 * or no claim is available (e.g. opponent's winning meld with no claim cards
 * for the human side).
 */
export function declarationHighlight(
  decl: PiquetDeclaration,
  humanIdx: number,
  elderIdx: number,
  hand: Card[],
): DeclarationHighlight | null {
  const cards = relevantClaimCards(decl, humanIdx, elderIdx);
  if (cards.length === 0) return null;

  const indices: number[] = [];
  for (const c of cards) {
    const idx = hand.findIndex((h) => h.design === c.design && h.value === c.value);
    if (idx >= 0 && !indices.includes(idx)) indices.push(idx);
  }
  if (indices.length === 0) return null;

  return {
    cardIndices: indices,
    labelKey: kindToKey(decl.kind),
    count: indices.length,
    won: decl.scoredBy === humanIdx,
  };
}

function relevantClaimCards(decl: PiquetDeclaration, humanIdx: number, elderIdx: number): Card[] {
  const humanClaim = humanIdx === elderIdx ? decl.elderClaim : decl.youngerClaim;
  if (decl.kind === PiquetDeclarationKind.SET && decl.scoredBy === humanIdx && decl.sets && decl.sets.length > 0) {
    return flattenClaims(decl.sets);
  }
  return claimCards(humanClaim);
}

function claimCards(claim: PiquetClaim | undefined): Card[] {
  return claim?.cards ?? [];
}

function flattenClaims(claims: PiquetClaim[]): Card[] {
  const out: Card[] = [];
  for (const c of claims) out.push(...c.cards);
  return out;
}

function kindToKey(kind: number): DeclarationHighlight['labelKey'] {
  switch (kind) {
    case PiquetDeclarationKind.POINT:
      return 'meldPoint';
    case PiquetDeclarationKind.SEQUENCE:
      return 'meldSequence';
    default:
      return 'meldSet';
  }
}
