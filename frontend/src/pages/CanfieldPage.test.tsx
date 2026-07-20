import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { canfieldApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CanfieldResponse, Card, CardDesign } from '../types/card';
import { CanfieldPage } from './CanfieldPage';

vi.mock('../api/gameApi', () => ({
  canfieldApi: { exec: vi.fn() },
  actionLogApi: { canfield: vi.fn() },
}));

const mockExec = vi.mocked(canfieldApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: CanfieldResponse = {
  tableau: [
    [{ card: card('SPADE', 7) }],
    [{ card: card('HEART', 8) }],
    [{ card: card('CLOVER', 9) }],
    [{ card: card('DIAMOND', 10) }],
  ],
  reserve: [card('SPADE', 3)],
  stockCount: 34,
  waste: [card('HEART', 4)],
  foundation: [[card('SPADE', 5)], [], [], []],
  baseRank: 5,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  message: '',
  messageCode: 'canfield.playing',
};

const gameClearState: CanfieldResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'canfield.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: CanfieldResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'canfield.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('CanfieldPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows base rank', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByText(/ベースランク/)).toBeInTheDocument());
  });

  it('reserve stack reflects remaining count via data-reserve-layers', async () => {
    const fullReserveState: CanfieldResponse = {
      ...playingState,
      reserve: Array.from({ length: 13 }, (_, i) => card('SPADE', (i % 13) + 1)),
    };
    mockExec.mockResolvedValue(fullReserveState);
    renderWithProviders(<CanfieldPage />);
    const stack = await screen.findByTestId('canfield-reserve-stack');
    expect(stack.getAttribute('data-reserve-layers')).toBe('5');
  });

  it('reserve stack shows zero layers when only one card remains', async () => {
    renderWithProviders(<CanfieldPage />);
    const stack = await screen.findByTestId('canfield-reserve-stack');
    expect(stack.getAttribute('data-reserve-layers')).toBe('0');
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('clicks draw button (stock) to fire draw command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const stockButton = screen.getByRole('button', { name: /山札/ });
    stockButton.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('moves reserve to foundation', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: /リザーブ→組札/ });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'foundation' }));
  });

  it('moves waste to foundation', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: /ウェイスト→組札/ });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows the hint source → destination text after clicking hint (tableau → foundation)', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The hint response carries the suggested move: tableau col 0 top card → foundation.
    mockExec.mockResolvedValueOnce({
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 2 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    const display = await screen.findByTestId('cf-hint-display');
    // Visible line names both source (場札 0) and destination (組札).
    await waitFor(() => expect(display).toHaveTextContent('場札 0'));
    expect(display).toHaveTextContent('組札');
    // Screen-reader announcement resolves the hinted card face + destination.
    expect(screen.getByTestId('cf-hint-announcement')).toHaveTextContent('♠ 7');
  });

  it('announces a reserve → tableau hint using the reserve top card', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockResolvedValueOnce({
      ...playingState,
      hint: { fromZone: 'reserve', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 1 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    const display = await screen.findByTestId('cf-hint-display');
    await waitFor(() => expect(display).toHaveTextContent('リザーブ'));
    expect(display).toHaveTextContent('場札 1');
    // Reserve top card is ♠ 3 in playingState.
    expect(screen.getByTestId('cf-hint-announcement')).toHaveTextContent('♠ 3');
  });

  it('clears the hint display after the next move (response without a hint field)', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockResolvedValueOnce({
      ...playingState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 0 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByTestId('cf-hint-display')).toHaveTextContent('ヒントがあります'));
    // A subsequent move returns the plain state (no hint) → the display empties.
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: /ウェイスト→組札/ }));
    await waitFor(() => expect(screen.getByTestId('cf-hint-display')).not.toHaveTextContent('ヒントがあります'));
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('collapses per-column actions behind a details disclosure on mobile', async () => {
    const orig = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      renderWithProviders(<CanfieldPage />);
      const details = await screen.findByTestId('cf-col-actions-0');
      expect(details.tagName.toLowerCase()).toBe('details');
      // Collapsed by default (no open attribute).
      expect(details).not.toHaveAttribute('open');
      // The <summary> itself carries the (0-based, matching the UI) column number
      // so the accordion labels are distinguishable to a screen reader.
      expect(details.querySelector('summary')).toHaveTextContent('列 0 の操作');
      expect(screen.getByTestId('cf-col-actions-1').querySelector('summary')).toHaveTextContent('列 1 の操作');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: orig });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('shows per-column actions directly (no disclosure) on desktop', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('cf-col-actions-0')).not.toBeInTheDocument();
  });

  it('collapses per-column actions on desktop when the setting is toggled on', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Initially expanded (no disclosure) on desktop.
    expect(screen.queryByTestId('cf-col-actions-0')).not.toBeInTheDocument();
    // Toggle the "collapse column actions" setting.
    const toggle = screen.getByRole('checkbox', { name: '列の操作ボタンを折りたたむ' });
    fireEvent.click(toggle);
    // Now each column's actions live behind a <details> disclosure, collapsed by default.
    const details = await screen.findByTestId('cf-col-actions-0');
    expect(details.tagName.toLowerCase()).toBe('details');
    expect(details).not.toHaveAttribute('open');
    expect(details.querySelector('summary')).toHaveTextContent('列 0 の操作');
    // The move buttons remain reachable inside the disclosure (functionality preserved).
    expect(screen.getAllByRole('button', { name: '→組' }).length).toBeGreaterThan(0);
  });

  it('persists the collapse setting across remounts via localStorage', async () => {
    localStorage.setItem('canfield-collapse-col-actions', 'true');
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Restored as collapsed on desktop without re-toggling.
    expect(await screen.findByTestId('cf-col-actions-0')).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: '列の操作ボタンを折りたたむ' })).toBeChecked();
  });

  it('per-column action buttons dispatch the matching move', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Column 0's action group. Await each call before the next click, since a
    // pending exec sets loading=true and disables the buttons in between.
    fireEvent.click(screen.getByRole('button', { name: '捨→0' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 0 }));
    fireEvent.click(screen.getByRole('button', { name: '予→0' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'tableau', col: 0 }),
    );
    fireEvent.click(screen.getAllByRole('button', { name: '→組' })[0]);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }),
    );
    fireEvent.click(screen.getAllByRole('button', { name: '→1' })[0]);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 1 },
      ),
    );
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('renders empty waste and reserve placeholders', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [], reserve: [] });
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The reserve/waste sections still render without their top cards
    expect(screen.getByText(/リザーブ: 0/)).toBeInTheDocument();
  });

  it('tableau-to-tableau button fires move command with correct args', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Column 0 → column 1 button (ja moveToCol label is exactly "→1"; each non-empty column renders one per other column)
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

  describe('drag and drop', () => {
    function buildDataTransfer() {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    }

    it('waste card is draggable when playing', async () => {
      renderWithProviders(<CanfieldPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const wasteImg = screen.getByAltText('♥ 4');
      const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
      expect(wasteButton).toHaveAttribute('draggable', 'true');
    });

    it('reserve card is draggable when playing', async () => {
      renderWithProviders(<CanfieldPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const reserveImg = screen.getByAltText('♠ 3');
      const reserveButton = reserveImg.closest('button') as HTMLButtonElement;
      expect(reserveButton).toHaveAttribute('draggable', 'true');
    });

    it('tableau card is draggable when playing', async () => {
      renderWithProviders(<CanfieldPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const cardImg = screen.getByAltText('♠ 7');
      const cardButton = cardImg.closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging waste card to tableau dispatches move', async () => {
      renderWithProviders(<CanfieldPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const wasteImg = screen.getByAltText('♥ 4');
      const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();

      fireEvent.dragStart(wasteButton, { dataTransfer });

      // Drop on a tableau column DropZone (use the column header text to find it)
      const colHeaders = screen.getAllByText(/^#\d$/);
      const dropZone = colHeaders[1].closest('.flex.flex-col')?.querySelector('[role="presentation"]');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'waste' }),
          expect.objectContaining({ zone: 'tableau' }),
        ),
      );
    });

    it('dragging reserve card to foundation dispatches move', async () => {
      renderWithProviders(<CanfieldPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const reserveImg = screen.getByAltText('♠ 3');
      const reserveButton = reserveImg.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();

      fireEvent.dragStart(reserveButton, { dataTransfer });

      // Drop on foundation DropZone
      const foundationTexts = screen.getAllByText('組札');
      const dropZone = foundationTexts[0].closest('[role="presentation"]');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'reserve' }),
          expect.objectContaining({ zone: 'foundation' }),
        ),
      );
    });
  });
});
