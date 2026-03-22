import { renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useGamePageSetup } from './useGamePageSetup';

const mockT = vi.fn((key: string) => key);
const mockTc = vi.fn((key: string) => key);

vi.mock('react-i18next', () => ({
  useTranslation: vi.fn((ns: string) => (ns === 'common' ? { t: mockTc } : { t: mockT })),
}));

const mockActionLog = { actionLog: null, showActionLog: vi.fn(), hideActionLog: vi.fn() };
vi.mock('./useActionLog', () => ({
  useActionLog: vi.fn(() => mockActionLog),
}));

const mockConfirmDialog = { isOpen: false, requestConfirm: vi.fn(), confirm: vi.fn(), cancel: vi.fn() };
vi.mock('./useConfirmDialog', () => ({
  useConfirmDialog: vi.fn(() => mockConfirmDialog),
}));

import { useTranslation } from 'react-i18next';
import { useActionLog } from './useActionLog';
import { useConfirmDialog } from './useConfirmDialog';

describe('useGamePageSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls useTranslation with game name and common', () => {
    renderHook(() => useGamePageSetup('blackjack'));
    expect(useTranslation).toHaveBeenCalledWith('blackjack');
    expect(useTranslation).toHaveBeenCalledWith('common');
  });

  it('calls useActionLog with game name', () => {
    renderHook(() => useGamePageSetup('poker'));
    expect(useActionLog).toHaveBeenCalledWith('poker');
  });

  it('calls useConfirmDialog', () => {
    renderHook(() => useGamePageSetup('blackjack'));
    expect(useConfirmDialog).toHaveBeenCalled();
  });

  it('returns t and tc from useTranslation', () => {
    const { result } = renderHook(() => useGamePageSetup('blackjack'));
    expect(result.current.t).toBe(mockT);
    expect(result.current.tc).toBe(mockTc);
  });

  it('returns actionLog fields', () => {
    const { result } = renderHook(() => useGamePageSetup('blackjack'));
    expect(result.current.actionLog).toBe(mockActionLog.actionLog);
    expect(result.current.showActionLog).toBe(mockActionLog.showActionLog);
    expect(result.current.hideActionLog).toBe(mockActionLog.hideActionLog);
  });

  it('returns confirmDialog fields with renamed keys', () => {
    const { result } = renderHook(() => useGamePageSetup('blackjack'));
    expect(result.current.confirmOpen).toBe(mockConfirmDialog.isOpen);
    expect(result.current.requestConfirm).toBe(mockConfirmDialog.requestConfirm);
    expect(result.current.confirmReset).toBe(mockConfirmDialog.confirm);
    expect(result.current.cancelReset).toBe(mockConfirmDialog.cancel);
  });

  it('sets document.title with game name', () => {
    renderHook(() => useGamePageSetup('blackjack'));
    expect(document.title).toBe('nav.blackjack - Trump Cards');
  });

  it('restores document.title on unmount', () => {
    const { unmount } = renderHook(() => useGamePageSetup('blackjack'));
    expect(document.title).toBe('nav.blackjack - Trump Cards');
    unmount();
    expect(document.title).toBe('Trump Cards');
  });
});
