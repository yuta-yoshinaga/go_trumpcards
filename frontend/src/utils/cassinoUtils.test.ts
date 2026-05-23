import { describe, expect, it } from 'vitest';
import type { Card, CassinoBuild } from '../types/card';
import { isCassinoFaceCard, suggestCassinoAction } from './cassinoUtils';

const card = (value: number): Card => ({ design: 'SPADE', value }) as Card;

const build = (value: number): CassinoBuild => ({
  ownerIdx: 1,
  value,
  groups: [],
  isMulti: false,
});

describe('isCassinoFaceCard', () => {
  it('marks J/Q/K (11/12/13) as face cards', () => {
    expect(isCassinoFaceCard(card(11))).toBe(true);
    expect(isCassinoFaceCard(card(12))).toBe(true);
    expect(isCassinoFaceCard(card(13))).toBe(true);
  });

  it('treats Ace and 2-10 as numeric', () => {
    expect(isCassinoFaceCard(card(1))).toBe(false);
    expect(isCassinoFaceCard(card(2))).toBe(false);
    expect(isCassinoFaceCard(card(10))).toBe(false);
  });
});

describe('suggestCassinoAction', () => {
  it('suggests take when table sum equals hand value', () => {
    const result = suggestCassinoAction({
      handCard: card(9),
      hand: [card(9)],
      handIndex: 0,
      selectedTableCards: [card(4), card(5)],
      selectedBuilds: [],
    });
    expect(result).toEqual({ type: 'take', reason: 'sum', value: 9 });
  });

  it('suggests take when single table card equals hand value', () => {
    const result = suggestCassinoAction({
      handCard: card(7),
      hand: [card(7)],
      handIndex: 0,
      selectedTableCards: [card(7)],
      selectedBuilds: [],
    });
    expect(result).toEqual({ type: 'take', reason: 'sum', value: 7 });
  });

  it('suggests build when hand + table sum equals another hand card value', () => {
    const result = suggestCassinoAction({
      handCard: card(4),
      hand: [card(4), card(9)],
      handIndex: 0,
      selectedTableCards: [card(5)],
      selectedBuilds: [],
    });
    expect(result).toEqual({ type: 'build', declaredValue: 9 });
  });

  it('does not suggest build when no matching capture card is in hand', () => {
    const result = suggestCassinoAction({
      handCard: card(4),
      hand: [card(4), card(2)],
      handIndex: 0,
      selectedTableCards: [card(5)],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });

  it('does not suggest build when declared value would not exceed hand value', () => {
    const result = suggestCassinoAction({
      handCard: card(5),
      hand: [card(5), card(5)],
      handIndex: 0,
      selectedTableCards: [],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });

  it('caps build declared value at 10', () => {
    const result = suggestCassinoAction({
      handCard: card(7),
      hand: [card(7), card(11)],
      handIndex: 0,
      selectedTableCards: [card(4)],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });

  it('suggests face-match take when every selected table card matches the rank', () => {
    const result = suggestCassinoAction({
      handCard: card(12),
      hand: [card(12)],
      handIndex: 0,
      selectedTableCards: [card(12), card(12)],
      selectedBuilds: [],
    });
    expect(result).toEqual({ type: 'take', reason: 'faceMatch', value: 12 });
  });

  it('returns null for face card with mismatched table selection', () => {
    const result = suggestCassinoAction({
      handCard: card(12),
      hand: [card(12)],
      handIndex: 0,
      selectedTableCards: [card(11)],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });

  it('suggests build-match take when selected build matches the played card', () => {
    const result = suggestCassinoAction({
      handCard: card(8),
      hand: [card(8)],
      handIndex: 0,
      selectedTableCards: [],
      selectedBuilds: [build(8)],
    });
    expect(result).toEqual({ type: 'take', reason: 'buildMatch', value: 8 });
  });

  it('returns null when selected build value does not match the played card', () => {
    const result = suggestCassinoAction({
      handCard: card(8),
      hand: [card(8)],
      handIndex: 0,
      selectedTableCards: [],
      selectedBuilds: [build(7)],
    });
    expect(result).toBeNull();
  });

  it('returns null when a build-match selection mixes a face card on the table', () => {
    const result = suggestCassinoAction({
      handCard: card(8),
      hand: [card(8)],
      handIndex: 0,
      selectedTableCards: [card(8), card(11)],
      selectedBuilds: [build(8)],
    });
    expect(result).toBeNull();
  });

  it('returns null when no table or builds are selected', () => {
    const result = suggestCassinoAction({
      handCard: card(8),
      hand: [card(8)],
      handIndex: 0,
      selectedTableCards: [],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });

  it('returns null when selected table contains a face card with numeric hand card', () => {
    const result = suggestCassinoAction({
      handCard: card(5),
      hand: [card(5)],
      handIndex: 0,
      selectedTableCards: [card(11)],
      selectedBuilds: [],
    });
    expect(result).toBeNull();
  });
});
