import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bristolApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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

  it('labels foundations and fans with the top card + count, or empty', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BristolPage />);
    // All foundations empty.
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 0: 空' })).toBeInTheDocument());
    // Fan 0 holds ♥4; fans 1 and 2 are empty (rendered as role=img placeholders).
    expect(screen.getByRole('button', { name: 'ファン 0: ♥ 4（1枚）' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'ファン 1: 空' })).toBeInTheDocument();
  });

  it('includes the top card + count on a non-empty foundation', async () => {
    mockExec.mockResolvedValue({ ...playingState, foundation: [[card('SPADE', 1)], [], [], []] });
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 0: ♠ A（1枚）' })).toBeInTheDocument());
  });

  it('gives tableau columns a contextual aria-label (1-based number, role, depth) and a rule header', async () => {
    renderWithProviders(<BristolPage />);
    // Column 1 (0-based idx 0) holds 1 card → "降順ビルド列 1（1枚）".
    await waitFor(() => expect(screen.getByRole('button', { name: '降順ビルド列 1（1枚）' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '降順ビルド列 8（1枚）' })).toBeInTheDocument();
    // The tableau header conveys the build-down rule with the actual column count (8).
    expect(screen.getByTestId('br-tableau-rule')).toHaveTextContent('8列の降順ビルド');
  });

  it('shows a stacked-count badge on fans with 2+ cards and hides it otherwise', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      fan: [[card('HEART', 4), card('SPADE', 5), card('CLOVER', 6)], [card('DIAMOND', 9), card('SPADE', 2)], []],
    });
    renderWithProviders(<BristolPage />);
    // Fan 0 has 3 cards → badge shows the count.
    await waitFor(() => expect(screen.getByTestId('br-fan-count-0')).toHaveTextContent('3'));
    // Fan 1 has exactly 2 cards → boundary case, badge shows.
    expect(screen.getByTestId('br-fan-count-1')).toHaveTextContent('2');
  });

  it('hides the fan count badge for single-card fans', async () => {
    mockExec.mockResolvedValue({ ...playingState, fan: [[card('HEART', 4)], [], []] });
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Single-card fan → no badge.
    expect(screen.queryByTestId('br-fan-count-0')).not.toBeInTheDocument();
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
    screen.getByRole('button', { name: /降順ビルド列 1/ }).click();
    // Wait for the source selection to render before clicking the destination,
    // so the destination handler reads the updated `selected` state.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /降順ビルド列 1/ })).toHaveAttribute('aria-pressed', 'true'),
    );
    screen.getByRole('button', { name: /降順ビルド列 2/ }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }),
    );
  });

  it('selects a tableau column then moves it to a foundation', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: /降順ビルド列 1/ }).click();
    await waitFor(() => expect(screen.getByRole('button', { name: /^組札 0/ })).toBeEnabled());
    screen.getByRole('button', { name: /^組札 0/ }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 0 }),
    );
  });

  it('selects a fan then moves it to a foundation', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: /^ファン 0/ }).click();
    await waitFor(() => expect(screen.getByRole('button', { name: /^組札 1/ })).toBeEnabled());
    screen.getByRole('button', { name: /^組札 1/ }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'fan', col: 0 }, { zone: 'foundation', col: 1 }),
    );
  });

  it('foundations are disabled until a source is selected', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /^組札 0/ })).toBeDisabled();
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

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
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

  it('keyboard: "d" draws, "h" hints, "a" auto-completes', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('keyboard: "d" is a no-op when the stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('draw');
  });

  it('keyboard: "z" undoes only when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeEnabled());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard: "z" is a no-op when canUndo is false', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('undo');
  });

  it('keyboard: "g" opens the give-up confirm and only gives up after confirming', async () => {
    renderWithProviders(<BristolPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'g' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('advertises the keyboard shortcuts on the action buttons', async () => {
    renderWithProviders(<BristolPage />);
    const draw = await screen.findByRole('button', { name: '配る' });
    expect(draw).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(draw.querySelector('kbd')?.textContent).toBe('D');
    expect(screen.getByRole('button', { name: 'ヒント' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: '自動完成' })).toHaveAttribute('aria-keyshortcuts', 'a');
    expect(screen.getByRole('button', { name: '元に戻す' })).toHaveAttribute('aria-keyshortcuts', 'z');
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toHaveAttribute('aria-keyshortcuts', 'g');
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

    it('tableau column top card is draggable while playing', async () => {
      renderWithProviders(<BristolPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      expect(screen.getByRole('button', { name: /降順ビルド列 1/ })).toHaveAttribute('draggable', 'true');
    });

    it('dragging a tableau column onto a foundation dispatches move', async () => {
      renderWithProviders(<BristolPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const source = screen.getByRole('button', { name: /降順ビルド列 1/ });
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(source, { dataTransfer });

      const dropZone = screen.getByRole('button', { name: /^組札 0/ }).closest('div') as HTMLElement;
      mockExec.mockClear();
      fireEvent.dragOver(dropZone, { dataTransfer });
      fireEvent.drop(dropZone, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 0 }),
      );
    });

    it('dragging a fan top onto a tableau column dispatches move', async () => {
      renderWithProviders(<BristolPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const source = screen.getByRole('button', { name: /^ファン 0/ });
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(source, { dataTransfer });

      const dropZone = screen.getByRole('button', { name: /降順ビルド列 2/ }).closest('div') as HTMLElement;
      mockExec.mockClear();
      fireEvent.dragOver(dropZone, { dataTransfer });
      fireEvent.drop(dropZone, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', { zone: 'fan', col: 0 }, { zone: 'tableau', col: 1 }),
      );
    });

    it('a drop with no active drag issues no move', async () => {
      renderWithProviders(<BristolPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const dropZone = screen.getByRole('button', { name: /^組札 0/ }).closest('div') as HTMLElement;
      mockExec.mockClear();
      const dataTransfer = buildDataTransfer();
      fireEvent.drop(dropZone, { dataTransfer });

      // No source was dragged, so getData returns '' and no move is dispatched.
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
    });
  });
});
