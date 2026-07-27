import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { Textarea } from '@/components/ui/textarea';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import {
  listSourceConfigs,
  uploadSourceConfig,
  fetchSubscription,
  confirmSourceConfig,
  deleteSourceConfig,
  refreshSourceConfig,
  reuploadSourceConfig,
  type SourceConfig,
  type UploadSourceConfigResponse,
} from '@/shared/api/clash';
import {
  Upload,
  Link as LinkIcon,
  CheckCircle,
  Trash2,
  FileText,
  DownloadCloud,
  AlertCircle,
  FileCode,
  UploadCloud,
  X,
  RefreshCw,
} from 'lucide-react';
import { ConfirmDialog } from '@/shared/components/confirm-dialog';

export function SourceConfigsPage() {
  const [sources, setSources] = useState<SourceConfig[]>([]);
  const [loading, setLoading] = useState(true);

  // Upload modal state
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false);
  const [uploadMode, setUploadMode] = useState<'file' | 'text'>('file');
  const [filename, setFilename] = useState('');
  const [rawContent, setRawContent] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploading, setUploading] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // Subscription modal state
  const [subDialogOpen, setSubDialogOpen] = useState(false);
  const [subUrl, setSubUrl] = useState('');
  const [fetchingSub, setFetchingSub] = useState(false);

  // Parse Result Preview state
  const [previewData, setPreviewData] =
    useState<UploadSourceConfigResponse | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [syncingId, setSyncingId] = useState<number | null>(null);

  // Reupload existing source
  const [reuploadTarget, setReuploadTarget] = useState<SourceConfig | null>(
    null,
  );
  const [reuploadContent, setReuploadContent] = useState('');
  const [reuploadFilename, setReuploadFilename] = useState('');
  const [reuploading, setReuploading] = useState(false);

  const fetchSources = async (ignoreCheck?: unknown) => {
    const isIgnored = () => typeof ignoreCheck === 'function' && ignoreCheck();
    try {
      setLoading(true);
      const res = await listSourceConfigs(1, 50);
      if (isIgnored()) return;
      setSources(res.items);
    } catch (err) {
      console.error('Failed to list source configs:', err);
    } finally {
      if (!isIgnored()) setLoading(false);
    }
  };

  useEffect(() => {
    let ignore = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchSources(() => ignore);
    return () => {
      ignore = true;
    };
  }, []);

  const handleFileRead = (file: File) => {
    if (!file) return;
    setSelectedFile(file);
    setFilename(file.name);

    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      setRawContent(text || '');
    };
    reader.readAsText(file);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFileRead(e.target.files[0]);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileRead(e.dataTransfer.files[0]);
    }
  };

  const clearSelectedFile = () => {
    setSelectedFile(null);
    setFilename('');
    setRawContent('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!rawContent.trim()) {
      toast.warning('请选择 YAML 文件或手动输入配置文件内容');
      return;
    }
    try {
      setUploading(true);
      const res = await uploadSourceConfig(
        filename || 'config.yaml',
        rawContent,
      );
      setUploadDialogOpen(false);
      clearSelectedFile();
      setPreviewData(res);
      await fetchSources();
      toast.success('配置解析成功，请确认导入节点');
    } catch (err) {
      toast.error(
        `上传失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setUploading(false);
    }
  };

  const handleFetchSubscription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!subUrl.trim()) {
      toast.warning('请输入订阅地址');
      return;
    }
    try {
      setFetchingSub(true);
      const res = await fetchSubscription(subUrl.trim());
      setSubDialogOpen(false);
      setSubUrl('');
      setPreviewData(res);
      await fetchSources();
      toast.success('订阅拉取并解析成功，请确认导入节点');
    } catch (err) {
      toast.error(
        `拉取订阅失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setFetchingSub(false);
    }
  };

  const handleConfirmImport = async (id: number) => {
    try {
      setConfirming(true);
      const res = await confirmSourceConfig(id);
      toast.success(`成功导入 ${res.imported_nodes} 个节点入库！`);
      setPreviewData(null);
      await fetchSources();
    } catch (err) {
      toast.error(
        `确认导入失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setConfirming(false);
    }
  };

  // Delete confirmation state
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [deleting, setDeleting] = useState(false);

  const handleDeleteClick = (id: number) => {
    setDeleteId(id);
  };

  const handleConfirmDelete = async () => {
    if (!deleteId) return;
    setDeleting(true);
    try {
      await deleteSourceConfig(deleteId);
      await fetchSources();
      toast.success('删除配置源及绑定节点成功');
      setDeleteId(null);
    } catch (err) {
      toast.error(
        `删除失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setDeleting(false);
    }
  };

  const handleRefresh = async (id: number) => {
    try {
      setSyncingId(id);
      const res = await refreshSourceConfig(id);
      toast.success(
        `刷新完成：入库 ${res.imported_nodes} 个节点（重复 ${res.duplicate_nodes}）`,
      );
      await fetchSources();
    } catch (err) {
      toast.error(
        `刷新失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setSyncingId(null);
    }
  };

  const openReupload = (item: SourceConfig) => {
    setReuploadTarget(item);
    setReuploadFilename(item.filename || 'config.yaml');
    setReuploadContent('');
  };

  const handleReupload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!reuploadTarget) return;
    if (!reuploadContent.trim()) {
      toast.warning('请粘贴 YAML 内容');
      return;
    }
    try {
      setReuploading(true);
      const res = await reuploadSourceConfig(
        reuploadTarget.id,
        reuploadFilename,
        reuploadContent,
      );
      toast.success(
        `重新上传完成：入库 ${res.imported_nodes} 个节点（重复 ${res.duplicate_nodes}）`,
      );
      setReuploadTarget(null);
      setReuploadContent('');
      await fetchSources();
    } catch (err) {
      toast.error(
        `重新上传失败: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setReuploading(false);
    }
  };

  return (
    <div className='space-y-6 p-6'>
      <div className='flex flex-col md:flex-row md:items-center justify-between gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight text-foreground'>
            配置源导入
          </h1>
          <p className='text-sm text-muted-foreground'>
            支持拖拽/点击选择 YAML 文件、粘贴文本输入或通过 HTTP
            订阅地址一键解析节点
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button onClick={() => setUploadDialogOpen(true)} className='gap-2'>
            <Upload className='h-4 w-4' /> 上传 YAML 文件
          </Button>
          <Button
            onClick={() => setSubDialogOpen(true)}
            variant='outline'
            className='gap-2'
          >
            <LinkIcon className='h-4 w-4' /> 拉取订阅 URL
          </Button>
        </div>
      </div>

      {/* Sources List Table */}
      <div className='rounded-xl border bg-card shadow-sm'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称/类型</TableHead>
              <TableHead>节点统计</TableHead>
              <TableHead>哈希摘要</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>导入时间</TableHead>
              <TableHead className='text-right'>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className='text-center py-8 text-muted-foreground'
                >
                  加载中...
                </TableCell>
              </TableRow>
            ) : sources.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className='text-center py-8 text-muted-foreground'
                >
                  暂无配置源。请点击上方“上传 YAML 文件”或“拉取订阅 URL”。
                </TableCell>
              </TableRow>
            ) : (
              sources.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <div className='font-medium flex items-center gap-2'>
                      <FileText className='h-4 w-4 text-primary' />{' '}
                      {item.filename}
                    </div>
                    <div className='text-xs text-muted-foreground'>
                      {item.source_type === 'subscription_url'
                        ? item.source_url
                        : '本地上传'}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className='text-xs space-y-0.5'>
                      <div>
                        有效节点:{' '}
                        <span className='font-bold text-emerald-600'>
                          {item.valid_nodes}
                        </span>
                      </div>
                      <div>
                        重复节点:{' '}
                        <span className='text-amber-600'>
                          {item.duplicate_nodes}
                        </span>
                      </div>
                      <div>
                        已入库:{' '}
                        <span className='font-bold text-primary'>
                          {item.imported_nodes}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className='font-mono text-xs text-muted-foreground'>
                    {item.content_hash.slice(0, 12)}...
                  </TableCell>
                  <TableCell>
                    {item.status === 'imported' ? (
                      <Badge
                        variant='outline'
                        className='bg-emerald-500/10 text-emerald-600 border-emerald-500/20'
                      >
                        已入库
                      </Badge>
                    ) : (
                      <Badge variant='secondary'>已解析未确认</Badge>
                    )}
                  </TableCell>
                  <TableCell className='text-xs text-muted-foreground'>
                    {new Date(item.created_at).toLocaleString('zh-CN')}
                  </TableCell>
                  <TableCell className='text-right space-x-1'>
                    {item.status !== 'imported' && (
                      <Button
                        size='sm'
                        onClick={() => handleConfirmImport(item.id)}
                        disabled={confirming}
                        className='gap-1'
                      >
                        <CheckCircle className='h-3.5 w-3.5' /> 确认导入
                      </Button>
                    )}
                    {item.status === 'imported' &&
                      item.source_type === 'subscription_url' && (
                        <Button
                          size='sm'
                          variant='outline'
                          className='gap-1'
                          disabled={syncingId === item.id}
                          onClick={() => handleRefresh(item.id)}
                        >
                          <RefreshCw
                            className={`h-3.5 w-3.5 ${syncingId === item.id ? 'animate-spin' : ''}`}
                          />
                          刷新
                        </Button>
                      )}
                    {item.status === 'imported' &&
                      item.source_type === 'upload' && (
                        <Button
                          size='sm'
                          variant='outline'
                          className='gap-1'
                          onClick={() => openReupload(item)}
                        >
                          <Upload className='h-3.5 w-3.5' /> 重新上传
                        </Button>
                      )}
                    {item.status === 'imported' && (
                      <Button
                        size='sm'
                        variant='outline'
                        onClick={() => handleConfirmImport(item.id)}
                        disabled={confirming}
                        className='gap-1'
                      >
                        <RefreshCw className='h-3.5 w-3.5' /> 重新同步
                      </Button>
                    )}
                    <Button
                      size='sm'
                      variant='ghost'
                      className='text-destructive'
                      onClick={() => handleDeleteClick(item.id)}
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

      {/* Upload YAML Modal */}
      <Dialog open={uploadDialogOpen} onOpenChange={setUploadDialogOpen}>
        <DialogContent className='max-w-xl'>
          <DialogHeader>
            <DialogTitle>上传 YAML 配置文件</DialogTitle>
          </DialogHeader>

          <Tabs
            value={uploadMode}
            onValueChange={(v: string) => setUploadMode(v as 'file' | 'text')}
            className='w-full'
          >
            <TabsList className='grid grid-cols-2 w-full mb-4'>
              <TabsTrigger value='file' className='gap-2'>
                <UploadCloud className='h-4 w-4' /> 文件拖拽 / 选择
              </TabsTrigger>
              <TabsTrigger value='text' className='gap-2'>
                <FileCode className='h-4 w-4' /> 手动文本输入
              </TabsTrigger>
            </TabsList>

            <form onSubmit={handleUpload} className='space-y-4'>
              <TabsContent value='file' className='space-y-4 m-0'>
                {/* Drag and Drop Zone */}
                <input
                  ref={fileInputRef}
                  type='file'
                  accept='.yaml,.yml,.txt'
                  onChange={handleFileSelect}
                  className='hidden'
                />

                {!selectedFile ? (
                  <div
                    onDragOver={handleDragOver}
                    onDragLeave={handleDragLeave}
                    onDrop={handleDrop}
                    onClick={() => fileInputRef.current?.click()}
                    className={`border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-all ${
                      isDragging
                        ? 'border-primary bg-primary/5 scale-[1.01]'
                        : 'border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/30'
                    }`}
                  >
                    <UploadCloud className='h-10 w-10 mx-auto mb-3 text-muted-foreground' />
                    <p className='font-semibold text-sm'>
                      点击选择文件，或将 .yaml / .yml 文件拖拽至此处
                    </p>
                    <p className='text-xs text-muted-foreground mt-1'>
                      支持 Clash / Mihomo 配置文件
                    </p>
                  </div>
                ) : (
                  <div className='rounded-xl border bg-muted/40 p-4 flex items-center justify-between'>
                    <div className='flex items-center gap-3'>
                      <FileText className='h-8 w-8 text-primary' />
                      <div>
                        <div className='font-semibold text-sm'>
                          {selectedFile.name}
                        </div>
                        <div className='text-xs text-muted-foreground font-mono'>
                          {(selectedFile.size / 1024).toFixed(1)} KB | YAML
                          已读取 ({rawContent.length} 字符)
                        </div>
                      </div>
                    </div>
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      onClick={clearSelectedFile}
                    >
                      <X className='h-4 w-4 text-muted-foreground hover:text-destructive' />
                    </Button>
                  </div>
                )}
              </TabsContent>

              <TabsContent value='text' className='space-y-4 m-0'>
                <div>
                  <label className='text-sm font-medium'>
                    配置名称 / 文件名
                  </label>
                  <Input
                    placeholder='custom_config.yaml'
                    value={filename}
                    onChange={(e) => setFilename(e.target.value)}
                    className='mt-1'
                  />
                </div>
                <div>
                  <label className='text-sm font-medium'>YAML 内容</label>
                  <Textarea
                    placeholder='粘贴包含 proxies 的 Clash / Mihomo YAML 内容...'
                    rows={8}
                    value={rawContent}
                    onChange={(e) => setRawContent(e.target.value)}
                    className='font-mono text-xs mt-1'
                  />
                </div>
              </TabsContent>

              <DialogFooter className='pt-2'>
                <Button
                  type='button'
                  variant='outline'
                  onClick={() => setUploadDialogOpen(false)}
                >
                  取消
                </Button>
                <Button
                  type='submit'
                  disabled={uploading || !rawContent.trim()}
                >
                  {uploading ? '解析中...' : '开始解析并预览'}
                </Button>
              </DialogFooter>
            </form>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* Fetch Subscription Modal */}
      <Dialog open={subDialogOpen} onOpenChange={setSubDialogOpen}>
        <DialogContent className='max-w-md'>
          <DialogHeader>
            <DialogTitle>拉取订阅 URL</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleFetchSubscription} className='space-y-4 py-2'>
            <div>
              <label className='text-sm font-medium'>HTTP/HTTPS 订阅地址</label>
              <Input
                placeholder='https://example.com/sub/yaml'
                value={subUrl}
                onChange={(e) => setSubUrl(e.target.value)}
                className='mt-1'
              />
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => setSubDialogOpen(false)}
              >
                取消
              </Button>
              <Button type='submit' disabled={fetchingSub} className='gap-2'>
                <DownloadCloud className='h-4 w-4' />{' '}
                {fetchingSub ? '拉取中...' : '开始拉取'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Reupload Modal */}
      <Dialog
        open={!!reuploadTarget}
        onOpenChange={(open) => !open && setReuploadTarget(null)}
      >
        <DialogContent className='max-w-xl'>
          <DialogHeader>
            <DialogTitle>重新上传配置源</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleReupload} className='space-y-4 py-2'>
            <p className='text-sm text-muted-foreground'>
              将覆盖「{reuploadTarget?.filename}
              」的内容，并以全删全插方式同步绑定节点。
            </p>
            <div>
              <label className='text-sm font-medium'>配置名称 / 文件名</label>
              <Input
                value={reuploadFilename}
                onChange={(e) => setReuploadFilename(e.target.value)}
                className='mt-1'
              />
            </div>
            <div>
              <label className='text-sm font-medium'>YAML 内容</label>
              <Textarea
                placeholder='粘贴包含 proxies 的 Clash / Mihomo YAML 内容...'
                rows={10}
                value={reuploadContent}
                onChange={(e) => setReuploadContent(e.target.value)}
                className='font-mono text-xs mt-1'
              />
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => setReuploadTarget(null)}
              >
                取消
              </Button>
              <Button
                type='submit'
                disabled={reuploading || !reuploadContent.trim()}
              >
                {reuploading ? '同步中...' : '覆盖并同步节点'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Parse Result Preview Modal */}
      {previewData && (
        <Dialog open={!!previewData} onOpenChange={() => setPreviewData(null)}>
          <DialogContent className='max-w-2xl max-h-[80vh] overflow-y-auto'>
            <DialogHeader>
              <DialogTitle>配置解析结果报告</DialogTitle>
            </DialogHeader>
            <div className='space-y-4 py-2 text-sm'>
              <div className='grid grid-cols-4 gap-2 bg-muted p-3 rounded-lg text-center font-semibold'>
                <div>总项数: {previewData.config.total_nodes}</div>
                <div className='text-emerald-600'>
                  有效节点: {previewData.config.valid_nodes}
                </div>
                <div className='text-amber-600'>
                  重复节点: {previewData.config.duplicate_nodes}
                </div>
                <div className='text-rose-600'>
                  无效解析: {previewData.config.invalid_nodes}
                </div>
              </div>

              {previewData.parse_result.issues.length > 0 && (
                <div className='rounded-lg border border-amber-500/20 bg-amber-500/10 p-3 space-y-1 text-xs text-amber-700'>
                  <div className='font-bold flex items-center gap-1'>
                    <AlertCircle className='h-4 w-4' /> 异常与格式警告:
                  </div>
                  {previewData.parse_result.issues.map((issue, idx) => (
                    <div key={idx}>
                      Line #{issue.index}: {issue.message}
                    </div>
                  ))}
                </div>
              )}

              <div>
                <h4 className='font-semibold mb-2'>节点解析明细 (前 10 个)</h4>
                <div className='border rounded-md divide-y max-h-48 overflow-y-auto font-mono text-xs'>
                  {previewData.parse_result.nodes.slice(0, 10).map((n, idx) => (
                    <div key={idx} className='p-2 flex justify-between'>
                      <span className='font-bold truncate max-w-[200px]'>
                        {n.name}
                      </span>
                      <span className='uppercase text-muted-foreground'>
                        {n.type}
                      </span>
                      <span>
                        {n.server}:{n.port}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant='outline' onClick={() => setPreviewData(null)}>
                稍后确认
              </Button>
              <Button
                onClick={() => handleConfirmImport(previewData.config.id)}
                disabled={confirming}
                className='gap-2'
              >
                <CheckCircle className='h-4 w-4' /> 确认入库 (
                {previewData.config.valid_nodes -
                  previewData.config.duplicate_nodes}{' '}
                新节点)
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title='删除配置源'
        description='确定删除该配置源？将同时删除其绑定的全部节点，并从端口配置中移除这些节点。此操作不可撤销。'
        loading={deleting}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
