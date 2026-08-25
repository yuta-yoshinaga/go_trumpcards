import { describe, expect, it } from 'vitest';
import { parseWhistCommand, WHIST_HELP } from './whistCommands';

describe('parseWhistCommand', () => {
  // **札のインデックスは 2 番目のスロット。** ここが 1 番目だったころは、
  // クライアントもそう読んでいたので試験と実装が互いに一致したまま
  // サーバには cardIndex が届いていなかった (#6227)。
  it('parses play with the card index in the second slot', () => {
    expect(parseWhistCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseWhistCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    const result = parseWhistCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseWhistCommand('n')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseWhistCommand('next')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseWhistCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
    expect(parseWhistCommand('nextround')).toEqual({
      args: ['nextround', undefined, undefined],
    });
  });

  it('parses hint and reset', () => {
    expect(parseWhistCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseWhistCommand('hint')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseWhistCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseWhistCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseWhistCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help text', () => {
    expect(WHIST_HELP.length).toBeGreaterThan(0);
  });
});
