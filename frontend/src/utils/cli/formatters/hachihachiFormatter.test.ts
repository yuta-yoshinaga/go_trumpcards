import { describe, expect, it } from 'vitest';
import { makeHachiHachiState } from '../../../test/stateFactories';
import { formatHachiHachiState } from './hachihachiFormatter';

describe('formatHachiHachiState', () => {
  it('formats a play-phase state with header, field, and player lines', () => {
    const out = formatHachiHachiState(makeHachiHachiState());
    expect(out).toContain('Hachi-Hachi');
    expect(out).toContain('phase: Play');
    expect(out).toContain('field:');
    expect(out).toContain('raw=0');
    // The human's hand is shown.
    expect(out).toMatch(/\[0\]/);
  });

  it('shows (empty) when the field is empty', () => {
    const out = formatHachiHachiState(makeHachiHachiState({ fieldCards: [] }));
    expect(out).toContain('field: (empty)');
  });

  it('renders per-player yaku when present', () => {
    const state = makeHachiHachiState();
    state.players[0].yaku = [{ key: 'sanko', points: 40 }];
    const out = formatHachiHachiState(state);
    expect(out).toContain('yaku=[sanko:40]');
  });

  it('renders the round-end settlement with a best marker and signed deltas', () => {
    const out = formatHachiHachiState(
      makeHachiHachiState({
        phase: 1,
        lastRoundResult: {
          best: 0,
          scores: [
            { playerIdx: 0, rawScore: 100, yaku: [{ key: 'sanko', points: 40 }], bonus: 40, delta: 52 },
            { playerIdx: 1, rawScore: 80, yaku: [], bonus: 0, delta: -8 },
          ],
        },
      }),
    );
    expect(out).toContain('result:');
    expect(out).toContain('+52');
    expect(out).toContain('-8');
    expect(out).toContain('*');
  });

  it('renders a hint line when a hint is present', () => {
    const out = formatHachiHachiState(
      makeHachiHachiState({ hint: { cardIndex: 2, fieldIndex: 1, reason: 'capture' } }),
    );
    expect(out).toContain('HINT: play 2 field 1 (capture)');
  });

  it('includes a message when present', () => {
    const out = formatHachiHachiState(makeHachiHachiState({ message: 'ゲーム終了！' }));
    expect(out).toContain('ゲーム終了！');
  });
});
