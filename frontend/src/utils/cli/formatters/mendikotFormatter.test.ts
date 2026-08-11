import { describe, expect, it } from 'vitest';
import type { Card, MendikotResponse } from '../../../types/card';
import { formatMendikotState } from './mendikotFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 10), card('SPADE', 1)] : [],
  tens: 0,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<MendikotResponse> = {}): MendikotResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    handNumber: 2,
    trickNumber: 3,
    trumpSuit: 0,
    trumpChooserIdx: -1,
    tensInDeck: 4,
    teamTens: [2, 1],
    teamTricks: [4, 2],
    scores: [1, 2],
    lastHandWinner: -1,
    lastHandKind: '',
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 3 },
    message: '',
    ...over,
  }) as unknown as MendikotResponse;

describe('formatMendikotState', () => {
  it('reports loading for a null state', () => {
    expect(formatMendikotState(null)).toBe('Loading...');
  });

  it('shows the hand and target', () => {
    const out = formatMendikotState(state());
    expect(out).toContain('hand 2');
    expect(out).toContain('first to 3 hands');
    expect(out).toContain('PLAY');
  });

  // **勝敗を決めるのは 10 の枚数。** 盤面から読めないので必ず出す。
  it('leads with the ten count, not the trick count', () => {
    const out = formatMendikotState(state());
    expect(out).toContain('tens: yours=2 theirs=1 (of 4; three takes the hand)');
    expect(out.indexOf('tens: yours')).toBeLessThan(out.indexOf('tricks: yours'));
    expect(out).toContain('tricks: yours=4 theirs=2 (decides a 2-2 split)');
  });

  // 切り札は未決定と確定の両側を踏む。
  it('says how trump gets decided while it is undecided', () => {
    expect(formatMendikotState(state())).toContain('trump: undecided (the first player who cannot follow');
  });

  it('names the suit and who set it once trump is fixed', () => {
    const out = formatMendikotState(state({ trumpSuit: 3, trumpChooserIdx: 2 }));
    expect(out).toContain('trump: ♥ (set by the first player who could not follow)');
    expect(out).toContain('[set trump]');
  });

  it('marks the playable cards in your hand', () => {
    const out = formatMendikotState(state());
    expect(out).toMatch(/your hand: \[0\].*\*/);
  });

  it('shows the current trick when one is in progress', () => {
    const out = formatMendikotState(state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 5) }] }));
    expect(out).toContain('trick: ');
  });

  // **決まり方で 1/2/3 点と変わる。** 4 通りすべてが別の文言になる。
  it.each([
    ['tens', 'took more tens (+1)'],
    ['tricks', 'wins on tricks (+1)'],
    ['mendikot', 'all four tens: Mendikot, +2'],
    ['whitewash', 'every trick: Whitewash, +3'],
  ])('explains a hand decided by %s', (kind, expected) => {
    const out = formatMendikotState(state({ phase: 1, lastHandWinner: 0, lastHandKind: kind }));
    expect(out).toContain(expected);
  });

  it('falls back to a plain line for an unknown outcome', () => {
    const out = formatMendikotState(state({ phase: 1, lastHandWinner: 1, lastHandKind: 'nope' }));
    expect(out).toContain('hand over — team 1 takes it');
  });

  it.each([
    [0, 'team 0 wins'],
    [1, 'team 1 wins'],
    [-1, 'tie'],
  ])('reports the game result for winnerTeam %s', (winnerTeam, expected) => {
    const out = formatMendikotState(state({ phase: 2, gameEndFlag: true, winnerTeam }));
    expect(out).toContain(expected);
  });

  // **配列も PHASE_NAMES も欠けることがある。** 既定値側の枝を踏む。
  it('falls back to zeroes and a raw phase when the arrays are empty', () => {
    const out = formatMendikotState(
      state({ phase: 99, teamTens: [], teamTricks: [], scores: [] } as Partial<MendikotResponse>),
    );
    expect(out).toContain('tens: yours=0 theirs=0');
    expect(out).toContain('tricks: yours=0 theirs=0');
    expect(out).toContain('hand points: yours=0 theirs=0');
    expect(out).toContain('99');
  });

  it('falls back to ? for an unknown trump suit code', () => {
    expect(formatMendikotState(state({ trumpSuit: 9 }))).toContain('trump: ?');
  });

  it('says the hand is empty rather than printing nothing', () => {
    const out = formatMendikotState(
      state({ players: [{ ...seat(0), cards: [] }, seat(1), seat(2), seat(3)] } as Partial<MendikotResponse>),
    );
    expect(out).toContain('your hand: (empty)');
  });

  it('omits the hand block when no seat is human', () => {
    const out = formatMendikotState(state({ players: [seat(1), seat(2), seat(3)] } as Partial<MendikotResponse>));
    expect(out).not.toContain('your hand:');
  });

  it('marks nothing when no card is playable', () => {
    const out = formatMendikotState(state({ validPlays: [] } as Partial<MendikotResponse>));
    expect(out).not.toContain('*');
  });

  // 終局時の得点欠落も既定値に落ちる。
  it('reports a game result with no scores recorded', () => {
    const out = formatMendikotState(
      state({ phase: 2, gameEndFlag: true, winnerTeam: 0, scores: [] } as Partial<MendikotResponse>),
    );
    expect(out).toContain('team 0 wins (0 - 0)');
  });

  it('keeps the hand-end block hidden before any hand has ended', () => {
    expect(formatMendikotState(state({ phase: 1, lastHandWinner: -1 }))).not.toContain('hand over');
  });

  it('echoes a server message', () => {
    expect(formatMendikotState(state({ message: 'follow suit' }))).toContain('follow suit');
  });
});
