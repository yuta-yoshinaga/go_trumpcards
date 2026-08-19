import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { looApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeLooState } from '../test/stateFactories';
import { LooPage } from './LooPage';

vi.mock('../api/gameApi', () => ({
  looApi: { exec: vi.fn() },
  actionLogApi: { loo: vi.fn() },
}));

const mockExec = vi.mocked(looApi.exec);

const playPhaseState = makeLooState();
const decidePhaseState = makeLooState({
  phase: 0,
  decidePlayerIdx: 0,
  isHumanTurn: true,
});
const roundEndState = makeLooState({
  phase: 3,
  lastDealDetail: {
    potStart: 12,
    trumpSuit: 1,
    playing: [true, true, true, false],
    tricks: { 0: 3, 1: 2, 2: 0, 3: 0 },
    gained: { 0: 4, 1: 1, 2: -12, 3: 0 },
    looed: [2],
    potCarry: 12,
  },
});
const cpuTurnState = makeLooState({ currentTurn: 1, isHumanTurn: false });
const cpuDecideState = makeLooState({ phase: 0, decidePlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('LooPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<LooPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<LooPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, ante: 3 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<LooPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('marks each player Play/Pass with an icon and a status aria-label', async () => {
    renderWithProviders(<LooPage />); // player 0 playing, player 3 passing
    const playing = await screen.findByTestId('loo-status-0');
    expect(playing).toHaveAttribute('role', 'status');
    expect(playing).toHaveTextContent('●');
    expect(playing).toHaveAttribute('aria-label', expect.stringContaining('参加'));
    expect(playing).toHaveAttribute('title'); // risk tooltip
    const passing = screen.getByTestId('loo-status-3');
    expect(passing).toHaveTextContent('○');
    expect(passing).toHaveAttribute('aria-label', expect.stringContaining('降り'));
  });

  // #5693: ルーの罰金 (looed = ポット全額) には下限が無いので、チップ残高は
  // 実際に赤字になる。色だけに頼らず、記号と aria-label でも警告する。
  it('warns when a chip balance has gone negative', async () => {
    mockExec.mockResolvedValue(
      makeLooState({
        players: [
          { id: 0, isHuman: true, cardCount: 3, cards: [], trickCount: 0, playing: true, chips: -12 },
          { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0, playing: true, chips: 0 },
          { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 0, playing: true, chips: 5 },
          { id: 3, isHuman: false, cardCount: 3, cards: [], trickCount: 0, playing: false, chips: -1 },
        ],
      }),
    );
    renderWithProviders(<LooPage />);

    const inDebt = await screen.findByTestId('loo-chips-0');
    expect(inDebt).toHaveClass('text-ds-error');
    expect(inDebt).toHaveTextContent('▼');
    expect(inDebt).toHaveAttribute('aria-label', expect.stringContaining('12'));

    // 0 と正の残高は現状どおり中立。
    expect(screen.getByTestId('loo-chips-1')).not.toHaveClass('text-ds-error');
    expect(screen.getByTestId('loo-chips-1')).not.toHaveTextContent('▼');
    expect(screen.getByTestId('loo-chips-2')).not.toHaveClass('text-ds-error');
    expect(screen.getByTestId('loo-chips-3')).toHaveClass('text-ds-error');
  });

  it('renders the decide phase with play and pass buttons', async () => {
    mockExec.mockResolvedValue(decidePhaseState);
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(screen.getByTestId('loo-decide-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '参加' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument();
  });

  it('deciding to play dispatches decide with play=true', async () => {
    mockExec.mockResolvedValue(decidePhaseState);
    renderWithProviders(<LooPage />);
    const playBtn = await screen.findByRole('button', { name: '参加' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(decidePhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('decide', { play: true }));
  });

  it('deciding to pass dispatches decide with play=false', async () => {
    mockExec.mockResolvedValue(decidePhaseState);
    renderWithProviders(<LooPage />);
    const passBtn = await screen.findByRole('button', { name: '降りる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(decidePhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('decide', { play: false }));
  });

  it('shows the pot reward and loo risk at the decide phase', async () => {
    // Default deal pot/potStart = 12 → 2 per trick, so a sweep pays 10, not 12:
    // the remainder stays in the pot. The penalty is the full 12 (#4921).
    mockExec.mockResolvedValue(decidePhaseState);
    renderWithProviders(<LooPage />);
    const potRisk = await screen.findByTestId('loo-pot-risk');
    expect(potRisk).toHaveTextContent('+10');
    expect(potRisk).toHaveTextContent('+2');
    expect(potRisk).toHaveTextContent('-12');
    // **ポット全額は入らない。**「最大 +12」は実際より多く見せることになる。
    expect(potRisk).not.toHaveTextContent('+12');
  });

  it('does not show the pot-risk block outside the human decide turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByTestId('loo-pot-risk')).not.toBeInTheDocument();
  });

  it('shows a CPU decide notice on a CPU decide turn', async () => {
    mockExec.mockResolvedValue(cpuDecideState);
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(screen.getByTestId('loo-decide-cpu')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '参加' })).not.toBeInTheDocument();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<LooPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders deal end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('clicking next deal dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<LooPage />);
    const nextBtn = await screen.findByRole('button', { name: '次のディール' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], decision: null, reason: 'x' } });
    renderWithProviders(<LooPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], decision: null, reason: 'x' },
      messageCode: 'loo.hintRequested',
    });
    renderWithProviders(<LooPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
