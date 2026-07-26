import { useCallback, useState } from 'react';

/** Extracts keys of T whose values extend number. */
export type NumberKeys<T> = { [K in keyof T]: T[K] extends number ? K : never }[keyof T];

/** Extracts keys of T whose values extend boolean. */
export type BooleanKeys<T> = { [K in keyof T]: T[K] extends boolean ? K : never }[keyof T];

/** Generic hook for managing game configuration state with number parsing and boolean toggling. */
export function useGameConfig<T extends object>(defaultConfig: T) {
  const [config, setConfig] = useState<T>(defaultConfig);

  const handleConfigChange = useCallback((key: NumberKeys<T>, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

  const handleToggle = useCallback((key: BooleanKeys<T>, value: boolean) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  }, []);

  return { config, setConfig, handleConfigChange, handleToggle } as const;
}
