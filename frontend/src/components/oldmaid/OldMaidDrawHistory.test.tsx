import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DrawHistoryEntry, OldMaidPlayerData } from '../../types/card';
import { OldMaidDrawHistory } from './OldMaidDrawHistory';

const players: OldMaidPlayerData[] = [
  { id: 0, isHuman: true, cardCount: 5, cards: [], displayName: 'YOU' },
  { id: 1, isHuman: false, cardCount: 5, cards: [], displayName: 'CPU1' },
  { id: 2, isHuman: false, cardCount: 5, cards: [], displayName: 'CPU2' },
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
});
