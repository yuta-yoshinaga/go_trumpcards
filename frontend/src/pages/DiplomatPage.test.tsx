import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { diplomatApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, DiplomatResponse } from '../types/card';
import { DiplomatPage } from './DiplomatPage';

vi.mock('../api/gameApi', () => ({
  diplomatApi: { exec: vi.fn() },
  actionLogApi: { diplomat: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(diplomatApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(piles: Card[][]): Card[][] {
  return Array.from({ length: 8 }, (_, i) => piles[i] ?? []);
}

const playingState: DiplomatResponse = {
  tableau: makeTableau([[card('SPADE', 9)], [card('HEART', 8)], [card('CLOVER', 1)]]),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 72,
  waste: [],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: DiplomatResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'diplomat.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: DiplomatResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'diplomat.gameOver',
};

describe('DiplomatPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByText(/ディプロマット/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り72枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // An empty pile takes only a stock or waste card, so the label says so and a
  // tableau selection must not be droppable there.
  it('labels an empty column as taking any card', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の列 3 (どのカードでも置けます)' })).toBeInTheDocument(),
    );
  });

  // **空き列はタブローからも埋められる。**Congress と違って列を移動元にできる
  // のが Diplomat の主要な逃げ道なので、押せなければならない。
  it('enables an empty column once a tableau card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    const empty = await screen.findByRole('button', { name: /空の列 3/ });
    // まだ何も選んでいなければ押せない。
    expect(empty).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    expect(empty).toBeEnabled();
  });

  // **ドラッグ経路も同じ規則に従う。**クリックとドラッグで結果が変わらないこと
  // を確かめる（ドラッグは dispatchMove を直接通る）。
  it('accepts a tableau card dragged onto an empty column', async () => {
    const buildDataTransfer = () => {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    };

    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toBeInTheDocument());
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getByRole('button', { name: '♠ 9' }), { dataTransfer });
    const empty = screen.getByRole('button', { name: /空の列 3/ });
    fireEvent.dragOver(empty, { dataTransfer });
    fireEvent.drop(empty, { dataTransfer });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }),
    );
  });

  // 捨て札からも同じように置ける。
  it('accepts a waste card dragged onto an empty column', async () => {
    const buildDataTransfer = () => {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    };

    mockExec.mockResolvedValue({ ...playingState, waste: [card('DIAMOND', 4)] });
    renderWithProviders(<DiplomatPage />);
    const wasteCard = await screen.findByRole('button', { name: '♦ 4' });
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(wasteCard, { dataTransfer });
    const empty = screen.getByRole('button', { name: /空の列 3/ });
    fireEvent.dragOver(empty, { dataTransfer });
    fireEvent.drop(empty, { dataTransfer });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 3 }));
  });

  // **山札は移動元にならない。**空き列は列か捨て札から埋めるので、カードを
  // 選んだ状態で山札を押しても「めくる」以外は起きない。
  it('never uses the stock as a move source', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り72枚/ });

    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    expect(mockExec).not.toHaveBeenCalledWith('move', { zone: 'stock' }, expect.anything());
  });

  it('sends a pile top to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
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
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
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
    renderWithProviders(<DiplomatPage />);
    const summary = await screen.findByTestId('cg-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('cg-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<DiplomatPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['column', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '列5'],
    // The stock only ever draws here, so its hint names the stock and no pile.
    ['draw', { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('DiplomatPage keyboard shortcuts', () => {
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
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<DiplomatPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
