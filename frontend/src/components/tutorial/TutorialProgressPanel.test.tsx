import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { gameRoutes } from '../../constants/gameRoutes';
import { TutorialProgressPanel } from './TutorialProgressPanel';

/**
 * How many games the panel lists.
 *
 * **Derived, not written down.** This used to be the literal 259, so adding a
 * game failed three assertions here with no hint that the count was the point
 * (#4652 is the same shape one layer up). `useTutorialProgress` reads
 * `gameRoutes`, so reading it here keeps the two in step by construction.
 */
const GAME_COUNT = gameRoutes.length;

function renderPanel() {
  return render(
    <MemoryRouter>
      <TutorialProgressPanel />
    </MemoryRouter>,
  );
}

describe('TutorialProgressPanel', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('renders progress summary with 0 completed', () => {
    renderPanel();
    expect(screen.getByText(/0/)).toBeInTheDocument();
    expect(screen.getByText(new RegExp(String(GAME_COUNT)))).toBeInTheDocument();
  });

  it('shows correct completed count', () => {
    localStorage.setItem('tutorial_completed_blackjack', 'true');
    localStorage.setItem('tutorial_completed_poker', 'true');
    localStorage.setItem('tutorial_completed_hearts', 'true');
    renderPanel();
    expect(screen.getByText(/3/)).toBeInTheDocument();
  });

  it('renders game links as icons', () => {
    renderPanel();
    const links = screen.getAllByRole('link');
    expect(links.length).toBe(GAME_COUNT);
  });

  it('shows checkmark for completed games', () => {
    localStorage.setItem('tutorial_completed_blackjack', 'true');
    renderPanel();
    const completedMarkers = screen.getAllByText('✓');
    expect(completedMarkers.length).toBe(1);
  });

  it('shows circle for incomplete games', () => {
    renderPanel();
    const incompleteMarkers = screen.getAllByText('○');
    expect(incompleteMarkers.length).toBe(GAME_COUNT);
  });

  it('renders as details/summary collapsible', () => {
    const { container } = renderPanel();
    expect(container.querySelector('details')).toBeInTheDocument();
  });

  it('renders progress bar', () => {
    renderPanel();
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });
});
