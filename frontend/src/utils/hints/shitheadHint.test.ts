import { describe, expect, it } from 'vitest';
import type { ShitheadResponse } from '../../types/card';
import { getShitheadHint } from './shitheadHint';

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

describe('getShitheadHint', () => {
  it('returns null when game ended', () => {
    expect(getShitheadHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getShitheadHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('hints blind play during facedown phase', () => {
    const r = getShitheadHint(makeState({ currentSource: 'facedown' }));
    expect(r?.reason).toBe('hint.blindPlay');
  });

  it('hints pickup when no playable cards', () => {
    const r = getShitheadHint(makeState());
    expect(r?.reason).toBe('hint.pickup');
  });

  it('hints lead-lowest when discard pile empty', () => {
    const base = makeState();
    base.players[0] = { ...base.players[0], handCards: [{ design: 'SPADE', value: 5 }] };
    expect(getShitheadHint(base)?.reason).toBe('hint.leadLowest');
  });

  it('suggests burning ten on high top', () => {
    const base = makeState({ discardPile: [{ design: 'HEART', value: 11 }] });
    base.players[0] = {
      ...base.players[0],
      handCards: [
        { design: 'SPADE', value: 5 },
        { design: 'CLOVER', value: 10 },
      ],
    };
    expect(getShitheadHint(base)?.reason).toBe('hint.burnTen');
  });

  it('suggests reset 2 when facing high top', () => {
    const base = makeState({ discardPile: [{ design: 'HEART', value: 13 }] });
    base.players[0] = {
      ...base.players[0],
      handCards: [
        { design: 'SPADE', value: 5 },
        { design: 'CLOVER', value: 2 },
      ],
    };
    expect(getShitheadHint(base)?.reason).toBe('hint.resetTwo');
  });

  it('falls back to play-lowest hint', () => {
    const base = makeState({ discardPile: [{ design: 'HEART', value: 4 }] });
    base.players[0] = { ...base.players[0], handCards: [{ design: 'SPADE', value: 5 }] };
    expect(getShitheadHint(base)?.reason).toBe('hint.playLowest');
  });
});
