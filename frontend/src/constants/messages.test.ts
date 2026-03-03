import { describe, expect, it } from 'vitest';
import { NETWORK_ERROR_MESSAGE } from './messages';

describe('NETWORK_ERROR_MESSAGE', () => {
  it('has the expected value', () => {
    expect(NETWORK_ERROR_MESSAGE).toBe('通信エラーが発生しました。もう一度お試しください。');
  });
});
