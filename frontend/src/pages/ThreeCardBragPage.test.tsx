import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { threeCardBragApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeThreeCardBragState } from '../test/stateFactories';
import { ThreeCardBragPage } from './ThreeCardBragPage';

vi.mock('../api/gameApi', () => ({
  threeCardBragApi: { exec: vi.fn() },
  actionLogApi: { threecardbrag: vi.fn() },
}));

const mockExec = vi.mocked(threeCardBragApi.exec);

// Default fixture: a human Betting turn (seat 0), still Blind.
const bettingState = makeThreeCardBragState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true });
// A CPU turn.
const cpuTurnState = makeThreeCardBragState({ phase: 0, currentPlayerIdx: 1, isHumanTurn: false });
const roundEndState = makeThreeCardBragState({ phase: 2, roundWinnerIdx: 0, isHumanTurn: false });
const gameEndState = makeThreeCardBragState({
  phase: 3,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  isHumanTurn: false,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('ThreeCardBragPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ThreeCardBragPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, ante: 1, startingChips: 100 },
      }),
    );
  });

  it('shows the betting action buttons on a human turn', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '手札を見る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット (1)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('dispatches see when the See button is clicked', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    const btn = await screen.findByRole('button', { name: '手札を見る' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('see'));
  });

  it('dispatches bet when the Bet button is clicked', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    const btn = await screen.findByRole('button', { name: 'ベット (1)' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet'));
  });

  it('labels the raise steppers and announces the amount', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    const inc = await screen.findByRole('button', { name: 'レイズ額を増やす' });
    const dec = screen.getByRole('button', { name: 'レイズ額を減らす' });
    const amount = screen.getByTestId('tcb-raise-amount');
    expect(amount).toHaveAttribute('aria-live', 'polite');
    const before = amount.textContent;
    fireEvent.click(inc);
    expect(amount.textContent).not.toBe(before);
    // The decrease button exists and is operable.
    expect(dec).toBeEnabled();
  });

  it('dispatches fold when the Fold button is clicked', async () => {
    renderWithProviders(<ThreeCardBragPage />);
    const btn = await screen.findByRole('button', { name: 'フォールド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('shows the Show button only when canShow is true', async () => {
    mockExec.mockResolvedValue(
      makeThreeCardBragState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true, canShow: true }),
    );
    renderWithProviders(<ThreeCardBragPage />);
    const btn = await screen.findByRole('button', { name: 'ショー' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('hides action buttons on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ThreeCardBragPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '手札を見る' })).not.toBeInTheDocument();
  });

  it('shows the next-deal button at deal end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ThreeCardBragPage />);
    const btn = await screen.findByRole('button', { name: '次のディール' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ThreeCardBragPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });
});
