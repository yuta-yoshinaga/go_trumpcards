import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, postJson } from './gameExec';

describe('postJson HTTP errors', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('preserves a JSON server message on ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        json: async () => ({ message: 'param error: invalid move zones.' }),
      }),
    );
    await expect(postJson('/game/exec', {})).rejects.toMatchObject({
      status: 400,
      serverMessage: 'param error: invalid move zones.',
    });
  });

  it('throws ApiError without a message for HTML error responses', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        json: async () => {
          throw new SyntaxError('Unexpected token <');
        },
      }),
    );
    await expect(postJson('/game/exec', {})).rejects.toBeInstanceOf(ApiError);
  });
});
