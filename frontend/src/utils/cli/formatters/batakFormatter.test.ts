import { describe, expect, it } from 'vitest';
import { makeBatakState } from '../../../test/stateFactories';
import { formatBatakState } from './batakFormatter';

describe('formatBatakState', () => {
  it('renders header and round/trick line', () => {
    const out = formatBatakState(makeBatakState());
    expect(out).toContain('Batak');
    expect(out).toContain('round: 1/5');
    expect(out).toContain('trick: 1');
    expect(out).toContain('spades broken: no');
  });

  it('renders spades broken yes when broken', () => {
    expect(formatBatakState(makeBatakState({ spadesBroken: true }))).toContain('spades broken: yes');
  });

  it('formats raw integer scores', () => {
    const out = formatBatakState(makeBatakState());
    // CPU 1 in the factory has cumulativeScore=4
    expect(out).toContain('total=4');
  });

  it('formats negative integer scores', () => {
    const state = makeBatakState();
    state.players[0].cumulativeScore = -40;
    state.players[0].roundScore = -40;
    const out = formatBatakState(state);
    expect(out).toContain('total=-40');
    expect(out).toContain('round=-40');
  });

  it('renders declarer when declarerIdx is set', () => {
    const state = makeBatakState({ declarerIdx: 0 });
    const out = formatBatakState(state);
    expect(out).toContain('declarer:');
    expect(out).toContain('[declarer]');
  });

  it('renders pass when bid is 0', () => {
    const state = makeBatakState();
    state.players[1].bid = 0;
    const out = formatBatakState(state);
    expect(out).toContain('bid=pass');
  });

  it('lists human cards with indices when human has cards', () => {
    const out = formatBatakState(makeBatakState());
    expect(out).toContain('[0]');
  });

  it('does not render bidding hint outside bid phase', () => {
    const out = formatBatakState(makeBatakState({ phase: 1 }));
    expect(out).not.toContain('Bidding phase');
  });

  it('renders bidding hint in phase 0', () => {
    expect(formatBatakState(makeBatakState({ phase: 0 }))).toContain('Bidding phase (5-13 or pass)');
  });

  it('renders current trick line when trick has cards', () => {
    const state = makeBatakState({
      currentTrick: [
        { playerIdx: 0, card: { design: 'HEART', value: 5 } },
        { playerIdx: 1, card: { design: 'HEART', value: 9 } },
      ],
    });
    const out = formatBatakState(state);
    // formatCard renders suits as glyphs (♥), so match values + the suit glyph.
    expect(out).toMatch(/trick:.*♥5.*♥9/);
  });

  it('omits trick line when no cards played', () => {
    expect(formatBatakState(makeBatakState())).not.toMatch(/^trick:/m);
  });

  it('renders bid hint when hint.bid is present', () => {
    const out = formatBatakState(
      makeBatakState({
        messageCode: 'batak.hintRequested',
        hint: { bid: 5, reason: 'strategic_bid' },
      }),
    );
    expect(out).toContain('HINT: bid 5');
  });

  it('renders pass hint when hint.bid is 0', () => {
    const out = formatBatakState(
      makeBatakState({
        messageCode: 'batak.hintRequested',
        hint: { bid: 0, reason: 'pass_weak_hand' },
      }),
    );
    expect(out).toContain('HINT: pass (pass_weak_hand)');
  });

  it('renders play hint when hint.cardIndex is present', () => {
    const out = formatBatakState(
      makeBatakState({
        messageCode: 'batak.hintRequested',
        hint: { cardIndex: 2, reason: 'follow_suit' },
      }),
    );
    expect(out).toContain('HINT: play [2]');
  });

  it('emits message line when message is non-empty', () => {
    expect(formatBatakState(makeBatakState({ message: 'hello' }))).toContain('hello');
  });

  it('emits Game Over when gameEndFlag is true', () => {
    expect(formatBatakState(makeBatakState({ gameEndFlag: true }))).toContain('Game Over');
  });

  it('does not emit Game Over when gameEndFlag is false', () => {
    expect(formatBatakState(makeBatakState())).not.toContain('Game Over');
  });

  it('handles unknown player index in trick gracefully', () => {
    const state = makeBatakState({
      currentTrick: [{ playerIdx: 99, card: { design: 'HEART', value: 5 } }],
    });
    expect(() => formatBatakState(state)).not.toThrow();
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { bid: 4, reason: 'strategic_bid' };
    expect(formatBatakState(makeBatakState({ hint, messageCode: 'batak.hintRequested' }))).toContain('HINT');
    expect(formatBatakState(makeBatakState({ hint, messageCode: 'batak.playing' }))).not.toContain('HINT');
  });
});
