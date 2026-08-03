import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { gameRoutes } from '../../constants/gameRoutes';
import { TutorialProgressPanel } from './TutorialProgressPanel';

// **ゲーム数はここに数字で書かない。**書けばゲームを 1 本足すたびに無関係な
// テストが赤くなるだけで、何も守っていない (#4652)。
const TOTAL_GAMES = gameRoutes.length;

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
    expect(screen.getByText(new RegExp(String(TOTAL_GAMES)))).toBeInTheDocument();
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
    expect(links.length).toBe(TOTAL_GAMES);
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
    expect(incompleteMarkers.length).toBe(TOTAL_GAMES);
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
