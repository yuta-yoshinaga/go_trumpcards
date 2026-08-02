import { describe, expect, it } from 'vitest';
import type { RussianBankResponse } from '../../types/card';
import { getRussianBankHint } from './russianbankHint';

function makeState(overrides?: Partial<RussianBankResponse>): RussianBankResponse {
  return { phase: 1, message: '', ...overrides } as RussianBankResponse;
}

describe('getRussianBankHint', () => {
  it('rates a foundation move as strong', () => {
    const hint = getRussianBankHint(
      makeState({ hint: { zone: 2, fromOpponent: false, col: 3, toFoundation: true, toCol: -1 } }),
    );
    expect(hint?.targetAction).toBe('tableau-3');
    expect(hint?.confidence).toBe('strong');
  });

  it('rates a tableau move as moderate', () => {
    const hint = getRussianBankHint(
      makeState({ hint: { zone: 2, fromOpponent: false, col: 0, toFoundation: false, toCol: 5 } }),
    );
    expect(hint?.reason).toBe('frontendHint.russianbankToTableau');
    expect(hint?.confidence).toBe('moderate');
  });

  // **リザーブと廃札は 1 山なので列が無い。**col を付けると reserve--1 になる。
  it('names a single pile without a column', () => {
    const reserve = getRussianBankHint(
      makeState({ hint: { zone: 0, fromOpponent: false, col: -1, toFoundation: true, toCol: -1 } }),
    );
    expect(reserve?.targetAction).toBe('reserve');

    const waste = getRussianBankHint(
      makeState({ hint: { zone: 1, fromOpponent: false, col: -1, toFoundation: false, toCol: 2 } }),
    );
    expect(waste?.targetAction).toBe('waste');
  });

  it('returns null without a hint', () => {
    expect(getRussianBankHint(makeState())).toBeNull();
  });

  it('returns null for a zone it does not know', () => {
    expect(
      getRussianBankHint(makeState({ hint: { zone: 9, fromOpponent: false, col: 0, toFoundation: true, toCol: -1 } })),
    ).toBeNull();
  });
});
