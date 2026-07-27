import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, sultanApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SultanResponse } from '../types/card';
import { SultanPage } from './SultanPage';

vi.mock('../api/gameApi', () => ({
  sultanApi: { exec: vi.fn() },
  actionLogApi: { sultan: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(sultanApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SultanResponse = {
  foundation: [
    [card('SPADE', 13)],
    [card('SPADE', 13)],
    [card('CLOVER', 13)],
    [card('CLOVER', 13)],
    [card('HEART', 13)],
    [card('HEART', 13)],
    [card('DIAMOND', 13)],
    [card('DIAMOND', 13)],
  ],
  divan: [card('CLOVER', 3), null, card('HEART', 5), null, null, null, null, null],
  stockCount: 60,
  waste: [card('CLOVER', 9)],
  redealCount: 0,
  canRedeal: false,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const playingNoWasteState: SultanResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: SultanResponse = {
  ...playingState,
  stockCount: 0,
};

const canRedealState: SultanResponse = {
  ...playingState,
  stockCount: 0,
  canRedeal: true,
};

const gameClearState: SultanResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'sultan.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: SultanResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'sultan.gameOver',
};

const withHintState: SultanResponse = {
  ...playingState,
  hint: { fromZone: 'waste', fromIdx: -1, toFoundation: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('SultanPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SultanPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('gives each empty divan slot a role=img with a numbered aria-label', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SultanPage />);
    // Empty divan slots (idx 1, 3..7) are now announced with their slot number,
    // matching the visible 0-based header.
    await waitFor(() => expect(screen.getByRole('img', { name: '空のディヴァン枠 1' })).toBeInTheDocument());
    expect(screen.getByRole('img', { name: '空のディヴァン枠 7' })).toBeInTheDocument();
    // Slots 0 and 2 hold cards (buttons), so they are not empty-slot images.
    expect(screen.queryByRole('img', { name: '空のディヴァン枠 0' })).not.toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(60\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const images = screen.getAllByRole('img');
    expect(images.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation piles with King bases', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // 8 foundations each pre-seeded with a King → at least 8 card images plus waste/divan.
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(8);
  });

  it('labels each foundation with its position number and suit so they are distinguishable', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Position number + King-base suit disambiguates the two piles per suit.
    expect(screen.getByTestId('sultan-foundation-label-0')).toHaveTextContent('1 ♠');
    expect(screen.getByTestId('sultan-foundation-label-1')).toHaveTextContent('2 ♠');
    expect(screen.getByTestId('sultan-foundation-label-4')).toHaveTextContent('5 ♥');
    expect(screen.getByTestId('sultan-foundation-label-7')).toHaveTextContent('8 ♦');
    // Each foundation cell carries a distinct aria-label (not a uniform "K").
    const labels = [1, 2, 3, 4, 5, 6, 7, 8].map(
      (n) => screen.getByTestId(`sultan-foundation-label-${(n - 1).toString()}`).textContent,
    );
    expect(new Set(labels).size).toBe(8);
  });

  it('exposes a numbered aria-label for each foundation card', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Two spade piles both top a King, but their aria-labels differ by position.
    expect(screen.getByRole('img', { name: '組札1 一番上 ♠ K' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: '組札2 一番上 ♠ K' })).toBeInTheDocument();
  });

  it('renders empty foundation placeholder with position number and K build hint', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[], [], [], [], [], [], [], []],
    });
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // "K" build hint is retained, and each empty slot is announced by position.
    expect(screen.getAllByText('K').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByRole('img', { name: '組札1 空（Kから積む）' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: '組札8 空（Kから積む）' })).toBeInTheDocument();
  });

  it('renders divan slots, empty ones show placeholder', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // 6 of the 8 divan slots are null → empty placeholders present.
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    fireEvent.click(drawBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('redeal button shown and dispatches redeal when canRedeal', async () => {
    mockExec.mockResolvedValue(canRedealState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リディール/ })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(canRedealState);
    fireEvent.click(screen.getByRole('button', { name: /リディール/ }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('redeal button hidden when canRedeal is false', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /リディール/ })).not.toBeInTheDocument();
  });

  it('redeal button shows the remaining redeal count', async () => {
    // redealCount 0 of a max of 2 → 2 redeals remaining.
    mockExec.mockResolvedValue(canRedealState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('sultan-redeal-count')).toHaveTextContent('リディール（残り2回）'));
  });

  it('redeal button count reflects one redeal already used', async () => {
    // redealCount 1 of a max of 2 → 1 redeal remaining (matches CUI display).
    mockExec.mockResolvedValue({ ...canRedealState, redealCount: 1 });
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('sultan-redeal-count')).toHaveTextContent('リディール（残り1回）'));
  });

  it('clicking waste card dispatches move from waste', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const wasteImg = screen.getByAltText('♣ 9');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.objectContaining({ zone: 'waste' })));
  });

  it('clicking a divan card dispatches move from divan', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const divanImg = screen.getByAltText('♣ 3');
    const divanButton = divanImg.closest('button') as HTMLButtonElement;
    fireEvent.click(divanButton);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', expect.objectContaining({ zone: 'divan', divanIdx: 0 })),
    );
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete when ready', async () => {
    const readyState: SultanResponse = { ...playingState, stockCount: 0, waste: [] };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('auto complete button is disabled while stock or waste has cards', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('clicking give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking reset button dispatches reset', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows hint text after clicking hint', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.sultan);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→foundation' }],
    });

    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'カードを移動してください',
      messageCode: 'sultan.playing',
    });
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getAllByText('カードを移動してください').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('draw button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    expect(drawBtn).toBeDisabled();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('SultanPage keyboard shortcuts', () => {
  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SultanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    expect(mockExec).not.toHaveBeenCalled();
  });
});
