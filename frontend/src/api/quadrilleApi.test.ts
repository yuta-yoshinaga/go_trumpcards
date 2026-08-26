import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { quadrilleApi, sessionId } from './gameApi';

describe('quadrilleApi', () => {
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

  it('reset hits /quadrille/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await quadrilleApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/quadrille/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quadrilleApi.exec('reset', { config: { cpuDifficulty: 2, targetRounds: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetRounds: 7 }, sessionId }),
      }),
    );
  });

  it('pass bid sends command=bid with bid=0', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quadrilleApi.exec('bid', { bid: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 0, sessionId }) }),
    );
  });

  it('entrar bid sends bid=1 with the chosen trump suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quadrilleApi.exec('bid', { bid: 1, trumpSuit: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 1, trumpSuit: 3, sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quadrilleApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quadrilleApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await quadrilleApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/quadrille/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(quadrilleApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
