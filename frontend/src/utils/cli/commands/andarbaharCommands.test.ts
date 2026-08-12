import { describe, expect, it } from 'vitest';
import { AndarBaharColumn, AndarBaharSideBand } from '../../../types/phases';
import { ANDARBAHAR_CLI_HELP, parseAndarBaharCommand } from './andarbaharCommands';

describe('parseAndarBaharCommand', () => {
  it('parses a bet on either column', () => {
    expect(parseAndarBaharCommand('bet andar 100')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE],
    });
    expect(parseAndarBaharCommand('b a 100')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE],
    });
    expect(parseAndarBaharCommand('bet bahar 50')).toEqual({
      args: ['bet', 50, AndarBaharColumn.BAHAR, 0, AndarBaharSideBand.NONE],
    });
    expect(parseAndarBaharCommand('b b 50')).toEqual({
      args: ['bet', 50, AndarBaharColumn.BAHAR, 0, AndarBaharSideBand.NONE],
    });
  });

  it('accepts the column shortcuts', () => {
    expect(parseAndarBaharCommand('andar 100')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE],
    });
    expect(parseAndarBaharCommand('bahar 100')).toEqual({
      args: ['bet', 100, AndarBaharColumn.BAHAR, 0, AndarBaharSideBand.NONE],
    });
  });

  it('parses a side bet only when both stake and band are given', () => {
    expect(parseAndarBaharCommand('bet andar 100 50 2')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 50, AndarBaharSideBand.SIX_TO_TEN],
    });
    // **帯だけでは賭けない。** band 0 は有効な値なので、金額の無い帯を送ると
    // サーバに賭け金 0 のサイドベットとして拒否されます。
    expect(parseAndarBaharCommand('bet andar 100 50')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE],
    });
  });

  it('rejects an out-of-range side band', () => {
    expect(parseAndarBaharCommand('bet andar 100 50 7')).toEqual({ error: 'Side bet band must be 0-6' });
    expect(parseAndarBaharCommand('bet andar 100 50 -1')).toEqual({ error: 'Side bet band must be 0-6' });
    // 境界は通す。
    expect(parseAndarBaharCommand('bet andar 100 50 0')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 50, AndarBaharSideBand.FIRST],
    });
    expect(parseAndarBaharCommand('bet andar 100 50 6')).toEqual({
      args: ['bet', 100, AndarBaharColumn.ANDAR, 50, AndarBaharSideBand.THIRTYSIX_PLUS],
    });
  });

  it('rejects a missing or unparsable amount and an unknown column', () => {
    expect(parseAndarBaharCommand('bet andar')).toHaveProperty('error');
    expect(parseAndarBaharCommand('bet andar abc')).toHaveProperty('error');
    expect(parseAndarBaharCommand('bet sideways 100')).toHaveProperty('error');
    expect(parseAndarBaharCommand('bet andar 100 abc 2')).toHaveProperty('error');
    expect(parseAndarBaharCommand('bet andar 100 50 abc')).toHaveProperty('error');
  });

  it('parses the simple commands', () => {
    expect(parseAndarBaharCommand('clear')).toEqual({ args: ['clear'] });
    expect(parseAndarBaharCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseAndarBaharCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAndarBaharCommand('log')).toEqual({ args: ['log'] });
    expect(parseAndarBaharCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseAndarBaharCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near-miss and rejects nonsense', () => {
    const near = parseAndarBaharCommand('bett');
    expect(near).toHaveProperty('error');
    expect((near as { error: string }).error).toContain('Did you mean');
    expect(parseAndarBaharCommand('zzzz')).toHaveProperty('error');
  });

  it('documents every command it accepts', () => {
    expect(ANDARBAHAR_CLI_HELP.length).toBeGreaterThanOrEqual(5);
    const help = ANDARBAHAR_CLI_HELP.join('\n');
    for (const token of ['bet', 'clear', 'hint', 'log', 'reset']) {
      expect(help).toContain(token);
    }
  });
});
