import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const marriageJa = JSON.parse(
  readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), '../src/i18n/locales/ja/marriage.json'), 'utf8'),
) as { discardHistory: { show: string } };

/** Japanese labels for shared and page-rendered auxiliary actions. */
export const AUXILIARY_ACTION_LABELS = [
  'リセット',
  '次のゲーム',
  'リセット確認',
  '投了確認',
  '投了',
  'ギブアップ',
  'ヒント',
  '棋譜を見る',
  '棋譜',
  '設定',
  'チュートリアル',
  'CLI',
  '説明を表示',
  '説明を非表示',
  'ヘルプ',
  'ルール',
  '言語',
  '日本語',
  'English',
  'ゲームナビゲーション',
  'お気に入りに追加',
  marriageJa.discardHistory.show,
] as const;
