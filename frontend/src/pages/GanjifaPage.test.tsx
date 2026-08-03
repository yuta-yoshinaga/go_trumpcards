import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ganjifaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGanjifaState } from '../test/stateFactories';
import { GanjifaPage } from './GanjifaPage';

vi.mock('../api/gameApi', () => ({
  ganjifaApi: { exec: vi.fn() },
  actionLogApi: { ganjifa: vi.fn() },
}));

const mockExec = vi.mocked(ganjifaApi.exec);

const playPhaseState = makeGanjifaState();
const cpuTurnState = makeGanjifaState({ isHumanTurn: false, currentPlayerIdx: 1, playableIndices: [] });
const trickEndState = makeGanjifaState({ phase: 1, isHumanTurn: false, playableIndices: [] });
const roundEndState = makeGanjifaState({ phase: 2, isHumanTurn: false, playableIndices: [], roundTricks: [14, 10, 8] });
const gameEndState = makeGanjifaState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('GanjifaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GanjifaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<GanjifaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 3 },
      }),
    );
  });

  // The rank direction is the only thing a player cannot infer from the cards,
  // so the readout has to change with the trump group, not just exist.
  it('states that higher numbers win when trump is a strong suit', async () => {
    mockExec.mockResolvedValue(makeGanjifaState({ trumpSuit: 3 }));
    renderWithProviders(<GanjifaPage />);
    const readout = await screen.findByTestId('ganjifa-trump-group');
    expect(readout).toHaveTextContent('数字が大きいほど強い');
    expect(readout).toHaveClass('text-ds-info');
  });

  it('states that lower numbers win when trump is a weak suit', async () => {
    mockExec.mockResolvedValue(makeGanjifaState({ trumpSuit: 6 }));
    renderWithProviders(<GanjifaPage />);
    const readout = await screen.findByTestId('ganjifa-trump-group');
    expect(readout).toHaveTextContent('数字が小さいほど強い');
    expect(readout).toHaveClass('text-ds-warning');
  });

  it('names the trump suit with its glyph', async () => {
    mockExec.mockResolvedValue(makeGanjifaState({ trumpSuit: 2 }));
    renderWithProviders(<GanjifaPage />);
    expect(await screen.findByText(/Shamsher/)).toBeInTheDocument();
  });

  it('plays the selected card', async () => {
    renderWithProviders(<GanjifaPage />);
    const playButton = await screen.findByRole('button', { name: '出す' });
    // The play button stays inert until exactly one card is selected.
    expect(playButton).toBeDisabled();

    // Procedural cards get their accessible name from label + glyph ("12 \u265b").
    fireEvent.click(screen.getByRole('button', { name: '12 \u265b' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('hides the play control when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GanjifaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<GanjifaPage />);
    const button = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(trickEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round and shows the trick tally', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GanjifaPage />);
    const button = await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GanjifaPage />);
    expect(await screen.findByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument();
  });

  it('has no bid controls — Ganjifa has no bidding phase', async () => {
    renderWithProviders(<GanjifaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'lead_high' } });
    renderWithProviders(<GanjifaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'lead_high' },
      messageCode: 'ganjifa.hintRequested',
    });
    renderWithProviders(<GanjifaPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
