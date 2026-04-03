import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, golfApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, GolfCard, GolfResponse } from '../types/card';
import { GolfPage } from './GolfPage';

vi.mock('../api/gameApi', () => ({
  golfApi: { exec: vi.fn() },
  actionLogApi: { golf: vi.fn() },
}));

const mockExec = vi.mocked(golfApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeGolfCard(c: Card | null, removed: boolean, exposed: boolean): GolfCard {
  return { card: c, removed, exposed };
}

/** Build a minimal Golf layout for testing (7 cols × 5 rows). */
function makeTestLayout(): GolfCard[][] {
  const layout: GolfCard[][] = [];
  for (let col = 0; col < 7; col++) {
    const column: GolfCard[] = [];
    for (let row = 0; row < 5; row++) {
      if (row === 4) {
        column.push(makeGolfCard(card('SPADE', (col % 13) + 1), false, true));
      } else {
        column.push(makeGolfCard(card('HEART', (col * 5 + row) % 13 + 1), false, false));
      }
    }
    layout.push(column);
  }
  return layout;
}

const playingState: GolfResponse = {
  layout: makeTestLayout(),
  stockCount: 16,
  waste: [card('CLOVER', 4)],
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

const gameClearState: GolfResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'golf.gameClear',
  messageParams: { moveCount: '35' },
};

const gameOverState: GolfResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'golf.gameOver',
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
});

describe('GolfPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GolfPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders stock count', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
    expect(screen.getByText(/\(16\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 3/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [] });
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(drawButtons[drawButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking giveup button dispatches giveup', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking undo button dispatches undo', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('renders game clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
  });

  it('renders game over state', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
  });

  it('hides action buttons when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  it('disables undo button when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });
});
