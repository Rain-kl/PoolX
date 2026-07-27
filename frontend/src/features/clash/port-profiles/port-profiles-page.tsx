import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Switch } from '@/components/ui/switch';
import {
  listPortProfiles,
  createPortProfile,
  updatePortProfile,
  deletePortProfile,
  previewPortProfile,
  listProxyNodes,
  type PortProfileWithNodes,
  type PortProfileProxySettings,
  type ProxyNode,
  type RenderResult,
} from '@/shared/api/clash';
import { Plus, Edit3, Trash2, Eye, Settings, ChevronDown } from 'lucide-react';
import { ConfirmDialog } from '@/shared/components/confirm-dialog';

export function PortProfilesPage() {
  const [profiles, setProfiles] = useState<PortProfileWithNodes[]>([]);
  const [nodesPool, setNodesPool] = useState<ProxyNode[]>([]);
  const [loading, setLoading] = useState(true);

  // Profile Form Modal state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState<PortProfileWithNodes | null>(null);
  const [name, setName] = useState('');
  const [mixedPort, setMixedPort] = useState(7890);
  const [listenHost, setListenHost] = useState('0.0.0.0');
  const [strategyType, setStrategyType] = useState<'select' | 'url-test' | 'fallback' | 'load-balance'>('select');
  const [testUrl, setTestUrl] = useState('https://cp.cloudflare.com/generate_204');
  const [testInterval, setTestInterval] = useState(300);
  const [loadBalanceStrategy, setLoadBalanceStrategy] = useState<'consistent-hashing' | 'round-robin'>('consistent-hashing');
  const [udpEnabled, setUdpEnabled] = useState(true);
  const [authEnabled, setAuthEnabled] = useState(false);
  const [authUsername, setAuthUsername] = useState('admin');
  const [authPassword, setAuthPassword] = useState('');
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [selectedNodeIds, setSelectedNodeIds] = useState<number[]>([]);
  const [saving, setSaving] = useState(false);

  // YAML Preview Sheet state
  const [previewResult, setPreviewResult] = useState<RenderResult | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [pRes, nRes] = await Promise.all([
        listPortProfiles(),
        listProxyNodes({ pageSize: 200, enabled: true }),
      ]);
      setProfiles(pRes.items);
      setNodesPool(nRes.items);
    } catch (err) {
      console.error('Failed to load port profiles:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const openCreateDialog = () => {
    setEditingProfile(null);
    setName(`Profile-${Math.floor(1000 + Math.random() * 9000)}`);
    setMixedPort(7890 + profiles.length);
    setListenHost('0.0.0.0');
    setStrategyType('select');
    setTestUrl('https://cp.cloudflare.com/generate_204');
    setTestInterval(300);
    setLoadBalanceStrategy('consistent-hashing');
    setUdpEnabled(true);
    setAuthEnabled(false);
    setAuthUsername('admin');
    setAuthPassword(Math.random().toString(36).slice(-8));
    setAdvancedOpen(false);
    setSelectedNodeIds(nodesPool.slice(0, 3).map((n) => n.id));
    setDialogOpen(true);
  };

  const openEditDialog = (item: PortProfileWithNodes) => {
    setEditingProfile(item);
    setName(item.profile.name);
    setMixedPort(item.profile.mixed_port);
    setListenHost(item.profile.listen_host);
    const settings = item.profile.proxy_settings || {};
    setStrategyType(settings.strategy_type || 'select');
    setTestUrl(settings.test_url || 'https://cp.cloudflare.com/generate_204');
    setTestInterval(settings.test_interval_seconds || 300);
    setLoadBalanceStrategy(settings.load_balance_strategy || 'consistent-hashing');
    setUdpEnabled(settings.udp_enabled ?? true);
    setAuthEnabled(settings.auth_enabled ?? false);
    setAuthUsername(settings.auth_username || 'admin');
    setAuthPassword(settings.auth_password || '');
    setAdvancedOpen(Boolean(settings.auth_enabled || settings.udp_enabled === false));
    setSelectedNodeIds(item.node_ids || []);
    setDialogOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.warning('请输入配置名称');
      return;
    }
    if (mixedPort <= 0) {
      toast.warning('请输入有效的混合端口');
      return;
    }

    const proxySettings: PortProfileProxySettings = {
      strategy_type: strategyType,
      test_url: testUrl,
      test_interval_seconds: testInterval,
      load_balance_strategy: loadBalanceStrategy,
      udp_enabled: udpEnabled,
      auth_enabled: authEnabled,
      auth_username: authEnabled ? authUsername : undefined,
      auth_password: authEnabled ? authPassword : undefined,
    };

    try {
      setSaving(true);
      if (editingProfile) {
        await updatePortProfile(editingProfile.profile.id, {
          name,
          mixed_port: mixedPort,
          listen_host: listenHost,
          proxy_settings: proxySettings,
          include_in_runtime: true,
          node_ids: selectedNodeIds,
        });
        toast.success('更新端口配置成功');
      } else {
        await createPortProfile({
          name,
          mixed_port: mixedPort,
          listen_host: listenHost,
          proxy_settings: proxySettings,
          include_in_runtime: true,
          node_ids: selectedNodeIds,
        });
        toast.success('新建端口配置成功');
      }
      setDialogOpen(false);
      await fetchData();
    } catch (err) {
      toast.error(`保存失败: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setSaving(false);
    }
  };

  // Delete confirm state
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [deleting, setDeleting] = useState(false);

  const handleDeleteClick = (id: number) => {
    setDeleteId(id);
  };

  const handleConfirmDelete = async () => {
    if (!deleteId) return;
    setDeleting(true);
    try {
      await deletePortProfile(deleteId);
      await fetchData();
      toast.success('删除端口配置成功');
      setDeleteId(null);
    } catch (err) {
      toast.error(`删除失败: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setDeleting(false);
    }
  };

  const handlePreview = async (id: number) => {
    try {
      const res = await previewPortProfile(id);
      setPreviewResult(res);
      setPreviewOpen(true);
    } catch (err) {
      toast.error(`生成预览失败: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">工作台 (端口配置)</h1>
          <p className="text-sm text-muted-foreground">定义代理监听入口、选择出口节点集合与策略（轮询/自动选优/故障转移）</p>
        </div>
        <Button onClick={openCreateDialog} className="gap-2">
          <Plus className="h-4 w-4" /> 新建端口配置
        </Button>
      </div>

      {/* Profile Cards Grid */}
      {loading ? (
        <div className="text-center py-12 text-muted-foreground">加载工作台端口中...</div>
      ) : profiles.length === 0 ? (
        <div className="text-center py-12 border rounded-xl bg-card text-muted-foreground">
          暂无端口配置。点击右上角“新建端口配置”创建监听服务。
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {profiles.map((item) => (
            <div key={item.profile.id} className="rounded-xl border bg-card p-5 shadow-sm space-y-4 hover:border-primary/50 transition-all">
              <div className="flex items-center justify-between">
                <span className="font-bold text-lg">{item.profile.name}</span>
                <Badge variant="outline" className="uppercase font-mono text-xs bg-primary/5">
                  {item.profile.proxy_settings?.strategy_type || 'select'}
                </Badge>
              </div>

              <div className="space-y-2 text-xs text-muted-foreground border-y py-3">
                <div className="flex justify-between items-center">
                  <span>混合监听端口 (Mixed):</span>
                  <span className="font-mono font-bold text-sm text-foreground">{item.profile.mixed_port}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>监听绑定地址:</span>
                  <span className="font-mono text-foreground">{item.profile.listen_host}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>包含在运行内核:</span>
                  <span className={item.profile.include_in_runtime ? 'text-emerald-500 font-bold' : 'text-slate-400'}>
                    {item.profile.include_in_runtime ? '是' : '否'}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span>已绑定出口节点:</span>
                  <span className="font-bold text-primary">{item.node_ids?.length || 0} 个</span>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-1">
                <Button size="sm" variant="ghost" onClick={() => handlePreview(item.profile.id)} className="gap-1 text-xs">
                  <Eye className="h-3.5 w-3.5" /> 预览 YAML
                </Button>
                <Button size="sm" variant="outline" onClick={() => openEditDialog(item)} className="gap-1 text-xs">
                  <Edit3 className="h-3.5 w-3.5" /> 编辑
                </Button>
                <Button size="sm" variant="ghost" className="text-destructive" onClick={() => handleDeleteClick(item.profile.id)}>
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Profile Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingProfile ? '编辑端口配置' : '新建端口配置'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSave} className="space-y-4 py-2 text-sm">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="font-medium">配置名称</label>
                <Input value={name} onChange={(e) => setName(e.target.value)} className="mt-1" />
              </div>
              <div>
                <label className="font-medium">混合监听端口 (Mixed Port)</label>
                <Input type="number" value={mixedPort} onChange={(e) => setMixedPort(Number(e.target.value))} className="mt-1" />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="font-medium">监听地址 (Listen Host)</label>
                <Input value={listenHost} onChange={(e) => setListenHost(e.target.value)} className="mt-1" />
              </div>
              <div>
                <label className="font-medium">策略类型 (Strategy)</label>
                <Select value={strategyType} onValueChange={(val: any) => setStrategyType(val)}>
                  <SelectTrigger className="mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="select">手动选择 (select)</SelectItem>
                    <SelectItem value="url-test">自动优选 (url-test)</SelectItem>
                    <SelectItem value="fallback">故障转移 (fallback)</SelectItem>
                    <SelectItem value="load-balance">负载均衡 (load-balance)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {strategyType === 'load-balance' && (
              <div className="space-y-1.5 bg-muted/40 p-3 rounded-lg border">
                <label className="text-xs font-semibold text-foreground">负载均衡策略</label>
                <Select
                  value={loadBalanceStrategy}
                  onValueChange={(val: any) => setLoadBalanceStrategy(val)}
                >
                  <SelectTrigger className="h-8 text-xs bg-background">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="consistent-hashing">一致性哈希 (consistent-hashing)</SelectItem>
                    <SelectItem value="round-robin">轮询 (round-robin)</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-[11px] text-muted-foreground leading-normal">
                  一致性哈希会尽量让相同顶级域名走同一节点，轮询会平均分配请求。
                </p>
              </div>
            )}



            {/* Advanced Settings */}
            <div className="border rounded-lg overflow-hidden bg-card">
              <button
                type="button"
                onClick={() => setAdvancedOpen(!advancedOpen)}
                className="w-full flex items-center justify-between p-3 text-xs font-semibold bg-muted/30 hover:bg-muted/60 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <Settings className="h-3.5 w-3.5 text-muted-foreground" />
                  <span>高级设置</span>
                  <span className="text-[11px] font-normal text-muted-foreground">
                    (UDP 默认开启，监听鉴权默认关闭)
                  </span>
                </div>
                <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform duration-200 ${advancedOpen ? 'rotate-180' : ''}`} />
              </button>

              {advancedOpen && (
                <div className="p-4 space-y-4 border-t text-xs">
                  {/* UDP Switch */}
                  <div className="flex items-center justify-between gap-4">
                    <div className="space-y-0.5">
                      <div className="font-semibold text-foreground">UDP 转发</div>
                      <div className="text-[11px] text-muted-foreground">
                        控制当前监听入口是否允许 UDP 转发。
                      </div>
                    </div>
                    <Switch
                      checked={udpEnabled}
                      onCheckedChange={setUdpEnabled}
                    />
                  </div>

                  {/* Auth Switch */}
                  <div className="space-y-3 border-t pt-3">
                    <div className="flex items-center justify-between gap-4">
                      <div className="space-y-0.5">
                        <div className="font-semibold text-foreground">开启鉴权</div>
                        <div className="text-[11px] text-muted-foreground">
                          开启后会为当前监听入口生成 <code className="text-[10px] bg-muted px-1 py-0.5 rounded font-mono">users</code> 凭据列表。
                        </div>
                      </div>
                      <Switch
                        checked={authEnabled}
                        onCheckedChange={setAuthEnabled}
                      />
                    </div>

                    {authEnabled && (
                      <div className="grid grid-cols-2 gap-3 bg-muted/40 p-3 rounded-md border mt-2">
                        <div>
                          <label className="text-[11px] font-medium">鉴权用户名</label>
                          <Input
                            value={authUsername}
                            onChange={(e) => setAuthUsername(e.target.value)}
                            className="mt-1 h-8 text-xs bg-background"
                            placeholder="admin"
                          />
                        </div>
                        <div>
                          <label className="text-[11px] font-medium">鉴权密码</label>
                          <Input
                            type="password"
                            value={authPassword}
                            onChange={(e) => setAuthPassword(e.target.value)}
                            className="mt-1 h-8 text-xs bg-background"
                            placeholder="设置密码"
                          />
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Node Selection */}
            <div className="space-y-2">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <label className="font-semibold text-sm">
                  绑定出口节点 (已选择 <span className="text-primary font-bold">{selectedNodeIds.length}</span> / {nodesPool.length} 个)
                </label>
                <div className="flex items-center gap-1.5">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-7 text-xs px-2"
                    onClick={() => setSelectedNodeIds(nodesPool.map((n) => n.id))}
                  >
                    全选
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-7 text-xs px-2"
                    onClick={() =>
                      setSelectedNodeIds(
                        nodesPool.filter((n) => !selectedNodeIds.includes(n.id)).map((n) => n.id)
                      )
                    }
                  >
                    反选
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    className="h-7 text-xs px-2 text-muted-foreground hover:text-foreground"
                    onClick={() => setSelectedNodeIds([])}
                  >
                    清空
                  </Button>
                </div>
              </div>

              <div className="border rounded-lg p-3 max-h-56 overflow-y-auto space-y-2 bg-card">
                {nodesPool.length === 0 ? (
                  <div className="text-xs text-muted-foreground text-center py-4">节点池中暂无可用节点</div>
                ) : (
                  nodesPool.map((n) => {
                    const checked = selectedNodeIds.includes(n.id);
                    return (
                      <div key={n.id} className="flex items-center justify-between text-xs py-1 border-b last:border-b-0">
                        <label className="flex items-center gap-2 cursor-pointer select-none">
                          <Checkbox
                            checked={checked}
                            onCheckedChange={(c) => {
                              if (c) setSelectedNodeIds((prev) => [...prev, n.id]);
                              else setSelectedNodeIds((prev) => prev.filter((i) => i !== n.id));
                            }}
                          />
                          <span className="font-semibold">{n.name}</span>
                          <Badge variant="secondary" className="uppercase font-mono text-[10px] px-1 py-0">{n.type}</Badge>
                        </label>
                        <span className="font-mono text-muted-foreground">{n.server}:{n.port}</span>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>取消</Button>
              <Button type="submit" disabled={saving}>
                {saving ? '保存中...' : '保存端口配置'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* YAML Preview Drawer */}
      <Sheet open={previewOpen} onOpenChange={setPreviewOpen}>
        <SheetContent className="sm:max-w-xl">
          <SheetHeader>
            <SheetTitle>Mihomo YAML 渲染预览</SheetTitle>
          </SheetHeader>
          <div className="py-4 space-y-3">
            {previewResult && (
              <>
                <div className="flex items-center justify-between text-xs text-muted-foreground font-mono">
                  <span>Checksum: {previewResult.checksum.slice(0, 16)}...</span>
                  <Badge variant="secondary">mihomo</Badge>
                </div>
                <pre className="bg-slate-950 text-slate-100 font-mono text-xs p-4 rounded-lg overflow-x-auto max-h-[75vh]">
                  {previewResult.content}
                </pre>
              </>
            )}
          </div>
        </SheetContent>
      </Sheet>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title="删除端口配置"
        description="确定删除该端口配置？此操作不可撤销。"
        loading={deleting}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
