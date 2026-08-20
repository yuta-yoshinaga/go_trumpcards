import { describe, expect, it } from 'vitest';
import type { HeartsResponse } from '../../../types/card';
import { formatHeartsState } from './heartsFormatter';

function makeState(overrides?: Partial<HeartsResponse>): HeartsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [{ design: 'SPADE', value: 5 }],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
        tookOmnibusJD: false,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
        tookOmnibusJD: false,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
        tookOmnibusJD: false,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
        tookOmnibusJD: false,
      },
    ],
    phase: 1,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    heartsBroken: false,
    passDirection: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 100, omnibusJD: false },
    ...overrides,
  };
}

describe('formatHeartsState', () => {
  it('formats basic state', () => {
    const output = formatHeartsState(makeState());
    expect(output).toContain('Hearts');
    expect(output).toContain('round: 1');
    expect(output).toContain('hearts broken: no');
  });

  it('shows hearts broken', () => {
    const output = formatHeartsState(makeState({ heartsBroken: true }));
    expect(output).toContain('hearts broken: yes');
  });

  it('shows pass phase', () => {
    const output = formatHeartsState(makeState({ phase: 0 }));
    expect(output).toContain('Pass phase: Left');
  });

  it('shows current trick', () => {
    const output = formatHeartsState(
      makeState({ currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 3 } }] }),
    );
    expect(output).toContain('trick:');
  });

  it('shows game over with winner', () => {
    const output = formatHeartsState(makeState({ gameEndFlag: true, winnerIdx: 0 }));
    expect(output).toContain('Game Over');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'follow_suit' };
    expect(formatHeartsState(makeState({ hint, messageCode: 'hearts.hintRequested' }))).toContain('HINT');
    expect(formatHeartsState(makeState({ hint, messageCode: 'hearts.playing' }))).not.toContain('HINT');
  });
});
