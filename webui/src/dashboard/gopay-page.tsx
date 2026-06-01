import { KeyRound, RefreshCw, WalletCards } from 'lucide-react';
import {
  AccountPhoneProbeToolbox,
  ToastMessage,
  WorkflowStatusPanel,
  WorkspaceTabbedPanel,
  accountCarrierID,
  useAccountProbeAction,
  useAccountManagementController,
  useQuery,
  useToastMessage,
} from '@byte-v-forge/common-ui';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import type { GoPayAccountProjection, GoPayPhoneCheckResponse } from './gopay-api';
import { checkGoPayPhone, getGoPayActionCatalog, getGoPayHealth, goPayKeys } from './gopay-api';
import { GoPayAccountsTab, goPayAccountControllerOptions } from './gopay-accounts-tab';
import { GoPayPhoneResult } from './gopay-phone-result';
import { resolveGoPayPhone, type GoPayResolvedPhone } from './gopay-phone-utils';

type GoPayTab = 'accounts' | 'toolbox' | 'workflows';

export function GoPayPage() {
  const toast = useToastMessage();
  const health = useQuery({ queryKey: goPayKeys.health, queryFn: getGoPayHealth, refetchInterval: 10000 });
  const actionCatalog = useQuery({ queryKey: goPayKeys.actionCatalog, queryFn: getGoPayActionCatalog, staleTime: 60000 });
  const accounts = useAccountManagementController<GoPayAccountProjection, ListGopayAccountsResponse>(goPayAccountControllerOptions);
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
            content: <ToolboxTab phone={phoneProbe.subject} result={phoneProbe.result} busy={phoneProbe.busy} onCheck={phoneProbe.run} onError={toast.showError} />,
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
  return (
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
