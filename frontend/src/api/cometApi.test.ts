import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cometApi, sessionId } from './gameApi';

describe('cometApi', () => {
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
  const payload = { players: [], phase: 'play', message: '' };

  it('reset hits /comet/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await cometApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/comet/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('sends the hand index with play', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cometApi.exec('play', { handIndex: 2 });
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toMatchObject({ command: 'play', handIndex: 2 });
  });

  // **パスは本物の手。** 連なりが止まるゲームなので、出せない席は言う必要がある。
  it('sends pass with no hand index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cometApi.exec('pass');
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toMatchObject({ command: 'pass' });
    expect(body.handIndex).toBeUndefined();
  });

  it('sends the config on reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cometApi.exec('reset', { config: { cpuDifficulty: 2, players: 5, targetScore: 200 } });
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body.config).toEqual({ cpuDifficulty: 2, players: 5, targetScore: 200 });
  });

  it('rejects on a server error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(cometApi.exec('reset')).rejects.toThrow();
  });
});
