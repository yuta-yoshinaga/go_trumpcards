import { describe, expect, it } from 'vitest';
import type { Card, TrashResponse } from '../../types/card';
import { TrashPhase } from '../../types/phases';
import { getTrashHint } from './trashHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

/** `faceUp` が true の位置は既に埋まっている。 */
const slots = (filled: number[]) =>
  Array.from({ length: 10 }, (_, i) => ({ card: card('SPADE', i + 1), faceUp: filled.includes(i) }));

function base(overrides: Partial<TrashResponse> = {}) {
  return {
    phase: TrashPhase.PLAYER_TURN,
    current: 0,
    players: [
      { slots: slots([]), isCpu: false },
      { slots: slots([]), isCpu: true },
    ],
    stockSize: 20,
    discardSize: 3,
    suggestedWildSlot: -1,
    moveCount: 0,
    winner: -1,
    ...overrides,
  } as TrashResponse;
}

describe('getTrashHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getTrashHint(base({ phase: TrashPhase.GAME_OVER }))).toBeNull();
  });

  it("stays quiet on the opponent's turn", () => {
    expect(getTrashHint(base({ current: 1 }))).toBeNull();
  });

  it('suggests drawing on an ordinary turn', () => {
    expect(getTrashHint(base())?.targetAction).toBe('draw');
  });

  it('places a wild in the domain-suggested slot', () => {
    const s = base({
      phase: TrashPhase.AWAIT_WILD,
      players: [
        { slots: slots([0, 1, 2]), isCpu: false },
        { slots: slots([]), isCpu: true },
      ],
    });
    const suggested = { ...s, suggestedWildSlot: 7 };
    expect(getTrashHint(suggested)?.targetAction).toBe('slot-7');
  });

  it('keeps slot 0 as a valid answer', () => {
    // **位置 0 も正当。**真偽値で見ると先頭だけ落ちる。
    const s = base({
      phase: TrashPhase.AWAIT_WILD,
      suggestedWildSlot: 0,
      players: [
        { slots: slots([1, 2, 3]), isCpu: false },
        { slots: slots([]), isCpu: true },
      ],
    });
    expect(getTrashHint(s)?.targetAction).toBe('slot-0');
  });

  it('uses the domain suggestion when lower slots are filled', () => {
    const s = base({
      phase: TrashPhase.AWAIT_WILD,
      suggestedWildSlot: 8,
      players: [
        { slots: slots([0, 1, 2, 3, 4, 6]), isCpu: false },
        { slots: slots([]), isCpu: true },
      ],
    });
    expect(getTrashHint(s)?.targetAction).toBe('slot-8');
  });

  it('says nothing when every slot is already filled', () => {
    const s = base({
      phase: TrashPhase.AWAIT_WILD,
      suggestedWildSlot: -1,
      players: [
        { slots: slots([0, 1, 2, 3, 4, 5, 6, 7, 8, 9]), isCpu: false },
        { slots: slots([]), isCpu: true },
      ],
    });
    expect(getTrashHint(s)).toBeNull();
  });
});
