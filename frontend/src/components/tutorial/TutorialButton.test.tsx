import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TutorialButton } from './TutorialButton';

const mockStart = vi.fn();
vi.mock('../../providers/TutorialProvider', () => ({
  useTutorialContext: () => ({ start: mockStart }),
}));

vi.mock('../../styles/buttonStyles', () => ({
  btnSecondary: 'btn-secondary',
}));

describe('TutorialButton', () => {
  it('renders a button with ? text', () => {
    render(<TutorialButton />);
    const btn = screen.getByRole('button', { name: 'チュートリアル' });
    expect(btn).toBeInTheDocument();
    expect(btn).toHaveTextContent('?');
  });

  it('calls start on click', () => {
    render(<TutorialButton />);
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    expect(mockStart).toHaveBeenCalledTimes(1);
  });

  it('has correct className', () => {
    render(<TutorialButton />);
    const btn = screen.getByRole('button', { name: 'チュートリアル' });
    expect(btn.className).toContain('btn-secondary');
    expect(btn.className).toContain('text-xs');
  });
});
