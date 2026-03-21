import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { renderHook } from '@testing-library/react';
import * as reactI18next from 'react-i18next';
import * as useActionLogModule from './useActionLog';
import * as useConfirmDialogModule from './useConfirmDialog';
import { useGamePageSetup } from './useGamePageSetup';

describe('useGamePageSetup', () => {
  const mockT = vi.fn((key: string) => key);
  const mockTc = vi.fn((key: string) => key);
  const mockActionLog = { actionLog: null, showActionLog: vi.fn(), hideActionLog: vi.fn() };
  const mockConfirmDialog = { isOpen: false, requestConfirm: vi.fn(), confirm: vi.fn(), cancel: vi.fn() };

  let useTranslationSpy: ReturnType<typeof vi.spyOn>;
  let useActionLogSpy: ReturnType<typeof vi.spyOn>;
  let useConfirmDialogSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.clearAllMocks();
    useTranslationSpy = vi
      .spyOn(reactI18next, 'useTranslation')
      .mockImplementation(
        (ns?: string) =>
          (ns === 'common' ? { t: mockTc } : { t: mockT }) as unknown as ReturnType<typeof reactI18next.useTranslation>,
      );
    useActionLogSpy = vi
      .spyOn(useActionLogModule, 'useActionLog')
      .mockReturnValue(mockActionLog as unknown as ReturnType<typeof useActionLogModule.useActionLog>);
    useConfirmDialogSpy = vi
      .spyOn(useConfirmDialogModule, 'useConfirmDialog')
      .mockReturnValue(mockConfirmDialog as unknown as ReturnType<typeof useConfirmDialogModule.useConfirmDialog>);
  });

  afterEach(() => {
    useTranslationSpy.mockRestore();
    useActionLogSpy.mockRestore();
    useConfirmDialogSpy.mockRestore();
  });

  it('calls useTranslation with game name and common', () => {
    renderHook(() => useGamePageSetup('blackjack'));
    expect(useTranslationSpy).toHaveBeenCalledWith('blackjack');
    expect(useTranslationSpy).toHaveBeenCalledWith('common');
  });

  it('calls useActionLog with game name', () => {
    renderHook(() => useGamePageSetup('poker'));
    expect(useActionLogSpy).toHaveBeenCalledWith('poker');
  });

  it('calls useConfirmDialog', () => {
    renderHook(() => useGamePageSetup('blackjack'));
    expect(useConfirmDialogSpy).toHaveBeenCalled();
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
});
