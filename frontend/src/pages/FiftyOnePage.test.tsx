import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import type { FiftyOneResponse } from '../types/card';

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: () => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  }),
}));

const mockExec = vi.fn();
vi.mock('../api/gameApi', () => ({
  fiftyoneApi: { exec: (...args: unknown[]) => mockExec(...args) },
}));

const baseState: FiftyOneResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 3 },
        { design: 'CLOVER', value: 2 },
      ] as never[],
      score: 21,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], score: 0 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], score: 0 },
  ],
  tableCards: [
    { design: 'SPADE', value: 13 },
    { design: 'HEART', value: 9 },
    { design: 'DIAMOND', value: 12 },
    { design: 'CLOVER', value: 6 },
    { design: 'SPADE', value: 8 },
  ] as never[],
  phase: 0,
  currentTurn: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 1,
  stopCallerIdx: -1,
  lastAction: '',
  lastHandIdx: -1,
  lastTableIdx: -1,
  message: '',
  config: { cpuDifficulty: 1 },
};

const gameEndState: FiftyOneResponse = {
  ...baseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'You win!',
  messageCode: 'fiftyone.result.humanWin',
};

describe('FiftyOnePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('calls reset on mount', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders player score after load', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByText(/21/)).toBeInTheDocument());
  });

  it('exchange all button calls exchangeall', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const exchangeAllBtn = screen.getByTestId('exchange-all-button');
    fireEvent.click(exchangeAllBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchangeall'));
  });

  it('stop button calls stop', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const stopBtn = screen.getByTestId('stop-button');
    fireEvent.click(stopBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stop'));
  });

  it('disables buttons on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('exchange-all-button')).toBeDisabled());
    expect(screen.getByTestId('stop-button')).toBeDisabled();
  });
});
