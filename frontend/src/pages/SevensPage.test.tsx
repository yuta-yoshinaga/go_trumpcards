import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevensApi } from '../api/gameApi';
import type { SevensResponse } from '../types/card';
import { SevensPage } from './SevensPage';

vi.mock('../api/gameApi', () => ({
  sevensApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(sevensApi.exec);

// tableMinVals/tableMaxVals: index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
// With all 7s placed: min=7, max=7 → value 6 or 8 of any suit is playable
const humanTurnState: SevensResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      rank: -1,
      cardCount: 3,
      passesUsed: 0,
      maxPasses: 5,
      cards: [
        { design: 'SPADE', value: 6 }, // playable: 6 == 7-1
        { design: 'HEART', value: 5 }, // not playable
        { design: 'CLOVER', value: 8 }, // playable: 8 == 7+1
      ],
    },
    { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 4, passesUsed: 0, maxPasses: 5, cards: [] },
    { id: 2, isHuman: false, isFinished: false, rank: -1, cardCount: 3, passesUsed: 1, maxPasses: 5, cards: [] },
    { id: 3, isHuman: false, isFinished: false, rank: -1, cardCount: 5, passesUsed: 0, maxPasses: 5, cards: [] },
  ],
  currentTurn: 0,
  tableMinVals: [0, 7, 7, 7, 7],
  tableMaxVals: [0, 7, 7, 7, 7],
  gameEndFlag: false,
  cpuActions: [],
  humanAction: null,
  message: '',
};

const cpuTurnState: SevensResponse = {
  ...humanTurnState,
  currentTurn: 1,
  humanAction: { playerIdx: 0, playedCard: { design: 'SPADE', value: 6 } },
  tableMinVals: [0, 6, 7, 7, 7],
  tableMaxVals: [0, 7, 7, 7, 7],
};

const gameEndState: SevensResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  message: '1位: あなた',
};

const passesExhaustedState: SevensResponse = {
  ...humanTurnState,
  players: [{ ...humanTurnState.players[0], passesUsed: 5, maxPasses: 5 }, ...humanTurnState.players.slice(1)],
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('SevensPage', () => {
  it('renders nothing before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = render(<SevensPage />);
    expect(container.firstChild).toBeNull();
  });

  it('calls reset command on mount', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', -1));
  });

  it('renders human player area labeled あなた', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
  });

  it('renders CPU player areas with correct labels', async () => {
    render(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('renders the board section', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
  });

  it('shows human player face-up cards', async () => {
    render(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByAltText('SPADE 6')).toBeInTheDocument();
      expect(screen.getByAltText('HEART 5')).toBeInTheDocument();
      expect(screen.getByAltText('CLOVER 8')).toBeInTheDocument();
    });
  });

  it('pass button is enabled on human turn with passes remaining', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('pass button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when passes are exhausted', async () => {
    mockExec.mockResolvedValue(passesExhaustedState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('calls play with -1 when pass button is clicked', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', -1));
  });

  it('calls play with card index when a playable card is clicked', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 6')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByAltText('SPADE 6'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('does not call play when a non-playable card is clicked', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('HEART 5')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByAltText('HEART 5'));
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('calls reset when reset button is clicked', async () => {
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', -1));
  });

  it('shows human action log after play', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたが出しました/)).toBeInTheDocument());
  });

  it('shows pass action log for human pass', async () => {
    const passState: SevensResponse = {
      ...cpuTurnState,
      humanAction: { playerIdx: 0, playedCard: null },
    };
    mockExec.mockResolvedValue(passState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたがパスしました/)).toBeInTheDocument());
  });

  it('shows CPU action log when cpuActions is non-empty', async () => {
    const stateWithCpuActions: SevensResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, playedCard: { design: 'SPADE', value: 8 } },
        { playerIdx: 2, playedCard: null },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument();
  });

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText('1位: あなた')).toBeInTheDocument());
  });

  it('shows rank badge for finished players', async () => {
    const stateWithFinished: SevensResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, passesUsed: 0, maxPasses: 5, cards: [] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithFinished);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText('1位')).toBeInTheDocument());
  });

  it('shows thinking indicator on current CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<SevensPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
  });

  it('shows pass counts for CPU players', async () => {
    render(<SevensPage />);
    await waitFor(() => {
      // CPU 2 has passesUsed=1
      expect(screen.getByText(/パス: 1\/5/)).toBeInTheDocument();
    });
  });
});
