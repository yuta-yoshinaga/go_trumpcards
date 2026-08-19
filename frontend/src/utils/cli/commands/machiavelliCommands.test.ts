import { describe, expect, it } from 'vitest';
import { parseMachiavelliCommand } from './machiavelliCommands';

describe('parseMachiavelliCommand', () => {
  it('parses draw', () => {
    expect(parseMachiavelliCommand('dr')).toEqual({ args: ['draw'] });
    expect(parseMachiavelliCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses newmeld with 3+ indices', () => {
    expect(parseMachiavelliCommand('nm 0 1 2')).toEqual({ args: ['newmeld', { handIndices: [0, 1, 2] }] });
    expect(parseMachiavelliCommand('newmeld 3 4 5 6')).toEqual({ args: ['newmeld', { handIndices: [3, 4, 5, 6] }] });
  });

  it('returns error for newmeld with fewer than 3 indices', () => {
    expect('error' in parseMachiavelliCommand('nm 0 1')).toBe(true);
  });

  it('returns error for newmeld with a non-numeric index', () => {
    expect('error' in parseMachiavelliCommand('nm 0 x 2')).toBe(true);
  });

  it('parses layoff with meld and hand indices', () => {
    expect(parseMachiavelliCommand('lo 1 4')).toEqual({ args: ['layoff', { meldIdx: 1, handIndex: 4 }] });
    expect(parseMachiavelliCommand('layoff 0 2')).toEqual({ args: ['layoff', { meldIdx: 0, handIndex: 2 }] });
  });

  it('returns error for layoff without both indices', () => {
    expect('error' in parseMachiavelliCommand('lo 1')).toBe(true);
    expect('error' in parseMachiavelliCommand('lo')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseMachiavelliCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMachiavelliCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseMachiavelliCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseMachiavelliCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMachiavelliCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseMachiavelliCommand('drw');
    expect('error' in result && result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMachiavelliCommand('xyz')).toBe(true);
  });

  // #5704: 場の組み替えは CUI にも Web CLI にも無かった。書式は Go 側の
  // parseMachiavelliRearrange と同じでなければならない。
  it('parses rearrange into a play with the whole new table', () => {
    expect(parseMachiavelliCommand('ra s5,h5,d5;c7,c8,c9 / 2,4')).toEqual({
      args: [
        'play',
        {
          tableMelds: [
            [
              { design: 1, value: 5 },
              { design: 3, value: 5 },
              { design: 4, value: 5 },
            ],
            [
              { design: 2, value: 7 },
              { design: 2, value: 8 },
              { design: 2, value: 9 },
            ],
          ],
          handIndices: [2, 4],
        },
      ],
    });
  });

  it('accepts rank letters in rearrange', () => {
    expect(parseMachiavelliCommand('rearrange sA,hA,dK / 0')).toEqual({
      args: [
        'play',
        {
          tableMelds: [
            [
              { design: 1, value: 1 },
              { design: 3, value: 1 },
              { design: 4, value: 13 },
            ],
          ],
          handIndices: [0],
        },
      ],
    });
  });

  it('tolerates spacing and empty segments in rearrange', () => {
    expect(parseMachiavelliCommand('ra  s5, h5 , d5 ; ; c7,c8,c9 ,  / 2 , 4 ,')).toEqual({
      args: [
        'play',
        {
          tableMelds: [
            [
              { design: 1, value: 5 },
              { design: 3, value: 5 },
              { design: 4, value: 5 },
            ],
            [
              { design: 2, value: 7 },
              { design: 2, value: 8 },
              { design: 2, value: 9 },
            ],
          ],
          handIndices: [2, 4],
        },
      ],
    });
  });

  it.each(['ra', 'ra s5,h5,d5', 'ra / 1', 'ra s5,x9 / 1', 'ra s5,h5,d5 / x', 'ra s5,h5,d5 /', 'ra s5 / 1 / 2'])(
    'rejects malformed rearrange: %s',
    (input) => {
      const result = parseMachiavelliCommand(input);

      expect(result).toHaveProperty('error');
    },
  );
});
