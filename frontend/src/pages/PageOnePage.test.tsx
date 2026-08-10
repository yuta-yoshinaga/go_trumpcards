import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pageoneApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PageOneResponse } from '../types/card';
import { PageOnePage } from './PageOnePage';

vi.mock('../api/gameApi', () => ({
  pageoneApi: { exec: vi.fn() },
  actionLogApi: { pageone: vi.fn() },
}));

const mockExec = vi.mocked(pageoneApi.exec);

const playPhaseState: PageOneResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 3, cumulativeScore: 10, hasDeclared: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 5, cumulativeScore: 20, hasDeclared: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 5, hasDeclared: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

const mustDeclareState: PageOneResponse = {
  ...playPhaseState,
  phase: 1,
  players: [
    { ...playPhaseState.players[0], cardCount: 1, cards: [{ design: 'SPADE', value: 1 }] },
    ...playPhaseState.players.slice(1),
  ],
};

const roundEndState: PageOneResponse = { ...playPhaseState, phase: 2 };

const gameEndState: PageOneResponse = {
  ...playPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: PageOneResponse = { ...playPhaseState, currentPlayerIdx: 1 };

const cpuAtOneCardState: PageOneResponse = {
  ...playPhaseState,
  players: [
    playPhaseState.players[0],
    { ...playPhaseState.players[1], cardCount: 1 },
    ...playPhaseState.players.slice(2),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('PageOnePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PageOnePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PageOnePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
      }),
    );
  });

  it('exposes the CPU last-card badge to assistive tech via a polite live region', async () => {
    mockExec.mockResolvedValue(cpuAtOneCardState);
    renderWithProviders(<PageOnePage />);
    const badge = await screen.findByTestId('po-cpu-1-last-card-badge');
    expect(badge).toHaveAttribute('role', 'status');
    expect(badge).toHaveAttribute('aria-live', 'polite');
    // The accessible name names the CPU so the announcement has context.
    expect(badge).toHaveAttribute('aria-label', expect.stringContaining('残り1枚'));
  });

  it('renders play and draw buttons when human turn', async () => {
    renderWithProviders(<PageOnePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
    });
  });

  it('shows the play-condition badge under the discard top', async () => {
    renderWithProviders(<PageOnePage />);
    // discardTop is ♥7 → playable cards are hearts or 7s.
    await waitFor(() => expect(screen.getByText('出せる条件: ♥ または 7')).toBeInTheDocument());
  });

  it('hides the play-condition badge when the discard pile is empty', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, discardTop: null });
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByText(/出せる条件/)).not.toBeInTheDocument();
  });

  it('does not show play/draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows declare/skip buttons during MustDeclare phase', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ページワン！' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'スキップ（ペナルティ）' })).toBeInTheDocument();
    });
  });

  it('calls declare command when declare button is clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ページワン！' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ページワン！' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('calls skip command when skip button is clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ（ペナルティ）' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ（ペナルティ）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip'));
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders game end state', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
  });

  it('shows the last-card warning banner when the human is at 1 undeclared card during play', async () => {
    const oneCardPlayState: PageOneResponse = {
      ...playPhaseState,
      players: [
        { ...playPhaseState.players[0], cardCount: 1, cards: [{ design: 'SPADE', value: 1 }] },
        ...playPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(oneCardPlayState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByTestId('po-last-card-banner')).toBeInTheDocument());
    expect(screen.getByTestId('po-last-card-banner')).toHaveAttribute('aria-live', 'assertive');
  });

  it('hides the last-card banner once the human has declared', async () => {
    const declaredState: PageOneResponse = {
      ...playPhaseState,
      players: [
        { ...playPhaseState.players[0], cardCount: 1, cards: [{ design: 'SPADE', value: 1 }], hasDeclared: true },
        ...playPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(declaredState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('po-last-card-banner')).not.toBeInTheDocument();
  });

  it('pulses the declare button when the human must declare', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByTestId('po-declare-btn')).toBeInTheDocument());
    expect(screen.getByTestId('po-declare-btn').className).toContain('animate-pulse');
  });

  it('highlights a CPU at 1 card with the warning badge', async () => {
    const cpuOneCardState: PageOneResponse = {
      ...playPhaseState,
      players: [
        playPhaseState.players[0],
        { ...playPhaseState.players[1], cardCount: 1 },
        playPhaseState.players[2],
        playPhaseState.players[3],
      ],
    };
    mockExec.mockResolvedValue(cpuOneCardState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByTestId('po-cpu-1-last-card-badge')).toBeInTheDocument());
    expect(screen.queryByTestId('po-cpu-2-last-card-badge')).not.toBeInTheDocument();
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
      }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('rings the cards that may be played on the discard top', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PageOnePage />);
    await waitFor(() => expect(document.querySelectorAll('[data-playable]').length).toBeGreaterThan(0));
    // Only matching cards, never the whole hand.
    const hand = playPhaseState.players.find((p) => p.isHuman)?.cards ?? [];
    expect(document.querySelectorAll('[data-playable]').length).toBeLessThan(hand.length);
    // The ring is decoration on a real control, so selection still works.
    const first = document.querySelector('[data-playable]') as HTMLButtonElement;
    fireEvent.click(first);
    await waitFor(() => expect(first).toHaveAttribute('aria-pressed', 'true'));
  });
});
