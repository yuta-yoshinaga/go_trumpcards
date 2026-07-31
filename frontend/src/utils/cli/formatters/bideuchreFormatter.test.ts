import { describe, expect, it } from 'vitest';
import type { BidEuchreHandResult, BidEuchrePlayer, BidEuchreResponse, CardDesign } from '../../../types/card';
import { formatBidEuchreState } from './bideuchreFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<BidEuchrePlayer>): BidEuchrePlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 11)] : [],
    tricksWon: 0,
    isDealer: id === 3,
    isDeclarer: id === 1,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<BidEuchreHandResult>): BidEuchreHandResult {
  return {
    points: [4, 2],
    tricks: [4, 2],
    made: true,
    bid: 3,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BidEuchreResponse>): BidEuchreResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 1, value: 4 },
    declarerIdx: 1,
    trump: 0,
    trumpSuit: 1,
    trumpChosen: true,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [0, 0],
    scores: [0, 0],
    lastResult: null,
    gameTarget: 32,
    minBid: 3,
    maxBid: 6,
    handSize: 6,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  };
}

describe('formatBidEuchreState', () => {
  it('shows the header, scores and the contract', () => {
    const out = formatBidEuchreState(makeState({ scores: [12, 20] }));
    expect(out).toContain('Bid Euchre');
    expect(out).toContain('phase: Play');
    expect(out).toContain('team0: 12 | team1: 20');
    expect(out).toContain('game is 32');
    expect(out).toContain('contract: 4 tricks, trump S');
  });

  // **落札直後は切札が未定。**
  it('says the trump is not named yet', () => {
    expect(formatBidEuchreState(makeState({ trumpChosen: false }))).toContain('not yet named');
  });

  // **キティが無く、伏せられているのは他家の手札だけ。**
  it('hides every other hand', () => {
    const out = formatBidEuchreState(makeState());
    expect(out).toContain('[0]');
    expect(out).toContain('hidden (2)');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[declarer]');
    expect(out.match(/hidden \(2\)/g)).toHaveLength(3);
  });

  // **親だけは同額で奪える。**
  it('says the dealer alone may equal the standing bid while bidding', () => {
    const out = formatBidEuchreState(makeState({ phase: 0, bidPlayerIdx: 0 }));
    expect(out).toContain('b <3-6>');
    expect(out).toContain('DEALER alone may EQUAL');
  });

  // **ノートランプが 2 種類あり、ローは序列が逆転する。**
  it('lists both no-trump forms when naming trump', () => {
    const out = formatBidEuchreState(makeState({ phase: 1, declarerIdx: 0, trumpChosen: false }));
    expect(out).toContain('4:NT-high');
    expect(out).toContain('5:NT-low');
    expect(out).toContain('ranking REVERSES');
  });

  it('lists the playable indexes on the human turn', () => {
    const out = formatBidEuchreState(makeState());
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('left bower counts as a trump');
  });

  // **未達で失うのは宣言額。**守備側は取ったトリックを得点する。
  it('explains a set', () => {
    const out = formatBidEuchreState(
      makeState({ phase: 3, lastResult: result({ made: false, bid: 5, points: [-5, 4], tricks: [2, 4] }) }),
    );
    expect(out).toContain('contract FAILED: bid 5');
    expect(out).toContain('points: team0 -5 / team1 4');
    expect(out).toContain('tricks: team0 2 / team1 4');
    expect(out).toContain('costs the BID, not the tricks taken');
  });

  it('omits the set note when the contract was made', () => {
    const out = formatBidEuchreState(makeState({ phase: 3, lastResult: result() }));
    expect(out).toContain('contract made: bid 3');
    expect(out).not.toContain('costs the BID');
  });

  it('shows the trick, the message and the winner', () => {
    const out = formatBidEuchreState(
      makeState({
        trick: [card('SPADE', 13)],
        message: 'boom',
        gameEndFlag: true,
        winnerTeam: 0,
      }),
    );
    expect(out).toContain('trick:');
    expect(out).toContain('boom');
    expect(out).toContain('Winning team: 0');
  });

  it('survives an unknown phase', () => {
    expect(formatBidEuchreState(makeState({ phase: 99 }))).toContain('phase: 99');
  });
});
