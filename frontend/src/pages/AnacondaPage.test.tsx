import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { anacondaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeAnacondaState } from '../test/stateFactories';
import { AnacondaPage } from './AnacondaPage';

vi.mock('../api/gameApi', () => ({
  anacondaApi: { exec: vi.fn() },
  actionLogApi: { anaconda: vi.fn() },
}));

const mockExec = vi.mocked(anacondaApi.exec);

const passState = makeAnacondaState({ phase: 0, passCount: 3, isHumanTurn: true });
const setState = makeAnacondaState({ phase: 1, passCount: 0, isHumanTurn: true });
const rollState = makeAnacondaState({ phase: 2, rollIndex: 1, isHumanTurn: true, canRaise: true, currentBet: 10 });
const rollWaitState = makeAnacondaState({ phase: 2, rollIndex: 1, isHumanTurn: false, currentPlayer: 1 });
const resultState = makeAnacondaState({ phase: 3, winnerIdx: 0, result: 1, isHumanTurn: false });
const gameEndState = makeAnacondaState({
  phase: 3,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  isHumanTurn: false,
  message: 'ゲーム終了！ あなたの勝利です！',
});

function cardButtons(): HTMLElement[] {
  return screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(passState);
});

describe('AnacondaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<AnacondaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<AnacondaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  it('shows the Pass button (disabled until the right count is selected) on the pass phase', async () => {
    renderWithProviders(<AnacondaPage />);
    const passBtn = await screen.findByRole('button', { name: 'パスする' });
    expect(passBtn).toBeDisabled();
  });

  it('dispatches pass with the selected indices after choosing passCount cards', async () => {
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'パスする' });
    const cards = cardButtons();
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    fireEvent.click(cards[2]);
    const passBtn = screen.getByRole('button', { name: 'パスする' });
    expect(passBtn).toBeEnabled();
    mockExec.mockClear();
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', [0, 1, 2]));
  });

  it('shows the Keep button on the set phase', async () => {
    mockExec.mockResolvedValue(setState);
    renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キープ' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パスする' })).not.toBeInTheDocument();
  });

  it('dispatches keep with 5 selected indices on the set phase', async () => {
    mockExec.mockResolvedValue(setState);
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'キープ' });
    const cards = cardButtons();
    for (let i = 0; i < 5; i++) fireEvent.click(cards[i]);
    const keepBtn = screen.getByRole('button', { name: 'キープ' });
    expect(keepBtn).toBeEnabled();
    mockExec.mockClear();
    fireEvent.click(keepBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('keep', [0, 1, 2, 3, 4]));
  });

  it('shows the bet buttons on the roll phase and dispatches bets', async () => {
    mockExec.mockResolvedValue(rollState);
    renderWithProviders(<AnacondaPage />);
    const callBtn = await screen.findByRole('button', { name: 'コール / チェック' });
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(callBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 'call'));
  });

  it('disables the Raise button when canRaise is false', async () => {
    mockExec.mockResolvedValue(makeAnacondaState({ phase: 2, isHumanTurn: true, canRaise: false }));
    renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeDisabled());
  });

  it('hides the bet buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(rollWaitState);
    renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByText(/ロールフェーズ/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コール / チェック' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'フォールド' })).not.toBeInTheDocument();
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<AnacondaPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });

  it('shows an under-count hint and keeps Pass disabled before enough cards are selected', async () => {
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'パスする' });
    const feedback = screen.getByTestId('anaconda-selection-feedback');
    expect(feedback).toHaveTextContent('あと 3 枚選択してください（0/3）');
    fireEvent.click(cardButtons()[0]);
    expect(feedback).toHaveTextContent('あと 2 枚選択してください（1/3）');
    expect(screen.getByRole('button', { name: 'パスする' })).toBeDisabled();
  });

  it('shows an exact-count confirmation and enables Pass at the required count', async () => {
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'パスする' });
    const cards = cardButtons();
    for (let i = 0; i < 3; i++) fireEvent.click(cards[i]);
    expect(screen.getByTestId('anaconda-selection-feedback')).toHaveTextContent('選択完了（3/3）');
    expect(screen.getByRole('button', { name: 'パスする' })).toBeEnabled();
  });

  it('shows an over-count hint and keeps Pass disabled when too many cards are selected', async () => {
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'パスする' });
    const cards = cardButtons();
    for (let i = 0; i < 4; i++) fireEvent.click(cards[i]);
    expect(screen.getByTestId('anaconda-selection-feedback')).toHaveTextContent('1 枚外してください（4/3）');
    expect(screen.getByRole('button', { name: 'パスする' })).toBeDisabled();
  });

  it('shows the over/exact feedback against KEEP_SIZE on the set phase', async () => {
    mockExec.mockResolvedValue(setState);
    renderWithProviders(<AnacondaPage />);
    await screen.findByRole('button', { name: 'キープ' });
    const cards = cardButtons();
    for (let i = 0; i < 5; i++) fireEvent.click(cards[i]);
    expect(screen.getByTestId('anaconda-selection-feedback')).toHaveTextContent('選択完了（5/5）');
    expect(screen.getByRole('button', { name: 'キープ' })).toBeEnabled();
    fireEvent.click(cards[5]);
    expect(screen.getByTestId('anaconda-selection-feedback')).toHaveTextContent('1 枚外してください（6/5）');
    expect(screen.getByRole('button', { name: 'キープ' })).toBeDisabled();
  });
});
