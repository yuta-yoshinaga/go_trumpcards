import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { terraceApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TerraceResponse } from '../types/card';
import { TerracePage } from './TerracePage';

vi.mock('../api/gameApi', () => ({
  terraceApi: { exec: vi.fn() },
  actionLogApi: { terrace: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(terraceApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(piles: Card[][]): Card[][] {
  return Array.from({ length: 9 }, (_, i) => piles[i] ?? []);
}

const playingState: TerraceResponse = {
  reserve: [card('CLOVER', 3), card('SPADE', 9)],
  tableau: makeTableau([[card('HEART', 8)], [card('SPADE', 7)]]),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 84,
  waste: [],
  baseRank: 5,
  awaitingBaseRank: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const awaitingBaseState: TerraceResponse = {
  ...playingState,
  baseRank: 0,
  awaitingBaseRank: true,
  moveCount: 0,
};

const gameClearState: TerraceResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'terrace.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: TerraceResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'terrace.gameOver',
};

describe('TerracePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading, base rank and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    // "テラス" is both the page title and the zone label, so target the heading.
    await waitFor(() => expect(screen.getByRole('heading', { name: 'テラス' })).toBeInTheDocument());
    expect(screen.getByText(/開始ランク: 5/)).toBeInTheDocument();
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and nine piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 9; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  // Nothing reaches a foundation until the rank is fixed, so the board says so.
  it('announces that the base rank is still open', async () => {
    mockExec.mockResolvedValue(awaitingBaseState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByText(/開始ランク: 未決定/)).toBeInTheDocument());
    expect(screen.getByText(/開始ランクになります/)).toBeInTheDocument();
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    const stock = await screen.findByRole('button', { name: /山札 残り84枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // The terrace's only destination is a foundation, and its depth matters
  // because it is never refilled.
  it('sends the terrace top to a foundation and shows its depth', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    const terrace = await screen.findByRole('button', { name: 'テラス 残り2枚（組札にのみ出せます）' });
    fireEvent.click(terrace);
    await waitFor(() => expect(terrace).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '空の組札0' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'foundation', col: 0 }),
    );
  });

  it('shows an empty terrace slot once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, reserve: [] });
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByLabelText('テラスは空です')).toBeInTheDocument());
  });

  // An empty pile refills itself, so it must not be an interactive target.
  it('renders an empty pile as a non-interactive slot', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByLabelText('空の山 3')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '空の山 3' })).not.toBeInTheDocument();
  });

  it('moves one card between piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    const red8 = await screen.findByRole('button', { name: '♥ 8' });
    fireEvent.click(red8);
    await waitFor(() => expect(red8).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }),
    );
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
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
    renderWithProviders(<TerracePage />);
    const summary = await screen.findByTestId('tr-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('tr-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<TerracePage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['terrace', { fromZone: 'reserve', fromIdx: -1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['pile', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '山5'],
    ['draw', { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('TerracePage keyboard shortcuts', () => {
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
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TerracePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
