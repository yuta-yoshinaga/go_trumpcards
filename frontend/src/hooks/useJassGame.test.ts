import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { jassApi } from '../api/gameApi';
import type { JassResponse } from '../types/card';
import { JassPhase } from '../types/phases';
import { DEFAULT_JASS_CONFIG, useJassGame } from './useJassGame';

vi.mock('../api/gameApi', () => ({
  jassApi: { exec: vi.fn() },
  actionLogApi: { jass: vi.fn() },
}));

const mockExec = vi.mocked(jassApi.exec);

function makeState(overrides: Partial<JassResponse> = {}): JassResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 9, cards: [], team: 0, trickCount: 0 },
      { id: 1, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 9, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
    ],
    phase: JassPhase.BID_TRUMP,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    forehandIdx: 0,
    trumpSuit: 0,
    schieben: false,
    makerTeam: -1,
    makerPlayerIdx: -1,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundWeisPoints: [0, 0],
    roundStockPoints: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 1000, lastTrickBonus: 5, enableWeis: true },
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('useJassGame', () => {
  it('dispatches reset on mount with the default config', async () => {
    renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_JASS_CONFIG));
  });

  it('handleCallTrump dispatches calltrump with the suit', async () => {
    const { result } = renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleCallTrump(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', 3));
  });

  it('handleSchieben dispatches schieben', async () => {
    const { result } = renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleSchieben());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('schieben'));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handlePlay());
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play with the selected card index', async () => {
    const { result } = renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useJassGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleConfigChange('targetScore', '500'));
    mockExec.mockClear();
    act(() => {
      void (result.current.exec as unknown as (c: string, s?: number, ci?: number, cfg?: unknown) => Promise<void>)(
        'reset',
        undefined,
        undefined,
        result.current.jassConfig,
      );
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        ...DEFAULT_JASS_CONFIG,
        targetScore: 500,
      }),
    );
  });
});
