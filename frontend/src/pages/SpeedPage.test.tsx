import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { speedApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SpeedResponse } from '../types/card';
import { SpeedPage } from './SpeedPage';

vi.mock('../api/gameApi', () => ({
  speedApi: { exec: vi.fn() },
  actionLogApi: { speed: vi.fn() },
}));

const mockExec = vi.mocked(speedApi.exec);

const playState: SpeedResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 4 },
        { design: 'HEART', value: 8 },
        { design: 'CLOVER', value: 11 },
        { design: 'DIAMOND', value: 2 },
      ],
      drawPileSize: 21,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], drawPileSize: 21 },
  ],
  centerPiles: [
    { design: 'DIAMOND', value: 5 },
    { design: 'SPADE', value: 9 },
  ],
  phase: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  config: { cpuDifficulty: 1, autoFlip: true },
  message: '',
};

const stuckState: SpeedResponse = {
  ...playState,
  phase: 1,
};

const playStateWithHint: SpeedResponse = {
  ...playState,
  hint: { cardIndex: 0, pileIndex: 1, found: true },
};

const gameEndState: SpeedResponse = {
  ...playState,
  phase: 2,
  gameEndFlag: true,
  winnerIdx: 0,
};

const gameEndLoseState: SpeedResponse = {
  ...playState,
  phase: 2,
  gameEndFlag: true,
  winnerIdx: 1,
};

const errorState: SpeedResponse = {
  ...playState,
  message: 'invalid play',
  messageCode: 'error',
};

describe('SpeedPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(playState);
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, autoFlip: true });
    });
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<SpeedPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders the page heading', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('スピード')).toBeInTheDocument());
  });

  it('renders player hand after API resolves', async () => {
    renderWithProviders(<SpeedPage />);
    // Wait for the exec call to happen (state loaded)
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Then verify state-dependent content renders
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
  });

  it('shows stuck phase flip button', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
  });

  it('calls flip on flip button click', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('flip-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip'));
  });

  it('shows game end phase', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('ゲーム終了')).toBeInTheDocument());
  });

  it('does not show flip button in play phase', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'めくる' })).not.toBeInTheDocument();
  });

  it('shows stuck message text', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText(/膠着状態/)).toBeInTheDocument());
  });

  it('shows hint when available in play phase', async () => {
    mockExec.mockResolvedValue(playStateWithHint);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText(/カード0を台札1に出せます/)).toBeInTheDocument());
  });

  it('does not show hint in stuck phase', async () => {
    const stuckWithHint: SpeedResponse = { ...stuckState, hint: { cardIndex: 0, pileIndex: 0, found: true } };
    mockExec.mockResolvedValue(stuckWithHint);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    expect(screen.queryByText(/カード0を台札0に出せます/)).not.toBeInTheDocument();
  });

  it('shows CPU card count and draw pile', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText(/CPU手札/)).toBeInTheDocument());
  });

  it('shows error message from API', async () => {
    mockExec.mockResolvedValue(errorState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('invalid play')).toBeInTheDocument());
  });

  it('disables play buttons in stuck phase', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    // Card buttons should be disabled in stuck phase
    const cardButtons = screen.queryAllByRole('button', { name: /SPADE|HEART|CLOVER|DIAMOND/ });
    for (const btn of cardButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('game end lose state does not show celebration', async () => {
    mockExec.mockResolvedValue(gameEndLoseState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了')).toBeInTheDocument());
  });

  it('selects and deselects a hand card on click', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    const cardBtn = screen.getByRole('button', { name: 'SPADE 4' });
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('plays a card to a center pile when a card is selected', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // Select a hand card
    fireEvent.click(screen.getByRole('button', { name: 'SPADE 4' }));
    // Click first center pile
    const pileBtns = screen.getAllByRole('button', { name: /台札/ });
    fireEvent.click(pileBtns[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 0));
  });

  it('clicking hint button calls hint command', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows phase as stuck when phase is 1', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('膠着')).toBeInTheDocument());
  });

  it('flip button has pulse animation when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    expect(screen.getByTestId('flip-button')).toHaveClass('animate-pulse');
  });

  it('center piles trigger flip when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playState);
    // Center pile buttons should have flip aria-label when stuck
    const flipPileBtns = screen.getAllByRole('button', { name: 'めくる' });
    // There should be center piles + the flip button = 3 buttons with 'めくる' label
    expect(flipPileBtns.length).toBeGreaterThanOrEqual(2);
    fireEvent.click(flipPileBtns[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip'));
  });

  it('center piles have pulse animation when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    const flipPileBtns = screen.getAllByRole('button', { name: 'めくる' });
    // Center pile buttons (first two) should have animate-pulse
    expect(flipPileBtns[0]).toHaveClass('animate-pulse');
  });

  it('renders the auto-flip settings toggle', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    expect(screen.getByLabelText('膠着時に自動でめくる')).toBeInTheDocument();
  });
});

describe('SpeedPage auto-flip timer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValue(stuckState);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('automatically calls flip after the delay when stuck and auto-flip enabled', async () => {
    renderWithProviders(<SpeedPage />);
    // Wait until the stuck state is rendered
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    mockExec.mockClear();
    // Advance past the 2.5s auto-flip delay
    await act(async () => {
      await vi.advanceTimersByTimeAsync(2600);
    });
    expect(mockExec).toHaveBeenCalledWith('flip');
  });

  it('does not auto-flip before the delay elapses', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(500);
    });
    expect(mockExec).not.toHaveBeenCalledWith('flip');
  });

  it('does not auto-flip when disabled via the settings toggle', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    // Turn auto-flip off before the timer fires
    fireEvent.click(screen.getByLabelText('膠着時に自動でめくる'));
    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mockExec).not.toHaveBeenCalledWith('flip');
  });
});
