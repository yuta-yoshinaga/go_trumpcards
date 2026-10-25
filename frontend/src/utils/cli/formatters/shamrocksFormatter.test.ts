import { describe, expect, it } from 'vitest';
import type { ShamrocksResponse } from '../../../types/card';
import { ShamrocksPhase } from '../../../types/phases';
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
    phase: ShamrocksPhase.PLAYING,
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
    // Fan 0, 1, 2 are movable (* on top card), fan 3 is empty
    expect(text).toContain('fan  0: ♠9*');
    expect(text).toContain('fan  1: ♣3 ♠8*');
    expect(text).toContain('fan  2: ♦A*');
    expect(text).toContain('fan  3: [empty]');
    expect(text).not.toContain('[0]♠9');
    // Moves guide and actions
    expect(text).toContain('Moves: use p <from> <to> or f <idx> for a fan marked *');
    expect(text).toContain('Other actions: ac, u, h, g, r');
    // No hint requested, so no HINT line
    expect(text).not.toContain('HINT');
  });

  it('marks the top (last) card of a movable fan with *, not the first card', () => {
    const text = formatShamrocksState(makeState());
    expect(text).toContain('fan  1: ♣3 ♠8*');
    expect(text).not.toContain('♣3*');
  });

  it('does not mark the top card of an unmovable fan with *', () => {
    const unmovableState = makeState({
      fans: [[{ design: 'HEART', value: 7 }]],
      legalTargets: {},
    });
    const text = formatShamrocksState(unmovableState);
    expect(text).toContain('fan  0: ♥7');
    expect(text).not.toContain('♥7*');
  });

  it('omits u from available actions when canUndo is false', () => {
    const text = formatShamrocksState(makeState({ canUndo: false }));
    expect(text).toContain('Other actions: ac, h, g, r');
    expect(text).not.toMatch(/Other actions:.*?\bu\b/);
  });

  it('shows only fan move guide and omits ac when only fan moves exist', () => {
    const fanOnlyText = formatShamrocksState(makeState({ legalTargets: { 'fan-0': { fan: [1], foundation: [] } } }));
    expect(fanOnlyText).toContain('Moves: use p <from> <to> for a fan marked *');
    expect(fanOnlyText).not.toContain('f <idx>');
    expect(fanOnlyText).toContain('Other actions: u, h, g, r');
    expect(fanOnlyText).not.toMatch(/Other actions:.*?\bac\b/);
  });

  it('shows only foundation move guide and includes ac when only foundation moves exist', () => {
    const fndOnlyText = formatShamrocksState(makeState({ legalTargets: { 'fan-2': { fan: [], foundation: [0] } } }));
    expect(fndOnlyText).toContain('Moves: use f <idx> for a fan marked *');
    expect(fndOnlyText).not.toContain('p <from> <to>');
    expect(fndOnlyText).toContain('Other actions: ac, u, h, g, r');
  });

  it('formats game clear and game over states with english message and suppresses server message', () => {
    const clearText = formatShamrocksState(
      makeState({ phase: ShamrocksPhase.GAME_CLEAR, messageCode: 'shamrocks.gameClear', message: 'ゲームクリア！' }),
    );
    expect(clearText).toContain('phase: GameClear');
    expect(clearText).toContain('Congratulations! You win!');
    expect(clearText).not.toContain('ゲームクリア！');
    expect(clearText).toContain('Game finished. Use r to start a new game.');
    expect(clearText).not.toContain('Other actions:');
    expect(clearText).not.toContain('Moves:');

    const overText = formatShamrocksState(
      makeState({
        phase: ShamrocksPhase.GAME_OVER,
        messageCode: 'shamrocks.gameOver',
        message: 'ゲームオーバー',
        canUndo: false,
      }),
    );
    expect(overText).toContain('phase: GameOver  moves: 5  undo: no');
    expect(overText).toContain('Game Over!');
    expect(overText).not.toContain('ゲームオーバー');
    expect(overText).toContain('Game finished. Use r to start a new game.');
    expect(overText).not.toContain('Other actions:');
    expect(overText).not.toContain('Moves:');
  });

  it('shows server message during playing phase', () => {
    const text = formatShamrocksState(makeState({ message: 'カードを動かしました' }));
    expect(text).toContain('カードを動かしました');
  });

  it('formats unknown phase gracefully', () => {
    const text = formatShamrocksState(makeState({ phase: 99 }));
    expect(text).toContain('phase: Phase 99');
  });

  it('detects deadlock during playing phase when no legal moves remain', () => {
    const deadlockedText = formatShamrocksState(makeState({ legalTargets: {} }));
    expect(deadlockedText).toContain('Deadlocked - no legal moves remain');
    expect(deadlockedText).not.toContain('Moves:');
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
