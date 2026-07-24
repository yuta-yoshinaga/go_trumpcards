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

  // Builds a human Betting turn at a given stake/seen status; CPUs stay blind.
  const humanBetTurn = (stake: number, seen: boolean, chips = 200) =>
    makeThreeCardBragState({
      phase: 0,
      currentPlayerIdx: 0,
      isHumanTurn: true,
      stake,
      players: [
        {
          id: 0,
          isHuman: true,
          chips,
          seen,
          folded: false,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [
            { design: 'SPADE' as const, value: 13 },
            { design: 'SPADE' as const, value: 12 },
            { design: 'SPADE' as const, value: 11 },
          ],
        },
        {
          id: 1,
          isHuman: false,
          chips: 100,
          seen: false,
          folded: false,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [],
        },
        {
          id: 2,
          isHuman: false,
          chips: 100,
          seen: false,
          folded: false,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [],
        },
        {
          id: 3,
          isHuman: false,
          chips: 100,
          seen: false,
          folded: false,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [],
        },
      ],
    });

  it('shows the nominal bet cost for a Blind player', async () => {
    mockExec.mockResolvedValue(humanBetTurn(20, false));
    renderWithProviders(<ThreeCardBragPage />);
    // Blind: cost equals the stake (20).
    expect(await screen.findByRole('button', { name: 'ベット (20)' })).toBeInTheDocument();
    expect(screen.getByTestId('tcb-cost-notice')).toHaveTextContent('20');
  });

  it('doubles the bet and raise cost on the button labels for a Seen player', async () => {
    mockExec.mockResolvedValue(humanBetTurn(20, true));
    renderWithProviders(<ThreeCardBragPage />);
    // Seen: a call to the stake (20) costs 40 chips.
    expect(await screen.findByRole('button', { name: 'ベット (40)' })).toBeInTheDocument();
    // The raise button starts at the minimum raise (stake + 1 = 21) -> 42 chips seen.
    expect(screen.getByRole('button', { name: 'レイズ (42)' })).toBeInTheDocument();
    // The Seen cost notice warns about the doubled cost.
    expect(screen.getByTestId('tcb-cost-notice')).toHaveTextContent('2倍');
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

  it('shows the raise range (min-max) beside the stepper', async () => {
    // Blind, stake 3, chips 40 -> min 4, max 40.
    mockExec.mockResolvedValue(
      makeThreeCardBragState({
        phase: 0,
        currentPlayerIdx: 0,
        isHumanTurn: true,
        stake: 3,
        players: [
          {
            id: 0,
            isHuman: true,
            chips: 40,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 1,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 2,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 3,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
        ],
      }),
    );
    renderWithProviders(<ThreeCardBragPage />);
    const range = await screen.findByTestId('tcb-raise-range');
    expect(range).toHaveTextContent('範囲: 4〜40');
    // Value syncs to the stake-based minimum.
    expect(screen.getByTestId('tcb-raise-amount')).toHaveTextContent('4');
  });

  it('halves the max for a Seen player and clamps the amount within it', async () => {
    // Seen, stake 2, chips 20 -> min 3, max 10.
    mockExec.mockResolvedValue(
      makeThreeCardBragState({
        phase: 0,
        currentPlayerIdx: 0,
        isHumanTurn: true,
        stake: 2,
        players: [
          {
            id: 0,
            isHuman: true,
            chips: 20,
            seen: true,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 1,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 2,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 3,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
        ],
      }),
    );
    renderWithProviders(<ThreeCardBragPage />);
    const range = await screen.findByTestId('tcb-raise-range');
    expect(range).toHaveTextContent('範囲: 3〜10');
    const inc = screen.getByRole('button', { name: 'レイズ額を増やす' });
    const amount = screen.getByTestId('tcb-raise-amount');
    // Click far past the max; the amount must never exceed 10.
    for (let i = 0; i < 20; i++) fireEvent.click(inc);
    expect(amount).toHaveTextContent('10');
    // The increase button is disabled once the max is reached.
    expect(inc).toBeDisabled();
  });

  it('does not step below the stake-based minimum', async () => {
    mockExec.mockResolvedValue(makeThreeCardBragState({ phase: 0, currentPlayerIdx: 0, isHumanTurn: true, stake: 5 }));
    renderWithProviders(<ThreeCardBragPage />);
    const amount = await screen.findByTestId('tcb-raise-amount');
    expect(amount).toHaveTextContent('6'); // stake + 1
    const dec = screen.getByRole('button', { name: 'レイズ額を減らす' });
    expect(dec).toBeDisabled();
    for (let i = 0; i < 10; i++) fireEvent.click(dec);
    expect(amount).toHaveTextContent('6');
  });

  it('disables the raise button and shows unavailable when chips are too low', async () => {
    // Seen, stake 2, chips 4 -> min 3, max 2 -> no legal raise.
    mockExec.mockResolvedValue(
      makeThreeCardBragState({
        phase: 0,
        currentPlayerIdx: 0,
        isHumanTurn: true,
        stake: 2,
        players: [
          {
            id: 0,
            isHuman: true,
            chips: 4,
            seen: true,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 1,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 2,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
          {
            id: 3,
            isHuman: false,
            chips: 100,
            seen: false,
            folded: false,
            out: false,
            roundBet: 1,
            cardCount: 3,
            cards: [],
          },
        ],
      }),
    );
    renderWithProviders(<ThreeCardBragPage />);
    const range = await screen.findByTestId('tcb-raise-range');
    expect(range).toHaveTextContent('チップ不足のためレイズできません');
    expect(screen.getByRole('button', { name: /レイズ \(/ })).toBeDisabled();
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
