import { describe, expect, it } from 'vitest';
import type { Card, PutPlayerData, PutResponse, PutTrickCard } from '../../types/card';
import { getPutHint, putCardStrength } from './putHint';

const card = (value: number, design = 'SPADE'): Card => ({ design, value }) as unknown as Card;

const player = (cards: Card[], overrides: Partial<PutPlayerData> = {}): PutPlayerData => ({
  id: 0,
  isHuman: true,
  cardCount: cards.length,
  cards,
  trickCount: 0,
  ...overrides,
});

/** PLAY=0, RESPOND=1, TRICK_END=2 (mirrors PutPhase). */
const makeState = (overrides: Partial<PutResponse> = {}): PutResponse =>
  ({
    players: [player([card(4)]), player([], { id: 1, isHuman: false })],
    phase: 0,
    handNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    responderIdx: -1,
    currentTrick: [],
    trickResults: [],
    canDeclarePut: false,
    gameEndFlag: false,
    ...overrides,
  }) as unknown as PutResponse;

const trick = (value: number, design = 'SPADE'): PutTrickCard =>
  ({ playerIdx: 1, card: card(value, design) }) as unknown as PutTrickCard;

describe('putCardStrength', () => {
  // The order is 3-2-A-K-Q-J-10-9-8-7-6-5-4 and suits do not matter. These
  // assertions used to encode Truco's ranking (four suit-specific matadores
  // above a common ladder, 8/9/10 worth 0) — the clone source's rules, not
  // Put's — so the suite was green while testing the wrong game.
  it('ranks 3 highest and 4 lowest, matching the Go domain', () => {
    expect(putCardStrength(card(3, 'HEART'))).toBe(13);
    expect(putCardStrength(card(2, 'CLOVER'))).toBe(12);
    expect(putCardStrength(card(1, 'SPADE'))).toBe(11);
    expect(putCardStrength(card(13, 'SPADE'))).toBe(10);
    expect(putCardStrength(card(4, 'HEART'))).toBe(1);
  });

  it('uses all 52 cards — 8, 9 and 10 are real ranks, not dead ones', () => {
    // Truco strips these from the deck; Put does not. Returning 0 made the
    // in-app hint recommend discarding them as worthless.
    expect(putCardStrength(card(10))).toBe(7);
    expect(putCardStrength(card(9))).toBe(6);
    expect(putCardStrength(card(8))).toBe(5);
  });

  it('ignores suit entirely', () => {
    for (const value of [1, 3, 7, 13]) {
      const want = putCardStrength(card(value, 'SPADE'));
      for (const design of ['HEART', 'DIAMOND', 'CLOVER']) {
        expect(putCardStrength(card(value, design))).toBe(want);
      }
    }
  });

  it('assigns each of the 13 ranks a distinct strength from 1 to 13', () => {
    const seen = new Set<number>();
    for (let v = 1; v <= 13; v++) seen.add(putCardStrength(card(v)));
    expect(seen.size).toBe(13);
    expect(Math.min(...seen)).toBe(1);
    expect(Math.max(...seen)).toBe(13);
  });

  it('returns 0 for undefined and out-of-range values', () => {
    expect(putCardStrength(undefined)).toBe(0);
    expect(putCardStrength(card(0))).toBe(0);
    expect(putCardStrength(card(14))).toBe(0);
  });
});

describe('getPutHint', () => {
  it('returns null when there is no state or the game has ended', () => {
    expect(getPutHint(null as unknown as PutResponse)).toBeNull();
    expect(getPutHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when the hand is empty', () => {
    expect(getPutHint(makeState({ players: [player([]), player([], { id: 1, isHuman: false })] }))).toBeNull();
  });

  it('recommends declaring Put with a strong hand when allowed', () => {
    const hint = getPutHint(makeState({ canDeclarePut: true, players: [player([card(1, 'SPADE'), card(4)])] }));
    expect(hint).toEqual({ targetAction: 'put', reason: 'hintReason.call', confidence: 'strong' });
  });

  it('recommends leading strong when the chosen lead card is strong', () => {
    const hint = getPutHint(makeState({ players: [player([card(1, 'SPADE')])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.leadStrong', confidence: 'strong' });
  });

  it('recommends leading low with a weak hand', () => {
    const hint = getPutHint(makeState({ players: [player([card(4)])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.leadLow', confidence: 'moderate' });
  });

  it('recommends the cheapest winning card when following', () => {
    const hint = getPutHint(
      makeState({ currentTrick: [trick(5, 'DIAMOND')], players: [player([card(4), card(3, 'HEART')])] }),
    );
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.followWin', confidence: 'strong' });
  });

  it('recommends dumping a weak card when it cannot beat the lead', () => {
    const hint = getPutHint(makeState({ currentTrick: [trick(3, 'HEART')], players: [player([card(4), card(5)])] }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hintReason.followDump', confidence: 'moderate' });
  });

  it('returns null during PLAY when it is not the human turn', () => {
    expect(getPutHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends accepting a call with a strong hand', () => {
    const hint = getPutHint(makeState({ phase: 1, responderIdx: 0, players: [player([card(3, 'HEART')])] }));
    expect(hint).toEqual({ targetAction: 'accept', reason: 'hintReason.accept', confidence: 'strong' });
  });

  it('recommends declining a call with a weak hand', () => {
    const hint = getPutHint(makeState({ phase: 1, responderIdx: 0, players: [player([card(4)])] }));
    expect(hint).toEqual({ targetAction: 'decline', reason: 'hintReason.decline', confidence: 'moderate' });
  });

  it('returns null during RESPOND when the human is not the responder', () => {
    expect(getPutHint(makeState({ phase: 1, responderIdx: 1 }))).toBeNull();
  });

  it('returns null in a non-decision phase (trick end)', () => {
    expect(getPutHint(makeState({ phase: 2 }))).toBeNull();
  });
});
