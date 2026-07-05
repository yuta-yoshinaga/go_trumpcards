import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ultiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeUltiState } from '../test/stateFactories';
import { UltiPage } from './UltiPage';

vi.mock('../api/gameApi', () => ({
  ultiApi: { exec: vi.fn() },
  actionLogApi: { ulti: vi.fn() },
}));

const mockExec = vi.mocked(ultiApi.exec);

const playPhaseState = makeUltiState();
const bidPhaseState = makeUltiState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  trumpSuit: -1,
  talonTaken: false,
  playableIndices: [],
});
const discardPhaseState = makeUltiState({
  phase: 1,
  isHumanTurn: true,
  isHumanBidTurn: false,
  contract: 1,
  trumpSuit: 1,
  playableIndices: [],
});
const trickEndState = makeUltiState({
  phase: 3,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeUltiState({
  phase: 4,
  isHumanTurn: false,
  outcome: 1,
});
const gameEndState = makeUltiState({
  phase: 5,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeUltiState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('UltiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<UltiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the declarer — the badge renders.
    expect(screen.getAllByText('デクレアラー').length).toBeGreaterThan(0);
  });

  it('renders the bid phase with Party, Betli and Durchmarsch buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パルティ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベトリ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ドゥルマルス' })).toBeInTheDocument();
  });

  it('Party is disabled until a trump suit is picked, then dispatches bid with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const partyBtn = await screen.findByRole('button', { name: 'パルティ' });
    expect(partyBtn).toBeDisabled();
    // Pick spades (♠) as trump.
    fireEvent.click(screen.getByRole('button', { name: 'スペード' }));
    expect(screen.getByRole('button', { name: 'パルティ' })).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'パルティ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'party', trumpSuit: 1 }));
  });

  it('declaring Betli dispatches bid with no trump requirement', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const betliBtn = await screen.findByRole('button', { name: 'ベトリ' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(betliBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'betli', trumpSuit: undefined }));
  });

  it('does not show bid controls on a CPU/non-bid turn', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パルティ' })).not.toBeInTheDocument();
  });

  it('renders the discard phase and dispatches discard for the two selected cards', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<UltiPage />);
    const first = await screen.findByAltText('♥ Q');
    fireEvent.click(first);
    fireEvent.click(screen.getByAltText('♥ K'));
    const discardBtn = screen.getByRole('button', { name: '捨てる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1] }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<UltiPage />);
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
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
