import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { royalcotillionApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, RoyalCotillionResponse } from '../types/card';
import { RoyalCotillionPage } from './RoyalCotillionPage';

vi.mock('../api/gameApi', () => ({
  royalcotillionApi: { exec: vi.fn() },
  actionLogApi: { royalcotillion: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(royalcotillionApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

// Sixteen slots holding one card each; slot 3 is left empty so the refill
// paths are reachable.
function makeTableau(cards: (Card | null)[]): (Card | null)[] {
  return Array.from({ length: 16 }, (_, i) => (i in cards ? cards[i] : card('DIAMOND', 7)));
}

const playingState: RoyalCotillionResponse = {
  tableau: makeTableau([card('SPADE', 9), card('HEART', 8), card('CLOVER', 1), null]),
  reserve: Array.from({ length: 4 }, (_, i) => [card('HEART', i + 2)]),
  foundationOdd: [true, true, true, true, false, false, false, false],
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 76,
  waste: [],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: RoyalCotillionResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'royalcotillion.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: RoyalCotillionResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'royalcotillion.gameOver',
};

describe('RoyalCotillionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByText(/ロイヤルコティヨン/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り76枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // An empty pile takes only a stock or waste card, so the label says so and a
  // tableau selection must not be droppable there.
  it('labels an empty slot with where it can be filled from', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の枠 3 (山札か捨て札から埋められます)' })).toBeInTheDocument(),
    );
  });

  // **空き山はタブローからは埋められない。**`MoveTableauToTableau` が明示的に
  // 拒否する。押せてしまうとサーバに弾かれるまで気づけない (#4906)。
  it('keeps an empty slot unclickable while a board card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
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
    renderWithProviders(<RoyalCotillionPage />);
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
    renderWithProviders(<RoyalCotillionPage />);
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
  // The reserve is the other source, and it can only ever go up. Emptying one
  // costs a slot for good, which is the game's central trade.
  it('sends a reserve top to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    // 読み上げ名に山番号が付いた (#5742)。番号ごと指定して取り違えを防ぐ。
    const reserveCard = await screen.findByRole('button', { name: 'リザーブ 0 ♥ 2（一番上）' });
    fireEvent.click(reserveCard);
    await waitFor(() => expect(reserveCard).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getAllByRole('button', { name: /組札/ })[0]);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 0 }, { zone: 'foundation', col: 0 }),
    );
  });

  // Half the foundations start at the Ace and half at the deuce; without the
  // marker the board does not say what a pile wants next.
  it('marks which foundations start at the Ace and which at the deuce', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getAllByText(/^A:/).length).toBe(4));
    expect(screen.getAllByText(/^2:/).length).toBe(4);
  });

  it('ignores a reserve card dragged onto an empty slot', async () => {
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
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リザーブ 0 ♥ 2（一番上）' })).toBeInTheDocument());
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getAllByRole('button', { name: 'リザーブ 0 ♥ 2（一番上）' })[0], { dataTransfer });
    const empty = screen.getByRole('button', { name: /空の枠 3/ });
    fireEvent.dragOver(empty, { dataTransfer });
    fireEvent.drop(empty, { dataTransfer });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('fills an empty slot straight from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り76枚/ });
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
    renderWithProviders(<RoyalCotillionPage />);
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
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
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
    renderWithProviders(<RoyalCotillionPage />);
    const summary = await screen.findByTestId('cg-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('cg-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<RoyalCotillionPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['pile', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '山5'],
    ['gap fill', { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('RoyalCotillionPage keyboard shortcuts', () => {
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
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<RoyalCotillionPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

// **16 枠 4 リザーブを番号で指定する設計なのに、読み上げには番号が無かった**
// (#5742)。空き枠には番号が入っていたので、埋まった瞬間に位置が読めなくなる。
describe('RoyalCotillionPage slot numbers in the accessible names', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    mockExec.mockResolvedValue(playingState);
  });

  it('names the slot a filled tableau card sits in', async () => {
    renderWithProviders(<RoyalCotillionPage />);
    expect(await screen.findByRole('button', { name: '枠 0 ♠ 9' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '枠 1 ♥ 8' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '枠 2 ♣ A' })).toBeInTheDocument();
    // 空き枠の形式は変えない (受け入れ条件3)。
    expect(screen.getByRole('button', { name: /^空の枠 3/ })).toBeInTheDocument();
  });

  it('names the reserve pile a card sits in, and says which card is on top', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [[card('SPADE', 3), card('HEART', 2)], [], [], []],
    });
    renderWithProviders(<RoyalCotillionPage />);

    expect(await screen.findByRole('button', { name: 'リザーブ 0 ♥ 2（一番上）' })).toBeInTheDocument();
    // 埋もれた札も山番号で位置が分かる。押せないことは disabled が示す。
    const buried = screen.getByRole('button', { name: 'リザーブ 0 ♠ 3（下に埋まっています）' });
    expect(buried).toBeDisabled();
    // 空のリザーブの読み上げは従来のまま。
    expect(screen.getByText(/空のリザーブ 1/)).toBeInTheDocument();
  });
});
