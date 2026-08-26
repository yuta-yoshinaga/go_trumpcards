import { describe, expect, it } from 'vitest';
import { CATCHTEN_HELP, parseCatchTenCommand } from './catchtenCommands';

describe('parseCatchTenCommand', () => {
  // **札のインデックスは 2 番目のスロット。** ここが 1 番目だったころは、
  // クライアントもそう読んでいたので試験と実装が互いに一致したまま
  // サーバには cardIndex が届いていなかった (#6227)。
  it('parses play with the card index in the second slot', () => {
    expect(parseCatchTenCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseCatchTenCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    const result = parseCatchTenCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseCatchTenCommand('n')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseCatchTenCommand('next')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseCatchTenCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
    expect(parseCatchTenCommand('nextround')).toEqual({
      args: ['nextround', undefined, undefined],
    });
  });

  it('parses hint and reset', () => {
    expect(parseCatchTenCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseCatchTenCommand('hint')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseCatchTenCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseCatchTenCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseCatchTenCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help text', () => {
    expect(CATCHTEN_HELP.length).toBeGreaterThan(0);
  });
});
