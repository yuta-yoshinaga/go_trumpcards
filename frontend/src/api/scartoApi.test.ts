import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { scartoApi, sessionId } from './gameApi';

describe('scartoApi', () => {
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

  it('reset hits /scarto/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await scartoApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/scarto/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await scartoApi.exec('reset', { config: { cpuDifficulty: 2, targetDeals: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/scarto/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetDeals: 7 }, sessionId }),
      }),
    );
  });

  it('scarto sends command=scarto with the 3 discard indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await scartoApi.exec('scarto', { cardIndices: [0, 1, 2] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/scarto/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'scarto', cardIndices: [0, 1, 2], sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await scartoApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/scarto/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await scartoApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/scarto/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await scartoApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/scarto/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(scartoApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
