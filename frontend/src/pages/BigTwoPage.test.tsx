import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bigtwoApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BigTwoResponse, Card } from '../types/card';
import { BigTwoPage } from './BigTwoPage';

vi.mock('../api/gameApi', () => ({
  bigtwoApi: { exec: vi.fn() },
  actionLogApi: { bigtwo: vi.fn() },
}));

const mockExec = vi.mocked(bigtwoApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<BigTwoResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function makeState(overrides: Partial<BigTwoResponse> = {}): BigTwoResponse {
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

describe('BigTwoPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BigTwoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders skeleton when fewer than 4 players are present', async () => {
    mockExec.mockResolvedValue(makeState({ players: [player(0, true, [])] }));
    renderWithProviders(<BigTwoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BigTwoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows play and pass buttons on the human turn', async () => {
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
    expect(screen.getByTestId('play-button')).toBeDisabled(); // nothing selected yet
  });

  it('selecting a card enables play and clicking plays it', async () => {
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('passes when the pass button is clicked', async () => {
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<BigTwoPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-bigtwo', 'true');
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('play-button')).not.toBeInTheDocument();
  });

  it('shows the shared ErrorAlert with a retry button when an action fails', async () => {
    renderWithProviders(<BigTwoPage />);
    const passBtn = await screen.findByTestId('pass-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(passBtn);
    // The error now surfaces via the shared ErrorAlert (role=alert), not a bare underline button.
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent(NETWORK_ERROR_MESSAGE());
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(screen.getByRole('button', { name: '再試行' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('renders the hand sort buttons', async () => {
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByTestId('bt-sort-strength')).toBeInTheDocument();
    expect(screen.getByTestId('bt-sort-suit')).toBeInTheDocument();
    expect(screen.getByTestId('bt-sort-number')).toBeInTheDocument();
  });

  it('keeps the selected card index stable across hand sorting', async () => {
    renderWithProviders(<BigTwoPage />);
    // Select ♦7 (original index 2), then re-sort by suit (which moves ♦ to the front).
    fireEvent.click(await screen.findByTestId('hand-card-2'));
    fireEvent.click(screen.getByTestId('bt-sort-suit'));
    // The play command still references the original dealt index.
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [2]));
  });

  it('shows the table play-type label for the cards in play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        tableCards: [card('SPADE', 3), card('SPADE', 5), card('SPADE', 7), card('SPADE', 9), card('SPADE', 11)],
        tablePlayType: 5, // flush
        currentTurn: 1,
      }),
    );
    renderWithProviders(<BigTwoPage />);
    const label = await screen.findByTestId('bt-table-playtype');
    expect(label).toHaveTextContent('フラッシュ');
  });
});
