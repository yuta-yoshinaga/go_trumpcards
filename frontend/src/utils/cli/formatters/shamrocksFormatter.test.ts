import { describe, expect, it } from 'vitest';
import type { ShamrocksResponse } from '../../../types/card';
import { formatShamrocksState } from './shamrocksFormatter';

function makeState(overrides: Partial<ShamrocksResponse> = {}): ShamrocksResponse {
  return {
    fans: [
      [{ design: 'SPADE', value: 9 }],
      [
        { design: 'CLOVER', value: 3 },
        { design: 'SPADE', value: 8 },
      ],
      [{ design: 'DIAMOND', value: 1 }],
      [],
    ],
    foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
    phase: 0,
    moveCount: 5,
    canUndo: true,
    legalTargets: {
      'fan-0': { fan: [1], foundation: [] },
      'fan-1': { fan: [0], foundation: [] },
      'fan-2': { fan: [], foundation: [0] },
    },
    message: '',
    ...overrides,
  };
}

describe('formatShamrocksState', () => {
  it('formats a regular playing state with header, foundations, fans, and markers', () => {
    const text = formatShamrocksState(makeState());

    expect(text).toContain('Shamrocks');
    expect(text).toContain('phase: Playing  moves: 5  undo: yes');
    // Foundation 0 has Heart A, others empty
    expect(text).toContain('f0:♥A  f1:[  ]  f2:[  ]  f3:[  ]');
    // Fan 0, 1, 2 are movable (*), fan 3 is empty
    expect(text).toContain('fan  0*: [0]♠9');
    expect(text).toContain('fan  1*: [0]♣3 [1]♠8');
    expect(text).toContain('fan  2*: [0]♦A');
    expect(text).toContain('fan  3 : [empty]');
    // No hint requested, so no HINT line
    expect(text).not.toContain('HINT');
  });

  it('formats game clear and game over states', () => {
    const clearText = formatShamrocksState(
      makeState({ phase: 1, messageCode: 'shamrocks.gameClear', message: 'Clear!' }),
    );
    expect(clearText).toContain('phase: GameClear');
    expect(clearText).toContain('Congratulations! You win!');
    expect(clearText).toContain('Clear!');

    const overText = formatShamrocksState(makeState({ phase: 2, messageCode: 'shamrocks.gameOver', canUndo: false }));
    expect(overText).toContain('phase: GameOver  moves: 5  undo: no');
    expect(overText).toContain('Game Over!');
  });

  it('formats unknown phase gracefully', () => {
    const text = formatShamrocksState(makeState({ phase: 99 }));
    expect(text).toContain('phase: Phase 99');
  });

  it('detects deadlock during playing phase when no legal moves remain', () => {
    const deadlockedText = formatShamrocksState(makeState({ legalTargets: {} }));
    expect(deadlockedText).toContain('Deadlocked - no legal moves remain');
  });

  it('shows hint only when explicitly requested (isRequestedHint is true)', () => {
    // Fan to fan hint
    const fanHintState = makeState({
      hint: { fromFan: 1, toFan: 0, toFoundation: false },
      messageCode: 'shamrocks.hintAvailable',
    });
    const fanHintText = formatShamrocksState(fanHintState);
    expect(fanHintText).toContain('HINT: fan 1 → fan 0');

    // Fan to foundation hint
    const fndHintState = makeState({
      hint: { fromFan: 2, toFan: -1, toFoundation: true },
      messageCode: 'shamrocks.hintAvailable',
    });
    const fndHintText = formatShamrocksState(fndHintState);
    expect(fndHintText).toContain('HINT: fan 2 → foundation');
  });

  it('does NOT show hint line for passive hint responses where hint was not requested', () => {
    // Server returns hint in regular output, but messageCode is shamrocks.playing or absent
    const passiveHintState = makeState({
      hint: { fromFan: 1, toFan: 0, toFoundation: false },
      messageCode: 'shamrocks.playing',
    });
    const passiveText = formatShamrocksState(passiveHintState);
    expect(passiveText).not.toContain('HINT:');

    const noCodeHintState = makeState({
      hint: { fromFan: 1, toFan: 0, toFoundation: false },
      messageCode: undefined,
    });
    expect(formatShamrocksState(noCodeHintState)).not.toContain('HINT:');
  });

  it('shows "No hint available." when server returns shamrocks.noHint', () => {
    const noHintState = makeState({
      messageCode: 'shamrocks.noHint',
    });
    const text = formatShamrocksState(noHintState);
    expect(text).toContain('No hint available.');
    expect(text).not.toContain('HINT:');
  });
});
