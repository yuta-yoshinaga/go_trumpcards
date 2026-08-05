import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { congressApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CongressResponse } from '../types/card';
import { CongressPage } from './CongressPage';

vi.mock('../api/gameApi', () => ({
  congressApi: { exec: vi.fn() },
  actionLogApi: { congress: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(congressApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(piles: Card[][]): Card[][] {
  return Array.from({ length: 8 }, (_, i) => piles[i] ?? []);
}

const playingState: CongressResponse = {
  tableau: makeTableau([[card('SPADE', 9)], [card('HEART', 8)], [card('CLOVER', 1)]]),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 96,
  waste: [],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: CongressResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'congress.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: CongressResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'congress.gameOver',
};

describe('CongressPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByText(/コングレス/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り96枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // An empty pile takes only a stock or waste card, so the label says so and a
  // tableau selection must not be droppable there.
  it('labels an empty pile as stock-or-waste only', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の山 3 (山札か捨て札からのみ置けます)' })).toBeInTheDocument(),
    );
  });

  // **空き山はタブローからは埋められない。**`MoveTableauToTableau` が明示的に
  // 拒否する。押せてしまうとサーバに弾かれるまで気づけない (#4906)。
  it('disables an empty pile while a tableau card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    const empty = await screen.findByRole('button', { name: /空の山 3/ });
    // まだ何も選んでいなければ、当然押せない。
    expect(empty).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    // タブローの札を選んでも押せないまま。
    expect(empty).toBeDisabled();
  });

  // The stock doubles as a move source: with a card selected it fills a gap
  // directly instead of turning to the waste.
  it('fills an empty pile straight from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り96枚/ });
    // First click draws, so select something else to enter selection mode.
    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(stock);
    await waitFor(() => expect(stock).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空の山 3/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', col: 3 }));
  });

  it('sends a pile top to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    const ace = await screen.findByRole('button', { name: '♣ A' });
    fireEvent.click(ace);
    await waitFor(() => expect(ace).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空の組札1/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 1 }),
    );
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
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

  it('counts the summary against 104 cards, not 52', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<CongressPage />);
    const summary = await screen.findByTestId('cg-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('cg-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<CongressPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['pile', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '山5'],
    ['gap fill', { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('CongressPage keyboard shortcuts', () => {
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
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CongressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
