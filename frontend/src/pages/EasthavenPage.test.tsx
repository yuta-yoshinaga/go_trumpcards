import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { easthavenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, EasthavenResponse } from '../types/card';
import { EasthavenPage } from './EasthavenPage';

vi.mock('../api/gameApi', () => ({
  easthavenApi: { exec: vi.fn() },
  actionLogApi: { easthaven: vi.fn() },
}));

const mockExec = vi.mocked(easthavenApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: EasthavenResponse = {
  tableau: [
    [
      { card: null, faceUp: false },
      { card: card('SPADE', 13), faceUp: true },
    ],
    [{ card: card('HEART', 8), faceUp: true }],
    [{ card: card('CLOVER', 5), faceUp: true }],
    [{ card: card('DIAMOND', 10), faceUp: true }],
    [{ card: card('SPADE', 3), faceUp: true }],
    [{ card: card('HEART', 7), faceUp: true }],
    [{ card: card('CLOVER', 2), faceUp: true }],
  ],
  foundation: [[], [], [], []],
  stockCount: 31,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'easthaven.playing',
};

const gameClearState: EasthavenResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'easthaven.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: EasthavenResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'easthaven.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('EasthavenPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count and stock', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
    expect(screen.getByText(/ストック/)).toBeInTheDocument();
  });

  it('deal button triggers deal command', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '配る' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('deal button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('warns on empty stock and flags columns that still hide a face-down card', async () => {
    // stock 0, column 0 still has a face-down card, columns 1+ are all face up.
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('eh-stock')).toBeInTheDocument());
    expect(screen.getByTestId('eh-stock').className).toContain('text-ds-warning');
    expect(screen.getByTestId('eh-col-header-0').className).toContain('bg-ds-warning');
    expect(screen.getByTestId('eh-col-header-1').className).not.toContain('bg-ds-warning');
  });

  it('does not flag stock or columns while the stock still has cards', async () => {
    mockExec.mockResolvedValue(playingState); // stockCount 31
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('eh-stock')).toBeInTheDocument());
    expect(screen.getByTestId('eh-stock').className).not.toContain('text-ds-warning');
    expect(screen.getByTestId('eh-col-header-0').className).not.toContain('bg-ds-warning');
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('autocomplete is disabled while stock remains', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('autocomplete triggers when all face-up and stock empty', async () => {
    const readyState: EasthavenResponse = {
      ...playingState,
      stockCount: 0,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [{ card: card('HEART', 8), faceUp: true }],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    screen.getByRole('button', { name: '自動完成' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('renders empty tableau column placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, tableau: [[], ...playingState.tableau.slice(1)] });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('renders foundation suit labels', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('selecting a card then a target column fires a tableau move', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const src = screen.getByRole('button', { name: '♠ K' }); // col0 top
    src.click();
    // The first click's selection must flush before the second click routes to target.
    await waitFor(() => expect(src.className).toContain('ring-ds-warning'));
    screen.getByRole('button', { name: '♥ 8' }).click(); // target col1 top
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: 0 }),
        expect.objectContaining({ zone: 'tableau', col: 1 }),
      ),
    );
  });

  it('toggles the empty-column aria-label when a source is selected', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [{ card: card('HEART', 8), faceUp: true }],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [], // empty column 6
      ],
    });
    renderWithProviders(<EasthavenPage />);
    const emptyCol = await screen.findByTestId('eh-empty-col-6');
    // No source selected → plain "empty" label.
    expect(emptyCol.getAttribute('aria-label')).not.toContain('ここに移動');
    // Selecting a source switches it to a "move here" prompt.
    screen.getByRole('button', { name: '♠ K' }).click();
    await waitFor(() =>
      expect(screen.getByTestId('eh-empty-col-6').getAttribute('aria-label')).toContain('ここに移動'),
    );
  });

  it('selecting a card then a foundation fires a foundation move', async () => {
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const src = screen.getByRole('button', { name: '♥ 8' }); // col1
    src.click();
    await waitFor(() => expect(src.className).toContain('ring-ds-warning'));
    screen.getByRole('button', { name: '空の組札 (♠)' }).click(); // empty ♠ foundation
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', expect.anything(), expect.objectContaining({ zone: 'foundation' })),
    );
  });

  it('double-clicking a foundation-playable top card auto-moves it to the foundation', async () => {
    const aceState: EasthavenResponse = {
      ...playingState,
      tableau: [
        [
          { card: null, faceUp: false },
          { card: card('SPADE', 13), faceUp: true },
        ],
        [{ card: card('HEART', 8), faceUp: true }],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 1), faceUp: true }], // ♠A → playable onto the empty ♠ foundation
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(aceState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(aceState);
    fireEvent.doubleClick(screen.getByRole('button', { name: '♠ A' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: 4 }),
        expect.objectContaining({ zone: 'foundation', col: 0 }),
      ),
    );
  });

  it('double-clicking a non-foundation-playable top card does nothing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // ♥8 has no legal foundation target (the empty ♥ pile needs an Ace first).
    fireEvent.doubleClick(screen.getByRole('button', { name: '♥ 8' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('renders a full from→to sentence for a to-foundation hint', async () => {
    // tableau[0][1] is ♠ K, so the sentence names the source column, card, and dest.
    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 0 } });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hint = await screen.findByTestId('eh-hint');
    expect(hint).toHaveTextContent('移動: 列 0 の ♠ K → 組札');
  });

  it('renders a full from→to sentence for a to-tableau hint', async () => {
    // toZone 'tableau' exercises the dest branch that names the destination column.
    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 3 } });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hint = await screen.findByTestId('eh-hint');
    expect(hint).toHaveTextContent('移動: 列 1 の ♥ 8 → 場札 3');
  });

  it('renders the hint sentence without a card name when the source card is face-down', async () => {
    // tableau[0][0] is a face-down (card: null) slot, hitting the empty-card-name branch.
    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 } });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hint = await screen.findByTestId('eh-hint');
    expect(hint).toHaveTextContent('移動: 列 0 の → 組札');
  });

  it('deal is guarded (no deal call) while an empty column exists', async () => {
    mockExec.mockResolvedValue({ ...playingState, tableau: [[], ...playingState.tableau.slice(1)] });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    screen.getByRole('button', { name: '配る' }).click();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('deal');
  });

  it('renders stalemate escape controls when stalemate', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('EasthavenPage keyboard shortcuts', () => {
  it.each([
    ['d', 'deal'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EasthavenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z', 'd']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
