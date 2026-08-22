import { describe, expect, it } from 'vitest';
import type { Card, SevenTwentySevenResponse } from '../../../types/card';
import { formatSevenTwentySevenState } from './seventwentysevenFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const player = (
  id: number,
  overrides: Partial<SevenTwentySevenResponse['players'][number]> = {},
): SevenTwentySevenResponse['players'][number] => ({
  id,
  isHuman: id === 0,
  chips: 200,
  standing: false,
  out: false,
  roundBet: 10,
  cardCount: 2,
  cards: id === 0 ? [card('SPADE', 4), card('HEART', 13)] : [],
  lowScore: id === 0 ? '4.5' : '',
  highScore: id === 0 ? '4.5' : '',
  wonLow: false,
  wonHigh: false,
  ...overrides,
});

const state = (overrides: Partial<SevenTwentySevenResponse> = {}): SevenTwentySevenResponse =>
  ({
    players: [player(0), player(1)],
    phase: 0,
    roundNumber: 1,
    drawRound: 1,
    pot: 40,
    carryPot: 0,
    carryCount: 0,
    ante: 10,
    chips: 200,
    lowWinner: -1,
    highWinner: -1,
    matchWinnerIdx: -1,
    result: 0,
    gameEndFlag: false,
    config: { playerCount: 2, ante: 10, startingChips: 200, targetRounds: 10 },
    message: '',
    ...overrides,
  }) as SevenTwentySevenResponse;

describe('formatSevenTwentySevenState', () => {
  // **2 つの目標とカードの点数を毎回書く。** ここを落とすと、CLI から
  // 何を狙えばいいのかが読めない。
  it('always states both targets and the unusual card values', () => {
    const out = formatSevenTwentySevenState(state());
    expect(out).toContain('7 and 27');
    expect(out).toMatch(/faces = 0\.5/);
    expect(out).toMatch(/ace = 1 or 11/);
  });

  it('shows the round, the draw pass and the pot', () => {
    const out = formatSevenTwentySevenState(state({ roundNumber: 3, drawRound: 2, pot: 90 }));
    expect(out).toContain('round: 3');
    expect(out).toContain('draw: 2');
    expect(out).toContain('pot: 90');
  });

  it('names the phase', () => {
    expect(formatSevenTwentySevenState(state({ phase: 0 }))).toContain('Draw');
    expect(formatSevenTwentySevenState(state({ phase: 1 }))).toContain('Result');
  });

  // **両側の得点を出す。** 片方だけでは、いま何を狙えるのかが読めない。
  it('prints both totals for a visible hand', () => {
    const out = formatSevenTwentySevenState(state());
    expect(out).toContain('4.5 / 4.5');
  });

  it('prints a busted side as a dash', () => {
    const out = formatSevenTwentySevenState(
      state({ players: [player(0, { lowScore: '-', highScore: '19' }), player(1)] }),
    );
    expect(out).toContain('- / 19');
  });

  it.each([
    [{ out: true }, 'OUT'],
    [{ standing: true }, 'stood pat'],
    [{}, 'drawing'],
    [{ wonLow: true }, 'won low'],
    [{ wonHigh: true }, 'won high'],
    [{ wonLow: true, wonHigh: true }, 'SCOOP'],
  ])('labels the status %o as %s', (overrides, label) => {
    const out = formatSevenTwentySevenState(state({ players: [player(0, overrides), player(1)] }));
    expect(out).toContain(label);
  });

  // **総取りは専用のラベル。** 「won low」と「won high」を並べると読み違える。
  it('prefers SCOOP over the single-side labels', () => {
    const out = formatSevenTwentySevenState(
      state({ players: [player(0, { wonLow: true, wonHigh: true }), player(1)] }),
    );
    expect(out).toContain('SCOOP');
    expect(out).not.toContain('won low');
  });

  it('renders the hint only when it was requested', () => {
    const hint = { draw: true, reason: 'chase_twentyseven' };
    const requested = formatSevenTwentySevenState(state({ hint, messageCode: 'seventwentyseven.hintRequested' }));
    expect(requested).toContain('take a card');
    expect(requested).toContain('chase_twentyseven');

    // 受動ヒント (毎レスポンスに載る) は CLI には出さない。
    expect(formatSevenTwentySevenState(state({ hint }))).not.toContain('HINT:');
  });

  it('says stand pat for a non-drawing hint', () => {
    const out = formatSevenTwentySevenState(
      state({ hint: { draw: false, reason: 'exactly_seven' }, messageCode: 'seventwentyseven.hintRequested' }),
    );
    expect(out).toContain('stand pat');
  });

  it('announces the end of the match', () => {
    const out = formatSevenTwentySevenState(state({ gameEndFlag: true, matchWinnerIdx: 1 }));
    expect(out).toContain('Game Over!');
  });

  it('shows the message when there is one', () => {
    expect(formatSevenTwentySevenState(state({ message: 'you took neither side' }))).toContain('you took neither side');
  });
});
