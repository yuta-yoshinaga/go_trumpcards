import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tehonbikiApi } from '../api/games/tehonbiki';
import { useGameApi } from '../hooks/useGameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { TehonbikiResponse } from '../types/games/tehonbiki';
import { TehonbikiPage } from './TehonbikiPage';

vi.mock('../api/games/tehonbiki', () => ({ tehonbikiApi: { exec: vi.fn() } }));
vi.mock('../hooks/useGameApi', () => ({ useGameApi: vi.fn() }));
vi.mock('../components/skeleton/GameSkeleton', () => ({ GameSkeleton: () => null }));

const mockUseGameApi = vi.mocked(useGameApi);
const mockExec = vi.mocked(tehonbikiApi.exec);
const base: TehonbikiResponse = {
  phase: 0,
  numbers: [],
  betType: '',
  bet: 0,
  result: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 6,
  gameEndFlag: false,
  payoutNum: 9,
  payoutDen: 2,
  message: '',
};

const state = (over: Partial<TehonbikiResponse> = {}): TehonbikiResponse => ({ ...base, ...over });

beforeEach(() => {
  vi.clearAllMocks();
  mockUseGameApi.mockReturnValue({
    state: state(),
    loading: false,
    error: null,
    exec: mockExec,
    retry: vi.fn(),
  } as never);
});

describe('TehonbikiPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the six selectable numbers', () => {
    renderWithProviders(<TehonbikiPage />);
    for (let n = 1; n <= 6; n++) expect(screen.getByRole('button', { name: String(n) })).toBeInTheDocument();
  });

  it('sends the selected numbers, wager kind, and amount', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '張る' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'double' } });
    fireEvent.click(screen.getByRole('button', { name: '2' }));
    fireEvent.click(screen.getByRole('button', { name: '3' }));
    fireEvent.click(screen.getByRole('button', { name: '張る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { numbers: [2, 3], betType: 'double', bet: 50 }));
  });

  it('shows the revealed parent card and next-round action after a result', async () => {
    mockUseGameApi.mockReturnValue({
      state: state({ phase: 1, parentCard: 6, numbers: [6], betType: 'single', bet: 50, result: 1, payout: 225 }),
      loading: false,
      error: null,
      exec: mockExec,
      retry: vi.fn(),
    } as never);
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByText('親の札: 6')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次の勝負' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '張る' })).not.toBeInTheDocument();
  });
});
