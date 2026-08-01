import { describe, expect, it } from 'vitest';
import { makeTysiacState } from '../../../test/stateFactories';
import { formatTysiacState } from './tysiacFormatter';

describe('formatTysiacState', () => {
  it('renders the header, round/trick, trump, bid/contract and per-player scores', () => {
    const out = formatTysiacState(makeTysiacState({ playerScores: [30, 10, 20], currentBid: 110, contract: 110 }));
    expect(out).toContain('Tysiac');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('bid: 110');
    expect(out).toContain('contract: 110');
    expect(out).toContain('P0=30');
    expect(out).toContain('P1=10');
    expect(out).toContain('P2=20');
  });

  it('marks the Declarer role for the human (seat 0)', () => {
    const out = formatTysiacState(makeTysiacState());
    expect(out).toContain('Declarer');
    expect(out).toContain('Player');
  });

  it('renders an unset trump as a dash in the Bid phase', () => {
    const out = formatTysiacState(makeTysiacState({ phase: 0, trumpSuit: 0 }));
    expect(out).toContain('trump: -');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatTysiacState(makeTysiacState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatTysiacState(
      makeTysiacState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders card points and marriage during RoundEnd', () => {
    const out = formatTysiacState(
      makeTysiacState({ phase: 4, roundCardPoints: [55, 35, 30], roundMarriage: [40, 0, 0] }),
    );
    expect(out).toContain('round result: card pts P0=55 P1=35 P2=30');
    expect(out).toContain('marriage: P0=40');
  });

  it('renders a hint with card indices', () => {
    const out = formatTysiacState(
      makeTysiacState({ hint: { cardIndices: [1, 2], reason: 'follow_win' }, messageCode: 'tysiac.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatTysiacState(makeTysiacState({ phase: 5, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatTysiacState(makeTysiacState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatTysiacState(makeTysiacState({ hint, messageCode: 'tysiac.hintRequested' }))).toContain('HINT');
    expect(formatTysiacState(makeTysiacState({ hint, messageCode: 'tysiac.playing' }))).not.toContain('HINT');
  });
});
