import { describe, expect, it } from 'vitest';
import type { TarabishResponse } from '../../types/card';
import { getTarabishHint } from './tarabishHint';

const base = (hint?: TarabishResponse['hint']): TarabishResponse => ({ hint }) as unknown as TarabishResponse;

describe('getTarabishHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getTarabishHint(base())).toBeNull();
  });

  // **入札のヒントは札を指さない。** cardIndex が無くても null にしない。
  it.each([
    ['tarabishTakeTrump', 'take-trump'],
    ['tarabishPassTrump', 'pass-trump'],
  ])('turns %s into the %s decision', (reason, targetAction) => {
    expect(getTarabishHint(base({ reason }))).toEqual({
      targetAction,
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getTarabishHint(base({ cardIndex: 0, reason: 'tarabishWinTrick' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.tarabishWinTrick',
      confidence: 'moderate',
    });
  });

  // 味方に点を乗せる手はほぼ一択。両側を踏む。
  it('is more confident about feeding a winning partner', () => {
    expect(getTarabishHint(base({ cardIndex: 2, reason: 'tarabishFeedPartner' }))?.confidence).toBe('strong');
    expect(getTarabishHint(base({ cardIndex: 2, reason: 'tarabishWinTrick' }))?.confidence).toBe('moderate');
  });
});
