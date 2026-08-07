import { describe, expect, it } from 'vitest';
import type { CasinoHoldemResponse } from '../../types/card';
import { CasinoHoldemPhase } from '../../types/phases';
import { getCasinoHoldemHint } from './casinoholdemHint';

/** Build a CasinoHoldemResponse with sensible defaults so each test overrides
 * only the fields it cares about. */
function makeState(overrides: Partial<CasinoHoldemResponse> = {}): CasinoHoldemResponse {
  return {
    playerHand: [
      { design: 'SPADE', value: 7 },
      { design: 'HEART', value: 7 },
    ],
    dealerHand: [],
    community: [
      { design: 'DIAMOND', value: 3 },
      { design: 'CLOVER', value: 9 },
      { design: 'SPADE', value: 11 },
    ],
    phase: CasinoHoldemPhase.FLOP,
    chips: 1000,
    anteBet: 100,
    bonusBet: 0,
    callBet: 0,
    result: 0,
    dealerQualify: false,
    antePayout: 0,
    callPayout: 0,
    bonusPayout: 0,
    totalPayout: 0,
    playerHandRank: 1, // OnePair by default
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getCasinoHoldemHint', () => {
  it('returns call hint for pair-or-better at FLOP', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 1 }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns fold hint for High Card at FLOP', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 0 }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null for BET phase', () => {
    const hint = getCasinoHoldemHint(makeState({ phase: CasinoHoldemPhase.BET }));
    expect(hint).toBeNull();
  });

  it('returns null for END phase', () => {
    const hint = getCasinoHoldemHint(makeState({ phase: CasinoHoldemPhase.END }));
    expect(hint).toBeNull();
  });

  it('returns null when no hole cards have been dealt', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHand: [] }));
    expect(hint).toBeNull();
  });

  it('returns strong call for a Flush', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 5 }));
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('strong');
  });
});

// **CUI と Web が逆の助言を出していた (#4712)。**ドメインの RecommendCall は
// 「ワンペア以上、または5枚のどこかに A か K があればコール」。CUI はこれを
// そのまま使うのに、こちらはランクしか見ていなかった。
describe('getCasinoHoldemHint — ace/king rule (sync: CasinoHoldem.RecommendCall)', () => {
  it('calls on an ace in the hole with no made hand', () => {
    const hint = getCasinoHoldemHint(
      makeState({
        playerHand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 6 },
        ],
        community: [
          { design: 'DIAMOND', value: 3 },
          { design: 'CLOVER', value: 9 },
          { design: 'SPADE', value: 11 },
        ],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('call');
  });

  it('calls on a king in the hole with no made hand', () => {
    const hint = getCasinoHoldemHint(
      makeState({
        playerHand: [
          { design: 'SPADE', value: 13 },
          { design: 'HEART', value: 6 },
        ],
        community: [
          { design: 'DIAMOND', value: 3 },
          { design: 'CLOVER', value: 9 },
          { design: 'SPADE', value: 11 },
        ],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('call');
  });

  // **ボードの A / K も数える。**ドメインは hole と community を区別せず
  // 5枚すべてを走査する。ここだけ手札に限ると、また CUI とずれる。
  it('calls when the ace is on the board rather than in the hole', () => {
    const hint = getCasinoHoldemHint(
      makeState({
        playerHand: [
          { design: 'SPADE', value: 6 },
          { design: 'HEART', value: 4 },
        ],
        community: [
          { design: 'DIAMOND', value: 1 },
          { design: 'CLOVER', value: 9 },
          { design: 'SPADE', value: 11 },
        ],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('call');
  });

  // **A と K の両方は要らない。**oasispoker の hasAceKing とは規則が違う。
  it('does not require both an ace and a king', () => {
    const hint = getCasinoHoldemHint(
      makeState({
        playerHand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 4 },
        ],
        community: [
          { design: 'DIAMOND', value: 6 },
          { design: 'CLOVER', value: 9 },
          { design: 'SPADE', value: 11 },
        ],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('call');
  });

  it('still folds when no card is an ace or a king', () => {
    const hint = getCasinoHoldemHint(
      makeState({
        playerHand: [
          { design: 'SPADE', value: 6 },
          { design: 'HEART', value: 4 },
        ],
        community: [
          { design: 'DIAMOND', value: 3 },
          { design: 'CLOVER', value: 9 },
          { design: 'SPADE', value: 11 },
        ],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('fold');
  });
});
