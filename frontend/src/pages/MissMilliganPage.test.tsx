import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { missMilliganApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, MissMilliganResponse, MissMilliganTableauCard } from '../types/card';
import { MissMilliganPage } from './MissMilliganPage';

vi.mock('../api/gameApi', () => ({
  missMilliganApi: { exec: vi.fn() },
  actionLogApi: { missmilligan: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(missMilliganApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(cols: MissMilliganTableauCard[][]): MissMilliganTableauCard[][] {
  return Array.from({ length: 8 }, (_, i) => cols[i] ?? []);
}

const playingState: MissMilliganResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 9), faceUp: true },
      { card: card('HEART', 8), faceUp: true },
    ],
    [{ card: card('CLOVER', 4), faceUp: true }],
  ]),
  stockCount: 96,
  foundation: Array.from({ length: 8 }, () => []),
  waived: [],
  canWaive: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: MissMilliganResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'missmilligan.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: MissMilliganResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'missmilligan.gameOver',
};

describe('MissMilliganPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByText(/ミス・ミリガン/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight columns', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('deals a row from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り96枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  // Dealing while holding could bury the only square the held cards fit, so the
  // domain refuses it — the UI must not offer it either.
  it('blocks dealing while cards are waived', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, waived: [card('HEART', 8)] });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配り足す' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '配り足す' })).toHaveAttribute('title');
  });

  it('labels an empty column as Kings-only', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /空のタブロー列 2 \(キングのみ置けます\)/ })).toBeInTheDocument(),
    );
  });

  // The waive control only exists once the stock is gone, and only on columns
  // that actually hold something.
  it('offers waiving per column only when it is available', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /の連番を保持する/ })).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canWaive: true });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /の連番を保持する/ }).length).toBe(2));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'タブロー列 0 の連番を保持する' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('waive', { zone: 'tableau', col: 0, cardIndex: undefined }),
    );
  });

  it('shows the held cards and moves them back through the move API', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, waived: [card('HEART', 8)] });
    renderWithProviders(<MissMilliganPage />);
    const held = await screen.findByRole('button', { name: /保持中の札 1枚/ });
    fireEvent.click(held);
    await waitFor(() => expect(held).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空のタブロー列 2/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waived' }, { zone: 'tableau', col: 2 }));
  });

  it('announces when waiving is available but nothing is held', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canWaive: true });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByLabelText(/保持できます/)).toBeInTheDocument());
  });

  it('lets a buried card be selected as the head of a run', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    const buried = await screen.findByRole('button', { name: '♠ 9' });
    expect(buried).toBeEnabled();
    fireEvent.click(buried);
    await waitFor(() => expect(buried).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空のタブロー列 3/ }));
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
    const { unmount } = renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
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
    renderWithProviders(<MissMilliganPage />);
    const summary = await screen.findByTestId('mm-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('mm-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until an Ace reaches a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<MissMilliganPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['waived', { fromZone: 'waived', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: 3 }, 'タブロー列3'],
    ['deal', { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: -1 }, '各列へ配り足す'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(new RegExp(expected))).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('MissMilliganPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it.each([
    ['d', 'deal'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<MissMilliganPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
