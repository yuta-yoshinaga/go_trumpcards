import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { StreetsAndAlleysTableauCard } from '../types/games/streetsandalleys';
import { streetsAndAlleysLegalTargets } from './streetsAndAlleysLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): StreetsAndAlleysTableauCard => ({
  card: card(design, value),
  faceUp: true,
});
const emptyFoundation = (): Card[][] => [[], [], [], []];

describe('streetsAndAlleysLegalTargets', () => {
  it('returns nothing when no card is selected', () => {
    const t = streetsAndAlleysLegalTargets([[tc('SPADE', 7)]], emptyFoundation(), null);
    expect(t.tableau.size).toBe(0);
    expect(t.foundation.size).toBe(0);
  });

  it('accepts a column whose top is one rank higher', () => {
    const t = streetsAndAlleysLegalTargets([[tc('SPADE', 8)]], emptyFoundation(), card('HEART', 7));
    expect(t.tableau.has(0)).toBe(true);
  });

  // ここが本題。以前は選択中なら全列の一番上が無条件で光っていた。
  it('rejects a column whose rank does not line up', () => {
    const t = streetsAndAlleysLegalTargets(
      [[tc('SPADE', 8)], [tc('SPADE', 3)], [tc('HEART', 7)]],
      emptyFoundation(),
      card('HEART', 7),
    );
    expect(t.tableau.has(0)).toBe(true); // 8 の上に 7
    expect(t.tableau.has(1)).toBe(false); // 3 の上に 7 は置けない
    expect(t.tableau.has(2)).toBe(false); // 同ランクも置けない
  });

  // スートも色も見ない。姉妹ゲームの交互色ルールを写していないことの証明。
  it('ignores suit and colour', () => {
    const sameSuit = streetsAndAlleysLegalTargets([[tc('SPADE', 8)]], emptyFoundation(), card('SPADE', 7));
    const sameColour = streetsAndAlleysLegalTargets([[tc('SPADE', 8)]], emptyFoundation(), card('CLOVER', 7));
    const otherColour = streetsAndAlleysLegalTargets([[tc('SPADE', 8)]], emptyFoundation(), card('HEART', 7));
    expect(sameSuit.tableau.has(0)).toBe(true);
    expect(sameColour.tableau.has(0)).toBe(true);
    expect(otherColour.tableau.has(0)).toBe(true);
  });

  it('always accepts an empty column', () => {
    const t = streetsAndAlleysLegalTargets([[], [tc('SPADE', 3)]], emptyFoundation(), card('HEART', 12));
    expect(t.tableau.has(0)).toBe(true);
    expect(t.tableau.has(1)).toBe(false);
  });

  // 組札の移動先はサーバが findFoundation で決めるので、合法な山を全部光らせない。
  it('reports only the first empty foundation for an ace', () => {
    const t = streetsAndAlleysLegalTargets([[tc('SPADE', 3)]], emptyFoundation(), card('CLOVER', 1));
    expect([...t.foundation]).toEqual([0]);
  });

  it('continues a foundation in the same suit only', () => {
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 1)];
    const ok = streetsAndAlleysLegalTargets([[tc('SPADE', 3)]], foundation, card('HEART', 2));
    const wrongSuit = streetsAndAlleysLegalTargets([[tc('SPADE', 3)]], foundation, card('SPADE', 2));
    expect(ok.foundation.has(2)).toBe(true);
    expect(wrongSuit.foundation.has(2)).toBe(false);
  });
});
