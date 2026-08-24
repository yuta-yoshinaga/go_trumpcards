import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { piedmonteseTarotApi, sessionId } from './gameApi';

describe('piedmonteseTarotApi', () => {
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

  it('reset hits /piedmontesetarot/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await piedmonteseTarotApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/piedmontesetarot/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the table size', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await piedmonteseTarotApi.exec('reset', { config: { seats: 3, cpuDifficulty: 2, targetDeals: 6 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/piedmontesetarot/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { seats: 3, cpuDifficulty: 2, targetDeals: 6 },
          sessionId,
        }),
      }),
    );
  });

  // **捨てる枚数はサーバーが検める。** クライアントは選んだ番号をそのまま送る。
  it('scarto sends the chosen indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await piedmonteseTarotApi.exec('scarto', { cardIndices: [0, 1] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/piedmontesetarot/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'scarto', cardIndices: [0, 1], sessionId }) }),
    );
  });

  it('play sends the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await piedmonteseTarotApi.exec('play', { cardIndex: 4 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/piedmontesetarot/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 4, sessionId }) }),
    );
  });

  it('next, nextround, hint and log send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    for (const command of ['next', 'nextround', 'hint', 'log'] as const) {
      await piedmonteseTarotApi.exec(command);
      expect(mockFetch).toHaveBeenCalledWith(
        '/piedmontesetarot/exec',
        expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
      );
    }
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(piedmonteseTarotApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
