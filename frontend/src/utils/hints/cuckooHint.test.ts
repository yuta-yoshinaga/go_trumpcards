import { describe, expect, it } from 'vitest';
import type { Card, CuckooResponse } from '../../types/card';
import { CuckooPhase } from '../../types/phases';
import { getCuckooHint } from './cuckooHint';

const card = (value: number): Card => ({ design: 'HEART', value });

function base(overrides: Partial<CuckooResponse> = {}, humanCard: Card | null = card(4)): CuckooResponse {
  return {
    players: [
      { id: 0, isHuman: true, card: humanCard, lives: 3, isEliminated: false, kingRevealed: false, isTurn: true },
      { id: 1, isHuman: false, card: null, lives: 3, isEliminated: false, kingRevealed: false, isTurn: false },
    ],
    phase: CuckooPhase.TURN,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 1,
    stockCount: 40,
    gameEndFlag: false,
    winnerIdx: -1,
    pendingSwapFrom: -1,
    pendingSwapTo: -1,
    roundLowest: -1,
    roundLosers: [],
    message: '',
    config: { cpuDifficulty: 1, startLives: 3 },
    ...overrides,
  } as CuckooResponse;
}

describe('getCuckooHint', () => {
  it('stays quiet when it is not the human turn', () => {
    expect(getCuckooHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet once the game is over', () => {
    expect(getCuckooHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet while the human card is hidden', () => {
    expect(getCuckooHint(base({}, null))).toBeNull();
  });

  it('recommends swapping a low card', () => {
    expect(getCuckooHint(base({}, card(3)))).toEqual({
      targetAction: 'swap',
      reason: 'frontendHint.cuckooSwapLow',
      confidence: 'strong',
    });
  });

  it('recommends keeping a high card', () => {
    expect(getCuckooHint(base({}, card(11)))).toEqual({
      targetAction: 'keep',
      reason: 'frontendHint.cuckooKeepHigh',
      confidence: 'strong',
    });
  });

  // 閾値ちょうどは「交換したい側」。CPU の判断 (cuckooCpuSwapThreshold = 7) に合わせる。
  it('treats the threshold itself as low', () => {
    expect(getCuckooHint(base({}, card(7)))?.targetAction).toBe('swap');
    expect(getCuckooHint(base({}, card(8)))?.targetAction).toBe('keep');
  });

  // **King は交換されない側。**取られる心配が無いので必ず残す。
  it('always keeps a King', () => {
    expect(getCuckooHint(base({}, card(13)))).toEqual({
      targetAction: 'keep',
      reason: 'frontendHint.cuckooKeepKing',
      confidence: 'strong',
    });
  });

  // 親は隣ではなく山と交換する。行き先が違うので理由も分ける。
  it('names the stock when the dealer is the one swapping', () => {
    expect(getCuckooHint(base({ dealerIdx: 0 }, card(2)))?.reason).toBe('frontendHint.cuckooSwapStock');
  });

  it('recommends refusing with a King', () => {
    const state = base({ phase: CuckooPhase.REFUSE, pendingSwapTo: 0, pendingSwapFrom: 1 }, card(13));
    expect(getCuckooHint(state)).toEqual({
      targetAction: 'refuse',
      reason: 'frontendHint.cuckooRefuseKing',
      confidence: 'strong',
    });
  });

  // **King が無ければ拒否ボタンは押せない。**押せない手を勧めない。
  it('does not recommend refusing without a King', () => {
    const state = base({ phase: CuckooPhase.REFUSE, pendingSwapTo: 0, pendingSwapFrom: 1 }, card(9));
    expect(getCuckooHint(state)?.targetAction).toBe('accept');
  });

  it('stays quiet when someone else is the swap target', () => {
    const state = base({ phase: CuckooPhase.REFUSE, pendingSwapTo: 1, pendingSwapFrom: 0 }, card(13));
    expect(getCuckooHint(state)).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getCuckooHint(base({ phase: CuckooPhase.ROUND_END }))).toBeNull();
  });
});
