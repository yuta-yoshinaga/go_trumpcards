import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SevenCardStudPlayerData, SevenCardStudResult } from '../types/card';
import { StudHiLoSplit } from './StudHiLoSplit';

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

describe('StudHiLoSplit', () => {
  it('renders nothing without results', () => {
    const { container } = renderWithProviders(<StudHiLoSplit results={undefined} players={players} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing when nobody won anything', () => {
    const { container } = renderWithProviders(<StudHiLoSplit results={[result({ playerIdx: 0 })]} players={players} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('derives the high half from wonAmount minus wonLow', () => {
    // wonAmount はハイとローの合計。これをそのまま「ハイ」として読むと
    // スクープが二重計上される。
    renderWithProviders(
      <StudHiLoSplit
        results={[result({ playerIdx: 0, wonAmount: 201 }), result({ playerIdx: 1, wonAmount: 200, wonLow: 200 })]}
        players={players}
      />,
    );
    expect(screen.getByTestId('studhilo-hi-badge')).toHaveTextContent('201');
    expect(screen.getByTestId('studhilo-lo-badge')).toHaveTextContent('200');
    expect(screen.queryByTestId('studhilo-scoop-badge')).not.toBeInTheDocument();
  });

  it('calls a scoop when one seat took both halves', () => {
    renderWithProviders(
      <StudHiLoSplit results={[result({ playerIdx: 0, wonAmount: 400, wonLow: 200 })]} players={players} />,
    );
    const scoop = screen.getByTestId('studhilo-scoop-badge');
    expect(scoop).toHaveTextContent('400');
    // 合計 400 のうちハイは 200。バッジが 400 と出てはいけない。
    expect(screen.getByTestId('studhilo-hi-badge')).toHaveTextContent('200');
  });

  it('says the high took it all when no low qualified', () => {
    // ここが 8 or Better の肝。ローのバッジが無いことで察しろ、では伝わらない。
    renderWithProviders(<StudHiLoSplit results={[result({ playerIdx: 0, wonAmount: 400 })]} players={players} />);
    expect(screen.getByTestId('studhilo-hi-takes-all')).toBeInTheDocument();
    expect(screen.queryByTestId('studhilo-lo-badge')).not.toBeInTheDocument();
  });

  it('shows the cards that made the low', () => {
    renderWithProviders(
      <StudHiLoSplit
        results={[
          result({
            playerIdx: 1,
            wonAmount: 200,
            wonLow: 200,
            lowBestHand: [card('SPADE', 1), card('HEART', 2), card('CLOVER', 3), card('DIAMOND', 4), card('SPADE', 5)],
          }),
        ]}
        players={players}
      />,
    );
    expect(screen.getByTestId('studhilo-lo-badge')).toHaveTextContent('A 2 3 4 5');
  });
});
