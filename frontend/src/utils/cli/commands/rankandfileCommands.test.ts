import { describe, expect, it } from 'vitest';
import { parseRankandfileCommand, RANKANDFILE_HELP } from './rankandfileCommands';

describe('parseRankandfileCommand', () => {
  it.each([
    ['d', ['draw']],
    ['draw', ['draw']],
    ['g', ['giveup']],
    ['giveup', ['giveup']],
    ['ac', ['autocomplete']],
    ['autocomplete', ['autocomplete']],
    ['u', ['undo']],
    ['undo', ['undo']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['log', ['log']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseRankandfileCommand(input)).toEqual({ args: expected });
  });

  it('parses tableau-to-tableau moves', () => {
    expect(parseRankandfileCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau moves with an explicit card index', () => {
    expect(parseRankandfileCommand('m t2 3 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 2, cardIndex: 3 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waste-to-tableau moves', () => {
    expect(parseRankandfileCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste', cardIndex: undefined }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses moves to any foundation', () => {
    expect(parseRankandfileCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses moves to a specific foundation', () => {
    expect(parseRankandfileCommand('m w f3')).toEqual({
      args: ['move', { zone: 'waste', cardIndex: undefined }, { zone: 'foundation', col: 3 }],
    });
  });

  it('rejects a move with too few arguments', () => {
    const result = parseRankandfileCommand('m t0');
    expect(result).toHaveProperty('error');
  });

  it('rejects an invalid source zone', () => {
    expect(parseRankandfileCommand('m x0 t1')).toHaveProperty('error');
    expect(parseRankandfileCommand('m tx t1')).toHaveProperty('error');
  });

  it('rejects a bare or malformed source column index', () => {
    // 'm t t1' must not silently resolve to column 0 (Number('') === 0).
    expect(parseRankandfileCommand('m t t1')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t-1 t1')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t1.5 t2')).toHaveProperty('error');
  });

  it('rejects an invalid target zone', () => {
    expect(parseRankandfileCommand('m t0 x1')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t0 tx')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t0 fx')).toHaveProperty('error');
  });

  it('rejects a bare or malformed target index', () => {
    expect(parseRankandfileCommand('m t0 t')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t0 t-1')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t0 f-2')).toHaveProperty('error');
    expect(parseRankandfileCommand('m t0 f1.5')).toHaveProperty('error');
  });

  it('suggests a close command for typos', () => {
    const result = parseRankandfileCommand('mvoe');
    expect(result).toHaveProperty('error');
    expect((result as { error: string }).error).toContain('Unknown command');
  });

  it('rejects entirely unknown commands', () => {
    expect(parseRankandfileCommand('zzzzz')).toHaveProperty('error');
  });
});

describe('RANKANDFILE_HELP', () => {
  it('documents the draw, move, and reset commands', () => {
    const joined = RANKANDFILE_HELP.join('\n');
    expect(joined).toContain('d/draw');
    expect(joined).toContain('m t<c> t<c>');
    expect(joined).toContain('m w t<c>');
    expect(joined).toContain('r/reset');
  });
});
