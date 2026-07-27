import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  listProxyNodes,
  testProxyNodes,
  deleteProxyNode,
  deleteProxyNodesBatch,
  toggleProxyNodesBatch,
  type ProxyNode,
  type TestNodeResult,
} from '@/shared/api/clash';
import {
  Search,
  Zap,
  Trash2,
  Power,
  RefreshCw,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import { ConfirmDialog } from '@/shared/components/confirm-dialog';

export function NodePoolPage() {
  const [nodes, setNodes] = useState<ProxyNode[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(true);

  // Batch selection state
  const [selectedIds, setSelectedIds] = useState<number[]>([]);

  // Testing modal & state
  const [testing, setTesting] = useState(false);
  const [testResults, setTestResults] = useState<TestNodeResult[] | null>(null);

  const fetchNodes = async (ignoreCheck?: unknown) => {
    const isIgnored = () => typeof ignoreCheck === 'function' && ignoreCheck();
    try {
      setLoading(true);
      const res = await listProxyNodes({
        page,
        pageSize: 20,
        keyword: keyword.trim() || undefined,
      });
      if (isIgnored()) return;
      setNodes(res.items);
      setTotal(res.total);
    } catch (err) {
      console.error('Failed to list nodes:', err);
    } finally {
      if (!isIgnored()) setLoading(false);
    }
  };

  useEffect(() => {
    let ignore = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchNodes(() => ignore);
    return () => {
      ignore = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, keyword]);

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(nodes.map((n) => n.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleSelectOne = (id: number, checked: boolean) => {
    if (checked) {
      setSelectedIds((prev) => [...prev, id]);
    } else {
      setSelectedIds((prev) => prev.filter((i) => i !== id));
    }
  };

  const handleTestBatch = async () => {
    const idsToTest =
      selectedIds.length > 0 ? selectedIds : nodes.map((n) => n.id);
    if (idsToTest.length === 0) {
      toast.warning('请选择要测试的节点');
      return;
    }
    try {
      setTesting(true);
      const res = await testProxyNodes(idsToTest);
      setTestResults(res.results);
      await fetchNodes();
      toast.success(`节点测速完成，共完成 ${res.results.length} 个节点测速`);
    } catch (err) {
      toast.error(
        `测试节点失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setTesting(false);
    }
  };

  const handleToggleBatch = async (enabled: boolean) => {
    if (selectedIds.length === 0) return;
    try {
      await toggleProxyNodesBatch(selectedIds, enabled);
      setSelectedIds([]);
      await fetchNodes();
      toast.success(`批量更新 ${selectedIds.length} 个节点状态成功`);
    } catch (err) {
      toast.error(
        `变更节点状态失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    }
  };

  // Delete confirmation modal state
  const [deleteConfirmState, setDeleteConfirmState] = useState<{
    open: boolean;
    type: 'single' | 'batch';
    id?: number;
    deleting?: boolean;
  }>({ open: false, type: 'single' });

  const handleDeleteBatch = () => {
    if (selectedIds.length === 0) return;
    setDeleteConfirmState({ open: true, type: 'batch' });
  };

  const handleDeleteOne = (id: number) => {
    setDeleteConfirmState({ open: true, type: 'single', id });
  };

  const handleConfirmDelete = async () => {
    setDeleteConfirmState((prev) => ({ ...prev, deleting: true }));
    try {
      if (deleteConfirmState.type === 'batch') {
        await deleteProxyNodesBatch(selectedIds);
        setSelectedIds([]);
        toast.success(`批量删除 ${selectedIds.length} 个节点成功`);
      } else if (deleteConfirmState.id) {
        await deleteProxyNode(deleteConfirmState.id);
        toast.success('删除节点成功');
      }
      await fetchNodes();
      setDeleteConfirmState({ open: false, type: 'single' });
    } catch (err) {
      toast.error(
        `删除节点失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setDeleteConfirmState((prev) => ({ ...prev, deleting: false }));
    }
  };

  return (
    <div className='space-y-6 p-6'>
      <div className='flex flex-col md:flex-row md:items-center justify-between gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight text-foreground'>
            节点池管理
          </h1>
          <p className='text-sm text-muted-foreground'>
            搜索、筛选、测试与批量控制代理池资源 (共 {total} 个节点)
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            onClick={handleTestBatch}
            disabled={testing}
            variant='default'
            className='gap-2 bg-amber-600 hover:bg-amber-700'
          >
            <Zap className={`h-4 w-4 ${testing ? 'animate-spin' : ''}`} />{' '}
            {testing ? '测试中...' : '并发测速/延迟'}
          </Button>
          {selectedIds.length > 0 && (
            <>
              <Button
                onClick={() => handleToggleBatch(true)}
                size='sm'
                variant='outline'
                className='gap-1 text-emerald-600'
              >
                <Power className='h-3.5 w-3.5' /> 批量启用
              </Button>
              <Button
                onClick={() => handleToggleBatch(false)}
                size='sm'
                variant='outline'
                className='gap-1 text-slate-500'
              >
                <Power className='h-3.5 w-3.5' /> 批量禁用
              </Button>
              <Button
                onClick={handleDeleteBatch}
                size='sm'
                variant='destructive'
                className='gap-1'
              >
                <Trash2 className='h-3.5 w-3.5' /> 删除 ({selectedIds.length})
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Filter Bar */}
      <div className='flex items-center gap-3 bg-card p-4 rounded-xl border'>
        <div className='relative flex-1'>
          <Search className='absolute left-3 top-2.5 h-4 w-4 text-muted-foreground' />
          <Input
            placeholder='搜索节点名称、服务器地址或类型...'
            value={keyword}
            onChange={(e) => {
              setKeyword(e.target.value);
              setPage(1);
            }}
            className='pl-9'
          />
        </div>
        <Button onClick={fetchNodes} variant='ghost' size='icon'>
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      {/* Node Table */}
      <div className='rounded-xl border bg-card shadow-sm'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-12'>
                <Checkbox
                  checked={
                    nodes.length > 0 && selectedIds.length === nodes.length
                  }
                  onCheckedChange={(c) => handleSelectAll(!!c)}
                />
              </TableHead>
              <TableHead>节点名称</TableHead>
              <TableHead>协议/类型</TableHead>
              <TableHead>服务器地址与端口</TableHead>
              <TableHead>状态/延迟</TableHead>
              <TableHead>来源配置</TableHead>
              <TableHead className='text-right'>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className='text-center py-8 text-muted-foreground'
                >
                  加载节点列表中...
                </TableCell>
              </TableRow>
            ) : nodes.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className='text-center py-8 text-muted-foreground'
                >
                  节点池为空。请先在配置源中导入节点。
                </TableCell>
              </TableRow>
            ) : (
              nodes.map((n) => (
                <TableRow key={n.id}>
                  <TableCell>
                    <Checkbox
                      checked={selectedIds.includes(n.id)}
                      onCheckedChange={(c) => handleSelectOne(n.id, !!c)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className='font-semibold'>{n.name}</div>
                    {!n.enabled && (
                      <Badge
                        variant='outline'
                        className='text-xs text-muted-foreground mt-0.5'
                      >
                        已禁用
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant='secondary'
                      className='uppercase font-mono text-xs'
                    >
                      {n.type}
                    </Badge>
                  </TableCell>
                  <TableCell className='font-mono text-xs'>
                    {n.server}:{n.port}
                  </TableCell>
                  <TableCell>
                    {n.last_test_status === 'success' ? (
                      <span className='font-mono font-bold text-xs text-emerald-600 flex items-center gap-1'>
                        <CheckCircle2 className='h-3.5 w-3.5' />{' '}
                        {n.last_latency_ms} ms
                      </span>
                    ) : n.last_test_status === 'failed' ? (
                      <span
                        className='text-xs text-rose-500 flex items-center gap-1'
                        title={n.last_test_error}
                      >
                        <XCircle className='h-3.5 w-3.5' /> 失败
                      </span>
                    ) : (
                      <span className='text-xs text-muted-foreground font-mono'>
                        未测试
                      </span>
                    )}
                  </TableCell>
                  <TableCell className='text-xs text-muted-foreground'>
                    {n.source_config_name || '默认'}
                  </TableCell>
                  <TableCell className='text-right'>
                    <Button
                      size='sm'
                      variant='ghost'
                      className='text-destructive'
                      onClick={() => handleDeleteOne(n.id)}
                    >
                      <Trash2 className='h-4 w-4' />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Test Results Modal */}
      {testResults && (
        <Dialog open={!!testResults} onOpenChange={() => setTestResults(null)}>
          <DialogContent className='max-w-xl max-h-[75vh] overflow-y-auto'>
            <DialogHeader>
              <DialogTitle>批量测试结果</DialogTitle>
            </DialogHeader>
            <div className='space-y-2 py-2'>
              {testResults.map((r, idx) => (
                <div
                  key={idx}
                  className='flex items-center justify-between p-2.5 rounded-lg border text-sm'
                >
                  <span className='font-semibold font-mono truncate max-w-[240px]'>
                    {r.node_name}
                  </span>
                  {r.success ? (
                    <Badge
                      variant='outline'
                      className='text-emerald-600 border-emerald-500/20 font-mono'
                    >
                      {r.latency_ms} ms
                    </Badge>
                  ) : (
                    <span className='text-xs text-rose-500 flex items-center gap-1'>
                      <AlertCircle className='h-3.5 w-3.5' />{' '}
                      {r.error_message || '请求超时/错误'}
                    </span>
                  )}
                </div>
              ))}
            </div>
            <DialogFooter>
              <Button onClick={() => setTestResults(null)}>关闭</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      <ConfirmDialog
        open={deleteConfirmState.open}
        onOpenChange={(open) =>
          setDeleteConfirmState((prev) => ({ ...prev, open }))
        }
        title={
          deleteConfirmState.type === 'batch' ? '批量删除节点' : '删除节点'
        }
        description={
          deleteConfirmState.type === 'batch'
            ? `确定批量删除选中的 ${selectedIds.length} 个节点？此操作不可撤销。`
            : '确定删除该节点？此操作不可撤销。'
        }
        loading={deleteConfirmState.deleting}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
