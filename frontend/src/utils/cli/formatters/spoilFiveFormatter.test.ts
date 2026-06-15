import { describe, expect, it } from 'vitest';
import { makeSpoilFiveState } from '../../../test/stateFactories';
import { formatSpoilFiveState } from './spoilFiveFormatter';

describe('formatSpoilFiveState', () => {
  it('renders the header, round/trick, trump and pot', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState({ trumpSuit: 3, pot: 5 }));
    expect(out).toContain('Spoil Five');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('pot: 5');
  });

  it('renders per-player round tricks and score', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState());
    expect(out).toContain('roundTricks=0');
    expect(out).toContain('score=0');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatSpoilFiveState(
      makeSpoilFiveState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 7 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a Spoil banner at round end with no winner', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState({ phase: 2, roundWinnerIdx: -1 }));
    expect(out).toContain('SPOIL!');
  });

  it('renders a hint with card indices', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState({ hint: { cardIndices: [1, 2], reason: 'take_trick' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('take_trick');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState({ phase: 3, gameEndFlag: true, winnerPlayer: 0 }));
    expect(out).toContain('Game Over! Winner:');
  });

  it('renders an explicit message when present', () => {
    const out = formatSpoilFiveState(makeSpoilFiveState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
