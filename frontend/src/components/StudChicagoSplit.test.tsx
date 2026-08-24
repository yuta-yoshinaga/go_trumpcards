import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SevenCardStudPlayerData, SevenCardStudResult } from '../types/card';
import { StudChicagoSplit } from './StudChicagoSplit';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const players = [
  { id: 0, name: 'あなた', isHuman: true },
  { id: 1, name: 'CPU 1', isHuman: false },
] as unknown as SevenCardStudPlayerData[];

function result(overrides: Partial<SevenCardStudResult> & { playerIdx: number }): SevenCardStudResult {
  return {
    handRank: 0,
    handName: '',
    kickers: '',
    bestHand: [],
    wonAmount: 0,
    mucked: false,
    ...overrides,
  } as SevenCardStudResult;
}

describe('StudChicagoSplit', () => {
  it('renders nothing without results', () => {
    const { container } = renderWithProviders(<StudChicagoSplit results={undefined} players={players} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing when nobody won anything', () => {
    const { container } = renderWithProviders(
      <StudChicagoSplit results={[result({ playerIdx: 0 })]} players={players} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  // **wonAmount は役とスペードの合計。** そのまま「役」として読むとスクープが
  // 二重計上される。
  it('derives the high half from wonAmount minus wonSpade', () => {
    renderWithProviders(
      <StudChicagoSplit
        results={[
          result({ playerIdx: 0, wonAmount: 200 }),
          result({ playerIdx: 1, wonAmount: 200, wonSpade: 200, spadeCard: card('SPADE', 1) }),
        ]}
        players={players}
      />,
    );
    const hi = screen.getByTestId('studchicago-hi-badge');
    expect(hi).toHaveTextContent('あなた');
    expect(hi).toHaveTextContent('200');
    const spade = screen.getByTestId('studchicago-spade-badge');
    expect(spade).toHaveTextContent('CPU 1');
    expect(spade).toHaveTextContent('200');
    // 役側のバッジは 1 つだけ: スペードで勝った席は「役 0」なので出ない。
    expect(screen.getAllByTestId('studchicago-hi-badge')).toHaveLength(1);
  });

  // **どの 1 枚で半分を取ったのかを出す。** 出さないと、ポットが割れた理由が
  // 画面のどこにも現れない。
  it('names the spade that took the half', () => {
    renderWithProviders(
      <StudChicagoSplit
        results={[result({ playerIdx: 0, wonAmount: 150, wonSpade: 150, spadeCard: card('SPADE', 13) })]}
        players={players}
      />,
    );
    expect(screen.getByTestId('studchicago-spade-badge')).toHaveTextContent('♠');
  });

  it('marks a scoop when one seat takes both halves', () => {
    renderWithProviders(
      <StudChicagoSplit
        results={[result({ playerIdx: 0, wonAmount: 400, wonSpade: 200, spadeCard: card('SPADE', 1) })]}
        players={players}
      />,
    );
    const scoop = screen.getByTestId('studchicago-scoop-badge');
    expect(scoop).toHaveTextContent('あなた');
    expect(scoop).toHaveTextContent('400');
  });

  // **誰も伏せ札にスペードを持っていなければ総取り。** バッジが無いことではなく、
  // その一文で伝える。
  it('says the high takes it all when no spade won', () => {
    renderWithProviders(<StudChicagoSplit results={[result({ playerIdx: 0, wonAmount: 400 })]} players={players} />);
    expect(screen.getByTestId('studchicago-hi-takes-all')).toBeInTheDocument();
  });

  it('does not say that when a spade did win', () => {
    renderWithProviders(
      <StudChicagoSplit
        results={[
          result({ playerIdx: 0, wonAmount: 200 }),
          result({ playerIdx: 1, wonAmount: 200, wonSpade: 200, spadeCard: card('SPADE', 9) }),
        ]}
        players={players}
      />,
    );
    expect(screen.queryByTestId('studchicago-hi-takes-all')).not.toBeInTheDocument();
  });

  it('omits the card when the server sent none', () => {
    renderWithProviders(
      <StudChicagoSplit
        results={[result({ playerIdx: 1, wonAmount: 200, wonSpade: 200, spadeCard: null })]}
        players={players}
      />,
    );
    const spade = screen.getByTestId('studchicago-spade-badge');
    expect(spade).toHaveTextContent('CPU 1');
    expect(spade).not.toHaveTextContent('(');
  });
});
