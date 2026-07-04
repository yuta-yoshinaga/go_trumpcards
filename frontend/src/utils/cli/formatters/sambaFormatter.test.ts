import { describe, expect, it } from 'vitest';
import { makeSambaState } from '../../../test/stateFactories';
import { formatSambaState } from './sambaFormatter';

describe('formatSambaState', () => {
  it('renders the header, round, phase, and team scores', () => {
    const out = formatSambaState(makeSambaState({ teamScores: [120, 45] }));
    expect(out).toContain('Samba');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: DRAW');
    expect(out).toContain('team 0: 120');
    expect(out).toContain('team 1: 45');
  });

  it('renders the frozen marker when the pile is frozen', () => {
    const out = formatSambaState(makeSambaState({ isFrozen: true }));
    expect(out).toContain('[FROZEN]');
  });

  it('shows the empty discard placeholder when there is no top card', () => {
    const out = formatSambaState(makeSambaState({ discardTop: null }));
    expect(out).toContain('[  ]');
  });

  it('labels a completed set as a canasta and a sequence meld as a samba', () => {
    const state = makeSambaState({
      players: [
        {
          ...makeSambaState().players[0],
          melds: [
            {
              cards: [
                { design: 'SPADE', value: 7 },
                { design: 'CLOVER', value: 7 },
                { design: 'HEART', value: 7 },
                { design: 'DIAMOND', value: 7 },
                { design: 'SPADE', value: 7 },
                { design: 'CLOVER', value: 7 },
                { design: 'HEART', value: 7 },
              ],
              kind: 0,
              isNatural: true,
              isCanasta: true,
              isSamba: false,
              rank: 7,
            },
            {
              cards: [
                { design: 'HEART', value: 4 },
                { design: 'HEART', value: 5 },
                { design: 'HEART', value: 6 },
                { design: 'HEART', value: 7 },
                { design: 'HEART', value: 8 },
                { design: 'HEART', value: 9 },
                { design: 'HEART', value: 10 },
              ],
              kind: 1,
              isNatural: true,
              isCanasta: false,
              isSamba: true,
              rank: 4,
            },
          ],
          hasCanasta: true,
          hasSamba: true,
        },
        ...makeSambaState().players.slice(1),
      ],
    });
    const out = formatSambaState(state);
    expect(out).toContain('(canasta)');
    expect(out).toContain('(samba)');
    expect(out).toContain('Canasta');
    expect(out).toContain('Samba');
  });

  it('renders a game-over line with the winning team when the game has ended', () => {
    const out = formatSambaState(makeSambaState({ gameEndFlag: true, winnerIdx: 1, phase: 4 }));
    expect(out).toContain('Game Over');
    expect(out).toContain('Winning team: 1');
  });

  it('renders the turn line while the game is in progress', () => {
    const out = formatSambaState(makeSambaState());
    expect(out).toContain('turn:');
  });
});
