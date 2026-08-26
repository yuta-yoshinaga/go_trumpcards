import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { perseveranceApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PerseveranceResponse, PerseveranceTableauCard } from '../types/card';
import { PerseverancePage } from './PerseverancePage';

vi.mock('../api/gameApi', () => ({
  perseveranceApi: { exec: vi.fn() },
  actionLogApi: { perseverance: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: ReactNode }) => children,
  useSound: () => mockSoundValue,
  // AnimatedCard AND the central taps (useGameApi / GamePageShell / ErrorAlert)
  // consume useOptionalSound; route it to the same spy and assert on specific
  // sound names so per-card deal sounds don't interfere.
  useOptionalSound: () => mockSoundValue,
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(perseveranceApi.exec);

function makeTableau(cols: PerseveranceTableauCard[][]): PerseveranceTableauCard[][] {
  const result: PerseveranceTableauCard[][] = [];
  // 12 列。クローン元の Baker's Dozen は 13 列だが、Perseverance は A 4 枚を
  // 先に組札へ抜くので 48 枚 = 12 列になる。
  for (let i = 0; i < 12; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: PerseveranceResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 13), faceUp: true },
      { card: card('SPADE', 5), faceUp: true },
    ],
    // **同スートでなければ ♠5 は載らない。**クローン元 (Baker's Dozen) は
    // ランクだけを見るので ♥6 で通っていたが、Perseverance では違法になる。
    [{ card: card('SPADE', 6), faceUp: true }],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
  ]),
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  redealsLeft: 2,
  message: '',
};

const gameClearState: PerseveranceResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'perseverance.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: PerseveranceResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'perseverance.gameOver',
};

describe('PerseverancePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPlaySound.mockReset();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('marks the last card in a column with a dashed warning ring', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    const lastCardBtn = await screen.findByTestId('bd-last-card-1');
    expect(lastCardBtn.className).toContain('ring-dashed');
    // Column 0 has 2 cards so it's NOT the last-card scenario
    expect(screen.queryByTestId('bd-last-card-0')).not.toBeInTheDocument();
  });

  it('persistently marks single-card columns with a dashed warning ring and tooltip', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    // Column 1 holds a single card → flagged before any selection.
    const oneCardCol = await screen.findByTestId('bd-onecard-col-1');
    expect(oneCardCol.className).toContain('ring-dashed');
    // The tooltip lives on the (interactive) card button, not the wrapper div.
    expect(screen.getByTestId('bd-last-card-1')).toHaveAttribute('title', 'この列は二度と使えなくなります');
    // Column 0 (2 cards) and empty columns are not flagged.
    expect(screen.queryByTestId('bd-onecard-col-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bd-onecard-col-2')).not.toBeInTheDocument();
  });

  it('does not flag any card on a fresh deal with multi-card columns only', async () => {
    const fullState: PerseveranceResponse = {
      ...playingState,
      tableau: makeTableau([
        [
          { card: card('SPADE', 13), faceUp: true },
          { card: card('SPADE', 5), faceUp: true },
        ],
        [
          { card: card('HEART', 6), faceUp: true },
          { card: card('HEART', 7), faceUp: true },
        ],
      ]),
    };
    mockExec.mockResolvedValue(fullState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.queryByTestId('bd-last-card-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bd-last-card-1')).not.toBeInTheDocument();
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByText(/パーシビアランス/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getAllByText('A').length).toBeGreaterThanOrEqual(4));
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('advertises the keyboard shortcuts on the control buttons', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    const giveUp = await screen.findByRole('button', { name: 'ギブアップ' });
    expect(giveUp).toHaveAttribute('aria-keyshortcuts', 'g');
    expect(giveUp.querySelector('kbd')?.textContent).toBe('G');
    // The KbdBadge text is aria-hidden, so the hint button's accessible name stays clean.
    const hint = screen.getByRole('button', { name: 'ヒント' });
    expect(hint).toHaveAttribute('aria-keyshortcuts', 'h');
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('renders auto-complete button enabled while playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: PerseveranceResponse = { ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('renders foundation pile top card', async () => {
    const withFoundation: PerseveranceResponse = {
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], [], [], []],
    };
    mockExec.mockResolvedValue(withFoundation);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => {
      // Foundation top card aria-label uses suit + count (見出しは "♠ 組札 2枚")
      expect(screen.getByLabelText(/♠ 組札 2枚/)).toBeInTheDocument();
    });
  });

  it('renders hint banner when hint state is set', async () => {
    let resolveHint: ((res: PerseveranceResponse) => void) | undefined;
    mockExec.mockImplementation((cmd: string) => {
      if (cmd === 'hint') {
        return new Promise<PerseveranceResponse>((resolve) => {
          resolveHint = resolve;
        });
      }
      return Promise.resolve(playingState);
    });

    renderWithProviders(<PerseverancePage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    fireEvent.click(hintBtn);

    // Resolve the in-flight hint call with a hint payload.
    resolveHint?.({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 1 },
    });

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  // #5955: ヒントは無言で現れていた。**空のまま先にマウントしてある**領域の中身が
  // 変わることが読み上げの条件なので、hint がある間だけ現れる内側の div ではなく、
  // 常設のラッパーがライブ領域でなければならない。
  it('announces the hint through a region that was already mounted', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    const region = screen.getByTestId('bd-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('');

    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // **同じ要素**の中身が変わる (別の要素が現れるのではない)。
    await waitFor(() => expect(region).toHaveTextContent(/ヒントがあります/));
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    // Pick the top card of column 0 by its aria-label (cardAlt produces "♠ 5").
    const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  it('makes the tableau horizontally scrollable so 13 columns stay legible', async () => {
    mockExec.mockResolvedValue(playingState);
    const { container } = renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="bd-tableau"]')).toBeInTheDocument());
    expect(container.querySelector('[data-tutorial="bd-tableau"]')).toHaveClass('overflow-x-auto');
  });

  it('centers the selected tableau column in view on mobile', async () => {
    const originalScrollIntoView = Element.prototype.scrollIntoView;
    const originalWidth = window.innerWidth;
    const scrollIntoView = vi.fn();
    // jsdom lacks scrollIntoView; spy on it and force a mobile viewport so the effect runs.
    Element.prototype.scrollIntoView = scrollIntoView;
    window.innerWidth = 375;
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<PerseverancePage />);
      const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
      fireEvent.click(sourceBtn);
      await waitFor(() => expect(scrollIntoView).toHaveBeenCalledWith(expect.objectContaining({ inline: 'center' })));
    } finally {
      Element.prototype.scrollIntoView = originalScrollIntoView;
      window.innerWidth = originalWidth;
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('forwards reset commands to the API on initial load', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(mockExec.mock.calls.some((c) => c[0] === 'reset')).toBe(true));
  });

  it('plays the cardPlace sound when a move succeeds (moveCount advances)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    const source = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(source);
    mockPlaySound.mockClear();
    // The move resolves with an advanced moveCount, signalling a server-confirmed move.
    mockExec.mockResolvedValue({ ...playingState, moveCount: 4 });
    fireEvent.click(screen.getByRole('button', { name: '♠ 6' }));
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('stays silent (no cardPlace) and buzzes when a move fails', async () => {
    mockExec.mockResolvedValueOnce(playingState); // initial reset succeeds
    renderWithProviders(<PerseverancePage />);
    const source = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(source);
    mockPlaySound.mockClear();
    mockExec.mockRejectedValue(new Error('illegal move')); // move rejects → moveCount unchanged
    fireEvent.click(screen.getByRole('button', { name: '♠ 6' }));
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
    expect(mockPlaySound).not.toHaveBeenCalledWith('cardPlace');
  });

  it('plays the shuffle sound when starting a new game via reset', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PerseverancePage />);
    const nextBtn = await screen.findByRole('button', { name: '次のゲーム' });
    mockPlaySound.mockClear();
    fireEvent.click(nextBtn);
    // The central tap plays after the reset exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('shuffle'));
  });
});

// **選択後は押すまで正誤が分からず、クリック→サーバーエラーのループになって
// いた (#4795)。**13列 + 4組札で移動先候補が多い。姉妹の Wasp / Accordion は
// 選択時に合法な移動先をリング表示している。
describe('PerseverancePage legal targets', () => {
  const selectSpadeFive = async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 5' }));
  };

  // ♠5 は ♠6 の列 (1) にだけ乗る。**スートも一致していなければならない。**
  it('rings the column whose top card is one rank higher', async () => {
    await selectSpadeFive();
    await waitFor(() => expect(document.querySelectorAll('[data-legal-target="true"]').length).toBeGreaterThan(0));
  });

  // **空き列は光らせない。**Perseverance も Baker's Dozen も空き列は埋められない。
  it('never rings an empty column', async () => {
    await selectSpadeFive();
    await waitFor(() => expect(document.querySelectorAll('[data-legal-target="true"]').length).toBe(1));
  });

  // **A の唯一の行き先は空の組札。**そこが光らないと「置ける先が無い」と
  // 読めてしまう (#5958)。リングは組札を包む要素側に付いているので、空札の
  // ボタンからは closest で辿る。
  it('rings the empty foundations when an ace is selected', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([[{ card: card('SPADE', 1), faceUp: true }]]),
    });
    renderWithProviders(<PerseverancePage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ A' }));

    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の組札 (♠)' }).closest('[data-legal-target="true"]')).not.toBeNull(),
    );
    // A はどの組札にも置ける。4 つとも光る。
    expect(document.querySelectorAll('[data-legal-target="true"]').length).toBe(4);
  });

  it('rings nothing before a card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    await screen.findByRole('button', { name: '♠ 5' });
    expect(document.querySelectorAll('[data-legal-target="true"]').length).toBe(0);
  });

  // **押せなくはしない。**押せなくすると E2E の「最初の列をクリック」が
  // 別の列を掴む。
  it('leaves the illegal targets clickable', async () => {
    await selectSpadeFive();
    // **合法な移動先を選んではいけない。**♠6 は ♠5 の合法な置き先なので、
    // 「押せなくしない」ことの検証にならない。空の組札は ♠5 では絶対に
    // 合法にならない (A しか置けない) illegal target。
    const emptyFoundation = screen.getByRole('button', { name: '空の組札 (♠)' });
    expect(emptyFoundation.closest('[data-legal-target="true"]')).toBeNull();
    expect(emptyFoundation).toBeEnabled();
  });
});

// 選ぶ前に行き先が見える (#4454)。
describe('PerseverancePage destination preview', () => {
  const render = async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    return screen.findByRole('button', { name: '♠ 5' });
  };
  const targets = () => document.querySelectorAll('[data-legal-target="true"]');
  const previews = () => document.querySelectorAll('[data-preview-target="true"]');

  it('rings the destination while a card is hovered', async () => {
    const spadeFive = await render();
    expect(targets().length).toBe(0);

    fireEvent.mouseEnter(spadeFive);
    await waitFor(() => expect(targets().length).toBeGreaterThan(0));
    // プレビュー中は弱いリング。選択後と見分けが付く。
    expect(previews().length).toBe(targets().length);
    expect(targets()[0]?.className).toContain('ring-ds-success/70');

    fireEvent.mouseLeave(spadeFive);
    await waitFor(() => expect(targets().length).toBe(0));
  });

  it('rings the destination on focus', async () => {
    const spadeFive = await render();
    fireEvent.focus(spadeFive);
    await waitFor(() => expect(previews().length).toBeGreaterThan(0));
    fireEvent.blur(spadeFive);
    await waitFor(() => expect(targets().length).toBe(0));
  });

  // 選択が hover に勝つ ── 選んだあとは実線で、カーソルが動いても消えない。
  it('switches to the solid ring once the card is selected', async () => {
    const spadeFive = await render();
    fireEvent.click(spadeFive);
    await waitFor(() => expect(targets().length).toBeGreaterThan(0));
    expect(previews().length).toBe(0);
    expect(targets()[0]?.className).not.toContain('ring-ds-success/70');
  });
});

// **リディールは Perseverance だけの機能** — クローン元の Baker's Dozen には無い。
// 残り 0 で押せなくなること、そして押せば `redeal` が飛ぶことを見る。
describe('PerseverancePage redeal', () => {
  it('sends redeal and clears the selection', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PerseverancePage />);
    const button = await screen.findByTestId('redeal-button');
    expect(button).toBeEnabled();

    mockExec.mockClear();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('disables the button once both redeals are spent', async () => {
    mockExec.mockResolvedValue({ ...playingState, redealsLeft: 0 });
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('redeal-button')).toBeDisabled());
  });

  it('shows how many redeals remain', async () => {
    mockExec.mockResolvedValue({ ...playingState, redealsLeft: 1 });
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('redeals-left')).toHaveTextContent('1'));
  });

  // 手詰まりのときだけ目を引かせる。**残り 0 なら点滅もしない** — 押せない
  // ボタンを光らせるのは、出口があるという嘘になる。
  it('pulses only while stuck with a redeal still in hand', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, redealsLeft: 1 });
    const { unmount } = renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('redeal-button').className).toContain('animate-pulse'));
    unmount();

    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, redealsLeft: 0 });
    renderWithProviders(<PerseverancePage />);
    await waitFor(() => expect(screen.getByTestId('redeal-button')).toBeDisabled());
    expect(screen.getByTestId('redeal-button').className).not.toContain('animate-pulse');
  });
});

// **並びの一括移動は、盤から掴めなければ存在しないのと同じ。**ドメインは
// cardIndex を受け取れるのに、UI が上札しか選ばせないと看板ルールが死ぬ。
// クローン元の Baker's Dozen は 1 枚ずつなので isTop 判定で正しかった。
describe('PerseverancePage run moves', () => {
  // 列0 は ♥K ♠9 ♠8 — 上2枚が同スート降順の並び。♥K はその下で切れている。
  const runState: PerseveranceResponse = {
    ...playingState,
    tableau: makeTableau([
      [
        { card: card('HEART', 13), faceUp: true },
        { card: card('SPADE', 9), faceUp: true },
        { card: card('SPADE', 8), faceUp: true },
      ],
      [{ card: card('SPADE', 10), faceUp: true }],
    ]),
  };

  it('lets a buried card that starts a run be selected', async () => {
    mockExec.mockResolvedValue(runState);
    renderWithProviders(<PerseverancePage />);
    const runStart = await screen.findByRole('button', { name: '♠ 9' });
    expect(runStart).toBeEnabled();

    fireEvent.click(runStart);
    await waitFor(() => expect(runStart).toHaveAttribute('aria-pressed', 'true'));
  });

  // 負のコントロール: 並びが切れた下の札は掴めないまま。
  it('still refuses a card below the break', async () => {
    mockExec.mockResolvedValue(runState);
    renderWithProviders(<PerseverancePage />);
    const buried = await screen.findByRole('button', { name: '♥ K' });
    expect(buried).toBeDisabled();
  });

  // **サーバに cardIndex が届かなければ、選べても一括移動にはならない。**
  it('sends the run start index, not the top card', async () => {
    mockExec.mockResolvedValue(runState);
    renderWithProviders(<PerseverancePage />);
    fireEvent.click(await screen.findByRole('button', { name: '♠ 9' }));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♠ 10' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 1 },
        expect.objectContaining({ zone: 'tableau', col: 1 }),
      ),
    );
  });
});
