import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { schafkopfApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSchafkopfState } from '../test/stateFactories';
import { SchafkopfPage } from './SchafkopfPage';

vi.mock('../api/gameApi', () => ({
  schafkopfApi: { exec: vi.fn() },
  actionLogApi: { schafkopf: vi.fn() },
}));

const mockExec = vi.mocked(schafkopfApi.exec);

const playPhaseState = makeSchafkopfState();
const pickPhaseState = makeSchafkopfState({
  phase: 0,
  pickerIdx: -1,
  currentPlayerIdx: 0,
  beatableContracts: [0, 1, 2],
});
const callPhaseState = makeSchafkopfState({ phase: 1, pickerIdx: 0, currentPlayerIdx: 0, callableSuits: [1, 2] });
const trickEndState = makeSchafkopfState({
  phase: 3,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 13 } },
  ],
});
const roundEndState = makeSchafkopfState({
  phase: 4,
  roundPickerPoints: 70,
  roundMultiplier: 2,
  roundPickerWon: true,
});
const gameEndState = makeSchafkopfState({
  phase: 5,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeSchafkopfState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SchafkopfPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SchafkopfPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, baseChips: 1, startChips: 100, targetChips: 200 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SchafkopfPage />);
    const card = await screen.findByAltText('♠ A');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('play phase: a number key selects a card and Enter plays it', async () => {
    renderWithProviders(<SchafkopfPage />);
    await screen.findByAltText('♠ A');
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    // "1" toggles the first hand card (index 0); Enter confirms the play.
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('keyboard play is disabled on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('renders every contract button in the pick phase on the human turn', async () => {
    mockExec.mockResolvedValue(pickPhaseState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Rufspiel（A呼び）' })).toBeInTheDocument());
    // **All three contracts have to be reachable.** Only Rufspiel used to be,
    // which left Wenz and Solo settable by the domain but not by any player.
    expect(screen.getByRole('button', { name: 'Wenz（Unterのみ切り札）' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♠ スペードを切り札にした Solo を宣言する' })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'パスする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pick', { pick: false }));
  });

  it('declaring Wenz sends the contract, and a Solo carries the suit it names', async () => {
    mockExec.mockResolvedValue(pickPhaseState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Wenz（Unterのみ切り札）' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'Wenz（Unterのみ切り札）' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('pick', { pick: true, contract: 1, soloSuit: undefined }),
    );

    // ♥ Solo, not ♠ — a Solo button that dropped its suit would still pass a
    // test that only checked `contract: 2`.
    fireEvent.click(screen.getByRole('button', { name: '♥ ハートを切り札にした Solo を宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pick', { pick: true, contract: 2, soloSuit: 3 }));
  });

  // **上回れない契約はボタンに出さない。** 出すと、押せるのに必ず拒否される
  // 操作面ができる。
  it('offers only the contracts that outrank the standing bid', async () => {
    mockExec.mockResolvedValue(
      makeSchafkopfState({ phase: 0, pickerIdx: -1, currentPlayerIdx: 0, beatableContracts: [2] }),
    );
    renderWithProviders(<SchafkopfPage />);

    await waitFor(() => expect(screen.getAllByRole('button', { name: /Solo/ })).toHaveLength(4));
    expect(screen.queryByRole('button', { name: 'Rufspiel（A呼び）' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Wenz（Unterのみ切り札）' })).not.toBeInTheDocument();
    // パスは常に残る。
    expect(screen.getByRole('button', { name: 'パスする' })).toBeInTheDocument();
  });

  it('names the contract in play so the trump structure is readable', async () => {
    mockExec.mockResolvedValue(makeSchafkopfState({ contract: 2, soloSuit: 3 }));
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByText('契約: Solo（♥ ハート）')).toBeInTheDocument());
  });

  it('renders callable suit buttons in the call phase', async () => {
    mockExec.mockResolvedValue(callPhaseState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /♠ スペードを呼ぶ/ })).toBeInTheDocument());
    // The aria-label adds partner-context while still containing the visible label (WCAG 2.5.3).
    expect(screen.getByRole('button', { name: /♠ スペードを呼ぶ/ })).toHaveAttribute(
      'aria-label',
      '♠ スペードを呼ぶ（パートナー指定）',
    );
    fireEvent.click(screen.getByRole('button', { name: /♣ クラブを呼ぶ/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', { callSuit: 2 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], suit: 0, pick: false, reason: 'x' } });
    renderWithProviders(<SchafkopfPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], suit: 0, pick: false, reason: 'x' },
      messageCode: 'schafkopf.hintRequested',
    });
    renderWithProviders(<SchafkopfPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
