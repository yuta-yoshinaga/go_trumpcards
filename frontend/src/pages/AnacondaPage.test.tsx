import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { anacondaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeAnacondaState } from '../test/stateFactories';
import type { AnacondaResponse, Card } from '../types/card';
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
const resultState = makeAnacondaState({ phase: 3, winnerIdx: 0, result: 1, lastPayout: 60, isHumanTurn: false });
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
  // ヒントのトグルは localStorage に残る。消さないと次のテストが
  // チェック済みで始まり、クリックで off になる。
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(passState);
});

describe('AnacondaPage', () => {
  it('shows the configured target rounds', async () => {
    for (const { targetRounds, text } of [
      { targetRounds: 4, text: 'ラウンド 1 / 4' },
      { targetRounds: 12, text: 'ラウンド 1 / 12' },
    ]) {
      mockExec.mockResolvedValueOnce(makeAnacondaState({ config: { ...makeAnacondaState().config, targetRounds } }));
      const { unmount } = renderWithProviders(<AnacondaPage />);
      await waitFor(() => expect(screen.getByText(text)).toBeInTheDocument());
      unmount();
    }
  });

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

  it('announces the raise count through the atomic polite live region', async () => {
    mockExec.mockResolvedValue(rollState);
    renderWithProviders(<AnacondaPage />);

    const live = await screen.findByTestId('anaconda-raise-count-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveAttribute('aria-atomic', 'true');
    expect(live).toHaveTextContent('レイズ 0/3回');
  });

  it('announces the current call amount with the same text shown on screen', async () => {
    mockExec.mockResolvedValueOnce(rollState).mockResolvedValueOnce({ ...rollState, currentBet: 25 });
    const { container } = renderWithProviders(<AnacondaPage />);

    const live = await screen.findByTestId('anaconda-current-bet-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveAttribute('aria-atomic', 'true');
    expect(live).toHaveTextContent('コール額: 10');
    expect(container.querySelector('[data-testid="anaconda-info"]')).toHaveTextContent('コール額: 10');

    fireEvent.click(screen.getByRole('button', { name: 'コール / チェック' }));
    await waitFor(() => {
      expect(live).toHaveTextContent('コール額: 25');
      expect(container.querySelector('[data-testid="anaconda-info"]')).toHaveTextContent('コール額: 25');
    });
  });

  it('shows the additional call amount on the human turn, including when no chips are needed', async () => {
    // The domain owes currentBet - streetBet (this betting round only); roundBet also
    // includes earlier rounds, so it is set higher here to prove it is not used.
    for (const { streetBet, amountText } of [
      { streetBet: 10, amountText: 'コールに必要な追加額: 15' },
      { streetBet: 25, amountText: 'コールに追加支払いは不要です' },
    ]) {
      mockExec.mockResolvedValueOnce(
        makeAnacondaState({
          phase: 2,
          rollIndex: 1,
          isHumanTurn: true,
          currentBet: 25,
          players: makeAnacondaState().players.map((player) =>
            player.isHuman ? { ...player, streetBet, roundBet: streetBet + 40 } : player,
          ),
        }),
      );
      const { unmount } = renderWithProviders(<AnacondaPage />);
      expect(await screen.findByTestId('anaconda-call-needed')).toHaveTextContent(amountText);
      unmount();
    }
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

  it('shows the winner and the paid pot amount, and omits payout when there is no winner', async () => {
    mockExec.mockResolvedValueOnce(resultState);
    const { unmount } = renderWithProviders(<AnacondaPage />);
    expect(await screen.findByText('あなた がポット 60 を獲得しました。')).toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeAnacondaState({ phase: 3, winnerIdx: -1, lastPayout: 0, result: 0 }));
    renderWithProviders(<AnacondaPage />);
    expect(await screen.findByText('勝者なしでラウンドが終了しました。')).toBeInTheDocument();
    expect(screen.queryByText(/ポット .* を獲得しました/)).not.toBeInTheDocument();
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

  const card = (design: Card['design'], value: number): Card => ({ design, value });

  /** A roll-phase state where CPU seat 1 has `revealed` exposed cards. */
  function rollRevealState(revealed: number): AnacondaResponse {
    const cpuCards = [
      card('SPADE', 14),
      card('HEART', 13),
      card('CLOVER', 5),
      card('DIAMOND', 9),
      card('SPADE', 2),
    ].slice(0, revealed);
    const base = makeAnacondaState({
      phase: 2,
      rollIndex: revealed,
      isHumanTurn: true,
      canRaise: true,
      currentBet: 10,
    });
    return {
      ...base,
      players: base.players.map((p) => (p.id === 1 ? { ...p, cards: cpuCards } : p)),
    };
  }

  it('renders exposed cards for revealed slots and face-down placeholders for unrevealed ones by rollIndex', async () => {
    mockExec.mockResolvedValue(rollRevealState(3));
    renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByTestId('anaconda-roll-1')).toBeInTheDocument());
    // 3 revealed slots present, remaining 2 (of KEEP_SIZE=5) are placeholders.
    expect(screen.getByTestId('anaconda-reveal-1-0')).toBeInTheDocument();
    expect(screen.getByTestId('anaconda-reveal-1-2')).toBeInTheDocument();
    expect(screen.queryByTestId('anaconda-reveal-1-3')).not.toBeInTheDocument();
    expect(screen.queryByTestId('anaconda-reveal-1-4')).not.toBeInTheDocument();
    expect(screen.getByTestId('anaconda-roll-1').children).toHaveLength(5);
  });

  it('announces every newly revealed CPU card without mixing it with call amount notices', async () => {
    mockExec.mockResolvedValueOnce(rollRevealState(1)).mockResolvedValueOnce(rollRevealState(3));
    renderWithProviders(<AnacondaPage />);

    const live = await screen.findByTestId('anaconda-revealed-cards-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(live).toHaveTextContent('CPU 1が公開したカード: ♠ 14'));
    expect(screen.getByTestId('anaconda-current-bet-live')).toHaveTextContent('コール額: 10');

    fireEvent.click(screen.getByRole('button', { name: 'コール / チェック' }));
    await waitFor(() => expect(live).toHaveTextContent('CPU 1が公開したカード: ♥ K、♣ 5'));
  });

  it('advances the roll-reveal emphasis one card at a time through the exposed cards', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      mockExec.mockResolvedValue(rollRevealState(3));
      renderWithProviders(<AnacondaPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      await waitFor(() => expect(screen.getByTestId('anaconda-reveal-1-0')).toBeInTheDocument());

      // Step 1: the first exposed card is emphasized.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      expect(screen.getByTestId('anaconda-reveal-1-0')).toHaveAttribute('data-emphasized', 'true');
      expect(screen.getByTestId('anaconda-reveal-1-1')).not.toHaveAttribute('data-emphasized');

      // Step 2: the emphasis advances to the second card.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      expect(screen.getByTestId('anaconda-reveal-1-1')).toHaveAttribute('data-emphasized', 'true');
      expect(screen.getByTestId('anaconda-reveal-1-0')).not.toHaveAttribute('data-emphasized');

      // Step 3: the emphasis lands on the most recently revealed card.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      expect(screen.getByTestId('anaconda-reveal-1-2')).toHaveAttribute('data-emphasized', 'true');
      expect(screen.getByTestId('anaconda-reveal-1-1')).not.toHaveAttribute('data-emphasized');
    } finally {
      vi.useRealTimers();
    }
  });

  // **バックエンドはどの札かまで計算している。**ツールチップは「3枚パス」で
  // 止まっていて、選ぶのはプレイヤー任せだった (#4851)。
  it('rings the cards the hint recommends passing', async () => {
    mockExec.mockResolvedValue(
      makeAnacondaState({ hint: { action: 'pass', cardIndices: [0, 2], reason: 'pass_weakest' } }),
    );
    renderWithProviders(<AnacondaPage />);
    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(toggle);

    await waitFor(() => expect(document.querySelectorAll('[data-hint-card="true"]')).toHaveLength(2));
    // **リングはインラインで置く。**手札には selectedCardStyle の
    // `boxShadow: 'none'` が乗るので、Tailwind の ring-* (同じ box-shadow) は
    // 未選択のあいだ潰される。印だけ見ていても気づけない。
    const ringed = document.querySelector('[data-hint-card="true"]') as HTMLElement;
    expect(ringed.style.outline).toContain('var(--color-ds-warning)');
    expect(ringed).toHaveAttribute('aria-describedby', 'anaconda-hint-target-description');
    expect(screen.getByText('ヒント対象', { selector: '#anaconda-hint-target-description' })).toBeInTheDocument();
    expect(ringed).toHaveAccessibleDescription('ヒント対象');
    expect(ringed).toHaveAttribute('aria-pressed', 'false');
  });

  it('describes suggested keep cards to screen readers', async () => {
    mockExec.mockResolvedValue(makeAnacondaState({ hint: { action: 'keep', cardIndices: [1], reason: 'keep_best' } }));
    renderWithProviders(<AnacondaPage />);
    fireEvent.click(await screen.findByRole('checkbox', { name: 'ヒント表示' }));
    await waitFor(() =>
      expect(document.querySelector('[data-hint-card="true"]')).toHaveAccessibleDescription('ヒント対象'),
    );
  });

  it('rings nothing for a betting suggestion or while hints are off', async () => {
    mockExec.mockResolvedValue(makeAnacondaState({ hint: { action: 'raise', reason: 'strong_hand' } }));
    renderWithProviders(<AnacondaPage />);
    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(toggle);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-hint-card="true"]')).toHaveLength(0);
    expect(cardButtons().every((button) => !button.hasAttribute('aria-describedby'))).toBe(true);
  });

  // #5703: 「左隣」は脱落者を飛ばすので席番号 +1 とは限らない。CUI は受取人を
  // 名指ししていたのに、Web は「左隣へ」としか出しておらず、実際に誰へ渡るのかが
  // 卓の状況次第で分からなかった。
  it('names the seat that will receive the passed cards', async () => {
    mockExec.mockResolvedValue(passState);
    renderWithProviders(<AnacondaPage />);

    const notice = await screen.findByTestId('anaconda-pass-notice');

    expect(notice).toHaveTextContent('CPU 1');
  });

  it('skips eliminated seats when naming the recipient', async () => {
    mockExec.mockResolvedValue(
      makeAnacondaState({
        phase: 0,
        passCount: 3,
        isHumanTurn: true,
        players: passState.players.map((p) => (p.id === 1 ? { ...p, out: true } : p)),
      }),
    );
    renderWithProviders(<AnacondaPage />);

    const notice = await screen.findByTestId('anaconda-pass-notice');

    expect(notice).toHaveTextContent('CPU 2');
    expect(notice).not.toHaveTextContent('CPU 1');
  });
  // **ボタンが押せない理由は 2 つある。**上限到達とチップ不足を区別しないと、
  // 選択肢が黙って消えたようにしか見えない (#6500)。Bouillotte / Primero が
  // 先行していて、Anaconda だけがこの説明を欠いていた。
  it('says why raising is unavailable', async () => {
    const base = makeAnacondaState({ phase: 2, rollIndex: 1, isHumanTurn: true, currentBet: 10 });
    const cases: Array<[string, ReturnType<typeof makeAnacondaState>]> = [
      ['レイズ 1/3回', { ...base, canRaise: true, raiseCount: 1 }],
      ['レイズ上限（3回）に達しました', { ...base, canRaise: false, raiseCount: 3 }],
      [
        'チップが足りません',
        {
          ...base,
          canRaise: false,
          raiseCount: 1,
          players: [{ ...base.players[0], chips: 3, roundBet: 0 }, ...base.players.slice(1)],
        },
      ],
    ];

    // 席に人間がいない盤（観戦状態）でも落ちない ── その場合は 0 として扱う。
    mockExec.mockResolvedValue({
      ...base,
      canRaise: false,
      players: base.players.map((p) => ({ ...p, isHuman: false })),
    });
    const spectator = renderWithProviders(<AnacondaPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    spectator.unmount();

    for (const [text, state] of cases) {
      mockExec.mockResolvedValue(state);
      const { container, unmount } = renderWithProviders(<AnacondaPage />);
      await screen.findByRole('button', { name: 'コール / チェック' });
      // 読み上げ対象そのものを掴む ── 同じ文字列を持つ別の要素で通らないように。
      const live = container.querySelector('[data-testid="anaconda-raise-count"][aria-live="polite"]');
      expect(live).not.toBeNull();
      expect(live).toHaveTextContent(text);
      expect(live?.textContent).not.toContain('{{');
      unmount();
    }
  });
});
