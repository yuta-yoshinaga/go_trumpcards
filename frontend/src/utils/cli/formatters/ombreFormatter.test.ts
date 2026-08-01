import { describe, expect, it } from 'vitest';
import { makeOmbreState } from '../../../test/stateFactories';
import { formatOmbreState } from './ombreFormatter';

describe('formatOmbreState', () => {
  it('renders the header, round/trick, bid, trump and per-player scores', () => {
    const out = formatOmbreState(makeOmbreState({ playerScores: [3, 1, 2], winningBid: 1, trumpSuit: 1 }));
    expect(out).toContain('Ombre');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('bid: entrar');
    expect(out).toContain('trump: spade');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('marks the Ombre role for the human (seat 0)', () => {
    const out = formatOmbreState(makeOmbreState());
    expect(out).toContain('Ombre');
    expect(out).toContain('Coalition');
  });

  it('renders a solo bid label', () => {
    const out = formatOmbreState(makeOmbreState({ winningBid: 2 }));
    expect(out).toContain('bid: solo');
  });

  it('renders trump as dash when unset', () => {
    const out = formatOmbreState(makeOmbreState({ trumpSuit: -1 }));
    expect(out).toContain('trump: -');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatOmbreState(makeOmbreState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatOmbreState(
      makeOmbreState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the outcome during RoundEnd', () => {
    const out = formatOmbreState(makeOmbreState({ phase: 3, outcome: 3 }));
    expect(out).toContain('round result: Codille');
  });

  it('renders a hint with card indices', () => {
    const out = formatOmbreState(
      makeOmbreState({ hint: { cardIndices: [1, 2], reason: 'follow_win' }, messageCode: 'ombre.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatOmbreState(makeOmbreState({ phase: 4, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatOmbreState(makeOmbreState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatOmbreState(makeOmbreState({ hint, messageCode: 'ombre.hintRequested' }))).toContain('HINT');
    expect(formatOmbreState(makeOmbreState({ hint, messageCode: 'ombre.playing' }))).not.toContain('HINT');
  });
});
