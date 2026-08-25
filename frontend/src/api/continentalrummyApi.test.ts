import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { continentalrummyApi, sessionId } from './gameApi';

describe('continentalrummyApi', () => {
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
  const payload = { players: [], phase: 'draw', message: '' };

  it('reset hits /continentalrummy/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await continentalrummyApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/continentalrummy/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // **山と捨て札は別のコマンドとして届く。**
  it.each([
    ['stock', 'stock'],
    ['take', 'take'],
    ['next', 'next'],
  ] as const)('%s reaches the server as its own command', async (cmd, wire) => {
    mockFetch.mockReturnValue(ok(payload));
    await continentalrummyApi.exec(cmd);
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toEqual({ command: wire, sessionId });
  });

  // **上がるときも捨てる 1 枚を名指す。**
  it.each([
    ['discard', 3],
    ['goout', 15],
  ] as const)('%s carries the hand index', async (cmd, idx) => {
    mockFetch.mockReturnValue(ok(payload));
    await continentalrummyApi.exec(cmd, { handIndex: idx });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toMatchObject({ command: cmd, handIndex: idx });
  });

  it('sends the config on reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await continentalrummyApi.exec('reset', { config: { cpuDifficulty: 2, totalRounds: 10 } });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body).config).toEqual({ cpuDifficulty: 2, totalRounds: 10 });
  });

  it('rejects on a server error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(continentalrummyApi.exec('reset')).rejects.toThrow();
  });
});
