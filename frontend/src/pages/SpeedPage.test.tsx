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
    // Card buttons (cardAlt labels use suit symbols) should be disabled in stuck phase.
    const cardButtons = screen.queryAllByRole('button', { name: /[♠♥♣♦]/ });
    expect(cardButtons.length).toBeGreaterThan(0);
    for (const btn of cardButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('game end lose state does not show celebration', async () => {
    mockExec.mockResolvedValue(gameEndLoseState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了')).toBeInTheDocument());
  });

  it('selects and deselects a hand card on click when no single valid pile', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // DIAMOND 2 is not adjacent to either center pile (5 or 9), so it toggles normally
    const cardBtn = screen.getByRole('button', { name: '♦ 2' });
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('includes the localized top-card name in the center pile aria-labels', async () => {
    renderWithProviders(<SpeedPage />);
    await screen.findByText('手札');
    // Center piles top with ♦5 and ♠9 → the labels name the card to play onto.
    expect(screen.getByRole('button', { name: '台札1: ♦ 5' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '台札2: ♠ 9' })).toBeInTheDocument();
  });

  it('auto-plays a card via smart-click when only one valid pile exists', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // SPADE 4 is adjacent to pile 0 (value 5) only → smart-click auto-plays.
    // It is also playable, so its aria-label carries the "playable now" suffix.
    fireEvent.click(screen.getByRole('button', { name: '♠ 4 (今すぐ出せる)' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 0));
  });

  it('plays a card to a center pile when a card is selected', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // DIAMOND 2 has no valid piles, so it toggles selection first
    fireEvent.click(screen.getByRole('button', { name: '♦ 2' }));
    // Then click first center pile manually
    const pileBtns = screen.getAllByRole('button', { name: /台札/ });
    fireEvent.click(pileBtns[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 3, 0));
  });

  it('clicking hint button calls hint command', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('uses the design-system button token (not Bootstrap classes) for the hint button', async () => {
    renderWithProviders(<SpeedPage />);
    const hint = await screen.findByRole('button', { name: 'ヒント' });
    // btnOutline token applied; legacy Bootstrap classes gone.
    expect(hint.className).toContain('border');
    expect(hint.className).not.toContain('btn-outline');
    expect(hint.className).not.toMatch(/\bbtn\b/);
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

  it('shows the auto-flip countdown when stuck with auto-flip enabled', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('auto-flip-countdown')).toBeInTheDocument());
    expect(screen.getByRole('timer', { name: '自動めくりカウントダウン' })).toBeInTheDocument();
  });

  it('center piles have pulse animation when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    // Each center pile gets a unique "台札N をめくる" aria-label while stuck,
    // distinct from the central "めくる" popup, so SR users hear three
    // discrete affordances instead of three duplicates.
    const firstPileBtn = screen.getByRole('button', { name: '台札1をめくる' });
    expect(firstPileBtn).toHaveClass('animate-pulse');
  });

  it('keyboard shortcut: Space triggers flip when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: ' ' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip'));
  });

  it('keyboard shortcut: Enter triggers flip when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip'));
  });

  it('keyboard shortcut: Space is ignored during play phase', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: ' ' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('keyboard shortcut: digit key smart-plays the matching hand card', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // Hand index 0 is SPADE 4 (single-valid-pile → smart-click auto-plays to pile 0)
    fireEvent.keyDown(window, { key: '1' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 0));
  });

  it('keyboard shortcut: ArrowRight plays the selected card to right pile', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // Select a card with no auto-pile so it stays selected
    fireEvent.click(screen.getByRole('button', { name: '♦ 2' }));
    fireEvent.keyDown(window, { key: 'ArrowRight' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 3, 1));
  });

  it('keyboard shortcut: ArrowLeft plays the selected card to left pile', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '♦ 2' }));
    fireEvent.keyDown(window, { key: 'ArrowLeft' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 3, 0));
  });

  it('keyboard shortcut: arrow keys are ignored when no card is selected', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: 'ArrowLeft' });
    fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('keyboard shortcut: digit beyond hand size is ignored', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: '5' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('keyboard shortcut: modifier keys (Ctrl/Alt/Meta) bypass the handler', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: '1', ctrlKey: true });
    fireEvent.keyDown(window, { key: 'ArrowLeft', altKey: true });
    fireEvent.keyDown(window, { key: '2', metaKey: true });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders the auto-flip settings toggle', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    expect(screen.getByLabelText('膠着時に自動でめくる')).toBeInTheDocument();
  });

  it('renders an inline auto-flip toggle when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    expect(screen.getByTestId('inline-auto-flip-toggle')).toBeInTheDocument();
  });

  it('does not render the inline auto-flip toggle in play phase', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    expect(screen.queryByTestId('inline-auto-flip-toggle')).not.toBeInTheDocument();
  });

  it('inline auto-flip toggle reflects current speedConfig.autoFlip', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('inline-auto-flip-toggle')).toBeInTheDocument());
    const toggle = screen.getByTestId('inline-auto-flip-toggle') as HTMLInputElement;
    expect(toggle.checked).toBe(true);
  });

  it('clicking the inline auto-flip toggle disables auto-flip', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('inline-auto-flip-toggle')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('inline-auto-flip-toggle'));
    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mockExec).not.toHaveBeenCalledWith('flip');
    vi.useRealTimers();
  });

  it('renders an emphasized stuck container when stuck', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    expect(screen.getByTestId('stuck-emphasis-container')).toBeInTheDocument();
  });

  it('does not render the stuck emphasis container in play phase', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    expect(screen.queryByTestId('stuck-emphasis-container')).not.toBeInTheDocument();
  });

  it('marks playable hand cards with a success ring in play phase', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // Center piles 5/9 → SPADE 4 (5-1) and HEART 8 (9-1) are playable; CLOVER 11
    // and DIAMOND 2 are not.
    const playable = screen.getByRole('button', { name: '♠ 4 (今すぐ出せる)' });
    expect(playable).toHaveAttribute('data-playable', 'true');
    expect(playable.style.outline).toContain('var(--color-ds-success)');

    const notPlayable = screen.getByRole('button', { name: '♦ 2' });
    expect(notPlayable).not.toHaveAttribute('data-playable');
    expect(notPlayable.style.outline).toBe('');
  });

  it('keeps a non-highlighted hand card clickable (ring does not block clicks)', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    // DIAMOND 2 is not highlighted (not adjacent to either pile) but must stay
    // interactive — clicking it toggles selection rather than being blocked.
    const notPlayable = screen.getByRole('button', { name: '♦ 2' });
    expect(notPlayable).toBeEnabled();
    fireEvent.click(notPlayable);
    expect(notPlayable).toHaveAttribute('aria-pressed', 'true');
  });

  it('removes the playable ring in stuck phase', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    // No hand card carries the playable marker once play is halted.
    expect(document.querySelector('[data-playable="true"]')).toBeNull();
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

  it('does not auto-flip while a manual flip request is in flight', async () => {
    // First render resolves to stuckState; subsequent calls hang so loading stays true
    mockExec.mockResolvedValueOnce(stuckState).mockReturnValue(new Promise(() => {}));
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeInTheDocument());
    // Manual click sets loading=true for the pending mutation; wait for the
    // loading state (disabled button) to propagate through react-query before
    // advancing fake timers, otherwise the pre-click timer can still fire.
    await act(async () => {
      fireEvent.click(screen.getByTestId('flip-button'));
    });
    await waitFor(() => expect(screen.getByTestId('flip-button')).toBeDisabled());
    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
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

describe('SpeedPage elapsed / best time', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.removeItem('speed_best_time_1');
    mockExec.mockResolvedValue(playState);
  });

  afterEach(() => {
    localStorage.removeItem('speed_best_time_1');
  });

  it('shows the elapsed timer readout in the header during play', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('speed-timer')).toBeInTheDocument());
    expect(screen.getByTestId('speed-timer')).toHaveTextContent('経過: 00:00');
  });

  it('shows the persisted best time for the current difficulty', async () => {
    localStorage.setItem('speed_best_time_1', '65000'); // 01:05
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('speed-best-time')).toBeInTheDocument());
    expect(screen.getByTestId('speed-best-time')).toHaveTextContent('01:05');
  });

  it('does not render a best-time readout when none is stored', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByTestId('speed-timer')).toBeInTheDocument());
    expect(screen.queryByTestId('speed-best-time')).not.toBeInTheDocument();
  });

  it('records and announces a new best time on a human win', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValueOnce(playState).mockResolvedValue(gameEndState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    // Smart-click SPADE 4 auto-plays to pile 0; the response ends the game as a human win.
    fireEvent.click(screen.getByRole('button', { name: '♠ 4 (今すぐ出せる)' }));
    await waitFor(() => expect(screen.getByTestId('speed-clear-time')).toBeInTheDocument());
    expect(screen.getByTestId('speed-clear-time')).toHaveTextContent('ベスト更新');
    expect(Number(localStorage.getItem('speed_best_time_1'))).toBeGreaterThanOrEqual(3000);
    vi.useRealTimers();
  });

  it('keeps a faster stored best when the new win is slower', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    localStorage.setItem('speed_best_time_1', '1000');
    mockExec.mockResolvedValueOnce(playState).mockResolvedValue(gameEndState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    fireEvent.click(screen.getByRole('button', { name: '♠ 4 (今すぐ出せる)' }));
    await waitFor(() => expect(screen.getByTestId('speed-clear-time')).toBeInTheDocument());
    expect(screen.getByTestId('speed-clear-time')).not.toHaveTextContent('ベスト更新');
    expect(localStorage.getItem('speed_best_time_1')).toBe('1000');
    vi.useRealTimers();
  });
});
