import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, stHelenaApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, StHelenaResponse, StHelenaTableauCard } from '../types/card';
import { StHelenaPage } from './StHelenaPage';

vi.mock('../api/gameApi', () => ({
  stHelenaApi: { exec: vi.fn() },
  actionLogApi: { sthelena: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(stHelenaApi.exec);

// **12 列。**クローン元のクレセントは 16 列なので、16 で埋めると存在しない列を
// 描いた盤でテストすることになる。
function makeTableau(cols: StHelenaTableauCard[][] | Record<number, StHelenaTableauCard[]>): StHelenaTableauCard[][] {
  const result: StHelenaTableauCard[][] = [];
  for (let i = 0; i < 12; i++) {
    result.push((cols as Record<number, StHelenaTableauCard[]>)[i] ?? []);
  }
  return result;
}

// **横の列に置く。**初回の配りでは上 4 列は K 段、下 4 列は A 段にしか送れない。
// クレセントにこの制限は無いので、クローンした盤は列 0 から A 段へ送っていた。
// 横の列 (4, 5, 10, 11) はどちらへも送れるので、制限そのものを測る専用テスト
// 以外はここに置く。
const SIDE_A = 4;
const SIDE_B = 5;
const SIDE_C = 10;

function sideTableau(...cols: StHelenaTableauCard[][]): StHelenaTableauCard[][] {
  const at: StHelenaTableauCard[][] = [];
  for (const [i, col] of [SIDE_A, SIDE_B, SIDE_C].entries()) {
    if (cols[i]) at[col] = cols[i];
  }
  return makeTableau(at);
}

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

const playingState: StHelenaResponse = {
  tableau: sideTableau(
    [
      { card: card('SPADE', 5), faceUp: true },
      { card: card('SPADE', 4), faceUp: true },
    ],
    [{ card: card('HEART', 6), faceUp: true }],
  ),
  foundation: [
    [card('SPADE', 1)],
    [card('CLOVER', 1)],
    [card('HEART', 1)],
    [card('DIAMOND', 1)],
    [card('SPADE', 13)],
    [card('CLOVER', 13)],
    [card('HEART', 13)],
    [card('DIAMOND', 13)],
  ],
  redealsRemaining: 3,
  restrictionsActive: true,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: StHelenaResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'sthelena.gameClear',
};

const gameOverState: StHelenaResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'sthelena.gameOver',
};

const noRedealsState: StHelenaResponse = {
  ...playingState,
  redealsRemaining: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('StHelenaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<StHelenaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders redeals remaining in header', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByText(/残り再配り回数: 3/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders ascending and descending foundation suit headers', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getAllByText(/♠ ↑/).length).toBeGreaterThanOrEqual(1));
    expect(screen.getAllByText(/♣ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♥ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♦ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♠ ↓/).length).toBeGreaterThanOrEqual(1);
  });

  it('color-codes the foundation direction badges and names the top card', async () => {
    renderWithProviders(<StHelenaPage />);
    const asc = await screen.findByTestId('foundation-dir-0'); // row 0 = ascending
    expect(asc.className).toContain('text-ds-success');
    const desc = screen.getByTestId('foundation-dir-4'); // row 1 = descending
    expect(desc.className).toContain('text-ds-warning');
    // Ascending ♠ pile tops out at A → aria-label is localized and names the top card.
    expect(screen.getByLabelText(/昇順ファンデーション ♠ 残り1枚 トップ ♠ A/)).toBeInTheDocument();
  });

  it('redeal button shows remaining count', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /再配り \(3\)/ })).toBeInTheDocument();
  });

  it('exposes the redeal count via an aria-live status region', async () => {
    renderWithProviders(<StHelenaPage />);
    const status = await screen.findByText(/残り再配り回数/);
    expect(status).toHaveAttribute('aria-live', 'polite');
  });

  it('announces a stalemate via a role=alert region', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2 });
    renderWithProviders(<StHelenaPage />);
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toMatch(/手詰まり/);
    expect(alert.textContent).toMatch(/2/);
  });

  it('defaults the escape count to 0 when undoToEscape is absent', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: undefined });
    renderWithProviders(<StHelenaPage />);
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toMatch(/手詰まり/);
    expect(alert.textContent).toMatch(/0/);
  });

  it('clicking redeal dispatches redeal', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り \(3\)/ })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: /再配り \(3\)/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('redeal button disabled when no redeals remain', async () => {
    mockExec.mockResolvedValue(noRedealsState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り \(0\)/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /再配り \(0\)/ })).toBeDisabled();
  });

  it('clicking hint dispatches hint', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking tableau top card selects it as source', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const cardImg = screen.getByAltText('♠ 4');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  // **初回の配りの制限をページが守ること。**サーバは拒むが、押せてしまうと
  // 「なぜ動かないのか」が分からないまま手が止まる。
  it('refuses a foundation the selected column cannot reach on the first deal', async () => {
    // 上の列 (0) に ♠4 を置く。A 段 (組札0) には届かない。
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau({ 0: [{ card: card('SPADE', 2), faceUp: true }] }),
      restrictionsActive: true,
    });
    renderWithProviders(<StHelenaPage />);
    const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
    mockExec.mockClear();

    const aceRow = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(aceRow).toBeDisabled();
    // 光らせてもいけない。無効化だけだと「押せないが緑」の矛盾した見た目になる。
    // (無効化は到達判定だけ、リングは到達 AND ランク。別の述語なので両方見る。)
    expect(aceRow.closest('[class*="ring-ds-success"]')).toBeNull();
    fireEvent.click(aceRow);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // **ドラッグ経路も同じ規則を守る。**クリックはボタンの無効化で防いでいるが、
  // ドラッグは dispatchMove を直接通る。
  it('ignores a card dragged onto a foundation its column cannot reach', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau({ 0: [{ card: card('SPADE', 2), faceUp: true }] }),
      restrictionsActive: true,
    });
    renderWithProviders(<StHelenaPage />);
    const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(cardButton, { dataTransfer });
    const aceRow = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.dragOver(aceRow, { dataTransfer });
    fireEvent.drop(aceRow, { dataTransfer });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // 負のコントロール: ドロップ経路そのものが死んでいるのではないこと。
  it('accepts the same drag once the restriction is lifted', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau({ 0: [{ card: card('SPADE', 2), faceUp: true }] }),
      restrictionsActive: false,
    });
    renderWithProviders(<StHelenaPage />);
    const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;
    mockExec.mockClear();

    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(cardButton, { dataTransfer });
    const aceRow = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.dragOver(aceRow, { dataTransfer });
    fireEvent.drop(aceRow, { dataTransfer });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: 0 }),
        expect.objectContaining({ zone: 'foundation', col: 0 }),
      ),
    );
  });

  // 負のコントロール: 制限が解ければ同じ手が通る。解けても false のままだと、
  // 上のテストは通るのに後半どこにも送れなくなる。
  it('accepts the same move once the restriction is lifted', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau({ 0: [{ card: card('SPADE', 2), faceUp: true }] }),
      restrictionsActive: false,
    });
    renderWithProviders(<StHelenaPage />);
    const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
    mockExec.mockClear();

    // 負のコントロール: 制限が解ければ同じ組札が光る。リングの述語から到達
    // 判定を落とすと、上の「光らない」だけが残って何も測らなくなる。
    const aceRow = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(aceRow.closest('[class*="ring-ds-success"]')).not.toBeNull();

    fireEvent.click(aceRow);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: 0 }),
        expect.objectContaining({ zone: 'foundation', col: 0 }),
      ),
    );
  });

  it('selecting source then foundation dispatches move', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // 横の列 (SIDE_A) の上札 ♠4 を選ぶ。上の列だと初回の配りでは A 段に
    // 届かないので、送れないのが正しい。
    const cardImg = screen.getByAltText('♠ 4');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click a foundation (♠ A ascending)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const foundationImg = screen.getByAltText('♠ A');
    const foundationButton = foundationImg.closest('button') as HTMLButtonElement;
    fireEvent.click(foundationButton);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: SIDE_A }),
        expect.objectContaining({ zone: 'foundation', col: 0 }),
      ),
    );
  });

  it('rings valid destinations when a source is selected (no hover needed)', async () => {
    // col0 top ♠4 can go onto foundation ♠3 (asc) and onto tableau ♠5; ♥6 is invalid.
    const highlightState: StHelenaResponse = {
      ...playingState,
      tableau: sideTableau(
        [{ card: card('SPADE', 4), faceUp: true }],
        [{ card: card('SPADE', 5), faceUp: true }],
        [{ card: card('HEART', 6), faceUp: true }],
      ),
      foundation: [
        [card('SPADE', 3)],
        [card('CLOVER', 1)],
        [card('HEART', 1)],
        [card('DIAMOND', 1)],
        [card('SPADE', 13)],
        [card('CLOVER', 13)],
        [card('HEART', 13)],
        [card('DIAMOND', 13)],
      ],
    };
    mockExec.mockResolvedValue(highlightState);

    const { container } = renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const foundations = container.querySelector('[data-tutorial="sthelena-foundations"]') as HTMLElement;
    const tableau = container.querySelector('[data-tutorial="sthelena-tableau"]') as HTMLElement;
    // No selection yet -> nothing highlighted.
    expect(foundations.querySelectorAll('.ring-ds-success')).toHaveLength(0);
    expect(tableau.querySelectorAll('.ring-ds-success')).toHaveLength(0);

    // Select ♠4 as the source.
    const cardButton = screen.getByAltText('♠ 4').closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);

    await waitFor(() => expect(foundations.querySelectorAll('.ring-ds-success').length).toBe(1));
    // Exactly the ♠5 column rings; the ♥6 column and the source do not.
    expect(tableau.querySelectorAll('.ring-ds-success')).toHaveLength(1);
    const invalidTop = screen.getByAltText('♥ 6').closest('button') as HTMLButtonElement;
    expect(invalidTop.className).toContain('opacity-40');
  });

  it('clicking reset dispatches reset (after confirm)', async () => {
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over hides play buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('renders empty foundation placeholder', async () => {
    const emptyFndState: StHelenaResponse = {
      ...playingState,
      foundation: [[], [], [], [], [], [], [], []],
    };
    mockExec.mockResolvedValue(emptyFndState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getAllByText(/♠ ↑/).length).toBeGreaterThanOrEqual(1));
    // Empty asc foundations show "A"; descending show "K".
    expect(screen.getAllByText('A').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('K').length).toBeGreaterThanOrEqual(1);
  });

  it('action log fetches and shows log entries', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.sthelena);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'redeal', detail: 'shuffle' }],
    });

    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  // **列の位置が規則そのもの。**初回の配りでは上 4 列は K 段、下 4 列は A 段に
  // しか送れないので、どの列がどの帯にいるかが読めないと打てない。クローン元の
  // クレセントは 16 列の三日月アーチを描くが、その形はクレセントのもの。
  it('groups the twelve columns into the top, side and bottom bands', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('sthelena-band-top')).toBeInTheDocument());

    for (const [band, cols] of [
      ['top', [0, 1, 2, 3]],
      ['side', [11, 4, 10, 5]],
      ['bottom', [6, 7, 8, 9]],
    ] as const) {
      const badges = screen
        .getByTestId(`sthelena-band-${band}`)
        .querySelectorAll('[data-testid^="sthelena-col-badge-"]');
      expect(Array.from(badges).map((b) => b.textContent)).toEqual(cols.map((c) => `[${c.toString()}]`));
    }
  });

  it('renders a [0]..[11] column-number badge above each tableau pile', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('sthelena-col-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('sthelena-col-badge-0')).toHaveTextContent('[0]');
    // 12 列。クローン元の 16 のままだと存在しない列を主張する。
    expect(screen.getByTestId('sthelena-col-badge-11')).toHaveTextContent('[11]');
    expect(screen.queryByTestId('sthelena-col-badge-12')).not.toBeInTheDocument();
    // The badge is decorative — it must not add to the SR card label noise.
    expect(screen.getByTestId('sthelena-col-badge-7')).toHaveAttribute('aria-hidden', 'true');
  });

  it('announces how many legal destinations the selected card has', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StHelenaPage />);
    const status = await screen.findByTestId('cr-selection-status');
    // Nothing selected yet, so the region stays empty rather than announcing noise.
    expect(status).toHaveTextContent('');
    expect(status).toHaveAttribute('aria-live', 'polite');

    // ♠4 needs ♠2 and ♠3 on the ascending ♠A first, so it has nowhere to go —
    // the dead end has to be announced too, not just left silent.
    fireEvent.click(screen.getByAltText('♠ 4').closest('button') as HTMLButtonElement);
    await waitFor(() => expect(status).toHaveTextContent('置ける場所はありません'));
  });

  it('counts the legal destinations of a playable card', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // ♠2 goes straight onto the ascending ♠A foundation.
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: sideTableau([{ card: card('SPADE', 2), faceUp: true }]),
    });
    renderWithProviders(<StHelenaPage />);
    const status = await screen.findByTestId('cr-selection-status');
    fireEvent.click(screen.getByAltText('♠ 2').closest('button') as HTMLButtonElement);
    await waitFor(() => expect(status).toHaveTextContent(/置ける場所が\d+箇所/));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('StHelenaPage keyboard shortcuts', () => {
  it.each([
    ['d', 'redeal'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StHelenaPage />);
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
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<StHelenaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // #5590: 8 組札 (昇順4 + 降順4) は Congress より複雑なのに、どこまで進んで
  // いたかを自分で数えるしかなかった。
  describe('game-over summary', () => {
    it('reports the foundation count and the percentage', async () => {
      mockExec.mockResolvedValue(gameOverState);
      renderWithProviders(<StHelenaPage />);
      const summary = await screen.findByTestId('cr-gameover-summary');
      // 種札 8 枚のみの盤面 → 8/104 = 8%。**部分一致で見ない** ── 分母の 104 に
      // "10" も "4" も含まれるので、数字だけを探すと何とでも一致してしまう。
      expect(summary.textContent).toBe('組札 8/104 枚（8%）まで到達');
    });

    // **数えているのは実際の枚数。**固定値を出す実装では通らない。
    it('counts what the foundations actually hold', async () => {
      mockExec.mockResolvedValue({
        ...gameOverState,
        foundation: [[card('SPADE', 1), card('SPADE', 2), card('SPADE', 3)], ...gameOverState.foundation.slice(1)],
      });
      renderWithProviders(<StHelenaPage />);
      const summary = await screen.findByTestId('cr-gameover-summary');
      expect(summary.textContent).toBe('組札 10/104 枚（10%）まで到達');
    });

    it('stays away on a clear and during play', async () => {
      mockExec.mockResolvedValue(gameClearState);
      const { unmount } = renderWithProviders(<StHelenaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('cr-gameover-summary')).not.toBeInTheDocument();
      unmount();

      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<StHelenaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('cr-gameover-summary')).not.toBeInTheDocument();
    });
  });
});
