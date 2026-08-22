import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { putApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PutResponse } from '../types/card';
import { PutPage } from './PutPage';

vi.mock('../api/gameApi', () => ({
  putApi: { exec: vi.fn() },
  actionLogApi: { put: vi.fn() },
}));

const mockExec = vi.mocked(putApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<PutResponse> = {}): PutResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 5), card('DIAMOND', 11)],
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0 },
    ],
    phase: 0,
    handNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    responderIdx: -1,
    currentTrick: [],
    trickResults: [],
    leadPlayerIdx: 0,
    manoIdx: 1,
    dealerIdx: 0,
    handStake: 1,
    acceptedLevel: 0,
    pendingLevel: 0,
    putCallerIdx: -1,
    canDeclarePut: true,
    matchTarget: 15,
    matchPoints: [0, 0],
    handWinnerIdx: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, matchTarget: 15 },
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockResolvedValue(makeState());
});

describe('PutPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));
  });

  it('plays a card with number keys and declares put with t during the play turn', async () => {
    mockExec.mockResolvedValue(makeState()); // PLAY phase, human turn, canDeclarePut
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: '1' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
    fireEvent.keyDown(document, { key: 't' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('put'));
  });

  it('accepts and declines with a/d during the respond phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, responderIdx: 0, putCallerIdx: 1 }));
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('accept'));
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('decline'));
  });

  it('ignores play/respond shortcuts outside their phases', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, gameEndFlag: true })); // GAME_END
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: 'a' });
    fireEvent.keyDown(document, { key: 't' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', 0);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('accept');
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('put');
  });

  it('renders match score and stake header', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByText(/マッチ得点/)).toBeInTheDocument());
    expect(screen.getByTestId('put-header')).toHaveTextContent(/目標/);
    expect(screen.getByTestId('put-stake')).toBeInTheDocument();
  });

  it('exposes the tutorial target elements for the guided tour', async () => {
    // A card on the table so the trick area (conditionally rendered) is present.
    mockExec.mockResolvedValue(makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 4) }] }));
    const { container } = renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByText(/マッチ得点/)).toBeInTheDocument());
    for (const target of ['put-score', 'put-rankref', 'put-trick', 'put-hand', 'put-call']) {
      expect(container.querySelector(`[data-tutorial="${target}"]`)).not.toBeNull();
    }
  });

  it('renders the collapsible card-strength reference panel', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByTestId('put-rank-ref')).toBeInTheDocument());
    // Panel content mirrors the Go domain strength order: one tier, suits
    // irrelevant. **Put has no matadores and no stripped deck** — the clone
    // source (Truco) does, and showing its two-tier panel here would state a
    // rule Put does not have.
    expect(screen.getByText('カードの強さ（強い順）')).toBeInTheDocument();
    expect(screen.getByText('3 ＞ 2 ＞ A ＞ K ＞ Q ＞ J ＞ 10 ＞ 9 ＞ 8 ＞ 7 ＞ 6 ＞ 5 ＞ 4')).toBeInTheDocument();
    expect(screen.queryByText(/マタドール/)).not.toBeInTheDocument();
    expect(screen.queryByText(/8・9・10 は使いません/)).not.toBeInTheDocument();
  });

  it('shows human hand as 3 play buttons', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ J を出す' })).toBeInTheDocument();
  });

  it('fires play with the selected card index when a card is clicked', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♥ 5 を出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('shows the Put button on the human turn and dispatches put', async () => {
    renderWithProviders(<PutPage />);
    const btn = await screen.findByRole('button', { name: 'Put を宣言' });

    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('put'));
  });

  it('shows accept/decline buttons in respond phase and dispatches accept', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, responderIdx: 0, putCallerIdx: 1, pendingLevel: 1 }));
    renderWithProviders(<PutPage />);
    const acceptBtn = await screen.findByRole('button', { name: '受諾' });
    expect(screen.getByRole('button', { name: '拒否' })).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(acceptBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('accept'));
  });

  it('shows Next button at baza end and dispatches next', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<PutPage />);
    const next = await screen.findByRole('button', { name: '次へ' });

    mockExec.mockClear();
    fireEvent.click(next);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows youWin banner when winnerIdx is 0', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, gameEndFlag: true, winnerIdx: 0, matchPoints: [15, 8] }));
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝ち！.*15.*8/)).toBeInTheDocument());
  });

  it('shows cpuWin banner when winnerIdx is 1', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4, gameEndFlag: true, winnerIdx: 1, matchPoints: [8, 15] }));
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByText(/CPUの勝ち.*8.*15/)).toBeInTheDocument());
  });

  it('disables play buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, canDeclarePut: false }));
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeDisabled());
  });

  it('shows confirm dialog on reset click and runs reset on accept', async () => {
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));
  });

  it('hides the hint tooltip by default and reveals a reasoned recommendation when the toggle is enabled', async () => {
    localStorage.removeItem('hint_enabled_put');
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(screen.getByTestId('put-hint-toggle')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).toBeNull();

    fireEvent.click(screen.getByRole('checkbox'));
    // Default hand leads with 1♠ (top matador) and canDeclarePut → declare-Put advice.
    await waitFor(() =>
      expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('強い手です。Put を宣言しましょう'),
    );
  });

  it('shows the follow-to-win hint when a stronger card can beat the lead', async () => {
    localStorage.setItem('hint_enabled_put', 'true');
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 5) }],
        canDeclarePut: false,
        players: [
          { id: 0, isHuman: true, cardCount: 2, cards: [card('HEART', 4), card('HEART', 3)], trickCount: 0 },
          { id: 1, isHuman: false, cardCount: 2, cards: [], trickCount: 0 },
        ],
      }),
    );
    renderWithProviders(<PutPage />);
    await waitFor(() =>
      expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('強い札を出してこのトリックを取りましょう'),
    );
    localStorage.removeItem('hint_enabled_put');
  });

  it('sends the chosen match target on reset', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 15 }));

    fireEvent.change(screen.getByLabelText('マッチ目標点'), { target: { value: '24' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { matchTarget: 24 }));
  });
});
