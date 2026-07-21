import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { pigtailApi, sessionId } from './gameApi';

describe('pigtailApi', () => {
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

  const payload = { players: [], circleCount: 52, gameEndFlag: false, message: '' };

  it('reset omits the player count when not supplied', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await pigtailApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/pigtail/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset forwards the selected player count', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await pigtailApi.exec('reset', undefined, 6);
    expect(mockFetch).toHaveBeenCalledWith(
      '/pigtail/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'reset', playerCount: 6, sessionId }) }),
    );
  });

  it('draw sends command=draw', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await pigtailApi.exec('draw');
    expect(mockFetch).toHaveBeenCalledWith(
      '/pigtail/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'draw', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(pigtailApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
