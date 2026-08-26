import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, sutdaApi } from './gameApi';

describe('sutdaApi', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const ok = (data: unknown) => Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(data) });
  const err = () => Promise.resolve({ ok: false, status: 500, json: () => Promise.resolve(null) });
  const payload = { players: [], phase: 'bet', message: '' };

  it('reset hits /sutda/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await sutdaApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/sutda/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the table setup', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await sutdaApi.exec('reset', { config: { cpuDifficulty: 2, seats: 5, startChips: 5000 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/sutda/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { cpuDifficulty: 2, seats: 5, startChips: 5000 },
          sessionId,
        }),
      }),
    );
  });

  // **どの手も札の指定を持たない。** ソッタの手札は配られたまま変わらない。
  it.each(['call', 'raise', 'fold', 'nexthand', 'hint', 'log'] as const)(
    '%s carries no extra fields',
    async (command) => {
      mockFetch.mockReturnValue(ok(payload));
      await sutdaApi.exec(command);
      expect(mockFetch).toHaveBeenCalledWith(
        '/sutda/exec',
        expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
      );
    },
  );

  it('throws when the server fails', async () => {
    mockFetch.mockReturnValue(err());
    await expect(sutdaApi.exec('call')).rejects.toThrow();
  });
});
