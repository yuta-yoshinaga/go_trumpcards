import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { settemezzoApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SetteEMezzoHand, SetteEMezzoResponse } from '../types/card';
import { SetteEMezzoPage } from './SetteEMezzoPage';

vi.mock('../api/gameApi', () => ({
  settemezzoApi: { exec: vi.fn() },
  actionLogApi: { settemezzo: vi.fn() },
}));

const mockExec = vi.mocked(settemezzoApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<SetteEMezzoHand>): SetteEMezzoHand {
  return {
    cards: [card('SPADE', 4)],
    bet: 100,
    totalHalves: 8,
    totalLabel: '4',
    mattaHalves: 0,
    hasMatta: false,
    stood: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<SetteEMezzoResponse>): SetteEMezzoResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hand: hand() },
      { name: 'CPU1', isCpu: true },
      { name: 'CPU2', isCpu: true, hand: hand({ bet: 20 }) },
    ],
    bankerHand: hand({ bet: 0 }),
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 900,
    activeSeat: 0,
    nextBanker: -1,
    lastResult: '',
    phase: 2,
    targetHalves: 15,
    cpuStandHalves: 11,
    canHit: true,
    canStand: true,
    canSetMatta: false,
    message: '',
    ...overrides,
  };
}

const bettingState = makeState({ phase: 1, bankerHand: undefined });

describe('SetteEMezzoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders chips, the banker and the target', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());
    expect(screen.getByText(/親: CPU1/)).toBeInTheDocument();
    // 7.5 comes from the server in halves so it is not hardcoded twice.
    expect(screen.getByText(/目標: 7\.5/)).toBeInTheDocument();
  });

  // The server withholds a hidden hand's cards; the page renders backs from
  // `hidden` rather than deciding for itself what may be seen.
  it('renders a hand the server marked hidden as backs', async () => {
    mockExec.mockResolvedValue(
      makeState({
        bankerHand: hand({ hidden: true, cards: [null], totalLabel: '' }),
        seats: [
          { name: 'あなた', isCpu: false, hand: hand() },
          { name: 'CPU1', isCpu: true },
          { name: 'CPU2', isCpu: true, hand: hand({ hidden: true, cards: [null], totalLabel: '', bet: 20 }) },
        ],
      }),
    );
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByLabelText('親の手は伏せられています')).toBeInTheDocument());
    expect(screen.getByLabelText('CPU2 の手は伏せられています')).toBeInTheDocument();
  });

  it('reveals a hand the server did not mark hidden', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, lastResult: '親は 4' }));
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByLabelText(/親の手 合計4/)).toBeInTheDocument());
  });

  it('offers the bet buttons while betting', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByRole('button', { name: '100' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  it('disables a stake above the stack', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, chips: 50, bankerHand: undefined }));
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '500' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '10' })).toBeEnabled();
  });

  it('shows a deal button instead of stakes when the human banks', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, isHumanBanker: true, bankerIdx: 0, bankerHand: undefined }));
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByRole('button', { name: '配る' });
    expect(screen.queryByRole('button', { name: '100' })).not.toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it.each([
    ['引く', 'hit'],
    ['止める', 'stand'],
  ])('%s dispatches %s', async (label, command) => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByRole('button', { name: label });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('hides the draw button when the server says it is illegal', async () => {
    mockExec.mockResolvedValue(makeState({ canHit: false }));
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '止める' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  // The matta's value is a choice, and it is offered only when the hand holds
  // one. Each button sends HALVES.
  it('offers the matta choices only when the hand holds one', async () => {
    mockExec.mockResolvedValue(makeState());
    const { unmount } = renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    expect(screen.queryByTestId('matta-6')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeState({ canSetMatta: true }));
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByTestId('matta-6');
    expect(btn).toHaveTextContent('3');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('matta', 6));
  });

  it('offers the half point as its own choice', async () => {
    mockExec.mockResolvedValue(makeState({ canSetMatta: true }));
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByTestId('matta-1');
    expect(btn).toHaveTextContent('0.5');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('matta', 1));
  });

  // The matta stays adjustable until the hand stands, so the board has to show
  // what it currently counts as.
  it('shows the matta at its current value', async () => {
    mockExec.mockResolvedValue(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ hasMatta: true, mattaHalves: 6, totalLabel: '7' }) }],
      }),
    );
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByText(/マッタ = 3/)).toBeInTheDocument());
  });

  it('shows an unassigned matta as half a point', async () => {
    mockExec.mockResolvedValue(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ hasMatta: true, mattaHalves: 0 }) }],
      }),
    );
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByText(/マッタ = 0\.5/)).toBeInTheDocument());
  });

  it.each([
    ['親として引く', 'bankerhit'],
    ['親として止める', 'bankerstand'],
  ])("the banker's %s dispatches %s", async (label, command) => {
    mockExec.mockResolvedValue(makeState({ phase: 3, isHumanBanker: true, bankerIdx: 0 }));
    renderWithProviders(<SetteEMezzoPage />);
    const btn = await screen.findByRole('button', { name: label });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('shows the payout after the round', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 4,
        lastResult: '親は 4',
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ payout: 100, totalLabel: '7.5' }) }],
      }),
    );
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByText(/\+100/)).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument());
  });

  it('marks the chosen matta value as pressed', async () => {
    // Earlier tests queue one-shot resolutions and can leave CLI mode persisted.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(
      makeState({
        canSetMatta: true,
        seats: [
          { name: 'あなた', isCpu: false, hand: hand({ mattaHalves: 4, hasMatta: true }) },
          { name: 'CPU1', isCpu: true },
          { name: 'CPU2', isCpu: true, hand: hand({ bet: 20 }) },
        ],
      }),
    );
    renderWithProviders(<SetteEMezzoPage />);
    const pressed = await screen.findByTestId('matta-4');
    expect(pressed).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('matta-2')).toHaveAttribute('aria-pressed', 'false');
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('SetteEMezzoPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    ['h', 'hit'],
    ['s', 'stand'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('ignores shortcuts outside the player turn', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4 }));
    renderWithProviders(<SetteEMezzoPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 's']) {
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
    renderWithProviders(<SetteEMezzoPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});

// #5566: 相手がいつ引くのをやめるかが分からないまま、賭け続けるか降りるかを
// 決めさせていた。
describe('SetteEMezzoPage stand threshold', () => {
  it('shows the threshold in points, not in halves', async () => {
    mockExec.mockResolvedValue(makeState({}));
    renderWithProviders(<SetteEMezzoPage />);
    const line = await screen.findByTestId('settemezzo-cpu-stand');
    expect(line).toHaveTextContent('5.5');
    // **半点の内部表現をそのまま出さない。**11 と読めては意味が逆になる。
    expect(line).not.toHaveTextContent('11');
  });

  // サーバが別の閾値を返せばそのまま出ること (定数を再実装していない証拠)。
  it('renders whatever the server sends', async () => {
    mockExec.mockResolvedValue(makeState({ cpuStandHalves: 13 }));
    renderWithProviders(<SetteEMezzoPage />);
    const line = await screen.findByTestId('settemezzo-cpu-stand');
    expect(line).toHaveTextContent('6.5');
    expect(line).not.toHaveTextContent('5.5');
  });
});
