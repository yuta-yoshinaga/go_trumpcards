import type { TFunction } from 'i18next';

/** Resolves a server message code and its nested translation-key parameters. */
export function resolveMessageCode(
  t: TFunction,
  messageCode: string | undefined,
  messageParams: Record<string, string> | undefined,
  fallback: string | undefined,
): string {
  let displayMessage = fallback ?? '';
  if (messageCode) {
    const params = messageParams
      ? Object.fromEntries(
          Object.entries(messageParams).map(([key, value]) =>
            key.endsWith('Key') ? [key.slice(0, -3), t(value, messageParams)] : [key, value],
          ),
        )
      : {};
    const translated = t(`messageCode.${messageCode}`, params);
    if (translated !== `messageCode.${messageCode}`) {
      displayMessage = translated;
    }
  }
  return displayMessage;
}
