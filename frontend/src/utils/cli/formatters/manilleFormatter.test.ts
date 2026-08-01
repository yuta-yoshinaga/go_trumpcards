import { describe, expect, it } from 'vitest';
import { makeManilleState } from '../../../test/stateFactories';
import { formatManilleState } from './manilleFormatter';

describe('formatManilleState', () => {
  it('renders the header, round/trick, trump and team scores', () => {
    const out = formatManilleState(makeManilleState({ teamScores: [80, 60] }));
    expect(out).toContain('Manille');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♦');
    expect(out).toContain('score: A=80  B=60');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatManilleState(makeManilleState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatManilleState(
      makeManilleState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the round result during RoundEnd', () => {
    const out = formatManilleState(makeManilleState({ phase: 2, roundCardPoints: [35, 25] }));
    expect(out).toContain('round result: A card pts=35 B card pts=25');
  });

  it('renders a hint with card indices', () => {
    const out = formatManilleState(
      makeManilleState({ hint: { cardIndices: [1, 2], reason: 'follow_win' }, messageCode: 'manille.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatManilleState(makeManilleState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatManilleState(makeManilleState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  // このゲーム群は hintAvailable がラベルとして埋まっているため hintRequested。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], reason: 'follow_suit' };
    expect(formatManilleState(makeManilleState({ hint, messageCode: 'manille.hintRequested' }))).toContain('HINT');
    expect(formatManilleState(makeManilleState({ hint, messageCode: 'manille.playing' }))).not.toContain('HINT');
  });
});
