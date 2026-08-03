import { describe, expect, it } from 'vitest';
import type { CinchPlayer, CinchResponse } from '../../../types/card';
import { formatCinchState } from './cinchFormatter';

function makePlayer(overrides?: Partial<CinchPlayer>): CinchPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 6,
    cards: [{ design: 'SPADE', value: 1 }],
    trickCount: 2,
    bid: 3,
    totalScore: 14,
    ...overrides,
  };
}

function makeState(overrides?: Partial<CinchResponse>): CinchResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], totalScore: 9 })],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    totalTricks: 6,
    dealerIdx: 1,
    currentTurn: 0,
    bidPlayerIdx: 0,
    currentBid: 3,
    bidWinnerIdx: 0,
    trumpSuit: 1,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    playableIndices: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinners: [],
    isHumanTurn: true,
    config: { cpuDifficulty: 1, pointLimit: 21 },
    message: '',
    ...overrides,
  };
}

describe('formatCinchState', () => {
  it('renders the header, the high bid and the trump', () => {
    const out = formatCinchState(makeState());
    expect(out).toContain('Cinch');
    expect(out).toContain('high bid: 3');
    expect(out).toContain('trump:');
  });

  it('renders every player score on the summary line', () => {
    const out = formatCinchState(makeState());
    expect(out).toContain('P0=14');
    expect(out).toContain('P1=9');
  });

  // **入札を取った席にだけ印が付く。**取っていなければ誰にも付かない。
  it('marks the bid winner', () => {
    expect(formatCinchState(makeState())).toContain('(Bidder)');
    expect(formatCinchState(makeState({ bidWinnerIdx: -1 }))).not.toContain('(Bidder)');
  });

  it("renders each player's cards, tricks and bid", () => {
    const out = formatCinchState(makeState());
    expect(out).toContain('cards=6');
    expect(out).toContain('tricks=2');
    expect(out).toContain('bid=3');
  });

  it('renders the message when present', () => {
    expect(formatCinchState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'follow_suit' };
    expect(formatCinchState(makeState({ hint, messageCode: 'cinch.hintRequested' }))).toContain('HINT');
    expect(formatCinchState(makeState({ hint, messageCode: 'cinch.playing' }))).not.toContain('HINT');
  });
});
