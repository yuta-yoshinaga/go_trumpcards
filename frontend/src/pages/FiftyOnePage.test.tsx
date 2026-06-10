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
    await waitFor(() => expect(screen.getByText(/スコア: 21/)).toBeInTheDocument());
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

  it('exchange button is disabled until both hand and table cards are selected', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const exchangeBtn = screen.getByTestId('exchange-button');
    expect(exchangeBtn).toBeDisabled();
  });

  it('guides the next selection while the exchange button is disabled', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Nothing selected → prompt for a hand card.
    expect(screen.getByText('手札を選択してください')).toBeInTheDocument();

    // Hand selected → prompt for a table card.
    fireEvent.click(screen.getByRole('button', { name: '♠ A' }));
    expect(screen.getByText('場札を選択してください')).toBeInTheDocument();
    expect(screen.queryByText('手札を選択してください')).not.toBeInTheDocument();

    // Both selected → guide disappears, button enabled.
    fireEvent.click(screen.getByRole('button', { name: '♠ K' }));
    expect(screen.queryByText(/を選択してください/)).not.toBeInTheDocument();
    expect(screen.getByTestId('exchange-button')).not.toBeDisabled();
  });

  it('hides the selection guide when it is not the human turn', async () => {
    mockExec.mockResolvedValue({ ...baseState, currentTurn: 1 });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByText(/を選択してください/)).not.toBeInTheDocument();
  });

  it('renders suit score badges and highlights the leading suit', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument());
    // SPADE=11+10=21, CLOVER=2, HEART=5, DIAMOND=3 → SPADE leads.
    const spade = screen.getByTestId('suit-badge-SPADE');
    expect(spade).toHaveTextContent('21');
    expect(spade.className).toContain('bg-ds-accent');
    const heart = screen.getByTestId('suit-badge-HEART');
    expect(heart).toHaveTextContent('5');
    expect(heart.className).not.toContain('bg-ds-accent');
  });
});
