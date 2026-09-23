import { describe, expect, it } from 'vitest';
import { makeMarjapussiState } from '../../../test/stateFactories';
import { formatMarjapussiState } from './marjapussiFormatter';

describe('formatMarjapussiState', () => {
  it('renders the header, round/trick, trump, team scores and per-player scores', () => {
    const out = formatMarjapussiState(
      makeMarjapussiState({
        teamScores: [30, 20],
        playerScores: [15, 10, 15, 10],
        trumpSuit: 3,
      }),
    );
    expect(out).toContain('Marjapussi');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('team scores: Team0=30  Team1=20');
    expect(out).toContain('P0=15');
    expect(out).toContain('P1=10');
  });

  it('marks the Team role for players', () => {
    const out = formatMarjapussiState(makeMarjapussiState());
    expect(out).toContain('Team 0');
    expect(out).toContain('Team 1');
  });

  it('renders an unset trump as a dash when trumpSuit is 0', () => {
    const out = formatMarjapussiState(makeMarjapussiState({ phase: 0, trumpSuit: 0 }));
    expect(out).toContain('trump: -');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatMarjapussiState(makeMarjapussiState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatMarjapussiState(
      makeMarjapussiState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders card points, marriage, and pussi winner during RoundEnd', () => {
    const out = formatMarjapussiState(
      makeMarjapussiState({
        phase: 2,
        roundCardPoints: [55, 35],
        roundMarriage: [40, 0],
        pussiWinnerTeam: 0,
      }),
    );
    expect(out).toContain('round result: card pts Team0=55 Team1=35');
    expect(out).toContain('marriage: Team0=40');
    expect(out).toContain('pussi won by: Team 0');
  });

  it('renders a hint with card indices', () => {
    const out = formatMarjapussiState(
      makeMarjapussiState({
        hint: { cardIndices: [1, 2], reason: 'follow_win' },
        messageCode: 'marjapussi.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatMarjapussiState(makeMarjapussiState({ phase: 3, gameEndFlag: true, winnerTeam: 0 }));
    expect(out).toContain('Game Over! Winner: Team 0');
  });

  it('renders an explicit message when present', () => {
    const out = formatMarjapussiState(makeMarjapussiState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatMarjapussiState(makeMarjapussiState({ hint, messageCode: 'marjapussi.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatMarjapussiState(makeMarjapussiState({ hint, messageCode: 'marjapussi.playing' }))).not.toContain(
      'HINT',
    );
  });
});
