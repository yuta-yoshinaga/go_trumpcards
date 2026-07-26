import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ManualButton } from './ManualButton';

// Manuals load per game now, so the mock is a loader rather than a map, and
// assertions on its output must await the resolution.
vi.mock('../constants/manualTexts', () => ({
  loadManualText: (gamePath: string) =>
    Promise.resolve({ '/': '# BlackJack\n\nManual content', '/poker': '# Poker\n\nPoker content' }[gamePath] ?? ''),
}));

describe('ManualButton', () => {
  it('renders a button with translated label', () => {
    render(<ManualButton gamePath="/" />);
    const btn = screen.getByRole('button', { name: 'マニュアル' });
    expect(btn).toBeInTheDocument();
  });

  it('opens modal when clicked', async () => {
    render(<ManualButton gamePath="/" />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'マニュアル' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('BlackJack')).toBeInTheDocument();
  });

  it('closes modal when close button is clicked', async () => {
    render(<ManualButton gamePath="/" />);
    fireEvent.click(screen.getByRole('button', { name: 'マニュアル' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('passes correct gamePath to modal', async () => {
    render(<ManualButton gamePath="/poker" />);
    fireEvent.click(screen.getByRole('button', { name: 'マニュアル' }));
    expect(await screen.findByText('Poker')).toBeInTheDocument();
  });
});
