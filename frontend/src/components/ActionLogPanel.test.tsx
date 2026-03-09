import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ActionLogEntry } from '../types/card';
import { ActionLogPanel } from './ActionLogPanel';

describe('ActionLogPanel', () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const sampleEntries: ActionLogEntry[] = [
    { turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'drew a card', cards: [{ design: 'SPADE', value: 1 }] },
    { turnNumber: 2, playerIdx: -1, actionType: 'deal', detail: 'dealt cards' },
    { turnNumber: 3, playerIdx: 1, actionType: 'stand', detail: 'stood', cards: [] },
  ];

  it('renders entries with correct formatting', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);

    // Turn 1: Player 0 with cards
    expect(screen.getByText(/T1 \[Player 0\] hit: drew a card \[SPADE 1\]/)).toBeInTheDocument();
    // Turn 2: System (playerIdx < 0), no cards
    expect(screen.getByText(/T2 \[SYSTEM\] deal: dealt cards/)).toBeInTheDocument();
    // Turn 3: Player 1, empty cards array (no card bracket appended)
    expect(screen.getByText(/T3 \[Player 1\] stand: stood/)).toBeInTheDocument();
  });

  it('shows empty message when entries is empty', () => {
    render(<ActionLogPanel entries={[]} onClose={vi.fn()} />);
    expect(screen.getByText('棋譜はありません。')).toBeInTheDocument();
  });

  it('copy button calls navigator.clipboard.writeText', async () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledTimes(1);
    });
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining('T1 [Player 0] hit: drew a card'));
  });

  it('shows "コピーしました" after copy', async () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'コピーしました' })).toBeInTheDocument();
    });
  });

  it('download button creates and clicks a download link', () => {
    const createObjectURL = vi.fn().mockReturnValue('blob:http://localhost/fake');
    const revokeObjectURL = vi.fn();
    vi.stubGlobal('URL', { createObjectURL, revokeObjectURL });

    const clickSpy = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    const createElementSpy = vi
      .spyOn(document, 'createElement')
      .mockImplementation((tag: string, options?: ElementCreationOptions) => {
        if (tag === 'a') {
          return { href: '', download: '', click: clickSpy } as unknown as HTMLAnchorElement;
        }
        return originalCreateElement(tag, options);
      });

    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const downloadButton = screen.getByRole('button', { name: 'ダウンロード' });
    fireEvent.click(downloadButton);

    expect(createObjectURL).toHaveBeenCalledTimes(1);
    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:http://localhost/fake');

    createElementSpy.mockRestore();
    vi.unstubAllGlobals();
  });

  it('close button calls onClose', () => {
    const onClose = vi.fn();
    render(<ActionLogPanel entries={sampleEntries} onClose={onClose} />);
    const closeButton = screen.getByRole('button', { name: '閉じる' });
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
