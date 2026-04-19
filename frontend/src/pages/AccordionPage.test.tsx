import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { accordionApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AccordionResponse, Card, CardDesign } from '../types/card';
import { AccordionPage } from './AccordionPage';

vi.mock('../api/gameApi', () => ({
  accordionApi: { exec: vi.fn() },
  actionLogApi: { accordion: vi.fn() },
}));

const mockExec = vi.mocked(accordionApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: AccordionResponse = {
  piles: [
    { cards: [card('SPADE', 1)], size: 1 },
    { cards: [card('HEART', 2)], size: 1 },
    { cards: [card('CLOVER', 3)], size: 1 },
    { cards: [card('DIAMOND', 4)], size: 1 },
  ],
  pileCount: 4,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'accordion.playing',
};

const gameClearState: AccordionResponse = {
  ...playingState,
  piles: [{ cards: [card('SPADE', 1)], size: 52 }],
  pileCount: 1,
  phase: 1,
  moveCount: 51,
  messageCode: 'accordion.gameClear',
  messageParams: { moveCount: '51' },
};

const gameOverState: AccordionResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'accordion.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('AccordionPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count and pile count', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
    expect(screen.getByText(/パイル数/)).toBeInTheDocument();
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('giveup button triggers giveup command', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ギブアップ' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button disabled when canUndo is false', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('reset button opens confirmation dialog', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
  });

  it('clicking pile selects it, clicking again deselects', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // Pile 0 ♠1
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(pile0.className).toMatch(/ring-/));
    fireEvent.click(pile0);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('selecting pile 1 then clicking pile 0 (offset=1) dispatches a move', async () => {
    // Same rank 2 on pile 0 and 1 for a valid move
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 2)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('DIAMOND', 4)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile1 = screen.getByRole('button', { name: /^1:/ });
    fireEvent.click(pile1);
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('selecting pile 3 then clicking pile 0 (offset=3) dispatches a move', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile3 = screen.getByRole('button', { name: /^3:/ });
    fireEvent.click(pile3);
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('selecting then clicking a pile 2 away re-selects that pile', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile2 = screen.getByRole('button', { name: /^2:/ });
    fireEvent.click(pile2);
    await waitFor(() => expect(pile2.className).toMatch(/ring-/));
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    // offset=2 is invalid; pile0 becomes the new selection instead
    expect(mockExec).not.toHaveBeenCalled();
    await waitFor(() => expect(pile0.className).toMatch(/ring-/));
  });

  it('renders inline hint when state.hint is set', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromIdx: 3, toIdx: 0 },
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintBox = screen.getByRole('status');
    expect(hintBox.textContent).toMatch(/3/);
    expect(hintBox.textContent).toMatch(/0/);
  });

  it('CLI toggle enables terminal mode', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const cliToggle = screen.getByRole('button', { name: /CLI|GUI/i });
    fireEvent.click(cliToggle);
    await waitFor(() => {
      expect(screen.getByLabelText(/コマンドを入力/)).toBeInTheDocument();
    });
  });

  it('shows action log button in ended phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /棋譜|action log|アクション/i })).toBeInTheDocument();
  });

  it('shows StalemateEscapeButton when isStalemate is true', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument();
  });
});
