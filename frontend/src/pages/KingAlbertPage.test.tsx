import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kingAlbertApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, KingAlbertResponse, KingAlbertTableauCard } from '../types/card';
import { KingAlbertPage } from './KingAlbertPage';

vi.mock('../api/gameApi', () => ({
  kingAlbertApi: { exec: vi.fn() },
  actionLogApi: { kingalbert: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(kingAlbertApi.exec);

function makeTableau(cols: KingAlbertTableauCard[][]): KingAlbertTableauCard[][] {
  const result: KingAlbertTableauCard[][] = [];
  for (let i = 0; i < 9; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: KingAlbertResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 13), faceUp: true },
      { card: card('HEART', 5), faceUp: true },
    ],
    [{ card: card('CLOVER', 6), faceUp: true }],
  ]),
  reserve: [card('DIAMOND', 7), null, null, null, null, null, null],
  foundation: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: KingAlbertResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'kingalbert.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: KingAlbertResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'kingalbert.gameOver',
};

describe('KingAlbertPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByText(/キング・アルバート/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札 1枚/).length).toBe(4));
  });

  it('renders a reserve card', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♦ 7' })).toBeInTheDocument());
  });

  it('gives each empty reserve slot a role=img with a numbered aria-label', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    // Reserve slots 2..7 (1-based) are empty and each is now an announced,
    // numbered slot instead of an anonymous div.
    await waitFor(() => expect(screen.getByRole('img', { name: '空のリザーブ枠 2' })).toBeInTheDocument());
    expect(screen.getByRole('img', { name: '空のリザーブ枠 7' })).toBeInTheDocument();
    // Slot 1 holds ♦7 (a button), so it is not an empty-slot image.
    expect(screen.queryByRole('img', { name: '空のリザーブ枠 1' })).not.toBeInTheDocument();
  });

  it('labels all 7 reserve slots with 0-based numbers matching the hint text', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByTestId('ka-reserve-label-0')).toBeInTheDocument());
    for (let i = 0; i < 7; i++) {
      const label = screen.getByTestId(`ka-reserve-label-${i.toString()}`);
      expect(label).toHaveTextContent(`r${i.toString()}`);
    }
  });

  it('selecting a reserve card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    const reserveBtn = await screen.findByRole('button', { name: '♦ 7' });
    fireEvent.click(reserveBtn);
    await waitFor(() => expect(reserveBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows the foundation progress summary on game over', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces on foundations → 4/52 (8%)
    renderWithProviders(<KingAlbertPage />);
    const summary = await screen.findByTestId('ka-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('shows the hint button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: KingAlbertResponse = {
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('KingAlbertPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KingAlbertPage />);
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
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<KingAlbertPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    expect(mockExec).not.toHaveBeenCalled();
  });
});
