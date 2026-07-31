import { describe, expect, it } from 'vitest';
import type { CardDesign, SixBidSoloHandResult, SixBidSoloPlayer, SixBidSoloResponse } from '../../../types/card';
import { formatSixBidSoloState } from './sixbidsoloFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<SixBidSoloPlayer>): SixBidSoloPlayer {
  return {
    id,
    isHuman,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 10)] : [],
    points: 0,
    tricksWon: 0,
    score: 0,
    isDealer: id === 2,
    isDeclarer: id === 1,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<SixBidSoloHandResult>): SixBidSoloHandResult {
  return {
    kind: 1,
    declarer: 1,
    declarerPoints: 65,
    widowPoints: 25,
    target: 61,
    made: true,
    value: 10,
    deltas: [-10, 20, -10],
    ...overrides,
  };
}

function makeState(overrides?: Partial<SixBidSoloResponse>): SixBidSoloResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false)],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 2,
    bids: [],
    highBid: { player: 1, kind: 4 },
    declarerIdx: 1,
    trumpSuit: 1,
    declared: true,
    calledCard: null,
    spreadOpen: false,
    widow: [],
    widowSize: 3,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 0,
    lastResult: null,
    bidTargets: [0, 61, 61, 0, 80, 0, 120],
    totalPoints: 120,
    baseTarget: 60,
    handSize: 11,
    targetHands: 6,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, targetHands: 6 },
    ...overrides,
  };
}

describe('formatSixBidSoloState', () => {
  it('shows the header, the contract and its target', () => {
    const out = formatSixBidSoloState(makeState());
    expect(out).toContain('Six-Bid Solo');
    expect(out).toContain('hand: 1/6');
    expect(out).toContain('phase: Play');
    expect(out).toContain('card points in play: 120');
    // **♠ のギャランティーは 80 点。**
    expect(out).toContain('contract: Guarantee Solo, trump S, target 80');
  });

  it('says the trump is not named yet', () => {
    expect(formatSixBidSoloState(makeState({ declared: false }))).toContain('not yet named');
  });

  // **ウィドウは精算まで伏せたまま。**
  it('keeps the widow face down until it is revealed', () => {
    expect(formatSixBidSoloState(makeState())).toContain('widow: 3 face down');
    const revealed = formatSixBidSoloState(
      makeState({ widow: [card('DIAMOND', 1), card('DIAMOND', 10), card('DIAMOND', 6)] }),
    );
    expect(revealed).not.toContain('3 face down');
  });

  it('hides every other hand', () => {
    const out = formatSixBidSoloState(makeState());
    expect(out).toContain('[0]');
    expect(out).toContain('hidden (2)');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[declarer]');
    expect(out.match(/hidden \(2\)/g)).toHaveLength(2);
  });

  // **通常ビッドは 60 を超えることが要る。**
  it('lists the ladder while bidding and states the real floor', () => {
    const out = formatSixBidSoloState(makeState({ phase: 0, bidPlayerIdx: 0 }));
    expect(out).toContain('1:Solo');
    expect(out).toContain('6:Call Solo');
    expect(out).toContain('must EXCEED 60, so 61 points');
  });

  it('explains the declaration step, including the call solo form', () => {
    const out = formatSixBidSoloState(makeState({ phase: 1, declarerIdx: 0 }));
    expect(out).toContain('d <1-4>');
    expect(out).toContain('call solo continues');
  });

  it('lists the playable indexes on the human turn', () => {
    expect(formatSixBidSoloState(makeState())).toContain('playable: 0 1');
  });

  // **ウィドウ加算はミゼール系では 0。**
  it('reports the settlement with the widow credit', () => {
    const made = formatSixBidSoloState(makeState({ phase: 3, lastResult: result() }));
    expect(made).toContain('Solo made: 65 points, needed 61');
    expect(made).toContain('widow credited: 25');
    expect(made).toContain('settlement: 10 per opponent');

    const set = formatSixBidSoloState(
      makeState({
        phase: 3,
        lastResult: result({ kind: 3, made: false, declarerPoints: 8, widowPoints: 0, target: 0 }),
      }),
    );
    expect(set).toContain('Misere SET');
    expect(set).toContain('widow credited: 0');
    expect(set).toContain('excluded at either misere');
  });

  it('shows the called card, the trick, the message and the winner', () => {
    const out = formatSixBidSoloState(
      makeState({
        calledCard: card('HEART', 1),
        trick: [card('SPADE', 13)],
        message: 'boom',
        gameEndFlag: true,
        winnerIdx: 0,
      }),
    );
    expect(out).toContain('called card:');
    expect(out).toContain('had to exchange it');
    expect(out).toContain('trick:');
    expect(out).toContain('boom');
    expect(out).toContain('Winning seat: 0');
  });

  it('survives an unknown phase and a missing contract', () => {
    expect(formatSixBidSoloState(makeState({ phase: 99 }))).toContain('phase: 99');
    expect(formatSixBidSoloState(makeState({ highBid: null }))).not.toContain('contract:');
  });
});
