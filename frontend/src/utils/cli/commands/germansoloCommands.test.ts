import { describe, expect, it } from 'vitest';
import { GERMAN_SOLO_HELP, parseGermanSoloCommand } from './germansoloCommands';

describe('parseGermanSoloCommand', () => {
  it('parses bid pass', () => {
    expect(parseGermanSoloCommand('bid pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseGermanSoloCommand('b pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseGermanSoloCommand('b p')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses bid frage with a trump-suit letter', () => {
    expect(parseGermanSoloCommand('bid frage s')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 1 }] });
    expect(parseGermanSoloCommand('b f c')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 2 }] });
  });

  it('parses bid solo with a trump-suit letter', () => {
    expect(parseGermanSoloCommand('bid solo h')).toEqual({ args: ['bid', { bid: 3, trumpSuit: 3 }] });
    expect(parseGermanSoloCommand('b s d')).toEqual({ args: ['bid', { bid: 3, trumpSuit: 4 }] });
  });

  it('parses bid tout with a trump-suit letter', () => {
    expect(parseGermanSoloCommand('bid tout d')).toEqual({ args: ['bid', { bid: 4, trumpSuit: 4 }] });
    expect(parseGermanSoloCommand('b t h')).toEqual({ args: ['bid', { bid: 4, trumpSuit: 3 }] });
  });

  // **Mussfrage は宣言できない。** 卓が押し付ける契約なので、打てても
  // サーバに弾かれるだけ。
  it('does not accept mussfrage as a declaration', () => {
    expect('error' in parseGermanSoloCommand('bid mussfrage s')).toBe(true);
  });

  // **エース呼びを抜ける唯一の入力。** これが無いと Frage 落札の直後に
  // CLI から先へ進めなくなる。
  it('parses the ace call (short and long)', () => {
    expect(parseGermanSoloCommand('ace c')).toEqual({ args: ['ace', { aceSuit: 2 }] });
    expect(parseGermanSoloCommand('a h')).toEqual({ args: ['ace', { aceSuit: 3 }] });
  });

  it('returns error for an ace call without a valid suit', () => {
    expect('error' in parseGermanSoloCommand('ace')).toBe(true);
    expect('error' in parseGermanSoloCommand('ace x')).toBe(true);
  });

  it('returns error for a contract without a valid suit', () => {
    expect('error' in parseGermanSoloCommand('bid frage')).toBe(true);
    expect('error' in parseGermanSoloCommand('bid solo x')).toBe(true);
    expect('error' in parseGermanSoloCommand('bid tout')).toBe(true);
  });

  it('returns error for bid without a valid argument', () => {
    expect('error' in parseGermanSoloCommand('bid')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseGermanSoloCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseGermanSoloCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseGermanSoloCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseGermanSoloCommand('n')).toEqual({ args: ['next'] });
    expect(parseGermanSoloCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseGermanSoloCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGermanSoloCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseGermanSoloCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGermanSoloCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseGermanSoloCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGermanSoloCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseGermanSoloCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseGermanSoloCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(GERMAN_SOLO_HELP.length).toBeGreaterThan(0);
    expect(GERMAN_SOLO_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
