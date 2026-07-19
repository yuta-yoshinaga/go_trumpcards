import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { wattenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeWattenState } from '../test/stateFactories';
import { WattenPage } from './WattenPage';

vi.mock('../api/gameApi', () => ({
  wattenApi: { exec: vi.fn() },
  actionLogApi: { watten: vi.fn() },
}));

const mockExec = vi.mocked(wattenApi.exec);

const playPhaseState = makeWattenState();
const declarePhaseState = makeWattenState({
  phase: 0,
  dealerIdx: 0,
  currentPlayerIdx: -1,
  schlagRank: 0,
  criticalSuit: -1,
  canRaise: false,
});
const raisePhaseState = makeWattenState({ phase: 1, currentPlayerIdx: 0, canRaise: true });
const respondPhaseState = makeWattenState({
  phase: 2,
  currentPlayerIdx: 1,
  responderIdx: 0,
  raiserTeam: 1,
  pendingStake: 3,
  canRaise: false,
});
const roundEndState = makeWattenState({ phase: 4, dealWinnerTeam: 0, canRaise: false });
const gameEndState = makeWattenState({
  phase: 5,
  gameEndFlag: true,
  winnerTeam: 0,
  canRaise: false,
  message: 'ゲーム終了！ チーム0の勝ち！',
});
const cpuTurnState = makeWattenState({ phase: 1, currentPlayerIdx: 1, canRaise: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('WattenPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<WattenPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<WattenPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 15,
        maxRaises: 4,
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<WattenPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ K')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('renders the declare phase with the Schlag rank, suit and declare buttons', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'A' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ハート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument();
  });

  it('declare is disabled until a rank and suit are picked, then dispatches declare', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<WattenPage />);
    const declareBtn = await screen.findByRole('button', { name: '宣言' });
    expect(declareBtn).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: 'K' }));
    expect(screen.getByRole('button', { name: '宣言' })).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    expect(screen.getByRole('button', { name: '宣言' })).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(declarePhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 13, 3));
  });

  it('previews the top-trump panel with the permanent trumps in the declare phase', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<WattenPage />);
    // Panel + rule explanation render.
    await waitFor(() => expect(screen.getByTestId('watten-trump-panel')).toBeInTheDocument());
    expect(screen.getByText('強札プレビュー')).toBeInTheDocument();
    // Base hand (♥K, ♦7, ♠A): Max + Spitz are permanent trumps → count 2 before any pick.
    expect(screen.getByTestId('watten-trump-count')).toHaveTextContent('あなたの強札: 2枚');
    // The permanent trumps are ringed in the hand.
    expect(screen.getByRole('button', { name: '♥ K' })).toHaveAttribute('data-trump', 'true');
  });

  it('updates the top-trump preview and hand ring when a Schlag rank is picked', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByTestId('watten-trump-count')).toHaveTextContent('2枚'));
    // ♠A is plain until Schlag = A is chosen.
    expect(screen.getByRole('button', { name: '♠ A' })).not.toHaveAttribute('data-trump');
    fireEvent.click(screen.getByRole('button', { name: 'A' }));
    // ♠A now becomes a Schlag trump → count rises to 3 and the card is ringed.
    expect(screen.getByTestId('watten-trump-count')).toHaveTextContent('あなたの強札: 3枚');
    expect(screen.getByRole('button', { name: '♠ A' })).toHaveAttribute('data-trump', 'true');
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<WattenPage />);
    const card = await screen.findByAltText('♥ K');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 0));
  });

  it('shows the raise button when canRaise and dispatches raise', async () => {
    mockExec.mockResolvedValue(raisePhaseState);
    renderWithProviders(<WattenPage />);
    const raiseBtn = await screen.findByRole('button', { name: /レイズ/ });
    mockExec.mockClear();
    mockExec.mockResolvedValue(raisePhaseState);
    fireEvent.click(raiseBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise'));
  });

  it('renders the respond phase with hold and fold and dispatches respond', async () => {
    mockExec.mockResolvedValue(respondPhaseState);
    renderWithProviders(<WattenPage />);
    const holdBtn = await screen.findByRole('button', { name: /hold/ });
    const foldBtn = screen.getByRole('button', { name: /fold/ });
    mockExec.mockClear();
    mockExec.mockResolvedValue(respondPhaseState);
    fireEvent.click(holdBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', undefined, undefined, undefined, true));
    fireEvent.click(foldBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', undefined, undefined, undefined, false));
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ チーム0の勝ち！')).toBeInTheDocument());
  });

  it('does not show the play or raise buttons on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByAltText('♥ K')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /レイズ/ })).not.toBeInTheDocument();
  });
});
