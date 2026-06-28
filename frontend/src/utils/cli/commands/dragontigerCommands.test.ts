import { describe, expect, it } from 'vitest';
import { DragonTigerBetType } from '../../../types/phases';
import { DRAGONTIGER_CLI_HELP, parseDragonTigerCommand } from './dragontigerCommands';

describe('parseDragonTigerCommand', () => {
  it('parses bet with an explicit target and amount', () => {
    expect(parseDragonTigerCommand('bet dragon 100')).toEqual({
      args: ['bet', 100, DragonTigerBetType.DRAGON],
    });
    expect(parseDragonTigerCommand('bet tiger 50')).toEqual({
      args: ['bet', 50, DragonTigerBetType.TIGER],
    });
    expect(parseDragonTigerCommand('bet tie 25')).toEqual({
      args: ['bet', 25, DragonTigerBetType.TIE],
    });
  });

  it('parses the d/t/e shortcuts', () => {
    expect(parseDragonTigerCommand('d 100')).toEqual({ args: ['bet', 100, DragonTigerBetType.DRAGON] });
    expect(parseDragonTigerCommand('t 100')).toEqual({ args: ['bet', 100, DragonTigerBetType.TIGER] });
    expect(parseDragonTigerCommand('e 100')).toEqual({ args: ['bet', 100, DragonTigerBetType.TIE] });
  });

  it('returns error for bet without a valid target', () => {
    expect('error' in parseDragonTigerCommand('bet 100')).toBe(true);
    expect('error' in parseDragonTigerCommand('bet sideways 100')).toBe(true);
  });

  it('returns error for a non-numeric amount', () => {
    expect('error' in parseDragonTigerCommand('bet dragon abc')).toBe(true);
    expect('error' in parseDragonTigerCommand('d')).toBe(true);
  });

  it('parses clear', () => {
    expect(parseDragonTigerCommand('clear')).toEqual({ args: ['clear'] });
  });

  it('parses log', () => {
    expect(parseDragonTigerCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseDragonTigerCommand('r')).toEqual({ args: ['reset'] });
    expect(parseDragonTigerCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseDragonTigerCommand('rese');
    expect('error' in result && result.error).toContain('reset');
  });

  it('returns error for an unknown command', () => {
    expect('error' in parseDragonTigerCommand('xyz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(DRAGONTIGER_CLI_HELP.length).toBeGreaterThan(0);
    expect(DRAGONTIGER_CLI_HELP.some((line) => line.includes('bet'))).toBe(true);
  });
});
