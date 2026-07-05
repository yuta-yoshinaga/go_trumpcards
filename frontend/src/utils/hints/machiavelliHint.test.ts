import { describe, expect, it } from 'vitest';
import type { Card, MachiavelliMeld, MachiavelliPlayer, MachiavelliResponse } from '../../types/card';
import { MachiavelliPhase } from '../../types/phases';
import { canLayoff, findHandMeld, getMachiavelliHint } from './machiavelliHint';

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

function player(overrides: Partial<MachiavelliPlayer> = {}): MachiavelliPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    deadwood: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<MachiavelliResponse> = {}): MachiavelliResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    table: [],
    phase: MachiavelliPhase.TURN,
    roundNumber: 1,
    targetRounds: 3,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    drawPileCount: 40,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 3 },
    message: '',
    messageCode: '',
    messageParams: {},
    ...overrides,
  } as MachiavelliResponse;
}

describe('getMachiavelliHint', () => {
  it('returns null when no human player', () => {
    expect(getMachiavelliHint(makeState({ players: [player({ isHuman: false })] }))).toBeNull();
  });

  it('returns null when human has no cards', () => {
    expect(getMachiavelliHint(makeState())).toBeNull();
  });

  it('returns null when gameEndFlag is true', () => {
    const state = makeState({ gameEndFlag: true, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getMachiavelliHint(state)).toBeNull();
  });

  it('returns null when not the human turn', () => {
    const state = makeState({ currentPlayerIdx: 1, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getMachiavelliHint(state)).toBeNull();
  });

  it('returns null in round-end phase', () => {
    const state = makeState({ phase: MachiavelliPhase.ROUND_END, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getMachiavelliHint(state)).toBeNull();
  });

  it('suggests newMeld when the hand contains a run', () => {
    const state = makeState({
      players: [player({ cards: [card('SPADE', 3), card('SPADE', 4), card('SPADE', 5), card('HEART', 9)] })],
    });
    const hint = getMachiavelliHint(state);
    expect(hint?.targetAction).toBe('newMeld');
    expect(hint?.reason).toBe('hint.newMeld');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests newMeld when the hand contains a set', () => {
    const state = makeState({
      players: [player({ cards: [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 2)] })],
    });
    expect(getMachiavelliHint(state)?.targetAction).toBe('newMeld');
  });

  it('suggests layoff when a card extends a table run', () => {
    const table: MachiavelliMeld[] = [{ cards: [card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)], kind: 1 }];
    const state = makeState({
      table,
      players: [player({ cards: [card('SPADE', 6), card('HEART', 11)] })],
    });
    const hint = getMachiavelliHint(state);
    expect(hint?.targetAction).toBe('layoff');
    expect(hint?.reason).toBe('hint.layoff');
  });

  it('suggests draw when neither a meld nor a layoff is possible', () => {
    const state = makeState({
      players: [player({ cards: [card('SPADE', 2), card('HEART', 6), card('CLOVER', 11)] })],
    });
    const hint = getMachiavelliHint(state);
    expect(hint?.targetAction).toBe('draw');
    expect(hint?.reason).toBe('hint.draw');
  });
});

describe('findHandMeld', () => {
  it('returns indices of a set (distinct suits)', () => {
    const hand = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)];
    expect(findHandMeld(hand)).toEqual([0, 1, 2]);
  });

  it('does not treat duplicate suits as a set', () => {
    const hand = [card('SPADE', 7), card('SPADE', 7), card('SPADE', 7)];
    expect(findHandMeld(hand)).toBeNull();
  });

  it('returns indices of a run', () => {
    const hand = [card('HEART', 8), card('HEART', 9), card('HEART', 10)];
    expect(findHandMeld(hand)).toEqual([0, 1, 2]);
  });

  it('returns null when no meld exists', () => {
    const hand = [card('SPADE', 2), card('HEART', 6), card('CLOVER', 11)];
    expect(findHandMeld(hand)).toBeNull();
  });
});

describe('canLayoff', () => {
  it('extends a run at the low end', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 4), card('SPADE', 5), card('SPADE', 6)], kind: 1 };
    expect(canLayoff(card('SPADE', 3), meld)).toBe(true);
  });

  it('extends a run at the high end', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 4), card('SPADE', 5), card('SPADE', 6)], kind: 1 };
    expect(canLayoff(card('SPADE', 7), meld)).toBe(true);
  });

  it('rejects a wrong suit for a run', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 4), card('SPADE', 5), card('SPADE', 6)], kind: 1 };
    expect(canLayoff(card('HEART', 7), meld)).toBe(false);
  });

  it('adds a new suit to a set', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9)], kind: 0 };
    expect(canLayoff(card('DIAMOND', 9), meld)).toBe(true);
  });

  it('rejects a duplicate suit for a set', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9)], kind: 0 };
    expect(canLayoff(card('SPADE', 9), meld)).toBe(false);
  });

  it('rejects a wrong value for a set', () => {
    const meld: MachiavelliMeld = { cards: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9)], kind: 0 };
    expect(canLayoff(card('DIAMOND', 8), meld)).toBe(false);
  });

  it('returns false for an empty meld', () => {
    expect(canLayoff(card('SPADE', 3), { cards: [], kind: 1 })).toBe(false);
  });
});
