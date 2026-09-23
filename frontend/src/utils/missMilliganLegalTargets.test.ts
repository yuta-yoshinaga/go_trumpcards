import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, MissMilliganTableauCard } from '../types/card';
import { missMilliganLegalTargets } from './missMilliganLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const column = (...cards: Card[]): MissMilliganTableauCard[] => cards.map((card) => ({ card, faceUp: true }));
const emptyTableau = (): MissMilliganTableauCard[][] => Array.from({ length: 8 }, () => []);
const emptyFoundation = (): Card[][] => Array.from({ length: 8 }, () => []);

describe('missMilliganLegalTargets', () => {
  it('returns no targets without a card', () => {
    const result = missMilliganLegalTargets(emptyTableau(), emptyFoundation(), null);
    expect(result.tableau.size).toBe(0);
    expect(result.foundation.size).toBe(0);
  });

  it('accepts an alternating-colour card and rejects a same-colour card', () => {
    const tableau = emptyTableau();
    tableau[0] = column(card('HEART', 10));
    expect([...missMilliganLegalTargets(tableau, emptyFoundation(), card('SPADE', 9)).tableau]).toEqual([0]);
    expect([...missMilliganLegalTargets(tableau, emptyFoundation(), card('DIAMOND', 9)).tableau]).toEqual([]);
  });

  it('excludes the source tableau column from destinations', () => {
    const tableau = emptyTableau();
    tableau[0] = column(card('HEART', 10));
    tableau[1] = column(card('HEART', 10));
    expect([...missMilliganLegalTargets(tableau, emptyFoundation(), card('SPADE', 9), 0).tableau]).toEqual([1]);
  });

  it('allows only Kings on empty tableau columns', () => {
    const result = missMilliganLegalTargets(emptyTableau(), emptyFoundation(), card('SPADE', 13));
    expect(result.tableau.size).toBe(8);
    expect(missMilliganLegalTargets(emptyTableau(), emptyFoundation(), card('SPADE', 12)).tableau.size).toBe(0);
  });

  it('matches fixed-suit empty foundations and continues an existing foundation', () => {
    const foundation = emptyFoundation();
    foundation[0] = [card('SPADE', 1)];
    const result = missMilliganLegalTargets(emptyTableau(), foundation, card('SPADE', 2));
    expect([...result.foundation]).toEqual([0]);
    expect(missMilliganLegalTargets(emptyTableau(), emptyFoundation(), card('HEART', 1)).foundation.size).toBe(2);
  });

  it('offers foundations only for the top tableau card, while retaining tableau targets for a buried card', () => {
    const tableau = emptyTableau();
    tableau[0] = column(card('SPADE', 9), card('HEART', 8));
    tableau[1] = column(card('HEART', 10));
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 7)];

    const topTargets = missMilliganLegalTargets(tableau, foundation, tableau[0][1].card, 0, 1);
    expect([...topTargets.foundation]).toEqual([2]);

    const buriedTargets = missMilliganLegalTargets(tableau, foundation, tableau[0][0].card, 0, 0);
    expect([...buriedTargets.foundation]).toEqual([]);
    expect([...buriedTargets.tableau]).toEqual([1]);
  });

  it('continues to offer foundations for a waived card', () => {
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 7)];
    const result = missMilliganLegalTargets(emptyTableau(), foundation, card('HEART', 8));
    expect([...result.foundation]).toEqual([2]);
  });
});
