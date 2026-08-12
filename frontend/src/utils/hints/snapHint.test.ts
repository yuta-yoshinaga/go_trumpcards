import { describe, expect, it } from 'vitest';
import type { SnapResponse } from '../../types/card';
import { getSnapHint } from './snapHint';

const state = (hint?: SnapResponse['hint']): SnapResponse => ({ hint }) as SnapResponse;

describe('getSnapHint', () => {
  it('returns null without a hint', () => {
    expect(getSnapHint(state())).toBeNull();
  });

  // **宣言できる瞬間だけが「正解のある手」。**
  it('is confident when a call would be correct', () => {
    expect(getSnapHint(state({ snap: true, reason: 'snapDeclare' }))).toEqual({
      targetAction: 'snap',
      reason: 'hint.snapDeclare',
      confidence: 'strong',
    });
  });

  it.each([
    ['snapStep', 'あなたの番'],
    ['snapWait', '相手の番'],
  ])('advises waiting for %s', (reason) => {
    expect(getSnapHint(state({ snap: false, reason }))).toEqual({
      targetAction: 'wait',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });
});
