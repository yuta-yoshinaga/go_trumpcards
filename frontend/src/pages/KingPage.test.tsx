import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kingApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKingState } from '../test/stateFactories';
import { KingPage } from './KingPage';

vi.mock('../api/gameApi', () => ({
  kingApi: { exec: vi.fn() },
  actionLogApi: { king: vi.fn() },
}));

const mockExec = vi.mocked(kingApi.exec);

const playPhaseState = makeKingState();
const selectPhaseState = makeKingState({
  phase: 'selectContract',
  dealerIdx: 0,
  currentTurn: 0,
  isHumanTurn: true,
  currentContract: -1,
});
const cpuSelectState = makeKingState({
  phase: 'selectContract',
  dealerIdx: 1,
  currentTurn: 1,
  isHumanTurn: false,
  currentContract: -1,
});
const dealEndState = makeKingState({
  phase: 'dealEnd',
  lastDealDetail: { contract: 0, trumpSuit: -1, dealerIdx: 0, gained: { 0: -20, 1: -30, 2: -10, 3: -30 } },
});
const gameEndState = makeKingState({
  phase: 'gameEnd',
  gameEndFlag: true,
  roundWinners: [0],
  message: 'ゲーム終了！',
});
const cpuTurnState = makeKingState({ currentTurn: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('KingPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KingPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<KingPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Dealer badge', async () => {
    renderWithProviders(<KingPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // Seat 0 is the default dealer.
    expect(screen.getByText('親')).toBeInTheDocument();
  });

  it('renders the select-contract phase with contract buttons and a prompt', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /^ノートリック —/ })).toBeInTheDocument();
  });

  it('shows achieve/avoid badges on the contract buttons', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-prompt')).toBeInTheDocument());
    // No Hearts (contract 1) is an avoid contract targeting hearts.
    const avoidBadge = screen.getByTestId('king-contract-badge-1');
    expect(avoidBadge).toHaveTextContent('回避');
    expect(avoidBadge).toHaveTextContent('♥');
    // King (Trump) (contract 6) is the only achieve contract.
    const achieveBadge = screen.getByTestId('king-contract-badge-6');
    expect(achieveBadge).toHaveTextContent('獲得');
  });

  it('choosing a non-trump contract dispatches contract with trumpSuit -1', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    const btn = await screen.findByRole('button', { name: /^ノートリック —/ });
    mockExec.mockClear();
    mockExec.mockResolvedValue(selectPhaseState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 0, trumpSuit: -1 }));
  });

  it('choosing King (Trump) shows a trump picker then dispatches the chosen suit', async () => {
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    const kingBtn = await screen.findByRole('button', { name: /^キング（切り札） —/ });
    fireEvent.click(kingBtn);
    // The trump prompt and suit buttons appear.
    await waitFor(() => expect(screen.getByTestId('king-trump-prompt')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(selectPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♥' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 6, trumpSuit: 3 }));
  });

  it('shows a CPU-selecting message when the dealer is a CPU', async () => {
    mockExec.mockResolvedValue(cpuSelectState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-select-cpu')).toBeInTheDocument());
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<KingPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 0 }));
  });

  it('renders deal end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(dealEndState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('king-deal-result')).toBeInTheDocument();
  });

  it('shows the contract breakdown with the contract name, loss basis, and per-player points', async () => {
    mockExec.mockResolvedValue(dealEndState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-deal-breakdown')).toBeInTheDocument());
    // Contract 0 = No Tricks; its label and loss basis are rendered.
    const breakdown = screen.getByTestId('king-deal-breakdown');
    expect(breakdown).toHaveTextContent('契約: ノートリック');
    expect(breakdown).toHaveTextContent('失点根拠: トリックを取るごとに失点。');
    // Per-player point rows match the gained map from lastDealDetail.
    expect(screen.getByTestId('king-deal-breakdown-row-0')).toHaveTextContent('-20点');
    expect(screen.getByTestId('king-deal-breakdown-row-1')).toHaveTextContent('-30点');
    expect(screen.getByTestId('king-deal-breakdown-row-2')).toHaveTextContent('-10点');
  });

  it('shows the trump suit in the breakdown for the King (Trump) contract', async () => {
    mockExec.mockResolvedValue(
      makeKingState({
        phase: 'dealEnd',
        lastDealDetail: { contract: 6, trumpSuit: 3, dealerIdx: 0, gained: { 0: -8, 1: 0, 2: 0, 3: 0 } },
      }),
    );
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByTestId('king-deal-breakdown')).toBeInTheDocument());
    const breakdown = screen.getByTestId('king-deal-breakdown');
    expect(breakdown).toHaveTextContent('契約: キング（切り札）（切り札: ♥）');
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<KingPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      // このページのバナーは `cardIndices` を並べる。`cardIndex` は型に無い。
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'king.hintRequested',
    });
    renderWithProviders(<KingPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('carries the contract type and description into the accessible name', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(selectPhaseState);
    renderWithProviders(<KingPage />);
    // The badge is aria-hidden, so the same information has to reach the
    // accessible name or a screen reader hears only the contract's name.
    const avoid = await screen.findByTestId('king-contract-1');
    const achieve = await screen.findByTestId('king-contract-6');
    const descId = avoid.getAttribute('aria-describedby');
    expect(descId).toBe('king-contract-desc-1');
    expect(document.getElementById(descId ?? '')).toHaveTextContent(i18n.t('king:contractDesc.1'));
    expect(avoid).toHaveAccessibleName(new RegExp(i18n.t('king:contractType.avoid')));
    expect(achieve).toHaveAccessibleName(new RegExp(i18n.t('king:contractType.achieve')));
  });
});
