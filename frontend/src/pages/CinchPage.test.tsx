import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cinchApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCinchState } from '../test/stateFactories';
import { CinchPage } from './CinchPage';

vi.mock('../api/gameApi', () => ({
  cinchApi: { exec: vi.fn() },
  actionLogApi: { cinch: vi.fn() },
}));

const mockExec = vi.mocked(cinchApi.exec);

const playPhaseState = makeCinchState();
const bidPhaseState = makeCinchState({
  phase: 0,
  bidPlayerIdx: 0,
  bidWinnerIdx: -1,
  currentBid: 0,
  isHumanTurn: true,
});
const nameTrumpState = makeCinchState({
  phase: 1,
  bidWinnerIdx: 0,
  trumpSuit: 0,
  isHumanTurn: true,
});
const roundEndState = makeCinchState({
  phase: 4,
  lastDealDetail: {
    trumpSuit: 1,
    bidderIdx: 0,
    bid: 6,
    setBack: false,
    points: { 0: 8, 1: 2, 2: 2, 3: 2 },
    gained: { 0: 6, 1: -1, 2: -1, 3: -1 },
  },
});
const gameEndState = makeCinchState({
  phase: 5,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeCinchState({ currentTurn: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CinchPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CinchPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CinchPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, pointLimit: 21 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<CinchPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('renders the bid phase with pass and numeric bid buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-bid-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    // A raise button for a legal bid (e.g. 1) is present.
    expect(screen.getByRole('button', { name: '1' })).toBeInTheDocument();
  });

  it('passing dispatches bid with bid=0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0 }));
  });

  it('bidding a number dispatches bid with that value', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    const sixBtn = await screen.findByRole('button', { name: '6' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(sixBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 6 }));
  });

  it('renders the name-trump phase with four suit buttons', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-trump-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♠' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦' })).toBeInTheDocument();
  });

  it('naming a trump suit dispatches trump with the suit index', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);
    const spadeBtn = await screen.findByRole('button', { name: '♠' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(nameTrumpState);
    fireEvent.click(spadeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 1 }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<CinchPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders deal end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
