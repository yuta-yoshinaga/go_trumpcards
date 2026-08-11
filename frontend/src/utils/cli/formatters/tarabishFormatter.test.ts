import { describe, expect, it } from 'vitest';
import type { Card, TarabishResponse } from '../../../types/card';
import { formatTarabishState } from './tarabishFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  meldPoints: 0,
  runLen: 0,
  hasBella: false,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<TarabishResponse> = {}): TarabishResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    upCard: card('HEART', 9),
    trumpTakerIdx: -1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    scores: [120, 80],
    roundPoints: [0, 0],
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 500 },
    message: '',
    ...overrides,
  } as unknown as TarabishResponse;
}

describe('formatTarabishState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatTarabishState(null)).toBe('Loading...');
  });

  it('renders the round, trick, target, score and the trump order', () => {
    const out = formatTarabishState(makeState());
    expect(out).toContain('round 2');
    expect(out).toContain('trick 4/9');
    expect(out).toContain('first to 500');
    expect(out).toContain('yours=120 theirs=80');
    // **切り札の序列は盤面から読めない。** 常時出す。
    expect(out).toContain('J(Jass)=20 > 9(Menel)=14');
  });

  // 入札前は候補、決まったあとは切り札。両側を踏む。
  it('shows the turned card before trump is settled', () => {
    const out = formatTarabishState(makeState({ trumpTakerIdx: -1 }));
    expect(out).toContain('turned for trump:');
  });

  it('shows who took trump once it is settled', () => {
    const out = formatTarabishState(makeState({ trumpTakerIdx: 2 }));
    expect(out).toContain('taken by');
    expect(out).not.toContain('turned for trump:');
  });

  // メルドの内訳が出る。無い席との両側を踏む。
  it('renders the meld breakdown per seat', () => {
    const out = formatTarabishState(
      makeState({
        players: [seat(0, { meldPoints: 70, runLen: 4, hasBella: true }), seat(1), seat(2), seat(3)],
      } as Partial<TarabishResponse>),
    );
    expect(out).toContain('run of 4+bella=70');
    expect(out).toContain('no meld');
  });

  // チーム番号が出る。向かい合う席が味方であることは盤面から読めない。
  it('labels each seat with its team', () => {
    const out = formatTarabishState(makeState());
    expect(out).toContain('[T0]');
    expect(out).toContain('[T1]');
  });

  it('marks legal cards and leaves illegal ones unmarked', () => {
    const handLine = formatTarabishState(makeState({ validPlays: [0] }))
      .split('\n')
      .find((l) => l.startsWith('your hand:'));
    expect(handLine?.match(/\*/g)).toHaveLength(1);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatTarabishState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('HEART', 11) }] } as Partial<TarabishResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'team 0 wins'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerTeam %i', (winnerTeam, expected) => {
    const out = formatTarabishState(makeState({ gameEndFlag: true, winnerTeam, phase: 3 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatTarabishState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
