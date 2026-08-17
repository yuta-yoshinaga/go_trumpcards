import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flowerGardenApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FlowerGardenResponse, FlowerGardenTableauCard } from '../types/card';
import { FlowerGardenPage } from './FlowerGardenPage';

vi.mock('../api/gameApi', () => ({
  flowerGardenApi: { exec: vi.fn() },
  actionLogApi: { flowergarden: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(flowerGardenApi.exec);

function makeTableau(cols: FlowerGardenTableauCard[][]): FlowerGardenTableauCard[][] {
  const result: FlowerGardenTableauCard[][] = [];
  for (let i = 0; i < 6; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: FlowerGardenResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 13), faceUp: true },
      { card: card('HEART', 5), faceUp: true },
    ],
    [{ card: card('CLOVER', 6), faceUp: true }],
  ]),
  reserve: [card('DIAMOND', 7), ...Array.from({ length: 15 }, () => null)],
  foundation: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: FlowerGardenResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'flowergarden.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FlowerGardenResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'flowergarden.gameOver',
};

describe('FlowerGardenPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByText(/フラワーガーデン/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札 1枚/).length).toBe(4));
  });

  it('renders a reserve card', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /^♦ 7（リザーブ枠/ })).toBeInTheDocument());
  });

  it('labels all 16 bouquet slots with their 0-based index (matching hint text)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    // Slots are numbered #0..#15 to match formatHintZone's raw reserve col and
    // the CUI's [r0]..[r15], so hint text maps to a visible card.
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    for (let i = 0; i < 16; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('lays the 16-card bouquet out as a responsive grid (4 cols mobile, 8 cols sm+)', async () => {
    mockExec.mockResolvedValue(playingState);
    const { container } = renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    // The reserve slots live in a responsive grid so all 16 stay visible on mobile
    // instead of wrapping into a single cramped flex row (#3283).
    const grid = container.querySelector('[data-tutorial="fg-reserve"] .grid');
    expect(grid).not.toBeNull();
    expect(grid).toHaveClass('grid-cols-4', 'sm:grid-cols-8');
    // All 16 bouquet slots render inside the grid.
    for (let i = 0; i < 16; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('selecting a reserve card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    const reserveBtn = await screen.findByRole('button', { name: /^♦ 7（リザーブ枠/ });
    fireEvent.click(reserveBtn);
    await waitFor(() => expect(reserveBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
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

  it('shows the foundation progress summary on game over', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces on foundations → 4/52 (8%)
    renderWithProviders(<FlowerGardenPage />);
    const summary = await screen.findByTestId('fg-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('shows the hint button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: FlowerGardenResponse = {
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('names every bouquet reserve slot, filled or empty', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    // Filled slots carry the card plus its slot number; empty slots announce
    // the number alone, so no slot is a nameless blank to a screen reader.
    await waitFor(() => expect(screen.getAllByLabelText(/リザーブ枠 \d+/).length).toBeGreaterThan(0));
    expect(screen.getAllByLabelText(/空のリザーブ枠 \d+/).length).toBeGreaterThan(0);
  });

  // #5599: 「スートを問わない」という他のソリティアと違う規則が、初回だけ出る
  // チュートリアルにしか書かれていなかった。読み飛ばした後に思い出す手掛かりが
  // 盤面に無いので、**チュートリアルの状態に関係なく**出る注記を置く。
  it('states the suit-agnostic packing rule next to the tableau', async () => {
    localStorage.setItem('tutorial-flowergarden-done', 'true');
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const note = screen.getByTestId('fg-rules-note');
    expect(note).toHaveTextContent('スートは無視');
    expect(note).toHaveTextContent('1つ下のランク');
    // 空のベッドの扱いも書く ── 規則の半分だけ出すと、空きに置けないと誤解する。
    expect(note).toHaveTextContent('任意のカード');
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('FlowerGardenPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
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
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FlowerGardenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
