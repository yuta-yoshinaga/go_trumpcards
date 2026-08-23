import { describe, expect, it } from 'vitest';
import { makeMadrassoState } from '../../../test/stateFactories';
import { formatMadrassoState } from './madrassoFormatter';

describe('formatMadrassoState', () => {
  it('renders header, round/trick and team scores', () => {
    const out = formatMadrassoState(makeMadrassoState());
    expect(out).toContain('Madrasso');
    expect(out).toContain('round: 1');
    expect(out).toContain('Team A:');
    expect(out).toContain('Team B:');
  });

  it('shows the human hand with indexed cards', () => {
    const out = formatMadrassoState(makeMadrassoState());
    expect(out).toContain('[Team A]');
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick', () => {
    const out = formatMadrassoState(
      makeMadrassoState({
        currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 3 } }],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders hint line when present', () => {
    const out = formatMadrassoState(
      makeMadrassoState({ hint: { cardIndices: [1], reason: 'follow_win' }, messageCode: 'madrasso.hintRequested' }),
    );
    expect(out).toContain('HINT:');
    expect(out).toContain('follow_win');
  });

  it('renders game over with winning team', () => {
    const out = formatMadrassoState(makeMadrassoState({ gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('Team B');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'follow_win' };
    expect(formatMadrassoState(makeMadrassoState({ hint, messageCode: 'madrasso.hintRequested' }))).toContain('HINT');
    expect(formatMadrassoState(makeMadrassoState({ hint, messageCode: 'madrasso.playing' }))).not.toContain('HINT');
  });
});
