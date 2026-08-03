import { describe, expect, it } from 'vitest';
import { makeTarocchiniState } from '../../../test/stateFactories';
import { formatTarocchiniState } from './tarocchiniFormatter';

describe('formatTarocchiniState', () => {
  it('renders the header, round and team scores', () => {
    const text = formatTarocchiniState(makeTarocchiniState({ teamScores: [9, 6] }));
    expect(text).toContain('Tarocchini');
    expect(text).toContain('round: 1');
    expect(text).toContain('team0=9');
    expect(text).toContain('team1=6');
  });

  // 後出し優先は手札からは読めないので、端末表示にも必ず出す。
  it('states the papi rule', () => {
    expect(formatTarocchiniState(makeTarocchiniState())).toContain('LATER-played wins');
  });

  it('lists the phase name', () => {
    expect(formatTarocchiniState(makeTarocchiniState({ phase: 0 }))).toContain('phase: Scarto');
    expect(formatTarocchiniState(makeTarocchiniState({ phase: 2 }))).toContain('phase: TrickEnd');
  });

  it('marks the dealer and shows every seat with its team', () => {
    const text = formatTarocchiniState(makeTarocchiniState());
    expect(text).toContain('[dealer]');
    expect(text).toContain('(team 0)');
    expect(text).toContain('(team 1)');
  });

  it('shows the human hand with indices and hides the CPU hands', () => {
    const text = formatTarocchiniState(makeTarocchiniState());
    expect(text).toContain('[0]');
    expect(text).toContain('cards=15');
  });

  it('reports the buried cards once the scarto is done', () => {
    expect(formatTarocchiniState(makeTarocchiniState({ scartoCount: 2 }))).toContain('scarto: 2 cards buried');
    expect(formatTarocchiniState(makeTarocchiniState({ scartoCount: 0 }))).not.toContain('scarto:');
  });

  it('renders the current trick', () => {
    const text = formatTarocchiniState(
      makeTarocchiniState({
        currentTrick: [
          {
            playerIdx: 1,
            card: { design: 'JOKER', value: 2, glyph: '✦', label: 'Papa', color: 'green', deck: 'tarot' },
          },
        ],
      }),
    );
    expect(text).toContain('trick:');
    expect(text).toContain('Papa');
  });

  it('shows the round tally with team labels at round end', () => {
    const text = formatTarocchiniState(makeTarocchiniState({ phase: 3, roundTricks: [5, 4, 3, 3] }));
    expect(text).toContain('P0(T0)=5');
    expect(text).toContain('P1(T1)=4');
  });

  it('omits the round tally mid-play', () => {
    expect(formatTarocchiniState(makeTarocchiniState())).not.toContain('round result');
  });

  it('shows the hint only once it was requested', () => {
    const hint = { cardIndices: [2], reason: 'play_papa' };
    expect(formatTarocchiniState(makeTarocchiniState({ hint }))).not.toContain('HINT:');
    const requested = formatTarocchiniState(makeTarocchiniState({ hint, messageCode: 'tarocchini.hintRequested' }));
    expect(requested).toContain('HINT: card indices [2] (play_papa)');
  });

  it('names the winning team at game end', () => {
    const text = formatTarocchiniState(makeTarocchiniState({ phase: 4, gameEndFlag: true, winnerTeam: 1 }));
    expect(text).toContain('Game Over! Team 1 wins!');
  });

  // 引き分けは winnerTeam = -1。"Team -1 wins" と書いてはならない。
  it('reports a draw rather than team -1', () => {
    const text = formatTarocchiniState(makeTarocchiniState({ phase: 4, gameEndFlag: true, winnerTeam: -1 }));
    expect(text).toContain('Game Over! Draw.');
    expect(text).not.toContain('Team -1');
  });
});
