import { useEffect } from 'react';
import { Save, Settings } from 'lucide-react';
import { Badge, Button, DashboardField, Input, WorkspaceToolbar, useForm } from '@byte-v-forge/common-ui';
import type { GoPayRegisterIndonesiaWASettings } from '../proto/gopay_app';

const defaults: GoPayRegisterIndonesiaWASettings = {
  sms_acquire_wait_seconds: 90,
  sms_min_available_count: 1,
  sms_max_price_amount_decimal: '',
  sms_max_price_currency_code: 'USD'
};

export function GoPaySettingsTab({ settings, loading, saving, onSave }: {
  settings?: GoPayRegisterIndonesiaWASettings;
  loading?: boolean;
  saving?: boolean;
  onSave: (settings: GoPayRegisterIndonesiaWASettings) => void;
}) {
  const form = useForm<GoPayRegisterIndonesiaWASettings>({ defaultValues: defaults });
  useEffect(() => {
    form.reset(normalizeSettings(settings));
  }, [form, settings]);

  return (
    <form className="flex min-h-0 flex-1 flex-col" onSubmit={form.handleSubmit((values) => onSave(normalizeSettings(values)))}>
      <WorkspaceToolbar
        title={<span className="inline-flex items-center gap-2"><Settings size={16} />GoPay 配置</span>}
        meta="注册印尼 WA 必要项"
        actions={<Button aria-label="保存配置" disabled={loading || saving} size="icon-sm" type="submit"><Save size={14} /></Button>}
      />
      <div className="overflow-auto bg-muted/30 p-4">
        <section className="grid w-[420px] max-w-full gap-3 rounded-xl border border-border/70 bg-background p-4 shadow-sm">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h3 className="m-0 text-sm font-semibold">印尼 WA 注册取号</h3>
              <p className="m-0 mt-1 text-xs leading-5 text-muted-foreground">最低价优先，且要求库存不低于阈值。</p>
            </div>
            <Badge variant="secondary">SMS</Badge>
          </div>
          <div className="rounded-lg border border-border/70 bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
            SMS 应用固定 <span className="font-mono text-foreground">whatsapp</span>，国家固定 <span className="font-mono text-foreground">ID / +62</span>。
          </div>
          <div className="grid grid-cols-2 gap-3">
            <DashboardField label="取号等待秒数">
              <Input min={1} type="number" {...form.register('sms_acquire_wait_seconds', { valueAsNumber: true })} />
            </DashboardField>
            <DashboardField label="最小库存">
              <Input min={1} type="number" {...form.register('sms_min_available_count', { valueAsNumber: true })} />
            </DashboardField>
          </div>
          <div className="grid grid-cols-[1fr_110px] gap-3">
            <DashboardField label="SMS 最大单价">
              <Input placeholder="例如 0.50；留空不限制" {...form.register('sms_max_price_amount_decimal')} />
            </DashboardField>
            <DashboardField label="币种">
              <Input placeholder="USD" {...form.register('sms_max_price_currency_code')} />
            </DashboardField>
          </div>
        </section>
      </div>
    </form>
  );
}

function normalizeSettings(settings?: GoPayRegisterIndonesiaWASettings): GoPayRegisterIndonesiaWASettings {
  return {
    sms_acquire_wait_seconds: positiveNumber(settings?.sms_acquire_wait_seconds, defaults.sms_acquire_wait_seconds),
    sms_min_available_count: positiveNumber(settings?.sms_min_available_count, defaults.sms_min_available_count),
    sms_max_price_amount_decimal: normalizePrice(settings?.sms_max_price_amount_decimal),
    sms_max_price_currency_code: (settings?.sms_max_price_currency_code?.trim() || defaults.sms_max_price_currency_code).toUpperCase()
  };
}

function positiveNumber(value: unknown, fallback: number) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : fallback;
}

function normalizePrice(value?: string) {
  const trimmed = value?.trim() || '';
  return /^\d+(\.\d+)?$/.test(trimmed) ? trimmed : '';
}
