import { describe, expect, it } from 'vitest';
import { makeSedmaState } from '../../../test/stateFactories';
import { formatSedmaState } from './sedmaFormatter';

describe('formatSedmaState', () => {
  it('renders the header, round/trick and team scores (no trump line)', () => {
    const out = formatSedmaState(makeSedmaState({ teamScores: [80, 60] }));
    expect(out).toContain('Sedma');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('score: A=80  B=60');
    expect(out).not.toContain('trump');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatSedmaState(makeSedmaState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatSedmaState(
      makeSedmaState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 7 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the round result during RoundEnd', () => {
    const out = formatSedmaState(makeSedmaState({ phase: 2, roundCardPoints: [30, 20] }));
    expect(out).toContain('round result: A card pts=30 B card pts=20');
  });

  it('renders a hint with card indices', () => {
    const out = formatSedmaState(makeSedmaState({ hint: { cardIndices: [1, 2], reason: 'capture' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('capture');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatSedmaState(makeSedmaState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatSedmaState(makeSedmaState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
