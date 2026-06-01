import { useState, type ReactNode } from 'react';
import { KeyRound, RefreshCw, WalletCards } from 'lucide-react';
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Button,
  Card,
  CardContent,
  ToastMessage,
  WorkspaceTabbedPanel,
  accountCarrierID,
  useAccountManagementController,
  useAsyncActionRunner,
  useQuery,
  useToastMessage,
} from '@byte-v-forge/common-ui';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import type { GoPayAccountProjection, GoPayPhoneCheckResponse } from './gopay-api';
import { checkGoPayPhone, getGoPayActionCatalog, getGoPayHealth, goPayKeys } from './gopay-api';
import { GoPayAccountsTab, goPayAccountControllerOptions } from './gopay-accounts-tab';
import { GoPayPhoneCheckForm } from './gopay-phone-check-form';
import { GoPayPhoneResult } from './gopay-phone-result';
import type { GoPayResolvedPhone } from './gopay-phone-utils';

type GoPayTab = 'accounts' | 'toolbox' | 'workflows';

export function GoPayPage() {
  const toast = useToastMessage();
  const health = useQuery({ queryKey: goPayKeys.health, queryFn: getGoPayHealth, refetchInterval: 10000 });
  const actionCatalog = useQuery({ queryKey: goPayKeys.actionCatalog, queryFn: getGoPayActionCatalog, staleTime: 60000 });
  const accounts = useAccountManagementController<GoPayAccountProjection, ListGopayAccountsResponse>(goPayAccountControllerOptions);
  const [phone, setPhone] = useState('');
  const [result, setResult] = useState<GoPayPhoneCheckResponse | null>(null);
  const runner = useAsyncActionRunner();
  const workflows = health.data?.workflows || [{ key: 'gopay-account', label: 'GoPayAccount 编排', webhook_path: 'gopay-app/account' }];

  async function accountCreated(account: GoPayAccountProjection) {
    accounts.cacheAccount(account);
    accounts.setSelectedID(accountCarrierID(account));
    toast.showOK('GoPayAccount 已添加');
    await accounts.invalidate();
  }

  async function checkPhone(target: GoPayResolvedPhone) {
    setPhone(target.e164);
    setResult(null);
    await runner.tryRun('gopay-phone-check', async () => {
      const output = await checkGoPayPhone({ phone: target.phone, country_code: target.country_code });
      setResult(output);
      toast.showOK('GoPay 号码检测完成');
    }, { onError: toast.showError });
  }

  return (
    <>
      <ToastMessage toast={toast.toast} />
      <WorkspaceTabbedPanel<GoPayTab>
        defaultValue="accounts"
        title={<span className="inline-flex items-center gap-2"><WalletCards className="size-4" />GoPay</span>}
        meta={health.data?.n8n_webhook_configured ? 'n8n 已接入' : '等待 n8n'}
        tabs={[
          {
            value: 'accounts',
            label: '账号',
            content: (
              <GoPayAccountsTab
                controller={accounts}
                actionCatalog={actionCatalog.data}
                loadingCatalog={actionCatalog.isLoading}
                onCreated={accountCreated}
                onToast={toast.showToast}
                onError={toast.showError}
              />
            ),
          },
          {
            value: 'toolbox',
            label: '工具箱',
            content: <ToolboxTab phone={phone} result={result} busy={runner.busy} onCheck={checkPhone} onError={toast.showError} />,
          },
          {
            value: 'workflows',
            label: '工作流',
            content: <WorkflowTab configured={Boolean(health.data?.n8n_webhook_configured)} workflows={workflows} loading={health.isLoading} />,
          },
        ]}
      />
    </>
  );
}

function ToolboxTab(props: {
  phone: string;
  result: GoPayPhoneCheckResponse | null;
  busy: boolean;
  onCheck: (target: GoPayResolvedPhone) => void | Promise<void>;
  onError: (message: string) => void;
}) {
  const hasResult = props.busy || props.result || props.phone;
  return (
    <div className="p-3">
      <GoPayPhoneCheckForm
        disabled={props.busy}
        resultSlot={hasResult ? <GoPayPhoneResult phone={props.phone} result={props.result} loading={props.busy} /> : undefined}
        onCheck={props.onCheck}
        onError={props.onError}
      />
    </div>
  );
}

function WorkflowTab({ configured, workflows, loading }: {
  configured: boolean;
  loading?: boolean;
  workflows: Array<{ key: string; label: string; webhook_path: string }>;
}) {
  return (
    <div className="grid gap-4 p-4">
      <Alert>
        <AlertTitle>{configured ? 'GoPay n8n 编排已接入' : 'GoPay n8n webhook 未配置'}</AlertTitle>
        <AlertDescription>{loading ? '加载中...' : 'GoPayAccount 登录、注册、PIN、改号和注销流程由 gopay-app 拥有；号码检测为 gopay-app 直连接口。'}</AlertDescription>
      </Alert>
      <div className="grid gap-3 md:grid-cols-2">
        <InfoCard icon={<KeyRound size={16} />} title="账户状态" badge="gopay-app" text="GoPayAccount 持久化、状态缓存、Profile 与账号动作都在 gopay-app 侧。" />
        <InfoCard icon={<RefreshCw size={16} />} title="OTP" badge="channel+target" text="WA/SMS OTP 统一按 channel + target + otp 投递，不再绑定 GPT 账户接口。" />
      </div>
      <div className="grid gap-2">{workflows.map((item) => <div key={item.key} className="flex items-center justify-between rounded-xl border bg-card p-3 text-sm"><span>{item.label}</span><code className="text-xs text-muted-foreground">{item.webhook_path}</code></div>)}</div>
      <Button variant="outline" asChild><a href="/workflow" target="_blank" rel="noreferrer">打开 Workflow 状态页</a></Button>
    </div>
  );
}

function InfoCard({ icon, title, badge, text }: { icon: ReactNode; title: string; badge: string; text: string }) {
  return (
    <Card>
      <CardContent className="grid gap-2 p-4">
        <div className="flex items-center justify-between"><div className="flex items-center gap-2 font-medium">{icon}{title}</div><Badge variant="outline">{badge}</Badge></div>
        <p className="text-sm text-muted-foreground">{text}</p>
      </CardContent>
    </Card>
  );
}
