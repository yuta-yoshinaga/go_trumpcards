import { describe, expect, it } from 'vitest';
import type { MonteCarloResponse } from '../../../types/card';
import { formatMonteCarloState } from './montecarloFormatter';

const board = Array.from({ length: 5 }, (_, row) =>
  Array.from({ length: 5 }, (_, col) => ({
    card: row === 0 && col < 2 ? { design: col === 0 ? ('SPADE' as const) : ('HEART' as const), value: 7 } : null,
  })),
);
const base: MonteCarloResponse = {
  board,
  phase: 0,
  stockCount: 27,
  removedCount: 0,
  dealCount: 1,
  removablePairs: 1,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'montecarlo.playing',
};

describe('formatMonteCarloState', () => {
  it('prints a playable board with flat 0-based p indices and next action', () => {
    const output = formatMonteCarloState(base);
    expect(output).toContain('phase: Playing');
    expect(output).toContain('[0]♠7 [1]♥7');
    expect(output).toContain('[24][empty]');
    expect(output).toContain('p <a> <b>');
    expect(output).toContain('Available commands: p 0 1, d (deal), u (undo)');
  });

  it('suggests dealing when no pair remains and stock is available', () => {
    const output = formatMonteCarloState({
      ...base,
      board: base.board.map((row) => row.map(() => ({ card: null }))),
      removablePairs: 0,
      canUndo: false,
    });
    expect(output).toContain('Available commands: d (deal)');
    expect(output).toContain('undo available: no');
  });

  it('shows only currently available commands and reset after game end', () => {
    const unavailable = formatMonteCarloState({
      ...base,
      board: base.board.map((row) => row.map(() => ({ card: null }))),
      removablePairs: 0,
      stockCount: 0,
      canUndo: false,
    });
    // A stalemate still offers hint, give up and reset, but no pair, deal or undo.
    expect(unavailable).toContain('Available commands: h (hint), g (give up), r (reset)');
    expect(unavailable).not.toMatch(/\bp \d|d \(deal\)|u \(undo\)/);
    expect(formatMonteCarloState({ ...base, phase: 1 })).toContain('Available command: r (reset)');
    expect(formatMonteCarloState({ ...base, phase: 2 })).toContain('Available command: r (reset)');
  });

  it('formats requested remove and deal hints from server hintAvailable responses', () => {
    const removeHint = formatMonteCarloState({
      ...base,
      hint: { action: 'remove', fromR: 0, fromC: 0, toR: 0, toC: 1 },
      messageCode: 'montecarlo.hintAvailable',
    });
    expect(removeHint).toContain('HINT: Remove [0] and [1].');
    const dealHint = formatMonteCarloState({
      ...base,
      hint: { action: 'deal' },
      messageCode: 'montecarlo.hintAvailable',
    });
    expect(dealHint).toContain('HINT: Deal cards from the stock.');
    const partialHint = formatMonteCarloState({
      ...base,
      board: base.board.map((row, rowIndex) => (rowIndex === 0 ? row.map(() => ({ card: null })) : row)),
      hint: { action: 'remove', fromR: 1, fromC: 0, toR: 1, toC: 1 },
      messageCode: 'montecarlo.hintAvailable',
    });
    expect(partialHint).toContain('HINT: Remove [5] and [6].');
  });

  it('does not show passive or unmarked hints', () => {
    const passive = formatMonteCarloState({ ...base, hint: { action: 'deal' } });
    expect(passive).not.toContain('HINT:');
    const unmarked = formatMonteCarloState({
      ...base,
      hint: { action: 'deal' },
      messageCode: 'montecarlo.playing',
    });
    expect(unmarked).not.toContain('HINT:');
    expect(
      formatMonteCarloState({
        ...base,
        hint: { action: 'deal' },
        stockCount: 0,
        messageCode: 'montecarlo.hintAvailable',
      }),
    ).not.toContain('HINT:');
    expect(formatMonteCarloState({ ...base, messageCode: 'montecarlo.noHint' })).toContain('No hint available.');
  });

  it('formats clear and game over phases', () => {
    expect(formatMonteCarloState({ ...base, phase: 1, removedCount: 52 })).toContain('Game clear!');
    expect(formatMonteCarloState({ ...base, phase: 2, removablePairs: 0, stockCount: 0 })).toContain(
      'No more moves. Game over.',
    );
  });

  it('handles unknown phases and appends response messages', () => {
    const output = formatMonteCarloState({ ...base, phase: 9, message: 'Server message' });
    expect(output).toContain('phase: 9');
    expect(output).toContain('Server message');
  });
});
