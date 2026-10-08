import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, doubleattackApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, DoubleAttackResponse } from '../types/card';
import { DOUBLE_ATTACK_RESULT } from '../types/games/doubleattack';
import { DoubleAttackPhase } from '../types/phases';
import { DoubleAttackPage } from './DoubleAttackPage';

vi.mock('../api/gameApi', () => ({
  doubleattackApi: { exec: vi.fn() },
  actionLogApi: { doubleattack: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(doubleattackApi.exec);
const mockActionLogApi = vi.mocked(actionLogApi.doubleattack);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (value: number): Card => ({ design: 'SPADE', value });

const hand = (over: Partial<DoubleAttackResponse['hands'][number]> = {}) =>
  ({
    cards: [card(13), card(7)],
    score: 17,
    bet: 100,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
    ...over,
  }) as DoubleAttackResponse['hands'][number];

const base: DoubleAttackResponse = {
  phase: DoubleAttackPhase.BET,
  hands: [],
  activeHand: 0,
  dealerCards: [],
  dealerScore: 0,
  dealerHoleDealt: false,
  maxAttackBet: 0,
  canDouble: false,
  canSplit: false,
  anteBet: 0,
  attackBet: 0,
  bustItBet: 0,
  payout: 0,
  bustItPayout: 0,
  bustItPayouts: [
    { cards: 3, multiplier: 1 },
    { cards: 4, multiplier: 2 },
    { cards: 5, multiplier: 8 },
    { cards: 6, multiplier: 25 },
    { cards: 7, multiplier: 100 },
    { cards: 8, multiplier: 500 },
  ],
  chips: 1000,
  roundNumber: 1,
  remainingCards: 384,
  gameEndFlag: false,
  message: '',
};

const withState = (over: Partial<DoubleAttackResponse>): DoubleAttackResponse => ({ ...base, ...over });

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  } as unknown as ReturnType<typeof useCliMode>);
});

describe('DoubleAttackPage', () => {
  it('announces the newly active hand and score only when the active hand changes during play', async () => {
    mockApi
      .mockResolvedValueOnce(
        withState({
          phase: DoubleAttackPhase.PLAY,
          hands: [hand(), hand({ score: 19 })],
          activeHand: 0,
          canDouble: false,
          canSplit: false,
        }),
      )
      .mockResolvedValueOnce(
        withState({
          phase: DoubleAttackPhase.PLAY,
          hands: [hand({ score: 18 }), hand({ score: 19 })],
          activeHand: 1,
        }),
      );
    renderWithProviders(<DoubleAttackPage />);

    const announcement = await screen.findByTestId('da-active-hand-announcement');
    expect(announcement).toBeEmptyDOMElement();
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(announcement).toHaveTextContent('手札2'));
    expect(announcement).toHaveTextContent('19点');
  });

  it('marks the final Bust It payout row as covering that many cards or more', async () => {
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.BET }));
    renderWithProviders(<DoubleAttackPage />);

    const payoutTable = await screen.findByTestId('da-bustit-payouts');
    expect(payoutTable).toHaveTextContent('3枚: 1:1');
    expect(payoutTable).toHaveTextContent('8枚以上: 500:1');
  });

  // ディーラーのアップカードは、カードが配られた以降のフェーズで表示される
  it('shows dealer cards when dealer has cards', async () => {
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.PLAY, dealerCards: [card(10)] }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-dealer-cards')).toBeInTheDocument());
  });

  it('hides dealer cards when dealer has no cards', async () => {
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.BET, dealerCards: [] }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.queryByTestId('da-dealer-cards')).not.toBeInTheDocument());
  });

  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('賭けフェーズでは配るボタンを出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
  });

  it('賭ける前にサーバーから受け取った Bust It 配当を表示する', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    const payouts = await screen.findByTestId('da-bustit-payouts');
    expect(payouts).toHaveTextContent('Bust It の配当（ディーラーがバスト）');
    expect(payouts).toHaveTextContent('3枚: 1:1');
    expect(payouts).toHaveTextContent('8枚以上: 500:1');
  });

  it('配るはアンティと Bust It を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { ante: 50, bustIt: 0 }));
  });

  it('アンティを増やすと Bust It を残高内に切り下げ、減らしても戻さない', async () => {
    mockApi.mockResolvedValue(withState({ chips: 200 }));
    renderWithProviders(<DoubleAttackPage />);
    const anteInput = await screen.findByLabelText('アンティ');
    const bustItInput = screen.getByLabelText('Bust It');

    fireEvent.change(bustItInput, { target: { value: '150' } });
    fireEvent.change(anteInput, { target: { value: '100' } });
    expect(bustItInput).toHaveValue('100');
    expect(Number(anteInput.getAttribute('value')) + Number(bustItInput.getAttribute('value'))).toBeLessThanOrEqual(
      200,
    );

    fireEvent.change(anteInput, { target: { value: '50' } });
    expect(bustItInput).toHaveValue('100');
  });

  // **追加ベットの前は 1 枚だけで、点数も出ない。**
  //
  // これは伏せているのではなく、サーバがまだ 2 枚目を持っていないため。
  it('追加ベットの前はアップカードだけを出し、点数を出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        anteBet: 50,
        dealerCards: [card(6)],
        dealerHoleDealt: false,
        maxAttackBet: 50,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-dealer-hidden')).toBeInTheDocument());
    expect(screen.queryByTestId('da-dealer-score')).not.toBeInTheDocument();
    expect(screen.getByTestId('da-attack-notice')).toBeInTheDocument();
  });

  // **見送りは amount 0 を送ること。** 送らないのとは違う。
  it('見送りは amount 0 を送る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        anteBet: 50,
        dealerCards: [card(6)],
        maxAttackBet: 50,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '見送る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '見送る' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('attack', { amount: 0 }));
  });

  it('追加ベットのキーで指定額を送る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        maxAttackBet: 50,
        dealerCards: [card(6)],
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭け増す' })).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/追加ベット/), { target: { value: '30' } });
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('attack', { amount: 30 }));
  });

  it('追加ベットのボタンで指定額を送る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        maxAttackBet: 50,
        dealerCards: [card(6)],
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭け増す' })).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/追加ベット/), { target: { value: '30' } });
    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭け増す' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('attack', { amount: 30 }));
  });

  it('追加ベット額・サーバー上限・チップを入力に合わせて表示する', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.ATTACK, chips: 730, maxAttackBet: 50, dealerCards: [card(6)] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-attack-comparison')).toHaveTextContent('チップ: 730'));
    expect(screen.getByTestId('da-attack-comparison')).toHaveTextContent('追加ベット上限: 50');
    expect(screen.getByTestId('da-attack-comparison')).toHaveTextContent('入力額: 0');
    expect(screen.getByLabelText('追加ベット')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/追加ベット/), { target: { value: '30' } });
    expect(screen.getByTestId('da-attack-comparison')).toHaveTextContent('入力額: 30');
  });

  it('見送りのキーで attack に amount 0 を送る', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.ATTACK, maxAttackBet: 50, dealerCards: [card(6)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '見送る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'q' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('attack', { amount: 0 }));
  });

  it('canDouble が真のときダブルのキーが効く', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.PLAY, canDouble: true, dealerCards: [card(6), card(9)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-double')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('double'));

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('da-double'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('double'));
  });

  it('canSplit が真のときスプリットのキーが効く', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.PLAY, canSplit: true, dealerCards: [card(6), card(9)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-split')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'p' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('split'));
  });

  it('canDouble が偽のときダブルのキーは効かない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.PLAY, canDouble: false, dealerCards: [card(6), card(9)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockApi).not.toHaveBeenCalled());
  });

  it('canSplit が偽のときスプリットのキーは効かない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.PLAY, canSplit: false, dealerCards: [card(6), card(9)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'p' });
    await waitFor(() => expect(mockApi).not.toHaveBeenCalled());
  });

  it('プレイフェーズ以外ではダブルとスプリットのキーは効かない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.ATTACK,
        canDouble: true,
        canSplit: true,
        dealerCards: [card(6)],
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '見送る' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'p' });
    await waitFor(() => expect(mockApi).not.toHaveBeenCalled());
  });

  it('既存の h / s / n キーが従来どおり動く', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.PLAY, dealerCards: [card(6), card(9)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 's' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stand'));

    cleanup();
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.RESULT, hands: [hand()] }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'n' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('7つの操作ボタンに aria-keyshortcuts を付ける', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        canDouble: true,
        canSplit: true,
        dealerCards: [card(6), card(9)],
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-double')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: 'ヒット' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: 'スタンド' })).toHaveAttribute('aria-keyshortcuts', 's');
    expect(screen.getByRole('button', { name: 'ダブル' })).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(screen.getByRole('button', { name: 'スプリット' })).toHaveAttribute('aria-keyshortcuts', 'p');

    cleanup();
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.ATTACK, maxAttackBet: 50, dealerCards: [card(6)], hands: [hand()] }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭け増す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '賭け増す' })).toHaveAttribute('aria-keyshortcuts', 'a');
    expect(screen.getByRole('button', { name: '見送る' })).toHaveAttribute('aria-keyshortcuts', 'q');

    cleanup();
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.RESULT, hands: [hand()] }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次のラウンド' })).toHaveAttribute('aria-keyshortcuts', 'n');
  });

  it('2枚目が配られたら点数を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerScore: 15,
        dealerHoleDealt: true,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-dealer-score')).toHaveTextContent('15'));
    expect(screen.queryByTestId('da-dealer-hidden')).not.toBeInTheDocument();
  });

  it('プレイフェーズではヒットとスタンドを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand()],
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stand'));
  });

  // **押せるかどうかはサーバが決める。** 手札から計算し直さない。
  it('canDouble / canSplit が false のときボタンを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand()],
        canDouble: false,
        canSplit: false,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByTestId('da-double')).not.toBeInTheDocument();
    expect(screen.queryByTestId('da-split')).not.toBeInTheDocument();
  });

  it('canDouble / canSplit が true のときボタンを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand({ cards: [card(8), card(8)], score: 16 })],
        canDouble: true,
        canSplit: true,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-double')).toBeInTheDocument());
    expect(screen.getByTestId('da-split')).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('da-split'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('split'));
  });

  it('スプリット後は手札が複数出る', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        anteBet: 50,
        dealerCards: [card(6), card(9)],
        dealerHoleDealt: true,
        hands: [hand(), hand({ score: 19 })],
        activeHand: 1,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-hand-0')).toBeInTheDocument());
    expect(screen.getByTestId('da-hand-1')).toBeInTheDocument();
  });

  it('プレイ中は複数ハンドの操作対象だけをテキストで示し、結果では示さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.PLAY,
        hands: [hand(), hand({ score: 19 })],
        activeHand: 1,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-hand-1')).toBeInTheDocument());

    expect(screen.getByTestId('da-hand-1')).toHaveTextContent('現在操作中');
    expect(screen.getByTestId('da-hand-0')).not.toHaveTextContent('現在操作中');

    cleanup();
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.PLAY, hands: [hand()] }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-hand-0')).toBeInTheDocument());
    expect(screen.getByTestId('da-hand-0')).not.toHaveTextContent('現在操作中');

    cleanup();
    mockApi.mockResolvedValue(
      withState({ phase: DoubleAttackPhase.RESULT, hands: [hand(), hand({ score: 19 })], activeHand: 1 }),
    );
    cleanup();
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-result')).toBeInTheDocument());
    expect(screen.getByTestId('da-hand-1')).toHaveTextContent('手札2');
    expect(screen.getByTestId('da-hand-1')).not.toHaveTextContent('現在操作中');
  });

  it('決着では収支と次へを出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 20,
        dealerCards: [card(6), card(9), card(10)],
        dealerScore: 25,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.win, bet: 100 })],
        payout: 200,
        bustItPayout: 60,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-result')).toBeInTheDocument());
    // 賭け 100 + Bust It 20 = 120 に対し払い戻し 200 -> 収支 +80
    expect(screen.getByTestId('da-result')).toHaveTextContent('80');
    // **賭けたのに結果が見えない状態をなくす** (#5776)。
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('Bust It 的中');
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('60');
    expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('アクションログとリセットのボタンを実行する', async () => {
    mockApi.mockResolvedValue(withState({ phase: DoubleAttackPhase.PLAY, hands: [hand()] }));
    mockActionLogApi.mockResolvedValue({ entries: [] });
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'button.actionLog' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'button.actionLog' }));
    await waitFor(() => expect(mockActionLogApi).toHaveBeenCalled());
    expect(screen.getByTestId('copy-announcer')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  // **払い戻し 0 でも「外れ」と言う。** 何も出ないと、賭けたこと自体が
  // 無かったように見える (#5776)。
  it('Bust It が外れた回もそう言う', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 20,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.lose, bet: 100 })],
        payout: 0,
        bustItPayout: 0,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-bustit-result')).toBeInTheDocument());
    expect(screen.getByTestId('da-bustit-result')).toHaveTextContent('外れ');
  });

  // **負のコントロール: 賭けていない回に行は出さない。**
  it('Bust It に賭けていない回は行を出さない', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: DoubleAttackPhase.RESULT,
        anteBet: 50,
        attackBet: 50,
        bustItBet: 0,
        dealerHoleDealt: true,
        hands: [hand({ result: DOUBLE_ATTACK_RESULT.win, bet: 100 })],
        payout: 200,
        bustItPayout: 0,
      }),
    );
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-result')).toBeInTheDocument());
    expect(screen.queryByTestId('da-bustit-result')).not.toBeInTheDocument();
  });

  it('チップを見出しに出す', async () => {
    mockApi.mockResolvedValue(withState({ chips: 777 }));
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.getByTestId('da-chips')).toHaveTextContent('777'));
  });

  it('CLI モードでは端末を出す', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    } as unknown as ReturnType<typeof useCliMode>);
    mockApi.mockResolvedValue(base);
    renderWithProviders(<DoubleAttackPage />);
    await waitFor(() => expect(screen.queryByTestId('card-area')).not.toBeInTheDocument());
  });
});
