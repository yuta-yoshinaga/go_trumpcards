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

vi.mock('../hooks/useChainCombo', () => ({
  useChainCombo: vi.fn().mockReturnValue(0),
}));

import { useChainCombo } from '../hooks/useChainCombo';

const mockExec = vi.mocked(golfApi.exec);
const mockCombo = vi.mocked(useChainCombo);

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
        column.push(makeGolfCard(card('HEART', ((col * 5 + row) % 13) + 1), false, false));
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
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
  mockCombo.mockReturnValue(0);
});

describe('GolfPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GolfPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('rings exposed cards that are playable (±1 of the waste top)', async () => {
    // Waste top 4; exposed bottom-row values are 1..7, so only 3 and 5 are playable.
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getAllByTestId('golf-playable')).toHaveLength(2);
  });

  it('offers the frontend hint toggle in the settings panel', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    const toggle = screen.getByLabelText('ヒント表示');
    expect(toggle).not.toBeChecked();
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('shows the frontend hint tooltip when the toggle is enabled', async () => {
    localStorage.setItem('hint_enabled_golf', 'true');
    renderWithProviders(<GolfPage />);
    // playingState's waste top (♣4) has adjacent exposed cards → the removable suggestion appears.
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('renders stock count', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
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

  it('clicking giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
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

  it('clicking exposed card dispatches remove', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Find an exposed card button (col 0 bottom card is ♠A)
    const cardButtons = screen.getAllByRole('button', { name: /♠/ });
    fireEvent.click(cardButtons[0]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', expect.any(Number)));
  });

  it('clicking hint button dispatches hint', async () => {
    const hintState: GolfResponse = {
      ...playingState,
      hint: { type: 'remove', col: 0 },
      messageCode: 'golf.hintAvailable',
    };
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockResolvedValue(hintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('renders stalemate message', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      messageCode: 'golf.stalemate',
      message: '手詰まりです。元に戻すかギブアップしてください。',
    });
    renderWithProviders(<GolfPage />);
    await waitFor(() =>
      expect(screen.getByText('手詰まりです。元に戻すかギブアップしてください。')).toBeInTheDocument(),
    );
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });

  it('renders correctly on mobile viewport (isMobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 393 });
    try {
      renderWithProviders(<GolfPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
      // 7 columns should be rendered with effectiveCardWidth derived from viewport
      const colDivs = document.querySelectorAll('[data-tutorial="golf-columns"] > div');
      expect(colDivs.length).toBe(7);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders correctly on desktop viewport (non-mobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      renderWithProviders(<GolfPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('does not render the combo badge when combo < 2', async () => {
    mockCombo.mockReturnValue(1);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('combo-badge')).not.toBeInTheDocument();
  });

  it('renders the combo badge with blue styling when combo is 2', async () => {
    mockCombo.mockReturnValue(2);
    renderWithProviders(<GolfPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-info');
  });

  it('renders the combo badge with warning styling when combo is between 3 and 4', async () => {
    mockCombo.mockReturnValue(3);
    renderWithProviders(<GolfPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-warning');
  });

  it('renders the combo badge with error styling when combo >= 5', async () => {
    mockCombo.mockReturnValue(5);
    renderWithProviders(<GolfPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-error');
  });

  it('does not show the 9-hole scorecard by default', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByTestId('golf-scorecard')).not.toBeInTheDocument();
  });

  it('shows the scorecard when 9-hole mode is enabled via the settings toggle', async () => {
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('golf-ninehole-toggle'));
    expect(screen.getByTestId('golf-scorecard')).toBeInTheDocument();
    // Fresh card: hole 1 pending, total 0.
    expect(screen.getByTestId('golf-scorecard-total')).toHaveTextContent('0');
  });

  it('records the finished deal once as the current hole (remaining cards = score)', async () => {
    // 9-hole mode on; the test layout has all 35 cards present → hole score 35.
    localStorage.setItem('trumpcards-golf-ninehole', JSON.stringify({ enabled: true, scores: [] }));
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByTestId('golf-hole-1')).toHaveTextContent('35'));
    // Only hole 1 is recorded — no double-count on re-render.
    expect(screen.getByTestId('golf-hole-2')).toHaveTextContent('-');
    expect(screen.getByTestId('golf-scorecard-total')).toHaveTextContent('35');
    expect(JSON.parse(localStorage.getItem('trumpcards-golf-ninehole') ?? '{}').scores).toEqual([35]);
  });

  it('shows the completion message and total after 9 holes', async () => {
    localStorage.setItem(
      'trumpcards-golf-ninehole',
      JSON.stringify({ enabled: true, scores: [1, 2, 3, 4, 5, 6, 7, 8, 9] }),
    );
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByTestId('golf-scorecard')).toBeInTheDocument());
    expect(screen.getByText(/合計スコア: 45/)).toBeInTheDocument();
    expect(screen.getByTestId('golf-scorecard-restart')).toBeInTheDocument();
  });

  it('restart button clears the completed scorecard', async () => {
    localStorage.setItem(
      'trumpcards-golf-ninehole',
      JSON.stringify({ enabled: true, scores: [1, 2, 3, 4, 5, 6, 7, 8, 9] }),
    );
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByTestId('golf-scorecard-restart')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('golf-scorecard-restart'));
    expect(screen.getByTestId('golf-scorecard-total')).toHaveTextContent('0');
    expect(screen.getByTestId('golf-hole-1')).toHaveTextContent('-');
  });

  it('pulses the stock pile when a draw hint is active', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GolfPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    // The hint is fetched via the hint command; a draw hint should pulse the stock.
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'draw', col: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => {
      const stock = screen.getByTestId('golf-stock');
      expect(stock.className).toContain('ring-ds-warning');
      expect(stock.className).toContain('animate-pulse');
    });
  });
});
