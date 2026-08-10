import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { duchessApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, DuchessResponse, DuchessTableauCard } from '../types/card';
import { DuchessPage } from './DuchessPage';

vi.mock('../api/gameApi', () => ({
  duchessApi: { exec: vi.fn() },
  actionLogApi: { duchess: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(duchessApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(cols: DuchessTableauCard[][]): DuchessTableauCard[][] {
  return Array.from({ length: 4 }, (_, i) => cols[i] ?? []);
}

const playingState: DuchessResponse = {
  reserve: [[card('CLOVER', 2)], [card('DIAMOND', 7)], [], []],
  tableau: makeTableau([
    [
      { card: card('SPADE', 9), faceUp: true },
      { card: card('HEART', 8), faceUp: true },
    ],
    [{ card: card('CLOVER', 4), faceUp: true }],
  ]),
  foundation: Array.from({ length: 4 }, () => []),
  stockCount: 35,
  waste: [],
  baseRank: 5,
  awaitingBaseRank: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const awaitingBaseState: DuchessResponse = {
  ...playingState,
  reserve: [[card('CLOVER', 2)], [card('DIAMOND', 7)], [card('SPADE', 3)], [card('HEART', 6)]],
  baseRank: 0,
  awaitingBaseRank: true,
  moveCount: 0,
};

const gameClearState: DuchessResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'duchess.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: DuchessResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'duchess.gameOver',
};

describe('DuchessPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading, base rank and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByText(/ダッチェス/)).toBeInTheDocument());
    expect(screen.getByText(/開始ランク: 5/)).toBeInTheDocument();
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders four foundations and four columns', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(4));
    for (let i = 0; i < 4; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り35枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // Until the base rank is set nothing else is legal, so the reserve fan is a
  // rank picker rather than a move source, and everything else is barred.
  it('turns the reserve fans into base-rank pickers before the rank is set', async () => {
    mockExec.mockResolvedValue(awaitingBaseState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByText(/開始ランク: 未決定/)).toBeInTheDocument());
    expect(screen.getByText(/開始ランクになります/)).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リザーブ扇 2 の札を開始ランクにする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('base', { zone: 'reserve', col: 2 }));
  });

  it('blocks drawing until the base rank is chosen', async () => {
    mockExec.mockResolvedValue(awaitingBaseState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled());
    expect(screen.getByRole('button', { name: /山札 残り35枚/ })).toBeDisabled();
  });

  it('selects a reserve top as a move source once the rank is set', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    const fanTop = await screen.findByRole('button', { name: /^♣ 2（扇/ });
    fireEvent.click(fanTop);
    await waitFor(() => expect(fanTop).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /タブロー列 2/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 0 }, { zone: 'tableau', col: 2 }),
    );
  });

  it('renders an empty reserve fan as a non-interactive slot', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByLabelText('リザーブ扇 2 は空です')).toBeInTheDocument());
  });

  // The reserve-only rule is the heart of the game, so an empty column has to
  // say which of the two states it is in.
  it('marks an empty column reserve-only while the reserve remains', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<DuchessPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /空のタブロー列 2 \(リザーブからのみ置けます\)/ })).toBeInTheDocument(),
    );
    unmount();

    mockExec.mockResolvedValue({ ...playingState, reserve: [[], [], [], []] });
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '空のタブロー列 2' })).toBeInTheDocument());
  });

  it('selects the waste top and moves it', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('DIAMOND', 4)] });
    renderWithProviders(<DuchessPage />);
    const wasteTop = await screen.findByRole('button', { name: '♦ 4' });
    fireEvent.click(wasteTop);
    await waitFor(() => expect(wasteTop).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /タブロー列 2/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 2 }));
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('lets a buried card be selected as the head of a run', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    const buried = await screen.findByRole('button', { name: /^♠ 9（/ });
    expect(buried).toBeEnabled();
    fireEvent.click(buried);
    await waitFor(() => expect(buried).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /タブロー列 3/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 3 },
      ),
    );
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);

    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('counts the summary against a single 52-card deck', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: [[card('SPADE', 5)], [], [], []],
    });
    renderWithProviders(<DuchessPage />);
    const summary = await screen.findByTestId('du-gameover-summary');
    expect(summary).toHaveTextContent('1/52');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('du-gameover-summary')).not.toBeInTheDocument();
  });

  // Every foundation opens with the base rank, so one card on a pile is the
  // starting position — auto-complete only helps past it.
  it('disables auto-complete until a foundation is past the base rank', async () => {
    mockExec.mockResolvedValue({ ...playingState, foundation: [[card('SPADE', 5)], [], [], []] });
    const { unmount } = renderWithProviders(<DuchessPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 5), card('SPADE', 6)], [], [], []],
    });
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, cardIndex: 0, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['reserve', { fromZone: 'reserve', fromIdx: 1, cardIndex: -1, toZone: 'tableau', toIdx: 3 }, 'リザーブ扇1'],
    ['waste', { fromZone: 'waste', fromIdx: -1, cardIndex: -1, toZone: 'foundation', toIdx: 0 }, '捨て札'],
    ['draw', { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(new RegExp(expected))).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });

  it('names each tableau card with its position for screen readers', async () => {
    // Earlier tests in this file queue one-shot resolutions and can leave CLI
    // mode persisted in localStorage; reset both so the board actually renders.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/列\d+・上から\d+枚目/).length).toBeGreaterThan(0));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('DuchessPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<DuchessPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
