import { ResultSummaryPanel, type ResultTone } from '@byte-v-forge/common-ui';
import type { GoPayPhoneCheckResponse } from './gopay-api';

export function GoPayPhoneResult({ phone, result, loading }: { phone?: string; result?: GoPayPhoneCheckResponse | null; loading?: boolean }) {
  const rateLimited = result?.status === 'rate_limited';
  const failed = Boolean(result && result.success === false);
  const registered = result?.status === 'registered';
  const available = Boolean(result?.available === true || result?.status === 'available');
  const registerLabel = failed ? '未知' : registered ? '已注册' : available ? '未注册' : '未知';
  const registerTone: ResultTone = failed ? 'idle' : registered ? 'warn' : available ? 'ok' : 'idle';
  const requestLabel = loading ? '执行中' : rateLimited ? '限流' : result ? (failed ? '失败' : '成功') : '未执行';
  const requestTone: ResultTone = loading ? 'idle' : rateLimited ? 'warn' : failed ? 'bad' : result ? 'ok' : 'idle';
  const details = [
    statusDetail(result),
    failed ? detail('原因', result?.error_message, true, rateLimited ? 'warn' : 'bad') : null,
    detail('代理', result?.proxy_hash ? `#${result.proxy_hash}` : '')
  ].filter((item): item is NonNullable<typeof item> => Boolean(item));
  return (
    <ResultSummaryPanel
      title="检测结果"
      subject={phone}
      badge={{
        label: loading ? '执行中' : rateLimited ? '限流' : failed ? '失败' : '完成',
        variant: loading || rateLimited ? 'secondary' : failed ? 'destructive' : available ? 'default' : 'secondary',
      }}
      metrics={[
        { label: '注册', value: registerLabel, tone: registerTone },
        { label: '请求', value: requestLabel, tone: requestTone },
      ]}
      meta={details}
      metaLayout="grid"
    />
  );
}

function detail(label: string, value?: string, wide = false, tone: ResultTone = 'idle') {
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
