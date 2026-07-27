import * as React from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { HelpCircle, Trash2 } from 'lucide-react';

export interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: React.ReactNode;
  description?: React.ReactNode;
  confirmText?: string;
  cancelText?: string;
  variant?: 'default' | 'destructive';
  loading?: boolean;
  onConfirm: () => void | Promise<void>;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmText,
  cancelText,
  variant = 'destructive',
  loading = false,
  onConfirm,
}: ConfirmDialogProps) {
  const { t } = useTranslation();

  const handleConfirm = async () => {
    await onConfirm();
  };

  const isDestructive = variant === 'destructive';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <div className="flex items-start gap-3">
            <div
              className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${
                isDestructive
                  ? 'bg-destructive/10 text-destructive'
                  : 'bg-primary/10 text-primary'
              }`}
            >
              {isDestructive ? (
                <Trash2 className="h-5 w-5" />
              ) : (
                <HelpCircle className="h-5 w-5" />
              )}
            </div>
            <div className="flex flex-col gap-1 text-left pt-0.5">
              <DialogTitle>
                {title || (isDestructive ? t('common.delete') : t('common.confirm'))}
              </DialogTitle>
              {description && <DialogDescription>{description}</DialogDescription>}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter className="mt-2 flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2">
          <Button
            type="button"
            variant="outline"
            disabled={loading}
            onClick={() => onOpenChange(false)}
          >
            {cancelText || t('common.cancel')}
          </Button>
          <Button
            type="button"
            variant={variant}
            disabled={loading}
            onClick={handleConfirm}
          >
            {loading && <Spinner className="mr-2 h-4 w-4 animate-spin" />}
            {confirmText || (isDestructive ? t('common.delete') : t('common.confirm'))}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
