import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cegoApi, sessionId } from './gameApi';

describe('cegoApi', () => {
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

  it('reset hits /cego/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await cegoApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/cego/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('reset', { config: { cpuDifficulty: 2, targetDeals: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetDeals: 7 }, sessionId }),
      }),
    );
  });

  it('bid sends command=bid with the bid string', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('bid', { bid: 'play' });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 'play', sessionId }) }),
    );
  });

  it('pass sends its bare command', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('pass');
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'pass', sessionId }) }),
    );
  });

  it('contract sends command=contract with the contract string', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('contract', { contract: 'cego' });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'contract', contract: 'cego', sessionId }) }),
    );
  });

  it('discard sends command=discard with the single kept index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('discard', { cardIndices: [3] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'discard', cardIndices: [3], sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cegoApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await cegoApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/cego/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(cegoApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
