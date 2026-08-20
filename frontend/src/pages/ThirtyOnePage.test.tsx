import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { thirtyoneApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ThirtyOneResponse } from '../types/card';
import { ThirtyOnePhase } from '../types/phases';
import { ThirtyOnePage } from './ThirtyOnePage';

vi.mock('../api/gameApi', () => ({
  thirtyoneApi: { exec: vi.fn() },
  actionLogApi: { thirtyone: vi.fn() },
}));

const mockExec = vi.mocked(thirtyoneApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<ThirtyOneResponse['players'][number]> = {}) {
  return { id, isHuman, cardCount: cards.length, cards, lives: 3, score: 0, isEliminated: false, ...over };
}

function makeState(overrides: Partial<ThirtyOneResponse> = {}): ThirtyOneResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    phase: ThirtyOnePhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 2),
    drawPileCount: 39,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    thirtyOneIdx: -1,
    roundWinnerIdx: -1,
    roundLosers: [],
    message: '',
    config: { cpuDifficulty: 1, initialLives: 3, knockThresholds: { easy: 29, normal: 27, hard: 25 } },
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('ThirtyOnePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows draw and knock actions during the draw phase', async () => {
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('draw-stock-button')).toBeEnabled();
    expect(screen.getByTestId('draw-discard-button')).toBeEnabled();
    expect(screen.getByTestId('knock-button')).toBeEnabled();
  });

  it('draws from stock when the button is clicked', async () => {
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('draw-stock-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('knocks when the knock button is clicked', async () => {
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('knock-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock'));
  });

  it('allows selecting and discarding a card in the discard phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.DISCARD }));
    renderWithProviders(<ThirtyOnePage />);
    const discardBtn = await screen.findByTestId('discard-button');
    expect(discardBtn).toBeDisabled();

    // Select the first hand card, then discard.
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('discard-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('shows the next-round button at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.ROUND_END, roundWinnerIdx: 0, roundLosers: [1] }));
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('summarizes life losses at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.ROUND_END, roundWinnerIdx: 0, roundLosers: [1] }));
    renderWithProviders(<ThirtyOnePage />);
    const summary = await screen.findByTestId('thirtyone-round-summary');
    expect(summary).toHaveTextContent('CPU 1 がライフを1つ失った');
    // No 31 achiever and player 1 not eliminated in this fixture.
    expect(screen.queryByTestId('thirtyone-achiever')).not.toBeInTheDocument();
    expect(screen.queryByTestId('eliminated-1')).not.toBeInTheDocument();
  });

  it('highlights the 31 achiever and newly eliminated players in the summary', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: ThirtyOnePhase.ROUND_END,
        thirtyOneIdx: 2,
        roundLosers: [1, 3],
        players: [
          player(0, true, [card('SPADE', 3)]),
          player(1, false, [], { lives: 1 }),
          player(2, false, []),
          player(3, false, [], { lives: 0, isEliminated: true }),
        ],
      }),
    );
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('thirtyone-achiever')).toHaveTextContent('CPU 2 が31達成');
    // Player 3 dropped to 0 lives → shown as eliminated; player 1 merely lost a life.
    expect(screen.getByTestId('eliminated-3')).toBeInTheDocument();
    expect(screen.queryByTestId('eliminated-1')).not.toBeInTheDocument();
  });

  it('states when nobody lost a life at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.ROUND_END, roundLosers: [] }));
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('no-life-loss')).toBeInTheDocument();
  });

  it('reveals CPU lives at all times', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-stock-button');
    // Human + 3 CPU each render a lives indicator (❤ or 💀).
    expect(screen.getAllByLabelText(/lives-/).length).toBeGreaterThanOrEqual(1);
  });

  it('toggles card selection on and off in the discard phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.DISCARD }));
    renderWithProviders(<ThirtyOnePage />);
    const card0 = await screen.findByTestId('hand-card-0');
    // Select then re-click to deselect → discard button disabled again.
    fireEvent.click(card0);
    expect(screen.getByTestId('discard-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('discard-button')).toBeDisabled();
  });

  it('disables the knock button once someone has knocked', async () => {
    mockExec.mockResolvedValue(makeState({ knockerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('knock-button')).toBeDisabled();
  });

  it('shows the knock countdown banner with the knocker name on the human last turn', async () => {
    mockExec.mockResolvedValue(makeState({ knockerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    const banner = await screen.findByTestId('knock-countdown-banner');
    expect(banner).toHaveTextContent(/CPU 1/);
    expect(banner).toHaveTextContent(/最後のターン/);
  });

  it('shows the active knock banner without last-turn text when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ knockerIdx: 2, currentPlayerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    const banner = await screen.findByTestId('knock-countdown-banner');
    expect(banner).toHaveTextContent(/CPU 2/);
    expect(banner).not.toHaveTextContent(/最後のターン/);
  });

  it('hides the knock countdown banner at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.ROUND_END, knockerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('next-round-button');
    expect(screen.queryByTestId('knock-countdown-banner')).not.toBeInTheDocument();
  });

  it('names the human in the knock banner when the human is the knocker', async () => {
    mockExec.mockResolvedValue(makeState({ knockerIdx: 0, currentPlayerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    const banner = await screen.findByTestId('knock-countdown-banner');
    expect(banner).toHaveTextContent(/あなた/);
    expect(banner).not.toHaveTextContent(/最後のターン/);
  });

  it('hides the knock countdown banner at game end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.GAME_END, gameEndFlag: true, knockerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    await waitFor(() => expect(screen.getAllByLabelText(/lives-|out/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('knock-countdown-banner')).not.toBeInTheDocument();
  });

  it('highlights the leading suit badge', async () => {
    // Default human hand: SPADE 3, HEART 5, DIAMOND 7 → DIAMOND leads.
    renderWithProviders(<ThirtyOnePage />);
    const leader = await screen.findByTestId('suit-badge-DIAMOND');
    expect(leader.className).toContain('bg-ds-accent');
  });

  it('reveals CPU cards and scores and the next-round button at round end', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: ThirtyOnePhase.ROUND_END,
        roundWinnerIdx: 0,
        roundLosers: [1],
        players: [
          player(0, true, [card('SPADE', 1), card('SPADE', 13), card('SPADE', 5)], { score: 26 }),
          player(1, false, [card('HEART', 2), card('CLOVER', 3), card('DIAMOND', 4)], { score: 4 }),
          player(2, false, [card('SPADE', 9), card('HEART', 9), card('CLOVER', 2)], { score: 9 }),
          player(3, false, [card('DIAMOND', 7), card('DIAMOND', 8), card('SPADE', 1)], { score: 15 }),
        ],
      }),
    );
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('next-round-button')).toBeInTheDocument();
    // At round end, draw actions are disabled (not the human's turn).
    expect(screen.getByTestId('draw-stock-button')).toBeDisabled();
  });

  it('renders eliminated and zero-life indicators', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)], { lives: 0 }),
          player(1, false, [], { lives: -1, isEliminated: true }),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-stock-button');
    expect(screen.getByLabelText('out')).toBeInTheDocument(); // 💀 for eliminated
    expect(screen.getByLabelText('lives-0')).toBeInTheDocument(); // · fallback
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-thirtyone', 'true');
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('draw-stock-button')).not.toBeInTheDocument();
  });

  it('advertises keyboard shortcuts on the action buttons', async () => {
    renderWithProviders(<ThirtyOnePage />);
    const drawStock = await screen.findByTestId('draw-stock-button');
    expect(drawStock).toHaveAttribute('aria-keyshortcuts', 's');
    expect(drawStock).toHaveTextContent('S');
    expect(screen.getByTestId('draw-discard-button')).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(screen.getByTestId('knock-button')).toHaveAttribute('aria-keyshortcuts', 'k');
    expect(screen.getByTestId('discard-button')).toHaveAttribute('aria-keyshortcuts', 'x');
  });

  it('draws from stock when the "s" key is pressed', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-stock-button');
    fireEvent.keyDown(document.body, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('draws from the discard pile when the "d" key is pressed', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-discard-button');
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('knocks when the "k" key is pressed', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('knock-button');
    fireEvent.keyDown(document.body, { key: 'k' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock'));
  });

  it('discards the selected card when the "x" key is pressed in the discard phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.DISCARD }));
    renderWithProviders(<ThirtyOnePage />);
    // The key is inert until a card is selected (mirrors the disabled button).
    fireEvent.keyDown(document.body, { key: 'x' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('discard', expect.anything());
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.keyDown(document.body, { key: 'x' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('ignores action keys when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-stock-button');
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 's' });
    fireEvent.keyDown(document.body, { key: 'k' });
    await waitFor(() => expect(mockExec).not.toHaveBeenCalled());
  });

  it('shows a retry button when an action fails', async () => {
    renderWithProviders(<ThirtyOnePage />);
    const drawBtn = await screen.findByTestId('draw-stock-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(drawBtn);
    const retry = await screen.findByText(NETWORK_ERROR_MESSAGE());
    // Clicking it retries the last action (which now succeeds again).
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(retry);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });
});

// #5623: 難易度セレクトは Easy/Normal/Hard としか言わず、何が変わるのか
// (CPU がノックしてくる点数) は体験からしか学べなかった。数字はサーバーが
// 運んでくるので、説明文に書き写さない。
describe('ThirtyOnePage difficulty help', () => {
  it('explains the difficulty with the thresholds the server sent', async () => {
    mockExec.mockResolvedValue(
      makeState({ config: { cpuDifficulty: 1, initialLives: 3, knockThresholds: { easy: 29, normal: 27, hard: 25 } } }),
    );
    renderWithProviders(<ThirtyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getByText('設定'));
    fireEvent.click(screen.getAllByRole('button', { name: '説明を表示' })[0] as HTMLElement);

    const tip = await screen.findByRole('tooltip');
    expect(tip).toHaveTextContent('29');
    expect(tip).toHaveTextContent('27');
    expect(tip).toHaveTextContent('25');
  });

  // **サーバーの値を出す。**画面に焼き込んだ数字だと、定数が動いても気づけない。
  it('follows the server when the thresholds differ', async () => {
    mockExec.mockResolvedValue(
      makeState({ config: { cpuDifficulty: 1, initialLives: 3, knockThresholds: { easy: 31, normal: 30, hard: 28 } } }),
    );
    renderWithProviders(<ThirtyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getByText('設定'));
    fireEvent.click(screen.getAllByRole('button', { name: '説明を表示' })[0] as HTMLElement);

    const tip = await screen.findByRole('tooltip');
    expect(tip).toHaveTextContent('31');
    expect(tip).not.toHaveTextContent('29');
  });
});
