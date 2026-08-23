import { describe, expect, it } from 'vitest';
import { makeTrappolaState } from '../../../test/stateFactories';
import { formatTrappolaState } from './trappolaFormatter';

describe('formatTrappolaState', () => {
  it('renders header, round/trick and team scores', () => {
    const out = formatTrappolaState(makeTrappolaState());
    expect(out).toContain('Trappola');
    expect(out).toContain('round: 1');
    expect(out).toContain('Team A:');
    expect(out).toContain('Team B:');
  });

  it('shows the human hand with indexed cards', () => {
    const out = formatTrappolaState(makeTrappolaState());
    expect(out).toContain('[Team A]');
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick', () => {
    const out = formatTrappolaState(
      makeTrappolaState({
        currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 3 } }],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders hint line when present', () => {
    const out = formatTrappolaState(
      makeTrappolaState({ hint: { cardIndices: [1], reason: 'follow_win' }, messageCode: 'trappola.hintRequested' }),
    );
    expect(out).toContain('HINT:');
    expect(out).toContain('follow_win');
  });

  it('renders game over with winning team', () => {
    const out = formatTrappolaState(makeTrappolaState({ gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('Team B');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'follow_win' };
    expect(formatTrappolaState(makeTrappolaState({ hint, messageCode: 'trappola.hintRequested' }))).toContain('HINT');
    expect(formatTrappolaState(makeTrappolaState({ hint, messageCode: 'trappola.playing' }))).not.toContain('HINT');
  });
});
