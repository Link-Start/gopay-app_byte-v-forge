import { Badge } from '@byte-v-forge/common-ui';
import type { GoPayPhoneCheckResponse } from './gopay-api';

type Tone = 'ok' | 'warn' | 'bad' | 'idle';

export function GoPayPhoneResult({ phone, result, loading }: { phone?: string; result?: GoPayPhoneCheckResponse | null; loading?: boolean }) {
  const rateLimited = result?.status === 'rate_limited';
  const failed = Boolean(result && result.success === false);
  const registered = result?.status === 'registered';
  const available = Boolean(result?.available === true || result?.status === 'available');
  const registerLabel = failed ? '未知' : registered ? '已注册' : available ? '未注册' : '未知';
  const registerTone: Tone = failed ? 'idle' : registered ? 'warn' : available ? 'ok' : 'idle';
  const requestLabel = loading ? '执行中' : rateLimited ? '限流' : result ? (failed ? '失败' : '成功') : '未执行';
  const requestTone: Tone = loading ? 'idle' : rateLimited ? 'warn' : failed ? 'bad' : result ? 'ok' : 'idle';
  const details = [
    statusDetail(result),
    failed ? detail('原因', result?.error_message, true, rateLimited ? 'warn' : 'bad') : null,
    detail('代理', result?.proxy_hash ? `#${result.proxy_hash}` : '')
  ].filter((item): item is NonNullable<typeof item> => Boolean(item));
  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between gap-2">
        <div className="flex min-w-0 items-baseline gap-2">
          <span className="shrink-0 text-xs font-medium">检测结果</span>
          <span className="truncate font-mono text-[11px] text-muted-foreground">{phone || '-'}</span>
        </div>
        <Badge variant={loading ? 'secondary' : rateLimited ? 'secondary' : failed ? 'destructive' : available ? 'default' : 'secondary'}>{loading ? '执行中' : rateLimited ? '限流' : failed ? '失败' : '完成'}</Badge>
      </div>
      <div className="flex flex-wrap gap-1.5">
        <MetricChip label="注册" value={registerLabel} tone={registerTone} />
        <MetricChip label="请求" value={requestLabel} tone={requestTone} />
      </div>
      {details.length > 0 && <div className="grid grid-cols-2 gap-x-3 gap-y-1 rounded-md border bg-background/70 px-2.5 py-1.5 text-[11px]">{details.map((item) => <InfoItem key={item.label} {...item} />)}</div>}
    </div>
  );
}

function detail(label: string, value?: string, wide = false, tone: Tone = 'idle') {
  const text = value?.trim();
  return text ? { label, value: text, wide, tone } : null;
}

function statusDetail(result?: GoPayPhoneCheckResponse | null) {
  if (!result?.status || result.status === 'registered' || result.status === 'available') return null;
  return detail('状态', statusLabel(result.status), false, result.status === 'rate_limited' ? 'warn' : 'idle');
}

function statusLabel(status: string) {
  if (status === 'rate_limited') return '限流';
  if (status === 'proxy_unavailable') return '代理不可用';
  if (status === 'check_failed') return '探测失败';
  return status;
}

function MetricChip({ label, value, tone }: { label: string; value: string; tone: Tone }) {
  return <span className={`inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[11px] ${toneClass(tone, true)}`}><span className="text-muted-foreground">{label}</span><span className="font-semibold">{value}</span></span>;
}

function InfoItem({ label, value, wide, tone = 'idle' }: { label: string; value: string; wide?: boolean; tone?: Tone }) {
  return <div className={`min-w-0 ${wide ? 'col-span-2' : ''}`}><span className="mr-1 text-muted-foreground">{label}</span><span className={`break-words font-medium ${toneClass(tone)}`}>{value}</span></div>;
}

function toneClass(tone: Tone, chip = false) {
  const base = tone === 'ok' ? 'text-primary' : tone === 'bad' ? 'text-destructive' : tone === 'warn' ? 'text-amber-600 dark:text-amber-400' : 'text-muted-foreground';
  if (!chip) return base;
  if (tone === 'ok') return 'border-primary/30 bg-primary/5 text-primary';
  if (tone === 'bad') return 'border-destructive/30 bg-destructive/5 text-destructive';
  if (tone === 'warn') return 'border-amber-500/30 bg-amber-500/5 text-amber-700 dark:text-amber-300';
  return 'bg-muted/30 text-muted-foreground';
}
