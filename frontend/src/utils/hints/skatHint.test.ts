import { describe, expect, it } from 'vitest';
import type { SkatResponse } from '../../types/card';
import { getSkatHint } from './skatHint';

function makeState(overrides?: Partial<SkatResponse>): SkatResponse {
  return {
    players: [],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    currentTrick: [],
    forehandIdx: 0,
    middlehandIdx: 1,
    rearhandIdx: 2,
    dealerIdx: 0,
    declarerIdx: -1,
    currentBid: 0,
    activeBidActorIdx: 1,
    gameType: 0,
    trumpSuit: 0,
    pickedSkat: false,
    declarerCardPoints: 0,
    defendersCardPoints: 0,
    winnerSide: 0,
    gameValue: 0,
    gameEndFlag: false,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, targetScore: 1000 },
    ...overrides,
  };
}

describe('getSkatHint', () => {
  it('returns null when game ended', () => {
    const r = getSkatHint(makeState({ gameEndFlag: true, hint: { reason: 'x' } }));
    expect(r).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getSkatHint(makeState())).toBeNull();
  });

  it('maps bid phase to bid action', () => {
    const r = getSkatHint(makeState({ phase: 0, hint: { reason: 'r', bid: 18 } }));
    expect(r?.targetAction).toBe('bid');
    expect(r?.reason).toBe('r');
  });

  it('maps bid pass when no bid value', () => {
    const r = getSkatHint(makeState({ phase: 0, hint: { reason: 'r', bid: 0 } }));
    expect(r?.targetAction).toBe('pass');
  });

  it('maps skat pickup', () => {
    const r = getSkatHint(makeState({ phase: 1, hint: { reason: 'r', pickSkat: true } }));
    expect(r?.targetAction).toBe('pickSkat');
  });

  it('maps hand game when no pickup', () => {
    const r = getSkatHint(makeState({ phase: 1, hint: { reason: 'r', pickSkat: false } }));
    expect(r?.targetAction).toBe('handGame');
  });

  it('maps discard phase', () => {
    const r = getSkatHint(makeState({ phase: 2, hint: { reason: 'r', discardIndex: 3 } }));
    expect(r?.targetAction).toBe('discard');
  });

  it('maps declare phase', () => {
    const r = getSkatHint(makeState({ phase: 3, hint: { reason: 'r', gameType: 1 } }));
    expect(r?.targetAction).toBe('declare');
  });

  it('maps play phase', () => {
    const r = getSkatHint(makeState({ phase: 4, hint: { reason: 'r', cardIndex: 2 } }));
    expect(r?.targetAction).toBe('play');
  });

  it('returns null for non-interactive phases', () => {
    expect(getSkatHint(makeState({ phase: 5, hint: { reason: 'r' } }))).toBeNull();
  });
});
