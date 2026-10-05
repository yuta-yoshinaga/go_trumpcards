import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { badugiApi } from '../api/gameApi';
import type { BadugiPlayerData, BadugiResponse } from '../types/card';
import { BadugiPhase } from '../types/phases';
import { useBadugiGame } from './useBadugiGame';

vi.mock('../api/gameApi', () => ({
  badugiApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(badugiApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const humanPlayer: BadugiPlayerData = {
  id: 0,
  isHuman: true,
  cards: [{ design: 'SPADE', value: 1 }],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handSize: 0,
  handName: '',
  drawCount: 0,
  totalDraws: 0,
  playStyleName: '',
};

const baseState: BadugiResponse = {
  players: [humanPlayer],
  pot: 40,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: BadugiPhase.DRAW,
  drawIndex: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 10,
  ante: 10,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  cpuExchanges: [],
  message: '',
};

beforeEach(() => {
  mockExec.mockReset();
});

describe('useBadugiGame', () => {
  it('toggles card selection during the exchange phase', async () => {
    mockExec.mockResolvedValue(baseState);
    const { result } = renderHook(() => useBadugiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(baseState));

    act(() => result.current.toggleCard(0));

    expect(result.current.selected).toEqual([0]);
    expect(result.current.canExchange).toBe(true);
  });

  it('does not change card selection outside the exchange phase', async () => {
    mockExec.mockResolvedValue({ ...baseState, phase: BadugiPhase.BET });
    const { result } = renderHook(() => useBadugiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state?.phase).toBe(BadugiPhase.BET));

    act(() => result.current.toggleCard(0));

    expect(result.current.selected).toEqual([]);
    expect(result.current.canExchange).toBe(false);
  });
});
