import { RefreshCw } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { Button } from '@/components/ui/button';
import { DashboardActivity } from '@/features/example/dashboard-activity';
import {
  DashboardOverview,
  DashboardResources,
} from '@/features/example/dashboard-overview';
import { DashboardProviderDistribution } from '@/features/example/dashboard-provider-distribution';
import { DashboardTopServices } from '@/features/example/dashboard-top-services';
import { DashboardTrend } from '@/features/example/dashboard-trend';
import { getMockDashboard } from '@/features/example/mock-data';
import { PeriodSelector } from '@/shared/components/period-selector';
import { toPeriodValue, type PeriodDays } from '@/shared/lib/period';

export function ExampleDashboardPage() {
  const { t, i18n } = useTranslation();
  const [periodDays, setPeriodDays] = useState<PeriodDays>(30);
  const [tick, setTick] = useState(0);
  const period = toPeriodValue(periodDays);
  const dashboard = useMemo(() => {
    void tick;
    return getMockDashboard(period);
  }, [period, tick]);

  return (
    <div className='space-y-5'>
      <header className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
        <div>
          <h1 className='text-xl font-medium'>
            {t('example.dashboard.title')}
          </h1>
          <p className='mt-1 text-xs text-muted-foreground'>
            {t('example.dashboard.description')}
          </p>
        </div>
        <div className='flex min-w-0 shrink-0 items-center gap-2'>
          <PeriodSelector
            value={periodDays}
            onChange={setPeriodDays}
            ariaLabel={t('example.dashboard.usage')}
          />
          <Button
            type='button'
            variant='secondary'
            size='sm'
            onClick={() => setTick((value) => value + 1)}
          >
            <RefreshCw />
            {t('common.refresh')}
          </Button>
        </div>
      </header>

      <DashboardOverview
        dashboard={dashboard}
        locale={i18n.language}
        loading={false}
      />

      <div className='grid items-stretch gap-2 xl:grid-cols-[minmax(0,3fr)_minmax(360px,2fr)]'>
        <DashboardTrend
          dashboard={dashboard}
          locale={i18n.language}
          loading={false}
        />
        <DashboardProviderDistribution
          dashboard={dashboard}
          locale={i18n.language}
          loading={false}
        />
      </div>

      <div className='grid items-stretch gap-2 xl:grid-cols-[minmax(0,3fr)_minmax(360px,2fr)]'>
        <DashboardTopServices
          dashboard={dashboard}
          locale={i18n.language}
          loading={false}
        />
        <div className='grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-2 xl:h-full'>
          <DashboardActivity
            dashboard={dashboard}
            locale={i18n.language}
            loading={false}
          />
          <DashboardResources
            dashboard={dashboard}
            locale={i18n.language}
            loading={false}
          />
        </div>
      </div>
    </div>
  );
}
