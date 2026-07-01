import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calabresellaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCalabresellaState } from '../test/stateFactories';
import { CalabresellaPage } from './CalabresellaPage';

vi.mock('../api/gameApi', () => ({
  calabresellaApi: { exec: vi.fn() },
  actionLogApi: { calabresella: vi.fn() },
}));

const mockExec = vi.mocked(calabresellaApi.exec);

const playPhaseState = makeCalabresellaState();
const bidPhaseState = makeCalabresellaState({
  phase: 0,
  currentBidderIdx: 0,
  isHumanTurn: true,
  winningBid: 0,
});
const discardPhaseState = makeCalabresellaState({ phase: 1, soloistIdx: 0 });
const trickEndState = makeCalabresellaState({
  phase: 3,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeCalabresellaState({
  phase: 4,
  roundThirds: [20, 8, 5],
});
const gameEndState = makeCalabresellaState({
  phase: 5,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeCalabresellaState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CalabresellaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CalabresellaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 21 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Soloist badge', async () => {
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Soloist.
    expect(screen.getByText('ソリスト')).toBeInTheDocument();
  });

  it('renders the bid phase with chiamo, solo and pass buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キアーモ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ソロ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('declaring chiamo dispatches bid with bid=1', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const chiamoBtn = await screen.findByRole('button', { name: 'キアーモ' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(chiamoBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
  });

  it('renders the discard phase with the discard-card button and prompt', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByTestId('calabresella-discard-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'カードを捨てる' })).toBeInTheDocument();
  });

  it('selecting a card then discarding dispatches discard', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const discardBtn = await screen.findByRole('button', { name: 'カードを捨てる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<CalabresellaPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
