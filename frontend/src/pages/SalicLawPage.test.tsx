import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { salicLawApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SalicLawResponse } from '../types/card';
import { SalicLawPage } from './SalicLawPage';

vi.mock('../api/gameApi', () => ({
  salicLawApi: { exec: vi.fn() },
  actionLogApi: { saliclaw: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(salicLawApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

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

function makeTableau(piles: Card[][]): Card[][] {
  return Array.from({ length: 8 }, (_, i) => piles[i] ?? []);
}

// **どの開いた列も底が K。**クローン元の Congress は 1 枚ずつ配るので K の無い
// 列があり得たが、Salic Law でその盤は出ない。列2 は K だけ ＝ 唯一の置き場所。
const playingState: SalicLawResponse = {
  tableau: makeTableau([
    [card('SPADE', 13), card('SPADE', 9)],
    [card('HEART', 13), card('CLOVER', 1)],
    [card('DIAMOND', 13)],
  ]),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 95,
  queens: [card('SPADE', 12), card('HEART', 12)],
  openPiles: 3,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: SalicLawResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'saliclaw.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: SalicLawResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'saliclaw.gameOver',
};

describe('SalicLawPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByText(/サリカ法典/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations and eight piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    for (let i = 0; i < 8; i++) {
      // 組札は F0..F7、列は #0..#7。同じ見出しにすると
      // getByText がどちらを指すか決まらない。
      expect(screen.getByText(`F${i}`)).toBeInTheDocument();
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り95枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // **抜いた Q を見せる。**8 枚消えている理由は盤からは読めない。
  it('shows the queens that are out of play', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const queens = await screen.findByTestId('sl-queens');
    expect(queens.children).toHaveLength(2);
  });

  // **まだ K が出ていない列は置き場所ではない。**Congress の「空き山」とは
  // 別物で、配りが進むまで存在しないのと同じ。
  it('marks a column with no king yet as not open, and never enables it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const unopened = await screen.findByTestId('sl-unopened-3');
    expect(unopened).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    // 札を選んでも押せないまま。
    expect(unopened).toBeDisabled();
  });

  // **K だけの列がこのゲーム唯一の置き場所。**選択中の札があるときだけ押せる。
  it('accepts a card onto a column holding just its king', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const bareKing = await screen.findByTestId('sl-bare-king-2');
    // 何も選んでいなければ押せない (土台の K 自体は動かせない)。
    expect(bareKing).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();
    fireEvent.click(bareKing);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 2 }),
    );
  });

  // **クリック経路も置き先を絞る（レビュー指摘）。**`dispatchMove` の再検査は
  // ドラッグしか通らない。クリックは `handleSelectTarget` からフックへ直行するので、
  // ボタン側で止めないと、K だけでない列を押した瞬間にサーバが必ず拒む move が飛ぶ。
  it('ignores a click on a column that is not a bare king while a source is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    // 列1 の上札 (♣A) を選ぶ。
    const ace = await screen.findByRole('button', { name: '♣ A' });
    fireEvent.click(ace);
    await waitFor(() => expect(ace).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    // 列0 の上札 (♠9) は K だけの列ではないので、置き先にならない。
    const notATarget = screen.getByRole('button', { name: '♠ 9' });
    expect(notATarget).toBeDisabled();
    fireEvent.click(notATarget);

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // 選択中の列をもう一度押したら解除できること。上の禁止で「押せない」だけに
  // すると、選び直す手段がボタンから消える。
  it('lets the selected column be clicked again to deselect', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const ace = await screen.findByRole('button', { name: '♣ A' });
    fireEvent.click(ace);
    await waitFor(() => expect(ace).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(ace);
    await waitFor(() => expect(ace).toHaveAttribute('aria-pressed', 'false'));
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // **ドラッグ経路も同じ規則を守る。**クリックはボタンを無効化して防いでいるが、
  // ドラッグは dispatchMove を直接通る（レビュー指摘）。まだ K が出ていない列は
  // 置き先ではない。
  it('ignores a card dragged onto a column with no king yet', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toBeInTheDocument());
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getByRole('button', { name: '♠ 9' }), { dataTransfer });
    const unopened = screen.getByTestId('sl-unopened-3');
    fireEvent.dragOver(unopened, { dataTransfer });
    fireEvent.drop(unopened, { dataTransfer });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // 上の否定が「ドロップ経路そのものが死んでいる」ことで通っていないかを
  // 確かめる負のコントロール。K だけの列へのドロップは通る。
  it('accepts a card dragged onto a column holding just its king', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 9' })).toBeInTheDocument());
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getByRole('button', { name: '♠ 9' }), { dataTransfer });
    const bareKing = screen.getByTestId('sl-bare-king-2');
    fireEvent.dragOver(bareKing, { dataTransfer });
    fireEvent.drop(bareKing, { dataTransfer });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 2 }),
    );
  });

  // **土台の K はドラッグでも動かない。**列に 1 枚しかない K を剥がすと、
  // 勘定 (K 8 枚 + 組札 88 枚) が崩れる。
  //
  // 置き先も「K だけの列」にしてあるのが要点。普通の列へ落とすと**置き先側の
  // 判定**で止まるので、移動元の判定を消しても落ちない ── それでは何も測って
  // いない。K だけの列を 2 つ用意して、移動元の判定だけを裸にする。
  it('refuses to drag the base king off its column', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([
        [card('SPADE', 13), card('SPADE', 9)],
        [card('HEART', 13), card('CLOVER', 1)],
        [card('DIAMOND', 13)],
        [card('CLOVER', 13)],
      ]),
      openPiles: 4,
    });
    renderWithProviders(<SalicLawPage />);
    const source = await screen.findByTestId('sl-bare-king-2');
    const target = screen.getByTestId('sl-bare-king-3');
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(source, { dataTransfer });
    fireEvent.dragOver(target, { dataTransfer });
    fireEvent.drop(target, { dataTransfer });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('sends a pile top to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    const ace = await screen.findByRole('button', { name: '♣ A' });
    fireEvent.click(ace);
    await waitFor(() => expect(ace).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空の組札1/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 1 }, { zone: 'foundation', col: 1 }),
    );
  });

  it('disables the stock once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /めくり直しはありません/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
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

  // **88 枚が満点。**K 8 枚は土台に残り、Q 8 枚は場に出ない。104 で数えると
  // 100%% に到達できない進捗バーになる。
  it('counts the summary against 88 cards, not 104', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<SalicLawPage />);
    const summary = await screen.findByTestId('sl-gameover-summary');
    expect(summary).toHaveTextContent('1/88');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('sl-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete until a foundation is open', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<SalicLawPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['foundation', { fromZone: 'tableau', fromIdx: 1, toZone: 'foundation', toIdx: 2 }, '組札2'],
    ['bare king', { fromZone: 'tableau', fromIdx: 0, toZone: 'tableau', toIdx: 5 }, '列5'],
    // 「配れ」は列を持たない。移動の体裁に落とすと「列-1」が出る。
    ['deal', { fromZone: 'stock', fromIdx: -1, toZone: 'stock', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
    // 負のコントロール: 存在しない列を指していないこと。
    expect(screen.queryByText(/列-1/)).not.toBeInTheDocument();
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByText('F0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('SalicLawPage keyboard shortcuts', () => {
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
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SalicLawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
