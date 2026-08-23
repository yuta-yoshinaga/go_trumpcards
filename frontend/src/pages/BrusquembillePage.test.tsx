import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { brusquembilleApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BrusquembilleResponse, Card } from '../types/card';
import { BrusquembillePhase } from '../types/phases';
import { BrusquembillePage } from './BrusquembillePage';

vi.mock('../api/gameApi', () => ({
  brusquembilleApi: { exec: vi.fn() },
  actionLogApi: { brusquembille: vi.fn() },
}));

const mockExec = vi.mocked(brusquembilleApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BrusquembilleResponse> = {}): BrusquembilleResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 5), card('DIAMOND', 11)],
        points: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], points: 0, trickCount: 0 },
    ],
    phase: 0,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    trumpCard: card('SPADE', 13),
    dealerIdx: 0,
    leadPlayerIdx: 0,
    stockRemaining: 33,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, playerCnt: 2 },
    validIndices: [0, 1, 2],
    followRequired: false,
    ...overrides,
  };
}

beforeEach(() => {
  // ヒントのオン/オフは localStorage に残る。消さないと前のテストの状態を
  // 引き継いで、以降のアサーションが何も確かめなくなる。
  // (clear() だと初回訪問フラグまで消え、チュートリアル案内ダイアログが出る。)
  localStorage.removeItem('hint_enabled_brusquembille');
  mockExec.mockResolvedValue(makeState());
});

describe('BrusquembillePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 2 }));
  });

  it('renders header info (trick, stock, points)', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/トリック: 1/)).toBeInTheDocument());
    expect(screen.getByText(/山札: 33/)).toBeInTheDocument();
    expect(screen.getByText(/得点/)).toBeInTheDocument();
  });

  it('exposes the tutorial target elements for the guided tour', async () => {
    // A card on the table so the trick area (conditionally rendered) is present.
    mockExec.mockResolvedValue(makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 4) }] }));
    const { container } = renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/トリック: 1/)).toBeInTheDocument());
    for (const target of ['brusquembille-trump', 'brusquembille-trick', 'brusquembille-hand', 'brusquembille-score']) {
      expect(container.querySelector(`[data-tutorial="${target}"]`)).not.toBeNull();
    }
  });

  it('shows human hand as 3 play buttons', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ J を出す' })).toBeInTheDocument();
  });

  it('renders hand and trump cards with the AnimatedCard component', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    // 3 hand cards + 1 face-up trump card are rendered as animated cards.
    expect(screen.getAllByTestId('animated-card')).toHaveLength(4);
  });

  it('fires play with the selected card index when a card is clicked', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♥ 5 を出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('keyboard: a number key highlights a card and Enter plays it', async () => {
    renderWithProviders(<BrusquembillePage />);
    const secondCard = await screen.findByRole('button', { name: '♥ 5 を出す' });
    mockExec.mockClear();
    // "2" highlights the second hand card (index 1) without playing it.
    fireEvent.keyDown(document.body, { key: '2' });
    await waitFor(() => expect(secondCard).toHaveAttribute('aria-pressed', 'true'));
    expect(mockExec).not.toHaveBeenCalled();
    // Enter plays the highlighted card.
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('keyboard: Escape clears the highlight and Enter then plays nothing', async () => {
    renderWithProviders(<BrusquembillePage />);
    const firstCard = await screen.findByRole('button', { name: '♠ A を出す' });
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(firstCard).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.keyDown(document.body, { key: 'Escape' });
    await waitFor(() => expect(firstCard).toHaveAttribute('aria-pressed', 'false'));
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('keyboard: number keys do nothing on a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<BrusquembillePage />);
    const firstCard = await screen.findByRole('button', { name: '♠ A を出す' });
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(firstCard).toHaveAttribute('aria-pressed', 'false');
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('shows "Next trick" button on trick-end and dispatches next', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリックへ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のトリックへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows youWin banner when winnerIdx is 0', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 70, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 50, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝ち！.*70.*50/)).toBeInTheDocument());
  });

  it('shows cpuWin banner when winnerIdx is 1', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 50, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 70, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/CPUの勝ち.*70|CPUの勝ち.*50.*70/)).toBeInTheDocument());
  });

  it('shows tie banner when winnerIdx is -1 and game ended', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: -1,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 60, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 60, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/引き分け/)).toBeInTheDocument());
  });

  it('hides trump card label when stock is exhausted', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, stockRemaining: 0 }));
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByText(/トランプ: 使い切り/)).toBeInTheDocument());
  });

  it('disables play buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => {
      const btn = screen.getByRole('button', { name: '♠ A を出す' });
      expect(btn).toBeDisabled();
    });
  });

  it('shows confirm dialog on reset click and runs reset on accept', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 2 }));
  });

  // #5608: ゲームが終わった後は失うものが無いので、共通の GameResetButton は
  // 確認を飛ばして「次のゲーム」になる。Brusquembille だけ自前ボタンで、決着後も
  // 毎回確認を要求していた。
  it('starts the next game without a confirm once the game has ended', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BrusquembillePhase.GAME_END, gameEndFlag: true }));
    renderWithProviders(<BrusquembillePage />);

    const next = await screen.findByRole('button', { name: '次のゲーム' });
    expect(screen.queryByRole('button', { name: 'リセット' })).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(next);
    // ダイアログを挟まずに走る。
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 2 }));
  });

  it('shows the hint button on the human turn and requests a hint', async () => {
    renderWithProviders(<BrusquembillePage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('renders the server hint text with the translated reason when present', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { cardIndex: 2, reason: 'lead_trump' }, messageCode: 'brusquembille.hintRequested' }),
    );
    renderWithProviders(<BrusquembillePage />);
    const hint = await screen.findByTestId('brusquembille-hint');
    expect(hint).toHaveTextContent('切り札でリードして主導権を握りましょう');
  });

  // **ヒントの窓口は1つ (#4753)。**バナーとツールチップはどちらも同じ
  // `state.hint` から出ている (getBrusquembilleHint がそれを読んでいる) ので矛盾は
  // しないが、要求すると同じ助言が2箇所に重複して出ていた。Cassino と同じく、
  // 明示的に要求したサーバー側の答えを優先して片方に畳む。
  it('hides the frontend hint tooltip while the server hint is shown', async () => {
    localStorage.setItem('hint_enabled_brusquembille', 'true');
    mockExec.mockResolvedValue(
      makeState({ hint: { cardIndex: 2, reason: 'lead_trump' }, messageCode: 'brusquembille.hintRequested' }),
    );
    renderWithProviders(<BrusquembillePage />);

    await screen.findByTestId('brusquembille-hint');
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  // 逆側。**同じ state.hint** を積んだまま messageCode だけ外すと、バナーは
  // 出ずツールチップだけが残る -- 排他が「常にツールチップを消す」に退化して
  // いないことを確かめる。
  it('still shows the frontend hint tooltip when no server hint was requested', async () => {
    localStorage.setItem('hint_enabled_brusquembille', 'true');
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 2, reason: 'lead_trump' } }));
    renderWithProviders(<BrusquembillePage />);

    await screen.findByTestId('hint-tooltip');
    expect(screen.queryByTestId('brusquembille-hint')).not.toBeInTheDocument();
  });
});

// **後半 (追従必須) では出せない札を押せなくする。**
// クローン元のブリスコラは常に自由出しなので、非合法手という概念が無かった。
// ブリュスカンビーユは山札が尽きると追従義務が生まれるので、押せるように
// 見せて実行時にだけ拒否すると、プレイヤーには理由の無いエラーに見える。
// 合法かどうかはバックエンドが決める (validIndices) —— ここで規則を
// 書き直すと二重管理になる。
describe('BrusquembillePage legal moves', () => {
  it('disables the cards the backend says are illegal', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [1], followRequired: true }));
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /を出す/ }).length).toBeGreaterThan(0));

    const cards = screen.getAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
    expect(cards[1]).toBeEnabled();
    expect(cards[2]).toBeDisabled();
  });

  it('leaves every card playable while the stock lasts', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [0, 1, 2], followRequired: false }));
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /を出す/ }).length).toBeGreaterThan(0));

    for (const card of screen.getAllByRole('button', { name: /を出す/ })) {
      expect(card).toBeEnabled();
    }
  });
});

// **席数を選べなければ 2〜5 人卓は誰にも届かない。**
// ドメインが席数可変でも、フォームが無ければ既定の 2 人卓しか始まらない。
describe('BrusquembillePage table size', () => {
  it('offers every supported seat count and resets with the chosen one', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const select = screen.getByLabelText('席数');
    expect(select).toBeInTheDocument();
    const options = Array.from(select.querySelectorAll('option')).map((o) => o.textContent);
    expect(options).toEqual(['2', '3', '4', '5']);

    mockExec.mockClear();
    fireEvent.change(select, { target: { value: '4' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 4 }));
  });

  it('keeps the chosen seat count on a later reset', async () => {
    renderWithProviders(<BrusquembillePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    fireEvent.change(screen.getByLabelText('席数'), { target: { value: '5' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 5 }));

    // **リセットで既定に戻らないこと。** 戻ると選択が無かったことになる。
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 5 }));
  });
});
