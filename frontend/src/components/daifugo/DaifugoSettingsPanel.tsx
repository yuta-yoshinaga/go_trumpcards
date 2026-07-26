import { useTranslation } from 'react-i18next';
import type { DaifugoConfigInput } from '../../types/card';
import type { SettingsGroup } from '../common/SettingsPanel';
import { SettingsPanel } from '../common/SettingsPanel';

/** Props for {@link DaifugoSettingsPanel}. */
export interface DaifugoSettingsPanelProps {
  config: DaifugoConfigInput;
  onChange: (key: keyof DaifugoConfigInput, value: boolean | number) => void;
}

/** Renders the Daifugo game rule settings panel. */
export function DaifugoSettingsPanel({ config, onChange }: DaifugoSettingsPanelProps) {
  const { t } = useTranslation('daifugo');

  const checkbox = (key: keyof DaifugoConfigInput, label: string) => ({
    type: 'checkbox' as const,
    id: key,
    label,
    checked: config[key] as boolean,
    onToggle: (checked: boolean) => onChange(key, checked),
  });

  const groups: SettingsGroup[] = [
    {
      title: t('settings.groupBasic'),
      items: [
        checkbox('eightCutEnabled', t('settings.eightCut')),
        {
          type: 'select' as const,
          id: 'suit-lock-mode',
          label: t('settings.suitLockMode'),
          value: config.suitLockMode,
          options: [
            { value: 0, label: t('settings.suitLockNone') },
            { value: 1, label: t('settings.suitLockPartial') },
            { value: 2, label: t('settings.suitLockFull') },
          ],
          onSelect: (v: string) => onChange('suitLockMode', Number(v)),
        },
        checkbox('sequenceEnabled', t('settings.sequence')),
        checkbox('cardExchangeEnabled', t('settings.cardExchange')),
      ],
    },
    {
      title: t('settings.groupSpecial'),
      items: [
        checkbox('elevenBackEnabled', t('settings.elevenBack')),
        checkbox('capitalFallEnabled', t('settings.capitalFall')),
        checkbox('fiveSkipEnabled', t('settings.fiveSkip')),
        checkbox('sevenPassEnabled', t('settings.sevenPass')),
        checkbox('tenDiscardEnabled', t('settings.tenDiscard')),
        checkbox('spadeThreeEnabled', t('settings.spadeThree')),
        checkbox('nineReverseEnabled', t('settings.nineReverse')),
        checkbox('coupDetatEnabled', t('settings.coupDetat')),
        checkbox('numberLockEnabled', t('settings.numberLock')),
        checkbox('sandstormEnabled', t('settings.sandstorm')),
        checkbox('emperorEnabled', t('settings.emperor')),
        checkbox('sequenceRevolutionEnabled', t('settings.sequenceRevolution')),
        {
          ...checkbox('sequenceLockEnabled', t('settings.sequenceLock')),
          disabled: !config.sequenceEnabled,
        },
        checkbox('illegalFinishEnabled', t('settings.illegalFinish')),
        checkbox('queenBomberEnabled', t('settings.queenBomber')),
        {
          ...checkbox('blindExchangeEnabled', t('settings.blindExchange')),
          disabled: !config.cardExchangeEnabled,
        },
      ],
    },
    {
      title: t('settings.groupGame'),
      items: [
        {
          type: 'select' as const,
          id: 'joker-count',
          label: t('settings.jokerCount'),
          value: config.jokerCount,
          options: [
            { value: 0, label: '0' },
            { value: 1, label: '1' },
            { value: 2, label: '2' },
          ],
          onSelect: (v: string) => onChange('jokerCount', Number(v)),
        },
        {
          type: 'select' as const,
          id: 'cpu-difficulty',
          label: t('settings.cpuDifficulty'),
          value: config.cpuDifficulty,
          options: [
            { value: 1, label: t('settings.difficultyEasy') },
            { value: 0, label: t('settings.difficultyNormal') },
            { value: 2, label: t('settings.difficultyHard') },
          ],
          onSelect: (v: string) => onChange('cpuDifficulty', Number(v)),
        },
        {
          type: 'select' as const,
          id: 'five-skip-count',
          label: t('settings.fiveSkipCount'),
          value: config.fiveSkipCount,
          options: [1, 2, 3, 4, 5].map((n) => ({ value: n, label: String(n) })),
          onSelect: (v: string) => onChange('fiveSkipCount', Number(v)),
          disabled: !config.fiveSkipEnabled,
        },
      ],
    },
  ];

  return <SettingsPanel title={t('settings.title')} groups={groups} />;
}
