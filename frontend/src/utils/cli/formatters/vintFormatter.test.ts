import { describe, expect, it } from 'vitest';
import type { CardDesign, VintHandResult, VintPlayer, VintResponse } from '../../../types/card';
import { formatVintState } from './vintFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<VintPlayer>): VintPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 2)] : [],
    tricksWon: 0,
    isDealer: id === 3,
    isDeclarer: id === 1,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<VintHandResult>): VintHandResult {
  return {
    trickPoints: [210, 180],
    honourPoints: [600, 0],
    acePoints: [1200, 0],
    penalty: [0, 0],
    made: true,
    declarerTricks: 9,
    trickValue: 30,
    ...overrides,
  };
}

function makeState(overrides?: Partial<VintResponse>): VintResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 1,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: { player: 1, level: 3, denom: 3, trickValue: 30 },
    declarerIdx: 1,
    trumpSuit: 3,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [6, 7],
    below: [100, 200],
    above: [50, 100],
    gamesWon: [0, 1],
    lastResult: null,
    trickValues: [4, 6, 8, 10, 12],
    gameTarget: 500,
    minLevel: 1,
    maxLevel: 7,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  };
}

describe('formatVintState', () => {
  it('shows the hand, both score lines and the contract', () => {
    const out = formatVintState(makeState());
    expect(out).toContain('hand: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('team0: below 100 above 50 games 0');
    expect(out).toContain('team1: below 200 above 100 games 1');
    expect(out).toContain('contract: 3 H');
    expect(out).toContain('trick value 30');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[declarer]');
  });

  // **ダミーが無いので、味方の手札も伏せられる。**
  it('hides every hand but your own', () => {
    const out = formatVintState(makeState());
    expect((out.match(/hidden \(2\)/g) ?? []).length).toBe(3);
  });

  it('marks each seat with its team', () => {
    const out = formatVintState(makeState());
    expect(out).toContain('(T0)');
    expect(out).toContain('(T1)');
  });

  // **出せる札を出さないと操作できない。**追随が強制。
  it('lists the playable indexes on the human turn', () => {
    const out = formatVintState(makeState());
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('your turn');
  });

  // **♠ が最弱で NT が最強。**ブリッジと逆なので必ず見せる。
  it('shows the reversed ranking while bidding', () => {
    const out = formatVintState(makeState({ phase: 0, highBid: null }));
    expect(out).toContain('ranking:');
    expect(out).toContain('0:S(4)');
    expect(out).toContain('4:NT(12)');
    expect(out).toContain('spades are LOWEST');
    expect(out.indexOf('0:S')).toBeLessThan(out.indexOf('4:NT'));
  });

  // **両チームが線下に得点する。**issue の「宣言側だけ」は誤り。
  it('reports both sides below-the-line points', () => {
    const out = formatVintState(makeState({ phase: 2, lastResult: result() }));
    expect(out).toContain('contract made with 9 tricks');
    expect(out).toContain('below: team0 +210 / team1 +180');
    expect(out).toContain('BOTH sides score their tricks');
    expect(out).toContain('honours: team0 +600');
    expect(out).toContain('aces: team0 +1200');
  });

  it('reports the penalty only when one applies', () => {
    const clean = formatVintState(makeState({ phase: 2, lastResult: result() }));
    expect(clean).not.toContain('penalty:');
    const set = formatVintState(
      makeState({ phase: 2, lastResult: result({ made: false, declarerTricks: 6, penalty: [0, 1500] }) }),
    );
    expect(set).toContain('contract FAILED with 6 tricks');
    expect(set).toContain('penalty: team0 +0 / team1 +1500');
  });

  it('shows the trick and announces the rubber', () => {
    expect(formatVintState(makeState({ trick: [card('SPADE', 5)] }))).toContain('trick:');
    const end = formatVintState(makeState({ phase: 3, gameEndFlag: true, winnerTeam: 1, message: 'done' }));
    expect(end).toContain('Rubber over! Winning team: 1');
    expect(end).toContain('done');
  });
});
