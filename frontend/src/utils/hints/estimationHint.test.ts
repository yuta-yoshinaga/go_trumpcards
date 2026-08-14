import { describe, expect, it } from 'vitest';
import type { EstimationResponse } from '../../types/card';
import { getEstimationHint } from './estimationHint';

const base = (hint?: EstimationResponse['hint']): EstimationResponse => ({ hint }) as unknown as EstimationResponse;

describe('getEstimationHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getEstimationHint(base())).toBeNull();
  });

  // **切り札のヒントはスートを指す。** cardIndex が無くても null にしない。
  it('names the suit when it recommends a trump', () => {
    expect(getEstimationHint(base({ reason: 'estimationSelectTrump', value: 3 }))).toEqual({
      targetAction: 'trump-3',
      reason: 'hint.estimationSelectTrump',
      confidence: 'moderate',
    });
  });

  // **宣言のヒントは数を指す。** 0（Dash Call）も含めて。
  it.each([
    ['estimationBid', 4, 'bid-4'],
    ['estimationDashCall', 0, 'bid-0'],
    ['estimationAvoidRestricted', 2, 'bid-2'],
  ])('turns %s into %s', (reason, value, targetAction) => {
    expect(getEstimationHint(base({ reason, value }))).toEqual({
      targetAction,
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getEstimationHint(base({ cardIndex: 0, reason: 'estimationWinTrick', value: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.estimationWinTrick',
      confidence: 'moderate',
    });
  });

  // 宣言ぶん取ったあとに逃げる手はほぼ一択。両側を踏む。
  it('is more confident about ducking once the call is made', () => {
    expect(getEstimationHint(base({ cardIndex: 2, reason: 'estimationDuck', value: 0 }))?.confidence).toBe('strong');
    expect(getEstimationHint(base({ cardIndex: 2, reason: 'estimationWinTrick', value: 0 }))?.confidence).toBe(
      'moderate',
    );
  });
});
