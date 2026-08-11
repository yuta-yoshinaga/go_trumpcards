import { describe, expect, it } from 'vitest';
import type { ReversisResponse } from '../../types/card';
import { getReversisHint } from './reversisHint';

const base = (hint?: ReversisResponse['hint']): ReversisResponse => ({ hint }) as unknown as ReversisResponse;

describe('getReversisHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getReversisHint(base())).toBeNull();
  });

  it('returns null when the hint carries no card index', () => {
    expect(getReversisHint(base({ reason: 'reversisDumpHigh' }))).toBeNull();
  });

  it('accepts card index 0', () => {
    expect(getReversisHint(base({ cardIndex: 0, reason: 'reversisLeadSafe' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.reversisLeadSafe',
      confidence: 'moderate',
    });
  });

  // 印付きの札はチップまで取られるので、避ける手はほぼ一択。両側を踏む。
  it.each([
    ['reversisAvoidMarked', 'strong'],
    ['reversisAvoidPoints', 'moderate'],
    ['reversisLeadSafe', 'moderate'],
    ['reversisDumpHigh', 'moderate'],
  ])('reports %s as %s confidence', (reason, confidence) => {
    const r = getReversisHint(base({ cardIndex: 2, reason }));
    expect(r?.confidence).toBe(confidence);
    expect(r?.reason).toBe(`hint.${reason}`);
  });
});
