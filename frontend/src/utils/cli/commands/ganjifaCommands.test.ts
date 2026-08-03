import { describe, expect, it } from 'vitest';
import { GANJIFA_HELP, parseGanjifaCommand } from './ganjifaCommands';

describe('parseGanjifaCommand', () => {
  it('parses play with an index', () => {
    expect(parseGanjifaCommand('play 4')).toEqual({ args: ['play', { cardIndex: 4 }] });
    expect(parseGanjifaCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a numeric index', () => {
    expect(parseGanjifaCommand('play')).toEqual({ error: 'Usage: p <idx>' });
    expect(parseGanjifaCommand('play x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it('parses trick and round advancement', () => {
    expect(parseGanjifaCommand('n')).toEqual({ args: ['next'] });
    expect(parseGanjifaCommand('next')).toEqual({ args: ['next'] });
    expect(parseGanjifaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGanjifaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses setdifficulty as a reset carrying the new level', () => {
    expect(parseGanjifaCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range difficulty', () => {
    for (const bad of ['sd 3', 'sd -1', 'sd x']) {
      expect(parseGanjifaCommand(bad)).toEqual({ error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' });
    }
  });

  it('parses hint, log and reset', () => {
    expect(parseGanjifaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGanjifaCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseGanjifaCommand('l')).toEqual({ args: ['log'] });
    expect(parseGanjifaCommand('log')).toEqual({ args: ['log'] });
    expect(parseGanjifaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGanjifaCommand('reset')).toEqual({ args: ['reset'] });
  });

  // Ganjifa has no bidding phase, so these must be rejected outright rather than
  // silently mapping onto some other action.
  it('rejects bidding commands', () => {
    for (const cmd of ['bid 1', 'pass']) {
      const result = parseGanjifaCommand(cmd);
      // Assert the shape first: `if ('error' in result)` alone would skip the
      // body — and pass — precisely when the command was wrongly accepted.
      expect('error' in result).toBe(true);
    }
  });

  it('suggests a near miss', () => {
    const result = parseGanjifaCommand('nexr');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('reports a bare unknown command', () => {
    expect(parseGanjifaCommand('zzzzz')).toEqual({ error: 'Unknown command: zzzzz' });
  });

  it('documents every accepted command in the help text', () => {
    const help = GANJIFA_HELP.join('\n');
    for (const cmd of ['p <idx>', 'n / next', 'nr / nextround', 'sd <0-2>', 'h / hint', 'l / log', 'r / reset']) {
      expect(help).toContain(cmd);
    }
    // The help must not advertise an action the parser rejects.
    expect(help).not.toContain('bid');
    expect(help).not.toContain('pass');
  });
});
