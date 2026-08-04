import { describe, expect, it } from 'vitest';
import { hachiHachiAction, hachiHachiPendingCandidates } from './hachihachiKeyboard';

describe('hachiHachiPendingCandidates', () => {
  it('is empty when nothing is selected', () => {
    expect(hachiHachiPendingCandidates({ 0: [1, 2] }, null)).toEqual([]);
  });

  it('is empty when the selected card has no recorded options', () => {
    expect(hachiHachiPendingCandidates({ 0: [1, 2] }, 3)).toEqual([]);
  });

  it('returns the selected card options', () => {
    expect(hachiHachiPendingCandidates({ 2: [4, 5] }, 2)).toEqual([4, 5]);
  });
});

describe('hachiHachiAction', () => {
  it('plays a hand card with no match', () => {
    expect(hachiHachiAction({}, null, 1)).toEqual({ kind: 'play', handIndex: 1 });
  });

  it('plays a hand card with exactly one match, letting the backend capture', () => {
    expect(hachiHachiAction({ 1: [3] }, null, 1)).toEqual({ kind: 'play', handIndex: 1 });
  });

  it('only selects a hand card that matches two field cards', () => {
    expect(hachiHachiAction({ 1: [3, 4] }, null, 1)).toEqual({ kind: 'select', handIndex: 1 });
  });

  it('addresses the field once a choice is owed', () => {
    expect(hachiHachiAction({ 1: [3, 4] }, 1, 4)).toEqual({ kind: 'play', handIndex: 1, fieldIndex: 4 });
  });

  it('ignores a field index the selection cannot capture', () => {
    expect(hachiHachiAction({ 1: [3, 4] }, 1, 5)).toEqual({ kind: 'none' });
  });

  it('goes back to addressing the hand once the selection has one match left', () => {
    // Selected card 1 owes nothing (single match), so 2 names a hand card.
    expect(hachiHachiAction({ 1: [3], 2: [6, 7] }, 1, 2)).toEqual({ kind: 'select', handIndex: 2 });
  });
});
