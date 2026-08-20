import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { suecaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSuecaState } from '../test/stateFactories';
import { SuecaPage } from './SuecaPage';

vi.mock('../api/gameApi', () => ({
  suecaApi: { exec: vi.fn() },
  actionLogApi: { sueca: vi.fn() },
}));

const mockExec = vi.mocked(suecaApi.exec);

const playPhaseState = makeSuecaState();
const trickEndState = makeSuecaState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeSuecaState({ phase: 2, roundCardPoints: [70, 50] });
const gameEndState = makeSuecaState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeSuecaState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SuecaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SuecaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('exposes the trump suit by name in the header and sidebar (symbol hidden from SR)', async () => {
    mockExec.mockResolvedValue(playPhaseState); // trumpSuit 4 = ♦ → ダイヤ
    renderWithProviders(<SuecaPage />);
    const sidebar = await screen.findByTestId('sueca-sidebar-trump');
    expect(sidebar).toHaveAttribute('role', 'img');
    expect(sidebar).toHaveAttribute('aria-label', '切り札: ダイヤ');
    // Header + sidebar both expose a named trump element (glyph is aria-hidden).
    expect(screen.getAllByRole('img', { name: '切り札: ダイヤ' }).length).toBeGreaterThanOrEqual(2);
  });

  it('falls back to the raw symbol when the trump suit is unset', async () => {
    // trumpSuit 0 has no suit-name key → the label falls back to the symbol string.
    mockExec.mockResolvedValue(makeSuecaState({ trumpSuit: 0 }));
    renderWithProviders(<SuecaPage />);
    const sidebar = await screen.findByTestId('sueca-sidebar-trump');
    expect(sidebar).toHaveAttribute('role', 'img');
    expect(sidebar).toHaveAttribute('aria-label', '切り札: ');
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SuecaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetGamePoints: 4 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SuecaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SuecaPage />);
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
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the trump suit in the always-visible sidebar', async () => {
    renderWithProviders(<SuecaPage />); // default trump ♦
    await waitFor(() => expect(screen.getByTestId('sueca-sidebar-trump')).toHaveTextContent('切り札: ♦'));
  });

  it('announces the trick-winning team at trick end (leadPlayerIdx → team)', async () => {
    mockExec.mockResolvedValue(trickEndState); // leadPlayerIdx 0 → Team A
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByTestId('sueca-trick-winner')).toHaveTextContent('チームA がトリック獲得'));
  });

  it('names Team B when an odd seat wins the trick', async () => {
    mockExec.mockResolvedValue(makeSuecaState({ phase: 1, leadPlayerIdx: 1 }));
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByTestId('sueca-trick-winner')).toHaveTextContent('チームB がトリック獲得'));
  });

  it('shows the live captured card points against the 61 win line during play', async () => {
    mockExec.mockResolvedValue(makeSuecaState({ roundCardPoints: [45, 20] }));
    renderWithProviders(<SuecaPage />);
    const panel = await screen.findByTestId('sueca-round-points');
    expect(panel).toHaveTextContent('チームA: 45 / 61');
    expect(panel).toHaveTextContent('チームB: 20 / 61');
    expect(panel).toHaveAttribute('role', 'status');
  });

  it('hides the live points panel once the round ends (round result takes over)', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByText('ラウンド結果')).toBeInTheDocument());
    expect(screen.queryByTestId('sueca-round-points')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  // #5642: 配点は A=11 / 7=10 / K=4 / J=3 / Q=2 と直感に反するのに、画面には
  // 61点という目標だけが出ていた。姉妹ゲームの Tressette には既に tr-point-legend
  // がある。
  it('lists what each card is worth in a collapsed legend', async () => {
    renderWithProviders(<SuecaPage />);

    const legend = await screen.findByTestId('sueca-point-legend');
    expect(legend).toHaveTextContent('カード得点');
    // suecaCardPoints (internal/domain/Sueca.go) と同じ値。
    expect(legend).toHaveTextContent('11点');
    expect(legend).toHaveTextContent('10点');
    expect(legend).toHaveTextContent('4点');
    expect(legend).toHaveTextContent('3点');
    expect(legend).toHaveTextContent('2点');
    expect(legend).toHaveTextContent('0点');
  });

  it('keeps the point legend collapsed by default', async () => {
    renderWithProviders(<SuecaPage />);

    const legend = await screen.findByTestId('sueca-point-legend');
    expect(legend).not.toHaveAttribute('open');
  });

  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'sueca.hintRequested',
    });
    renderWithProviders(<SuecaPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
