import { useCallback, useRef, useState } from 'react';

export function useConfirmDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const callbackRef = useRef<(() => void) | null>(null);

  const requestConfirm = useCallback((callback: () => void) => {
    callbackRef.current = callback;
    setIsOpen(true);
  }, []);

  const confirm = useCallback(() => {
    setIsOpen(false);
    callbackRef.current?.();
    callbackRef.current = null;
  }, []);

  const cancel = useCallback(() => {
    setIsOpen(false);
    callbackRef.current = null;
  }, []);

  return { isOpen, requestConfirm, confirm, cancel };
}
