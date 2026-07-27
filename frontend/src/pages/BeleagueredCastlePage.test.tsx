import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beleagueredCastleApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BeleagueredCastleResponse, BeleagueredCastleTableauCard, Card, CardDesign } from '../types/card';
import { BeleagueredCastlePage } from './BeleagueredCastlePage';

vi.mock('../api/gameApi', () => ({
  beleagueredCastleApi: { exec: vi.fn() },
  actionLogApi: { beleagueredcastle: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(beleagueredCastleApi.exec);

function makeTableau(cols: BeleagueredCastleTableauCard[][]): BeleagueredCastleTableauCard[][] {
  const result: BeleagueredCastleTableauCard[][] = [];
  for (let i = 0; i < 8; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: BeleagueredCastleResponse = {
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

const gameClearState: BeleagueredCastleResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'beleagueredcastle.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: BeleagueredCastleResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'beleagueredcastle.gameOver',
};

describe('BeleagueredCastlePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByText(/包囲された城/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札 1枚/).length).toBe(4));
  });

  it('labels all eight tableau columns with their 0-based index (matching hint text)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    // Columns are numbered #0..#7 to match formatHintZone's raw fromCol/toCol.
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('gives each empty tableau column a distinct column-numbered aria-label', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    // Columns 3 and 8 (1-based) are empty and each reads distinctly, unlike the
    // previous shared "empty" text.
    await waitFor(() => expect(screen.getByRole('button', { name: '空のタブロー列 3' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '空のタブロー列 8' })).toBeInTheDocument();
    // The two filled columns (1, 2) are not rendered as empty-column buttons.
    expect(screen.queryByRole('button', { name: '空のタブロー列 1' })).not.toBeInTheDocument();
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('shows the foundation progress summary on game over', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces on foundations → 4/52 (8%)
    renderWithProviders(<BeleagueredCastlePage />);
    const summary = await screen.findByTestId('bc-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('bc-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables the auto-complete button and shows a reason when no foundation has progressed', async () => {
    mockExec.mockResolvedValue(playingState); // foundations hold only aces
    renderWithProviders(<BeleagueredCastlePage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn.className).not.toContain('animate-pulse');
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses the auto-complete button once a foundation builds past its ace', async () => {
    const readyState: BeleagueredCastleResponse = {
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<BeleagueredCastlePage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
    expect(btn.className).toContain('ring-ds-success');
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: BeleagueredCastleResponse = {
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
    const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('BeleagueredCastlePage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BeleagueredCastlePage />);
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
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BeleagueredCastlePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    expect(mockExec).not.toHaveBeenCalled();
  });
});
