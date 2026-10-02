import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { rummy500MeldScore } from './rummy500MeldScore';

const meldScoreCases: { name: string; cards: Card[]; expected: number }[] = [
  {
    name: 'set of 7s = 21',
    cards: [
      { design: 'SPADE', value: 7 },
      { design: 'HEART', value: 7 },
      { design: 'CLOVER', value: 7 },
    ],
    expected: 21,
  },
  {
    name: 'low run A-2-3 = 6',
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'SPADE', value: 2 },
      { design: 'SPADE', value: 3 },
    ],
    expected: 6,
  },
  {
    name: 'high run Q-K-A = 10+10+15 = 35',
    cards: [
      { design: 'SPADE', value: 12 },
      { design: 'SPADE', value: 13 },
      { design: 'SPADE', value: 1 },
    ],
    expected: 35,
  },
  {
    name: 'face cards K-K-K = 30',
    cards: [
      { design: 'SPADE', value: 13 },
      { design: 'HEART', value: 13 },
      { design: 'CLOVER', value: 13 },
    ],
    expected: 30,
  },
];

describe('rummy500MeldScore', () => {
  it.each(meldScoreCases)('$name', ({ cards, expected }) => {
    expect(rummy500MeldScore(cards)).toBe(expected);
  });
});
