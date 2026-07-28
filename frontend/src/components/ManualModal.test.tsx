import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ManualModal } from './ManualModal';

vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockResolvedValue({ svg: '<svg>mermaid</svg>' }),
  },
}));

// Manuals are fetched per game now, so these mock the loaders rather than a
// prebuilt map. Every assertion on rendered content therefore has to await.
const WEB_MANUALS: Record<string, string> = {
  '/': '# BlackJack\n\nTest **bold** content\n\n| A | B |\n|---|---|\n| 1 | 2 |',
  '/poker': '# Poker\n\nPoker manual',
  '/mermaid': '# Flow\n\n```mermaid\nflowchart TD\n    A-->B\n```',
  '/code': '# Code\n\n```js\nconsole.log("hello");\n```',
};

vi.mock('../constants/manualTexts', () => ({
  loadManualText: (gamePath: string) => Promise.resolve(WEB_MANUALS[gamePath] ?? ''),
}));

const CUI_MANUALS: Record<string, string> = {
  '/': '# BlackJack CUI\n\nCUI manual text',
  '/poker': '# Poker CUI\n\nCUI poker manual',
};

// isCliModeEnabled stays real — the CLI-mode tests drive it through
// localStorage, which is the behaviour worth exercising.
vi.mock('../constants/cuiManualTexts', async () => {
  const actual = await vi.importActual<typeof import('../constants/cuiManualTexts')>('../constants/cuiManualTexts');
  return {
    ...actual,
    loadCuiManualText: (gamePath: string) => Promise.resolve(CUI_MANUALS[gamePath] ?? ''),
  };
});

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});

describe('ManualModal', () => {
  it('renders nothing when closed', () => {
    render(<ManualModal open={false} onClose={vi.fn()} gamePath="/" />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders markdown content when open', async () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('BlackJack')).toBeInTheDocument();
    expect(screen.getByText('bold')).toBeInTheDocument();
  });

  it('renders GFM table', async () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(await screen.findByRole('table')).toBeInTheDocument();
  });

  it('renders different manual for different gamePath', async () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/poker" />);
    expect(await screen.findByText('Poker')).toBeInTheDocument();
  });

  it('renders CUI manual when CLI mode is enabled for the game', async () => {
    localStorage.setItem('cli-mode-blackjack', 'true');
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(await screen.findByText('BlackJack CUI')).toBeInTheDocument();
    expect(screen.queryByText('BlackJack')).not.toBeInTheDocument();
  });

  it('renders web manual when CLI mode is disabled', async () => {
    localStorage.setItem('cli-mode-blackjack', 'false');
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(await screen.findByText('BlackJack')).toBeInTheDocument();
    expect(screen.queryByText('BlackJack CUI')).not.toBeInTheDocument();
  });

  it('renders CUI manual for /poker when CLI mode is enabled for poker', async () => {
    localStorage.setItem('cli-mode-poker', 'true');
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/poker" />);
    expect(await screen.findByText('Poker CUI')).toBeInTheDocument();
  });

  it('renders web manual when localStorage.getItem throws', async () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('localStorage unavailable');
    });
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    expect(await screen.findByText('BlackJack')).toBeInTheDocument();
    expect(screen.queryByText('BlackJack CUI')).not.toBeInTheDocument();
  });

  it('renders empty content for unknown gamePath', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/unknown" />);
    const dialog = screen.getByRole('dialog');
    expect(dialog).toBeInTheDocument();
  });

  it('renders mermaid diagram', async () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/mermaid" />);
    await waitFor(() => {
      const dialog = screen.getByRole('dialog');
      expect(dialog.querySelector('svg')).toBeTruthy();
    });
  });

  it('has opaque background instead of glass-panel', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const dialog = screen.getByRole('dialog');
    expect(dialog.className).toContain('bg-ds-surface');
    expect(dialog.className).not.toContain('glass-panel');
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('calls onClose when overlay is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    const overlay = screen.getByRole('presentation');
    fireEvent.click(overlay);
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('does not call onClose when dialog content is clicked', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('calls onClose on Escape key', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('has aria-modal and translated aria-label attributes', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-label', 'ゲームマニュアル');
  });

  it('close button has translated aria-label', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const closeBtn = screen.getByRole('button', { name: '閉じる' });
    expect(closeBtn).toHaveAttribute('aria-label', '閉じる');
  });

  it('wraps Tab focus from last to first focusable element', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const closeBtn = screen.getByRole('button', { name: '閉じる' });
    // The only focusable element is the close button, so focus should stay on it
    closeBtn.focus();
    expect(document.activeElement).toBe(closeBtn);
    fireEvent.keyDown(document, { key: 'Tab' });
    // With a single focusable element, last === first, so Tab wraps to itself
    expect(document.activeElement).toBe(closeBtn);
  });

  it('wraps Shift+Tab focus from first to last focusable element', () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/" />);
    const closeBtn = screen.getByRole('button', { name: '閉じる' });
    closeBtn.focus();
    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });
    expect(document.activeElement).toBe(closeBtn);
  });

  it('renders regular code block inside pre without unwrapping', async () => {
    render(<ManualModal open={true} onClose={vi.fn()} gamePath="/code" />);
    const dialog = screen.getByRole('dialog');
    // querySelector has no findBy equivalent, so wait for the loaded manual to
    // render before inspecting the DOM shape.
    await waitFor(() => expect(dialog.querySelector('pre')).toBeInTheDocument());
    expect(dialog.querySelector('code')).toBeInTheDocument();
  });

  it('ignores non-Tab/non-Escape keydown events', () => {
    const onClose = vi.fn();
    render(<ManualModal open={true} onClose={onClose} gamePath="/" />);
    fireEvent.keyDown(document, { key: 'Enter' });
    expect(onClose).not.toHaveBeenCalled();
  });
});
