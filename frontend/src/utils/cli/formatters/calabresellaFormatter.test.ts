import { describe, expect, it } from 'vitest';
import { makeCalabresellaState } from '../../../test/stateFactories';
import { formatCalabresellaState } from './calabresellaFormatter';

describe('formatCalabresellaState', () => {
  it('renders the header, round/trick, bid and per-player scores', () => {
    const out = formatCalabresellaState(makeCalabresellaState({ playerScores: [3, 1, 2], winningBid: 1 }));
    expect(out).toContain('Calabresella');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('bid: chiamo');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('marks the Soloist role for the human (seat 0)', () => {
    const out = formatCalabresellaState(makeCalabresellaState());
    expect(out).toContain('Soloist');
    expect(out).toContain('Coalition');
  });

  it('renders a solo bid label', () => {
    const out = formatCalabresellaState(makeCalabresellaState({ winningBid: 2 }));
    expect(out).toContain('bid: solo');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatCalabresellaState(makeCalabresellaState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatCalabresellaState(
      makeCalabresellaState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders thirds during RoundEnd', () => {
    const out = formatCalabresellaState(makeCalabresellaState({ phase: 4, roundThirds: [20, 8, 5] }));
    expect(out).toContain('round result: thirds P0=20 P1=8 P2=5');
  });

  it('renders a hint with card indices', () => {
    const out = formatCalabresellaState(
      makeCalabresellaState({
        hint: { cardIndices: [1, 2], reason: 'follow_win' },
        messageCode: 'calabresella.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatCalabresellaState(makeCalabresellaState({ phase: 5, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatCalabresellaState(makeCalabresellaState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(
      formatCalabresellaState(makeCalabresellaState({ hint, messageCode: 'calabresella.hintRequested' })),
    ).toContain('HINT');
    expect(formatCalabresellaState(makeCalabresellaState({ hint, messageCode: 'calabresella.playing' }))).not.toContain(
      'HINT',
    );
  });
});
