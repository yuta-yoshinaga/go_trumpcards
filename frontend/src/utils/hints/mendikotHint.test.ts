import { describe, expect, it } from 'vitest';
import type { MendikotResponse } from '../../types/card';
import { getMendikotHint } from './mendikotHint';

const base = (hint?: MendikotResponse['hint']): MendikotResponse => ({ hint }) as unknown as MendikotResponse;

describe('getMendikotHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getMendikotHint(base())).toBeNull();
  });

  // **切り札を宣言する場面が無いので、札を指さないヒントは存在しない。**
  it('returns null when the hint names no card', () => {
    expect(getMendikotHint(base({ reason: 'mendikotDuck' }))).toBeNull();
  });

  it('accepts card index 0', () => {
    expect(getMendikotHint(base({ cardIndex: 0, reason: 'mendikotDuck' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.mendikotDuck',
      confidence: 'moderate',
    });
  });

  // **場に 10 が出ているときだけ強く勧める。** ほかは判断の余地がある。
  it('is strong only when chasing a ten', () => {
    expect(getMendikotHint(base({ cardIndex: 4, reason: 'mendikotChaseTen' }))?.confidence).toBe('strong');
    expect(getMendikotHint(base({ cardIndex: 4, reason: 'mendikotFeedPartner' }))?.confidence).toBe('moderate');
    expect(getMendikotHint(base({ cardIndex: 4, reason: 'mendikotDuck' }))?.confidence).toBe('moderate');
  });
});
