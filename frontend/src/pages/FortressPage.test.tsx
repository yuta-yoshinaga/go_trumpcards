import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortressApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FortressResponse, FortressTableauCard } from '../types/card';
import { FortressPage } from './FortressPage';

vi.mock('../api/gameApi', () => ({
  fortressApi: { exec: vi.fn() },
  actionLogApi: { fortress: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(fortressApi.exec);

function makeTableau(cols: FortressTableauCard[][]): FortressTableauCard[][] {
  const result: FortressTableauCard[][] = [];
  for (let i = 0; i < 8; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: FortressResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 13), faceUp: true },
      { card: card('SPADE', 5), faceUp: true },
    ],
    [{ card: card('HEART', 6), faceUp: true }],
    [],
    [],
    [],
    [],
    [],
    [],
  ]),
  foundation: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: FortressResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'fortress.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FortressResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'fortress.gameOver',
};

describe('FortressPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByText(/フォートレス/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札 1枚/).length).toBe(4));
  });

  it('labels all eight tableau columns with their 0-based index (matching hint text)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    // Columns are numbered #0..#7 to match formatHintZone's raw fromCol/toCol.
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('gives each empty tableau column a distinct column-numbered aria-label', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    // Columns 3 and 8 (1-based) are empty and each reads distinctly, unlike the
    // previous shared "empty" text.
    await waitFor(() => expect(screen.getByRole('button', { name: '空のタブロー列 3' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '空のタブロー列 8' })).toBeInTheDocument();
    // The two filled columns (1, 2) are not rendered as empty-column buttons.
    expect(screen.queryByRole('button', { name: '空のタブロー列 1' })).not.toBeInTheDocument();
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
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

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('shows the foundation progress summary on game over', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces on foundations → 4/52 (8%)
    renderWithProviders(<FortressPage />);
    const summary = await screen.findByTestId('fortress-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('fortress-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables the auto-complete button and shows a reason when no foundation has progressed', async () => {
    mockExec.mockResolvedValue(playingState); // foundations hold only aces
    renderWithProviders(<FortressPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn.className).not.toContain('animate-pulse');
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses the auto-complete button once a foundation builds past its ace', async () => {
    const readyState: FortressResponse = {
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<FortressPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
    expect(btn.className).toContain('ring-ds-success');
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: FortressResponse = {
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  // 合法な移動先のリング表示 (#4799)。「選ぶまで光らない」側も踏まないと、
  // 常時全部を光らせる実装でも通ってしまう。
  describe('legal target highlighting', () => {
    /** リングが付いた列の見出し (`#0` など)。ファンデーションは見出しを持たない。 */
    const markedColumns = () =>
      [...document.querySelectorAll('[data-legal-target="true"]')]
        .map((el) => el.querySelector('[aria-hidden="true"]')?.textContent ?? '')
        .filter((label) => label.startsWith('#'));

    it('marks nothing until a card is selected', async () => {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<FortressPage />);
      await screen.findByRole('button', { name: '♠ 5' });
      expect(document.querySelectorAll('[data-legal-target="true"]')).toHaveLength(0);
    });

    // ♠5 は ♥6 の上 (スート不問) と、空き列すべてに置ける。自分の列 #0 は
    // 相手が ♠5 自身なので置けない。
    it('marks the ranks-down column and every empty column, but not the source column', async () => {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<FortressPage />);
      fireEvent.click(await screen.findByRole('button', { name: '♠ 5' }));
      await waitFor(() => expect(markedColumns()).toContain('#1'));
      expect(markedColumns().sort()).toEqual(['#1', '#2', '#3', '#4', '#5', '#6', '#7']);
    });

    // **空の組札の唯一の受け手は A。**そこが光らないと、A にとって置き先が
    // 無いように見える (#5958)。リングは組札を包む要素側にあるので closest で辿る。
    it('marks the empty foundations when an ace is selected', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        tableau: makeTableau([[{ card: card('SPADE', 1), faceUp: true }]]),
        foundation: [[], [], [], []],
      });
      renderWithProviders(<FortressPage />);
      fireEvent.click(await screen.findByRole('button', { name: '♠ A' }));

      await waitFor(() =>
        expect(
          screen.getByRole('button', { name: '空の組札 (♠)' }).closest('[data-legal-target="true"]'),
        ).not.toBeNull(),
      );
      // 組札 4 つ + 空き列 7 つ。組札側が 0 なら 7 で止まる。
      expect(document.querySelectorAll('[data-legal-target="true"]')).toHaveLength(11);
    });

    // ファンデーションは A の上に同スートの 2 だけ。♠5 では光らない。
    it('marks a foundation only for the card that continues it', async () => {
      mockExec.mockResolvedValue(playingState);
      const { unmount } = renderWithProviders(<FortressPage />);
      fireEvent.click(await screen.findByRole('button', { name: '♠ 5' }));
      await waitFor(() => expect(markedColumns()).toContain('#1'));
      expect(document.querySelectorAll('[data-legal-target="true"]')).toHaveLength(7);
      unmount();

      mockExec.mockResolvedValue({
        ...playingState,
        tableau: makeTableau([[{ card: card('SPADE', 2), faceUp: true }]]),
      });
      renderWithProviders(<FortressPage />);
      fireEvent.click(await screen.findByRole('button', { name: '♠ 2' }));
      await waitFor(() => expect(document.querySelectorAll('[data-legal-target="true"]').length).toBeGreaterThan(7));
    });
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('FortressPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
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
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

// 選ぶ前に行き先が見える (#4454)。
describe('FortressPage destination preview', () => {
  const render = async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    return screen.findByRole('button', { name: '♠ 5' });
  };
  const targets = () => document.querySelectorAll('[data-legal-target="true"]');
  const previews = () => document.querySelectorAll('[data-preview-target="true"]');

  it('marks the destinations while a card is hovered', async () => {
    const spadeFive = await render();
    expect(targets()).toHaveLength(0);

    fireEvent.mouseEnter(spadeFive);
    await waitFor(() => expect(targets().length).toBeGreaterThan(0));
    // 選択後と同じ集合が、弱いリングで出る。
    expect(previews().length).toBe(targets().length);
    expect(targets()[0]?.className).toContain('ring-ds-success/70');

    fireEvent.mouseLeave(spadeFive);
    await waitFor(() => expect(targets()).toHaveLength(0));
  });

  it('marks the destinations on focus', async () => {
    const spadeFive = await render();
    fireEvent.focus(spadeFive);
    await waitFor(() => expect(previews().length).toBeGreaterThan(0));
    fireEvent.blur(spadeFive);
    await waitFor(() => expect(targets()).toHaveLength(0));
  });

  // #5596: ヒント表示は無言で書き換わっていた。**空のまま先にマウントしてある**
  // 領域の中身が変わることが読み上げの条件なので、hint がある間だけ現れる内側の
  // div ではなく、常設のラッパーがライブ領域でなければならない。
  it('keeps the hint live region mounted before there is any hint to announce', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByText(/フォートレス/)).toBeInTheDocument());

    const region = screen.getByTestId('fortress-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('');
  });

  it('announces the recommended move through that same region', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortressPage />);
    await waitFor(() => expect(screen.getByText(/フォートレス/)).toBeInTheDocument());

    const region = screen.getByTestId('fortress-hint-live');
    expect(region).toHaveTextContent('');

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 1, cardIndex: 0, toZone: 'foundation', toCol: 2 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    // **同じ要素**の中身が変わる (別の要素が現れるのではない) ことが読み上げの条件。
    await waitFor(() => expect(region).toHaveTextContent(/→/));
    expect(region.textContent).toBe('ヒントがあります: タブロー列1 → 組札');
  });

  // hover と選択で同じ集合を指す ── プレビューが嘘をつかないことの検証。
  it('previews exactly the set the selection then commits to', async () => {
    const spadeFive = await render();
    fireEvent.mouseEnter(spadeFive);
    await waitFor(() => expect(targets().length).toBeGreaterThan(0));
    const hovered = targets().length;

    fireEvent.click(spadeFive);
    await waitFor(() => expect(previews()).toHaveLength(0));
    expect(targets().length).toBe(hovered);
    expect(targets()[0]?.className).not.toContain('ring-ds-success/70');
  });
});
