import { useTranslation } from 'react-i18next';
import { isReplaySpeed, type ReplaySpeed, useReplaySpeed } from '../../hooks/useReplaySpeed';
import { type SettingsGroup, SettingsPanel } from './SettingsPanel';

const SPEED_OPTIONS: ReadonlyArray<ReplaySpeed> = ['normal', 'fast', 'instant'];

/**
 * Renders a small settings panel that lets the player tune the CPU replay
 * animation speed (Normal / Fast / Instant). The choice is persisted globally
 * so it carries across every game that uses {@link hooks/gameReplay.runReplay | runReplay}.
 *
 * Designed to sit alongside an existing per-game `<SettingsPanel>` on pages
 * that animate CPU turns (Daifugo, Old Maid, Sevens, President — see #1649).
 */
export function ReplaySpeedSettingsPanel() {
  const { t } = useTranslation('common');
  const [speed, setSpeed] = useReplaySpeed();

  const groups: SettingsGroup[] = [
    {
      items: [
        {
          type: 'select',
          id: 'cpu-replay-speed',
          label: t('settings.replaySpeed.label'),
          tooltip: t('settings.replaySpeed.tooltip'),
          value: speed,
          options: SPEED_OPTIONS.map((value) => ({
            value,
            label: t(`settings.replaySpeed.${value}`),
          })),
          // SettingsPanel's onSelect emits the raw <option> value. The select
          // is rendered exclusively from SPEED_OPTIONS, but `isReplaySpeed`
          // makes the cast type-safe even if a future refactor adds a stray
          // option (and silences the assertion the linter would otherwise need).
          onSelect: (value) => {
            if (isReplaySpeed(value)) setSpeed(value);
          },
        },
      ],
    },
  ];

  return <SettingsPanel title={t('settings.replaySpeed.title')} groups={groups} />;
}
