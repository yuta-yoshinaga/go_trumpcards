import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { presidentApi } from '../api/gameApi';
import { usePresidentGame } from './usePresidentGame';

vi.mock('../api/gameApi', () => ({
  presidentApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(presidentApi.exec);

describe('usePresidentGame', () => {
  beforeEach(() => {
    mockExec.mockReset();
    mockExec.mockResolvedValue({} as Awaited<ReturnType<typeof presidentApi.exec>>);
  });

  it('tracks pending settings changes and clears them when resetting with config', async () => {
    const client = new QueryClient();
    const wrapper = ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client, children });
    const { result } = renderHook(() => usePresidentGame(), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    act(() => result.current.handleConfigChange('revolutionEnabled', false));
    expect(result.current.hasPendingConfigChanges).toBe(true);

    act(() => result.current.handleResetWithConfig());
    expect(result.current.hasPendingConfigChanges).toBe(false);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.objectContaining({ revolutionEnabled: false })),
    );
  });
});
