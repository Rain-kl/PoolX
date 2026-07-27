import { useTranslation } from 'react-i18next';

export function SiteFooter() {
  const { t } = useTranslation();

  return (
    <footer className='fixed right-0 bottom-0 z-20 flex h-10 w-fit max-w-[calc(100vw-2rem)] items-center justify-end gap-1.5 whitespace-nowrap px-5 text-right text-[11px] text-muted-foreground sm:px-6'>
      <span>{t('appName')}</span>
      <span>© 2026</span>
      <span aria-hidden='true'>·</span>
      <span>{t('shell.gallery')}</span>
    </footer>
  );
}
