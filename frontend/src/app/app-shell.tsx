import { zodResolver } from '@hookform/resolvers/zod';
import {
  ChevronRight,
  DownloadCloud,
  ExternalLink,
  KeyRound,
  Languages,
  LayoutDashboard,
  LogOut,
  Menu,
  Monitor,
  Moon,
  MoreHorizontal,
  Server,
  Settings,
  Sliders,
  Sun,
} from 'lucide-react';
import { useTheme } from 'next-themes';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Link, NavLink, Outlet } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import { useAuth } from '@/shared/auth/use-auth';
import { SiteFooter } from '@/shared/components/site-footer';
import { cn } from '@/shared/lib/cn';

const clashNavigation = [
  { href: '/clash/dashboard', label: '仪表盘', icon: LayoutDashboard },
  { href: '/clash/port-profiles', label: '工作台', icon: Sliders },
  { href: '/clash/source-configs', label: '配置源', icon: DownloadCloud },
  { href: '/clash/nodes', label: '节点池', icon: Server },
  {
    href: '/zashboard/',
    label: 'Zashboard',
    icon: ExternalLink,
    external: true,
  },
] as const;

export function AppShell() {
  const { t, i18n } = useTranslation();
  const { admin, logout, changePassword } = useAuth();
  const { setTheme } = useTheme();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [passwordOpen, setPasswordOpen] = useState(false);

  const passwordSchema = z.object({
    currentPassword: z.string().min(1, t('errors.required')),
    newPassword: z.string().min(8, t('errors.minPassword')),
  });
  type PasswordForm = z.infer<typeof passwordSchema>;
  const passwordForm = useForm<PasswordForm>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { currentPassword: '', newPassword: '' },
  });

  async function submitPassword(values: PasswordForm): Promise<void> {
    try {
      await changePassword(values.currentPassword, values.newPassword);
      toast.success(t('auth.passwordUpdated'));
      passwordForm.reset();
      setPasswordOpen(false);
      await logout();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('errors.generic'));
    }
  }

  function navigationLinks(items: typeof clashNavigation) {
    return items.map(({ href, label, icon: Icon, ...rest }) => {
      const isExternal = (rest as { external?: boolean }).external;
      if (isExternal) {
        return (
          <a
            key={href}
            href={href}
            target='_blank'
            rel='noopener noreferrer'
            onClick={() => setMobileOpen(false)}
            className='group flex h-8 items-center gap-2 rounded-md px-2.5 text-xs font-normal text-muted-foreground transition-colors hover:bg-secondary/55 hover:text-foreground'
          >
            <span className='flex size-5 shrink-0 items-center justify-center'>
              <Icon
                className='size-4 text-muted-foreground'
                strokeWidth={1.8}
              />
            </span>
            {label.startsWith('nav.') ? t(label) : label}
          </a>
        );
      }
      return (
        <NavLink
          key={href}
          to={href}
          onClick={() => setMobileOpen(false)}
          className={({ isActive }) =>
            cn(
              'group flex h-8 items-center gap-2 rounded-md px-2.5 text-xs font-normal text-muted-foreground transition-colors hover:bg-secondary/55 hover:text-foreground',
              isActive && 'bg-secondary/60 text-foreground font-semibold',
            )
          }
        >
          {({ isActive }) => (
            <>
              <span className='flex size-5 shrink-0 items-center justify-center'>
                <Icon
                  className={cn(
                    'size-4 text-muted-foreground',
                    isActive && 'text-foreground',
                  )}
                  fill={isActive ? 'currentColor' : 'none'}
                  fillOpacity={isActive ? 0.14 : 0}
                  strokeWidth={1.8}
                />
              </span>
              {label.startsWith('nav.') ? t(label) : label}
            </>
          )}
        </NavLink>
      );
    });
  }

  const navigationContent = (
    <nav
      className='mt-7 min-h-0 flex-1 overflow-y-auto overscroll-contain pr-2 pb-2 space-y-6'
      aria-label={t('shell.navigation')}
    >
      <div>
        <div className='px-2.5 pb-2 text-xs font-semibold text-foreground uppercase tracking-wider'>
          Clash Meta
        </div>
        <div className='space-y-1'>{navigationLinks(clashNavigation)}</div>
      </div>
    </nav>
  );

  const accountControl = (
    <div className='flex h-9 items-center gap-1 px-2.5'>
      <span className='min-w-0 flex-1 truncate text-xs font-normal capitalize text-muted-foreground'>
        {admin?.username}
      </span>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant='ghost'
            size='icon'
            className='size-7 shrink-0 text-muted-foreground hover:text-foreground'
            aria-label={t('common.actions')}
          >
            <MoreHorizontal />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align='end'
          side='top'
          sideOffset={8}
          className='w-56 p-1.5'
        >
          <DropdownMenuSub>
            <DropdownMenuSubTrigger className='h-8'>
              <Sun />
              {t('shell.appearance')}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              <DropdownMenuItem onClick={() => setTheme('light')}>
                <Sun />
                {t('shell.light')}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('dark')}>
                <Moon />
                {t('shell.dark')}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('system')}>
                <Monitor />
                {t('shell.system')}
              </DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuSub>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger className='h-8'>
              <Languages />
              {t('shell.language')}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              <DropdownMenuItem
                onClick={() => void i18n.changeLanguage('zh-CN')}
              >
                简体中文
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => void i18n.changeLanguage('en')}>
                English
              </DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuSub>
          <DropdownMenuItem
            className='h-8'
            onClick={() => setPasswordOpen(true)}
          >
            <KeyRound />
            {t('auth.changePassword')}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem className='h-8' onClick={() => void logout()}>
            <LogOut />
            {t('auth.signOut')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <NavLink
        to='/settings'
        onClick={() => setMobileOpen(false)}
        className={({ isActive }) =>
          cn(
            'flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-secondary/55 hover:text-foreground',
            isActive && 'bg-secondary/60 text-foreground',
          )
        }
        aria-label={t('nav.settings')}
      >
        <Settings className='size-4' strokeWidth={1.8} />
      </NavLink>
    </div>
  );

  return (
    <div className='min-h-screen bg-background'>
      <aside className='fixed inset-y-0 left-0 z-30 hidden h-screen w-[288px] flex-col overflow-hidden bg-sidebar px-4 py-6 lg:flex'>
        <div className='flex h-7 shrink-0 items-center justify-between px-2.5'>
          <Link
            to='/clash/dashboard'
            className='flex h-7 items-center gap-2 text-base font-semibold text-foreground'
          >
            <span>{t('appName')}</span>
            <ChevronRight className='size-3 text-muted-foreground' />
            <span className='text-xs font-normal text-muted-foreground'>
              Clash Control
            </span>
          </Link>
        </div>
        {navigationContent}
        <div className='relative z-10 mt-4 shrink-0 bg-sidebar pt-4'>
          {accountControl}
        </div>
      </aside>

      <div className='flex min-h-screen flex-col lg:pl-[288px]'>
        <header className='flex h-12 items-center justify-between border-b px-4 lg:hidden'>
          <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
            <SheetTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='size-8'
                aria-label={t('shell.openNavigation')}
              >
                <Menu className='size-4' />
              </Button>
            </SheetTrigger>
            <SheetContent
              side='left'
              className='flex w-72 flex-col gap-0 bg-sidebar px-3 py-4 [&>button]:right-2 [&>button]:top-3.5 [&>button]:flex [&>button]:size-7 [&>button]:items-center [&>button]:justify-center [&>nav]:mt-5 [&>nav]:pr-1'
            >
              <SheetHeader className='h-7 shrink-0 px-2.5 text-left'>
                <SheetTitle className='flex h-7 items-center text-base'>
                  {t('appName')}
                </SheetTitle>
                <SheetDescription className='sr-only'>
                  {t('shell.navigation')}
                </SheetDescription>
              </SheetHeader>
              {navigationContent}
              <div className='relative z-10 mt-3 shrink-0 bg-sidebar pt-3'>
                {accountControl}
              </div>
            </SheetContent>
          </Sheet>
          <span className='text-sm font-semibold'>{t('appName')}</span>
          <span className='w-8' aria-hidden='true' />
        </header>

        <main className='mx-auto w-full max-w-[1280px] flex-1 px-5 py-8 sm:px-8 lg:py-20'>
          <Outlet />
        </main>
        <SiteFooter />
      </div>

      <Dialog open={passwordOpen} onOpenChange={setPasswordOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('auth.changePassword')}</DialogTitle>
            <DialogDescription>{admin?.username}</DialogDescription>
          </DialogHeader>
          <form
            className='space-y-4'
            onSubmit={passwordForm.handleSubmit(submitPassword)}
          >
            <div className='space-y-2'>
              <Label htmlFor='current-password'>
                {t('auth.currentPassword')}
              </Label>
              <Input
                id='current-password'
                type='password'
                autoComplete='current-password'
                {...passwordForm.register('currentPassword')}
              />
              {passwordForm.formState.errors.currentPassword ? (
                <p className='text-xs text-destructive'>
                  {passwordForm.formState.errors.currentPassword.message}
                </p>
              ) : null}
            </div>
            <div className='space-y-2'>
              <Label htmlFor='new-password'>{t('auth.newPassword')}</Label>
              <Input
                id='new-password'
                type='password'
                autoComplete='new-password'
                {...passwordForm.register('newPassword')}
              />
              {passwordForm.formState.errors.newPassword ? (
                <p className='text-xs text-destructive'>
                  {passwordForm.formState.errors.newPassword.message}
                </p>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='secondary'
                size='sm'
                onClick={() => setPasswordOpen(false)}
              >
                {t('common.cancel')}
              </Button>
              <Button
                type='submit'
                size='sm'
                disabled={passwordForm.formState.isSubmitting}
              >
                {t('common.save')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
