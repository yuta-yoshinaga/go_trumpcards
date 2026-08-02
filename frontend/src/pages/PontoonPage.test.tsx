import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pontoonApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PontoonHand, PontoonResponse } from '../types/card';
import { PontoonPage } from './PontoonPage';

vi.mock('../api/gameApi', () => ({
  pontoonApi: { exec: vi.fn() },
  actionLogApi: { pontoon: vi.fn() },
}));

const mockExec = vi.mocked(pontoonApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<PontoonHand>): PontoonHand {
  return {
    cards: [card('SPADE', 10), card('HEART', 8)],
    bet: 100,
    total: 18,
    rank: 1,
    twisted: false,
    stuck: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PontoonResponse>): PontoonResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hands: [hand()] },
      { name: 'CPU1', isCpu: true, hands: [] },
      { name: 'CPU2', isCpu: true, hands: [hand({ bet: 20 })] },
    ],
    bankerHand: hand({ bet: 0 }),
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 900,
    activeSeat: 0,
    activeHand: 0,
    nextBanker: -1,
    lastResult: '',
    phase: 2,
    canStick: true,
    canTwist: true,
    canBuy: false,
    canSplit: false,
    message: '',
    ...overrides,
  };
}

const bettingState = makeState({ phase: 1, seats: makeState().seats, bankerHand: undefined });

describe('PontoonPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders chips and the banker', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());
    expect(screen.getByText(/親: CPU1/)).toBeInTheDocument();
  });

  // The server withholds a hidden hand's cards entirely, so the page renders
  // backs from `hidden` rather than deciding for itself what may be seen.
  it('renders a hand the server marked hidden as backs', async () => {
    mockExec.mockResolvedValue(
      makeState({
        bankerHand: hand({ hidden: true, cards: [null, null], total: 0, rank: 0 }),
        seats: [
          { name: 'あなた', isCpu: false, hands: [hand()] },
          { name: 'CPU1', isCpu: true, hands: [] },
          {
            name: 'CPU2',
            isCpu: true,
            hands: [hand({ hidden: true, cards: [null, null], total: 0, rank: 0, bet: 20 })],
          },
        ],
      }),
    );
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByLabelText('親の手は伏せられています')).toBeInTheDocument());
    expect(screen.getByLabelText('CPU2 の手は伏せられています')).toBeInTheDocument();
    // A hidden hand must not print a total, which is what a leak would look like.
    expect(screen.queryByText(/合計: 0/)).not.toBeInTheDocument();
  });

  it('reveals every hand once the round settles', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, lastResult: '親は 18' }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.queryByLabelText('親の手は伏せられています')).not.toBeInTheDocument());
    expect(screen.getByLabelText(/親の手 合計18/)).toBeInTheDocument();
  });

  it('offers the bet buttons while betting', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: '100' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  it('disables a stake above the stack', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, chips: 50, bankerHand: undefined }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '500' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '10' })).toBeEnabled();
  });

  // The banker takes the other players' bets rather than making one, so the
  // stake buttons give way to a single deal.
  it('shows a deal button instead of stakes when the human banks', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, isHumanBanker: true, bankerIdx: 0, bankerHand: undefined }));
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: '配る' });
    expect(screen.queryByRole('button', { name: '100' })).not.toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  // The server decides what is legal; the page renders exactly what it is told
  // rather than re-deriving the 15 minimum or the no-buy-after-twist rule.
  it('renders only the declarations the server allows', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スティック' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ツイスト' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'バイ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  it('hides stick when the total is below the minimum', async () => {
    mockExec.mockResolvedValue(makeState({ canStick: false }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ツイスト' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スティック' })).not.toBeInTheDocument();
  });

  it.each([
    ['スティック', 'stick'],
    ['ツイスト', 'twist'],
  ])('%s dispatches %s', async (label, command) => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: label });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // Buy raises by the current stake, so the amount has to come off the hand on
  // turn rather than a constant.
  it('buys for the current stake', async () => {
    mockExec.mockResolvedValue(makeState({ canBuy: true }));
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: 'バイ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('buy', 100));
  });

  it('splits when the server allows it', async () => {
    mockExec.mockResolvedValue(makeState({ canSplit: true }));
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: 'スプリット' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('split'));
  });

  it.each([
    ['引く', 'bankertwist'],
    ['止める', 'bankerstay'],
  ])("the banker's %s dispatches %s", async (label, command) => {
    mockExec.mockResolvedValue(makeState({ phase: 3, isHumanBanker: true, bankerIdx: 0 }));
    renderWithProviders(<PontoonPage />);
    const btn = await screen.findByRole('button', { name: label });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('labels the special ranks after the round', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 4,
        seats: [
          { name: 'あなた', isCpu: false, hands: [hand({ rank: 3, total: 21, payout: 200 })] },
          { name: 'CPU1', isCpu: true, hands: [] },
          { name: 'CPU2', isCpu: true, hands: [hand({ rank: 2, total: 19 })] },
        ],
      }),
    );
    renderWithProviders(<PontoonPage />);
    // "ポンツーン" is both the page title and the rank label, so the count is
    // what distinguishes "the rank rendered" from "only the heading exists".
    await waitFor(() => expect(screen.getAllByText('ポンツーン')).toHaveLength(2));
    expect(screen.getByText('ファイブカード・トリック')).toBeInTheDocument();
    expect(screen.getByText(/\+200/)).toBeInTheDocument();
  });

  it('marks the hand on turn', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByLabelText(/あなた の手 合計18/)).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ツイスト' })).not.toBeInTheDocument());
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('PontoonPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    ['s', 'stick'],
    ['t', 'twist'],
    ['p', 'split'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(makeState({ canBuy: true, canSplit: true }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing b buys for the current stake', async () => {
    mockExec.mockResolvedValue(makeState({ canBuy: true }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('buy', 100));
  });

  it('ignores shortcuts outside the player turn', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4 }));
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['s', 't', 'b', 'p']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PontoonPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('shows no tooltip while betting', async () => {
    localStorage.clear();
    localStorage.setItem('hint_enabled_pontoon', 'true');
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<PontoonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
    localStorage.clear();
  });
});
