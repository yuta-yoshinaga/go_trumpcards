import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { americanToadApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AmericanToadResponse, AmericanToadTableauCard, Card, CardDesign } from '../types/card';
import { AmericanToadPage } from './AmericanToadPage';

vi.mock('../api/gameApi', () => ({
  americanToadApi: { exec: vi.fn() },
  actionLogApi: { americantoad: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(americanToadApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(cols: AmericanToadTableauCard[][]): AmericanToadTableauCard[][] {
  return Array.from({ length: 8 }, (_, i) => cols[i] ?? []);
}

const playingState: AmericanToadResponse = {
  reserve: [card('CLOVER', 3), card('DIAMOND', 7)],
  tableau: makeTableau([
    [
      { card: card('SPADE', 9), faceUp: true },
      { card: card('SPADE', 8), faceUp: true },
    ],
    [{ card: card('CLOVER', 4), faceUp: true }],
  ]),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 75,
  waste: [],
  baseRank: 5,
  passesUsed: 0,
  canRedeal: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: AmericanToadResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'americantoad.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: AmericanToadResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'americantoad.gameOver',
};

describe('AmericanToadPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading, base rank and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByText(/アメリカン・トード/)).toBeInTheDocument());
    expect(screen.getByText(/開始ランク: 5/)).toBeInTheDocument();
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight columns', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り75枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // The stock button becomes the redeal once the stock is out, and there is
  // only one, so it is labelled and highlighted differently.
  it('turns the stock into a redeal button when the stock is out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canRedeal: true, waste: [card('HEART', 9)] });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByText(/めくり直しはあと1回/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /捨て札を山札に戻す/ })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'めくり直す' })).toBeEnabled();
  });

  it('disables the stock once the redeal is spent too', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canRedeal: false, passesUsed: 1 });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しも使い切りました/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('selects the reserve top and moves it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    const reserve = await screen.findByRole('button', { name: 'リザーブ 残り2枚' });
    fireEvent.click(reserve);
    await waitFor(() => expect(reserve).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空の組札0/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'foundation', col: 0 }),
    );
  });

  // While the reserve holds cards an empty column belongs to it and refills
  // automatically, so the player must not be able to drop anything there.
  it('locks an empty column to the reserve until the reserve is gone', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<AmericanToadPage />);
    const locked = await screen.findByRole('button', {
      name: '空のタブロー列 2 (リザーブから自動で埋まります)',
    });
    expect(locked).toBeDisabled();
    unmount();

    mockExec.mockResolvedValue({ ...playingState, reserve: [], waste: [card('HEART', 8)] });
    renderWithProviders(<AmericanToadPage />);
    const open = await screen.findByRole('button', { name: '空のタブロー列 2 (捨て札から埋められます)' });
    // Still disabled with nothing selected; selecting the waste enables it.
    fireEvent.click(screen.getByRole('button', { name: '♥ 8' }));
    await waitFor(() => expect(open).toBeEnabled());

    mockExec.mockClear();
    fireEvent.click(open);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 2 }));
  });

  it('shows an empty reserve slot once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, reserve: [] });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByLabelText('リザーブは空です')).toBeInTheDocument());
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('lets a buried card be selected as the head of a run', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    const buried = await screen.findByRole('button', { name: /^♠ 9（/ });
    expect(buried).toBeEnabled();
    fireEvent.click(buried);
    await waitFor(() => expect(buried).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /^♣ 4（/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 1 },
      ),
    );
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
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
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<AmericanToadPage />);
    const summary = await screen.findByTestId('at-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('at-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<AmericanToadPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, cardIndex: 0, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['reserve', { fromZone: 'reserve', fromIdx: -1, cardIndex: -1, toZone: 'tableau', toIdx: 3 }, 'タブロー列3'],
    ['draw', { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
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
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/列\d+・上から\d+枚目/).length).toBeGreaterThan(0));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('AmericanToadPage keyboard shortcuts', () => {
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
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AmericanToadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
