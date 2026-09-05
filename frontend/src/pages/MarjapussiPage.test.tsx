import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { marjapussiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMarjapussiState } from '../test/stateFactories';
import { MarjapussiPage } from './MarjapussiPage';

vi.mock('../api/gameApi', () => ({
  marjapussiApi: { exec: vi.fn() },
  actionLogApi: { marjapussi: vi.fn() },
}));

const mockExec = vi.mocked(marjapussiApi.exec);

const playPhaseState = makeMarjapussiState();
const trickEndState = makeMarjapussiState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeMarjapussiState({
  phase: 2,
  roundCardPoints: [55, 35],
  roundMarriage: [40, 0],
  pussiCount: 4,
  pussiWinnerTeam: 0,
  pussi: [
    { design: 'SPADE', value: 1 }, // A = 11
    { design: 'HEART', value: 10 }, // 10 = 10
    { design: 'DIAMOND', value: 13 }, // K = 4
    { design: 'CLOVER', value: 7 }, // 7 = 0
  ],
});
const gameEndState = makeMarjapussiState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeMarjapussiState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MarjapussiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MarjapussiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 500 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  // --- 必須テスト 1: 切り札が画面に出ていること ---
  it('displays the trump suit on screen (shows 未決定 when 0, and suit symbol when set)', async () => {
    // 1. 未決定のとき (trumpSuit = 0)
    mockExec.mockResolvedValue(makeMarjapussiState({ trumpSuit: 0 }));
    const { unmount } = renderWithProviders(<MarjapussiPage />);
    const trumpElUnset = await screen.findByTestId('marjapussi-trump');
    expect(trumpElUnset).toHaveTextContent('未決定');
    unmount();

    // 2. 切り札が ♠ (1) に設定されたとき
    mockExec.mockResolvedValue(makeMarjapussiState({ trumpSuit: 1 }));
    renderWithProviders(<MarjapussiPage />);
    const trumpElSpade = await screen.findByTestId('marjapussi-trump');
    expect(trumpElSpade).toHaveTextContent('♠');
  });

  // --- 必須テスト 2: チーム分けが画面に出ていること (席 0+2 と 1+3 が別チームとして描かれる) ---
  it('displays partnership teams on screen (seats 0+2 as us, seats 1+3 as them)', async () => {
    renderWithProviders(<MarjapussiPage />);
    // 席 0 (人間) と 席 2 (CPU 2) は味方チーム
    const p0 = await screen.findByTestId('marjapussi-player-team-0');
    expect(p0).toHaveTextContent('味方');
    const p2 = await screen.findByTestId('marjapussi-player-team-2');
    expect(p2).toHaveTextContent('味方');

    // 席 1 (CPU 1) と 席 3 (CPU 3) は相手チーム
    const p1 = await screen.findByTestId('marjapussi-player-team-1');
    expect(p1).toHaveTextContent('相手');
    const p3 = await screen.findByTestId('marjapussi-player-team-3');
    expect(p3).toHaveTextContent('相手');

    // チーム別の進捗バーも表示される
    expect(screen.getByTestId('marjapussi-progress-team-0')).toBeInTheDocument();
    expect(screen.getByTestId('marjapussi-progress-team-1')).toBeInTheDocument();
  });

  it('displays latest marriage declaration or none', async () => {
    // マリッジ宣言なし
    const { unmount: unmountNoMarriage } = renderWithProviders(<MarjapussiPage />);
    const noMarriage = await screen.findByTestId('marjapussi-last-marriage');
    expect(noMarriage).toHaveTextContent('直近のマリッジ宣言: なし');
    unmountNoMarriage();

    // マリッジ宣言あり (チーム0の人間が ♥ を宣言して 40点)
    mockExec.mockResolvedValue(
      makeMarjapussiState({
        trumpSuit: 3,
        roundMarriage: [40, 0],
        leadPlayerIdx: 0,
      }),
    );
    const { unmount } = renderWithProviders(<MarjapussiPage />);
    const declaredMarriage = await screen.findByTestId('marjapussi-last-marriage');
    expect(declaredMarriage).toHaveTextContent('直近のマリッジ宣言');
    expect(declaredMarriage).toHaveTextContent('♥');
    expect(declaredMarriage).toHaveTextContent('40');
    unmount();
  });

  it('shows a marriage available banner when human has K and Q and is leading', async () => {
    // 手札に ♥K と ♥Q があり、リード手番 (currentTrick = [])
    renderWithProviders(<MarjapussiPage />);
    const banner = await screen.findByTestId('marjapussi-marriage');
    expect(banner).toHaveTextContent('マリッジ可能');
    // trumpSuit = 0 (未決定) のため、新しい切り札マリッジとして 20 点
    expect(banner).toHaveTextContent('♥ K-Q (+20)');
  });

  it('labels marriage as 40 points when suit matches existing trump', async () => {
    // すでに切り札が ♥ (3) の場合、♥K+Q マリッジは 40 点
    mockExec.mockResolvedValue(makeMarjapussiState({ trumpSuit: 3 }));
    renderWithProviders(<MarjapussiPage />);
    const banner = await screen.findByTestId('marjapussi-marriage');
    expect(banner).toHaveTextContent('♥ K-Q (+40)');
  });

  it('displays the pussi (berry bag) as 4 hidden cards during play, and reveals winner and cards at round end', async () => {
    // プレイ中: 伏せ札表示
    renderWithProviders(<MarjapussiPage />);
    const pussiEl = await screen.findByTestId('marjapussi-pussi');
    expect(pussiEl).toHaveTextContent('ベリー袋: 4枚（伏せ札）');
    expect(screen.queryByTestId('marjapussi-pussi-result')).not.toBeInTheDocument();

    // ラウンド終了時: 結果表示 (チーム 0 が獲得、A+10+K+7 = 11+10+4+0 = 25点)
    mockExec.mockResolvedValue(roundEndState);
    const { unmount } = renderWithProviders(<MarjapussiPage />);
    const resultEl = await screen.findByTestId('marjapussi-pussi-result');
    expect(resultEl).toHaveTextContent('チーム 0 が獲得');
    expect(resultEl).toHaveTextContent('25');
    unmount();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<MarjapussiPage />);
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
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders a progress bar toward the target for each team', async () => {
    const state = makeMarjapussiState({
      teamScores: [250, 450],
      config: { cpuDifficulty: 1, targetPoints: 500 },
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<MarjapussiPage />);
    const bar0 = await screen.findByTestId('marjapussi-progress-team-0');
    expect(bar0).toHaveAttribute('role', 'progressbar');
    expect(bar0).toHaveAttribute('aria-valuemax', '500');
    expect(bar0).toHaveAttribute('aria-valuenow', '250');
    const bar1 = screen.getByTestId('marjapussi-progress-team-1');
    expect(bar1).toHaveAttribute('aria-valuenow', '450');

    // 250/500 = 50% <= 80%: accent fill
    expect(bar0.querySelector('.bg-ds-accent')).not.toBeNull();
    expect(bar0.querySelector('.bg-ds-warning')).toBeNull();

    // 450/500 = 90% > 80%: warning fill
    expect(bar1.querySelector('.bg-ds-warning')).not.toBeNull();
  });

  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<MarjapussiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'lead_low' },
      messageCode: 'marjapussi.hintRequested',
    });
    renderWithProviders(<MarjapussiPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('announces the prompt from an always-mounted live region', async () => {
    renderWithProviders(<MarjapussiPage />);
    const live = await screen.findByTestId('marjapussi-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toContainElement(await screen.findByTestId('marjapussi-play-prompt'));
  });
});
