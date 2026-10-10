import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doppelkopfApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeDoppelkopfState } from '../test/stateFactories';
import { DoppelkopfPage } from './DoppelkopfPage';

vi.mock('../api/gameApi', () => ({
  doppelkopfApi: { exec: vi.fn() },
  actionLogApi: { doppelkopf: vi.fn() },
}));

const mockExec = vi.mocked(doppelkopfApi.exec);

const playPhaseState = makeDoppelkopfState({ canAnnounce: false });
const announceState = makeDoppelkopfState({ canAnnounce: true, youAreRe: true });
const kontraAnnounceState = makeDoppelkopfState({ canAnnounce: true, youAreRe: false });
const trickEndState = makeDoppelkopfState({
  phase: 1,
  canAnnounce: false,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 10 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 12 } },
  ],
});
const roundEndState = makeDoppelkopfState({
  phase: 2,
  canAnnounce: false,
  teamsRevealed: true,
  reTeam: [true, false, true, false],
  roundRePoints: 130,
  liveRePoints: 130,
  liveKontraPoints: 110,
  roundReWon: true,
  roundGamePoints: 2,
});
const gameEndState = makeDoppelkopfState({
  phase: 3,
  canAnnounce: false,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeDoppelkopfState({ currentPlayerIdx: 1, canAnnounce: false });

// A hand mixing trumps (♥10, ♦K, ♣Q) with fail cards (♠A, ♣10) for highlight tests.
const mixedHandState = makeDoppelkopfState({
  canAnnounce: false,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART', value: 10 }, // trump (Dulle)
        { design: 'SPADE', value: 1 }, // fail (♠A)
        { design: 'DIAMOND', value: 13 }, // trump (♦K)
        { design: 'CLOVER', value: 10 }, // fail (♣10)
        { design: 'CLOVER', value: 12 }, // trump (♣Q)
      ],
      trickCount: 0,
      chips: 20,
      isRe: false,
      teamKnown: true,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false, teamKnown: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false, teamKnown: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false, teamKnown: false },
  ],
  playableIndices: [0, 1, 2, 3, 4],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('DoppelkopfPage', () => {
  it('keeps the human player summary outside the CPU accordion', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const accordion = await screen.findByTestId('cpu-accordion');
    expect(accordion).not.toContainElement(screen.getByText(/あなた.*チップ/));
    expect(screen.getByText(/あなた.*チップ/)).toBeVisible();
  });

  it('shows each player’s remaining hand count and refreshes it from the response', async () => {
    const initial = makeDoppelkopfState({
      players: makeDoppelkopfState().players.map((player, index) => ({ ...player, cardCount: 12 - index })),
    });
    const updated = makeDoppelkopfState({
      players: makeDoppelkopfState().players.map((player, index) => ({ ...player, cardCount: 8 - index })),
    });
    mockExec.mockResolvedValueOnce(initial).mockResolvedValue(updated);
    const { container } = renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(container).toHaveTextContent('12枚'));
    expect(container).toHaveTextContent('11枚');
    expect(container).toHaveTextContent('10枚');
    expect(container).toHaveTextContent('9枚');
    fireEvent.click(await screen.findByAltText('♥ 10'));
    fireEvent.click(await screen.findByRole('button', { name: '出す' }));
    await waitFor(() => expect(container).toHaveTextContent('8枚'));
  });

  it('renders the server supplied round by player chip history', async () => {
    mockExec.mockResolvedValue(makeDoppelkopfState({ roundScoreHistory: [[2, -2, 2, -2]] }));
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByTestId('dk-round-history')).toBeInTheDocument());
    expect(screen.getByTestId('dk-round-history')).toHaveTextContent('ラウンドごとのチップ増減');
    expect(screen.getByTestId('dk-round-history')).toHaveTextContent('+2');
    expect(screen.getByTestId('dk-round-history')).toHaveTextContent('-2');
  });
  it('shows a known team label before the teams are fully revealed', async () => {
    mockExec.mockResolvedValue(
      makeDoppelkopfState({
        players: [
          { id: 0, isHuman: true, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: true, teamKnown: true },
          { id: 1, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: true, teamKnown: true },
          { id: 2, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: false, teamKnown: false },
          { id: 3, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: false, teamKnown: false },
        ],
        teamsRevealed: false,
      }),
    );
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1 \[Re\]/)).toBeInTheDocument());
    expect(screen.queryByText(/CPU 2 \[Kontra\]/)).not.toBeInTheDocument();
  });

  it('keeps the team and announcement status compact on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
    mockExec.mockResolvedValue(
      makeDoppelkopfState({
        canAnnounce: false,
        youAreRe: true,
        reAnnounced: true,
      }),
    );
    const { unmount } = renderWithProviders(<DoppelkopfPage />);
    const teamStatus = await screen.findByText(/あなたのチーム:/);
    const status = teamStatus.parentElement;
    expect(status).toHaveClass('flex', 'flex-wrap');
    expect(status?.querySelectorAll('span')).toHaveLength(2);
    expect(status).toHaveTextContent('あなたのチーム: Re');
    expect(status).toHaveTextContent('宣言済み: Re');
    unmount();
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalWidth });
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoppelkopfPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, baseChips: 2, startChips: 20, targetChips: 40 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ 10')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const card = await screen.findByAltText('♥ 10');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows the hint button on the human turn and requests a hint', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows the Re announce button when the human is Re and can announce', async () => {
    mockExec.mockResolvedValue(announceState);
    renderWithProviders(<DoppelkopfPage />);
    const btn = await screen.findByRole('button', { name: /Re を宣言/ });
    // aria-label adds the timing context while still containing the visible label.
    expect(btn.getAttribute('aria-label')).toContain('Re を宣言');
    expect(btn.getAttribute('aria-label')).toContain('第1トリック');
    expect(btn).toHaveAttribute('title');
    expect(screen.getByTestId('dk-announce-stage')).toHaveTextContent('第1トリック');
    expect(screen.getByTestId('dk-announce-stage')).toHaveTextContent('Re');
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('announce'));
  });

  it('shows the Kontra announce button when the human is not Re', async () => {
    mockExec.mockResolvedValue(kontraAnnounceState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Kontra を宣言/ })).toBeInTheDocument());
  });

  it('does not show the announce button when canAnnounce is false', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 10')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /宣言/ })).not.toBeInTheDocument();
    expect(screen.queryByTestId('dk-announce-stage')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('names the trick winner and the points they took', async () => {
    mockExec.mockResolvedValue({
      ...trickEndState,
      messageCode: 'doppelkopf.trickEnd.cpuWin',
      messageParams: { winnerId: '2', points: '28' },
    });
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2 が 28点獲得/)).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 10')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders the collapsible trump ordering legend', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const legend = await screen.findByTestId('dk-trump-legend');
    expect(legend).toBeInTheDocument();
    expect(legend).not.toHaveAttribute('open');
    expect(screen.getByText('ルール')).toBeInTheDocument();
    expect(screen.getByText('切り札序列')).not.toBeVisible();
    fireEvent.click(screen.getByText('ルール'));
    expect(screen.getByText('切り札序列')).toBeVisible();
    // The strongest and weakest trumps appear in the ordering.
    expect(screen.getByText('♥10')).toBeInTheDocument();
    expect(screen.getByText('♦9')).toBeInTheDocument();
  });

  it('collapses player statistics on mobile and opens them on desktop', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
    const mobile = renderWithProviders(<DoppelkopfPage />);
    const mobileAccordion = await screen.findByTestId('cpu-accordion');
    expect(mobileAccordion).not.toHaveAttribute('open');
    expect(screen.getByText(/CPU 1:/)).not.toBeVisible();
    mobile.unmount();

    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1280 });
    renderWithProviders(<DoppelkopfPage />);
    const desktopAccordion = await screen.findByTestId('cpu-accordion');
    expect(desktopAccordion).toHaveAttribute('open');
    expect(screen.getByText(/CPU 1:/)).toBeVisible();
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalWidth });
  });

  it('rings trump cards in the hand and leaves fail cards unmarked', async () => {
    mockExec.mockResolvedValue(mixedHandState);
    renderWithProviders(<DoppelkopfPage />);
    await screen.findByAltText('♥ 10');
    // Trumps: ♥10 (Dulle), ♦K, ♣Q.
    for (const alt of ['♥ 10', '♦ K', '♣ Q']) {
      const button = screen.getByAltText(alt).closest('button');
      expect(button).toHaveAttribute('data-trump', 'true');
      expect(button).toHaveAccessibleName(`${alt} (切り札)`);
    }
    // Fail cards: ♠A, ♣10.
    for (const alt of ['♠ A', '♣ 10']) {
      const button = screen.getByAltText(alt).closest('button');
      expect(button).not.toHaveAttribute('data-trump');
      expect(button).toHaveAccessibleName(alt);
    }
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<DoppelkopfPage />);
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
      // このページのバナーは `cardIndices` を並べる。`cardIndex` を渡しても
      // 何も出ないので、押した側のテストにならない。
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'doppelkopf.hintRequested',
    });
    renderWithProviders(<DoppelkopfPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **途中経過は liveRePoints / liveKontraPoints をそのまま出す。**
  // 以前は roundRePoints (プレイ中常に 0) を使って Kontra を 240 - roundRePoints
  // で計算していたため、ラウンド中ずっと「Re: 0点 / Kontra: 240点」となっていた (#6435)。
  it('shows the running card points during play from liveRePoints and liveKontraPoints', async () => {
    mockExec.mockResolvedValue(makeDoppelkopfState({ phase: 0, liveRePoints: 30, liveKontraPoints: 20 }));
    renderWithProviders(<DoppelkopfPage />);

    const panel = await screen.findByTestId('dk-live-points');
    expect(panel).toHaveTextContent('30');
    // Kontra 側は 240 - 30 = 210 ではなく、実際に獲得した 20。
    expect(panel).toHaveTextContent('20');
    expect(panel).not.toHaveTextContent('210');
    expect(panel).toHaveTextContent('121');
  });

  // 逆側。ラウンド終了後は下の内訳が引き継ぐので、二重に出さない。
  it('hides the running panel once the round has ended', async () => {
    mockExec.mockResolvedValue(
      makeDoppelkopfState({ phase: 2, liveRePoints: 30, liveKontraPoints: 20, roundRePoints: 85 }),
    );
    renderWithProviders(<DoppelkopfPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('dk-live-points')).not.toBeInTheDocument();
  });
});
