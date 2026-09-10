import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { MarriageResponse } from '../../types/games/marriage';
import { sessionId } from '../gameApi';
import { marriageApi } from './marriage';

describe('marriageApi', () => {
  const mockFetch = vi.fn();
  const payload: MarriageResponse = {
    message: '',
    players: [],
    phase: 0,
    roundNumber: 1,
    targetRounds: 3,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    discardTop: null,
    drawPileCount: 0,
    wildJoker: null,
    wildRank: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    declarerIdx: -1,
    declarationValid: false,
    config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 3 },
  };

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve(payload) });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it.each([
    [
      'reset',
      undefined,
      { playerCount: 2, cpuDifficulty: 1, targetRounds: 3 },
      {
        command: 'reset',
        config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 3 },
        sessionId,
      },
    ],
    ['drawstock', undefined, undefined, { command: 'drawstock', sessionId }],
    ['drawdiscard', undefined, undefined, { command: 'drawdiscard', sessionId }],
    ['discard', 4, undefined, { command: 'discard', cardIndex: 4, sessionId }],
    ['declare', 2, undefined, { command: 'declare', cardIndex: 2, sessionId }],
    ['nextround', undefined, undefined, { command: 'nextround', sessionId }],
    ['log', undefined, undefined, { command: 'log', sessionId }],
  ] as [
    'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'declare' | 'nextround' | 'log',
    number | undefined,
    { playerCount?: number; cpuDifficulty?: number; targetRounds?: number } | undefined,
    Record<string, unknown>,
  ][])('%s sends its command body', async (command, cardIndex, config, body) => {
    await marriageApi.exec(command, cardIndex, config);
    expect(mockFetch).toHaveBeenCalledWith('/marriage/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  });

  it('returns the response and reports HTTP errors', async () => {
    await expect(marriageApi.exec('reset')).resolves.toEqual(payload);
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve(null) });
    await expect(marriageApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
