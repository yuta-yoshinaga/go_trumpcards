import { describe, expect, it } from 'vitest';
import type { CatchTenResponse } from '../../../types/card';
import { formatCatchTenState } from './catchtenFormatter';

function baseState(overrides: Partial<CatchTenResponse> = {}): CatchTenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'SPADE', value: 11 },
          { design: 'HEART', value: 5 },
        ],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        team: 0,
      },
      { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
      { id: 2, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 0 },
      { id: 3, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 3,
    dealerIdx: 3,
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, pointLimit: 41 },
    message: '',
    ...overrides,
  } as CatchTenResponse;
}

describe('formatCatchTenState', () => {
  it('renders header, round, trump and score', () => {
    const out = formatCatchTenState(baseState());
    expect(out).toContain('Catch the Ten');
    expect(out).toContain('round: 1');
    expect(out).toContain('trump: Heart');
    expect(out).toContain('score: Team0=0 Team1=0');
  });

  it('renders the human hand', () => {
    const out = formatCatchTenState(baseState());
    expect(out).toContain('[0]');
  });

  it('renders the current trick', () => {
    const out = formatCatchTenState(
      baseState({ currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 10 } }] }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a hint when present', () => {
    const out = formatCatchTenState(
      baseState({ hint: { cardIndex: 1, reason: 'follow_suit' }, messageCode: 'catchten.hintRequested' }),
    );
    expect(out).toContain('HINT: play [1]');
  });

  it('renders the game-over line with winning team', () => {
    const out = formatCatchTenState(baseState({ gameEndFlag: true, winnerTeam: 0 }));
    expect(out).toContain('Game Over! Winner: Team 0');
  });

  it('omits trump line when no trump set', () => {
    const out = formatCatchTenState(baseState({ trumpSuit: 0 }));
    expect(out).not.toContain('trump:');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 1, reason: 'follow_suit' };
    expect(formatCatchTenState(baseState({ hint, messageCode: 'catchten.hintRequested' }))).toContain('HINT');
    expect(formatCatchTenState(baseState({ hint, messageCode: 'catchten.playing' }))).not.toContain('HINT');
  });
});
