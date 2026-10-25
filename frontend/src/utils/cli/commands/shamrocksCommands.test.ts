import { describe, expect, it } from 'vitest';
import type { ShamrocksResponse } from '../../../types/card';
import { parseShamrocksCommand, SHAMROCKS_HELP, VALID_COMMANDS } from './shamrocksCommands';

function makeState(overrides: Partial<ShamrocksResponse> = {}): ShamrocksResponse {
  return {
    fans: [[{ design: 'SPADE', value: 9 }], [{ design: 'SPADE', value: 8 }], [{ design: 'DIAMOND', value: 1 }], []],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
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

describe('parseShamrocksCommand (stateless)', () => {
  it('parses fan-to-fan moves with m, move, and mf', () => {
    // Matches the args page passes to shamrocksApi.exec('mf', selected, idx)
    expect(parseShamrocksCommand('m 0 1')).toEqual({ args: ['mf', 0, 1] });
    expect(parseShamrocksCommand('move 0 1')).toEqual({ args: ['mf', 0, 1] });
    expect(parseShamrocksCommand('mf 0 1')).toEqual({ args: ['mf', 0, 1] });
    expect(parseShamrocksCommand('m fan-0 fan-1')).toEqual({ args: ['mf', 0, 1] });
    expect(parseShamrocksCommand('m fan0 fan1')).toEqual({ args: ['mf', 0, 1] });
  });

  it('parses fan-to-foundation moves with m, move, and ff', () => {
    // Matches the args page passes to shamrocksApi.exec('ff', selected)
    expect(parseShamrocksCommand('m 2 f')).toEqual({ args: ['ff', 2] });
    expect(parseShamrocksCommand('move 2 foundation')).toEqual({ args: ['ff', 2] });
    expect(parseShamrocksCommand('ff 2')).toEqual({ args: ['ff', 2] });
    expect(parseShamrocksCommand('ff fan-2')).toEqual({ args: ['ff', 2] });
  });

  it('parses action commands matching page exec calls', () => {
    expect(parseShamrocksCommand('ac')).toEqual({ args: ['ac'] });
    expect(parseShamrocksCommand('autocomplete')).toEqual({ args: ['ac'] });
    expect(parseShamrocksCommand('u')).toEqual({ args: ['u'] });
    expect(parseShamrocksCommand('undo')).toEqual({ args: ['u'] });
    expect(parseShamrocksCommand('h')).toEqual({ args: ['hint'] });
    expect(parseShamrocksCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseShamrocksCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseShamrocksCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseShamrocksCommand('r')).toEqual({ args: ['reset'] });
    expect(parseShamrocksCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects invalid or missing arguments without state', () => {
    expect(parseShamrocksCommand('m')).toEqual({ error: 'Usage: m <from> <to> or m <from> f' });
    expect(parseShamrocksCommand('m 0')).toEqual({ error: 'Usage: m <from> <to> or m <from> f' });
    expect(parseShamrocksCommand('m invalid 1')).toEqual({
      error: 'Invalid source fan: use <index> (e.g., 0, fan-0)',
    });
    expect(parseShamrocksCommand('m fan- 1')).toEqual({
      error: 'Invalid source fan: use <index> (e.g., 0, fan-0)',
    });
    expect(parseShamrocksCommand('m 0 invalid')).toEqual({
      error: 'Invalid target: use <fan> or f (foundation)',
    });
    expect(parseShamrocksCommand('m 0 fan-')).toEqual({
      error: 'Invalid target: use <fan> or f (foundation)',
    });

    expect(parseShamrocksCommand('mf')).toEqual({ error: 'Usage: mf <from> <to>' });
    expect(parseShamrocksCommand('mf 0')).toEqual({ error: 'Usage: mf <from> <to>' });
    expect(parseShamrocksCommand('mf invalid 1')).toEqual({ error: 'Usage: mf <from> <to>' });
    expect(parseShamrocksCommand('mf 0 invalid')).toEqual({ error: 'Usage: mf <from> <to>' });

    expect(parseShamrocksCommand('ff')).toEqual({ error: 'Usage: ff <from>' });
    expect(parseShamrocksCommand('ff invalid')).toEqual({ error: 'Usage: ff <from>' });
  });

  it('rejects unknown commands', () => {
    expect(parseShamrocksCommand('foobar')).toHaveProperty('error');
    // Log command is intentionally not implemented
    expect(parseShamrocksCommand('log')).toHaveProperty('error');
    expect(parseShamrocksCommand('l')).toHaveProperty('error');
  });

  it('provides valid commands and help text', () => {
    expect(VALID_COMMANDS.length).toBeGreaterThan(0);
    expect(SHAMROCKS_HELP.length).toBeGreaterThan(0);
  });
});

describe('parseShamrocksCommand (state-aware)', () => {
  const state = makeState();

  it('allows legal moves when state is provided', () => {
    expect(parseShamrocksCommand('m 0 1', state)).toEqual({ args: ['mf', 0, 1] });
    expect(parseShamrocksCommand('ff 2', state)).toEqual({ args: ['ff', 2] });
    expect(parseShamrocksCommand('m 2 f', state)).toEqual({ args: ['ff', 2] });
    expect(parseShamrocksCommand('u', state)).toEqual({ args: ['u'] });
    expect(parseShamrocksCommand('ac', state)).toEqual({ args: ['ac'] });
    expect(parseShamrocksCommand('h', state)).toEqual({ args: ['hint'] });
    expect(parseShamrocksCommand('g', state)).toEqual({ args: ['giveup'] });
    expect(parseShamrocksCommand('r', state)).toEqual({ args: ['reset'] });
  });

  it('rejects moves when game is already ended', () => {
    const clearState = makeState({ phase: 1 });
    const overState = makeState({ phase: 2 });

    for (const s of [clearState, overState]) {
      expect(parseShamrocksCommand('m 0 1', s)).toEqual({ error: 'Game is already over' });
      expect(parseShamrocksCommand('ff 2', s)).toEqual({ error: 'Game is already over' });
      expect(parseShamrocksCommand('ac', s)).toEqual({ error: 'Game is already over' });
      expect(parseShamrocksCommand('u', s)).toEqual({ error: 'Game is already over' });
      expect(parseShamrocksCommand('h', s)).toEqual({ error: 'Game is already over' });
      expect(parseShamrocksCommand('g', s)).toEqual({ error: 'Game is already over' });
      // Reset is still allowed
      expect(parseShamrocksCommand('r', s)).toEqual({ args: ['reset'] });
    }
  });

  it('rejects out of range or empty source fans', () => {
    expect(parseShamrocksCommand('m -1 1', state)).toEqual({
      error: 'Invalid source fan index: -1 (valid: 0-3)',
    });
    expect(parseShamrocksCommand('m 4 1', state)).toEqual({
      error: 'Invalid source fan index: 4 (valid: 0-3)',
    });
    expect(parseShamrocksCommand('ff 99', state)).toEqual({
      error: 'Invalid source fan index: 99 (valid: 0-3)',
    });
    // fan 3 is empty in makeState()
    expect(parseShamrocksCommand('m 3 0', state)).toEqual({ error: 'Fan 3 is empty' });
    expect(parseShamrocksCommand('ff 3', state)).toEqual({ error: 'Fan 3 is empty' });

    const sparseState = makeState({ fans: [undefined as unknown as []] });
    expect(parseShamrocksCommand('m 0 1', sparseState)).toEqual({ error: 'Fan 0 is empty' });
    expect(parseShamrocksCommand('ff 0', sparseState)).toEqual({ error: 'Fan 0 is empty' });
  });

  it('rejects self moves and out of range destination fans', () => {
    expect(parseShamrocksCommand('m 0 0', state)).toEqual({ error: 'Cannot move a fan onto itself' });
    expect(parseShamrocksCommand('m 0 -1', state)).toEqual({
      error: 'Invalid destination fan index: -1 (valid: 0-3)',
    });
    expect(parseShamrocksCommand('m 0 4', state)).toEqual({
      error: 'Invalid destination fan index: 4 (valid: 0-3)',
    });
  });

  it('rejects illegal moves not present in legalTargets', () => {
    // fan 0 can only move to fan 1 in makeState()
    expect(parseShamrocksCommand('m 0 2', state)).toEqual({
      error: 'Illegal move: cannot move fan 0 to fan 2',
    });
    // fan 0 has no foundation move in makeState()
    expect(parseShamrocksCommand('ff 0', state)).toEqual({
      error: 'Card from fan 0 cannot be moved to foundation',
    });

    const noTargetsState = makeState({ legalTargets: {} });
    expect(parseShamrocksCommand('m 0 1', noTargetsState)).toEqual({
      error: 'Illegal move: cannot move fan 0 to fan 1',
    });
    expect(parseShamrocksCommand('ff 2', noTargetsState)).toEqual({
      error: 'Card from fan 2 cannot be moved to foundation',
    });
  });

  it('rejects undo when canUndo is false', () => {
    const cannotUndoState = makeState({ canUndo: false });
    expect(parseShamrocksCommand('u', cannotUndoState)).toEqual({ error: 'Cannot undo' });
    expect(parseShamrocksCommand('undo', cannotUndoState)).toEqual({ error: 'Cannot undo' });
  });
});
