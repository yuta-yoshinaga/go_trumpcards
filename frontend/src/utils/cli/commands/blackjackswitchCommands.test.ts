import { describe, expect, it } from 'vitest';
import { parseBlackjackSwitchCommand } from './blackjackswitchCommands';

describe('parseBlackjackSwitchCommand', () => {
  it('parses bet with an amount', () => {
    expect(parseBlackjackSwitchCommand('bet 100')).toEqual({ args: ['bet', 100] });
    expect(parseBlackjackSwitchCommand('b 250')).toEqual({ args: ['bet', 250] });
  });

  it('errors on bet without a valid amount', () => {
    expect('error' in parseBlackjackSwitchCommand('bet')).toBe(true);
    expect('error' in parseBlackjackSwitchCommand('bet abc')).toBe(true);
  });

  it('parses switch and keep', () => {
    expect(parseBlackjackSwitchCommand('sw')).toEqual({ args: ['switch'] });
    expect(parseBlackjackSwitchCommand('switch')).toEqual({ args: ['switch'] });
    expect(parseBlackjackSwitchCommand('k')).toEqual({ args: ['keep'] });
    expect(parseBlackjackSwitchCommand('keep')).toEqual({ args: ['keep'] });
  });

  it('parses hit, stand, and double down', () => {
    expect(parseBlackjackSwitchCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseBlackjackSwitchCommand('s')).toEqual({ args: ['stand'] });
    expect(parseBlackjackSwitchCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseBlackjackSwitchCommand('dd')).toEqual({ args: ['doubledown'] });
    expect(parseBlackjackSwitchCommand('doubledown')).toEqual({ args: ['doubledown'] });
  });

  it('parses log and reset', () => {
    expect(parseBlackjackSwitchCommand('log')).toEqual({ args: ['log'] });
    expect(parseBlackjackSwitchCommand('l')).toEqual({ args: ['log'] });
    expect(parseBlackjackSwitchCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBlackjackSwitchCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseBlackjackSwitchCommand('stan');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('stand');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseBlackjackSwitchCommand('xyz')).toBe(true);
  });
});
