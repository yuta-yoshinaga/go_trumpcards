import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { teenPattiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTeenPattiState } from '../test/stateFactories';
import { TeenPattiPage } from './TeenPattiPage';

vi.mock('../api/gameApi', () => ({
  teenPattiApi: { exec: vi.fn() },
  actionLogApi: { teenpatti: vi.fn() },
}));

const mockExec = vi.mocked(teenPattiApi.exec);

// Default fixture: a human Betting turn (seat 0), still Blind.
const bettingState = makeTeenPattiState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true });
// A CPU turn.
const cpuTurnState = makeTeenPattiState({ phase: 0, currentPlayerIdx: 1, isHumanTurn: false });
const roundEndState = makeTeenPattiState({ phase: 3, roundWinnerIdx: 0, isHumanTurn: false });
const gameEndState = makeTeenPattiState({
  phase: 4,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  isHumanTurn: false,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('TeenPattiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TeenPattiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, ante: 1, startingChips: 100 },
      }),
    );
  });

  it('shows the betting action buttons on a human turn', async () => {
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '手札を見る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット (1)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('dispatches see when the See button is clicked', async () => {
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: '手札を見る' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('see'));
  });

  it('dispatches bet when the Bet button is clicked', async () => {
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: 'ベット (1)' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet'));
  });

  it('dispatches fold when the Fold button is clicked', async () => {
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: 'フォールド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('shows the Show button only when canShow is true', async () => {
    mockExec.mockResolvedValue(makeTeenPattiState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true, canShow: true }));
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: 'ショー' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('shows the Side Show button only when canRequestSideShow is true and dispatches sideshow', async () => {
    mockExec.mockResolvedValue(
      makeTeenPattiState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true, canRequestSideShow: true }),
    );
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: 'サイドショー' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sideshow'));
  });

  it('shows the unified Side Show panel with Accept/Decline when the human is the target and dispatches respond', async () => {
    mockExec.mockResolvedValue(
      makeTeenPattiState({ phase: 1, isHumanTurn: false, sideShowRequester: 1, sideShowTarget: 0 }),
    );
    renderWithProviders(<TeenPattiPage />);
    const panel = await screen.findByTestId('teenpatti-sideshow-panel');
    // The accept/decline buttons live inside the emphasized panel, not the footer.
    const accept = within(panel).getByRole('button', { name: '承諾' });
    const decline = within(panel).getByRole('button', { name: '拒否' });
    mockExec.mockClear();
    fireEvent.click(accept);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: true }));
    fireEvent.click(decline);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: false }));
  });

  it('shows the Side Show panel (without response buttons) when the human is not the target', async () => {
    mockExec.mockResolvedValue(
      makeTeenPattiState({ phase: 1, isHumanTurn: false, sideShowRequester: 1, sideShowTarget: 2 }),
    );
    renderWithProviders(<TeenPattiPage />);
    const panel = await screen.findByTestId('teenpatti-sideshow-panel');
    expect(within(panel).queryByRole('button', { name: '承諾' })).not.toBeInTheDocument();
    expect(within(panel).queryByRole('button', { name: '拒否' })).not.toBeInTheDocument();
  });

  it('shows the resolved Side Show comparison panel with both hands and the outcome', async () => {
    mockExec.mockResolvedValue(
      makeTeenPattiState({
        phase: 0,
        currentPlayerIdx: 0,
        isHumanTurn: true,
        lastSideShow: {
          requesterIdx: 0,
          targetIdx: 2,
          winnerIdx: 0,
          loserIdx: 2,
          requester: {
            playerIdx: 0,
            handName: 'trail',
            cards: [
              { design: 'SPADE', value: 5 },
              { design: 'HEART', value: 5 },
              { design: 'CLOVER', value: 5 },
            ],
          },
          target: {
            playerIdx: 2,
            handName: 'highcard',
            cards: [
              { design: 'DIAMOND', value: 2 },
              { design: 'CLOVER', value: 7 },
              { design: 'SPADE', value: 9 },
            ],
          },
        },
      }),
    );
    renderWithProviders(<TeenPattiPage />);
    const panel = await screen.findByTestId('teenpatti-sideshow-result');
    // The result title and outcome sentence naming winner/loser and their hand ranks.
    expect(within(panel).getByText('サイドショー結果')).toBeInTheDocument();
    expect(within(panel).getAllByText(/あなた/).length).toBeGreaterThanOrEqual(1);
    expect(within(panel).getAllByText(/トレイル/).length).toBeGreaterThanOrEqual(1);
    // Both participants' cards are rendered (3 + 3).
    expect(within(panel).getAllByRole('img').length).toBeGreaterThanOrEqual(6);
  });

  it('hides the resolved Side Show panel when there is no last side show', async () => {
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('teenpatti-sideshow-result')).not.toBeInTheDocument();
  });

  it('hides the Side Show panel outside the Side Show phase', async () => {
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('teenpatti-sideshow-panel')).not.toBeInTheDocument();
  });

  it('hides action buttons on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '手札を見る' })).not.toBeInTheDocument();
  });

  it('shows the next-deal button at deal end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TeenPattiPage />);
    const btn = await screen.findByRole('button', { name: '次のディール' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TeenPattiPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });
});
