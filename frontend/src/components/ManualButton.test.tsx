import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ManualButton } from './ManualButton';

vi.mock('../constants/manualTexts', () => ({
  manualTexts: {
    '/': '# BlackJack\n\nManual content',
    '/poker': '# Poker\n\nPoker content',
  },
}));

describe('ManualButton', () => {
  it('renders a button with Manual label', () => {
    render(<ManualButton gamePath="/" />);
    const btn = screen.getByRole('button', { name: 'Manual' });
    expect(btn).toBeInTheDocument();
  });

  it('opens modal when clicked', () => {
    render(<ManualButton gamePath="/" />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Manual' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('BlackJack')).toBeInTheDocument();
  });

  it('closes modal when close button is clicked', () => {
    render(<ManualButton gamePath="/" />);
    fireEvent.click(screen.getByRole('button', { name: 'Manual' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '×' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('passes correct gamePath to modal', () => {
    render(<ManualButton gamePath="/poker" />);
    fireEvent.click(screen.getByRole('button', { name: 'Manual' }));
    expect(screen.getByText('Poker')).toBeInTheDocument();
  });
});
