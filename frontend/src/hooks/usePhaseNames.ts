import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';

/** Hook that maps numeric phase constants to translated phase display names. */
export function usePhaseNames(
  namespace: string,
  phaseKeyMap: Readonly<Record<number, string>>,
): Record<number, string> {
  const { t } = useTranslation(namespace);
  return useMemo(() => {
    const result: Record<number, string> = {};
    for (const [phase, key] of Object.entries(phaseKeyMap)) {
      result[Number(phase)] = t(`phase.${key}`);
    }
    return result;
  }, [t, phaseKeyMap]);
}
