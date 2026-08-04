import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { soloWhistApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSoloWhistState } from '../test/stateFactories';
import { SoloWhistContract } from '../types/phases';
import { SoloWhistPage } from './SoloWhistPage';

vi.mock('../api/gameApi', () => ({
  soloWhistApi: { exec: vi.fn() },
  actionLogApi: { solowhist: vi.fn() },
}));

const mockExec = vi.mocked(soloWhistApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeSoloWhistState();
// A human play turn with a started trick (so the play control is shown).
const playPhaseState = makeSoloWhistState({
  phase: 1,
  declarerIdx: 0,
  contract: 1,
  trumpSuit: 3,
  isHumanBidTurn: false,
  isHumanTurn: true,
  playableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeSoloWhistState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeSoloWhistState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeSoloWhistState({ phase: 3, isHumanBidTurn: false, roundTricks: [8, 2, 2, 1] });
const gameEndState = makeSoloWhistState({
  phase: 4,
  isHumanBidTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});

/**
 * Builds a play-phase fixture with the human declarer holding a given contract,
 * tricks won, and cards left in hand — the three inputs to the contract-progress readout.
 */
function makeProgressState(contract: number, won: number, cardCount: number) {
  return makeSoloWhistState({
    phase: 1,
    declarerIdx: 0,
    contract,
    trumpSuit: contract === 2 ? 0 : 3,
    isHumanBidTurn: false,
    isHumanTurn: true,
    players: [
      { id: 0, isHuman: true, cardCount, cards: [], trickCount: won, score: 0, isDeclarer: true },
      { id: 1, isHuman: false, cardCount, cards: [], trickCount: 0, score: 0, isDeclarer: false },
      { id: 2, isHuman: false, cardCount, cards: [], trickCount: 0, score: 0, isDeclarer: false },
      { id: 3, isHuman: false, cardCount, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    ],
  });
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('SoloWhistPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SoloWhistPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 21 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-1')).toBeInTheDocument();
    expect(screen.getByTestId('bid-2')).toBeInTheDocument();
    expect(screen.getByTestId('bid-3')).toBeInTheDocument();
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<SoloWhistPage />);
    const bidSolo = await screen.findByTestId('bid-1');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bidSolo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [2, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    // Highest bid is 2 (Misère): Solo (1) and Misère (2) are disabled, Abundance (3) and Pass (0) enabled.
    await waitFor(() => expect(screen.getByTestId('bid-1')).toBeDisabled());
    expect(screen.getByTestId('bid-2')).toBeDisabled();
    expect(screen.getByTestId('bid-3')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('shows the current highest bid and a reason tooltip on too-low bids', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [2, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    const info = await screen.findByTestId('sw-highest-bid');
    expect(info).not.toHaveTextContent('まだ入札なし');
    // Too-low bids carry a tooltip on the wrapping span; valid bids do not.
    expect(screen.getByTestId('bid-wrap-1')).toHaveAttribute('title');
    expect(screen.getByTestId('bid-1')).toHaveAttribute('aria-label');
    expect(screen.getByTestId('bid-wrap-3')).not.toHaveAttribute('title');
  });

  it('shows "no bids yet" before anyone has bid', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [0, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByTestId('sw-highest-bid')).toHaveTextContent('まだ入札なし'));
  });

  it('exposes the declarer line as a polite live region', async () => {
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-declarer');
    expect(line).toHaveAttribute('role', 'status');
    expect(line).toHaveAttribute('aria-live', 'polite');
    expect(line.className).not.toContain('animate-pulse');
  });

  it('pulses the declarer line when the contract is decided', async () => {
    // Mount with an undecided contract, then a bid resolves it (declarerIdx -1 → 0).
    mockExec.mockResolvedValue(makeSoloWhistState({ declarerIdx: -1 }));
    renderWithProviders(<SoloWhistPage />);
    const bidSolo = await screen.findByTestId('bid-1');
    mockExec.mockResolvedValue(makeSoloWhistState({ declarerIdx: 0, contract: 1, isHumanBidTurn: false }));
    fireEvent.click(bidSolo);
    await waitFor(() => expect(screen.getByTestId('solowhist-declarer').className).toContain('animate-pulse'));
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('宣言者')).toBeInTheDocument();
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('shows the Solo contract progress in-progress (won < 8, still reachable)', async () => {
    // Solo targets 8 tricks; 3 won with 9 cards left is still reachable → in progress.
    mockExec.mockResolvedValue(makeProgressState(1, 3, 9));
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-contract-progress');
    expect(line).toHaveTextContent('宣言者の進捗: 3 / 8 トリック');
    expect(line).not.toHaveTextContent('達成');
    expect(line).not.toHaveTextContent('失敗確定');
    expect(line.className).toContain('text-ds-warning');
  });

  it('marks the Solo contract as made once the target is reached', async () => {
    // 8 tricks won meets the Solo target → made.
    mockExec.mockResolvedValue(makeProgressState(1, 8, 5));
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-contract-progress');
    expect(line).toHaveTextContent('宣言者の進捗: 8 / 8 トリック');
    expect(line).toHaveTextContent('達成');
    expect(line.className).toContain('text-ds-success');
  });

  it('marks the Solo contract as failed when the target is out of reach', async () => {
    // 2 won + 5 remaining = 7 < 8 → mathematically impossible → failed.
    mockExec.mockResolvedValue(makeProgressState(1, 2, 5));
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-contract-progress');
    expect(line).toHaveTextContent('失敗確定');
    expect(line.className).toContain('text-ds-error');
  });

  it('fails the Misère contract the instant a trick is won', async () => {
    // Misère (contract 2) targets 0 tricks; winning even 1 fails immediately.
    mockExec.mockResolvedValue(makeProgressState(2, 1, 8));
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-contract-progress');
    expect(line).toHaveTextContent('宣言者の進捗: 1 トリック（ミゼール・目標0）');
    expect(line).toHaveTextContent('失敗確定');
    expect(line.className).toContain('text-ds-error');
  });

  it('keeps a clean Misère in progress while no trick is won', async () => {
    // 0 tricks won with cards still in hand → still in progress.
    mockExec.mockResolvedValue(makeProgressState(2, 0, 8));
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-contract-progress');
    expect(line).toHaveTextContent('宣言者の進捗: 0 トリック（ミゼール・目標0）');
    expect(line).not.toHaveTextContent('失敗確定');
    expect(line).not.toHaveTextContent('達成');
    expect(line.className).toContain('text-ds-warning');
  });

  it('does not show the progress readout before the contract is decided', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ declarerIdx: -1 }));
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByTestId('solowhist-declarer')).toBeInTheDocument());
    expect(screen.queryByTestId('solowhist-contract-progress')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'soloWhist.hintRequested',
    });
    renderWithProviders(<SoloWhistPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('describes the misere contract for screen readers', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SoloWhistPage />);
    const misere = await screen.findByTestId(`bid-${SoloWhistContract.MISERE}`);
    const descId = misere.getAttribute('aria-describedby');
    expect(descId).toBe('solowhist-misere-desc');
    expect(document.getElementById(descId ?? '')).toHaveTextContent('1トリックも取らない');
    // Non-misere bids must not borrow the description.
    expect(screen.getByTestId(`bid-${SoloWhistContract.PASS}`)).not.toHaveAttribute('aria-describedby');
  });
});
