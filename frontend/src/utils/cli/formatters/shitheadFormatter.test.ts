import { describe, expect, it } from 'vitest';
import type { ShitheadResponse } from '../../../types/card';
import { formatShitheadState } from './shitheadFormatter';

function makeState(overrides?: Partial<ShitheadResponse>): ShitheadResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: 0,
        handCount: 3,
        handCards: [],
        faceUpCards: [],
        faceDownCount: 3,
      },
      {
        id: 1,
        isHuman: false,
        isFinished: false,
        rank: 0,
        handCount: 3,
        handCards: [],
        faceUpCards: [],
        faceDownCount: 3,
      },
    ],
    currentTurn: 0,
    currentSource: 'hand',
    discardPile: [],
    stockSize: 30,
    skipNext: false,
    sevenActive: false,
    gameEndFlag: false,
    config: {
      magicTwo: true,
      magicSeven: true,
      magicEight: true,
      magicTen: true,
      fourOfAKindBurn: true,
      cpuDifficulty: 1,
    },
    cpuActions: [],
    message: '',
    ...overrides,
  };
}

describe('formatShitheadState', () => {
  it('renders header and basic state', () => {
    const result = formatShitheadState(makeState());
    expect(result).toContain('Shithead');
    expect(result).toContain('stock: 30');
  });

  it('marks current turn with asterisk', () => {
    const result = formatShitheadState(makeState());
    expect(result).toMatch(/\*/);
  });

  it('shows human hand with indices', () => {
    const result = formatShitheadState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            isFinished: false,
            rank: 0,
            handCount: 1,
            handCards: [{ design: 'SPADE', value: 5 }],
            faceUpCards: [],
            faceDownCount: 3,
          },
        ],
      }),
    );
    expect(result).toContain('hand:');
    expect(result).toContain('[0]');
  });

  it('shows seven-active flag', () => {
    const result = formatShitheadState(makeState({ sevenActive: true }));
    expect(result).toContain('seven active');
  });

  it('shows skip-next flag', () => {
    const result = formatShitheadState(makeState({ skipNext: true }));
    expect(result).toContain('skipped');
  });

  it('renders empty discard pile placeholder', () => {
    const result = formatShitheadState(makeState({ discardPile: [] }));
    expect(result).toContain('discard top: [  ]');
  });

  it('renders human face-up cards when present', () => {
    const result = formatShitheadState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            isFinished: false,
            rank: 0,
            handCount: 0,
            handCards: [],
            faceUpCards: [{ design: 'SPADE', value: 11 }],
            faceDownCount: 3,
          },
        ],
      }),
    );
    expect(result).toContain('faceUp:');
  });

  it('renders finished player rank tag', () => {
    const result = formatShitheadState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            isFinished: true,
            rank: 1,
            handCount: 0,
            handCards: [],
            faceUpCards: [],
            faceDownCount: 0,
          },
        ],
      }),
    );
    expect(result).toContain('[done #1]');
  });

  it('renders message when present', () => {
    const result = formatShitheadState(makeState({ message: 'Magic 7 played' }));
    expect(result).toContain('Magic 7 played');
  });

  it('shows game over with loser', () => {
    const result = formatShitheadState(
      makeState({
        gameEndFlag: true,
        players: [
          {
            id: 0,
            isHuman: true,
            isFinished: false,
            rank: 0,
            handCount: 5,
            handCards: [],
            faceUpCards: [],
            faceDownCount: 0,
          },
          {
            id: 1,
            isHuman: false,
            isFinished: true,
            rank: 1,
            handCount: 0,
            handCards: [],
            faceUpCards: [],
            faceDownCount: 0,
          },
        ],
      }),
    );
    expect(result).toContain('Game Over');
  });
});
