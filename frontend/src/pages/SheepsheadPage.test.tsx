import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sheepsheadApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSheepsheadState } from '../test/stateFactories';
import { SheepsheadPage } from './SheepsheadPage';

vi.mock('../api/gameApi', () => ({
  sheepsheadApi: { exec: vi.fn() },
  actionLogApi: { sheepshead: vi.fn() },
}));

const mockExec = vi.mocked(sheepsheadApi.exec);

const playPhaseState = makeSheepsheadState();
const pickPhaseState = makeSheepsheadState({ phase: 0, pickerIdx: -1, currentPlayerIdx: 0, blindCount: 2 });
const buryPhaseState = makeSheepsheadState({ phase: 1, pickerIdx: 0, currentPlayerIdx: 0 });
const callPhaseState = makeSheepsheadState({ phase: 2, pickerIdx: 0, currentPlayerIdx: 0, callableSuits: [1, 2] });
const trickEndState = makeSheepsheadState({
  phase: 4,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 13 } },
  ],
});
const roundEndState = makeSheepsheadState({
  phase: 5,
  roundPickerPoints: 70,
  roundMultiplier: 2,
  roundPickerWon: true,
});
const gameEndState = makeSheepsheadState({
  phase: 6,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeSheepsheadState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SheepsheadPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SheepsheadPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, baseChips: 1, startChips: 100, targetChips: 200 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SheepsheadPage />);
    const card = await screen.findByAltText('♠ A');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('play phase: a number key selects a card and Enter plays it', async () => {
    renderWithProviders(<SheepsheadPage />);
    await screen.findByAltText('♠ A');
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    // "1" toggles the first hand card (index 0); Enter confirms the play.
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('bury phase: two number keys select cards and Enter buries them', async () => {
    mockExec.mockResolvedValue(buryPhaseState);
    renderWithProviders(<SheepsheadPage />);
    await screen.findByAltText('♠ A');
    mockExec.mockClear();
    mockExec.mockResolvedValue(buryPhaseState);
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: '2' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bury', { buryIndices: [0, 1] }));
  });

  it('keyboard play is disabled on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('renders pick / pass buttons in the pick phase on the human turn', async () => {
    mockExec.mockResolvedValue(pickPhaseState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ピックする' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'パスする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pick', { pick: false }));
  });

  it('visualizes the blind as face-down cards in the pick phase', async () => {
    mockExec.mockResolvedValue(pickPhaseState);
    renderWithProviders(<SheepsheadPage />);
    const blind = await screen.findByTestId('sh-blind-display');
    expect(blind).toBeInTheDocument();
    // blindCount is 2, so two face-down card backs are rendered.
    expect(screen.getAllByTestId('animated-card-back')).toHaveLength(2);
  });

  it('does not show the blind display outside the pick phase', async () => {
    mockExec.mockResolvedValue(buryPhaseState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /埋める/ })).toBeInTheDocument());
    expect(screen.queryByTestId('sh-blind-display')).not.toBeInTheDocument();
  });

  it('renders the bury button for the picker in the bury phase', async () => {
    mockExec.mockResolvedValue(buryPhaseState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /埋める/ })).toBeInTheDocument());
  });

  it('renders callable suit buttons in the call phase', async () => {
    mockExec.mockResolvedValue(callPhaseState);
    renderWithProviders(<SheepsheadPage />);
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
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], suit: 0, pick: false, reason: 'x' } });
    renderWithProviders(<SheepsheadPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
