import { useCallback, useRef, useState } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';

export function useGameApi<TState, TArgs extends unknown[]>(
  apiFn: (...args: TArgs) => Promise<TState>,
  options?: { onSuccess?: (res: TState) => void | Promise<void> },
): {
  state: TState | null;
  setState: React.Dispatch<React.SetStateAction<TState | null>>;
  loading: boolean;
  error: string | null;
  exec: (...args: TArgs) => Promise<void>;
} {
  const [state, setState] = useState<TState | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const apiFnRef = useRef(apiFn);
  apiFnRef.current = apiFn;
  const onSuccessRef = useRef(options?.onSuccess);
  onSuccessRef.current = options?.onSuccess;

  const exec = useCallback(async (...args: TArgs) => {
    setLoading(true);
    try {
      setError(null);
      const res = await apiFnRef.current(...args);
      setState(res);
      await onSuccessRef.current?.(res);
    } catch {
      setError(NETWORK_ERROR_MESSAGE);
    } finally {
      setLoading(false);
    }
  }, []);

  return { state, setState, loading, error, exec };
}
