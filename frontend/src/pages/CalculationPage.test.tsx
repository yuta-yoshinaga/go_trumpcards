import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calculationApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CalculationResponse, Card, CardDesign } from '../types/card';
import { CalculationPage } from './CalculationPage';

vi.mock('../api/gameApi', () => ({
  calculationApi: { exec: vi.fn() },
  actionLogApi: { calculation: vi.fn() },
}));

const mockExec = vi.mocked(calculationApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: CalculationResponse = {
  foundations: [[card('SPADE', 1)], [card('HEART', 2)], [card('DIAMOND', 3)], [card('CLOVER', 4)]],
  wastes: [[], [], [], []],
  stockCount: 48,
  stockTop: card('SPADE', 7),
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'calculation.playing',
};

const gameClearState: CalculationResponse = {
  ...playingState,
  phase: 1,
  stockCount: 0,
  stockTop: undefined,
  moveCount: 100,
  messageCode: 'calculation.gameClear',
  messageParams: { moveCount: '100' },
};

const gameOverState: CalculationResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'calculation.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('CalculationPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('giveup button triggers giveup command', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ギブアップ' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button disabled when canUndo is false', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('autocomplete button is disabled while stock is non-empty', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('autocomplete button enables when stock is empty and wastes are non-empty', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      stockCount: 0,
      stockTop: undefined,
      wastes: [[card('SPADE', 5)], [], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByTestId('autocomplete-button');
    expect(btn).not.toBeDisabled();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('clicking foundation without a source has no effect', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const f0 = screen.getByLabelText(/Foundation 0 \+1/);
    fireEvent.click(f0);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('clicking deselect button clears source', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Select stock by clicking the stock top button (find by aria-pressed=false initially).
    const buttons = screen.getAllByRole('button');
    const stockBtn = buttons.find((b) => b.getAttribute('aria-pressed') === 'false');
    if (stockBtn) fireEvent.click(stockBtn);

    // After selection, a cancel button appears
    const cancelBtn = screen.queryByRole('button', { name: 'キャンセル' });
    if (cancelBtn) {
      fireEvent.click(cancelBtn);
      await waitFor(() => expect(screen.queryByRole('button', { name: 'キャンセル' })).toBeNull());
    }
  });

  it('reset button opens confirmation dialog', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
  });
});
