import { describe, expect, it } from 'vitest';
import { ColourWhistContract } from '../../../types/phases';
import { COLOURWHIST_CLI_HELP, parseColourWhistCommand } from './colourwhistCommands';

describe('parseColourWhistCommand', () => {
  it('parses a play by hand index', () => {
    expect(parseColourWhistCommand('play 7')).toEqual({ args: ['play', 7] });
    expect(parseColourWhistCommand('p 0')).toEqual({ args: ['play', 0] });
    expect(parseColourWhistCommand('p')).toHaveProperty('error');
    expect(parseColourWhistCommand('p abc')).toHaveProperty('error');
  });

  it('parses the three biddable contracts', () => {
    expect(parseColourWhistCommand('bid samen')).toEqual({
      args: ['bid', undefined, ColourWhistContract.SAMEN],
    });
    expect(parseColourWhistCommand('bid alleen')).toEqual({
      args: ['bid', undefined, ColourWhistContract.ALLEEN],
    });
    expect(parseColourWhistCommand('bid miserie')).toEqual({
      args: ['bid', undefined, ColourWhistContract.MISERIE],
    });
    expect(parseColourWhistCommand('bid mis')).toEqual({
      args: ['bid', undefined, ColourWhistContract.MISERIE],
    });
  });

  // **troel は競れない。** 配りでしか成立しないので語彙に入れません。
  it('refuses troel with an explanation', () => {
    const out = parseColourWhistCommand('bid troel');
    expect(out).toHaveProperty('error');
    expect((out as { error: string }).error).toContain('troel is dealt, not bid');
  });

  it('sends pass as contract 0', () => {
    expect(parseColourWhistCommand('pass')).toEqual({ args: ['bid', undefined, ColourWhistContract.NONE] });
    expect(parseColourWhistCommand('bid pass')).toEqual({
      args: ['bid', undefined, ColourWhistContract.NONE],
    });
  });

  it('rejects a missing or unknown contract rather than defaulting', () => {
    expect(parseColourWhistCommand('bid')).toHaveProperty('error');
    expect(parseColourWhistCommand('bid nonsense')).toHaveProperty('error');
  });

  it('parses every trump and rejects a missing one', () => {
    expect(parseColourWhistCommand('call s')).toEqual({ args: ['call', undefined, undefined, 1] });
    expect(parseColourWhistCommand('call clover')).toEqual({ args: ['call', undefined, undefined, 2] });
    expect(parseColourWhistCommand('call h')).toEqual({ args: ['call', undefined, undefined, 3] });
    expect(parseColourWhistCommand('call diamond')).toEqual({ args: ['call', undefined, undefined, 4] });
    expect(parseColourWhistCommand('call')).toHaveProperty('error');
    expect(parseColourWhistCommand('call x')).toHaveProperty('error');
  });

  it('parses the simple commands', () => {
    expect(parseColourWhistCommand('next')).toEqual({ args: ['next'] });
    expect(parseColourWhistCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseColourWhistCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseColourWhistCommand('log')).toEqual({ args: ['log'] });
    expect(parseColourWhistCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near-miss and rejects nonsense', () => {
    const near = parseColourWhistCommand('bidd');
    expect(near).toHaveProperty('error');
    expect((near as { error: string }).error).toContain('Did you mean');
    expect(parseColourWhistCommand('zzzz')).toHaveProperty('error');
  });

  it('documents every command it accepts', () => {
    expect(COLOURWHIST_CLI_HELP.length).toBeGreaterThanOrEqual(8);
    const help = COLOURWHIST_CLI_HELP.join('\n');
    for (const token of ['play', 'bid', 'pass', 'call', 'next', 'giveup', 'hint', 'log']) {
      expect(help).toContain(token);
    }
    // ヘルプでも troel が競れないことを説明する。
    expect(help).toContain('troel is dealt, not bid');
  });
});
