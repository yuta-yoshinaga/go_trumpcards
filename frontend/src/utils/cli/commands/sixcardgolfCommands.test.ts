import { describe, expect, it } from 'vitest';
import { parseSixCardGolfCommand } from './sixcardgolfCommands';

describe('parseSixCardGolfCommand', () => {
  it('parses positionless commands and their aliases', () => {
    expect(parseSixCardGolfCommand('r')).toEqual({ args: [{ command: 'reset' }] });
    expect(parseSixCardGolfCommand('reset')).toEqual({ args: [{ command: 'reset' }] });
    expect(parseSixCardGolfCommand('ds')).toEqual({ args: [{ command: 'drawstock' }] });
    expect(parseSixCardGolfCommand('drawstock')).toEqual({ args: [{ command: 'drawstock' }] });
    expect(parseSixCardGolfCommand('dd')).toEqual({ args: [{ command: 'drawdiscard' }] });
    expect(parseSixCardGolfCommand('di')).toEqual({ args: [{ command: 'discard' }] });
    expect(parseSixCardGolfCommand('sf')).toEqual({ args: [{ command: 'skipflip' }] });
    expect(parseSixCardGolfCommand('nr')).toEqual({ args: [{ command: 'nextround' }] });
    expect(parseSixCardGolfCommand('l')).toEqual({ args: [{ command: 'log' }] });
  });

  it('parses position commands', () => {
    expect(parseSixCardGolfCommand('fi 2')).toEqual({ args: [{ command: 'flipinitial', position: 2 }] });
    expect(parseSixCardGolfCommand('sw 5')).toEqual({ args: [{ command: 'swap', position: 5 }] });
    expect(parseSixCardGolfCommand('fl 0')).toEqual({ args: [{ command: 'flip', position: 0 }] });
  });

  it('is case-insensitive on the command word', () => {
    expect(parseSixCardGolfCommand('FLIP 1')).toEqual({ args: [{ command: 'flip', position: 1 }] });
  });

  it('returns a localized usage error when a position is missing', () => {
    expect(parseSixCardGolfCommand('fi')).toEqual({ error: '使い方: fi <位置>' });
    expect(parseSixCardGolfCommand('sw')).toEqual({ error: '使い方: sw <位置>' });
    expect(parseSixCardGolfCommand('fl')).toEqual({ error: '使い方: fl <位置>' });
  });

  it('returns a localized usage error when a position is not numeric', () => {
    expect(parseSixCardGolfCommand('fi x')).toEqual({ error: '使い方: fi <位置>' });
  });

  it('returns a localized unknown-command error including the command', () => {
    expect(parseSixCardGolfCommand('zzz')).toEqual({ error: '不明なコマンド: zzz' });
  });
});
