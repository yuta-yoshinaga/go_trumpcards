import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bristolApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BristolResponse, Card, CardDesign } from '../types/card';
import { BristolPage } from './BristolPage';

vi.mock('../api/gameApi', () => ({
  bristolApi: { exec: vi.fn() },
  actionLogApi: { bristol: vi.fn() },
}));

const mockExec = vi.mocked(bristolApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: BristolResponse = {
  tableau: [
    [card('SPADE', 8)],
    [card('HEART', 9)],
    [card('CLOVER', 4)],
    [card('DIAMOND', 10)],
    [card('SPADE', 3)],
    [card('HEART', 6)],
    [card('CLOVER', 2)],
    [card('DIAMOND', 7)],
  ],
  fan: [[card('HEART', 4)], [], []],
  stockCount: 28,
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  message: '',
  messageCode: 'bristol.playing',
};

const gameClearState: BristolResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'bristol.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: BristolResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'bristol.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('BristolPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows a stacked-count badge on fans with 2+ cards and hides it otherwise', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      fan: [[card('HEART', 4), card('SPADE', 5), card('CLOVER', 6)], [card('DIAMOND', 9)], []],
    });
    renderWithProviders(<BristolPage />);
    // Fan 0 has 3 cards → badge shows the count.
    await waitFor(() => expect(screen.getByTestId('br-fan-count-0')).toHaveTextContent('3'));
    // Fan 1 has a single card → no badge.
    expect(screen.queryByTestId('br-fan-count-1')).not.toBeInTheDocument();
  });

  it('clicks stock to fire draw command', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '山札' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('selects a tableau column then moves it to another tableau column', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '場札 0' }).click();
    // Wait for the source selection to render before clicking the destination,
    // so the destination handler reads the updated `selected` state.
    await waitFor(() => expect(screen.getByRole('button', { name: '場札 0' })).toHaveAttribute('aria-pressed', 'true'));
    screen.getByRole('button', { name: '場札 1' }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }),
    );
  });

  it('selects a tableau column then moves it to a foundation', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '場札 0' }).click();
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 0' })).toBeEnabled());
    screen.getByRole('button', { name: '組札 0' }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 0 }),
    );
  });

  it('selects a fan then moves it to a foundation', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ファン 0' }).click();
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 1' })).toBeEnabled());
    screen.getByRole('button', { name: '組札 1' }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'fan', col: 0 }, { zone: 'foundation', col: 1 }),
    );
  });

  it('foundations are disabled until a source is selected', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '組札 0' })).toBeDisabled();
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '自動完成' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button triggers giveup command', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ギブアップ' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button fires undo when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase and hides action buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
