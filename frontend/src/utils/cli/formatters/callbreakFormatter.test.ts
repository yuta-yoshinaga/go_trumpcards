import { describe, expect, it } from 'vitest';
import { makeCallBreakState } from '../../../test/stateFactories';
import { formatCallBreakState } from './callbreakFormatter';

describe('formatCallBreakState', () => {
  it('renders header and round/trick line', () => {
    const out = formatCallBreakState(makeCallBreakState());
    expect(out).toContain('Call Break');
    expect(out).toContain('round: 1/5');
    expect(out).toContain('trick: 1');
    expect(out).toContain('spades broken: no');
  });

  it('renders spades broken yes when broken', () => {
    expect(formatCallBreakState(makeCallBreakState({ spadesBroken: true }))).toContain('spades broken: yes');
  });

  it('formats decimal scores via divmod 10 (41 → 4.1)', () => {
    const out = formatCallBreakState(makeCallBreakState());
    // CPU 1 in the factory has cumulativeScore=41
    expect(out).toContain('total=4.1');
  });

  it('formats negative decimal scores', () => {
    const state = makeCallBreakState();
    state.players[0].cumulativeScore = -40;
    state.players[0].roundScore = -40;
    const out = formatCallBreakState(state);
    expect(out).toContain('total=-4.0');
    expect(out).toContain('round=-4.0');
  });

  it('lists human cards with indices when human has cards', () => {
    const out = formatCallBreakState(makeCallBreakState());
    expect(out).toContain('[0]');
  });

  it('does not render bidding hint outside bid phase', () => {
    const out = formatCallBreakState(makeCallBreakState({ phase: 1 }));
    expect(out).not.toContain('Bidding phase');
  });

  it('renders bidding hint in phase 0', () => {
    expect(formatCallBreakState(makeCallBreakState({ phase: 0 }))).toContain('Bidding phase');
  });

  it('renders current trick line when trick has cards', () => {
    const state = makeCallBreakState({
      currentTrick: [
        { playerIdx: 0, card: { design: 'HEART', value: 5 } },
        { playerIdx: 1, card: { design: 'HEART', value: 9 } },
      ],
    });
    const out = formatCallBreakState(state);
    // formatCard renders suits as glyphs (♥), so match values + the suit glyph.
    expect(out).toMatch(/trick:.*♥5.*♥9/);
  });

  it('omits trick line when no cards played', () => {
    expect(formatCallBreakState(makeCallBreakState())).not.toMatch(/^trick:/m);
  });

  it('renders bid hint when hint.bid is present', () => {
    const out = formatCallBreakState(
      makeCallBreakState({
        messageCode: 'callbreak.hintRequested',
        hint: { bid: 4, reason: 'strategic_bid' },
      }),
    );
    expect(out).toContain('HINT: bid 4');
  });

  it('renders play hint when hint.cardIndex is present', () => {
    const out = formatCallBreakState(
      makeCallBreakState({
        messageCode: 'callbreak.hintRequested',
        hint: { cardIndex: 2, reason: 'follow_suit' },
      }),
    );
    expect(out).toContain('HINT: play [2]');
  });

  it('emits message line when message is non-empty', () => {
    expect(formatCallBreakState(makeCallBreakState({ message: 'hello' }))).toContain('hello');
  });

  it('emits Game Over when gameEndFlag is true', () => {
    expect(formatCallBreakState(makeCallBreakState({ gameEndFlag: true }))).toContain('Game Over');
  });

  it('does not emit Game Over when gameEndFlag is false', () => {
    expect(formatCallBreakState(makeCallBreakState())).not.toContain('Game Over');
  });

  it('handles unknown player index in trick gracefully', () => {
    const state = makeCallBreakState({
      currentTrick: [{ playerIdx: 99, card: { design: 'HEART', value: 5 } }],
    });
    expect(() => formatCallBreakState(state)).not.toThrow();
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { bid: 4, reason: 'strategic_bid' };
    expect(formatCallBreakState(makeCallBreakState({ hint, messageCode: 'callbreak.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatCallBreakState(makeCallBreakState({ hint, messageCode: 'callbreak.playing' }))).not.toContain('HINT');
  });
});
