import { KeyRound, MessageCircle, RefreshCw, WalletCards } from 'lucide-react';
import {
  AccountPhoneProbeToolbox,
  Button,
  ToastMessage,
  WorkflowStatusPanel,
  WorkspaceTabbedPanel,
  accountCarrierID,
  useAccountProbeAction,
  useAccountManagementController,
  useMutation,
  useQuery,
  useQueryClient,
  useToastMessage,
} from '@byte-v-forge/common-ui';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import type { GoPayAccountProjection, GoPayPhoneCheckResponse } from './gopay-api';
import { checkGoPayPhone, getGoPayActionCatalog, getGoPayHealth, getGoPaySettings, goPayKeys, saveGoPaySettings, startGoPayRegisterIndonesiaWA } from './gopay-api';
import { GoPayAccountsTab, goPayAccountControllerOptions } from './gopay-accounts-tab';
import { GoPayPhoneResult } from './gopay-phone-result';
import { resolveGoPayPhone, type GoPayResolvedPhone } from './gopay-phone-utils';
import { GoPaySettingsTab } from './gopay-settings-tab';

type GoPayTab = 'accounts' | 'toolbox' | 'settings' | 'workflows';

export function GoPayPage() {
  const toast = useToastMessage();
  const queryClient = useQueryClient();
  const health = useQuery({ queryKey: goPayKeys.health, queryFn: getGoPayHealth, refetchInterval: 10000 });
  const actionCatalog = useQuery({ queryKey: goPayKeys.actionCatalog, queryFn: getGoPayActionCatalog, staleTime: 60000 });
  const settings = useQuery({ queryKey: goPayKeys.settings, queryFn: getGoPaySettings });
  const accounts = useAccountManagementController<GoPayAccountProjection, ListGopayAccountsResponse>(goPayAccountControllerOptions);
  const saveSettings = useMutation({
    mutationFn: saveGoPaySettings,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: goPayKeys.settings });
      toast.showOK('GoPay 配置已保存');
    },
    onError: toast.showError
  });
  const registerIndonesiaWA = useMutation({
    mutationFn: startGoPayRegisterIndonesiaWA,
    onSuccess: (resp) => toast.showOK(`印尼 WA 注册流程已启动：${resp.job_id}`),
    onError: toast.showError
  });
  const phoneProbe = useAccountProbeAction<GoPayResolvedPhone, GoPayPhoneCheckResponse>({
    actionKey: 'gopay-phone-check',
    subjectOf: (target) => target.e164,
    probe: (target) => checkGoPayPhone({ phone: target.phone, country_code: target.country_code }),
    onSuccess: () => toast.showOK('GoPay 号码检测完成'),
    onError: toast.showError,
  });
  const workflows = health.data?.workflows || [{ key: 'gopay-account', label: 'GoPayAccount 编排', webhook_path: 'gopay-app/account' }];

  async function accountCreated(account: GoPayAccountProjection) {
    accounts.cacheAccount(account);
    accounts.setSelectedID(accountCarrierID(account));
    toast.showOK('GoPayAccount 已添加');
    await accounts.invalidate();
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
            content: <ToolboxTab phone={phoneProbe.subject} result={phoneProbe.result} busy={phoneProbe.busy} registerBusy={registerIndonesiaWA.isPending} onRegisterIndonesiaWA={() => registerIndonesiaWA.mutate()} onCheck={phoneProbe.run} onError={toast.showError} />,
          },
          {
            value: 'settings',
            label: '配置',
            content: <GoPaySettingsTab settings={settings.data} loading={settings.isLoading} saving={saveSettings.isPending} onSave={(values) => saveSettings.mutate(values)} />,
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
  registerBusy: boolean;
  onCheck: (target: GoPayResolvedPhone) => void | Promise<void>;
  onRegisterIndonesiaWA: () => void;
  onError: (message: string) => void;
}) {
  return (
    <div className="grid gap-3">
      <section className="rounded-xl border border-border/70 bg-background p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 className="m-0 text-sm font-semibold">注册印尼 WA</h3>
            <p className="m-0 mt-1 text-xs text-muted-foreground">固定 whatsapp / ID / +62，从 SMS 取号后并行检测 WA 与 GoPay 未注册。</p>
          </div>
          <Button disabled={props.registerBusy} onClick={props.onRegisterIndonesiaWA}>
            <MessageCircle size={14} />
            启动注册
          </Button>
        </div>
      </section>
      <AccountPhoneProbeToolbox<GoPayResolvedPhone, GoPayPhoneCheckResponse>
        title="GoPay 号码检测"
        subject={props.phone}
        result={props.result}
        busy={props.busy}
        emptyResultText="结果：是否已注册"
        countryPlaceholder="+62"
        phonePlaceholder="81234567890"
        actionLabel="检测 GoPay 号码状态"
        resolve={(values) => ({ target: resolveGoPayPhone(values.phone, values.country_calling_code), error: '请输入手机号和国家拨号码，例如手机号 81234567890、拨号码 62。' })}
        renderResult={({ subject, result, loading }) => <GoPayPhoneResult phone={subject} result={result} loading={loading} />}
        onSubmit={props.onCheck}
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
    <WorkflowStatusPanel
      configured={configured}
      loading={loading}
      configuredTitle="GoPay n8n 编排已接入"
      unconfiguredTitle="GoPay n8n webhook 未配置"
      description="GoPayAccount 登录、注册、PIN、改号和注销流程由 gopay-app 拥有；号码检测为 gopay-app 直连接口。"
      cards={[
        {
          id: 'account-state',
          icon: <KeyRound size={16} />,
          title: '账户状态',
          badge: 'gopay-app',
          text: 'GoPayAccount 持久化、状态缓存、Profile 与账号动作都在 gopay-app 侧。',
        },
        {
          id: 'otp',
          icon: <RefreshCw size={16} />,
          title: 'OTP',
          badge: 'channel+account',
          text: 'WA/SMS 自动 OTP 按 channel + target 投递；账号详情提供一次性手动兜底，不写入缓存或历史。',
        },
      ]}
      workflows={workflows}
    />
  );
}
