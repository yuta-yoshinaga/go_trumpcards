import { describe, expect, it } from 'vitest';
import type { SpiteAndMaliceResponse } from '../../types/card';
import { getSpiteAndMaliceHint } from './spiteAndMaliceHint';

const baseState: SpiteAndMaliceResponse = {
  phase: 0,
  current: 0,
  players: [
    { hand: [], goalSize: 0, sides: [[], [], [], []], isCpu: false },
    { hand: [], goalSize: 0, sides: [[], [], [], []], isCpu: true },
  ],
  foundations: [[], [], [], []],
  foundationTops: [0, 0, 0, 0],
  stockSize: 0,
  completedSize: 0,
  moveCount: 0,
  winner: -1,
  goalSize: 20,
  cpuDifficulty: 1,
  canAutoComplete: false,
  message: '',
};

describe('getSpiteAndMaliceHint', () => {
  it('returns null when game is over', () => {
    expect(getSpiteAndMaliceHint({ ...baseState, phase: 1 })).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(
      getSpiteAndMaliceHint({
        ...baseState,
        current: 1,
        hint: { source: 'goal', index: -1, foundationIdx: 0, discard: false },
      }),
    ).toBeNull();
  });

  it('returns null when no hint is provided', () => {
    expect(getSpiteAndMaliceHint(baseState)).toBeNull();
  });

  it('returns goal hint', () => {
    const r = getSpiteAndMaliceHint({
      ...baseState,
      hint: { source: 'goal', index: -1, foundationIdx: 2, discard: false },
    });
    expect(r).toEqual({ targetAction: 'goal-to-f2', reason: 'frontendHint.goalToFoundation', confidence: 'strong' });
  });

  it('returns hand hint', () => {
    const r = getSpiteAndMaliceHint({
      ...baseState,
      hint: { source: 'hand', index: 3, foundationIdx: 0, discard: false },
    });
    expect(r).toEqual({ targetAction: 'hand3-to-f0', reason: 'frontendHint.handToFoundation', confidence: 'strong' });
  });

  it('returns side hint', () => {
    const r = getSpiteAndMaliceHint({
      ...baseState,
      hint: { source: 'side', index: 1, foundationIdx: 3, discard: false },
    });
    expect(r).toEqual({ targetAction: 'side1-to-f3', reason: 'frontendHint.sideToFoundation', confidence: 'strong' });
  });

  it('returns discard hint', () => {
    const r = getSpiteAndMaliceHint({
      ...baseState,
      hint: { source: 'hand', index: 2, foundationIdx: 0, discard: true },
    });
    expect(r).toEqual({ targetAction: 'discard-2-0', reason: 'frontendHint.discard', confidence: 'moderate' });
  });

  it('returns null for unknown source', () => {
    const r = getSpiteAndMaliceHint({
      ...baseState,
      // intentionally invalid source to exercise the default branch
      hint: { source: 'bogus' as 'goal', index: 0, foundationIdx: 0, discard: false },
    });
    expect(r).toBeNull();
  });
});
