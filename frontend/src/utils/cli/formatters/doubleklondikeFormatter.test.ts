import { describe, expect, it } from 'vitest';
import type { DoubleKlondikeResponse } from '../../../types/card';
import { DoubleKlondikePhase } from '../../../types/phases';
import { formatDoubleKlondikeState } from './doubleklondikeFormatter';

const base: DoubleKlondikeResponse = {
  tableau: [
    [
      { card: { design: 'SPADE', value: 9 }, faceUp: true },
      { card: { design: 'HEART', value: 8 }, faceUp: true },
    ],
    [],
  ],
  stockCount: 4,
  waste: [
    { design: 'DIAMOND', value: 1 },
    { design: 'CLOVER', value: 13 },
  ],
  foundation: [[{ design: 'SPADE', value: 1 }], []],
  phase: DoubleKlondikePhase.PLAYING,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  legalTargets: {
    waste: { tableau: [1], foundation: [] },
    'tableau-0-1': { tableau: [1], foundation: [] },
  },
  message: '',
};

describe('formatDoubleKlondikeState', () => {
  it('formats a playing board with card-level move markers and available commands', () => {
    const output = formatDoubleKlondikeState(base);
    expect(output).toContain('phase: Playing');
    expect(output).toContain('[0]♠9 *[1]♥8');
    expect(output).toContain('waste: *♣K');
    expect(output).toContain('d (draw)');
  });
  it('marks the waste when it can move and omits draw guidance when stock and waste are empty', () => {
    const output = formatDoubleKlondikeState({
      ...base,
      stockCount: 0,
      waste: [],
      legalTargets: { waste: { tableau: [0], foundation: [] } },
    });
    expect(output).not.toContain('d (draw)');
    expect(output).toContain('waste: [empty]');
  });
  it('prints stalemate guidance and renders a real face-down card as hidden', () => {
    const output = formatDoubleKlondikeState({
      ...base,
      isStalemate: true,
      tableau: [[{ card: { design: 'SPADE', value: 7 }, faceUp: false }], []],
      legalTargets: {},
    });
    expect(output).toContain('Stalemate - no more moves possible');
    expect(output).toContain('[0]XX');
    expect(output).not.toContain('♠7');
  });
  it('reports a requested card hint with the server message code', () => {
    expect(
      formatDoubleKlondikeState({
        ...base,
        messageCode: 'doubleklondike.hintAvailable',
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 0 },
      }),
    ).toContain('HINT: Move column 0 card 1 to a foundation.');
  });
  it('prints no hint only for the server noHint response and ignores passive hints', () => {
    expect(formatDoubleKlondikeState({ ...base, messageCode: 'doubleklondike.noHint' })).toContain(
      'No hint available.',
    );
    expect(
      formatDoubleKlondikeState({
        ...base,
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 1 },
      }),
    ).not.toContain('HINT:');
  });
  it('formats the finished phase with reset only and no localized server message', () => {
    const output = formatDoubleKlondikeState({
      ...base,
      phase: DoubleKlondikePhase.GAME_CLEAR,
      message: 'ゲームクリア！',
    });
    expect(output).toContain('phase: Game Clear');
    expect(output).toContain('Available command: r (reset)');
    expect(output).not.toContain('ゲームクリア');
  });
  it('formats empty piles, hidden cards, unknown phase, and an English non-hint message', () => {
    const output = formatDoubleKlondikeState({
      ...base,
      phase: DoubleKlondikePhase.PLAYING,
      waste: [],
      foundation: [[]],
      tableau: [[{ card: null, faceUp: false }]],
      legalTargets: {},
      message: 'Move rejected',
      messageCode: 'other.error',
    });
    expect(output).toContain('phase: Playing');
    expect(output).toContain('waste: [empty]');
    expect(output).toContain('[0]XX');
    expect(output).toContain('Move rejected');
  });
  it('falls back to the raw phase for an unknown server phase', () => {
    expect(formatDoubleKlondikeState({ ...base, phase: 99 })).toContain('phase: 99');
  });
  it('formats a game over state and a waste hint', () => {
    const output = formatDoubleKlondikeState({
      ...base,
      phase: DoubleKlondikePhase.GAME_OVER,
      messageCode: 'doubleklondike.hintAvailable',
      hint: { fromZone: 'waste', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 1 },
    });
    expect(output).toContain('No more moves. Game over.');
    expect(output).toContain('HINT: Move waste to column 1.');
    expect(output).not.toContain('h (hint)');
  });
});
