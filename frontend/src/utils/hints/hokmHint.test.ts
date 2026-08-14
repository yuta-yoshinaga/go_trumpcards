import { describe, expect, it } from 'vitest';
import type { HokmResponse } from '../../types/card';
import { getHokmHint } from './hokmHint';

const base = (hint?: HokmResponse['hint']): HokmResponse => ({ hint }) as unknown as HokmResponse;

describe('getHokmHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getHokmHint(base())).toBeNull();
  });

  // **切り札のヒントは札を指さない。** cardIndex が無くても null にしない。
  it.each([1, 2, 3, 4])('names suit %s when it recommends a trump', (suit) => {
    expect(getHokmHint(base({ reason: 'hokmDeclareTrump', suit }))).toEqual({
      targetAction: `trump-${suit}`,
      reason: 'hint.hokmDeclareTrump',
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getHokmHint(base({ cardIndex: 0, reason: 'hokmWinTrick', suit: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.hokmWinTrick',
      confidence: 'moderate',
    });
  });

  // 味方が勝っているなら温存はほぼ一択。両側を踏む。
  it('is more confident about holding cards back', () => {
    expect(getHokmHint(base({ cardIndex: 2, reason: 'hokmSaveCards', suit: 0 }))?.confidence).toBe('strong');
    expect(getHokmHint(base({ cardIndex: 2, reason: 'hokmWinTrick', suit: 0 }))?.confidence).toBe('moderate');
  });
});
