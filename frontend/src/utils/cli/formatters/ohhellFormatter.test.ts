import { describe, expect, it } from 'vitest';
import type { OhHellPlayerData, OhHellResponse } from '../../../types/card';
import { formatOhhellState } from './ohhellFormatter';

function makePlayer(overrides?: Partial<OhHellPlayerData>): OhHellPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    bid: 1,
    roundScore: 0,
    cumulativeScore: 10,
    trickCount: 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<OhHellResponse>): OhHellResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [] })],
    phase: 0,
    roundNumber: 2,
    totalRounds: 7,
    handSize: 3,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    currentTrick: [],
    trumpCard: { design: 'HEART', value: 12 },
    trumpSuit: 3,
    restrictedBid: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, maxHandSize: 7, scoringVariant: 0, roundDirection: 0 },
    message: '',
    ...overrides,
  };
}

describe('formatOhhellState', () => {
  it('renders the header, round out of total, trick and hand size', () => {
    const out = formatOhhellState(makeState());
    expect(out).toContain('Oh Hell');
    expect(out).toContain('round: 2/7');
    expect(out).toContain('trick: 1');
    expect(out).toContain('hand size: 3');
  });

  it('names the trump card only when there is one', () => {
    expect(formatOhhellState(makeState())).toContain('trump:');
    expect(formatOhhellState(makeState({ trumpCard: null }))).not.toContain('trump:');
  });

  it("renders each player's score, bid and tricks", () => {
    const out = formatOhhellState(makeState());
    expect(out).toContain('total=10');
    expect(out).toContain('bid=1');
    expect(out).toContain('tricks=0');
  });

  // **入札とプレイでヒント行が変わる。**
  it('renders a bid hint and a play hint differently', () => {
    const bid = formatOhhellState(
      makeState({ hint: { bid: 2, reason: 'strategic_bid' }, messageCode: 'ohhell.hintRequested' }),
    );
    expect(bid).toContain('HINT: bid 2');

    const play = formatOhhellState(
      makeState({ hint: { cardIndex: 1, reason: 'follow_suit' }, messageCode: 'ohhell.hintRequested' }),
    );
    expect(play).toContain('HINT: play [1]');
  });

  it('renders the message when present', () => {
    expect(formatOhhellState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 1, reason: 'follow_suit' };
    expect(formatOhhellState(makeState({ hint, messageCode: 'ohhell.hintRequested' }))).toContain('HINT');
    expect(formatOhhellState(makeState({ hint, messageCode: 'ohhell.playing' }))).not.toContain('HINT');
  });
});
