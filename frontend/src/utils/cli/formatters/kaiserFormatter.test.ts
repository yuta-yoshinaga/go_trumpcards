import { describe, expect, it } from 'vitest';
import type { CardDesign, KaiserPlayer, KaiserResponse } from '../../../types/card';
import { formatKaiserState } from './kaiserFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<KaiserPlayer>): KaiserPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 2,
    cards: isHuman ? [card('HEART', 5), card('SPADE', 1)] : [],
    isDealer: id === 0,
    isDeclarer: id === 1,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KaiserResponse>): KaiserResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 1,
    dealerIdx: 0,
    bids: [],
    highBid: { player: 1, value: 8, contract: 0 },
    declarerIdx: 1,
    trumpSuit: 3,
    contract: 0,
    kittySize: 0,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 0,
    validPlays: [0, 1],
    teamHandPoints: [3, 2],
    teamScores: [10, 20],
    heartFiveBy: -1,
    spadeThreeBy: -1,
    bidMade: false,
    targetScore: 52,
    minBid: 7,
    maxBid: 12,
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 0, allowNoTrump: true },
    message: '',
    ...overrides,
  };
}

describe('formatKaiserState', () => {
  it('shows the hand, target, contract and both team scores', () => {
    const out = formatKaiserState(makeState());
    expect(out).toContain('hand: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('target: 52');
    expect(out).toContain('team0: 10');
    expect(out).toContain('contract: 8 with trump');
    expect(out).toContain('trump: H');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[declarer]');
  });

  // **パートナーが判らないと戦えない。**
  it('marks each seat with its team', () => {
    const out = formatKaiserState(makeState());
    expect(out).toContain('(T0)');
    expect(out).toContain('(T1)');
  });

  it('hides opponent hands but shows their size', () => {
    expect(formatKaiserState(makeState())).toContain('hidden (2)');
  });

  // **出せる札を出さないと操作できない。**追随が強制。
  it('lists the playable indexes on the human turn', () => {
    const out = formatKaiserState(makeState());
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('your turn');
  });

  // ♥5 と ♠3 の行方はトリック8点と同じ重み。
  it('reports where the scoring cards went', () => {
    const out = formatKaiserState(makeState({ heartFiveBy: 0, spadeThreeBy: 2 }));
    expect(out).toContain('H5 (+5) taken by');
    expect(out).toContain('S3 (-3) taken by');
  });

  it('prompts for the bid in points and for the kitty', () => {
    expect(formatKaiserState(makeState({ phase: 0, bidPlayerIdx: 0 }))).toContain('POINTS');
    const needsTrump = formatKaiserState(makeState({ phase: 1, declarerIdx: 0, trumpSuit: 0 }));
    expect(needsTrump).toContain('name trump');
    const needsDiscard = formatKaiserState(makeState({ phase: 1, declarerIdx: 0 }));
    expect(needsDiscard).toContain('H5 and S3 may not go');
  });

  it('shows the kitty only while it is face down', () => {
    expect(formatKaiserState(makeState({ phase: 0, kittySize: 2 }))).toContain('kitty: 2');
    expect(formatKaiserState(makeState())).not.toContain('kitty:');
  });

  // ベートは達成と字面で区別する。
  it('tells a set hand apart', () => {
    expect(formatKaiserState(makeState({ phase: 3, bidMade: false }))).toContain('SET');
    expect(formatKaiserState(makeState({ phase: 3, bidMade: true }))).toContain('made it');
  });

  it('shows the trick and announces the winning team', () => {
    expect(formatKaiserState(makeState({ trick: [card('HEART', 1)] }))).toContain('trick:');
    const end = formatKaiserState(makeState({ phase: 4, gameEndFlag: true, winnerTeam: 1, message: 'done' }));
    expect(end).toContain('Winning team: 1');
    expect(end).toContain('done');
  });
});
