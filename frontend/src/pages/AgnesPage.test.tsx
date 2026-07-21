import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { agnesApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AgnesResponse, Card, CardDesign } from '../types/card';
import { AgnesPage } from './AgnesPage';

vi.mock('../api/gameApi', () => ({
  agnesApi: { exec: vi.fn() },
  actionLogApi: { agnes: vi.fn() },
}));

const mockExec = vi.mocked(agnesApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const up = (design: CardDesign, value: number) => ({ card: card(design, value), faceUp: true });
const down = () => ({ card: null, faceUp: false });

const playingState: AgnesResponse = {
  tableau: [
    [up('SPADE', 7)],
    [down(), up('HEART', 8)],
    [down(), down(), up('CLOVER', 9)],
    [up('DIAMOND', 10)],
    [up('SPADE', 4)],
    [up('HEART', 5)],
    [up('CLOVER', 6)],
  ],
  stockCount: 23,
  foundation: [[card('SPADE', 5)], [], [], []],
  baseRank: 5,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  message: '',
  messageCode: 'agnes.playing',
};

const gameClearState: AgnesResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'agnes.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: AgnesResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'agnes.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('AgnesPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows base rank', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByText(/ベースランク/)).toBeInTheDocument());
  });

  it('shows stock count', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByText(/山札: 23/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('clicks stock card to fire deal command', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const stockButton = screen.getByRole('button', { name: /山札/ });
    stockButton.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('deal footer button fires deal command', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '配る' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('keyboard: "d" deals and "h" hints', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('keyboard: "z" undoes only when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeEnabled());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard: "z" is a no-op when canUndo is false', async () => {
    mockExec.mockResolvedValue(playingState); // canUndo: false
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    expect(mockExec).not.toHaveBeenCalledWith('undo');
  });

  it('keyboard: "g" opens the give-up confirm and only gives up after confirming', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'g' });
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('advertises the keyboard shortcuts on the action buttons', async () => {
    renderWithProviders(<AgnesPage />);
    const deal = await screen.findByRole('button', { name: '配る' });
    expect(deal).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(deal.querySelector('kbd')?.textContent).toBe('D');
    expect(screen.getByRole('button', { name: 'ヒント' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toHaveAttribute('aria-keyshortcuts', 'g');
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows the error inline without replacing the board (issue #3290)', async () => {
    // First load succeeds, then a subsequent action fails: the error must appear
    // inline while the board (stock count, base rank) stays rendered.
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByText(/山札: 23/)).toBeInTheDocument());
    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    // Board is still present alongside the error banner.
    expect(screen.getByText(/山札: 23/)).toBeInTheDocument();
    expect(screen.getByText(/ベースランク/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument();
  });

  it('per-column action button moves tableau end card to foundation', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getAllByRole('button', { name: '→組' })[0]);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }),
    );
  });

  it('tableau-to-tableau button fires move command with end-card index', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Column 0 (one card, end index 0) → column 1.
    const btn = screen.getAllByRole('button', { name: '→1' })[0];
    btn.click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 1 },
      ),
    );
  });

  it('only the face-up end card of a column is draggable', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Column 1's end card (♥ 8) is face-up and draggable.
    const endImg = screen.getByAltText('♥ 8');
    const endButton = endImg.closest('button') as HTMLButtonElement;
    expect(endButton).toHaveAttribute('draggable', 'true');
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('disables deal button when stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('collapses per-column actions behind a details disclosure on mobile', async () => {
    const orig = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      renderWithProviders(<AgnesPage />);
      const details = await screen.findByTestId('ag-col-actions-0');
      expect(details.tagName.toLowerCase()).toBe('details');
      expect(details).not.toHaveAttribute('open');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: orig });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('shows per-column actions directly (no disclosure) on desktop', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('ag-col-actions-0')).not.toBeInTheDocument();
  });

  // --- Auto-complete (issue #3289) ---

  it('disables auto-complete while the stock is not empty', async () => {
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('ag-autocomplete-button')).toBeDisabled();
  });

  it('enables auto-complete once stock is empty and all cards are face up, then sweeps to foundation', async () => {
    const ready: AgnesResponse = {
      ...playingState,
      stockCount: 0,
      tableau: [[up('SPADE', 6)], [up('HEART', 9)]],
      foundation: [[card('SPADE', 5)], [], [], []],
    };
    // After the sweep move, column 0 is empty and no foundation move remains.
    const swept: AgnesResponse = {
      ...ready,
      tableau: [[], [up('HEART', 9)]],
      foundation: [[card('SPADE', 5), card('SPADE', 6)], [], [], []],
      canUndo: true,
    };
    mockExec.mockResolvedValue(ready);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByTestId('ag-autocomplete-button')).toBeEnabled());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(swept);
    fireEvent.click(screen.getByTestId('ag-autocomplete-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }),
    );
    // Loop stops after the swept state (no further foundation move): exactly one move.
    await waitFor(() => expect(mockExec).toHaveBeenCalledTimes(1));
  });

  it('keyboard: "a" triggers auto-complete when ready', async () => {
    const ready: AgnesResponse = {
      ...playingState,
      stockCount: 0,
      tableau: [[up('SPADE', 6)]],
      foundation: [[card('SPADE', 5)], [], [], []],
    };
    mockExec.mockResolvedValue(ready);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByTestId('ag-autocomplete-button')).toBeEnabled());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce({
      ...ready,
      tableau: [[]],
      foundation: [[card('SPADE', 5), card('SPADE', 6)], [], [], []],
    });
    fireEvent.keyDown(document.body, { key: 'a' });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }),
    );
  });

  // --- Stalemate detection + undo-escape (issue #3289) ---

  it('shows the stalemate banner when no legal move remains', async () => {
    const stuck: AgnesResponse = {
      ...playingState,
      stockCount: 0,
      tableau: [[up('SPADE', 8)], [up('HEART', 3)]],
      foundation: [[], [], [], []],
    };
    mockExec.mockResolvedValue(stuck);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByTestId('ag-stalemate-banner')).toBeInTheDocument());
  });

  it('does not show the stalemate banner while a legal move exists', async () => {
    // Default playing state still has stock, so a deal is always possible.
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('ag-stalemate-banner')).not.toBeInTheDocument();
  });

  it('offers an undo-to-escape button in the stalemate banner and fires undo', async () => {
    const stuck: AgnesResponse = {
      ...playingState,
      stockCount: 0,
      canUndo: true,
      tableau: [[up('SPADE', 8)], [up('HEART', 3)]],
      foundation: [[], [], [], []],
    };
    mockExec.mockResolvedValue(stuck);
    renderWithProviders(<AgnesPage />);
    const escapeBtn = await screen.findByTestId('ag-stalemate-undo');
    mockExec.mockClear();
    fireEvent.click(escapeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('hides the undo-to-escape button when there is nothing to undo', async () => {
    const stuck: AgnesResponse = {
      ...playingState,
      stockCount: 0,
      canUndo: false,
      tableau: [[up('SPADE', 8)], [up('HEART', 3)]],
      foundation: [[], [], [], []],
    };
    mockExec.mockResolvedValue(stuck);
    renderWithProviders(<AgnesPage />);
    await waitFor(() => expect(screen.getByTestId('ag-stalemate-banner')).toBeInTheDocument());
    expect(screen.queryByTestId('ag-stalemate-undo')).not.toBeInTheDocument();
  });
});
