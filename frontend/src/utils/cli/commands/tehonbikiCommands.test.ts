import { describe, expect, it } from 'vitest';
import { parseTehonbikiCommand } from './tehonbikiCommands';

describe('parseTehonbikiCommand', () => {
  it('parses all covered numbers and the wager amount', () => {
    expect(parseTehonbikiCommand('bet double 2 5 50')).toEqual({
      args: ['bet', { betType: 'double', numbers: [2, 5], bet: 50 }],
    });
  });
});
