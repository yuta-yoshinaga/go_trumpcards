import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pasurApi } from '../api/gameApi';
import * as cardDimensions from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import enCommon from '../i18n/locales/en/common.json';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PasurResponse } from '../types/card';
import { PasurPage } from './PasurPage';

vi.mock('../api/gameApi', () => ({
  pasurApi: { exec: vi.fn() },
  actionLogApi: { pasur: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(pasurApi.exec);
const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
}));

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('DIAMOND', 4), card('SPADE', 13), card('HEART', 9), card('CLOVER', 6)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? hand : [],
  capturedCount: 0,
  soors: 0,
  score: 0,
  ...over,
});

function makeState(overrides: Partial<PasurResponse> = {}): PasurResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    table: [card('SPADE', 7), card('HEART', 4), card('CLOVER', 3)],
    // ♦4 は ♠7 単独、または ♥4+♣3。♠K は取れない。
    captureOptions: [[[0], [1, 2]], [], [], []],
    captureScores: [
      [
        { normal: 0, soorBonus: 0 },
        { normal: 0, soorBonus: 0 },
      ],
      [],
      [],
      [],
    ],
    deckRemaining: 32,
    packsDealt: 1,
    lastCaptureIdx: -1,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winners: [],
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as PasurResponse;
}

beforeEach(() => {
  mobileState.isMobile = true;
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('PasurPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('describes whether each hand card can capture using the server options', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    expect(cards[0]).toHaveAccessibleDescription('捕獲可能');
    expect(cards[1]).toHaveAccessibleDescription('場に置く');
  });

  it('describes hand cards as lay-down when capture options are missing', async () => {
    mockExec.mockResolvedValue(makeState({ captureOptions: [] }));

    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    expect(cards[0]).toHaveAccessibleDescription('場に置く');
  });

  it('describes capture availability in English', async () => {
    await i18n.changeLanguage('en');
    try {
      renderWithProviders(<PasurPage />);
      const cards = await screen.findAllByRole('button', { name: /Select .*$/ });

      expect(cards[0]).toHaveAccessibleDescription('Capture available');
      expect(cards[1]).toHaveAccessibleDescription('Can be laid down');
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  // **11 の合計と絵札の扱いが規則そのもの。**
  it('states the capture rule', async () => {
    renderWithProviders(<PasurPage />);
    const details = await screen.findByTestId('ps-rule-details');
    expect(details).not.toHaveAttribute('open');
    expect(screen.getByTestId('ps-rule')).toHaveTextContent('ルール');
    const ruleText = details.querySelector('div');
    expect(ruleText).toHaveTextContent(/合計が11/);
    expect(ruleText).not.toBeVisible();
  });

  it('opens the rule and CPU seats by default on desktop', async () => {
    vi.spyOn(cardDimensions, 'useIsMobile').mockReturnValue(false);
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-rule-details')).toHaveAttribute('open');
    expect(screen.getByTestId('cpu-accordion')).toHaveAttribute('open');
  });

  it('shows the table, and says when it is empty', async () => {
    const { unmount } = renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-table')).toHaveTextContent(/場/);
    unmount();

    mockExec.mockResolvedValue(makeState({ table: [] }));
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-table')).toHaveTextContent(/なし/);
  });

  it('shows matching zero-based indices on table cards and capture candidates', async () => {
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-table-card-0')).toHaveTextContent('場札 0');
    expect(screen.getByTestId('ps-table-card-1')).toHaveTextContent('場札 1');
    expect(screen.getByTestId('ps-table-card-2')).toHaveTextContent('場札 2');

    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[0]);
    expect(await screen.findByTestId('ps-take-0-btn')).toHaveTextContent('0：');
    expect(screen.getByTestId('ps-take-1-2-btn')).toHaveTextContent('1：');
    expect(screen.getByTestId('ps-take-1-2-btn')).toHaveTextContent('2：');
  });

  it('localizes table card indices in English', async () => {
    await i18n.changeLanguage('en');
    try {
      renderWithProviders(<PasurPage />);
      expect(await screen.findByTestId('ps-table-card-0')).toHaveTextContent('Table card 0');
      const cards = await screen.findAllByRole('button', { name: /Select .*$/ });
      fireEvent.click(cards[0]);
      expect(await screen.findByTestId('ps-take-0-btn')).toHaveTextContent('0:');
      expect(screen.getByTestId('ps-take-1-2-btn')).toHaveTextContent('1:');
      expect(screen.getByTestId('ps-take-1-2-btn')).toHaveTextContent('2:');
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('shows captures and soors for every seat', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { capturedCount: 6, soors: 2, score: 9 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<PasurPage />);
    const s0 = await screen.findByTestId('ps-seat-0');
    expect(s0.closest('details')).toBeNull();
    expect(s0).toHaveTextContent('捕獲6枚');
    expect(s0).toHaveTextContent('スール2');
    expect(s0).toHaveTextContent('得点9');
    expect(screen.getByTestId('ps-seat-3')).toBeInTheDocument();
  });

  it('marks the current human seat', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-turn-0')).toHaveAttribute('aria-current', 'step');
  });

  // **場に残った札の行き先が読めること。**
  it('marks the last capturer only once someone has captured', async () => {
    const { unmount } = renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('ps-seat-2')).not.toHaveTextContent(/最後に捕獲/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastCaptureIdx: 2 }));
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('ps-seat-2')).toHaveTextContent(/最後に捕獲/);
  });

  // **候補はサーバが送ったものだけ。** ここで 11 の部分集合を作り直さない。
  it('offers exactly the capture options the server sent', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[0]);

    expect(await screen.findByTestId('ps-options')).toBeInTheDocument();
    expect(screen.getByTestId('ps-take-0-btn')).toBeInTheDocument();
    expect(screen.getByTestId('ps-take-1-2-btn')).toBeInTheDocument();
    expect(screen.getByTestId('ps-score-0')).toHaveTextContent('通常 0点 + スール加算 0点 = 0点');
    // 送られていない組み合わせは出さない。
    expect(screen.queryByTestId('ps-take-1-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ps-take-0-1-2-btn')).not.toBeInTheDocument();
  });

  it('sends the chosen card and table indices', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[0]);

    mockExec.mockClear();
    fireEvent.click(await screen.findByTestId('ps-take-1-2-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, undefined, [1, 2]));
  });

  // **取れる組み合わせがあるときは場に置けない。** サーバが必ず拒否する。
  it('hides the lay-down button while a capture is available', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    fireEvent.click(cards[0]);
    expect(await screen.findByTestId('ps-must-capture')).toBeInTheDocument();
    expect(screen.queryByTestId('ps-trail-btn')).not.toBeInTheDocument();

    // ♠K は取れないので置ける。
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    expect(await screen.findByTestId('ps-trail-btn')).toBeEnabled();
    expect(screen.queryByTestId('ps-must-capture')).not.toBeInTheDocument();
  });

  it('sends an empty selection when laying a card down', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    fireEvent.click(cards[1]);

    mockExec.mockClear();
    fireEvent.click(await screen.findByTestId('ps-trail-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1, undefined, []));
  });

  it('lets you deselect and choose again', async () => {
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });

    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cards[0]);
    expect(cards[0]).toHaveAttribute('aria-pressed', 'false');
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();

    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByRole('button', { name: '選び直す' }));
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<PasurPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    expect(cards[0]).toBeDisabled();
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winners, expected] of [
      [[0], /あなたの勝ち/],
      [[2], /CPU2 の勝ち/],
      [[1, 3], /2 人が同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winners: [...winners] }));
      const { unmount } = renderWithProviders(<PasurPage />);
      expect(await screen.findByTestId('ps-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('shows every seat score breakdown beside the final result', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winners: [1],
        players: [
          seat(0, { capturedCount: 8, soors: 1, score: 11 }),
          seat(1, { capturedCount: 12, soors: 3, score: 18 }),
          seat(2, { capturedCount: 6, soors: 0, score: 4 }),
          seat(3, { capturedCount: 7, soors: 2, score: 9 }),
        ],
        captureScores: [
          [{ normal: 7, soorBonus: 4 }],
          [{ normal: 10, soorBonus: 8 }],
          [{ normal: 4, soorBonus: 0 }],
          [{ normal: 5, soorBonus: 4 }],
        ],
      }),
    );

    renderWithProviders(<PasurPage />);

    const finalResult = await screen.findByTestId('ps-final-result');
    expect(finalResult).toHaveTextContent('CPU1 の勝ちです。');
    expect(screen.getByTestId('ps-result')).toHaveAttribute('role', 'status');
    expect(screen.getByTestId('ps-result')).toHaveTextContent('CPU1 の勝ちです。');
    expect(finalResult).toHaveTextContent('最終得点内訳');
    expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('札点 7点');
    expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('スール加点 4点');
    expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('最終得点 11点');
    expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('捕獲 8枚');
    expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('スール 1回');
    expect(screen.getByTestId('ps-final-score-1')).toHaveTextContent('最終得点 18点');
    expect(screen.getByTestId('ps-final-score-1')).toHaveTextContent('札点 10点');
    expect(screen.getByTestId('ps-final-score-1')).toHaveTextContent('スール加点 8点');
    expect(screen.getByTestId('ps-final-score-3')).toHaveTextContent('スール 2回');
  });

  it('localizes the final score breakdown in English', async () => {
    await i18n.changeLanguage('en');
    try {
      mockExec.mockResolvedValue(
        makeState({
          gameEndFlag: true,
          phase: 1,
          winners: [0],
          players: [seat(0, { score: 5, capturedCount: 4, soors: 1 }), seat(1), seat(2), seat(3)],
          captureScores: [[{ normal: 3, soorBonus: 2 }], [], [], []],
        }),
      );
      renderWithProviders(<PasurPage />);
      const finalResult = await screen.findByTestId('ps-final-result');
      expect(finalResult).toHaveTextContent('You win!');
      expect(finalResult).toHaveTextContent('Final score breakdown');
      expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('Card points: 3 points');
      expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('Soor bonus: 2 points');
      expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('Final score: 5 points');
      expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('Captured: 4 cards');
      expect(screen.getByTestId('ps-final-score-0')).toHaveTextContent('Soors: 1');
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-0-take-0', reason: 'hint.pasurSoor', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PasurPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/場を空に/);
  });
});

// **スールは「取った結果、場が空になる」こと** (#5762)。倍化を狙うなら、どの
// 選択肢がそれに当たるかがボタンから読めないと選べない。
describe('PasurPage soor options', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('badges the option that clears the table and not the partial one', async () => {
    mockExec.mockResolvedValue(
      makeState({
        table: [card('SPADE', 3), card('HEART', 4), card('CLOVER', 7)],
        // 手札 0 で「場を全部」か「1 枚だけ」かを選べる状況。
        captureOptions: [[[0, 1, 2], [0]], [], [], []],
      } as Partial<PasurResponse>),
    );
    renderWithProviders(<PasurPage />);

    // 手札の 1 枚目を選ぶと、その札の取得候補がボタンになる。
    const handButtons = await screen.findAllByRole('button', { name: /を選ぶ/ });
    fireEvent.click(handButtons[0]);

    expect(await screen.findByTestId('ps-soor-0-1-2')).toHaveTextContent('スール');
    expect(screen.queryByTestId('ps-soor-0')).not.toBeInTheDocument();
    // 読み上げには倍化まで入る。
    expect(screen.getByTestId('ps-soor-0-1-2')).toHaveTextContent('2 倍');
  });

  it('does not badge a full-length option when the table is empty', async () => {
    mockExec.mockResolvedValue(makeState({ table: [], captureOptions: [[], [], [], []] } as Partial<PasurResponse>));
    renderWithProviders(<PasurPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('ps-options')).not.toBeInTheDocument();
  });
});

describe('Pasur result messages', () => {
  it.each([
    ['ja', 'you', 'あなたの勝ちです。', 'leftoverYou', 'あなたが場の残り 3 枚を取りました。', {}],
    ['ja', 'you', 'あなたの勝ちです。', 'leftoverCpu', 'CPU2 が場の残り 3 枚を取りました。', { leftoverIdx: '2' }],
    ['ja', 'cpu', 'CPU1 の勝ちです。', 'leftoverYou', 'あなたが場の残り 3 枚を取りました。', {}],
    ['ja', 'cpu', 'CPU1 の勝ちです。', 'leftoverCpu', 'CPU2 が場の残り 3 枚を取りました。', { leftoverIdx: '2' }],
    ['ja', 'tie', '4 人が同点です。', 'leftoverYou', 'あなたが場の残り 3 枚を取りました。', {}],
    ['ja', 'tie', '4 人が同点です。', 'leftoverCpu', 'CPU2 が場の残り 3 枚を取りました。', { leftoverIdx: '2' }],
    ['en', 'you', 'You win.', 'leftoverYou', 'You took the remaining 3 cards from the table.', {}],
    ['en', 'you', 'You win.', 'leftoverCpu', 'CPU2 took the remaining 3 cards from the table.', { leftoverIdx: '2' }],
    ['en', 'cpu', 'CPU1 wins.', 'leftoverYou', 'You took the remaining 3 cards from the table.', {}],
    ['en', 'cpu', 'CPU1 wins.', 'leftoverCpu', 'CPU2 took the remaining 3 cards from the table.', { leftoverIdx: '2' }],
    ['en', 'tie', '4 players tie.', 'leftoverYou', 'You took the remaining 3 cards from the table.', {}],
    [
      'en',
      'tie',
      '4 players tie.',
      'leftoverCpu',
      'CPU2 took the remaining 3 cards from the table.',
      { leftoverIdx: '2' },
    ],
  ])('%s renders the %s result for a %s recipient', async (lang, result, text, recipient, suffix, recipientParams) => {
    i18n.addResourceBundle('en', 'common', enCommon, true, true);
    await i18n.changeLanguage(lang);
    try {
      mockExec.mockResolvedValue(
        makeState({
          phase: 1,
          gameEndFlag: true,
          messageCode: `pasur.result.${result}.${recipient}`,
          messageParams: { idx: '1', n: '4', leftoverCount: '3', ...recipientParams },
        }),
      );
      renderWithProviders(<PasurPage />);
      expect(await screen.findByText(`${text} ${suffix}`)).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage('ja');
    }
  });
});
