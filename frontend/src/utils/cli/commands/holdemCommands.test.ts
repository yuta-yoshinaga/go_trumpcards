import { describe, expect, it } from 'vitest';
import { parseHoldemCommand } from './holdemCommands';

describe('parseHoldemCommand', () => {
  it('parses fold', () => {
    expect(parseHoldemCommand('f')).toEqual({ args: ['fold', undefined] });
    expect(parseHoldemCommand('fold')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parseHoldemCommand('ck')).toEqual({ args: ['check', undefined] });
    expect(parseHoldemCommand('check')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parseHoldemCommand('c')).toEqual({ args: ['call', undefined] });
    expect(parseHoldemCommand('call')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parseHoldemCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseHoldemCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parseHoldemCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseHoldemCommand('ra 200')).toEqual({ args: ['raise', 200] });
    expect(parseHoldemCommand('raise 300')).toEqual({ args: ['raise', 300] });
  });

  it('parses allin', () => {
    expect(parseHoldemCommand('a')).toEqual({ args: ['allin', undefined] });
    expect(parseHoldemCommand('allin')).toEqual({ args: ['allin', undefined] });
  });

  it('parses rebuy', () => {
    expect(parseHoldemCommand('rb')).toEqual({ args: ['rebuy', undefined] });
    expect(parseHoldemCommand('rebuy')).toEqual({ args: ['rebuy', undefined] });
  });

  it('parses skiprebuy', () => {
    expect(parseHoldemCommand('sr')).toEqual({ args: ['skiprebuy', undefined] });
    expect(parseHoldemCommand('skiprebuy')).toEqual({ args: ['skiprebuy', undefined] });
  });

  it('parses addon', () => {
    expect(parseHoldemCommand('ao')).toEqual({ args: ['addon', undefined] });
    expect(parseHoldemCommand('addon')).toEqual({ args: ['addon', undefined] });
  });

  it('parses skipaddon', () => {
    expect(parseHoldemCommand('sa')).toEqual({ args: ['skipaddon', undefined] });
    expect(parseHoldemCommand('skipaddon')).toEqual({ args: ['skipaddon', undefined] });
  });

  it('parses muck', () => {
    expect(parseHoldemCommand('mu')).toEqual({ args: ['muck', undefined] });
    expect(parseHoldemCommand('muck')).toEqual({ args: ['muck', undefined] });
  });

  it('parses show', () => {
    expect(parseHoldemCommand('sh')).toEqual({ args: ['show', undefined] });
    expect(parseHoldemCommand('show')).toEqual({ args: ['show', undefined] });
  });

  it('parses reset', () => {
    expect(parseHoldemCommand('r')).toEqual({ args: ['reset', undefined] });
    expect(parseHoldemCommand('reset')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseHoldemCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
