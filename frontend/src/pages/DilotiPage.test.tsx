import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { dilotiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeDilotiState } from '../test/stateFactories';
import { DilotiPage } from './DilotiPage';

vi.mock('../api/gameApi', () => ({
  dilotiApi: { exec: vi.fn() },
  actionLogApi: { diloti: vi.fn() },
}));

const mockExec = vi.mocked(dilotiApi.exec);

const playState = makeDilotiState();

/** Selects the human's card at the given hand index. */
// **手札の中だけを見る。** スート字形で画面全体を引くと、取り札ボタンの
// 読み上げ名にも札名が入っている (#7037) ので同じ網に掛かり、添字がずれる。
async function handCards() {
  await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
  const hand = document.querySelector<HTMLElement>('[data-tutorial="diloti-player-hand"]');
  if (hand === null) throw new Error('手札の領域が見つからない');
  return within(hand).getAllByRole('button', { name: /♠|♥|♦|♣/ });
}

async function pickHand(idx: number) {
  fireEvent.click((await handCards())[idx]);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('DilotiPage', () => {
  it('calls reset on mount with the configured target', async () => {
    renderWithProviders(<DilotiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetScore: 61 } }),
    );
  });

  it('shows the round and the stock', async () => {
    renderWithProviders(<DilotiPage />);
    expect(await screen.findByText('局 1（61 点勝負）')).toBeInTheDocument();
    expect(screen.getByTestId('diloti-deck')).toHaveTextContent('36');
  });

  // **場札には番号が要る。** 取る対象はこの番号で指すので、無いと組合せ捕獲も
  // 宣言の捕獲も打てない。
  it('numbers the four table cards', async () => {
    renderWithProviders(<DilotiPage />);
    const table = await screen.findByTestId('diloti-table');
    for (const n of ['0', '1', '2', '3']) {
      expect(table).toHaveTextContent(n);
    }
    expect(table.children).toHaveLength(4);
  });

  // **宣言も番号付きで見せる。** 見えないと取る対象を指せず、グループ宣言か
  // どうかも分からない。
  it('shows the declarations with their index and kind', async () => {
    renderWithProviders(<DilotiPage />);
    const decl = await screen.findByTestId('diloti-declaration-0');
    expect(decl).toHaveTextContent('[0] 5');
    expect(decl).toHaveTextContent('単一');
    expect(decl).not.toHaveTextContent('グループ');
  });

  it('marks a group declaration as unraisable', async () => {
    mockExec.mockResolvedValue(
      makeDilotiState({
        declarations: [
          {
            ownerIdx: 1,
            value: 5,
            groups: [[{ design: 'SPADE', value: 5, color: 'black' }], [{ design: 'HEART', value: 5, color: 'red' }]],
            isGroup: true,
          },
        ],
      }),
    );
    renderWithProviders(<DilotiPage />);
    expect(await screen.findByTestId('diloti-declaration-0')).toHaveTextContent('グループ');
  });

  // **打てる手はサーバが数えたものだけ。** 選ぶまでは候補が出ない。
  it('shows the moves only after a card is picked', async () => {
    renderWithProviders(<DilotiPage />);
    await screen.findByTestId('diloti-table');
    expect(screen.queryByTestId('diloti-move-options')).not.toBeInTheDocument();
    await pickHand(0);
    expect(await screen.findByTestId('diloti-move-options')).toBeInTheDocument();
  });

  it('sends the chosen capture with the card played', async () => {
    renderWithProviders(<DilotiPage />);
    await pickHand(0);
    fireEvent.click(await screen.findByTestId('diloti-take-2-3'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', {
        handIndex: 0,
        action: 'capture',
        tableIndices: [2, 3],
        declIndices: undefined,
      }),
    );
  });

  // **宣言を取る手も同じ 1 回の要求で届く。**
  it('sends a declaration capture', async () => {
    renderWithProviders(<DilotiPage />);
    await pickHand(1);
    fireEvent.click(await screen.findByTestId('diloti-take-d0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', {
        handIndex: 1,
        action: 'capture',
        tableIndices: undefined,
        declIndices: [0],
      }),
    );
  });

  it('sends the declared value', async () => {
    renderWithProviders(<DilotiPage />);
    await pickHand(1);
    fireEvent.click(await screen.findByTestId('diloti-declare-8-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', {
        handIndex: 1,
        action: 'declare',
        tableIndices: [0],
        declIndices: undefined,
        declValue: 8,
      }),
    );
  });

  // **置けない札には「場に置く」を出さない。** 出すと押しても弾かれるだけの
  // ボタンになる ── 手札 2 (♣J) は場に ♦J があるので置けない。
  it('offers lay off only when the card may be laid off', async () => {
    renderWithProviders(<DilotiPage />);
    await pickHand(0);
    expect(await screen.findByTestId('diloti-lay-off')).toBeInTheDocument();

    fireEvent.click((await handCards())[0]); // deselect
    await pickHand(2);
    await screen.findByTestId('diloti-move-options');
    expect(screen.queryByTestId('diloti-lay-off')).not.toBeInTheDocument();
  });

  it('lays the card off', async () => {
    renderWithProviders(<DilotiPage />);
    await pickHand(0);
    fireEvent.click(await screen.findByTestId('diloti-lay-off'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 0, action: 'trail' }));
  });

  it('shows the seats with their xeri count', async () => {
    renderWithProviders(<DilotiPage />);
    const scores = await screen.findByTestId('diloti-scores');
    expect(scores).toHaveTextContent('クセリ 0');
  });

  it('shows the round result and advances', async () => {
    mockExec.mockResolvedValue(
      makeDilotiState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastResult: {
          lines: [
            { key: 'cards', points: [4, 0] },
            { key: 'xeri', points: [10, 0] },
          ],
          totals: [14, 0],
          cardCounts: [30, 22],
          xeris: [1, 0],
        },
      }),
    );
    renderWithProviders(<DilotiPage />);
    const result = await screen.findByTestId('diloti-round-result');
    expect(result).toHaveTextContent('最多枚数');
    expect(result).toHaveTextContent('クセリ');
    expect(result).toHaveTextContent('14 - 0');

    fireEvent.click(screen.getByTestId('diloti-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows the winner at game end', async () => {
    mockExec.mockResolvedValue(makeDilotiState({ phase: 'gameEnd', gameEndFlag: true, winnerIdx: 0 }));
    renderWithProviders(<DilotiPage />);
    expect(await screen.findByTestId('diloti-winner')).toBeInTheDocument();
  });

  // **ヒントのライブ領域は常設。** 出る側と出ない側の両方を見る。
  it('announces a requested hint, and stays quiet otherwise', async () => {
    renderWithProviders(<DilotiPage />);
    const live = await screen.findByTestId('diloti-hint-live');
    expect(live).toBeEmptyDOMElement();

    mockExec.mockResolvedValue(
      makeDilotiState({
        messageCode: 'diloti.hintRequested',
        hintHandIdx: 1,
        hintAction: 'capture',
        hintReason: 'capture',
      }),
    );
    renderWithProviders(<DilotiPage />);
    await waitFor(() => {
      const lives = screen.getAllByTestId('diloti-hint-live');
      expect(lives[lives.length - 1]).not.toBeEmptyDOMElement();
    });
  });

  // **取り札ボタンの読み上げが場札の番号だけだった (#7037)。** どの札を取る手なのか
  // スクリーンリーダの利用者に伝わらない。Cirulla (#6628) と同じ形。
  describe('capture button accessible name', () => {
    it('names the cards, while the visible label keeps the indices', async () => {
      renderWithProviders(<DilotiPage />);
      await pickHand(0);
      const options = await screen.findByTestId('diloti-move-options');
      const takes = within(options)
        .getAllByRole('button')
        .filter((b) => b.getAttribute('data-testid')?.startsWith('diloti-take-'));
      expect(takes.length).toBeGreaterThan(0);

      // 読み上げ名には札名が入る。
      expect(takes[0]).toHaveAttribute('aria-label', expect.stringMatching(/[♠♥♦♣]/));
      expect(takes[0].getAttribute('aria-label')).not.toContain('{{');
      // **見えている文言は番号のまま** ── CUI のコマンドに打ち込む値なので消さない。
      expect(takes[0].textContent).toMatch(/\d/);
      expect(takes[0].textContent).not.toMatch(/[♠♥♦♣]/);
    });
  });
});
