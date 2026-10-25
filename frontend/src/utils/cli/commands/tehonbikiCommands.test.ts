import { describe, expect, it } from 'vitest';
import { parseTehonbikiCommand } from './tehonbikiCommands';

describe('parseTehonbikiCommand', () => {
  it('parses all covered numbers and the wager amount', () => {
    expect(parseTehonbikiCommand('bet double 2 5 50')).toEqual({
      args: ['bet', { betType: 'double', numbers: [2, 5], bet: 50 }],
    });
  });
  it('normalizes uppercase bet types and accepts the bet alias', () => {
    expect(parseTehonbikiCommand('bet DOUBLE 1 2 50')).toEqual({
      args: ['bet', { betType: 'double', numbers: [1, 2], bet: 50 }],
    });
    expect(parseTehonbikiCommand('b double 1 2 50')).toEqual({
      args: ['bet', { betType: 'double', numbers: [1, 2], bet: 50 }],
    });
  });
  it.each([
    ['next', ['next']],
    ['hint', ['hint']],
    ['h', ['hint']],
    ['reset', ['reset']],
    ['r', ['reset']],
  ])('maps %s to the page API action', (input, expected) => {
    expect(parseTehonbikiCommand(input)).toEqual({ args: expected });
  });
  it('treats log as an unknown command', () => {
    expect(parseTehonbikiCommand('log')).toEqual({ error: 'Unknown command: log' });
  });
  it.each([
    'bet',
    'bet single',
    'bet single 1',
    'bet single x 50',
    'bet single 0 50',
    'bet single 7 50',
    'bet single 1 0',
    'bet single 1 9',
    'bet single 1 501',
    'bet mystery 1 50',
    'bet double 1 50',
    'bet triple 1 2 2 50',
    'bet half 1 2 4 50',
    'unknown',
  ])('rejects invalid input %s', (input) => {
    expect(parseTehonbikiCommand(input)).toHaveProperty('error');
  });
  it('maps a legal half wager to the same values as the bet button', () => {
    expect(parseTehonbikiCommand('bet half 4 5 6 50')).toEqual({
      args: ['bet', { numbers: [4, 5, 6], betType: 'half', bet: 50 }],
    });
  });
  it('accepts the half alias command', () => {
    expect(parseTehonbikiCommand('half 1 2 3 50')).toEqual({
      args: ['bet', { numbers: [1, 2, 3], betType: 'half', bet: 50 }],
    });
  });
});
