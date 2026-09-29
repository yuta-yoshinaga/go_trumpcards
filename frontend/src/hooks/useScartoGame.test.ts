import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { scartoApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { useScartoGame } from './useScartoGame';

vi.mock('../api/gameApi', () => ({ scartoApi: { exec: vi.fn() }, actionLogApi: { scarto: vi.fn() } }));
const mockExec = vi.mocked(scartoApi.exec);
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { mutations: { retry: false } } }) },
    children,
  );

describe('useScartoGame config validation', () => {
  beforeEach(() => mockExec.mockReset());
  it('does not send a target deal count outside 1 through 100', async () => {
    const { result } = renderHook(() => useScartoGame(), { wrapper });
    act(() => result.current.handleConfigChange('targetDeals', '101'));
    act(() => result.current.reset());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
