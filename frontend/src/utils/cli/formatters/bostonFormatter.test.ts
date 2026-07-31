import { describe, expect, it } from 'vitest';
import type { BostonBidOption, BostonPlayer, BostonResponse, CardDesign } from '../../../types/card';
import { formatBostonState } from './bostonFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function option(level: number, name: string, kind: number, tricks: number): BostonBidOption {
  return {
    level,
    name,
    kind,
    tricks,
    needsTrump: kind === 1,
    exposed: false,
    canCallPartner: kind === 1,
    payout: level,
  };
}

/** The ladder in rank order, with the miseres between the trick bids. */
const LADDER: BostonBidOption[] = [
  option(1, 'five', 1, 5),
  option(2, 'six', 1, 6),
  option(3, 'littleMisere', 2, 0),
  option(4, 'seven', 1, 7),
  option(5, 'piccolissimo', 3, 1),
];

function seat(id: number, isHuman: boolean, overrides?: Partial<BostonPlayer>): BostonPlayer {
  return {
    id,
    isHuman,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 2)] : [],
    tricksWon: 0,
    chips: 0,
    isDealer: id === 3,
    isDeclarer: id === 1,
    isPartner: false,
    isDeclarerSide: id === 1,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BostonResponse>): BostonResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 1, level: 4, name: 'seven', suit: 3 },
    bidOptions: LADDER,
    declarerIdx: 1,
    partnerIdx: -1,
    trumpSuit: 3,
    exposed: false,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 0,
    declarerTricks: 4,
    bidMade: false,
    handSize: 13,
    targetHands: 8,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatBostonState', () => {
  it('shows the hand, contract and roles', () => {
    const out = formatBostonState(makeState());
    expect(out).toContain('hand: 1/8');
    expect(out).toContain('phase: Play');
    expect(out).toContain('contract: seven');
    expect(out).toContain('trump: H');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[declarer]');
  });

  it('hides opponent hands but shows their size', () => {
    expect(formatBostonState(makeState())).toContain('hidden (2)');
  });

  // **出せる札を出さないと操作できない。**追随が強制。
  it('lists the playable indexes on the human turn', () => {
    const out = formatBostonState(makeState());
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('your turn');
  });

  // **序列を見せないと競りの判断ができない。**ミゼールが間に挟まるため。
  it('shows the remaining ladder while bidding', () => {
    const out = formatBostonState(makeState({ phase: 0, highBid: null }));
    expect(out).toContain('ladder:');
    expect(out).toContain('LADDER STEP');
    // 並びは 2:six < 3:littleMisere < 4:seven。
    expect(out.indexOf('2:six')).toBeLessThan(out.indexOf('3:littleMisere'));
    expect(out.indexOf('3:littleMisere')).toBeLessThan(out.indexOf('4:seven'));
  });

  // 立っている宣言より上しか出さない。
  it('drops ladder steps that can no longer be bid', () => {
    const out = formatBostonState(
      makeState({ phase: 0, highBid: { player: 1, level: 3, name: 'littleMisere', suit: 0 } }),
    );
    expect(out).not.toContain('2:six');
    expect(out).toContain('4:seven');
  });

  it('prompts for the partner decision', () => {
    const out = formatBostonState(makeState({ phase: 1, declarerIdx: 0 }));
    expect(out).toContain('cp -1');
    expect(out).toContain('play alone against three');
  });

  // 達成と失敗は取ったトリック数つきで区別する。
  it('tells a failed contract apart', () => {
    expect(formatBostonState(makeState({ phase: 3, bidMade: false }))).toContain('FAILED with 4 tricks');
    expect(formatBostonState(makeState({ phase: 3, bidMade: true }))).toContain('made with 4 tricks');
  });

  it('marks a called partner', () => {
    const out = formatBostonState(
      makeState({
        partnerIdx: 3,
        players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false, { isPartner: true })],
      }),
    );
    expect(out).toContain('[partner]');
  });

  it('shows the trick and announces the winner', () => {
    expect(formatBostonState(makeState({ trick: [card('SPADE', 5)] }))).toContain('trick:');
    const end = formatBostonState(makeState({ phase: 4, gameEndFlag: true, winnerIdx: 2, message: 'done' }));
    expect(end).toContain('Game Over!');
    expect(end).toContain('done');
  });
});
