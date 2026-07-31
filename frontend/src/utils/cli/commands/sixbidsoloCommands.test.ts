import { describe, expect, it } from 'vitest';
import { parseSixBidSoloCommand, SIXBIDSOLO_HELP } from './sixbidsoloCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseSixBidSoloCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseSixBidSoloCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['ps', 'pass'],
      ['pass', 'pass'],
      ['n', 'next'],
      ['next', 'next'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseSixBidSoloCommand(input)).toEqual({ args: [want] });
    }
  });

  // **6 段階すべて指せる。**
  it('bids any of the six levels', () => {
    for (const n of [1, 2, 3, 4, 5, 6]) {
      expect(parseSixBidSoloCommand(`b ${n}`)).toEqual({ args: ['bid', { bid: n }] });
    }
    expect(parseSixBidSoloCommand('bid 6')).toEqual({ args: ['bid', { bid: 6 }] });
    for (const input of ['b', 'b abc', 'b 0', 'b 7']) {
      expect(parseError(input)).toContain('Usage: b <1-6>');
    }
  });

  it('declares a trump on its own', () => {
    for (const n of [1, 2, 3, 4]) {
      expect(parseSixBidSoloCommand(`d ${n}`)).toEqual({ args: ['declare', { suit: n }] });
    }
    expect(parseSixBidSoloCommand('declare 2')).toEqual({ args: ['declare', { suit: 2 }] });
    for (const input of ['d', 'd abc', 'd 0', 'd 5']) {
      expect(parseError(input)).toContain('Usage: d <1-4>');
    }
  });

  // **コール・ソロは指名札を続ける。**片方だけでは札が決まらない。
  it('declares a trump with a called card', () => {
    expect(parseSixBidSoloCommand('d 1 3 13')).toEqual({
      args: ['declare', { suit: 1, calledSuit: 3, calledValue: 13 }],
    });
    // 片方だけでは札が決まらない。
    expect(parseError('d 1 3')).toContain('both the called suit and its rank');
    for (const input of ['d 1 9 1', 'd 1 abc 1']) {
      expect(parseError(input)).toContain('called suit is 1-4');
    }
    for (const input of ['d 1 3 0', 'd 1 3 14', 'd 1 3 abc']) {
      expect(parseError(input)).toContain('called rank');
    }
  });

  // **手札は 11 枚。**
  it('parses a play and rejects an index outside the hand', () => {
    expect(parseSixBidSoloCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseSixBidSoloCommand('play 10')).toEqual({ args: ['play', { cardIndex: 10 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 11']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('pas')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['b <1-6>', 'ps / pass', 'd <1-4>', 'p <0-10>']) {
      expect(SIXBIDSOLO_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // 11 枚 + ウィドウ 3 枚と、上回る宣言だけが通ることを help でも言う。
    expect(SIXBIDSOLO_HELP.some((line) => line.includes('eleven each, plus a three-card widow'))).toBe(true);
    expect(SIXBIDSOLO_HELP.some((line) => line.includes('only a HIGHER bid stands'))).toBe(true);
  });
});
