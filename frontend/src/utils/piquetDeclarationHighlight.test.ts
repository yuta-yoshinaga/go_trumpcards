import { describe, expect, it } from 'vitest';
import type { Card, PiquetClaim, PiquetDeclaration } from '../types/card';
import { PiquetDeclarationKind } from '../types/phases';
import { declarationHighlight } from './piquetDeclarationHighlight';

function claim(cards: Card[]): PiquetClaim {
  return { length: cards.length, topRank: 0, pipTotal: 0, suit: 0, cards };
}

const hand: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'SPADE', value: 13 },
  { design: 'SPADE', value: 12 },
  { design: 'HEART', value: 1 },
  { design: 'CLOVER', value: 13 },
  { design: 'DIAMOND', value: 12 },
];

describe('declarationHighlight', () => {
  it('returns elder claim indices when human is elder', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.POINT,
      elderClaim: claim([
        { design: 'SPADE', value: 1 },
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
      ]),
      youngerClaim: claim([]),
      winner: 0,
      scoredBy: 0,
      score: 3,
    };
    expect(declarationHighlight(decl, 0, 0, hand)).toEqual({
      cardIndices: [0, 1, 2],
      labelKey: 'meldPoint',
      count: 3,
      won: true,
    });
  });

  it('uses younger claim cards when human is younger', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.SEQUENCE,
      elderClaim: claim([]),
      youngerClaim: claim([
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
      ]),
      winner: 1,
      scoredBy: 1,
      score: 3,
    };
    expect(declarationHighlight(decl, 1, 0, hand)).toEqual({
      cardIndices: [1, 2],
      labelKey: 'meldSequence',
      count: 2,
      won: true,
    });
  });

  it('returns null on tie when human has no relevant claim', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.POINT,
      elderClaim: claim([]),
      youngerClaim: claim([]),
      winner: -1,
      scoredBy: -1,
      score: 0,
    };
    expect(declarationHighlight(decl, 0, 0, hand)).toBeNull();
  });

  it('flattens all sets when human wins SET kind', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.SET,
      winner: 0,
      scoredBy: 0,
      score: 6,
      sets: [
        claim([
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 1 },
        ]),
        claim([
          { design: 'SPADE', value: 13 },
          { design: 'CLOVER', value: 13 },
        ]),
      ],
    };
    expect(declarationHighlight(decl, 0, 0, hand)?.cardIndices).toEqual([0, 3, 1, 4]);
  });

  it('marks won=false when opponent scored', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.POINT,
      elderClaim: claim([{ design: 'SPADE', value: 1 }]),
      youngerClaim: claim([]),
      winner: 1,
      scoredBy: 1,
      score: 5,
    };
    const out = declarationHighlight(decl, 0, 0, hand);
    expect(out?.won).toBe(false);
    expect(out?.cardIndices).toEqual([0]);
  });

  it('ignores claim cards not currently in hand (post-exchange edge)', () => {
    const decl: PiquetDeclaration = {
      kind: PiquetDeclarationKind.POINT,
      elderClaim: claim([
        { design: 'CLOVER', value: 5 },
        { design: 'SPADE', value: 1 },
      ]),
      youngerClaim: claim([]),
      winner: 0,
      scoredBy: 0,
      score: 2,
    };
    expect(declarationHighlight(decl, 0, 0, hand)?.cardIndices).toEqual([0]);
  });
});
