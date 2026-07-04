import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { michiganApi } from '../api/gameApi';
import { makeMichiganState } from '../test/stateFactories';
import { DEFAULT_MICHIGAN_CONFIG, useMichiganGame } from './useMichiganGame';

vi.mock('../api/gameApi', () => ({
  michiganApi: { exec: vi.fn() },
  actionLogApi: { michigan: vi.fn() },
}));

const mockExec = vi.mocked(michiganApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeMichiganState());
});

describe('useMichiganGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useMichiganGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_MICHIGAN_CONFIG));
  });

  it('handleBet dispatches bet with the boodle distribution', async () => {
    const { result } = renderHook(() => useMichiganGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBet([2, 2, 2, 2]));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', [2, 2, 2, 2]));
  });

  it('handlePlay dispatches play with the card index', async () => {
    const { result } = renderHook(() => useMichiganGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => useMichiganGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useMichiganGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('playerCount', '6'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        ...DEFAULT_MICHIGAN_CONFIG,
        playerCount: 6,
      }),
    );
  });
});
