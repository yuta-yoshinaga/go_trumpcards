import { describe, expect, it } from 'vitest';
import { makeMinchiateState } from '../../../test/stateFactories';
import { MINCHIATE_SURPLUS } from '../../../types/card';
import { formatMinchiateState } from './minchiateFormatter';

describe('formatMinchiateState', () => {
  it('renders the header, round and team scores', () => {
    const text = formatMinchiateState(makeMinchiateState({ teamScores: [9, 6] }));
    expect(text).toContain('Minchiate');
    expect(text).toContain('round: 1');
    expect(text).toContain('team0=9');
    expect(text).toContain('team1=6');
  });

  // 40枚という梯子の長さは手札からは読めないので、端末表示にも必ず出す。
  it('states the 40-trump count', () => {
    expect(formatMinchiateState(makeMinchiateState())).toContain('trumps: 40');
  });

  it('lists the phase name', () => {
    expect(formatMinchiateState(makeMinchiateState({ phase: 0 }))).toContain('phase: Scarto');
    expect(formatMinchiateState(makeMinchiateState({ phase: 2 }))).toContain('phase: TrickEnd');
  });

  it('marks the dealer and shows every seat with its team', () => {
    const text = formatMinchiateState(makeMinchiateState());
    expect(text).toContain('[dealer]');
    expect(text).toContain('(team 0)');
    expect(text).toContain('(team 1)');
  });

  it('shows the human hand with indices and hides the CPU hands', () => {
    const text = formatMinchiateState(makeMinchiateState());
    expect(text).toContain('[0]');
    expect(text).toContain('cards=21');
  });

  it('reports the buried cards once the scarto is done', () => {
    expect(formatMinchiateState(makeMinchiateState({ scartoCount: MINCHIATE_SURPLUS }))).toContain(
      `scarto: ${MINCHIATE_SURPLUS} cards buried`,
    );
    expect(formatMinchiateState(makeMinchiateState({ scartoCount: 0 }))).not.toContain('scarto:');
  });

  it('renders the current trick', () => {
    const text = formatMinchiateState(
      makeMinchiateState({
        currentTrick: [
          {
            playerIdx: 1,
            card: { design: 'JOKER', value: 39, glyph: '✦', label: 'Angelo', color: 'purple', deck: 'tarot' },
          },
        ],
      }),
    );
    expect(text).toContain('trick:');
    expect(text).toContain('Angelo');
  });

  it('shows the round tally with team labels at round end', () => {
    const text = formatMinchiateState(makeMinchiateState({ phase: 3, roundTricks: [5, 4, 3, 3] }));
    expect(text).toContain('P0(T0)=5');
    expect(text).toContain('P1(T1)=4');
  });

  it('omits the round tally mid-play', () => {
    expect(formatMinchiateState(makeMinchiateState())).not.toContain('round result');
  });

  it('shows the hint only once it was requested', () => {
    const hint = { cardIndices: [2], reason: 'lead_trump' };
    expect(formatMinchiateState(makeMinchiateState({ hint }))).not.toContain('HINT:');
    const requested = formatMinchiateState(makeMinchiateState({ hint, messageCode: 'minchiate.hintRequested' }));
    expect(requested).toContain('HINT: card indices [2] (lead_trump)');
  });

  it('names the winning team at game end', () => {
    const text = formatMinchiateState(makeMinchiateState({ phase: 4, gameEndFlag: true, winnerTeam: 1 }));
    expect(text).toContain('Game Over! Team 1 wins!');
  });

  // 引き分けは winnerTeam = -1。"Team -1 wins" と書いてはならない。
  it('reports a draw rather than team -1', () => {
    const text = formatMinchiateState(makeMinchiateState({ phase: 4, gameEndFlag: true, winnerTeam: -1 }));
    expect(text).toContain('Game Over! Draw.');
    expect(text).not.toContain('Team -1');
  });
});
