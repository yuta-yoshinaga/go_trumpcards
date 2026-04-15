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

  it('renders play and draw buttons when human turn', async () => {
    renderWithProviders(<PageOnePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
    });
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
});
