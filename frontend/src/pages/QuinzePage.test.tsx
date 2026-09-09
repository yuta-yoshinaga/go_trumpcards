import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { quinzeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, QuinzeHand, QuinzeResponse } from '../types/card';
import { QuinzePage } from './QuinzePage';

vi.mock('../api/gameApi', () => ({
  quinzeApi: { exec: vi.fn() },
  actionLogApi: { quinze: vi.fn() },
}));

const mockExec = vi.mocked(quinzeApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<QuinzeHand>): QuinzeHand {
  return {
    cards: [card('SPADE', 4)],
    bet: 100,
    totalPoints: 8,
    totalLabel: '4',

    stood: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<QuinzeResponse>): QuinzeResponse {
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
    targetPoints: 15,
    cpuStandPoints: 11,
    canHit: true,
    canStand: true,

    message: '',
    ...overrides,
  };
}

const bettingState = makeState({ phase: 1, bankerHand: undefined });

describe('QuinzePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders chips, the banker and the target', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());
    expect(screen.getByText(/親: CPU1/)).toBeInTheDocument();
    // 15 comes from the server in halves so it is not hardcoded twice.
    expect(screen.getByText(/目標: 15/)).toBeInTheDocument();
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
    renderWithProviders(<QuinzePage />);
    // 「伏せられている」だけでなく**何枚あるか**まで言う。晴眼者は CardBack を
    // 数えているので、そこが落ちると同じ情報が得られない (#6362)。
    await waitFor(() => expect(screen.getByLabelText(/^親の手は伏せられています \d+枚$/)).toBeInTheDocument());
    expect(screen.getByLabelText('CPU2 の手は伏せられています 1枚')).toBeInTheDocument();
    // 未解決のプレースホルダが出ていないこと。
    expect(document.body.textContent).not.toContain('{{count}}');
    expect(
      Array.from(document.querySelectorAll('[aria-label]'), (el) => el.getAttribute('aria-label') ?? '').join(' | '),
    ).not.toContain('{{');
  });

  it('reveals a hand the server did not mark hidden', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, lastResult: '親は 4' }));
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByLabelText(/親の手 合計4/)).toBeInTheDocument());
  });

  it('offers the bet buttons while betting', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<QuinzePage />);
    const btn = await screen.findByRole('button', { name: '100' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  it('disables a stake above the stack', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, chips: 50, bankerHand: undefined }));
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '500' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '10' })).toBeEnabled();
  });

  it('shows a deal button instead of stakes when the human banks', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, isHumanBanker: true, bankerIdx: 0, bankerHand: undefined }));
    renderWithProviders(<QuinzePage />);
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
    renderWithProviders(<QuinzePage />);
    const btn = await screen.findByRole('button', { name: label });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('hides the draw button when the server says it is illegal', async () => {
    mockExec.mockResolvedValue(makeState({ canHit: false }));
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '止める' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });
  it.each([
    ['親として引く', 'bankerhit'],
    ['親として止める', 'bankerstand'],
  ])("the banker's %s dispatches %s", async (label, command) => {
    mockExec.mockResolvedValue(makeState({ phase: 3, isHumanBanker: true, bankerIdx: 0 }));
    renderWithProviders(<QuinzePage />);
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
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ payout: 100, totalLabel: '15' }) }],
      }),
    );
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByText(/\+100/)).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<QuinzePage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument());
  });

  // Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
  // ActionShortcutsPanel; assert the keys actually run their action (#4429).
  describe('QuinzePage keyboard shortcuts', () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it.each([
      ['h', 'hit'],
      ['s', 'stand'],
    ])('pressing %s dispatches %s', async (key, command) => {
      mockExec.mockResolvedValue(makeState());
      renderWithProviders(<QuinzePage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
      mockExec.mockClear();
      fireEvent.keyDown(document, { key });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
    });

    it('ignores shortcuts outside the player turn', async () => {
      mockExec.mockResolvedValue(makeState({ phase: 4 }));
      renderWithProviders(<QuinzePage />);
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
      renderWithProviders(<QuinzePage />);

      const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
      expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

      fireEvent.click(toggle);
      expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
    });
  });

  // #5566: 相手がいつ引くのをやめるかが分からないまま、賭け続けるか降りるかを
  // 決めさせていた。
  describe('QuinzePage stand threshold', () => {
    it('shows the threshold in points, not in halves', async () => {
      mockExec.mockResolvedValue(makeState({}));
      renderWithProviders(<QuinzePage />);
      const line = await screen.findByTestId('quinze-cpu-stand');
      expect(line).toHaveTextContent('11');
    });

    // サーバが別の閾値を返せばそのまま出ること (定数を再実装していない証拠)。
    it('renders whatever the server sends', async () => {
      mockExec.mockResolvedValue(makeState({ cpuStandPoints: 13 }));
      renderWithProviders(<QuinzePage />);
      const line = await screen.findByTestId('quinze-cpu-stand');
      expect(line).toHaveTextContent('13');
      expect(line).not.toHaveTextContent('11');
    });
  });
});
