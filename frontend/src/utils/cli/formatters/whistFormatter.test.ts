import { describe, expect, it } from 'vitest';
import type { WhistPlayerData, WhistResponse } from '../../../types/card';
import { formatWhistState } from './whistFormatter';

function makePlayer(overrides?: Partial<WhistPlayerData>): WhistPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    roundScore: 2,
    cumulativeScore: 5,
    trickCount: 3,
    team: 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<WhistResponse>): WhistResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1 })],
    phase: 0,
    roundNumber: 2,
    trickNumber: 4,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    dealerIdx: 1,
    teamScores: [3, 1],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [],
    config: { cpuDifficulty: 1, pointLimit: 5 },
    message: '',
    ...overrides,
  };
}

describe('formatWhistState', () => {
  it('renders the header, round, trick and team scores', () => {
    const out = formatWhistState(makeState());
    expect(out).toContain('Whist');
    expect(out).toContain('round: 2');
    expect(out).toContain('trick: 4');
    expect(out).toContain('Team0=3');
    expect(out).toContain('Team1=1');
  });

  // **切り札なしの局面がある。**suit 0 は「切り札なし」で、行ごと出ない。
  it('names the trump only when there is one', () => {
    expect(formatWhistState(makeState())).toContain('trump:');
    expect(formatWhistState(makeState({ trumpSuit: 0 }))).not.toContain('trump:');
  });

  it("renders each player's team, totals and tricks", () => {
    const out = formatWhistState(makeState());
    expect(out).toContain('team=0');
    expect(out).toContain('total=5');
    expect(out).toContain('tricks=3');
  });

  it('renders a play hint', () => {
    const out = formatWhistState(
      makeState({ hint: { cardIndex: 2, reason: 'follow_suit' }, messageCode: 'whist.hintRequested' }),
    );
    expect(out).toContain('HINT: play [2]');
  });

  it('renders the message when present', () => {
    expect(formatWhistState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 2, reason: 'follow_suit' };
    expect(formatWhistState(makeState({ hint, messageCode: 'whist.hintRequested' }))).toContain('HINT');
    expect(formatWhistState(makeState({ hint, messageCode: 'whist.playing' }))).not.toContain('HINT');
  });
});
