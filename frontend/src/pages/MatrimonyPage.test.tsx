import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { matrimonyApi } from '../api/games/matrimony';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign } from '../types/card';
import type { MatrimonyResponse } from '../types/games/matrimony';
import { MatrimonyPage } from './MatrimonyPage';

vi.mock('../api/games/matrimony', () => ({
  matrimonyApi: { exec: vi.fn() },
}));

vi.mock('../styles/gameTheme', async () => {
  const actual = await vi.importActual<typeof import('../styles/gameTheme')>('../styles/gameTheme');
  return { ...actual, gameTheme: { ...actual.gameTheme, matrimony: actual.gameTheme.royalcotillion } };
});

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(matrimonyApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

// Sixteen slots holding one card each; slot 3 is left empty so the refill
// paths are reachable.
function makeTableau(cards: (Card | null)[]): (Card | null)[] {
  return Array.from({ length: 16 }, (_, i) => (i in cards ? cards[i] : card('DIAMOND', 7)));
}

const playingState: MatrimonyResponse = {
  tableau: makeTableau([card('SPADE', 9), card('HEART', 8), card('CLOVER', 1), null]),
  foundation: Array.from({ length: 4 }, () => []),
  stockCount: 88,
  redealCount: 0,
  waste: [],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: MatrimonyResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'matrimony.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: MatrimonyResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'matrimony.gameOver',
};

describe('MatrimonyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders four foundations', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(4));
    for (let i = 0; i < 4; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り88枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // An empty pile takes only a stock or waste card, so the label says so and a
  // tableau selection must not be droppable there.
  it('labels an empty slot with where it can be filled from', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の枠 3 (山札か捨て札から埋められます)' })).toBeInTheDocument(),
    );
  });

  // **空き山はタブローからは埋められない。**`MoveTableauToTableau` が明示的に
  // 拒否する。押せてしまうとサーバに弾かれるまで気づけない (#4906)。
  it('keeps an empty slot unclickable while a board card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    const empty = await screen.findByRole('button', { name: /空の枠 3/ });
    // まだ何も選んでいなければ、当然押せない。
    expect(empty).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '枠 0 ♠ 9' }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '枠 0 ♠ 9' })).toHaveAttribute('aria-pressed', 'true'),
    );
    // タブローの札を選んでも押せないまま。
    expect(empty).toBeDisabled();
  });

  // **ドラッグ経路も同じ規則を守る。**クリックはボタンを無効化して防いでいるが、
  // ドラッグは dispatchMove を直接通る（レビュー指摘）。
  it('ignores a tableau card dragged onto an empty slot', async () => {
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
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '枠 0 ♠ 9' })).toBeInTheDocument());
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getByRole('button', { name: '枠 0 ♠ 9' }), { dataTransfer });
    const empty = screen.getByRole('button', { name: /空の枠 3/ });
    fireEvent.dragOver(empty, { dataTransfer });
    fireEvent.drop(empty, { dataTransfer });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // 捨て札からのドラッグは通る。上の否定が「ドロップ経路そのものが死んでいる」
  // ことで通っていないかを確かめる。
  it('still accepts a waste card dragged onto an empty slot', async () => {
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
    renderWithProviders(<MatrimonyPage />);
    const wasteCard = await screen.findByRole('button', { name: '♦ 4' });
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(wasteCard, { dataTransfer });
    const empty = screen.getByRole('button', { name: /空の枠 3/ });
    fireEvent.dragOver(empty, { dataTransfer });
    fireEvent.drop(empty, { dataTransfer });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 3 }));
  });

  // The stock doubles as a move source: with a card selected it fills a gap
  // directly instead of turning to the waste.
  it('fills an empty slot straight from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り88枚/ });
    // First click draws, so select something else to enter selection mode.
    fireEvent.click(screen.getByRole('button', { name: '枠 0 ♠ 9' }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '枠 0 ♠ 9' })).toHaveAttribute('aria-pressed', 'true'),
    );
    fireEvent.click(stock);
    await waitFor(() => expect(stock).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空の枠 3/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', col: 3 }));
  });

  it('sends a pile top to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    // 読み上げ名に枠番号が付いた (#5742)。
    const ace = await screen.findByRole('button', { name: '枠 2 ♣ A' });
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
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
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

  it('shows 100% when all foundation piles are full', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: Array.from({ length: 4 }, () => Array.from({ length: 13 }, () => card('SPADE', 1))),
    });
    renderWithProviders(<MatrimonyPage />);
    const summary = await screen.findByTestId('cg-gameover-summary');
    expect(summary).toHaveTextContent('52/52');
    expect(summary).toHaveTextContent('100%');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('cg-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<MatrimonyPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['pile', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '山5'],
    ['gap fill', { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('MatrimonyPage keyboard shortcuts', () => {
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
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<MatrimonyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

// **16枠を番号で指定する設計なのに、読み上げには番号が無かった**
// (#5742)。空き枠には番号が入っていたので、埋まった瞬間に位置が読めなくなる。
describe('MatrimonyPage slot numbers in the accessible names', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    mockExec.mockResolvedValue(playingState);
  });

  it('names the slot a filled tableau card sits in', async () => {
    renderWithProviders(<MatrimonyPage />);
    expect(await screen.findByRole('button', { name: '枠 0 ♠ 9' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '枠 1 ♥ 8' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '枠 2 ♣ A' })).toBeInTheDocument();
    // 空き枠の形式は変えない (受け入れ条件3)。
    expect(screen.getByRole('button', { name: /^空の枠 3/ })).toBeInTheDocument();
  });
});
