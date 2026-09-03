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
// raiserTeam === human team (0) — the human's own team raised
const respondPhaseStateOwnRaiser = makeWattenState({
  phase: 2,
  currentPlayerIdx: 1,
  responderIdx: 0,
  raiserTeam: 0,
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

  // #5701: スタークの吊り上げ (レイズ / hold / フォールドの応酬) がこのゲームの
  // 核心なのに、履歴パネルは live region ではなく、スクリーンリーダー利用者には
  // 賭け金が動いたことが一切伝わらなかった。
  it('announces the stake history as a polite live region', async () => {
    mockActionLog.mockResolvedValue({
      entries: [logEntry(1, 'declare', 0), logEntry(2, 'raise', 1)],
    });
    renderWithProviders(<WattenPage />);

    const panel = await screen.findByTestId('watten-stake-history');

    expect(panel).toHaveAttribute('role', 'status');
    expect(panel).toHaveAttribute('aria-live', 'polite');
    // 読み上げ対象の中に、誰がいくらへ上げたかが入っていること。
    expect(panel).toHaveTextContent('CPU 1 レイズ→3');
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

  // **吊り上げたのが相手か味方かで hold/fold の判断は変わる。**サーバは
  // raiserTeam を毎回送っているのに、画面は pendingStake しか出していなかった (#6497)。
  it('shows opponent-raised label when the opponent team raised', async () => {
    mockExec.mockResolvedValue(respondPhaseState); // raiserTeam=1, human team=0
    renderWithProviders(<WattenPage />);
    const info = await screen.findByTestId('watten-raiser-info');
    expect(info).toHaveTextContent('相手チームにレイズされました。');
    expect(info).not.toHaveTextContent('{{');
  });

  it('shows own-raised label when the human team raised', async () => {
    mockExec.mockResolvedValue(respondPhaseStateOwnRaiser); // raiserTeam=0, human team=0
    renderWithProviders(<WattenPage />);
    const info = await screen.findByTestId('watten-raiser-info');
    expect(info).toHaveTextContent('自チームがレイズしました。');
    expect(info).not.toHaveTextContent('{{');
  });

  // 負のコントロール: まだ誰も吊り上げていない (-1) 局面では出さない。
  it('does not render the raiser-info element when raiserTeam is -1', async () => {
    mockExec.mockResolvedValue(
      makeWattenState({ phase: 2, responderIdx: 0, raiserTeam: -1, pendingStake: 3, canRaise: false }),
    );
    renderWithProviders(<WattenPage />);
    await screen.findByRole('button', { name: /hold/ });
    expect(screen.queryByTestId('watten-raiser-info')).not.toBeInTheDocument();
  });

  // 宣言の催促もフェーズが変わったときに現れるテキスト。**名前が
  // `-bid-prompt` でないだけ**で領域の外に取り残されていた (#6880 レビュー指摘)。
  it('announces the watten-declare-prompt from the always-mounted live region', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<WattenPage />);

    const live = await screen.findByTestId('watten-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 隣に置いただけの実装は属性の検査を通る。**中にあること**を見る。
    expect(live).toContainElement(await screen.findByTestId('watten-declare-prompt'));
  });
});
