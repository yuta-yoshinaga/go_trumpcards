import { describe, expect, it } from 'vitest';
import { parseOldmaidCommand } from './oldmaidCommands';

describe('parseOldmaidCommand', () => {
  it('parses draw with index', () => {
    expect(parseOldmaidCommand('d 2')).toEqual({ args: ['draw', 2] });
    expect(parseOldmaidCommand('draw 5')).toEqual({ args: ['draw', 5] });
  });

  it('returns error for draw without index', () => {
    const result = parseOldmaidCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses shuffle', () => {
    expect(parseOldmaidCommand('sh')).toEqual({ args: ['shuffle'] });
    expect(parseOldmaidCommand('shuffle')).toEqual({ args: ['shuffle'] });
  });

  it('parses reorder with indices', () => {
    expect(parseOldmaidCommand('ro 2 0 1 3')).toEqual({
      args: ['reorder', undefined, undefined, undefined, [2, 0, 1, 3]],
    });
    expect(parseOldmaidCommand('reorder 1 0')).toEqual({ args: ['reorder', undefined, undefined, undefined, [1, 0]] });
  });

  it('parses reorder without indices as empty', () => {
    expect(parseOldmaidCommand('ro')).toEqual({ args: ['reorder', undefined, undefined, undefined, []] });
  });

  it('parses reset', () => {
    expect(parseOldmaidCommand('r')).toEqual({ args: ['reset'] });
    expect(parseOldmaidCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseOldmaidCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
