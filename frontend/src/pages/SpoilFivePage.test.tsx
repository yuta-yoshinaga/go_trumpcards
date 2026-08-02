import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spoilFiveApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSpoilFiveState } from '../test/stateFactories';
import { SpoilFivePage } from './SpoilFivePage';

vi.mock('../api/gameApi', () => ({
  spoilFiveApi: { exec: vi.fn() },
  actionLogApi: { spoilfive: vi.fn() },
}));

const mockExec = vi.mocked(spoilFiveApi.exec);

const playPhaseState = makeSpoilFiveState();
const trickEndState = makeSpoilFiveState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 7 } },
  ],
});
const roundEndState = makeSpoilFiveState({ phase: 2, roundWinnerIdx: 0 });
const spoilState = makeSpoilFiveState({ phase: 2, roundWinnerIdx: -1, pot: 10 });
const gameEndState = makeSpoilFiveState({
  phase: 3,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeSpoilFiveState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SpoilFivePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpoilFivePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SpoilFivePage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows a transient +NN pot delta when the pot grows', async () => {
    renderWithProviders(<SpoilFivePage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockResolvedValue(spoilState); // pot 5 -> 10
    fireEvent.click(playBtn);
    const delta = await screen.findByTestId('spoilfive-pot-delta');
    expect(delta).toHaveTextContent('+5');
    // The visible pulse is decorative (colour-only, transient) → hidden from SR.
    expect(delta).toHaveAttribute('aria-hidden', 'true');
  });

  it('announces the pot change to screen readers with delta and total', async () => {
    renderWithProviders(<SpoilFivePage />);
    const card = await screen.findByAltText('♥ Q');
    const announce = screen.getByTestId('spoilfive-pot-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    // Empty until the pot grows.
    expect(announce).toHaveTextContent('');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockResolvedValue(spoilState); // pot 5 -> 10
    fireEvent.click(playBtn);
    await waitFor(() => expect(screen.getByTestId('spoilfive-pot-announce')).toHaveTextContent('ポット +5（合計 10）'));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the Spoil (流局) message on a spoiled round', async () => {
    mockExec.mockResolvedValue(spoilState);
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByText(/流局/)).toBeInTheDocument());
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders five player panels with pot and trump', async () => {
    renderWithProviders(<SpoilFivePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.getByText(/ポット/)).toBeInTheDocument();
    expect(screen.getByText('切り札 ♥')).toBeInTheDocument();
  });

  it('renders the top-trump order legend without a duplicate trump ace when Hearts is trump', async () => {
    // Default state has trumpSuit 3 (Hearts): the ♥A is the trump ace, so it must not be duplicated.
    renderWithProviders(<SpoilFivePage />);
    const legend = await screen.findByTestId('spoilfive-trump-legend');
    expect(legend).toHaveTextContent('トップトランプ順序');
    expect(legend).toHaveTextContent('5♥');
    expect(legend).toHaveTextContent('J♥');
    expect(legend).toHaveTextContent('A♥');
    expect(legend).toHaveTextContent('その他の切り札');
    expect(legend).toHaveTextContent('非切り札');
    // Hearts trump: A♥ appears exactly once (no separate trump-ace entry).
    expect(legend.textContent?.match(/A♥/g)?.length).toBe(1);
  });

  it('lists the trump ace separately when a black suit is trump', async () => {
    mockExec.mockResolvedValue(makeSpoilFiveState({ trumpSuit: 1 })); // ♠
    renderWithProviders(<SpoilFivePage />);
    const legend = await screen.findByTestId('spoilfive-trump-legend');
    expect(legend).toHaveTextContent('5♠');
    expect(legend).toHaveTextContent('A♠');
    expect(legend).toHaveTextContent('A♥');
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<SpoilFivePage />);
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
      messageCode: 'spoilFive.hintRequested',
    });
    renderWithProviders(<SpoilFivePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
