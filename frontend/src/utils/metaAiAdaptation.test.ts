import { describe, expect, it } from 'vitest';
import {
  type AdaptationLevel,
  deriveAdaptationLevel,
  deriveStrategyStyle,
  type StrategyStyle,
} from './metaAiAdaptation';

describe('deriveAdaptationLevel', () => {
  const cases: { gamesPlayed: number; expected: AdaptationLevel }[] = [
    { gamesPlayed: 0, expected: 'learning' },
    { gamesPlayed: 1, expected: 'learning' },
    { gamesPlayed: 4, expected: 'learning' },
    { gamesPlayed: 5, expected: 'adapting' },
    { gamesPlayed: 10, expected: 'adapting' },
    { gamesPlayed: 14, expected: 'adapting' },
    { gamesPlayed: 15, expected: 'adapted' },
    { gamesPlayed: 100, expected: 'adapted' },
  ];

  it.each(cases)('returns "$expected" when gamesPlayed=$gamesPlayed', ({ gamesPlayed, expected }) => {
    expect(deriveAdaptationLevel(gamesPlayed)).toBe(expected);
  });
});

describe('deriveStrategyStyle', () => {
  describe('OldMaid (edgePickRate only)', () => {
    const cases: { edgePickRate: number; expected: StrategyStyle }[] = [
      { edgePickRate: 0.6, expected: 'cautious' },
      { edgePickRate: 0.51, expected: 'cautious' },
      { edgePickRate: 0.5, expected: 'balanced' },
      { edgePickRate: 0.3, expected: 'balanced' },
      { edgePickRate: 0, expected: 'balanced' },
    ];

    it.each(cases)('returns "$expected" when edgePickRate=$edgePickRate', ({ edgePickRate, expected }) => {
      expect(deriveStrategyStyle({ edgePickRate })).toBe(expected);
    });
  });

  describe('Betting games (bluffRate + foldRate)', () => {
    const cases: { bluffRate: number; foldRate: number; expected: StrategyStyle }[] = [
      { bluffRate: 0.4, foldRate: 0.2, expected: 'aggressive' },
      { bluffRate: 0.31, foldRate: 0.9, expected: 'aggressive' },
      { bluffRate: 0.1, foldRate: 0.7, expected: 'defensive' },
      { bluffRate: 0.2, foldRate: 0.61, expected: 'defensive' },
      { bluffRate: 0.05, foldRate: 0.2, expected: 'observing' },
      { bluffRate: 0.0, foldRate: 0.0, expected: 'observing' },
      { bluffRate: 0.2, foldRate: 0.4, expected: 'balanced' },
      { bluffRate: 0.15, foldRate: 0.5, expected: 'balanced' },
    ];

    it.each(cases)(
      'returns "$expected" when bluffRate=$bluffRate, foldRate=$foldRate',
      ({ bluffRate, foldRate, expected }) => {
        expect(deriveStrategyStyle({ bluffRate, foldRate })).toBe(expected);
      },
    );
  });

  describe('edge cases', () => {
    it('returns "observing" when no rates are provided', () => {
      expect(deriveStrategyStyle({})).toBe('observing');
    });

    it('ignores edgePickRate when bluffRate is also present', () => {
      // When both bluffRate and edgePickRate are set, it should follow the betting path
      expect(deriveStrategyStyle({ bluffRate: 0.4, edgePickRate: 0.6 })).toBe('aggressive');
    });
  });
});
