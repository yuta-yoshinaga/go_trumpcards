import { describe, expect, it } from 'vitest';
import type { TrashResponse } from '../../types/card';
import { getTrashHint } from './trashHint';

function fixture(phase: number): TrashResponse {
  return {
    phase,
    current: 0,
    players: [
      { slots: Array.from({ length: 10 }, () => ({ faceUp: false })), isCpu: false },
      { slots: Array.from({ length: 10 }, () => ({ faceUp: false })), isCpu: true },
    ],
    stockSize: 34,
    discardSize: 0,
    moveCount: 0,
    winner: -1,
    message: '',
  };
}

describe('getTrashHint', () => {
  it('returns null during player turn', () => {
    expect(getTrashHint(fixture(0))).toBeNull();
  });

  it('returns null while awaiting wild placement', () => {
    expect(getTrashHint(fixture(1))).toBeNull();
  });

  it('returns null after game over', () => {
    expect(getTrashHint(fixture(2))).toBeNull();
  });
});
