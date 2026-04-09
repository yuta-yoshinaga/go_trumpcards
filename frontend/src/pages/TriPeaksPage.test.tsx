import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, tripeaksApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TriPeaksCard, TriPeaksResponse } from '../types/card';
import { TriPeaksPage } from './TriPeaksPage';

vi.mock('../api/gameApi', () => ({
  tripeaksApi: { exec: vi.fn() },
  actionLogApi: { tripeaks: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(tripeaksApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTriPeaksCard(c: Card | null, removed: boolean, exposed: boolean): TriPeaksCard {
  return { card: c, removed, exposed };
}

/** Build a minimal TriPeaks layout for testing (4 rows × 10 cols). */
function makeTestLayout(): TriPeaksCard[][] {
  const layout: TriPeaksCard[][] = [];
  for (let r = 0; r < 4; r++) {
    const row: TriPeaksCard[] = [];
    for (let c = 0; c < 10; c++) {
      row.push(makeTriPeaksCard(null, true, false));
    }
    layout.push(row);
  }
  layout[3][0] = makeTriPeaksCard(card('SPADE', 5), false, true);
  layout[3][1] = makeTriPeaksCard(card('HEART', 6), false, true);
  layout[3][2] = makeTriPeaksCard(card('CLOVER', 7), false, true);
  layout[0][0] = makeTriPeaksCard(card('DIAMOND', 10), false, false);
  layout[0][3] = makeTriPeaksCard(card('SPADE', 11), false, false);
  layout[0][6] = makeTriPeaksCard(card('HEART', 12), false, false);
  return layout;
}

const playingState: TriPeaksResponse = {
  layout: makeTestLayout(),
  stockCount: 20,
  waste: [card('CLOVER', 4)],
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

const gameClearState: TriPeaksResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'tripeaks.gameClear',
  messageParams: { moveCount: '28' },
};

const gameOverState: TriPeaksResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'tripeaks.gameOver',
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('TriPeaksPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TriPeaksPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders stock count', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
    expect(screen.getByText(/\(20\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 3/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [] });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(drawButtons[drawButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking giveup button dispatches giveup', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking undo button dispatches undo', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('renders game clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
  });

  it('renders game over state', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
  });

  it('hides action buttons when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  it('disables undo button when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('renders correctly on mobile viewport (isMobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 393 });
    try {
      renderWithProviders(<TriPeaksPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
      // 10-column tableau should render with effectiveCardWidth derived from viewport
      const tableauRows = document.querySelectorAll('[data-tutorial="tp-peaks"] > div');
      expect(tableauRows.length).toBe(4);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders correctly on desktop viewport (non-mobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      renderWithProviders(<TriPeaksPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3, canUndo: true });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('3');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 4, canUndo: true });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });
});
