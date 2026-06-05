import { describe, expect, it } from 'vitest';
import { makeTressetteState } from '../../../test/stateFactories';
import { formatTressetteState } from './tressetteFormatter';

describe('formatTressetteState', () => {
  it('renders header, round/trick and team scores', () => {
    const out = formatTressetteState(makeTressetteState());
    expect(out).toContain('Tressette');
    expect(out).toContain('round: 1');
    expect(out).toContain('Team A:');
    expect(out).toContain('Team B:');
  });

  it('shows the human hand with indexed cards', () => {
    const out = formatTressetteState(makeTressetteState());
    expect(out).toContain('[Team A]');
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick', () => {
    const out = formatTressetteState(
      makeTressetteState({
        currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 3 } }],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders hint line when present', () => {
    const out = formatTressetteState(makeTressetteState({ hint: { cardIndices: [1], reason: 'follow_win' } }));
    expect(out).toContain('HINT:');
    expect(out).toContain('follow_win');
  });

  it('renders game over with winning team', () => {
    const out = formatTressetteState(makeTressetteState({ gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('Team B');
  });
});
