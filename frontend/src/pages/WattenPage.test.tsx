import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, wattenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeWattenState } from '../test/stateFactories';
import type { ActionLogEntry } from '../types/card';
import { WattenPage } from './WattenPage';

vi.mock('../api/gameApi', () => ({
  wattenApi: { exec: vi.fn() },
  actionLogApi: { watten: vi.fn() },
}));

const mockExec = vi.mocked(wattenApi.exec);
const mockActionLog = vi.mocked(actionLogApi.watten);

/** Builds a minimal action-log entry for tests. */
function logEntry(turnNumber: number, actionType: string, playerIdx: number): ActionLogEntry {
  return { turnNumber, playerIdx, actionType, detail: '' };
}

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
  mockActionLog.mockReset();
  mockActionLog.mockResolvedValue({ entries: [] });
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

  it('renders the stake-escalation mini-history in order when the action log has raises', async () => {
    mockActionLog.mockResolvedValue({
      entries: [
        logEntry(1, 'declare', 0),
        logEntry(2, 'raise', 1),
        logEntry(3, 'hold', 0),
        logEntry(4, 'raise', 0),
        logEntry(5, 'hold', 1),
      ],
    });
    renderWithProviders(<WattenPage />);
    const panel = await screen.findByTestId('watten-stake-history');
    expect(panel).toHaveTextContent('開始 2');
    expect(panel).toHaveTextContent('CPU 1 レイズ→3');
    expect(panel).toHaveTextContent('あなた hold（3）');
    expect(panel).toHaveTextContent('あなた レイズ→4');
    expect(panel).toHaveTextContent('CPU 1 hold（4）');
    // Ordering: the first CPU raise precedes the human's later raise.
    expect(panel.textContent?.indexOf('CPU 1 レイズ→3')).toBeLessThan(
      panel.textContent?.indexOf('あなた レイズ→4') ?? -1,
    );
  });

  it('does not render the mini-history when the action log has no escalation', async () => {
    mockActionLog.mockResolvedValue({ entries: [logEntry(1, 'declare', 0), logEntry(2, 'play', 0)] });
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(screen.getByAltText('♥ K')).toBeInTheDocument());
    expect(screen.queryByTestId('watten-stake-history')).not.toBeInTheDocument();
  });

  it('shows the fold-loss amount during the respond phase', async () => {
    mockExec.mockResolvedValue(respondPhaseState);
    renderWithProviders(<WattenPage />);
    const loss = await screen.findByTestId('watten-respond-loss');
    // respondPhaseState keeps the default confirmed stake of 2.
    expect(loss).toHaveTextContent('fold すると相手チームに 2点を献上します。');
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { action: 'play', cardIndex: 0, reason: 'x' } });
    renderWithProviders(<WattenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { action: 'play', cardIndex: 0, reason: 'x' },
      messageCode: 'watten.hintRequested',
    });
    renderWithProviders(<WattenPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
