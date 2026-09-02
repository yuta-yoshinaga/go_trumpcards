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

  // **切り札は配りで決まる。**CUI は出しているのに CLI モードだけが落として
  // いた (#6615)。決まっている場合と未確定の場合の両方を踏む。
  it('shows the dealt trump suit, and a dash while it is undecided', () => {
    expect(formatMadrassoState(makeMadrassoState({ trumpSuit: 3 }))).toContain('trump: ♥');
    expect(formatMadrassoState(makeMadrassoState({ trumpSuit: 1 }))).toContain('trump: ♠');
    // 未確定 (0) をハートと言い張らないこと。
    const undecided = formatMadrassoState(makeMadrassoState({ trumpSuit: 0 }));
    expect(undecided).toContain('trump: -');
    expect(undecided).not.toContain('♥');

    // 表に無い値が来ても字形を捏造せず「未確定」と同じ扱いにする
    // (`SUIT_NAMES[...] ?? '-'` の右辺。0 は表に載っているので踏まない)。
    expect(formatMadrassoState(makeMadrassoState({ trumpSuit: 9 }))).toContain('trump: -');
  });
});
