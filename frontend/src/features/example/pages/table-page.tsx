import { MoreHorizontal, RefreshCw, Search, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { ConfirmDialog } from '@/shared/components/confirm-dialog';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableActionCell,
  TableActionHead,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { MOCK_TABLE_ROWS, type ExampleRow } from '@/features/example/mock-data';
import { EmptyState } from '@/shared/components/data-state';
import { DataTableFilters } from '@/shared/components/data-table-filters';
import { DataTableShell } from '@/shared/components/data-table-shell';
import { PageHeader } from '@/shared/components/page-header';
import { Pagination } from '@/shared/components/pagination';
import { SortableTableHead } from '@/shared/components/sortable-table-head';
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value';
import { formatDateTime, formatNumber } from '@/shared/lib/format';
import {
  nextTableSort,
  type SortOrder,
  type TableSort,
} from '@/shared/lib/table-sort';

export function ExampleTablePage() {
  const { t, i18n } = useTranslation();
  const [rows, setRows] = useState<ExampleRow[]>(MOCK_TABLE_ROWS);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [sort, setSort] = useState<TableSort>({
    field: 'updatedAt',
    order: 'desc',
  });
  const [selected, setSelected] = useState<Set<string>>(() => new Set());
  const debouncedSearch = useDebouncedValue(search);

  const categories = useMemo(
    () =>
      [...new Set(MOCK_TABLE_ROWS.map((row) => row.category))]
        .sort()
        .map((value) => ({ value, label: value })),
    [],
  );

  const filtered = useMemo(() => {
    const query = debouncedSearch.trim().toLowerCase();
    const next = rows.filter((row) => {
      if (statusFilter && row.status !== statusFilter) return false;
      if (categoryFilter && row.category !== categoryFilter) return false;
      if (!query) return true;
      return [row.name, row.owner, row.category, row.id].some((value) =>
        value.toLowerCase().includes(query),
      );
    });
    const orderFactor = sort.order === 'asc' ? 1 : -1;
    next.sort((left, right) => {
      const field = sort.field as keyof ExampleRow;
      const a = left[field];
      const b = right[field];
      if (typeof a === 'number' && typeof b === 'number')
        return (a - b) * orderFactor;
      return String(a).localeCompare(String(b)) * orderFactor;
    });
    return next;
  }, [categoryFilter, debouncedSearch, rows, sort, statusFilter]);

  const total = filtered.length;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const safePage = Math.min(page, pageCount);
  const pageRows = filtered.slice(
    (safePage - 1) * pageSize,
    safePage * pageSize,
  );
  const pageIDs = pageRows.map((row) => row.id);
  const selectedOnPage = pageIDs.filter((id) => selected.has(id));
  const allPageSelected =
    pageIDs.length > 0 && selectedOnPage.length === pageIDs.length;

  function changeSort(field: string, initialOrder: SortOrder): void {
    setSort((current) => nextTableSort(current, field, initialOrder));
  }

  function togglePage(checked: boolean): void {
    setSelected((current) => {
      const next = new Set(current);
      for (const id of pageIDs) {
        if (checked) next.add(id);
        else next.delete(id);
      }
      return next;
    });
  }

  function toggleRow(id: string, checked: boolean): void {
    setSelected((current) => {
      const next = new Set(current);
      if (checked) next.add(id);
      else next.delete(id);
      return next;
    });
  }

  const [deleteTarget, setDeleteTarget] = useState<{ type: 'single' | 'bulk'; id?: string } | null>(null);

  function confirmDelete(): void {
    if (!deleteTarget) return;
    if (deleteTarget.type === 'bulk') {
      setRows((current) => current.filter((row) => !selected.has(row.id)));
      toast.success(t('example.table.deleted', { count: selected.size }));
      setSelected(new Set());
    } else if (deleteTarget.id) {
      const targetId = deleteTarget.id;
      setRows((current) => current.filter((item) => item.id !== targetId));
      setSelected((current) => {
        const next = new Set(current);
        next.delete(targetId);
        return next;
      });
      toast.success(t('example.table.deleted', { count: 1 }));
    }
    setDeleteTarget(null);
  }

  function statusLabel(status: ExampleRow['status']): string {
    return t(`example.table.status.${status}`);
  }

  return (
    <div className='space-y-5'>
      <PageHeader
        title={t('example.table.title')}
        description={t('example.table.description')}
        actions={
          <>
            {selected.size > 0 ? (
              <Button variant='secondary' size='sm' onClick={() => setDeleteTarget({ type: 'bulk' })}>
                <Trash2 />
                {t('example.table.bulkDelete', { count: selected.size })}
              </Button>
            ) : null}
            <Button
              variant='secondary'
              size='sm'
              onClick={() => {
                setRows(MOCK_TABLE_ROWS);
                setSelected(new Set());
                toast.success(t('example.table.reset'));
              }}
            >
              <RefreshCw />
              {t('common.reset')}
            </Button>
          </>
        }
      />

      <DataTableShell
        toolbar={
          <div className='flex w-full flex-wrap items-center gap-2'>
            <div className='relative min-w-0 flex-1 sm:w-64 sm:flex-none'>
              <Search className='pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground' />
              <Input
                className='h-8 pl-9 text-xs'
                value={search}
                onChange={(event) => {
                  setSearch(event.target.value);
                  setPage(1);
                }}
                placeholder={t('example.table.search')}
                aria-label={t('example.table.search')}
              />
            </div>
            <DataTableFilters
              filters={[
                {
                  id: 'status',
                  label: t('example.table.columns.status'),
                  value: statusFilter,
                  onChange: (value) => {
                    setStatusFilter(value);
                    setPage(1);
                  },
                  options: [
                    {
                      value: 'active',
                      label: t('example.table.status.active'),
                    },
                    {
                      value: 'paused',
                      label: t('example.table.status.paused'),
                    },
                    {
                      value: 'archived',
                      label: t('example.table.status.archived'),
                    },
                  ],
                },
                {
                  id: 'category',
                  label: t('example.table.columns.category'),
                  value: categoryFilter,
                  onChange: (value) => {
                    setCategoryFilter(value);
                    setPage(1);
                  },
                  options: categories,
                },
              ]}
            />
          </div>
        }
        footer={
          total > 0 ? (
            <Pagination
              page={safePage}
              pageSize={pageSize}
              total={total}
              onPageChange={setPage}
              onPageSizeChange={(value) => {
                setPageSize(value);
                setPage(1);
              }}
            />
          ) : undefined
        }
      >
        {pageRows.length === 0 ? (
          <EmptyState message={t('example.table.empty')} />
        ) : (
          <Table className='min-w-[960px] table-fixed text-xs'>
            <TableHeader>
              <TableRow className='hover:bg-transparent'>
                <TableHead className='w-10'>
                  <Checkbox
                    checked={allPageSelected}
                    onCheckedChange={(checked) => togglePage(checked === true)}
                    aria-label={t('common.selectPage')}
                  />
                </TableHead>
                <SortableTableHead
                  field='name'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                >
                  {t('example.table.columns.name')}
                </SortableTableHead>
                <SortableTableHead
                  field='status'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                  className='w-28'
                >
                  {t('example.table.columns.status')}
                </SortableTableHead>
                <SortableTableHead
                  field='owner'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                  className='w-36'
                >
                  {t('example.table.columns.owner')}
                </SortableTableHead>
                <SortableTableHead
                  field='category'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                  className='w-28'
                >
                  {t('example.table.columns.category')}
                </SortableTableHead>
                <SortableTableHead
                  field='requests'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                  className='w-28'
                >
                  {t('example.table.columns.requests')}
                </SortableTableHead>
                <SortableTableHead
                  field='updatedAt'
                  sortBy={sort.field}
                  sortOrder={sort.order}
                  onSort={changeSort}
                  className='w-40'
                >
                  {t('example.table.columns.updatedAt')}
                </SortableTableHead>
                <TableActionHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {pageRows.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>
                    <Checkbox
                      checked={selected.has(row.id)}
                      onCheckedChange={(checked) =>
                        toggleRow(row.id, checked === true)
                      }
                      aria-label={t('common.selectItem', { name: row.name })}
                    />
                  </TableCell>
                  <TableCell>
                    <div className='min-w-0'>
                      <p className='truncate font-medium'>{row.name}</p>
                      <p className='mt-0.5 truncate text-[10px] text-muted-foreground'>
                        {row.id}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        row.status === 'active'
                          ? 'default'
                          : row.status === 'paused'
                            ? 'secondary'
                            : 'outline'
                      }
                    >
                      {statusLabel(row.status)}
                    </Badge>
                  </TableCell>
                  <TableCell className='truncate'>{row.owner}</TableCell>
                  <TableCell>{row.category}</TableCell>
                  <TableCell className='tabular-nums'>
                    {formatNumber(row.requests, i18n.language, 0)}
                  </TableCell>
                  <TableCell className='text-muted-foreground'>
                    {formatDateTime(row.updatedAt, i18n.language)}
                  </TableCell>
                  <TableActionCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='size-7'
                          aria-label={t('common.actions')}
                        >
                          <MoreHorizontal />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align='end'>
                        <DropdownMenuItem
                          onClick={() =>
                            toast.message(
                              t('example.table.viewed', { name: row.name }),
                            )
                          }
                        >
                          {t('example.table.view')}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => setDeleteTarget({ type: 'single', id: row.id })}
                        >
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableActionCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </DataTableShell>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title={deleteTarget?.type === 'bulk' ? t('example.table.bulkDelete', { count: selected.size }) : t('common.delete')}
        description={
          deleteTarget?.type === 'bulk'
            ? `确定删除选中的 ${selected.size} 项数据？此操作不可撤销。`
            : '确定删除该条数据？此操作不可撤销。'
        }
        onConfirm={confirmDelete}
      />
    </div>
  );
}
