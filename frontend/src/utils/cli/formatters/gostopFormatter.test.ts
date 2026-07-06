import { describe, expect, it } from 'vitest';
import { makeGoStopState } from '../../../test/stateFactories';
import { formatGoStopState } from './gostopFormatter';

describe('formatGoStopState', () => {
  it('formats a play-phase state with the human hand and breakdown', () => {
    const out = formatGoStopState(makeGoStopState());
    expect(out).toContain('Go-Stop');
    expect(out).toContain('phase: Play');
    expect(out).toContain('あなた');
    expect(out).toContain('gwang:0');
  });

  it('renders (empty) when the field is empty', () => {
    const out = formatGoStopState(makeGoStopState({ fieldCards: [] }));
    expect(out).toContain('field: (empty)');
  });

  it('formats the go/stop decision line', () => {
    const out = formatGoStopState(
      makeGoStopState({
        phase: 1,
        pendingPoints: 7,
        pendingBreakdown: {
          gwang: 3,
          godori: 0,
          tti: 2,
          yeol: 1,
          pi: 1,
          base: 7,
          goCount: 0,
          goMult: 1,
          goScore: 7,
          brightCount: 3,
          ribbonCount: 5,
          animalCount: 5,
          piCount: 10,
        },
      }),
    );
    expect(out).toContain('decision:');
    expect(out).toContain('(go / stop)');
  });

  it('formats a round result with bak flags', () => {
    const out = formatGoStopState(
      makeGoStopState({
        phase: 2,
        lastRoundResult: {
          winner: 0,
          breakdown: null,
          basePoints: 7,
          goScore: 7,
          bakMult: 2,
          total: 14,
          gwangBak: true,
          piBak: true,
          goBak: false,
          goCount: 1,
        },
      }),
    );
    expect(out).toContain('result: winner=あなた');
    expect(out).toContain('gwang-bak');
    expect(out).toContain('pi-bak');
  });

  it('formats a drawn round result', () => {
    const out = formatGoStopState(
      makeGoStopState({
        phase: 2,
        lastRoundResult: {
          winner: -1,
          breakdown: null,
          basePoints: 0,
          goScore: 0,
          bakMult: 1,
          total: 0,
          gwangBak: false,
          piBak: false,
          goBak: false,
          goCount: 0,
        },
      }),
    );
    expect(out).toContain('winner=draw');
  });

  it('renders a decision hint line', () => {
    const out = formatGoStopState(
      makeGoStopState({ hint: { cardIndex: -1, fieldIndex: -1, go: 1, reason: 'go_lowscore' } }),
    );
    expect(out).toContain('HINT: go');
  });

  it('renders a play hint line', () => {
    const out = formatGoStopState(
      makeGoStopState({ hint: { cardIndex: 2, fieldIndex: 1, go: -1, reason: 'capture' } }),
    );
    expect(out).toContain('HINT: play 2 field 1');
  });

  it('appends the message when present', () => {
    const out = formatGoStopState(makeGoStopState({ message: 'hello' }));
    expect(out).toContain('hello');
  });
});
