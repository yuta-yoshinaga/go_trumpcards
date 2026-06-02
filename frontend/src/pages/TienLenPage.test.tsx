import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tienlenApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TienLenResponse } from '../types/card';
import { TienLenPage } from './TienLenPage';

vi.mock('../api/gameApi', () => ({
  tienlenApi: { exec: vi.fn() },
  actionLogApi: { tienlen: vi.fn() },
}));

const mockExec = vi.mocked(tienlenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<TienLenResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function makeState(overrides: Partial<TienLenResponse> = {}): TienLenResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('TienLenPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<TienLenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows play and pass buttons on the human turn', async () => {
    renderWithProviders(<TienLenPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
    expect(screen.getByTestId('play-button')).toBeDisabled(); // nothing selected yet
  });

  it('selecting a card enables play and clicking plays it', async () => {
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('passes when the pass button is clicked', async () => {
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<TienLenPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-tienlen', 'true');
    renderWithProviders(<TienLenPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('play-button')).not.toBeInTheDocument();
  });

  it('shows a retry button when an action fails', async () => {
    renderWithProviders(<TienLenPage />);
    const passBtn = await screen.findByTestId('pass-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(passBtn);
    const retry = await screen.findByText(NETWORK_ERROR_MESSAGE());
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(retry);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });
});
