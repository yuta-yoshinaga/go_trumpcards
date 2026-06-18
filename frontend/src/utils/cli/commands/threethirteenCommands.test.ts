import { describe, expect, it } from 'vitest';
import { parseThreeThirteenCommand } from './threethirteenCommands';

describe('parseThreeThirteenCommand', () => {
  it('parses drawstock', () => {
    expect(parseThreeThirteenCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseThreeThirteenCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseThreeThirteenCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseThreeThirteenCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with index', () => {
    expect(parseThreeThirteenCommand('d 3')).toEqual({ args: ['discard', 3] });
    expect(parseThreeThirteenCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseThreeThirteenCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses knock with index', () => {
    expect(parseThreeThirteenCommand('k 2')).toEqual({ args: ['knock', 2] });
    expect(parseThreeThirteenCommand('knock 4')).toEqual({ args: ['knock', 4] });
  });

  it('returns error for knock without index', () => {
    const result = parseThreeThirteenCommand('k');
    expect('error' in result).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseThreeThirteenCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseThreeThirteenCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseThreeThirteenCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseThreeThirteenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseThreeThirteenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseThreeThirteenCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
