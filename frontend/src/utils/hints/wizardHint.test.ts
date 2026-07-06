import { describe, expect, it } from 'vitest';
import type { Card, WizardResponse } from '../../types/card';
import { WizardPhase } from '../../types/phases';
import { getWizardHint } from './wizardHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });
/** A Wizard special card (always wins). */
const wizard = (): Card => ({
  design: 'JOKER',
  value: 0,
  label: 'Wizard',
  deck: 'wizard',
  glyph: '✦',
  color: 'purple',
});
/** A Jester special card (always loses). */
const jester = (): Card => ({
  design: 'JOKER',
  value: 0,
  label: 'Jester',
  deck: 'wizard',
  glyph: '✦',
  color: 'purple',
});

function makeState(overrides: Partial<WizardResponse> = {}): WizardResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    ],
    phase: WizardPhase.BID,
    roundNumber: 5,
    totalRounds: 15,
    handSize: 5,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    trumpCard: card('HEART', 10),
    trumpSuit: 3,
    restrictedBid: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 0 },
    message: '',
    ...overrides,
  };
}

describe('getWizardHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getWizardHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getWizardHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getWizardHint(makeState({ phase: WizardPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getWizardHint(makeState({ phase: WizardPhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getWizardHint(makeState({ phase: WizardPhase.GAME_END }))).toBeNull();
  });

  // Bid phase
  it('returns null in BID phase when not human bid turn', () => {
    expect(getWizardHint(makeState({ phase: WizardPhase.BID, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('returns bid hint with strong confidence for many high cards', () => {
    const state = makeState({ phase: WizardPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('HEART', 12), card('DIAMOND', 11)];
    const result = getWizardHint(state);
    expect(result?.targetAction).toMatch(/^bid:\d+$/);
    expect(result?.reason).toBe('hint.bidEstimate');
    expect(result?.confidence).toBe('strong');
  });

  it('counts Wizards as near-certain tricks in the bid estimate', () => {
    const state = makeState({ phase: WizardPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [wizard(), wizard(), card('SPADE', 14), card('CLOVER', 3)];
    const result = getWizardHint(state);
    const bidValue = Number(result?.targetAction.split(':')[1]);
    // 2 Wizards (certain) + a high card → at least 2 tricks, strong confidence.
    expect(bidValue).toBeGreaterThanOrEqual(2);
    expect(result?.confidence).toBe('strong');
  });

  it('returns bid hint with moderate confidence for few high cards', () => {
    const state = makeState({ phase: WizardPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.bidEstimate');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase
  it('returns null in PLAY phase when not human turn', () => {
    expect(getWizardHint(makeState({ phase: WizardPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests playing a Wizard to guarantee the trick', () => {
    const state = makeState({ phase: WizardPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [wizard(), card('CLOVER', 5)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.playWizard');
    expect(result?.confidence).toBe('strong');
  });

  it('does not suggest a Wizard once one is already in the trick', () => {
    const state = makeState({
      phase: WizardPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: wizard() }],
    });
    state.players[0].cards = [wizard(), card('CLOVER', 5)];
    const result = getWizardHint(state);
    expect(result?.reason).not.toBe('hint.playWizard');
  });

  it('suggests strategic lead when leading', () => {
    const state = makeState({ phase: WizardPhase.PLAY, currentPlayerIdx: 0 });
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.leadStrategic');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests following suit', () => {
    const state = makeState({
      phase: WizardPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests dumping a Jester when void in led suit', () => {
    const state = makeState({
      phase: WizardPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [jester(), card('SPADE', 14)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.dumpJester');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests trumping when void in led suit and have trump', () => {
    const state = makeState({
      phase: WizardPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.trumpWithCard');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests discarding lowest when void with no trump', () => {
    const state = makeState({
      phase: WizardPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getWizardHint(state);
    expect(result?.reason).toBe('hint.discardLowest');
    expect(result?.confidence).toBe('moderate');
  });
});
