import { describe, expect, it } from 'vitest';
import { RikkenContract } from '../../../types/phases';
import { parseRikkenCommand, RIKKEN_CLI_HELP } from './rikkenCommands';

describe('parseRikkenCommand', () => {
  it('parses a play by hand index', () => {
    expect(parseRikkenCommand('play 5')).toEqual({ args: ['play', 5] });
    expect(parseRikkenCommand('p 0')).toEqual({ args: ['play', 0] });
    expect(parseRikkenCommand('p')).toHaveProperty('error');
    expect(parseRikkenCommand('p abc')).toHaveProperty('error');
  });

  it('parses every contract on the ladder', () => {
    expect(parseRikkenCommand('bid rik')).toEqual({ args: ['bid', undefined, RikkenContract.RIK] });
    expect(parseRikkenCommand('bid misere')).toEqual({ args: ['bid', undefined, RikkenContract.MISERE] });
    expect(parseRikkenCommand('bid mis')).toEqual({ args: ['bid', undefined, RikkenContract.MISERE] });
    expect(parseRikkenCommand('bid solo')).toEqual({ args: ['bid', undefined, RikkenContract.SOLO] });
    expect(parseRikkenCommand('bid open')).toEqual({ args: ['bid', undefined, RikkenContract.OPEN_MISERE] });
    expect(parseRikkenCommand('bid openmisere')).toEqual({
      args: ['bid', undefined, RikkenContract.OPEN_MISERE],
    });
  });

  // **パスは契約 0。** 別経路にせず、同じ値として送ります。
  it('sends pass as contract 0', () => {
    expect(parseRikkenCommand('pass')).toEqual({ args: ['bid', undefined, RikkenContract.NONE] });
    expect(parseRikkenCommand('bid pass')).toEqual({ args: ['bid', undefined, RikkenContract.NONE] });
  });

  it('rejects a missing or unknown contract rather than defaulting', () => {
    expect(parseRikkenCommand('bid')).toHaveProperty('error');
    expect(parseRikkenCommand('bid nonsense')).toHaveProperty('error');
  });

  it('parses every trump and rejects a missing one', () => {
    expect(parseRikkenCommand('call s')).toEqual({ args: ['call', undefined, undefined, 1] });
    expect(parseRikkenCommand('call clover')).toEqual({ args: ['call', undefined, undefined, 2] });
    expect(parseRikkenCommand('call h')).toEqual({ args: ['call', undefined, undefined, 3] });
    expect(parseRikkenCommand('call diamond')).toEqual({ args: ['call', undefined, undefined, 4] });
    expect(parseRikkenCommand('call')).toHaveProperty('error');
    expect(parseRikkenCommand('call x')).toHaveProperty('error');
  });

  it('parses the simple commands', () => {
    expect(parseRikkenCommand('next')).toEqual({ args: ['next'] });
    expect(parseRikkenCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseRikkenCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseRikkenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseRikkenCommand('log')).toEqual({ args: ['log'] });
    expect(parseRikkenCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseRikkenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near-miss and rejects nonsense', () => {
    const near = parseRikkenCommand('bidd');
    expect(near).toHaveProperty('error');
    expect((near as { error: string }).error).toContain('Did you mean');
    expect(parseRikkenCommand('zzzz')).toHaveProperty('error');
  });

  it('documents every command it accepts', () => {
    expect(RIKKEN_CLI_HELP.length).toBeGreaterThanOrEqual(8);
    const help = RIKKEN_CLI_HELP.join('\n');
    for (const token of ['play', 'bid', 'pass', 'call', 'next', 'giveup', 'hint', 'log']) {
      expect(help).toContain(token);
    }
  });
});
