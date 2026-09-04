import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { russianbankApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RussianBankPlayer, RussianBankResponse } from '../types/card';
import { RussianBankPage } from './RussianBankPage';

vi.mock('../api/gameApi', () => ({
  russianbankApi: { exec: vi.fn() },
  actionLogApi: { russianbank: vi.fn() },
}));

const mockExec = vi.mocked(russianbankApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<RussianBankPlayer> = {}): RussianBankPlayer {
  return {
    id: 0,
    isHuman: true,
    reserveCount: 13,
    reserveTop: card('DIAMOND', 7),
    handCount: 39,
    wasteCount: 0,
    wasteTop: undefined,
    stopPoints: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<RussianBankResponse> = {}): RussianBankResponse {
  return {
    phase: 1,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    isHumanTurn: true,
    canCallStop: false,
    canUndo: false,
    moveCount: 0,
    tableau: [[], [], [], []],
    foundations: [[], [], [], [], [], [], [], []],
    // 空の台はどのスートのエースも受ける (design 空文字 = 任意)。
    foundationNext: Array.from({ length: 8 }, () => ({ design: '', value: 1 })),
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false })],
    config: { cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('RussianBankPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<RussianBankPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders both players and the board', async () => {
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.getByTestId('player-1')).toBeInTheDocument();
    expect(screen.getByTestId('tableau-0')).toBeInTheDocument();
    expect(screen.getByTestId('foundation-0')).toBeInTheDocument();
  });

  it('selects a source then sends it to a foundation', async () => {
    renderWithProviders(<RussianBankPage />);
    const reserve = await screen.findByTestId('reserve-0');
    fireEvent.click(reserve);
    const toFnd = await screen.findByTestId('to-foundation');
    mockExec.mockClear();
    fireEvent.click(toFnd);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pf', { zone: 0, fromOpp: false, col: 0 }));
  });

  // **8 つの枠はどれを押しても同じコマンドを送っていた** ── 選択肢があるように
  // 見えて、実際には送り先がスートで決まる単一操作だった (#6473)。
  it('does not make the foundation slots clickable', async () => {
    renderWithProviders(<RussianBankPage />);
    fireEvent.click(await screen.findByTestId('reserve-0'));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('foundation-0'));
    fireEvent.click(screen.getByTestId('foundation-3'));
    await waitFor(() => expect(screen.getByTestId('to-foundation')).toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });

  // **実際に行く台だけを光らせる。**判定の材料はサーバの `foundationNext` で、
  // 規則をページに書き写さない。先頭一致なのはドメインの `rbFoundationFor` と同じ。
  it('rings the foundation the selected card will actually reach', async () => {
    mockExec.mockResolvedValue(
      makeState({
        // 3 番の台だけが ♦7 を受ける。他は別スート / 別ランク。
        foundationNext: [
          { design: 'SPADE', value: 7 },
          { design: 'DIAMOND', value: 3 },
          { design: '', value: 1 },
          { design: 'DIAMOND', value: 7 },
          { design: 'DIAMOND', value: 7 },
          { design: 'HEART', value: 7 },
          { design: 'CLOVER', value: 7 },
          { design: '', value: 1 },
        ],
      }),
    );
    renderWithProviders(<RussianBankPage />);
    // 自リザーブのトップは ♦7。
    fireEvent.click(await screen.findByTestId('reserve-0'));

    await waitFor(() => expect(screen.getByTestId('foundation-3').className).toContain('ring-ds-success'));
    // 同じ条件の 4 番は光らない ── 送り先は最初に受け取れる台。
    expect(screen.getByTestId('foundation-4').className).not.toContain('ring-ds-success');
    // 受け取れない台も光らない。
    expect(screen.getByTestId('foundation-0').className).not.toContain('ring-ds-success');
    expect(screen.getByTestId('foundation-2').className).not.toContain('ring-ds-success');
  });

  // 何も選んでいないうちは、どの台も光らない。
  it('rings nothing until a source is selected', async () => {
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('foundation-0')).toBeInTheDocument());
    for (const i of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`foundation-${i}`).className).not.toContain('ring-ds-success');
    }
  });

  it('labels slots by card + zone and marks the selected source aria-pressed', async () => {
    renderWithProviders(<RussianBankPage />);
    // Reserve top is ♦7 in the human's reserve.
    const reserve = await screen.findByTestId('reserve-0');
    expect(reserve).toHaveAttribute('aria-label', '♦ 7（自リザーブ）');
    expect(reserve).toHaveAttribute('aria-pressed', 'false');
    // An empty tableau column names its zone.
    expect(screen.getByTestId('tableau-0')).toHaveAttribute('aria-label', 'タブロー1（空き）');
    // Selecting the reserve flips aria-pressed and announces the source.
    fireEvent.click(reserve);
    expect(screen.getByTestId('reserve-0')).toHaveAttribute('aria-pressed', 'true');
    const src = screen.getByTestId('rb-selected-source');
    expect(src).toHaveAttribute('role', 'status');
    expect(src).toHaveTextContent('自リザーブ');
  });

  it('renders every card in a multi-card tableau column as a cascade', async () => {
    const column: Card[] = [card('SPADE', 10), card('HEART', 9), card('SPADE', 8), card('HEART', 7)];
    mockExec.mockResolvedValue(makeState({ tableau: [column, [], [], []] }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('tableau-0')).toBeInTheDocument());
    // All four buried cards are rendered (not just the top one).
    for (let ci = 0; ci < column.length; ci++) {
      expect(screen.getByTestId(`tableau-0-card-${ci}`)).toBeInTheDocument();
    }
    // The column aria-label still names the top card as the actionable target.
    expect(screen.getByTestId('tableau-0')).toHaveAttribute('aria-label', expect.stringContaining('タブロー1'));
  });

  it('selects a source then moves it to a tableau column', async () => {
    renderWithProviders(<RussianBankPage />);
    const reserve = await screen.findByTestId('reserve-0');
    fireEvent.click(reserve);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('tableau-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mt', { zone: 0, fromOpp: false, col: 0, toCol: 2 }));
  });

  it('discards to end the turn', async () => {
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('discard-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d'));
  });

  it('shows the stop button only when a violation can be caught', async () => {
    mockExec.mockResolvedValue(makeState({ canCallStop: true }));
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('stop-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('s'));
  });

  it('shows the undo button only when a move can be undone', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('undo-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('u'));
  });

  it('hides action controls when it is the CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ isHumanTurn: false, currentPlayerIdx: 1 }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.queryByTestId('discard-button')).not.toBeInTheDocument();
  });

  it('requests a hint and rings the suggested reserve source and foundation target', async () => {
    renderWithProviders(<RussianBankPage />);
    const hintBtn = await screen.findByTestId('hint-button');
    mockExec.mockResolvedValue(
      makeState({ hint: { zone: 0, fromOpponent: false, col: 0, toFoundation: true, toCol: 0 } }),
    );
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByTestId('reserve-0')).toHaveAttribute('data-hint-source', 'true'));
    // The foundation zone is ringed as the destination.
    expect(screen.getByTestId('foundation-0').closest('[data-hint-foundation]')).toHaveAttribute(
      'data-hint-foundation',
      'true',
    );
  });

  it('rings a suggested tableau source and tableau destination', async () => {
    renderWithProviders(<RussianBankPage />);
    const hintBtn = await screen.findByTestId('hint-button');
    mockExec.mockResolvedValue(
      makeState({ hint: { zone: 2, fromOpponent: false, col: 0, toFoundation: false, toCol: 2 } }),
    );
    fireEvent.click(hintBtn);
    await waitFor(() => expect(screen.getByTestId('tableau-0')).toHaveAttribute('data-hint-source', 'true'));
    expect(screen.getByTestId('tableau-2')).toHaveAttribute('data-hint-dest', 'true');
  });

  it('shows a no-move message and no board rings when there is no hint', async () => {
    renderWithProviders(<RussianBankPage />);
    const hintBtn = await screen.findByTestId('hint-button');
    // Server returns no hint object.
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(hintBtn);
    await waitFor(() => expect(screen.getByTestId('rb-hint-none')).toBeInTheDocument());
    expect(screen.getByTestId('reserve-0')).not.toHaveAttribute('data-hint-source');
  });

  it('shows the game-over label at game end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, isHumanTurn: false, winnerIdx: 0 }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getAllByText('ゲーム終了').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('discard-button')).not.toBeInTheDocument();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（codecov が #4596 で同じ 3 ファイルを未到達と報告した）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.removeItem('hint_enabled_russianbank');
    mockExec.mockResolvedValue(
      makeState({ hint: { zone: 2, fromOpponent: false, col: 3, toFoundation: true, toCol: -1 } }),
    );
    renderWithProviders(<RussianBankPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('shows a single-pile hint when the toggle is already on', async () => {
    localStorage.setItem('hint_enabled_russianbank', 'true');
    mockExec.mockResolvedValue(
      makeState({ hint: { zone: 0, fromOpponent: false, col: 0, toFoundation: false, toCol: 2 } }),
    );
    renderWithProviders(<RussianBankPage />);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
    localStorage.removeItem('hint_enabled_russianbank');
  });

  // **無言でボタンが増えるだけだった (#4817)。**CUI は同じ状態を黄色で明示している。
  it('explains why the stop button appeared', async () => {
    mockExec.mockResolvedValue(makeState({ canCallStop: true }));
    renderWithProviders(<RussianBankPage />);

    expect(await screen.findByTestId('rb-stop-available')).toBeInTheDocument();
    expect(screen.getByTestId('stop-button')).toBeInTheDocument();
  });

  it('shows no stop notice while nothing can be called', async () => {
    mockExec.mockResolvedValue(makeState({ canCallStop: false }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('rb-stop-available')).not.toBeInTheDocument();
  });
});
