import { describe, expect, it } from 'vitest';
import type { CardDesign, ZwickerCard, ZwickerPlayer, ZwickerResponse } from '../../../types/card';
import { formatZwickerState } from './zwickerFormatter';

const card = (design: CardDesign, value: number, values: number[]): ZwickerCard => ({ design, value, values });

function seat(id: number, isHuman: boolean, overrides?: Partial<ZwickerPlayer>): ZwickerPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 4,
    cards: isHuman ? [card('SPADE', 1, [1, 11])] : [],
    capturedCount: 6,
    zwicks: 1,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ZwickerResponse>): ZwickerResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 0,
    currentPlayerIdx: 0,
    stockCount: 20,
    tableCards: [card('HEART', 13, [4, 14]), card('CLOVER', 7, [7])],
    builds: [
      {
        owner: 1,
        value: 9,
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 4 },
        ],
      },
    ],
    teamScores: [12, 8],
    targetScore: 61,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  };
}

describe('formatZwickerState', () => {
  it('prints both rules every frame', () => {
    // 「A・絵札は2つの値」と「Zwickは場を空にすること」が最も誤解されやすい。
    const out = formatZwickerState(makeState());
    expect(out).toContain('A=1/11');
    expect(out).toContain('clearing the table is a Zwick');
  });

  it('prints the matching values, which are unreadable from the rank', () => {
    const out = formatZwickerState(makeState());
    expect(out).toContain('(4/14)'); // a king
    expect(out).toContain('(7)');
    expect(out).toContain('(1/11)'); // the ace in hand
  });

  it('prints the build with its declared value', () => {
    // ビルドは宣言値ちょうどでしか取れないので、値が出ていないと判断できない。
    expect(formatZwickerState(makeState())).toContain('build[0] worth 9 (seat 1)');
  });

  it('prints the scores, the target and each seat', () => {
    const out = formatZwickerState(makeState());
    expect(out).toContain('us 12 them 8');
    expect(out).toContain('first to 61');
    expect(out).toContain('6 taken, 1 zwick(s)');
    expect(out).toContain('4 cards'); // the hidden hand is a count
  });

  it('shows an empty table rather than nothing', () => {
    expect(formatZwickerState(makeState({ tableCards: [] }))).toContain('table: -');
  });

  it('says when the majority was level, so the total still adds up', () => {
    const out = formatZwickerState(
      makeState({
        lastRound: { cardPoints: [10, 10], cards: [27, 27], majorityTeam: -1, zwicks: [0, 0], total: [10, 10] },
      }),
    );
    expect(out).toContain('the card counts were level');
  });

  it('does not say that when a team did take the majority', () => {
    const out = formatZwickerState(
      makeState({
        lastRound: { cardPoints: [17, 10], cards: [30, 25], majorityTeam: 0, zwicks: [1, 0], total: [21, 10] },
      }),
    );
    expect(out).toContain('last deal: us 21, them 10');
    expect(out).not.toContain('the card counts were level');
  });

  it('reports each ending', () => {
    expect(formatZwickerState(makeState({ gameEndFlag: true, winnerTeam: 0 }))).toContain('your team wins');
    expect(formatZwickerState(makeState({ gameEndFlag: true, winnerTeam: 1 }))).toContain('the other team wins');
  });
});
