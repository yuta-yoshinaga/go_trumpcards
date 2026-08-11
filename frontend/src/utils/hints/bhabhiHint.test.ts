import { describe, expect, it } from 'vitest';
import type { BhabhiResponse } from '../../types/card';
import { getBhabhiHint } from './bhabhiHint';

const base = (hint?: BhabhiResponse['hint']): BhabhiResponse => ({ hint }) as unknown as BhabhiResponse;

describe('getBhabhiHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getBhabhiHint(base())).toBeNull();
  });

  // **札を指さないヒントは存在しない。** 出す以外の選択肢が無いゲーム。
  it('returns null when the hint names no card', () => {
    expect(getBhabhiHint(base({ reason: 'bhabhiLead' }))).toBeNull();
  });

  it('accepts card index 0', () => {
    expect(getBhabhiHint(base({ cardIndex: 0, reason: 'bhabhiLead' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.bhabhiLead',
      confidence: 'moderate',
    });
  });

  // **フォローできないときだけ強く勧める。** どのみち引き取るので迷いが無い。
  it('is strong only when dumping a high card', () => {
    expect(getBhabhiHint(base({ cardIndex: 4, reason: 'bhabhiDumpHigh' }))?.confidence).toBe('strong');
    expect(getBhabhiHint(base({ cardIndex: 4, reason: 'bhabhiLead' }))?.confidence).toBe('moderate');
    expect(getBhabhiHint(base({ cardIndex: 4, reason: 'bhabhiDuck' }))?.confidence).toBe('moderate');
  });
});
