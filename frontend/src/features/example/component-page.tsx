import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from 'recharts';
import { toast } from 'sonner';

import { ConfirmDialog } from '@/shared/components/confirm-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import { Switch } from '@/components/ui/switch';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import {
  MOCK_GALLERY_ITEMS,
  MOCK_TABLE_ROWS,
} from '@/features/example/mock-data';
import { PageHeader } from '@/shared/components/page-header';
import { cn } from '@/shared/lib/cn';
import { formatNumber } from '@/shared/lib/format';

function Section({
  id,
  title,
  description,
  children,
}: {
  id: string;
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <section
      id={id}
      className='scroll-mt-24 space-y-4 rounded-xl bg-card p-5 sm:p-6'
    >
      <header className='space-y-1'>
        <h2 className='text-sm font-medium'>{title}</h2>
        <p className='text-xs text-muted-foreground'>{description}</p>
      </header>
      {children}
    </section>
  );
}

export function ExampleComponentPage() {
  const { t, i18n } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [enabled, setEnabled] = useState(true);
  const [checked, setChecked] = useState(true);
  const [name, setName] = useState('PoolX gallery');
  const [role, setRole] = useState('designer');
  const [notes, setNotes] = useState('');

  const chartData = useMemo(
    () => [
      { label: 'Mon', requests: 120, tokens: 4200 },
      { label: 'Tue', requests: 168, tokens: 5100 },
      { label: 'Wed', requests: 142, tokens: 4600 },
      { label: 'Thu', requests: 201, tokens: 6200 },
      { label: 'Fri', requests: 188, tokens: 5800 },
      { label: 'Sat', requests: 96, tokens: 2900 },
      { label: 'Sun', requests: 110, tokens: 3300 },
    ],
    [],
  );

  const chartConfig = {
    requests: {
      label: t('example.component.charts.requests'),
      theme: { light: 'oklch(0.68 0.15 245)', dark: 'oklch(0.74 0.13 245)' },
    },
    tokens: {
      label: t('example.component.charts.tokens'),
      theme: { light: 'oklch(0.7 0.11 160)', dark: 'oklch(0.73 0.1 160)' },
    },
  } satisfies ChartConfig;

  const tablePreview = MOCK_TABLE_ROWS.slice(0, 5);
  const sections = [
    { id: 'buttons', label: t('example.component.sections.buttons') },
    { id: 'cards', label: t('example.component.sections.cards') },
    { id: 'table', label: t('example.component.sections.table') },
    { id: 'dialogs', label: t('example.component.sections.dialogs') },
    { id: 'tabs', label: t('example.component.sections.tabs') },
    { id: 'gallery', label: t('example.component.sections.gallery') },
    { id: 'forms', label: t('example.component.sections.forms') },
    { id: 'charts', label: t('example.component.sections.charts') },
  ];

  return (
    <div className='space-y-5'>
      <PageHeader
        title={t('example.component.title')}
        description={t('example.component.description')}
      />

      <nav
        className='flex flex-wrap gap-2'
        aria-label={t('example.component.toc')}
      >
        {sections.map((section) => (
          <a
            key={section.id}
            href={`#${section.id}`}
            className='rounded-full bg-secondary/70 px-3 py-1.5 text-[11px] text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground'
          >
            {section.label}
          </a>
        ))}
      </nav>

      <div className='space-y-4'>
        <Section
          id='buttons'
          title={t('example.component.sections.buttons')}
          description={t('example.component.buttonsDescription')}
        >
          <div className='flex flex-wrap items-center gap-2'>
            <Button size='sm'>{t('example.component.primary')}</Button>
            <Button size='sm' variant='secondary'>
              {t('example.component.secondary')}
            </Button>
            <Button size='sm' variant='outline'>
              {t('example.component.outline')}
            </Button>
            <Button size='sm' variant='ghost'>
              {t('example.component.ghost')}
            </Button>
            <Button size='sm' variant='destructive'>
              {t('example.component.destructive')}
            </Button>
            <Button size='sm' disabled>
              {t('common.loading')}
            </Button>
          </div>
          <div className='flex flex-wrap items-center gap-2'>
            <Badge>{t('example.component.badgeDefault')}</Badge>
            <Badge variant='secondary'>
              {t('example.component.badgeSecondary')}
            </Badge>
            <Badge variant='outline'>
              {t('example.component.badgeOutline')}
            </Badge>
            <Badge variant='destructive'>
              {t('example.component.badgeDestructive')}
            </Badge>
          </div>
        </Section>

        <Section
          id='cards'
          title={t('example.component.sections.cards')}
          description={t('example.component.cardsDescription')}
        >
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3'>
            {[
              {
                title: t('example.component.cardOneTitle'),
                body: t('example.component.cardOneBody'),
              },
              {
                title: t('example.component.cardTwoTitle'),
                body: t('example.component.cardTwoBody'),
              },
              {
                title: t('example.component.cardThreeTitle'),
                body: t('example.component.cardThreeBody'),
              },
            ].map((card) => (
              <article
                key={card.title}
                className='rounded-lg border border-border/70 bg-background/40 p-4'
              >
                <h3 className='text-sm font-medium'>{card.title}</h3>
                <p className='mt-2 text-xs leading-5 text-muted-foreground'>
                  {card.body}
                </p>
              </article>
            ))}
          </div>
        </Section>

        <Section
          id='table'
          title={t('example.component.sections.table')}
          description={t('example.component.tableDescription')}
        >
          <Table className='min-w-[640px] table-fixed text-xs'>
            <TableHeader>
              <TableRow className='hover:bg-transparent'>
                <TableHead>{t('example.table.columns.name')}</TableHead>
                <TableHead className='w-28'>
                  {t('example.table.columns.status')}
                </TableHead>
                <TableHead className='w-32'>
                  {t('example.table.columns.owner')}
                </TableHead>
                <TableHead className='w-28 text-right'>
                  {t('example.table.columns.requests')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tablePreview.map((row) => (
                <TableRow key={row.id}>
                  <TableCell className='font-medium'>{row.name}</TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        row.status === 'active' ? 'default' : 'secondary'
                      }
                    >
                      {t(`example.table.status.${row.status}`)}
                    </Badge>
                  </TableCell>
                  <TableCell>{row.owner}</TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {formatNumber(row.requests, i18n.language, 0)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Section>

        <Section
          id='dialogs'
          title={t('example.component.sections.dialogs')}
          description={t('example.component.dialogsDescription')}
        >
          <div className='flex flex-wrap gap-2'>
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
              <DialogTrigger asChild>
                <Button size='sm' variant='secondary'>
                  {t('example.component.openDialog')}
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>
                    {t('example.component.dialogTitle')}
                  </DialogTitle>
                  <DialogDescription>
                    {t('example.component.dialogDescription')}
                  </DialogDescription>
                </DialogHeader>
                <p className='text-sm text-muted-foreground'>
                  {t('example.component.dialogBody')}
                </p>
                <DialogFooter>
                  <Button
                    size='sm'
                    variant='secondary'
                    onClick={() => setDialogOpen(false)}
                  >
                    {t('common.close')}
                  </Button>
                  <Button
                    size='sm'
                    onClick={() => {
                      setDialogOpen(false);
                      toast.success(t('example.component.dialogSaved'));
                    }}
                  >
                    {t('common.save')}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>

            <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
              <SheetTrigger asChild>
                <Button size='sm' variant='secondary'>
                  {t('example.component.openSheet')}
                </Button>
              </SheetTrigger>
              <SheetContent>
                <SheetHeader>
                  <SheetTitle>{t('example.component.sheetTitle')}</SheetTitle>
                  <SheetDescription>
                    {t('example.component.sheetDescription')}
                  </SheetDescription>
                </SheetHeader>
                <div className='mt-6 space-y-3 text-sm text-muted-foreground'>
                  <p>{t('example.component.sheetBody')}</p>
                  <Button size='sm' onClick={() => setSheetOpen(false)}>
                    {t('common.close')}
                  </Button>
                </div>
              </SheetContent>
            </Sheet>

            <Button
              size='sm'
              variant='outline'
              onClick={() => setConfirmOpen(true)}
            >
              {t('example.component.openAlert')}
            </Button>
            <ConfirmDialog
              open={confirmOpen}
              onOpenChange={setConfirmOpen}
              title={t('example.component.alertTitle')}
              description={t('example.component.alertDescription')}
              confirmText={t('common.confirm', { defaultValue: 'Confirm' })}
              cancelText={t('common.cancel')}
              variant='default'
              onConfirm={() => {
                toast.message(t('example.component.alertConfirmed'));
                setConfirmOpen(false);
              }}
            />
          </div>
        </Section>

        <Section
          id='tabs'
          title={t('example.component.sections.tabs')}
          description={t('example.component.tabsDescription')}
        >
          <Tabs defaultValue='overview'>
            <TabsList>
              <TabsTrigger value='overview'>
                {t('example.component.tabOverview')}
              </TabsTrigger>
              <TabsTrigger value='details'>
                {t('example.component.tabDetails')}
              </TabsTrigger>
              <TabsTrigger value='activity'>
                {t('example.component.tabActivity')}
              </TabsTrigger>
            </TabsList>
            <TabsContent
              value='overview'
              className='mt-4 text-sm text-muted-foreground'
            >
              {t('example.component.tabOverviewBody')}
            </TabsContent>
            <TabsContent
              value='details'
              className='mt-4 text-sm text-muted-foreground'
            >
              {t('example.component.tabDetailsBody')}
            </TabsContent>
            <TabsContent
              value='activity'
              className='mt-4 text-sm text-muted-foreground'
            >
              {t('example.component.tabActivityBody')}
            </TabsContent>
          </Tabs>
        </Section>

        <Section
          id='gallery'
          title={t('example.component.sections.gallery')}
          description={t('example.component.galleryDescription')}
        >
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3'>
            {MOCK_GALLERY_ITEMS.map((item) => (
              <article
                key={item.id}
                className='overflow-hidden rounded-lg border border-border/70'
              >
                <div className={cn('h-28 bg-gradient-to-br', item.accent)} />
                <div className='space-y-1 p-3'>
                  <h3 className='text-sm font-medium'>{item.title}</h3>
                  <p className='text-[11px] text-muted-foreground'>
                    {item.meta}
                  </p>
                </div>
              </article>
            ))}
          </div>
        </Section>

        <Section
          id='forms'
          title={t('example.component.sections.forms')}
          description={t('example.component.formsDescription')}
        >
          <form
            className='grid max-w-xl gap-4'
            onSubmit={(event) => {
              event.preventDefault();
              toast.success(t('example.component.formSaved', { name }));
            }}
          >
            <div className='space-y-2'>
              <Label htmlFor='example-name'>
                {t('example.component.formName')}
              </Label>
              <Input
                id='example-name'
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label>{t('example.component.formRole')}</Label>
              <Select value={role} onValueChange={setRole}>
                <SelectTrigger className='h-9'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='designer'>
                    {t('example.component.roleDesigner')}
                  </SelectItem>
                  <SelectItem value='engineer'>
                    {t('example.component.roleEngineer')}
                  </SelectItem>
                  <SelectItem value='pm'>
                    {t('example.component.rolePm')}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-2'>
              <Label htmlFor='example-notes'>
                {t('example.component.formNotes')}
              </Label>
              <Textarea
                id='example-notes'
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
                placeholder={t('example.component.formNotesPlaceholder')}
              />
            </div>
            <div className='flex flex-wrap items-center gap-6'>
              <label className='flex items-center gap-2 text-xs'>
                <Checkbox
                  checked={checked}
                  onCheckedChange={(value) => setChecked(value === true)}
                />
                {t('example.component.formCheckbox')}
              </label>
              <label className='flex items-center gap-2 text-xs'>
                <Switch checked={enabled} onCheckedChange={setEnabled} />
                {t('example.component.formSwitch')}
              </label>
            </div>
            <div>
              <Button type='submit' size='sm'>
                {t('common.save')}
              </Button>
            </div>
          </form>
        </Section>

        <Section
          id='charts'
          title={t('example.component.sections.charts')}
          description={t('example.component.chartsDescription')}
        >
          <div className='grid gap-4 xl:grid-cols-2'>
            <div className='rounded-lg border border-border/70 p-3'>
              <p className='mb-3 text-xs text-muted-foreground'>
                {t('example.component.charts.requests')}
              </p>
              <ChartContainer
                config={chartConfig}
                className='h-56 w-full aspect-auto'
              >
                <BarChart
                  data={chartData}
                  margin={{ left: 0, right: 8, top: 8, bottom: 0 }}
                >
                  <CartesianGrid vertical={false} strokeDasharray='3 3' />
                  <XAxis
                    dataKey='label'
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                  />
                  <YAxis tickLine={false} axisLine={false} width={36} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Bar
                    dataKey='requests'
                    fill='var(--color-requests)'
                    radius={[4, 4, 0, 0]}
                  />
                </BarChart>
              </ChartContainer>
            </div>
            <div className='rounded-lg border border-border/70 p-3'>
              <p className='mb-3 text-xs text-muted-foreground'>
                {t('example.component.charts.tokens')}
              </p>
              <ChartContainer
                config={chartConfig}
                className='h-56 w-full aspect-auto'
              >
                <AreaChart
                  data={chartData}
                  margin={{ left: 0, right: 8, top: 8, bottom: 0 }}
                >
                  <CartesianGrid vertical={false} strokeDasharray='3 3' />
                  <XAxis
                    dataKey='label'
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                  />
                  <YAxis tickLine={false} axisLine={false} width={40} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Area
                    dataKey='tokens'
                    type='monotone'
                    stroke='var(--color-tokens)'
                    fill='var(--color-tokens)'
                    fillOpacity={0.18}
                  />
                </AreaChart>
              </ChartContainer>
            </div>
          </div>
        </Section>
      </div>
    </div>
  );
}
