import { useTranslation } from 'react-i18next';

export function usePhaseNames(
  namespace: string,
  phaseKeyMap: Readonly<Record<number, string>>,
): Record<number, string> {
  const { t } = useTranslation(namespace);
  const result: Record<number, string> = {};
  for (const [phase, key] of Object.entries(phaseKeyMap)) {
    result[Number(phase)] = t(`phase.${key}`);
  }
  return result;
}
