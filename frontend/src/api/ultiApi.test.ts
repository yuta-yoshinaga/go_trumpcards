import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, ultiApi } from './gameApi';

describe('ultiApi', () => {
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

  it('reset hits /ulti/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await ultiApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/ulti/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('reset', { config: { cpuDifficulty: 2, targetRounds: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetRounds: 7 }, sessionId }),
      }),
    );
  });

  it('party bid sends command=bid with the contract and trump suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('bid', { contract: 'party', trumpSuit: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'bid', contract: 'party', trumpSuit: 3, sessionId }),
      }),
    );
  });

  it('betli bid sends command=bid with just the contract', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('bid', { contract: 'betli' });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', contract: 'betli', sessionId }) }),
    );
  });

  it('discard sends command=discard with the card indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('discard', { cardIndices: [0, 3] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'discard', cardIndices: [0, 3], sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ultiApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await ultiApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/ulti/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(ultiApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
