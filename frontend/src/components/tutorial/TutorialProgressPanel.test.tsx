import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { TutorialProgressPanel } from './TutorialProgressPanel';

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
    expect(screen.getByText(/232/)).toBeInTheDocument();
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
    expect(links.length).toBe(232);
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
    expect(incompleteMarkers.length).toBe(232);
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
