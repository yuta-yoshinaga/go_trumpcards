import { describe, expect, it } from 'vitest';
import type { Card, TrucoPlayerData, TrucoResponse, TrucoTrickCard } from '../../types/card';
import { getTrucoHint, trucoCardStrength } from './trucoHint';

const card = (value: number, design = 'SPADE'): Card => ({ design, value }) as unknown as Card;

const player = (cards: Card[], overrides: Partial<TrucoPlayerData> = {}): TrucoPlayerData => ({
  id: 0,
  isHuman: true,
  cardCount: cards.length,
  cards,
  trickCount: 0,
  ...overrides,
});

/** PLAY=0, RESPOND=1, TRICK_END=2 (mirrors TrucoPhase). */
const makeState = (overrides: Partial<TrucoResponse> = {}): TrucoResponse =>
  ({
    players: [player([card(4)]), player([], { id: 1, isHuman: false })],
    phase: 0,
    handNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    responderIdx: -1,
    currentTrick: [],
    trickResults: [],
    canDeclareTruco: false,
    gameEndFlag: false,
    ...overrides,
  }) as unknown as TrucoResponse;

const trick = (value: number, design = 'SPADE'): TrucoTrickCard =>
  ({ playerIdx: 1, card: card(value, design) }) as unknown as TrucoTrickCard;

describe('trucoCardStrength', () => {
  it('ranks the four matadores above every common card', () => {
    expect(trucoCardStrength(card(1, 'SPADE'))).toBe(14); // 1 de Espadas
    expect(trucoCardStrength(card(1, 'CLOVER'))).toBe(13); // 1 de Bastos
    expect(trucoCardStrength(card(7, 'SPADE'))).toBe(12); // 7 de Espadas
    expect(trucoCardStrength(card(7, 'DIAMOND'))).toBe(11); // 7 de Oros
    expect(trucoCardStrength(card(3, 'HEART'))).toBe(10); // strongest common
    expect(trucoCardStrength(card(4, 'HEART'))).toBe(1); // weakest common
  });

  it('returns 0 for unused ranks and undefined', () => {
    expect(trucoCardStrength(card(8))).toBe(0);
    expect(trucoCardStrength(undefined)).toBe(0);
  });
});

describe('getTrucoHint', () => {
  it('returns null when there is no state or the game has ended', () => {
    expect(getTrucoHint(null as unknown as TrucoResponse)).toBeNull();
    expect(getTrucoHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when the hand is empty', () => {
    expect(getTrucoHint(makeState({ players: [player([]), player([], { id: 1, isHuman: false })] }))).toBeNull();
  });

  it('recommends declaring Truco with a strong hand when allowed', () => {
    const hint = getTrucoHint(makeState({ canDeclareTruco: true, players: [player([card(1, 'SPADE'), card(4)])] }));
    expect(hint).toEqual({ targetAction: 'truco', reason: 'hintReason.call', confidence: 'strong' });
  });

  it('recommends leading strong when the chosen lead card is strong', () => {
    const hint = getTrucoHint(makeState({ players: [player([card(1, 'SPADE')])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.leadStrong', confidence: 'strong' });
  });

  it('recommends leading low with a weak hand', () => {
    const hint = getTrucoHint(makeState({ players: [player([card(4)])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.leadLow', confidence: 'moderate' });
  });

  it('recommends the cheapest winning card when following', () => {
    const hint = getTrucoHint(
      makeState({ currentTrick: [trick(5, 'DIAMOND')], players: [player([card(4), card(3, 'HEART')])] }),
    );
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.followWin', confidence: 'strong' });
  });

  it('recommends dumping a weak card when it cannot beat the lead', () => {
    const hint = getTrucoHint(makeState({ currentTrick: [trick(3, 'HEART')], players: [player([card(4), card(5)])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.followDump', confidence: 'moderate' });
  });

  it('returns null during PLAY when it is not the human turn', () => {
    expect(getTrucoHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends accepting a call with a strong hand', () => {
    const hint = getTrucoHint(makeState({ phase: 1, responderIdx: 0, players: [player([card(3, 'HEART')])] }));
    expect(hint).toEqual({ targetAction: 'accept', reason: 'hintReason.accept', confidence: 'strong' });
  });

  it('recommends declining a call with a weak hand', () => {
    const hint = getTrucoHint(makeState({ phase: 1, responderIdx: 0, players: [player([card(4)])] }));
    expect(hint).toEqual({ targetAction: 'decline', reason: 'hintReason.decline', confidence: 'moderate' });
  });

  it('returns null during RESPOND when the human is not the responder', () => {
    expect(getTrucoHint(makeState({ phase: 1, responderIdx: 1 }))).toBeNull();
  });

  it('returns null in a non-decision phase (trick end)', () => {
    expect(getTrucoHint(makeState({ phase: 2 }))).toBeNull();
  });
});
