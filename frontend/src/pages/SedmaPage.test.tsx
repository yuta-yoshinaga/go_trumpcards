import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sedmaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSedmaState } from '../test/stateFactories';
import { SedmaPage } from './SedmaPage';

vi.mock('../api/gameApi', () => ({
  sedmaApi: { exec: vi.fn() },
  actionLogApi: { sedma: vi.fn() },
}));

const mockExec = vi.mocked(sedmaApi.exec);

const playPhaseState = makeSedmaState();
const trickEndState = makeSedmaState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 7 } },
  ],
});
const roundEndState = makeSedmaState({ phase: 2, roundCardPoints: [30, 20] });
const gameEndState = makeSedmaState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeSedmaState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SedmaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SedmaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SedmaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 101 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SedmaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SedmaPage />);
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
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the live captured card points during play', async () => {
    mockExec.mockResolvedValue(makeSedmaState({ roundCardPoints: [40, 20] }));
    renderWithProviders(<SedmaPage />);
    const panel = await screen.findByTestId('sedma-round-points');
    expect(panel).toHaveTextContent('獲得カード点（現ラウンド）');
    expect(panel).toHaveTextContent('チームAのカード点: 40');
    expect(panel).toHaveTextContent('チームBのカード点: 20');
  });

  it('hides the live captured card points once the round ends', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('sedma-round-points')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('colour-codes the player list by 2-vs-2 team', async () => {
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByTestId('sedma-player-0')).toBeInTheDocument());
    // Even ids → team A (blue), odd ids → team B (red).
    expect(screen.getByTestId('sedma-player-0')).toHaveAttribute('data-team', '0');
    expect(screen.getByTestId('sedma-player-0')).toHaveClass('border-ds-info');
    expect(screen.getByTestId('sedma-player-2')).toHaveAttribute('data-team', '0');
    expect(screen.getByTestId('sedma-player-1')).toHaveAttribute('data-team', '1');
    expect(screen.getByTestId('sedma-player-1')).toHaveClass('border-ds-error');
    expect(screen.getByTestId('sedma-player-3')).toHaveAttribute('data-team', '1');
  });

  it('marks each team with a colour-independent label (badge + sr-only name)', async () => {
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(screen.getByTestId('sedma-player-0')).toBeInTheDocument());
    // Team A (even id): visible 'A' badge + sr-only 'チームA'.
    const teamA = screen.getByTestId('sedma-player-0');
    expect(teamA).toHaveTextContent('A');
    expect(teamA).toHaveTextContent('チームA');
    // Team B (odd id): visible 'B' badge + sr-only 'チームB'.
    const teamB = screen.getByTestId('sedma-player-1');
    expect(teamB).toHaveTextContent('B');
    expect(teamB).toHaveTextContent('チームB');
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<SedmaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
