import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tichuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { TichuPlayerData, TichuResponse } from '../types/card';
import { TichuPage } from './TichuPage';

vi.mock('../api/gameApi', () => ({
  tichuApi: { exec: vi.fn() },
  actionLogApi: { tichu: vi.fn() },
}));

const mockExec = vi.mocked(tichuApi.exec);

function player(overrides: Partial<TichuPlayerData>): TichuPlayerData {
  return {
    id: 0,
    isHuman: false,
    isFinished: false,
    team: 0,
    rank: 0,
    declType: 0,
    cardCount: 14,
    cards: [],
    ...overrides,
  };
}

function makeState(overrides: Partial<TichuResponse> = {}): TichuResponse {
  return {
    players: [
      player({ id: 0, isHuman: true, team: 0 }),
      player({ id: 1, team: 1 }),
      player({ id: 2, team: 0 }),
      player({ id: 3, team: 1 }),
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

  it('renders CPU player areas and declaration labels after load', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player({ id: 0, isHuman: true, team: 0 }),
          player({ id: 1, team: 1, declType: 1 }),
          player({ id: 2, team: 0, declType: 2 }),
          player({ id: 3, team: 1 }),
        ],
      }),
    );
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('declaration phase: all three buttons dispatch declare', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'declare', currentTurn: 0 }));
    renderWithProviders(<TichuPage />);
    fireEvent.click(await screen.findByRole('button', { name: '宣言しない' }));
    fireEvent.click(screen.getByRole('button', { name: 'ティチュー宣言' }));
    fireEvent.click(screen.getByRole('button', { name: 'グランド宣言' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'declare', declType: 2 }));
    });
  });

  it('play phase: selecting a card enables Play, and Pass is shown when following', async () => {
    mockExec.mockResolvedValue(
      makeState({
        tableCards: [{ design: 'HEART', value: 7 }],
        tableCombo: 'single',
        lastPlayIdx: 1,
        players: [
          player({
            id: 0,
            isHuman: true,
            team: 0,
            cardCount: 2,
            cards: [
              { design: 'SPADE', value: 9 },
              { design: 'CLOVER', value: 5 },
            ],
          }),
          player({ id: 1, team: 1 }),
          player({ id: 2, team: 0 }),
          player({ id: 3, team: 1 }),
        ],
      }),
    );
    const { container } = renderWithProviders(<TichuPage />);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    expect(playBtn).toBeDisabled();

    const cardBtn = container.querySelector('[data-tutorial="tichu-hand"] button');
    expect(cardBtn).not.toBeNull();
    fireEvent.click(cardBtn as Element);
    expect(screen.getByRole('button', { name: '出す' })).toBeEnabled();

    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p' }));
    });

    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', indices: [] }));
    });
  });

  it('end phase: shows team scores, the win banner, and the one-two note', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'end', gameEndFlag: true, isOneTwo: true, scores: [200, -100] }));
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(screen.getByText(/チームA/)).toBeInTheDocument();
    });
    expect(screen.getByText(/チームB/)).toBeInTheDocument();
    expect(screen.getByText(/あなたのチームの勝利/)).toBeInTheDocument();
    expect(screen.getByText(/ワンツー/)).toBeInTheDocument();
  });

  it('end phase: shows the losing banner when opponents win', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'end', gameEndFlag: true, scores: [0, 100] }));
    renderWithProviders(<TichuPage />);
    await waitFor(() => {
      expect(screen.getByText(/相手チームの勝利/)).toBeInTheDocument();
    });
  });

  it('toggles CLI mode', async () => {
    renderWithProviders(<TichuPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cliToggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(cliToggle);
    // toggling hides the GUI CPU areas in favour of the CLI terminal
    await waitFor(() => expect(screen.queryByText(/CPU 1/)).not.toBeInTheDocument());
  });
});
