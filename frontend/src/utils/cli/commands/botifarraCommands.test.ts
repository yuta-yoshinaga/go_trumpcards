import { describe, expect, it } from 'vitest';
import { BOTIFARRA_NO_TRUMP } from '../../../types/games/botifarra';
import { BOTIFARRA_CLI_HELP, parseBotifarraCommand } from './botifarraCommands';

describe('parseBotifarraCommand', () => {
  it('parses a play by hand index', () => {
    expect(parseBotifarraCommand('play 3')).toEqual({ args: ['play', 3] });
    expect(parseBotifarraCommand('p 0')).toEqual({ args: ['play', 0] });
    expect(parseBotifarraCommand('p')).toHaveProperty('error');
    expect(parseBotifarraCommand('p abc')).toHaveProperty('error');
  });

  // **切り札なしは -1 という有効な値。** 引数が無いのとは違います。
  it('parses every trump including no trump', () => {
    expect(parseBotifarraCommand('declare s')).toEqual({ args: ['declare', undefined, 1] });
    expect(parseBotifarraCommand('declare spade')).toEqual({ args: ['declare', undefined, 1] });
    expect(parseBotifarraCommand('declare c')).toEqual({ args: ['declare', undefined, 2] });
    expect(parseBotifarraCommand('declare club')).toEqual({ args: ['declare', undefined, 2] });
    expect(parseBotifarraCommand('declare h')).toEqual({ args: ['declare', undefined, 3] });
    expect(parseBotifarraCommand('declare d')).toEqual({ args: ['declare', undefined, 4] });
    expect(parseBotifarraCommand('declare none')).toEqual({
      args: ['declare', undefined, BOTIFARRA_NO_TRUMP],
    });
    expect(parseBotifarraCommand('declare n')).toEqual({ args: ['declare', undefined, BOTIFARRA_NO_TRUMP] });
  });

  // **引数無しが「スペード」に化けない。** スートは 1..4 なので 0 は無効です。
  it('rejects a missing or unknown trump rather than defaulting', () => {
    expect(parseBotifarraCommand('declare')).toHaveProperty('error');
    expect(parseBotifarraCommand('declare x')).toHaveProperty('error');
  });

  it('parses the simple commands', () => {
    expect(parseBotifarraCommand('delegate')).toEqual({ args: ['delegate'] });
    expect(parseBotifarraCommand('double')).toEqual({ args: ['double'] });
    expect(parseBotifarraCommand('pass')).toEqual({ args: ['passdouble'] });
    expect(parseBotifarraCommand('next')).toEqual({ args: ['next'] });
    expect(parseBotifarraCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseBotifarraCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBotifarraCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBotifarraCommand('log')).toEqual({ args: ['log'] });
    expect(parseBotifarraCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseBotifarraCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near-miss and rejects nonsense', () => {
    const near = parseBotifarraCommand('declar');
    expect(near).toHaveProperty('error');
    expect((near as { error: string }).error).toContain('Did you mean');
    expect(parseBotifarraCommand('zzzz')).toHaveProperty('error');
  });

  it('documents every command it accepts', () => {
    expect(BOTIFARRA_CLI_HELP.length).toBeGreaterThanOrEqual(8);
    const help = BOTIFARRA_CLI_HELP.join('\n');
    for (const token of ['play', 'declare', 'delegate', 'double', 'pass', 'next', 'giveup', 'hint', 'log']) {
      expect(help).toContain(token);
    }
  });
});
