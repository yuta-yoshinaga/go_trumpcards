import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bourreApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BourrePlayerData, BourreResponse } from '../types/card';
import { BourrePage } from './BourrePage';

vi.mock('../api/gameApi', () => ({
  bourreApi: { exec: vi.fn() },
  actionLogApi: { bourre: vi.fn() },
}));

const mockExec = vi.mocked(bourreApi.exec);

function player(overrides: Partial<BourrePlayerData>): BourrePlayerData {
  return {
    id: 0,
    isHuman: false,
    isFinished: false,
    folded: false,
    decided: false,
    drawn: false,
    bourreed: false,
    chips: 100,
    tricks: 0,
    cardCount: 5,
    cards: [],
    ...overrides,
  };
}

function makeState(overrides: Partial<BourreResponse> = {}): BourreResponse {
  return {
    players: [
      player({ id: 0, isHuman: true }),
      player({ id: 1 }),
      player({ id: 2 }),
      player({ id: 3 }),
      player({ id: 4 }),
    ],
    phase: 'decide',
    currentPlayerIdx: 0,
    dealerIdx: 4,
    pot: 25,
    carryPot: 0,
    trumpSuit: 'SPADE',
    trumpCard: { design: 'SPADE', value: 5 },
    trickNumber: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    leadPlayerIdx: -1,
    handNumber: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    validPlays: [],
    results: [],
    config: { cpuDifficulty: 0 },
    message: '',
    ...overrides,
  } as BourreResponse;
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BourrePage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BourrePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'reset' }));
    });
  });

  it('renders CPU player areas after load', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('decide phase: play and fold buttons dispatch decide', async () => {
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '参加（アンティ）' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'decide', decide: true }));
    });
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'decide', decide: false }));
    });
  });

  it('play phase: clicking a legal card dispatches play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'play',
        currentPlayerIdx: 0,
        validPlays: [0],
        players: [
          player({
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 9 },
              { design: 'CLOVER', value: 5 },
            ],
          }),
          player({ id: 1 }),
          player({ id: 2 }),
          player({ id: 3 }),
          player({ id: 4 }),
        ],
      }),
    );
    const { container } = renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBtn = container.querySelector('[data-tutorial="bourre-hand"] button');
    expect(cardBtn).not.toBeNull();
    fireEvent.click(cardBtn as Element);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', cardIndex: 0 }));
    });
  });

  it('roundEnd phase: next hand button dispatches next', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        results: [
          { playerIdx: 0, tricks: 3, wonAmount: 25, bourreed: false, folded: false },
          { playerIdx: 1, tricks: 0, wonAmount: 0, bourreed: true, folded: false },
        ],
      }),
    );
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のハンド' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'next' }));
    });
  });

  it('end phase: shows the win banner when the human wins', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: 0,
        results: [{ playerIdx: 0, tricks: 5, wonAmount: 50, bourreed: false, folded: false }],
      }),
    );
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(screen.getByText(/あなたの勝ち/)).toBeInTheDocument();
    });
  });

  it('toggles CLI mode', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cliToggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(cliToggle);
    await waitFor(() => expect(screen.queryByText(/CPU 1/)).not.toBeInTheDocument());
  });
});
