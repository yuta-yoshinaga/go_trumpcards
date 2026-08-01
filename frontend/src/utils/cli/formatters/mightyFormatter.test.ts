import { describe, expect, it } from 'vitest';
import type { MightyPlayerData, MightyResponse } from '../../../types/card';
import { formatMightyState } from './mightyFormatter';

function makePlayer(overrides: Partial<MightyPlayerData> = {}): MightyPlayerData {
  return {
    id: 0,
    isHuman: false,
    cardCount: 10,
    cards: [],
    bid: -1,
    bidNoTrump: false,
    isDeclarer: false,
    isPartner: false,
    partnerRevealed: false,
    pointCards: 0,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<MightyResponse> = {}): MightyResponse {
  return {
    players: [
      makePlayer({ id: 0, isHuman: true, cards: [{ design: 'SPADE', value: 1 }] }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
      makePlayer({ id: 4 }),
    ],
    phase: 3,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 0,
    declarerIdx: -1,
    partnerIdx: -1,
    partnerRevealed: false,
    highestBid: 0,
    highestBidder: -1,
    winningBidNoTrump: false,
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, minBid: 13, noTrumpExtra: 2, pointLimit: 100 },
    ...overrides,
  };
}

describe('formatMightyState', () => {
  it('renders the header and round/trick line', () => {
    const out = formatMightyState(makeState());
    expect(out).toContain('Mighty');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
  });

  it('renders trump as No-Trump when winningBidNoTrump is true', () => {
    const out = formatMightyState(makeState({ winningBidNoTrump: true, trumpSuit: 0 }));
    expect(out).toContain('trump: No-Trump');
  });

  it('renders trump suit name when set and not no-trump', () => {
    const out = formatMightyState(makeState({ trumpSuit: 1 }));
    expect(out).toContain('trump: Spade');
  });

  it('falls back to ? for an unknown trump suit value', () => {
    const out = formatMightyState(makeState({ trumpSuit: 99 }));
    expect(out).toContain('trump: ?');
  });

  it('omits trump line when trumpSuit is 0 and not no-trump', () => {
    const out = formatMightyState(makeState({ trumpSuit: 0, winningBidNoTrump: false }));
    expect(out).not.toContain('trump:');
  });

  it('renders the highest bid when > 0', () => {
    const out = formatMightyState(makeState({ highestBid: 14 }));
    expect(out).toContain('bid: 14');
  });

  it('omits the bid line when highestBid is 0', () => {
    const out = formatMightyState(makeState({ highestBid: 0 }));
    expect(out).not.toContain('bid:');
  });

  it('renders the partner card when present', () => {
    const out = formatMightyState(makeState({ partnerCard: { design: 'HEART', value: 1 } }));
    expect(out).toContain('partner:');
  });

  it('omits the partner card line when null', () => {
    const out = formatMightyState(makeState({ partnerCard: null }));
    expect(out).not.toContain('partner:');
  });

  it('shows Declarer role badge on the declarer row', () => {
    const out = formatMightyState(
      makeState({
        players: [
          makePlayer({ id: 0, isHuman: true, isDeclarer: true }),
          makePlayer({ id: 1 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
          makePlayer({ id: 4 }),
        ],
      }),
    );
    expect(out).toContain('[Declarer]');
  });

  it('shows Partner role only when revealed', () => {
    const revealed = formatMightyState(
      makeState({
        players: [
          makePlayer({ id: 0, isHuman: true }),
          makePlayer({ id: 1, isPartner: true, partnerRevealed: true }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
          makePlayer({ id: 4 }),
        ],
      }),
    );
    expect(revealed).toContain('[Partner]');

    const hidden = formatMightyState(
      makeState({
        players: [
          makePlayer({ id: 0, isHuman: true }),
          makePlayer({ id: 1, isPartner: true, partnerRevealed: false }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
          makePlayer({ id: 4 }),
        ],
      }),
    );
    expect(hidden).not.toContain('[Partner]');
  });

  it('renders human hand with indexed cards', () => {
    const out = formatMightyState(
      makeState({
        players: [
          makePlayer({
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'HEART', value: 10 },
            ],
          }),
          makePlayer({ id: 1 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
          makePlayer({ id: 4 }),
        ],
      }),
    );
    expect(out).toMatch(/\[0\]/);
    expect(out).toMatch(/\[1\]/);
  });

  it('omits indexed hand for empty cards', () => {
    const out = formatMightyState(
      makeState({
        players: [
          makePlayer({ id: 0, isHuman: true, cards: [] }),
          makePlayer({ id: 1 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
          makePlayer({ id: 4 }),
        ],
      }),
    );
    expect(out).not.toMatch(/\[0\]/);
  });

  it('renders kitty cards only in phase 2 (kitty exchange)', () => {
    const inKitty = formatMightyState(makeState({ phase: 2, kitty: [{ design: 'SPADE', value: 5 }] }));
    expect(inKitty).toContain('kitty:');

    const outsideKitty = formatMightyState(makeState({ phase: 3, kitty: [{ design: 'SPADE', value: 5 }] }));
    expect(outsideKitty).not.toContain('kitty:');
  });

  it('skips kitty line when kitty is undefined or empty', () => {
    expect(formatMightyState(makeState({ phase: 2, kitty: undefined }))).not.toContain('kitty:');
    expect(formatMightyState(makeState({ phase: 2, kitty: [] }))).not.toContain('kitty:');
  });

  it('renders the current trick when cards are present', () => {
    const out = formatMightyState(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
          { playerIdx: 1, card: { design: 'HEART', value: 13 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('falls back to non-human when player index in trick is out of range', () => {
    const out = formatMightyState(
      makeState({
        currentTrick: [{ playerIdx: 99, card: { design: 'SPADE', value: 1 } }],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a hint line when hint is present', () => {
    const out = formatMightyState(
      makeState({ hint: { reason: 'strategic_bid' }, messageCode: 'mighty.hintRequested' }),
    );
    expect(out).toContain('HINT: strategic_bid');
  });

  it('renders a message line when message is present', () => {
    const out = formatMightyState(makeState({ message: 'Your turn' }));
    expect(out).toContain('Your turn');
  });

  it('renders game-over line when gameEndFlag is true', () => {
    const out = formatMightyState(makeState({ gameEndFlag: true, winnerTeam: 0 }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('Team 0');
  });

  it('always emits a trailing separator', () => {
    const out = formatMightyState(makeState());
    expect(out.split('\n').filter((l) => l.length > 0).length).toBeGreaterThan(0);
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  // このフォーマッタは波括弧なしの 1 行形式で書かれている。
  it('shows the hint only when the hint was requested', () => {
    const hint = { reason: 'strategic_bid' };
    expect(formatMightyState(makeState({ hint, messageCode: 'mighty.hintRequested' }))).toContain('HINT');
    expect(formatMightyState(makeState({ hint, messageCode: 'mighty.playing' }))).not.toContain('HINT');
  });
});
