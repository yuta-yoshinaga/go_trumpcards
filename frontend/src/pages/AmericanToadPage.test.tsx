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

// #5559: 8列 + 8基礎札 + リザーブ + 捨て札と候補が多いのに、どこに置けるかは
// クリックしてサーバーのエラーを見るまで分からなかった。
describe('AmericanToadPage destination highlight', () => {
  const targets = () => document.querySelectorAll('[data-legal-target]');

  // **空の組札の唯一の受け手は基準ランクの札。**そこが光らないと、置き先が
  // 無いように見える (#5958)。リングは組札を包む要素側にあるので closest で辿る。
  it('rings the empty foundations of the matching suit for a base-rank card', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      tableau: makeTableau([[{ card: card('SPADE', 5), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    fireEvent.click(await screen.findByRole('button', { name: /♠ 5/ }));

    await waitFor(() =>
      expect(
        screen.getAllByRole('button', { name: /^空の組札\d+ \(♠\)$/ })[0]?.closest('[data-legal-target="true"]'),
      ).not.toBeNull(),
    );
    // ♠ の組札は 2 つ (2 デッキ)。他スートの 6 つは光らない。
    expect(targets()).toHaveLength(2);
  });

  it('rings the legal column once a card is selected', async () => {
    // ♠8 を選ぶと ♠9 の上には置けない (降順なので ♠7 が要る) が、
    // 別の列の ♣4 の上でもない。合法な列だけが光ること。
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      tableau: makeTableau([
        [{ card: card('SPADE', 8), faceUp: true }],
        [{ card: card('SPADE', 9), faceUp: true }],
        [{ card: card('CLOVER', 4), faceUp: true }],
      ]),
    });
    renderWithProviders(<AmericanToadPage />);
    const source = await screen.findByRole('button', { name: /♠ 8/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));

    // ♠9 の列 (index 1) **だけ**がリング。♣4 は不一致、空列はタブロー由来の札を
    // 受け取らない (下の回帰テスト)。
    expect(targets()).toHaveLength(1);
    expect(document.querySelector('[data-preview-target]')).toBeNull();
  });

  // **判定を二重に持たない。**hover のプレビューも同じ計算を通る。
  it('previews the same targets on hover, marked as a preview', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }], [{ card: card('SPADE', 9), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const source = await screen.findByRole('button', { name: /♠ 8/ });
    fireEvent.mouseEnter(source);
    await waitFor(() => expect(document.querySelector('[data-preview-target]')).not.toBeNull());
  });

  // 山札由来の札 (捨て札) を選んでも同じ判定を通る。ページの previewedCard は
  // ゾーンごとに分岐しているので、tableau 以外も一度通しておく。
  it('highlights targets for the waste card too', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      waste: [card('SPADE', 7)],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const waste = await screen.findByRole('button', { name: /♠ 7/ });
    fireEvent.click(waste);
    await waitFor(() => expect(document.querySelectorAll('[data-legal-target]').length).toBeGreaterThan(0));
  });

  // リザーブの札を選んだときも同じ。
  it('highlights targets for the reserve card too', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [card('SPADE', 7)],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const reserve = await screen.findByRole('button', { name: /リザーブ 残り/ });
    fireEvent.click(reserve);
    await waitFor(() => expect(document.querySelectorAll('[data-legal-target]').length).toBeGreaterThan(0));
  });

  // **リザーブが残っている間は空列を光らせない。**自動補充の対象で、手では置けない。
  it('does not offer an empty column while the reserve holds cards', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [card('CLOVER', 3)],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const source = await screen.findByRole('button', { name: /♠ 8/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));
    // 空列は 7 つあるが、どれも光らない。♠8 の下に置ける札も無い。
    expect(targets()).toHaveLength(0);
  });

  // **リザーブが尽きても、空列はタブロー由来の札には開かない。**
  // `MoveTableauToTableau` が拒む (#4417) ので、光らせると押して弾かれる —
  // このPRが消そうとしているループそのものが残る。
  it('does not offer an empty column to a tableau card once the reserve is gone', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const source = await screen.findByRole('button', { name: /♠ 8/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));
    expect(targets()).toHaveLength(0);
  });

  // 負のコントロール: 同じ空列が、捨て札からなら開く。上のテストが
  // 「空列を常に光らせない」だけの実装でも通ってしまわないこと。
  it('does offer those empty columns to the waste card', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      reserve: [],
      waste: [card('DIAMOND', 9)],
      tableau: makeTableau([[{ card: card('SPADE', 8), faceUp: true }]]),
    });
    renderWithProviders(<AmericanToadPage />);
    const waste = await screen.findByRole('button', { name: /♦ 9/ });
    fireEvent.click(waste);
    await waitFor(() => expect(targets().length).toBeGreaterThan(0));
  });
});
