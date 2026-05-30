import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DrawHistoryEntry, OldMaidPlayerData } from '../../types/card';
import { OldMaidDrawHistory, PLAYER_PALETTE } from './OldMaidDrawHistory';

/** sRGB relative luminance of a `#rrggbb` color (WCAG 2.1 formula). */
function relativeLuminance(hex: string): number {
  const channels = [hex.slice(1, 3), hex.slice(3, 5), hex.slice(5, 7)].map((h) => {
    const c = Number.parseInt(h, 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
}

/** Contrast ratio between a `#rrggbb` color and pure white (`#ffffff`). */
function contrastWithWhite(hex: string): number {
  return 1.05 / (relativeLuminance(hex) + 0.05);
}

const players: OldMaidPlayerData[] = [
  { id: 0, isHuman: true, isFinished: false, cardCount: 5, cards: [] },
  { id: 1, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
  { id: 2, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
];

const baseEntry = (overrides: Partial<DrawHistoryEntry> = {}): DrawHistoryEntry => ({
  drawPlayerIdx: 0,
  drawFromIdx: 1,
  discardedPairs: 0,
  drawerFinished: false,
  targetFinished: false,
  ...overrides,
});

describe('OldMaidDrawHistory (graphical timeline)', () => {
  it('renders nothing when entries are empty', () => {
    const { container } = render(<OldMaidDrawHistory entries={[]} players={players} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders one row per entry with drawer + target chips', () => {
    render(
      <OldMaidDrawHistory entries={[baseEntry(), baseEntry({ drawPlayerIdx: 1, drawFromIdx: 2 })]} players={players} />,
    );
    expect(screen.getAllByTestId('draw-history-entry')).toHaveLength(2);
  });

  it('shows a discard burst when discardedPairs > 0', () => {
    render(<OldMaidDrawHistory entries={[baseEntry({ discardedPairs: 1 })]} players={players} />);
    expect(screen.getByTestId('discard-burst').textContent).toBe('💥');
  });

  it('omits the discard burst when no pair was discarded', () => {
    render(<OldMaidDrawHistory entries={[baseEntry({ discardedPairs: 0 })]} players={players} />);
    expect(screen.queryByTestId('discard-burst')).not.toBeInTheDocument();
  });

  it('flags the arrow red when the target is a suspect', () => {
    const suspectPins = new Set([1]);
    render(<OldMaidDrawHistory entries={[baseEntry()]} players={players} suspectPins={suspectPins} />);
    const entry = screen.getByTestId('draw-history-entry');
    expect(entry.dataset.suspectTarget).toBe('true');
    expect(entry.querySelector('.text-ds-error')).not.toBeNull();
  });

  it('does not flag suspect when pins do not include the target', () => {
    const suspectPins = new Set([2]);
    render(
      <OldMaidDrawHistory entries={[baseEntry({ drawFromIdx: 1 })]} players={players} suspectPins={suspectPins} />,
    );
    const entry = screen.getByTestId('draw-history-entry');
    expect(entry.dataset.suspectTarget).toBe('false');
  });

  it('falls back gracefully when suspectPins is undefined', () => {
    render(<OldMaidDrawHistory entries={[baseEntry()]} players={players} />);
    expect(screen.getByTestId('draw-history-entry').dataset.suspectTarget).toBe('false');
  });

  it('shows a finished tag for the drawer when drawerFinished is true', () => {
    render(<OldMaidDrawHistory entries={[baseEntry({ drawerFinished: true })]} players={players} />);
    // i18n: "history.finished" → "[{{name}}上がり]"; ja init in test/setup.
    expect(screen.getByText(/あなた.*上がり/)).toBeInTheDocument();
  });

  it('shows a finished tag for the target when targetFinished is true', () => {
    render(<OldMaidDrawHistory entries={[baseEntry({ targetFinished: true })]} players={players} />);
    expect(screen.getByText(/CPU\s?1.*上がり/)).toBeInTheDocument();
  });

  it('shows both finished tags when drawer and target finished on the same draw', () => {
    render(
      <OldMaidDrawHistory entries={[baseEntry({ drawerFinished: true, targetFinished: true })]} players={players} />,
    );
    expect(screen.getByText(/あなた.*上がり/)).toBeInTheDocument();
    expect(screen.getByText(/CPU\s?1.*上がり/)).toBeInTheDocument();
  });

  it('every chip palette color meets WCAG AA contrast (>=4.5:1) against white text', () => {
    for (const color of PLAYER_PALETTE) {
      expect(contrastWithWhite(color)).toBeGreaterThanOrEqual(4.5);
    }
  });
});
