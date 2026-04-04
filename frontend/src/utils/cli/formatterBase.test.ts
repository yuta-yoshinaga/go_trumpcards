import { describe, expect, it } from 'vitest';
import type { Card } from '../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from './formatterBase';

const S = '\u2660'; // spade
const H = '\u2665'; // heart
const D = '\u2666'; // diamond
const C = '\u2663'; // clover

describe('formatCard', () => {
  it('formats spade card', () => {
    expect(formatCard({ design: 'SPADE', value: 5 })).toBe(`${S}5`);
  });

  it('formats heart card', () => {
    expect(formatCard({ design: 'HEART', value: 10 })).toBe(`${H}10`);
  });

  it('formats diamond ace', () => {
    expect(formatCard({ design: 'DIAMOND', value: 1 })).toBe(`${D}A`);
  });

  it('formats clover king', () => {
    expect(formatCard({ design: 'CLOVER', value: 13 })).toBe(`${C}K`);
  });

  it('formats joker', () => {
    expect(formatCard({ design: 'JOKER', value: 0 })).toBe('\uD83C\uDCCF');
  });

  it('formats jack', () => {
    expect(formatCard({ design: 'SPADE', value: 11 })).toBe(`${S}J`);
  });

  it('formats queen', () => {
    expect(formatCard({ design: 'HEART', value: 12 })).toBe(`${H}Q`);
  });
});

describe('formatCardList', () => {
  it('formats multiple cards comma-separated', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 10 },
    ];
    expect(formatCardList(cards)).toBe(`${S}5, ${H}10`);
  });

  it('returns empty string for empty array', () => {
    expect(formatCardList([])).toBe('');
  });
});

describe('formatIndexedCards', () => {
  it('formats cards with indices', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 13 },
    ];
    expect(formatIndexedCards(cards)).toBe(`[0]${S}5  [1]${H}K`);
  });

  it('returns empty string for empty array', () => {
    expect(formatIndexedCards([])).toBe('');
  });
});

describe('formatSeparator', () => {
  it('returns separator line', () => {
    expect(formatSeparator()).toBe('==========');
  });
});

describe('formatHeader', () => {
  it('wraps title in separators', () => {
    expect(formatHeader('BlackJack')).toBe('==========\nBlackJack\n==========');
  });
});

describe('formatPlayerName', () => {
  it('returns you label for human', () => {
    expect(formatPlayerName(0, true)).toBe('\u3042\u306a\u305f');
  });

  it('returns CPU label for non-human', () => {
    expect(formatPlayerName(1, false)).toBe('CPU 1');
  });

  it('returns CPU label with given index', () => {
    expect(formatPlayerName(3, false)).toBe('CPU 3');
  });
});
