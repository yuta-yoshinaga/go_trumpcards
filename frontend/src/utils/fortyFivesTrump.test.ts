import { describe, expect, it } from 'vitest';
import { formatFortyFivesTrumpOrder } from './fortyFivesTrump';

describe('formatFortyFivesTrumpOrder', () => {
  it('formats the order for a non-heart trump suit', () => {
    expect(formatFortyFivesTrumpOrder('♠')).toEqual([
      '♠5',
      '♠J',
      '♥A',
      '♠A',
      '♠K',
      '♠Q',
      '♠10',
      '♠9',
      '♠8',
      '♠7',
      '♠6',
      '♠4',
      '♠3',
      '♠2',
    ]);
  });

  it('removes the duplicate ace when hearts are trump', () => {
    expect(formatFortyFivesTrumpOrder('♥')).toEqual([
      '♥5',
      '♥J',
      '♥A',
      '♥K',
      '♥Q',
      '♥10',
      '♥9',
      '♥8',
      '♥7',
      '♥6',
      '♥4',
      '♥3',
      '♥2',
    ]);
  });
});
