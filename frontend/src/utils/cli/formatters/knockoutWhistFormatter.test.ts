import { describe, expect, it } from 'vitest';
import { makeKnockoutWhistState } from '../../../test/stateFactories';
import { formatKnockoutWhistState } from './knockoutWhistFormatter';

describe('formatKnockoutWhistState', () => {
  it('renders the header, round/hand/trick, trump and active count', () => {
    const out = formatKnockoutWhistState(makeKnockoutWhistState({ trumpSuit: 3, activeCount: 4 }));
    expect(out).toContain('Knockout Whist');
    expect(out).toContain('round: 1');
    expect(out).toContain('hand: 7');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('active: 4');
  });

  it('renders per-player Dogbones and round tricks for active players', () => {
    const out = formatKnockoutWhistState(makeKnockoutWhistState());
    expect(out).toContain('dogbones=1');
    expect(out).toContain('roundTricks=0');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatKnockoutWhistState(makeKnockoutWhistState());
    expect(out).toMatch(/\[0\]/);
  });

  it('marks eliminated players', () => {
    const state = makeKnockoutWhistState();
    state.players[1] = { ...state.players[1], eliminated: true };
    const out = formatKnockoutWhistState(state);
    expect(out).toContain('ELIMINATED');
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatKnockoutWhistState(
      makeKnockoutWhistState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 7 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a hint with card indices', () => {
    const out = formatKnockoutWhistState(
      makeKnockoutWhistState({ hint: { cardIndices: [1, 2], reason: 'follow_win' } }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatKnockoutWhistState(makeKnockoutWhistState({ phase: 3, gameEndFlag: true, winnerPlayer: 0 }));
    expect(out).toContain('Game Over! Winner:');
  });

  it('renders an explicit message when present', () => {
    const out = formatKnockoutWhistState(makeKnockoutWhistState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
