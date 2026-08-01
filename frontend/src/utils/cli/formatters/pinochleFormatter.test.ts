import { describe, expect, it } from 'vitest';
import type { PinochlePlayerData, PinochleResponse } from '../../../types/card';
import { formatPinochleState } from './pinochleFormatter';

function makePlayer(overrides?: Partial<PinochlePlayerData>): PinochlePlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    team: 0,
    trickCount: 2,
    bid: 0,
    hasPassed: false,
    meldScore: 40,
    trickPoints: 12,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PinochleResponse>): PinochleResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1 })],
    phase: 0,
    roundNumber: 2,
    trickNumber: 5,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    trumpSuit: 1,
    highestBid: 250,
    highestBidder: 0,
    currentTrick: [],
    teamScores: [120, 80],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    playerMelds: [[], []],
    config: { cpuDifficulty: 1, pointLimit: 1500 },
    message: '',
    ...overrides,
  };
}

describe('formatPinochleState', () => {
  it('renders the header, round, trick and team scores', () => {
    const out = formatPinochleState(makeState());
    expect(out).toContain('Pinochle');
    expect(out).toContain('round: 2');
    expect(out).toContain('trick: 5');
    expect(out).toContain('Team0=120');
    expect(out).toContain('Team1=80');
  });

  // **切り札も宣言額も競りが終わるまで無い。**どちらも行ごと出ない。
  it('omits the trump and the bid until the auction settles them', () => {
    const out = formatPinochleState(makeState({ trumpSuit: 0, highestBid: 0 }));
    expect(out).not.toContain('trump:');
    expect(out).not.toContain('bid:');
    expect(formatPinochleState(makeState())).toContain('bid: 250');
  });

  it("renders each player's meld and trick points", () => {
    const out = formatPinochleState(makeState());
    expect(out).toContain('meld=40');
    expect(out).toContain('tricks=12');
  });

  // **パスした席にだけ印が付く。**
  it('marks a seat that has passed', () => {
    expect(formatPinochleState(makeState())).not.toContain('[Passed]');
    expect(formatPinochleState(makeState({ players: [makePlayer({ hasPassed: true })] }))).toContain('[Passed]');
  });

  it('announces the winner once the game ends', () => {
    const out = formatPinochleState(makeState({ gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatPinochleState(makeState({ hint, messageCode: 'pinochle.hintRequested' }))).toContain('HINT');
    expect(formatPinochleState(makeState({ hint, messageCode: 'pinochle.playing' }))).not.toContain('HINT');
  });
});
