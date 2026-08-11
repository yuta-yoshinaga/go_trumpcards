import { describe, expect, it } from 'vitest';
import type { RamsResponse } from '../../types/card';
import { getRamsHint } from './ramsHint';

const base = (hint?: RamsResponse['hint']): RamsResponse => ({ hint }) as unknown as RamsResponse;

describe('getRamsHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getRamsHint(base())).toBeNull();
  });

  // **選択フェーズのヒントは札を指さない。** cardIndex が無くても null にしない。
  it.each([
    ['ramsPlayIn', 'play-in'],
    ['ramsPassOut', 'pass-out'],
  ])('turns %s into the %s decision', (reason, targetAction) => {
    expect(getRamsHint(base({ reason }))).toEqual({
      targetAction,
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getRamsHint(base({ cardIndex: 0, reason: 'ramsAlreadySafe' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.ramsAlreadySafe',
      confidence: 'moderate',
    });
  });

  // 1トリック目の確保は追加支払いを避ける唯一の手なので、確信度が高い。
  it('is more confident about banking the first trick', () => {
    expect(getRamsHint(base({ cardIndex: 2, reason: 'ramsTakeTrick' }))?.confidence).toBe('strong');
    expect(getRamsHint(base({ cardIndex: 2, reason: 'ramsAlreadySafe' }))?.confidence).toBe('moderate');
  });
});
