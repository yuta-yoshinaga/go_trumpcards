import { describe, expect, it } from 'vitest';
import { parsePineappleCommand } from './pineappleCommands';

describe('parsePineappleCommand', () => {
  it('parses fold', () => {
    expect(parsePineappleCommand('f')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parsePineappleCommand('ck')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parsePineappleCommand('c')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parsePineappleCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('returns error for bet without amount', () => {
    const result = parsePineappleCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parsePineappleCommand('ra 200')).toEqual({ args: ['raise', 200] });
  });

  it('parses allin', () => {
    expect(parsePineappleCommand('a')).toEqual({ args: ['allin', undefined] });
  });

  it('parses rebuy', () => {
    expect(parsePineappleCommand('rb')).toEqual({ args: ['rebuy', undefined] });
    expect(parsePineappleCommand('rebuy')).toEqual({ args: ['rebuy', undefined] });
  });

  it('parses skiprebuy', () => {
    expect(parsePineappleCommand('sr')).toEqual({ args: ['skiprebuy', undefined] });
    expect(parsePineappleCommand('skiprebuy')).toEqual({ args: ['skiprebuy', undefined] });
  });

  it('parses addon', () => {
    expect(parsePineappleCommand('ao')).toEqual({ args: ['addon', undefined] });
    expect(parsePineappleCommand('addon')).toEqual({ args: ['addon', undefined] });
  });

  it('parses skipaddon', () => {
    expect(parsePineappleCommand('sa')).toEqual({ args: ['skipaddon', undefined] });
    expect(parsePineappleCommand('skipaddon')).toEqual({ args: ['skipaddon', undefined] });
  });

  it('parses muck', () => {
    expect(parsePineappleCommand('mu')).toEqual({ args: ['muck', undefined] });
    expect(parsePineappleCommand('muck')).toEqual({ args: ['muck', undefined] });
  });

  it('parses show', () => {
    expect(parsePineappleCommand('sh')).toEqual({ args: ['show', undefined] });
    expect(parsePineappleCommand('show')).toEqual({ args: ['show', undefined] });
  });

  it('parses discard with index', () => {
    expect(parsePineappleCommand('dis 2')).toEqual({ args: ['discard', undefined, { cardIdx: 2 }] });
    expect(parsePineappleCommand('discard 4')).toEqual({ args: ['discard', undefined, { cardIdx: 4 }] });
  });

  it('returns error for discard without index', () => {
    const result = parsePineappleCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses reset', () => {
    expect(parsePineappleCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parsePineappleCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
