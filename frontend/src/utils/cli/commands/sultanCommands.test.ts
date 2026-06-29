import { describe, expect, it } from 'vitest';
import { parseSultanCommand, SULTAN_HELP } from './sultanCommands';

describe('parseSultanCommand', () => {
  it.each([
    ['d', ['draw']],
    ['draw', ['draw']],
    ['rd', ['redeal']],
    ['redeal', ['redeal']],
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
    expect(parseSultanCommand(input)).toEqual({ args: expected });
  });

  it('parses divan-to-foundation moves (m d <idx>)', () => {
    expect(parseSultanCommand('m d 3')).toEqual({
      args: ['move', { zone: 'divan', divanIdx: 3 }],
    });
  });

  it('parses divan-to-foundation shorthand (m <idx>)', () => {
    expect(parseSultanCommand('m 5')).toEqual({
      args: ['move', { zone: 'divan', divanIdx: 5 }],
    });
  });

  it('parses waste-to-foundation moves (m w)', () => {
    expect(parseSultanCommand('m w')).toEqual({
      args: ['move', { zone: 'waste' }],
    });
  });

  it('rejects a move with no arguments', () => {
    expect(parseSultanCommand('m')).toHaveProperty('error');
  });

  it('rejects a divan move with a missing or malformed index', () => {
    expect(parseSultanCommand('m d')).toHaveProperty('error');
    expect(parseSultanCommand('m d x')).toHaveProperty('error');
    expect(parseSultanCommand('m d -1')).toHaveProperty('error');
    expect(parseSultanCommand('m d 1.5')).toHaveProperty('error');
  });

  it('rejects an invalid source token', () => {
    expect(parseSultanCommand('m x')).toHaveProperty('error');
    expect(parseSultanCommand('m t0')).toHaveProperty('error');
  });

  it('suggests a close command for typos', () => {
    const result = parseSultanCommand('mvoe');
    expect(result).toHaveProperty('error');
    expect((result as { error: string }).error).toContain('Unknown command');
  });

  it('rejects entirely unknown commands', () => {
    expect(parseSultanCommand('zzzzz')).toHaveProperty('error');
  });
});

describe('SULTAN_HELP', () => {
  it('documents the draw, redeal, move, and reset commands', () => {
    const joined = SULTAN_HELP.join('\n');
    expect(joined).toContain('d/draw');
    expect(joined).toContain('rd/redeal');
    expect(joined).toContain('m d <idx>');
    expect(joined).toContain('m w');
    expect(joined).toContain('r/reset');
  });
});
