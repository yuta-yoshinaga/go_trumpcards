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

  it('selecting stock and clicking a foundation dispatches a stock→foundation move', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    fireEvent.click(screen.getByLabelText(/Foundation 2 \+3/));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'foundation', idx: 2 }),
    );
  });

  it('selecting stock and clicking an empty waste dispatches a stock→waste move', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    fireEvent.click(screen.getByTestId('calc-waste-button-1'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'waste', idx: 1 }));
  });

  it('selecting a waste and clicking a foundation dispatches a waste→foundation move', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[], [card('HEART', 4)], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-waste-button-1'));
    fireEvent.click(screen.getByLabelText(/Foundation 1 \+2/));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste', idx: 1 }, { zone: 'foundation', idx: 1 }),
    );
  });

  it('clicking the same source twice toggles the selection off', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const stockBtn = screen.getByTestId('calc-stock-button');
    fireEvent.click(stockBtn);
    expect(stockBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(stockBtn);
    expect(stockBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('cancel button clears the active selection', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    const cancelBtn = await screen.findByRole('button', { name: 'キャンセル' });
    fireEvent.click(cancelBtn);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'キャンセル' })).toBeNull());
  });

  it('clicking different waste tops switches the selection', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[card('SPADE', 9)], [card('HEART', 5)], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const w0 = screen.getByTestId('calc-waste-button-0');
    const w1 = screen.getByTestId('calc-waste-button-1');
    fireEvent.click(w0);
    expect(w0).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(w1);
    expect(w0).toHaveAttribute('aria-pressed', 'false');
    expect(w1).toHaveAttribute('aria-pressed', 'true');
  });

  it('renders the backend hint banner when state.hint is a stock hint', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'stock', wasteIdx: -1, foundationIdx: 2 },
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument();
  });

  it('renders the backend hint banner when state.hint is a waste hint', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[card('SPADE', 5)], [], [], []],
      hint: { fromZone: 'waste', wasteIdx: 0, foundationIdx: 1 },
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/W0 → F1/)).toBeInTheDocument();
  });

  it('renders a stalemate escape button when the game is stalled', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /脱出する/ })).toBeInTheDocument();
  });

  it('reset confirmation dialog dispatches a reset when confirmed', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);

    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    const confirmBtn = await screen.findByRole('button', { name: '確認' });
    fireEvent.click(confirmBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('CLI toggle switches to the CLI terminal', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cliToggle = screen.getByRole('button', { name: 'CLIモードに切り替え' });
    fireEvent.click(cliToggle);

    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
