import { describe, expect, it } from 'vitest';
import { parseBlackjackCommand } from './blackjackCommands';

describe('parseBlackjackCommand', () => {
  it('parses hit', () => {
    expect(parseBlackjackCommand('h')).toEqual({ args: ['hit'] });
    expect(parseBlackjackCommand('hit')).toEqual({ args: ['hit'] });
  });

  it('parses stand', () => {
    expect(parseBlackjackCommand('s')).toEqual({ args: ['stand'] });
    expect(parseBlackjackCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses doubledown', () => {
    expect(parseBlackjackCommand('d')).toEqual({ args: ['doubledown'] });
  });

  it('parses split', () => {
    expect(parseBlackjackCommand('sp')).toEqual({ args: ['split'] });
  });

  it('parses insurance', () => {
    expect(parseBlackjackCommand('i')).toEqual({ args: ['insurance'] });
    expect(parseBlackjackCommand('di')).toEqual({ args: ['declineinsurance'] });
  });

  it('parses surrender', () => {
    expect(parseBlackjackCommand('sur')).toEqual({ args: ['surrender'] });
  });

  it('parses early surrender', () => {
    expect(parseBlackjackCommand('es')).toEqual({ args: ['earlysurrender'] });
    expect(parseBlackjackCommand('des')).toEqual({ args: ['declineearlysurrender'] });
  });

  it('parses bet with amount', () => {
    expect(parseBlackjackCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseBlackjackCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parseBlackjackCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses toggle commands', () => {
    expect(parseBlackjackCommand('hint')).toEqual({ args: ['togglehint'] });
    expect(parseBlackjackCommand('soft17')).toEqual({ args: ['togglesoft17'] });
    expect(parseBlackjackCommand('counting')).toEqual({ args: ['togglecounting'] });
    expect(parseBlackjackCommand('das')).toEqual({ args: ['toggledas'] });
  });

  it('parses set commands with argument', () => {
    expect(parseBlackjackCommand('sd 6')).toEqual({ args: ['setdeckcount', 6] });
    expect(parseBlackjackCommand('scc 2')).toEqual({ args: ['setcpucount', 2] });
    expect(parseBlackjackCommand('scs 1')).toEqual({ args: ['setcountingsystem', 1] });
    expect(parseBlackjackCommand('pen 75')).toEqual({ args: ['setpenetration', 75] });
    expect(parseBlackjackCommand('ssr 1')).toEqual({ args: ['setsurrenderrule', 1] });
  });

  it('parses reset', () => {
    expect(parseBlackjackCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBlackjackCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for typo', () => {
    const result = parseBlackjackCommand('hti');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error for unknown command', () => {
    const result = parseBlackjackCommand('zzzzzzz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Unknown command');
    }
  });

  it('is case insensitive', () => {
    expect(parseBlackjackCommand('HIT')).toEqual({ args: ['hit'] });
    expect(parseBlackjackCommand('B 100')).toEqual({ args: ['bet', 100] });
  });
});
