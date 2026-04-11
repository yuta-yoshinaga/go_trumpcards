import { describe, expect, it } from 'vitest';
import type { CanastaConfig, CanastaPlayerData, CanastaResponse, Card } from '../../types/card';
import { CanastaPhase } from '../../types/phases';
import { getCanastaHint } from './canastaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: CanastaConfig = { cpuDifficulty: 0, pointLimit: 5000 };

function player(overrides: Partial<CanastaPlayerData> = {}): CanastaPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 11,
    cards: [],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    hasCanasta: false,
    hasInitMeld: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<CanastaResponse> = {}): CanastaResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    phase: CanastaPhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: null,
    drawPileCount: 40,
    discardPileCount: 0,
    isFrozen: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getCanastaHint', () => {
  it('returns null when game has ended', () => {
    expect(getCanastaHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getCanastaHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends drawing from stock at start of turn', () => {
    expect(getCanastaHint(makeState())?.reason).toBe('hint.drawStock');
  });

  it('recommends taking the discard pile when player has meld and pile is not frozen', () => {
    const state = makeState({
      discardTop: card('HEART', 5),
      discardPileCount: 3,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getCanastaHint(state)?.reason).toBe('hint.takeDiscardPile');
  });

  it('recommends initial meld in meld phase when player has none', () => {
    const hint = getCanastaHint(makeState({ phase: CanastaPhase.MELD }));
    expect(hint?.reason).toBe('hint.meldInitial');
  });

  it('recommends extending melds in meld phase when player already melded', () => {
    const state = makeState({
      phase: CanastaPhase.MELD,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getCanastaHint(state)?.reason).toBe('hint.meldExtend');
  });

  it('recommends discarding a high safe card in discard phase', () => {
    const hint = getCanastaHint(makeState({ phase: CanastaPhase.DISCARD }));
    expect(hint?.reason).toBe('hint.discardHighSafe');
  });
});
