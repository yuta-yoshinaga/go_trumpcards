import { describe, expect, it } from 'vitest';
import type { Card, SixCardGolfResponse, SixCardGolfSlot } from '../../types/card';
import { getSixcardgolfHint } from './sixcardgolfHint';

function makeCard(value: number, design: Card['design'] = 'SPADE'): Card {
  return { value, design };
}

function makeSlot(card: Card | null, faceUp: boolean): SixCardGolfSlot {
  return { card, faceUp };
}

function makeDefaultGrid(): SixCardGolfSlot[] {
  return [
    makeSlot(makeCard(5), true),
    makeSlot(makeCard(8), true),
    makeSlot(makeCard(3), false),
    makeSlot(makeCard(7), false),
    makeSlot(makeCard(9), false),
    makeSlot(makeCard(4), false),
  ];
}

function makeState(overrides: Partial<SixCardGolfResponse> = {}): SixCardGolfResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        grid: makeDefaultGrid(),
        roundScore: 0,
        cumulativeScore: 0,
        allFaceUp: false,
      },
      {
        id: 1,
        isHuman: false,
        grid: makeDefaultGrid(),
        roundScore: 0,
        cumulativeScore: 0,
        allFaceUp: false,
      },
    ],
    phase: 1,
    roundNumber: 1,
    totalRounds: 9,
    currentPlayerIdx: 0,
    discardTop: makeCard(7),
    drawPileCount: 30,
    drawnCard: null,
    drawnFromDiscard: false,
    canFlip: false,
    finalTurnTrigger: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { playerCount: 2, cpuDifficulty: 1, rounds: 9 },
    ...overrides,
  };
}

describe('getSixcardgolfHint', () => {
  it('returns null for roundOver phase (3)', () => {
    expect(getSixcardgolfHint(makeState({ phase: 3 }))).toBeNull();
  });

  it('returns null for gameOver phase (4)', () => {
    expect(getSixcardgolfHint(makeState({ phase: 4 }))).toBeNull();
  });

  it('returns drawDiscard hint when discard top is a King (value 13)', () => {
    const result = getSixcardgolfHint(makeState({ phase: 1, discardTop: makeCard(13) }));
    expect(result).toEqual({
      targetAction: 'drawDiscard',
      reason: 'hintReason.drawDiscardLow',
      confidence: 'strong',
    });
  });

  it('returns drawStock hint when discard top is high', () => {
    const result = getSixcardgolfHint(makeState({ phase: 1, discardTop: makeCard(10) }));
    expect(result).toEqual({
      targetAction: 'drawStock',
      reason: 'hintReason.drawStock',
      confidence: 'moderate',
    });
  });

  it('returns swap hint when drawn card is low', () => {
    const result = getSixcardgolfHint(makeState({ phase: 2, drawnCard: makeCard(2) }));
    expect(result).toEqual({
      targetAction: 'swap',
      reason: 'hintReason.swapHigh',
      confidence: 'strong',
      // **どのマスかまで返す。**返さないと 6 マスから目視で探し直させる (#4887)。
      // 既定グリッドの表向きは 5(0) と 8(1)。最も高いのは 8 なので位置 1。
      targetPos: 1,
    });
  });

  it('returns discard hint when drawn card is high', () => {
    const result = getSixcardgolfHint(makeState({ phase: 2, drawnCard: makeCard(12) }));
    expect(result).toEqual({
      targetAction: 'discard',
      reason: 'hintReason.discardHigh',
      confidence: 'strong',
    });
  });

  it('returns null when no human player found', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getSixcardgolfHint(state)).toBeNull();
  });

  it('returns flip hint when canFlip is true in playerTurn phase', () => {
    const result = getSixcardgolfHint(makeState({ phase: 1, canFlip: true }));
    expect(result).toEqual({
      targetAction: 'flip',
      reason: 'hintReason.flipCard',
      confidence: 'moderate',
    });
  });

  it('returns column match hint when drawn card matches a visible column partner', () => {
    const grid = makeDefaultGrid();
    // grid[0] is face-up with value 5, grid[3] is face-down with value 7
    // Drawn card value 5 should suggest column match (swap into grid[3])
    const state = makeState({
      phase: 2,
      drawnCard: makeCard(5),
      players: [
        { id: 0, isHuman: true, grid, roundScore: 0, cumulativeScore: 0, allFaceUp: false },
        { id: 1, isHuman: false, grid: makeDefaultGrid(), roundScore: 0, cumulativeScore: 0, allFaceUp: false },
      ],
    });
    const result = getSixcardgolfHint(state);
    expect(result).toEqual({
      targetAction: 'swap',
      reason: 'hintReason.columnMatch',
      confidence: 'strong',
      // 列を揃えられるマスそのもの（grid[0] の相方 = grid[3]）を指す。
      targetPos: 3,
    });
  });
});
