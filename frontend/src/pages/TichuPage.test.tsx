import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tichuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { TichuResponse } from '../types/card';
import { TichuPage } from './TichuPage';

vi.mock('../api/gameApi', () => ({
  tichuApi: { exec: vi.fn() },
  actionLogApi: { tichu: vi.fn() },
}));

const mockExec = vi.mocked(tichuApi.exec);

function makeState(overrides: Partial<TichuResponse> = {}): TichuResponse {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, team: 0, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 1, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 2, isHuman: false, isFinished: false, team: 0, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 3, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
    ],
    phase: 'play',
    currentTurn: 0,
    tableCards: [],
    tableCombo: '',
    lastPlayIdx: -1,
    startLeader: 0,
    finishOrder: [],
    scores: [0, 0],
    isOneTwo: false,
    bombCount: 0,
    gameEndFlag: false,
    config: { cpuDifficulty: 0 },
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  } as TichuResponse;
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('TichuPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TichuPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'reset' }));
    });
  });

  it('renders CPU player areas after load', async () => {
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('shows declaration buttons during declare phase on human turn', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'declare', currentTurn: 0 }));
    renderWithProviders(<TichuPage />);
    const tichuBtn = await screen.findByRole('button', { name: 'ティチュー宣言' });
    fireEvent.click(tichuBtn);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'declare' }));
    });
  });

  it('plays selected cards in play phase', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            isFinished: false,
            team: 0,
            rank: 0,
            declType: 0,
            cardCount: 1,
            cards: [{ design: 'SPADE', value: 9 }],
          },
          { id: 1, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
          { id: 2, isHuman: false, isFinished: false, team: 0, rank: 0, declType: 0, cardCount: 14, cards: [] },
          { id: 3, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
        ],
      }),
    );
    renderWithProviders(<TichuPage />);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    expect(playBtn).toBeDisabled();
  });

  it('shows team scores at game end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'end', gameEndFlag: true, scores: [120, -20] }));
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(screen.getByText(/チームA/)).toBeInTheDocument();
    });
  });
});
