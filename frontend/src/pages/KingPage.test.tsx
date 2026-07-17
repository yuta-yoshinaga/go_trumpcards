import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kingApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKingState } from '../test/stateFactories';
import { KingPage } from './KingPage';

vi.mock('../api/gameApi', () => ({
  kingApi: { exec: vi.fn() },
  actionLogApi: { king: vi.fn() },
}));

const mockExec = vi.mocked(kingApi.exec);

const playPhaseState = makeKingState();
const selectPhaseState = makeKingState({
  phase: 'selectContract',
  dealerIdx: 0,
  currentTurn: 0,
  isHumanTurn: true,
  currentContract: -1,
});
const cpuSelectState = makeKingState({
  phase: 'selectContract',
  dealerIdx: 1,
  currentTurn: 1,
  isHumanTurn: false,
  currentContract: -1,
});
const dealEndState = makeKingState({
  phase: 'dealEnd',
  lastDealDetail: { contract: 0, trumpSuit: -1, dealerIdx: 0, gained: { 0: -20, 1: -30, 2: -10, 3: -30 } },
});
const gameEndState = makeKingState({
  phase: 'gameEnd',
  gameEndFlag: true,
  roundWinners: [0],
  message: 'ゲーム終了！',
});
const cpuTurnState = makeKingState({ currentTurn: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('KingPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KingPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<KingPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Dealer badge', async () => {
    renderWithProviders(<KingPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // Seat 0 is the default dealer.
    expect(screen.getByText('親')).toBeInTheDocument();
  });

  it('renders the select-contract phase with contract buttons and a prompt', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ノートリック' })).toBeInTheDocument();
  });

  it('shows achieve/avoid badges on the contract buttons', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-prompt')).toBeInTheDocument());
    // No Hearts (contract 1) is an avoid contract targeting hearts.
    const avoidBadge = screen.getByTestId('king-contract-badge-1');
    expect(avoidBadge).toHaveTextContent('回避');
    expect(avoidBadge).toHaveTextContent('♥');
    // King (Trump) (contract 6) is the only achieve contract.
    const achieveBadge = screen.getByTestId('king-contract-badge-6');
    expect(achieveBadge).toHaveTextContent('獲得');
  });

  it('choosing a non-trump contract dispatches contract with trumpSuit -1', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    const btn = await screen.findByRole('button', { name: 'ノートリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(selectPhaseState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 0, trumpSuit: -1 }));
  });

  it('choosing King (Trump) shows a trump picker then dispatches the chosen suit', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    const kingBtn = await screen.findByRole('button', { name: 'キング（切り札）' });
    fireEvent.click(kingBtn);
    // The trump prompt and suit buttons appear.
    await waitFor(() => expect(screen.getByTestId('king-trump-prompt')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(selectPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♥' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 6, trumpSuit: 3 }));
  });

  it('shows a CPU-selecting message when the dealer is a CPU', async () => {
    mockExec.mockResolvedValue(cpuSelectState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-cpu')).toBeInTheDocument());
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<KingPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 0 }));
  });

  it('renders deal end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(dealEndState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('king-deal-result')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
