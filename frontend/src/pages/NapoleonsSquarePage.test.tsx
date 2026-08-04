import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { napoleonsSquareApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, NapoleonsSquareResponse, NapoleonsSquareTableauCard } from '../types/card';
import { NapoleonsSquarePage } from './NapoleonsSquarePage';

vi.mock('../api/gameApi', () => ({
  napoleonsSquareApi: { exec: vi.fn() },
  actionLogApi: { napoleonssquare: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(napoleonsSquareApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(cols: NapoleonsSquareTableauCard[][]): NapoleonsSquareTableauCard[][] {
  return Array.from({ length: 12 }, (_, i) => cols[i] ?? []);
}

const aces: Card[][] = [
  [card('SPADE', 1)],
  [card('CLOVER', 1)],
  [card('HEART', 1)],
  [card('DIAMOND', 1)],
  [card('SPADE', 1)],
  [card('CLOVER', 1)],
  [card('HEART', 1)],
  [card('DIAMOND', 1)],
];

const playingState: NapoleonsSquareResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 7), faceUp: true },
      { card: card('SPADE', 6), faceUp: true },
    ],
    [{ card: card('HEART', 9), faceUp: true }],
  ]),
  stockCount: 48,
  waste: [],
  foundation: aces,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: NapoleonsSquareResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'napoleonssquare.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: NapoleonsSquareResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'napoleonssquare.gameOver',
};

describe('NapoleonsSquarePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByText(/ナポレオンズ・スクエア/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders all eight foundations', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札\d+ 1枚/).length).toBe(8));
  });

  it('labels all twelve tableau columns with their 0-based index', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    for (let i = 0; i < 12; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('shows the stock count and lets the player draw', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    const stock = await screen.findByRole('button', { name: /山札 残り48枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('disables the stock once it is exhausted', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /山札 残り0枚/ })).toBeDisabled());
  });

  // Any card can head a run, so every card is selectable — not just the top one.
  it('lets a buried card be selected as the head of a run', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    const buried = await screen.findByRole('button', { name: /^♠ 7（/ });
    expect(buried).toBeEnabled();
    fireEvent.click(buried);
    await waitFor(() => expect(buried).toHaveAttribute('aria-pressed', 'true'));
  });

  it('sends the selected run head with the move', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    const buried = await screen.findByRole('button', { name: /^♠ 7（/ });
    fireEvent.click(buried);
    await waitFor(() => expect(buried).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '空のタブロー列 4' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 4 },
      ),
    );
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
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
    mockExec.mockResolvedValue(gameOverState); // 8 aces on the foundations
    renderWithProviders(<NapoleonsSquarePage />);
    const summary = await screen.findByTestId('ns-gameover-summary');
    expect(summary).toHaveTextContent('8/104');
    expect(summary).toHaveTextContent('8%');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('ns-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete while every foundation still holds only its Ace', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses auto-complete once a foundation builds past its Ace', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], ...aces.slice(1)],
    });
    renderWithProviders(<NapoleonsSquarePage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('names each tableau card with its position for screen readers', async () => {
    // Earlier tests in this file queue one-shot resolutions and can leave CLI
    // mode persisted in localStorage; reset both so the board actually renders.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getAllByLabelText(/列\d+・上から\d+枚目/).length).toBeGreaterThan(0));
  });
});

// The waste, the hint row and CLI mode are all whole branches of the page that
// the happy-path state never reaches.
describe('NapoleonsSquarePage waste, hints and CLI mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  const withWaste: NapoleonsSquareResponse = { ...playingState, waste: [card('DIAMOND', 4)] };

  it('labels the waste as empty when there is nothing on it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('selects the waste top and sends it to a foundation', async () => {
    mockExec.mockResolvedValue(withWaste);
    renderWithProviders(<NapoleonsSquarePage />);
    const wasteCard = await screen.findByRole('button', { name: '♦ 4' });
    fireEvent.click(wasteCard);
    await waitFor(() => expect(wasteCard).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    // Foundation 3 is the diamond pile of the first deck.
    fireEvent.click(screen.getByRole('button', { name: '♦ 組札3 1枚' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', col: 3 }),
    );
  });

  it.each([
    ['foundation', { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 2 }, '組札2'],
    ['tableau', { fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 4 }, 'タブロー列4'],
    ['stock', { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 }, '捨て札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(new RegExp(expected))).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('NapoleonsSquarePage keyboard shortcuts', () => {
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
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<NapoleonsSquarePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
