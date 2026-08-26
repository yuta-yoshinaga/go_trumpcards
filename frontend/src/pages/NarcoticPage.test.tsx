import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, narcoticApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, NarcoticCard, NarcoticResponse } from '../types/card';
import { NarcoticPage } from './NarcoticPage';

vi.mock('../api/gameApi', () => ({
  narcoticApi: { exec: vi.fn() },
  actionLogApi: { narcotic: vi.fn() },
}));

const mockExec = vi.mocked(narcoticApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeCard(c: Card, opts: Partial<Omit<NarcoticCard, 'card'>> = {}): NarcoticCard {
  return { card: c, top: opts.top ?? false, removable: opts.removable ?? false, movable: opts.movable ?? false };
}

/**
 * col0: ♠5, col1: ♠9 (stackable onto a matching pile to its left), col2: empty, col3: ♦6.
 *
 * **removable は盤面全体の性質。**露出4枚が揃ったときだけ真で、そのときは4列とも
 * 真になる。ここでは揃っていないので全列 false。
 */
function makeColumns(): NarcoticCard[][] {
  return [
    [makeCard(card('SPADE', 5), { top: true })],
    [makeCard(card('SPADE', 9), { top: true, movable: true })],
    [],
    [makeCard(card('DIAMOND', 6), { top: true })],
  ];
}

/** 露出4枚がすべて 7: 揃っているので4列とも removable。 */
function matchedColumns(): NarcoticCard[][] {
  return [
    [makeCard(card('SPADE', 7), { top: true, removable: true })],
    [makeCard(card('HEART', 7), { top: true, removable: true })],
    [makeCard(card('CLOVER', 7), { top: true, removable: true })],
    [makeCard(card('DIAMOND', 7), { top: true, removable: true })],
  ];
}

const playingState: NarcoticResponse = {
  columns: makeColumns(),
  stockCount: 44,
  discardCount: 4,
  redealCount: 0,
  discardTop: card('CLOVER', 7),
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

const gameClearState: NarcoticResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'narcotic.gameClear',
  messageParams: { moveCount: '20' },
};

const gameOverState: NarcoticResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'narcotic.gameOver',
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
});

describe('NarcoticPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<NarcoticPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders stock count', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(44\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 3/));
  });

  it('renders the four columns', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    const colDivs = document.querySelectorAll('[data-tutorial="narcotic-columns"] > div');
    expect(colDivs.length).toBe(4);
  });

  it('renders the discard pile with progress and the top card', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-discard-pile')).toBeInTheDocument());
    // Progress readout: discardCount out of the 48-card goal.
    expect(screen.getByText(/捨て札/)).toBeInTheDocument();
    expect(screen.getByText(/\(4\/48\)/)).toBeInTheDocument();
    // The most recently removed card is shown face-up.
    expect(screen.getByTestId('narcotic-discard-top')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /♣ 7/ })).toBeInTheDocument();
  });

  it('renders an empty discard placeholder when nothing has been removed', async () => {
    mockExec.mockResolvedValue({ ...playingState, discardCount: 0, discardTop: null });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-discard-pile')).toBeInTheDocument());
    expect(screen.getByTestId('narcotic-discard-empty')).toBeInTheDocument();
    expect(screen.getByText(/\(0\/48\)/)).toBeInTheDocument();
    expect(screen.queryByTestId('narcotic-discard-top')).not.toBeInTheDocument();
  });

  it('renders empty column placeholder', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking deal button dispatches draw', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Both the stock card-back and the footer button expose the "配る" label.
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // **札のクリックは「重ねる」。**捨てるのは4枚まとまりなので別ボタン。
  it('clicking a stackable top card dispatches move', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getAllByRole('button', { name: /♠ 9/ })[0]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  // **列を渡さない。**4枚まとめて捨てるので、選ぶ余地が無い。
  it('clicking discard dispatches remove with no column', async () => {
    mockExec.mockResolvedValue({ ...playingState, columns: matchedColumns() });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-remove-all')).toBeEnabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('narcotic-remove-all'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove'));
  });

  it('disables discard when the four ranks do not match', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-remove-all')).toBeInTheDocument());
    expect(screen.getByTestId('narcotic-remove-all')).toBeDisabled();
  });

  // **クローン元には無いボタン。**山札が尽きても場を集めれば続けられる。
  it('offers redeal only once the stock is spent', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-redeal')).toBeInTheDocument());
    expect(screen.getByTestId('narcotic-redeal')).toBeDisabled();
  });

  it('clicking redeal dispatches redeal', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('narcotic-redeal')).toBeEnabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('narcotic-redeal'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  // **落とし先は空の山ではない。**Narcotic の行き先は「同ランクを露出している
  // 最も左の山」で、空の山は何も露出していないので絶対に合法にならない。
  // クローン元 (Aces Up) は空き列へドロップさせる。
  it('dragging a stackable card onto the destination pile dispatches move', async () => {
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

    mockExec.mockResolvedValue({ ...playingState, columns: matchedColumns() });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /♥ 7/ })[0]).toBeInTheDocument());

    mockExec.mockClear();
    const dataTransfer = buildDataTransfer();
    const source = screen.getAllByRole('button', { name: /♥ 7/ })[0];
    const target = screen.getAllByRole('button', { name: /♠ 7/ })[0];
    if (!source || !target) throw new Error('drag fixtures missing');

    fireEvent.dragStart(source, { dataTransfer });
    fireEvent.dragOver(target, { dataTransfer });
    fireEvent.drop(target, { dataTransfer });

    // **動かすのは掴んだ山 (col1 の ♥7)。**行き先はサーバが「同ランクを露出して
    // いる最も左の山」として決めるので、送るのは掴んだ列番号だけ。
    // expect.any(Number) では別の山を動かしても通ってしまう。
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  it('offers no drop zone on an emptied pile', async () => {
    renderWithProviders(<NarcoticPage />);
    const empty = await screen.findByTestId('narcotic-empty-2');
    // 空の山は presentation の div のままで、ドロップを受ける region にならない。
    expect(empty.closest('[role="presentation"]')).toBeNull();
  });

  it('clicking giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking undo button dispatches undo', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    // **除去は列を名指ししない。**4枚まとめてなので col は -1。
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'remove', col: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByText(/4枚のランクが揃っている/)).toBeInTheDocument());
  });

  it('renders a move hint via HintTooltip with the column', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'move', col: 2 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // **「空き列へ移動」ではない。**行き先は同ランクを出している左の山。
    await waitFor(() => expect(screen.getByText(/列\[2\].*左の山へ重ね/)).toBeInTheDocument());
    expect(screen.queryByText(/空き列/)).not.toBeInTheDocument();
  });

  it('renders a draw hint via HintTooltip without a column', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'draw', col: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/4つの山へ1枚ずつ配ります/)).toBeInTheDocument());
  });

  // **クローン元には無いヒント。**山札が尽きても場を集めれば続けられる。
  it('renders a redeal hint via HintTooltip', async () => {
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, hint: { type: 'redeal', col: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/集めれば配り直せます/)).toBeInTheDocument());
  });

  it('renders game clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
  });

  it('renders game over state and hides action buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
  });

  it('disables undo button when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('disables deal button when stock empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('renders stalemate message', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      undoToEscape: 1,
      messageCode: 'narcotic.stalemate',
      message: '手詰まりです。',
    });
    renderWithProviders(<NarcoticPage />);
    // messageCode の訳が勝つ (#5291)。サーバの message はフォールバック。
    // **文面はループ検出を指す。**「元に戻すかギブアップ」はクローン元の文面で、
    // 配り直しが無制限にあるこのゲームでは誤った案内になる。
    await waitFor(() =>
      expect(screen.getByText('同じ盤面が繰り返しています。これ以上は進みません')).toBeInTheDocument(),
    );
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('NarcoticPage keyboard shortcuts', () => {
  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<NarcoticPage />);
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
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<NarcoticPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
