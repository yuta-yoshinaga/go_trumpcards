import { useCallback, useRef, useState } from 'react';

/** Hook that manages confirm dialog open/close state with callback. */
export function useConfirmDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const callbackRef = useRef<(() => void | Promise<void>) | null>(null);

  const requestConfirm = useCallback((callback: () => void | Promise<void>) => {
    callbackRef.current = callback;
    setIsOpen(true);
  }, []);

  const confirm = useCallback(async () => {
    setIsOpen(false);
    await callbackRef.current?.();
    callbackRef.current = null;
  }, []);

  const cancel = useCallback(() => {
    setIsOpen(false);
    callbackRef.current = null;
  }, []);

  return { isOpen, requestConfirm, confirm, cancel };
}
