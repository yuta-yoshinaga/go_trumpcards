import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { oasispokerApi } from './gameApi';

describe('oasispokerApi', () => {
  const fetchMock = vi.fn();
  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
    fetchMock.mockReset();
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts reset to /oasispoker/exec', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ phase: 1, message: '' }),
    });
    await oasispokerApi.exec('reset');
    expect(fetchMock).toHaveBeenCalled();
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toMatch(/\/oasispoker\/exec$/);
    const body = JSON.parse((init as RequestInit).body as string);
    expect(body).toMatchObject({ command: 'reset' });
    expect(body.sessionId).toBeTruthy();
  });

  it('posts bet with amount and jackpotBet', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ phase: 1 }),
    });
    await oasispokerApi.exec('bet', 100, 10);
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const body = JSON.parse(init.body as string);
    expect(body).toMatchObject({ command: 'bet', amount: 100, jackpotBet: 10 });
  });

  it('posts exchange with indices', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ phase: 3 }),
    });
    await oasispokerApi.exec('exchange', undefined, undefined, [0, 2]);
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const body = JSON.parse(init.body as string);
    expect(body).toMatchObject({ command: 'exchange', indices: [0, 2] });
  });

  it('posts stand with no extras', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ phase: 3 }),
    });
    await oasispokerApi.exec('stand');
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const body = JSON.parse(init.body as string);
    expect(body.command).toBe('stand');
  });

  it('throws on non-2xx', async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 500 });
    await expect(oasispokerApi.exec('reset')).rejects.toThrow(/500/);
  });
});
