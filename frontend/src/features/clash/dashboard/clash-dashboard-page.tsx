import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import {
  getRuntimeStatus,
  listProxyNodes,
  listPortProfiles,
  startKernel,
  stopKernel,
  reloadKernel,
  updatePortProfile,
  getActiveRuntimeConfig,
  getKernelLogs,
  downloadKernelBinary,
  uploadKernelBinary,
  inspectKernelBinary,
  type RuntimeStatusView,
  type ProxyNode,
  type PortProfileWithNodes,
  type RenderResult,
  type InstalledKernelBinary,
} from '@/shared/api/clash';
import {
  Activity,
  Server,
  Cpu,
  Play,
  Square,
  RefreshCw,
  Radio,
  CheckCircle2,
  XCircle,
  Eye,
  Copy,
  Terminal,
  Search,
  Download,
  Upload,
  Info,
} from 'lucide-react';

export function ClashDashboardPage() {
  const [status, setStatus] = useState<RuntimeStatusView | null>(null);
  const [nodes, setNodes] = useState<ProxyNode[]>([]);
  const [profiles, setProfiles] = useState<PortProfileWithNodes[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  // Live Logs State
  const [logs, setLogs] = useState<string[]>([]);
  const [logFilter, setLogFilter] = useState('');

  // Active Config Drawer State
  const [activeConfig, setActiveConfig] = useState<RenderResult | null>(null);
  const [configDrawerOpen, setConfigDrawerOpen] = useState(false);

  // Kernel Management Modal State
  const [kernelModalOpen, setKernelModalOpen] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [installedInfo, setInstalledInfo] =
    useState<InstalledKernelBinary | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchDashboardData = async (ignoreCheck?: unknown) => {
    const isIgnored = () => typeof ignoreCheck === 'function' && ignoreCheck();
    try {
      setLoading(true);
      const [stRes, nodesRes, profRes, logRes] = await Promise.all([
        getRuntimeStatus(),
        listProxyNodes({ pageSize: 100 }),
        listPortProfiles(),
        getKernelLogs(),
      ]);
      if (isIgnored()) return;
      setStatus(stRes);
      setNodes(nodesRes.items);
      setProfiles(profRes.items || profRes);
      setLogs(logRes.logs || []);
    } catch (err) {
      console.error('Failed to load dashboard data:', err);
    } finally {
      if (!isIgnored()) setLoading(false);
    }
  };

  useEffect(() => {
    let ignore = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchDashboardData(() => ignore);

    const interval = setInterval(async () => {
      try {
        const [stRes, logRes] = await Promise.all([
          getRuntimeStatus(),
          getKernelLogs(),
        ]);
        if (ignore) return;
        setStatus(stRes);
        setLogs(logRes.logs || []);
      } catch {
        // silent loop update
      }
    }, 3000);
    return () => {
      ignore = true;
      clearInterval(interval);
    };
  }, []);

  const handleStart = async () => {
    try {
      setActionLoading(true);
      await startKernel({});
      await fetchDashboardData();
      toast.success('Mihomo 进程启动成功');
    } catch (err) {
      toast.error(
        `启动失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleStop = async () => {
    try {
      setActionLoading(true);
      await stopKernel();
      await fetchDashboardData();
      toast.success('Mihomo 进程已停止');
    } catch (err) {
      toast.error(
        `停止失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleReload = async () => {
    try {
      setActionLoading(true);
      await reloadKernel({});
      await fetchDashboardData();
      toast.success('Mihomo 配置重载成功');
    } catch (err) {
      toast.error(
        `重载失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleViewConfig = async () => {
    try {
      setActionLoading(true);
      const res = await getActiveRuntimeConfig();
      setActiveConfig(res);
      setConfigDrawerOpen(true);
    } catch (err) {
      toast.error(
        `获取运行配置失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleCopyConfig = () => {
    if (!activeConfig?.content) return;
    navigator.clipboard.writeText(activeConfig.content);
    toast.success('配置文件内容已复制到剪贴板');
  };

  const handleToggleProfile = async (
    item: PortProfileWithNodes,
    enabled: boolean,
  ) => {
    try {
      await updatePortProfile(item.profile.id, {
        name: item.profile.name,
        listen_host: item.profile.listen_host,
        mixed_port: item.profile.mixed_port,
        socks_port: item.profile.socks_port,
        http_port: item.profile.http_port,
        proxy_settings: item.profile.proxy_settings,
        include_in_runtime: enabled,
        node_ids: item.node_ids,
      });
      toast.success(
        enabled
          ? `已启用配置 "${item.profile.name}"`
          : `已禁用配置 "${item.profile.name}"`,
      );
      await fetchDashboardData();
    } catch (err) {
      toast.error(
        `修改配置状态失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    }
  };

  const handleInspect = async () => {
    try {
      setInstalling(true);
      const res = await inspectKernelBinary();
      setInstalledInfo(res);
      toast.success(`内核版本校验成功！版本: ${res.detected_version}`);
    } catch (err) {
      toast.error(
        `内核检查失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setInstalling(false);
    }
  };

  const handleAutoDownload = async () => {
    try {
      setInstalling(true);
      const res = await downloadKernelBinary();
      setInstalledInfo(res);
      await fetchDashboardData();
      toast.success(
        `成功下载并安装最新版 Mihomo 内核 (版本: ${res.detected_version})`,
      );
    } catch (err) {
      toast.error(
        `自动下载内核失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setInstalling(false);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      setInstalling(true);
      const res = await uploadKernelBinary(file);
      setInstalledInfo(res);
      await fetchDashboardData();
      toast.success(`手动上传 Mihomo 内核成功 (版本: ${res.detected_version})`);
    } catch (err) {
      toast.error(
        `上传内核失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setInstalling(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const isRunning =
    status?.is_process_alive || status?.kernel_instance?.status === 'running';
  const successNodesCount = nodes.filter(
    (n) => n.last_test_status === 'success',
  ).length;
  const filteredLogs = logs.filter((l) =>
    !logFilter ? true : l.toLowerCase().includes(logFilter.toLowerCase()),
  );

  return (
    <div className='space-y-6 p-6'>
      {/* Header */}
      <div className='flex flex-col md:flex-row md:items-center justify-between gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight text-foreground'>
            Clash Meta 控制台仪表盘
          </h1>
          <p className='text-sm text-muted-foreground'>
            实时监控内核进程状态、生命周期控制、动态代理分配与健康日志
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            onClick={() => setKernelModalOpen(true)}
            variant='outline'
            className='gap-2 text-xs'
          >
            <Cpu className='h-4 w-4 text-primary' /> 内核
          </Button>
          <Button
            onClick={handleViewConfig}
            disabled={actionLoading}
            variant='outline'
            className='gap-2 text-xs'
          >
            <Eye className='h-4 w-4' /> 查看
          </Button>
          {!isRunning ? (
            <Button
              onClick={handleStart}
              disabled={actionLoading}
              className='gap-2 bg-emerald-600 hover:bg-emerald-700'
            >
              <Play className='h-4 w-4' /> 启动内核
            </Button>
          ) : (
            <>
              <Button
                onClick={handleReload}
                disabled={actionLoading}
                variant='outline'
                className='gap-2'
              >
                <RefreshCw
                  className={`h-4 w-4 ${actionLoading ? 'animate-spin' : ''}`}
                />{' '}
                重载
              </Button>
              <Button
                onClick={handleStop}
                disabled={actionLoading}
                variant='destructive'
                className='gap-2'
              >
                <Square className='h-4 w-4' /> 停止内核
              </Button>
            </>
          )}
          <Button onClick={fetchDashboardData} variant='ghost' size='icon'>
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      {status?.kernel_instance?.last_error && (
        <div className='rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-destructive space-y-1 shadow-sm'>
          <div className='flex items-center gap-2 font-bold text-sm'>
            <XCircle className='h-5 w-5 shrink-0' />
            <span>内核报错信息 / 启动失败原因</span>
          </div>
          <p className='text-xs font-mono break-all pl-7 text-destructive/90'>
            {status.kernel_instance.last_error}
          </p>
        </div>
      )}

      {/* Metrics Grid */}
      <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4'>
        {/* Core Status Card */}
        <div className='rounded-xl border bg-card p-5 shadow-sm transition-all hover:shadow-md'>
          <div className='flex items-center justify-between text-muted-foreground'>
            <span className='text-xs font-semibold uppercase'>
              内核进程状态
            </span>
            <Cpu className='h-5 w-5 text-primary' />
          </div>
          <div className='mt-4 flex items-center justify-between'>
            <div className='text-xl font-bold'>
              {isRunning ? (
                <span className='text-emerald-500 flex items-center gap-1.5'>
                  <CheckCircle2 className='h-5 w-5' /> 运行中
                </span>
              ) : (
                <span className='text-slate-400 flex items-center gap-1.5'>
                  <XCircle className='h-5 w-5' /> 已停止
                </span>
              )}
            </div>
            {status?.kernel_instance?.pid && (
              <Badge variant='secondary' className='font-mono text-xs'>
                PID: {status.kernel_instance.pid}
              </Badge>
            )}
          </div>
          <p className='mt-2 text-xs text-muted-foreground'>
            版本: {status?.mihomo_version || 'Mihomo Kernel'}
          </p>
        </div>

        {/* Listen Ports Card */}
        <div className='rounded-xl border bg-card p-5 shadow-sm transition-all hover:shadow-md'>
          <div className='flex items-center justify-between text-muted-foreground'>
            <span className='text-xs font-semibold uppercase'>
              活动监听端口
            </span>
            <Radio className='h-5 w-5 text-indigo-500' />
          </div>
          <div className='mt-4 text-3xl font-bold'>
            {status?.active_profiles?.length ||
              profiles.filter((p) => p.profile.include_in_runtime).length}
          </div>
          <p className='mt-2 text-xs text-muted-foreground'>
            包含 {profiles.length} 个端口工作台配置
          </p>
        </div>

        {/* Node Count Card */}
        <div className='rounded-xl border bg-card p-5 shadow-sm transition-all hover:shadow-md'>
          <div className='flex items-center justify-between text-muted-foreground'>
            <span className='text-xs font-semibold uppercase'>
              节点池节点数
            </span>
            <Server className='h-5 w-5 text-blue-500' />
          </div>
          <div className='mt-4 text-3xl font-bold'>{nodes.length}</div>
          <p className='mt-2 text-xs text-muted-foreground'>
            测试通过:{' '}
            <span className='text-emerald-500 font-semibold'>
              {successNodesCount}
            </span>{' '}
            个
          </p>
        </div>

        {/* Controller Health Card */}
        <div className='rounded-xl border bg-card p-5 shadow-sm transition-all hover:shadow-md'>
          <div className='flex items-center justify-between text-muted-foreground'>
            <span className='text-xs font-semibold uppercase'>
              API 控制连通性
            </span>
            <Activity className='h-5 w-5 text-violet-500' />
          </div>
          <div className='mt-4 text-xl font-bold'>
            {status?.controller_ok ? (
              <span className='text-emerald-500 flex items-center gap-1.5'>
                <CheckCircle2 className='h-5 w-5' /> 可连接
              </span>
            ) : (
              <span className='text-amber-500 flex items-center gap-1.5'>
                <XCircle className='h-5 w-5' /> 未连接
              </span>
            )}
          </div>
          <p className='mt-2 text-xs text-muted-foreground truncate'>
            {status?.controller_ok
              ? `${status?.kernel_instance?.controller_address || '127.0.0.1:9090'} 通讯正常`
              : !isRunning
                ? '内核未启动 (启动后自动连通)'
                : '端口 127.0.0.1:9090 无法响应'}
          </p>
        </div>
      </div>

      {/* Active Profiles Section */}
      <div className='rounded-xl border bg-card p-6 shadow-sm space-y-4'>
        <div className='flex items-center justify-between'>
          <div>
            <h2 className='text-lg font-semibold'>端口工作台配置</h2>
            <p className='text-xs text-muted-foreground'>
              可独立控制各项监听端口配置是否包含在内核运行加载项中
            </p>
          </div>
          <Badge variant='secondary' className='font-mono text-xs'>
            已启用 {profiles.filter((p) => p.profile.include_in_runtime).length}{' '}
            / {profiles.length} 项
          </Badge>
        </div>

        {profiles.length === 0 ? (
          <div className='text-center py-8 text-muted-foreground text-sm border rounded-lg'>
            暂无端口配置。请先在工作台中新建并关联节点。
          </div>
        ) : (
          <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
            {profiles.map((item) => {
              const enabled = Boolean(item.profile.include_in_runtime);
              return (
                <div
                  key={item.profile.id}
                  className={`rounded-lg border p-4 transition-all ${
                    enabled
                      ? 'border-primary/40 bg-card'
                      : 'border-border/60 bg-muted/20 opacity-80'
                  }`}
                >
                  <div className='flex items-center justify-between mb-3'>
                    <div className='flex items-center gap-2'>
                      <span className='font-semibold text-base'>
                        {item.profile.name}
                      </span>
                      <Badge
                        variant='outline'
                        className='uppercase font-mono text-[10px] px-1.5 py-0'
                      >
                        {item.profile.proxy_settings?.strategy_type || 'select'}
                      </Badge>
                    </div>

                    <div className='flex items-center gap-2'>
                      <span
                        className={`text-xs font-semibold ${enabled ? 'text-emerald-500' : 'text-muted-foreground'}`}
                      >
                        {enabled ? '已启用' : '已禁用'}
                      </span>
                      <Switch
                        checked={enabled}
                        onCheckedChange={(val) =>
                          handleToggleProfile(item, val)
                        }
                      />
                    </div>
                  </div>

                  <div className='text-xs text-muted-foreground space-y-1.5 border-t pt-3'>
                    <div className='flex justify-between items-center'>
                      <span>混合监听端口:</span>
                      <span className='font-mono font-bold text-foreground'>
                        {item.profile.mixed_port}
                      </span>
                    </div>
                    <div className='flex justify-between items-center'>
                      <span>监听绑定地址:</span>
                      <span className='font-mono'>
                        {item.profile.listen_host}
                      </span>
                    </div>
                    <div className='flex justify-between items-center'>
                      <span>绑定出口节点:</span>
                      <span className='font-bold text-foreground'>
                        {item.node_ids?.length || 0} 个
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Live Log Stream Panel */}
      <div className='rounded-xl border bg-card shadow-sm p-6 space-y-4'>
        <div className='flex flex-col sm:flex-row sm:items-center justify-between gap-3'>
          <div className='flex items-center gap-2'>
            <Terminal className='h-5 w-5 text-primary' />
            <h2 className='text-base font-semibold'>内核运行日志</h2>
          </div>
          <div className='relative w-full sm:w-64'>
            <Search className='absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground' />
            <Input
              placeholder='过滤日志关键字...'
              value={logFilter}
              onChange={(e) => setLogFilter(e.target.value)}
              className='pl-8 text-xs h-8 bg-background'
            />
          </div>
        </div>

        <pre className='bg-slate-950 text-slate-100 font-mono text-xs p-4 rounded-xl overflow-y-auto max-h-[350px] leading-relaxed space-y-1'>
          {filteredLogs.length === 0 ? (
            <span className='text-slate-500'>暂无日志输出...</span>
          ) : (
            filteredLogs.map((l, i) => (
              <div key={i} className='whitespace-pre-wrap'>
                {l}
              </div>
            ))
          )}
        </pre>
      </div>

      {/* Active Config Sheet Drawer */}
      <Sheet open={configDrawerOpen} onOpenChange={setConfigDrawerOpen}>
        <SheetContent className='sm:max-w-xl'>
          <SheetHeader>
            <SheetTitle className='flex items-center justify-between pr-6'>
              <span>当前实际运行 YAML 配置</span>
              <Button
                size='sm'
                variant='outline'
                onClick={handleCopyConfig}
                className='gap-1.5 text-xs'
              >
                <Copy className='h-3.5 w-3.5' /> 复制配置
              </Button>
            </SheetTitle>
          </SheetHeader>
          <div className='py-4 space-y-3'>
            {activeConfig && (
              <>
                <div className='flex items-center justify-between text-xs text-muted-foreground font-mono bg-muted/40 p-2.5 rounded-md border'>
                  <span>校验和: {activeConfig.checksum.slice(0, 16)}...</span>
                  <Badge
                    variant='secondary'
                    className='uppercase font-mono text-[10px]'
                  >
                    {activeConfig.kernel_type || 'mihomo'}
                  </Badge>
                </div>
                <pre className='bg-slate-950 text-slate-100 font-mono text-xs p-4 rounded-lg overflow-x-auto max-h-[75vh] whitespace-pre-wrap break-all leading-relaxed'>
                  {activeConfig.content}
                </pre>
              </>
            )}
          </div>
        </SheetContent>
      </Sheet>

      {/* Kernel Binary Management Modal */}
      <Dialog open={kernelModalOpen} onOpenChange={setKernelModalOpen}>
        <DialogContent className='max-w-md'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <Cpu className='h-5 w-5 text-primary' />
              <span>Mihomo 内核文件管理</span>
            </DialogTitle>
          </DialogHeader>

          <div className='space-y-4 py-2 text-xs'>
            <div className='bg-muted/40 p-3 rounded-lg border space-y-1.5'>
              <div className='flex justify-between items-center'>
                <span className='text-muted-foreground'>检测版本:</span>
                <span className='font-mono font-bold text-foreground'>
                  {installedInfo?.detected_version ||
                    status?.mihomo_version ||
                    '未检测到二进制'}
                </span>
              </div>
              <div className='flex justify-between items-center'>
                <span className='text-muted-foreground'>当前来源:</span>
                <span className='font-mono text-foreground'>
                  {installedInfo?.binary_source ||
                    (status?.kernel_instance?.work_dir
                      ? '本地独立安装'
                      : '系统/全局内置')}
                </span>
              </div>
            </div>

            <div className='grid grid-cols-1 gap-2.5 pt-1'>
              <Button
                onClick={handleInspect}
                disabled={installing}
                variant='outline'
                className='w-full justify-start gap-2 text-xs'
              >
                <Info className='h-4 w-4 text-blue-500' />
                <span>校验本地内核版本与可用性</span>
              </Button>

              <Button
                onClick={handleAutoDownload}
                disabled={installing}
                variant='outline'
                className='w-full justify-start gap-2 text-xs'
              >
                <Download className='h-4 w-4 text-emerald-500' />
                <span>一键从官方源自动下载/升级最新内核</span>
              </Button>

              <div>
                <input
                  type='file'
                  ref={fileInputRef}
                  onChange={handleFileUpload}
                  className='hidden'
                  accept='*'
                />
                <Button
                  type='button'
                  onClick={() => fileInputRef.current?.click()}
                  disabled={installing}
                  variant='outline'
                  className='w-full justify-start gap-2 text-xs'
                >
                  <Upload className='h-4 w-4 text-violet-500' />
                  <span>手动上传自定义 Mihomo 二进制文件</span>
                </Button>
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
