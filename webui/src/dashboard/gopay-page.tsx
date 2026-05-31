import type { ReactNode } from 'react';
import { Activity, KeyRound, RefreshCw, WalletCards, Workflow } from 'lucide-react';
import { ACCOUNT_PAGE_SIZE, AccountCarrierPanel, Alert, AlertDescription, AlertTitle, Badge, Button, Card, CardContent, ToastMessage, WorkspaceTabbedPanel, accountSubject, useAccountPages, useQuery, useToastMessage, type AccountListPagination } from '@byte-v-forge/common-ui';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import { getGoPayAccounts, getGoPayActionCatalog, getGoPayHealth, goPayKeys, type GoPayAccountProjection } from './gopay-api';

type GoPayTab = 'accounts' | 'workflows';

export function GoPayPage() {
  const toast = useToastMessage();
  const health = useQuery({ queryKey: goPayKeys.health, queryFn: getGoPayHealth, refetchInterval: 10000 });
  const catalog = useQuery({ queryKey: goPayKeys.actionCatalog, queryFn: getGoPayActionCatalog });
  const accounts = useAccountPages<GoPayAccountProjection, ListGopayAccountsResponse>({
    queryKey: goPayKeys.accounts,
    queryFn: getGoPayAccounts,
    refetchInterval: 10000,
    pageSize: ACCOUNT_PAGE_SIZE
  });
  const workflows = health.data?.workflows || [{ key: 'gopay-account', label: 'GoPayAccount 编排', webhook_path: 'gopay-app/account' }];
  return <><ToastMessage toast={toast.toast} /><WorkspaceTabbedPanel<GoPayTab> defaultValue="accounts" title={<span className="inline-flex items-center gap-2"><WalletCards className="size-4" />GoPay</span>} meta={health.data?.n8n_webhook_configured ? 'n8n 已接入' : '等待 n8n'} tabs={[
    { value: 'accounts', label: '账号', content: <AccountsTab accounts={accounts.accounts} loading={accounts.isLoading} pagination={accounts.pagination} actionCount={catalog.data?.actions?.length || 0} /> },
    { value: 'workflows', label: '工作流', content: <WorkflowTab configured={Boolean(health.data?.n8n_webhook_configured)} workflows={workflows} loading={health.isLoading} /> }
  ]} /></>;
}

function AccountsTab({ accounts, loading, pagination, actionCount }: { accounts: GoPayAccountProjection[]; loading?: boolean; pagination?: AccountListPagination; actionCount: number }) {
  const ready = accounts.filter((item) => item.account?.status?.value === 'ready').length;
  return <div className="grid gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_360px]"><AccountCarrierPanel title="GoPayAccount" carriers={accounts} loading={loading} loadingText="加载 GoPayAccount..." emptyText="暂无已持久化 GoPayAccount" pagination={pagination} config={{ icon: () => <WalletCards size={15} />, title: (record) => <span className="font-mono">{accountSubject(record) || record.key?.account_id}</span>, subtitle: (record) => record.key?.account_id || '', meta: (record) => <span className="text-xs text-muted-foreground">{record.status?.label || record.status?.value || '-'}</span> }} renderChildren={(carrier) => <GoPayAccountMeta account={carrier} />} /><div className="grid content-start gap-3"><Summary icon={<WalletCards size={16} />} label="账号" value={accounts.length} /><Summary icon={<Activity size={16} />} label="Ready" value={ready} /><Summary icon={<Workflow size={16} />} label="动作" value={actionCount} /></div></div>;
}

function GoPayAccountMeta({ account }: { account: GoPayAccountProjection }) {
  return <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground"><Badge variant="outline">{account.country_code || '+62'}</Badge>{account.phone && <Badge variant="secondary">{account.phone}</Badge>}{account.balance_currency && <Badge variant="outline">{account.balance_amount} {account.balance_currency}</Badge>}</div>;
}

function Summary({ icon, label, value }: { icon: ReactNode; label: string; value: number }) {
  return <Card><CardContent className="flex items-center gap-3 p-3"><div className="rounded-lg bg-primary/10 p-2 text-primary">{icon}</div><div><div className="text-xs text-muted-foreground">{label}</div><div className="text-lg font-semibold leading-none">{value}</div></div></CardContent></Card>;
}

function WorkflowTab({ configured, workflows, loading }: { configured: boolean; loading?: boolean; workflows: Array<{ key: string; label: string; webhook_path: string }> }) {
  return <div className="grid gap-4 p-4"><Alert><AlertTitle>{configured ? 'GoPay n8n 编排已接入' : 'GoPay n8n webhook 未配置'}</AlertTitle><AlertDescription>{loading ? '加载中...' : 'GoPayAccount 登录、注册、PIN、改号和注销流程由 gopay-app 拥有；GPT 仅保留 checkout/stripe/payment。'}</AlertDescription></Alert><div className="grid gap-3 md:grid-cols-2"><InfoCard icon={<KeyRound size={16} />} title="账户状态" badge="gopay-app" text="GoPayAccount 持久化、状态缓存、Profile 与账号动作都在 gopay-app 侧。" /><InfoCard icon={<RefreshCw size={16} />} title="OTP" badge="channel+target" text="WA/SMS OTP 统一按 channel + target + otp 投递，不再绑定 GPT 账户接口。" /></div><div className="grid gap-2">{workflows.map((item) => <div key={item.key} className="flex items-center justify-between rounded-xl border bg-card p-3 text-sm"><span>{item.label}</span><code className="text-xs text-muted-foreground">{item.webhook_path}</code></div>)}</div><Button variant="outline" asChild><a href="/workflow" target="_blank" rel="noreferrer">打开 Workflow 状态页</a></Button></div>;
}

function InfoCard({ icon, title, badge, text }: { icon: ReactNode; title: string; badge: string; text: string }) {
  return <Card><CardContent className="grid gap-2 p-4"><div className="flex items-center justify-between"><div className="flex items-center gap-2 font-medium">{icon}{title}</div><Badge variant="outline">{badge}</Badge></div><p className="text-sm text-muted-foreground">{text}</p></CardContent></Card>;
}
