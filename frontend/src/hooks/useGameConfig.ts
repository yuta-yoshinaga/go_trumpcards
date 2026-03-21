import { useCallback, useState } from 'react';

/** Generic hook for managing game configuration state with number parsing and boolean toggling. */
export function useGameConfig<T extends object>(defaultConfig: T) {
  const [config, setConfig] = useState<T>(defaultConfig);

  const handleConfigChange = useCallback((key: keyof T, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

  const handleToggle = useCallback((key: keyof T, value: boolean) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  }, []);

  return { config, setConfig, handleConfigChange, handleToggle } as const;
}
