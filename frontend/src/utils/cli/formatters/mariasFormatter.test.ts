import { describe, expect, it } from 'vitest';
import { makeMariasState } from '../../../test/stateFactories';
import { formatMariasState } from './mariasFormatter';

describe('formatMariasState', () => {
  it('renders the header, round/trick, trump and per-player scores', () => {
    const out = formatMariasState(makeMariasState({ playerScores: [3, 1, 2] }));
    expect(out).toContain('Marias');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('marks the Soloist role for the human (seat 0)', () => {
    const out = formatMariasState(makeMariasState());
    expect(out).toContain('Soloist');
    expect(out).toContain('Defender');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatMariasState(makeMariasState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatMariasState(
      makeMariasState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders card points and marriage during RoundEnd', () => {
    const out = formatMariasState(
      makeMariasState({ phase: 2, roundCardPoints: [55, 35, 30], roundMarriage: [40, 0, 0] }),
    );
    expect(out).toContain('round result: card pts P0=55 P1=35 P2=30');
    expect(out).toContain('marriage: P0=40');
  });

  it('renders a hint with card indices', () => {
    const out = formatMariasState(makeMariasState({ hint: { cardIndices: [1, 2], reason: 'follow_win' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatMariasState(makeMariasState({ phase: 3, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatMariasState(makeMariasState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
