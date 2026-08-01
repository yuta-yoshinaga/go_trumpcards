import { describe, expect, it } from 'vitest';
import { makeSuecaState } from '../../../test/stateFactories';
import { formatSuecaState } from './suecaFormatter';

describe('formatSuecaState', () => {
  it('renders the header, round/trick, trump and team game points', () => {
    const out = formatSuecaState(makeSuecaState({ teamGamePoints: [2, 1] }));
    expect(out).toContain('Sueca');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♦');
    expect(out).toContain('game points: A=2  B=1');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatSuecaState(makeSuecaState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatSuecaState(
      makeSuecaState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the round result during RoundEnd', () => {
    const out = formatSuecaState(makeSuecaState({ phase: 2, roundCardPoints: [70, 50] }));
    expect(out).toContain('round result: A card pts=70 B card pts=50');
  });

  it('renders a hint with card indices', () => {
    const out = formatSuecaState(
      makeSuecaState({ hint: { cardIndices: [1, 2], reason: 'follow_win' }, messageCode: 'sueca.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatSuecaState(makeSuecaState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatSuecaState(makeSuecaState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  // このゲーム群は hintAvailable がラベルとして埋まっているため hintRequested。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], reason: 'follow' };
    expect(formatSuecaState(makeSuecaState({ hint, messageCode: 'sueca.hintRequested' }))).toContain('HINT');
    expect(formatSuecaState(makeSuecaState({ hint, messageCode: 'sueca.playing' }))).not.toContain('HINT');
  });
});
