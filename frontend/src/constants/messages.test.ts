import { describe, expect, it } from 'bun:test';
import { NETWORK_ERROR_MESSAGE } from './messages';

describe('NETWORK_ERROR_MESSAGE', () => {
  it('returns the expected value', () => {
    expect(NETWORK_ERROR_MESSAGE()).toBe('通信エラーが発生しました。もう一度お試しください。');
  });
});
