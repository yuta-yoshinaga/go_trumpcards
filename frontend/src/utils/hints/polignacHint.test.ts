import { describe, expect, it } from 'vitest';
import type { PolignacResponse } from '../../types/card';
import { getPolignacHint } from './polignacHint';

const base = (hint?: PolignacResponse['hint']): PolignacResponse => ({ hint }) as unknown as PolignacResponse;

describe('getPolignacHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getPolignacHint(base())).toBeNull();
  });

  it('returns null when the hint carries no card index', () => {
    expect(getPolignacHint(base({ reason: 'polignacDumpJack' }))).toBeNull();
  });

  it('accepts card index 0', () => {
    expect(getPolignacHint(base({ cardIndex: 0, reason: 'polignacLeadSafe' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.polignacLeadSafe',
      confidence: 'moderate',
    });
  });

  // ほぼ一択の場面と、判断が割れる場面を両方踏む。
  it.each([
    ['polignacAvoidJack', 'strong'],
    ['polignacBlockCapot', 'strong'],
    // 自分の capot 中は全トリック取るしかなく、選択の余地がほぼ無い。
    ['polignacWinCapot', 'strong'],
    ['polignacDumpJack', 'moderate'],
    ['polignacLeadSafe', 'moderate'],
  ])('reports %s as %s confidence', (reason, confidence) => {
    const r = getPolignacHint(base({ cardIndex: 2, reason }));
    expect(r?.confidence).toBe(confidence);
    expect(r?.reason).toBe(`hint.${reason}`);
  });
});
