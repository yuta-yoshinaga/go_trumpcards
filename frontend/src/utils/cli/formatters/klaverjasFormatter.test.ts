import { describe, expect, it } from 'vitest';
import { makeKlaverjasState } from '../../../test/stateFactories';
import { formatKlaverjasState } from './klaverjasFormatter';

describe('formatKlaverjasState', () => {
  it('renders the header, round/trick, trump and team scores', () => {
    const out = formatKlaverjasState(makeKlaverjasState({ teamScores: [120, 90] }));
    expect(out).toContain('Klaverjas');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♦');
    expect(out).toContain('score: A=120  B=90');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatKlaverjasState(makeKlaverjasState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatKlaverjasState(
      makeKlaverjasState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the round result and roem during RoundEnd', () => {
    const out = formatKlaverjasState(makeKlaverjasState({ phase: 2, roundCardPoints: [70, 50], roundRoem: [20, 0] }));
    expect(out).toContain('round result: A card pts=70 B card pts=50');
    expect(out).toContain('roem: A=20 B=0');
  });

  it('renders a hint with card indices', () => {
    const out = formatKlaverjasState(
      makeKlaverjasState({
        hint: { cardIndices: [1, 2], reason: 'follow_win' },
        messageCode: 'klaverjas.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatKlaverjasState(makeKlaverjasState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatKlaverjasState(makeKlaverjasState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatKlaverjasState(makeKlaverjasState({ hint, messageCode: 'klaverjas.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatKlaverjasState(makeKlaverjasState({ hint, messageCode: 'klaverjas.playing' }))).not.toContain('HINT');
  });
});
