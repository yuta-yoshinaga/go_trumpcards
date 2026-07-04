import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, wattenApi } from './gameApi';

describe('wattenApi', () => {
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

  const payload = { players: [], phase: 0, message: '' };

  it('reset hits /watten/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await wattenApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/watten/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('reset', undefined, undefined, undefined, undefined, {
      cpuDifficulty: 2,
      targetScore: 21,
      maxRaises: 4,
    });
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { cpuDifficulty: 2, targetScore: 21, maxRaises: 4 },
          sessionId,
        }),
      }),
    );
  });

  it('declare sends command=declare with rank and suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('declare', 13, 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'declare', rank: 13, suit: 3, sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('play', undefined, undefined, 2);
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('raise sends a bare command=raise', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('raise');
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'raise', sessionId }) }),
    );
  });

  it('respond sends command=respond with the hold flag (hold and fold)', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('respond', undefined, undefined, undefined, true);
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'respond', hold: true, sessionId }) }),
    );
    await wattenApi.exec('respond', undefined, undefined, undefined, false);
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'respond', hold: false, sessionId }) }),
    );
  });

  it('nextround sends its bare command', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await wattenApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/watten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(wattenApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
