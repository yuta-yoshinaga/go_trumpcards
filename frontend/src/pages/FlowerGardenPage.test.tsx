import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flowerGardenApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
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
    await waitFor(() => expect(screen.getByRole('button', { name: '♦ 7' })).toBeInTheDocument());
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

  it('selecting a reserve card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FlowerGardenPage />);
    const reserveBtn = await screen.findByRole('button', { name: '♦ 7' });
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
});
