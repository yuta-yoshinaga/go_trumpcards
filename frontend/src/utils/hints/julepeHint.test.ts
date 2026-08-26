import { describe, expect, it } from 'vitest';
import type { JulepeResponse } from '../../types/card';
import { getJulepeHint } from './julepeHint';

const base = (hint?: JulepeResponse['hint']): JulepeResponse => ({ hint }) as unknown as JulepeResponse;

describe('getJulepeHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getJulepeHint(base())).toBeNull();
  });

  // **選択フェーズのヒントは札を指さない。** cardIndex が無くても null にしない。
  it.each([
    ['julepePlayIn', 'play-in'],
    ['julepePassOut', 'pass-out'],
  ])('turns %s into the %s decision', (reason, targetAction) => {
    expect(getJulepeHint(base({ reason }))).toEqual({
      targetAction,
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getJulepeHint(base({ cardIndex: 0, reason: 'julepeAlreadySafe' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.julepeAlreadySafe',
      confidence: 'moderate',
    });
  });

  // 1トリック目の確保は追加支払いを避ける唯一の手なので、確信度が高い。
  it('is more confident about banking the first trick', () => {
    expect(getJulepeHint(base({ cardIndex: 2, reason: 'julepeTakeTrick' }))?.confidence).toBe('strong');
    expect(getJulepeHint(base({ cardIndex: 2, reason: 'julepeAlreadySafe' }))?.confidence).toBe('moderate');
  });
});
