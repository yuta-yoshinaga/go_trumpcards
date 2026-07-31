import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, SjavsPlayer, SjavsResponse } from '../../../types/card';
import { formatSjavsState } from './sjavsFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<SjavsPlayer>): SjavsPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 8,
    cards: isHuman ? [card('CLOVER', 12), card('HEART', 1)] : [],
    bid: 0,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<SjavsResponse>): SjavsResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 1,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    trumpSuit: 2,
    trumpCount: 13,
    bidderIdx: 0,
    bidLength: 6,
    minBid: 5,
    myLongest: 6,
    trick: [],
    trickNo: 0,
    validIndices: [0],
    teamPoints: [30, 20],
    remaining: [24, 24],
    crosses: [0, 0],
    carryOver: 0,
    gameEndFlag: false,
    winnerTeam: -1,
    doubleVictory: false,
    message: '',
    ...overrides,
  };
}

describe('formatSjavsState', () => {
  it('prints the permanent trumps every frame', () => {
    // 切札スートの札しか切札でないと思い込むのが定番の誤解なので、毎回出す。
    expect(formatSjavsState(makeState())).toContain('CQ > SQ > CJ > SJ > HJ > DJ');
  });

  it('says the trump is undecided rather than inventing one', () => {
    const bidding = makeState({ trumpSuit: -1, trumpCount: 0, phase: 0 });
    expect(formatSjavsState(bidding)).toContain('trump undecided');
    // 未確定のうちは枚数も出さない。
    expect(formatSjavsState(bidding)).not.toContain('trumps in this suit');
    expect(formatSjavsState(makeState())).toContain('trumps in this suit: 13');
  });

  it('prints the rubber countdown and the hand points', () => {
    const out = formatSjavsState(makeState({ remaining: [10, 24] }));
    expect(out).toContain('to go: us 10 / them 24');
    expect(out).toContain('us 30 / them 20 (120 in all)');
  });

  it('shows each seat team and bid, hiding only the hand', () => {
    const out = formatSjavsState(makeState({ players: [seat(0, true), seat(1, false, { bid: 6 })] }));
    expect(out).toContain('you [team 0]');
    expect(out).toContain('cpu1 [team 1] bid 6');
    expect(out).toContain('8 cards');
  });

  it('reports the 60-60 tie distinctly from a score', () => {
    // 何も出ないと、入力を取りこぼしたように見える。
    const tie = formatSjavsState(
      makeState({
        handResult: {
          declarerTeam: 0,
          declarerPoints: 60,
          scoringTeam: -1,
          amount: 0,
          vol: false,
          trumpWasClubs: false,
        },
      }),
    );
    expect(tie).toContain('60-60');

    const scored = formatSjavsState(
      makeState({
        handResult: {
          declarerTeam: 0,
          declarerPoints: 95,
          scoringTeam: 0,
          amount: 4,
          vol: false,
          trumpWasClubs: false,
        },
      }),
    );
    expect(scored).toContain('team 0 takes 4 off');
  });

  it('marks a slam and a double victory', () => {
    const vol = formatSjavsState(
      makeState({
        handResult: {
          declarerTeam: 0,
          declarerPoints: 120,
          scoringTeam: 0,
          amount: 12,
          vol: true,
          trumpWasClubs: false,
        },
      }),
    );
    expect(vol).toContain('(all tricks)');

    const dbl = formatSjavsState(makeState({ gameEndFlag: true, winnerTeam: 0, doubleVictory: true }));
    expect(dbl).toContain('you won the rubber');
    expect(dbl).toContain('double victory');
  });
});
