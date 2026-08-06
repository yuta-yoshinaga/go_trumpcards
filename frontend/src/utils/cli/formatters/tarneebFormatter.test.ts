import { describe, expect, it } from 'vitest';
import type { TarneebPlayerData, TarneebResponse } from '../../../types/card';
import { formatTarneebState } from './tarneebFormatter';

function makePlayer(overrides?: Partial<TarneebPlayerData>): TarneebPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [{ design: 'SPADE', value: 1 }],
    team: 0,
    bid: 7,
    roundScore: 5,
    cumulativeScore: 21,
    trickCount: 3,
    ...overrides,
  };
}

function makeState(overrides?: Partial<TarneebResponse>): TarneebResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1, bid: -1 })],
    teamScores: [21, 14],
    phase: 1,
    roundNumber: 2,
    trickNumber: 4,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    bidWinnerIdx: 0,
    highestBid: 7,
    trumpSuit: 1,
    redealCount: 0,
    dealerIdx: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [],
    config: { cpuDifficulty: 1, pointLimit: 31, minBid: 7 },
    message: '',
    ...overrides,
  };
}

describe('formatTarneebState', () => {
  it('renders the header, round, trick and team scores', () => {
    const out = formatTarneebState(makeState());
    expect(out).toContain('Tarneeb');
    expect(out).toContain('round: 2');
    expect(out).toContain('trick: 4');
    expect(out).toContain('team 0: 21');
    expect(out).toContain('team 1: 14');
  });

  // **宣言前は入札額が無い。**0 のとき - を出す。
  it('shows a dash for the bid before anyone has bid', () => {
    expect(formatTarneebState(makeState())).toContain('bid: 7');
    expect(formatTarneebState(makeState({ highestBid: 0 }))).toContain('bid: -');
  });

  // **入札していない席は - になる。**bid が -1 のとき。
  it('shows a dash for a seat that has not bid', () => {
    const out = formatTarneebState(makeState());
    expect(out).toContain('bid=7');
    expect(out).toContain('bid=-');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatTarneebState(makeState({ hint, messageCode: 'tarneeb.hintRequested' }))).toContain('HINT');
    expect(formatTarneebState(makeState({ hint, messageCode: 'tarneeb.playing' }))).not.toContain('HINT');
  });
});
