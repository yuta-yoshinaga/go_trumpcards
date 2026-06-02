import { describe, expect, it } from 'vitest';
import type { Card, ScopaPlayerData, ScopaResponse } from '../../types/card';
import { getScopaHint } from './scopaHint';

const card = (value: number, design = 'SPADE'): Card => ({ design, value }) as unknown as Card;

const player = (cards: Card[], overrides: Partial<ScopaPlayerData> = {}): ScopaPlayerData => ({
  id: 0,
  isHuman: true,
  cardCount: cards.length,
  cards,
  capturedCount: 0,
  scopaCount: 0,
  totalScore: 0,
  ...overrides,
});

const makeState = (overrides: Partial<ScopaResponse> = {}): ScopaResponse =>
  ({
    players: [player([card(7)]), player([], { id: 1, isHuman: false })],
    currentTurn: 0,
    tableCards: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 11, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 30,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  }) as ScopaResponse;

describe('getScopaHint', () => {
  it('returns null when state is missing', () => {
    expect(getScopaHint(null as unknown as ScopaResponse)).toBeNull();
  });

  it('returns null when the game has ended', () => {
    expect(getScopaHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when there is no human player', () => {
    expect(getScopaHint(makeState({ players: [player([], { id: 1, isHuman: false })] }))).toBeNull();
  });

  it('returns null when the human hand is empty', () => {
    expect(getScopaHint(makeState({ players: [player([]), player([], { id: 1, isHuman: false })] }))).toBeNull();
  });

  it('recommends taking the settebello (7 of diamonds)', () => {
    const state = makeState({
      players: [player([card(7)]), player([], { id: 1, isHuman: false })],
      tableCards: [card(7, 'DIAMOND')],
    });
    const hint = getScopaHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.reason).toBe('hint.take.settebello');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends taking diamonds when capturing diamonds (not the 7)', () => {
    const state = makeState({
      players: [player([card(5)]), player([], { id: 1, isHuman: false })],
      tableCards: [card(5, 'DIAMOND')],
    });
    const hint = getScopaHint(state);
    expect(hint?.reason).toBe('hint.take.diamonds');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends a generic capture when no diamonds are involved', () => {
    const state = makeState({
      players: [player([card(5)]), player([], { id: 1, isHuman: false })],
      tableCards: [card(2, 'SPADE'), card(3, 'CLOVER')],
    });
    const hint = getScopaHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.reason).toBe('hint.take.capture');
    expect(hint?.confidence).toBe('moderate');
  });

  it('recommends laying when no capture is possible', () => {
    const state = makeState({
      players: [player([card(5)]), player([], { id: 1, isHuman: false })],
      tableCards: [card(8, 'SPADE'), card(9, 'CLOVER')],
    });
    const hint = getScopaHint(state);
    expect(hint?.targetAction).toBe('lay');
    expect(hint?.reason).toBe('hint.lay.safe');
  });
});
