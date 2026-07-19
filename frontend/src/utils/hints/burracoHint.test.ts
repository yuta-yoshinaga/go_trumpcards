import { describe, expect, it } from 'vitest';
import type { BurracoConfig, BurracoPlayerData, BurracoResponse, Card } from '../../types/card';
import { BurracoPhase } from '../../types/phases';
import { getBurracoHint } from './burracoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: BurracoConfig = { cpuDifficulty: 0, pointLimit: 5000 };

function player(overrides: Partial<BurracoPlayerData> = {}): BurracoPlayerData {
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
    hasBurraco: false,
    hasInitMeld: false,
    tookPozzetto: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<BurracoResponse> = {}): BurracoResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    phase: BurracoPhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: null,
    discardPile: [],
    drawPileCount: 40,
    discardPileCount: 0,
    pozzettoCount: 2,
    isFrozen: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getBurracoHint', () => {
  it('returns null when game has ended', () => {
    expect(getBurracoHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getBurracoHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends drawing from stock at start of turn', () => {
    expect(getBurracoHint(makeState())?.reason).toBe('hint.drawStock');
  });

  it('recommends taking the discard pile when player has meld and pile is not frozen', () => {
    const state = makeState({
      discardTop: card('HEART', 5),
      discardPileCount: 3,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getBurracoHint(state)?.reason).toBe('hint.takeDiscardPile');
  });

  it('recommends initial meld in meld phase when player has none', () => {
    const hint = getBurracoHint(makeState({ phase: BurracoPhase.MELD }));
    expect(hint?.reason).toBe('hint.meldInitial');
  });

  it('recommends extending melds in meld phase when player already melded', () => {
    const state = makeState({
      phase: BurracoPhase.MELD,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getBurracoHint(state)?.reason).toBe('hint.meldExtend');
  });

  it('recommends discarding a high safe card in discard phase', () => {
    const hint = getBurracoHint(makeState({ phase: BurracoPhase.DISCARD }));
    expect(hint?.reason).toBe('hint.discardHighSafe');
  });
});
