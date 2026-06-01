import { useState } from 'react';
import {
  ActionButtonGroup,
  ActionSection,
  Badge,
  KVList,
  accountActionButton,
  accountCarrierID,
  accountSubject,
  useQuery,
  type ActionButtonDescriptor,
} from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import { getGoPayAccountProfile, goPayKeys, type GoPayAccountProjection } from './gopay-api';
import {
  GOPAY_ACCOUNT_LIFECYCLE_ACTIONS,
  GOPAY_ACCOUNT_PRIMARY_ACTIONS,
  GOPAY_ACCOUNT_TOOL_ACTIONS,
  type GoPayAccountActionSpec,
} from './gopay-account-action-specs';
import { GoPayAccountPINDialog } from './gopay-account-pin-dialog';

export function GoPayAccountDetails({
  account,
  actionCatalog,
  busy,
  onRunAction,
}: {
  account: GoPayAccountProjection;
  actionCatalog?: AccountActionCatalog;
  busy?: boolean;
  onRunAction: (spec: GoPayAccountActionSpec, payload?: Record<string, unknown>) => Promise<boolean>;
}) {
  const [pinAction, setPinAction] = useState<GoPayAccountActionSpec | null>(null);
  const accountID = accountCarrierID(account);
  const profile = useQuery({
    queryKey: goPayKeys.profile(accountID),
    queryFn: () => getGoPayAccountProfile(accountID),
    enabled: !!accountID,
  });

  async function run(spec: GoPayAccountActionSpec, payload: Record<string, unknown> = {}) {
    if (spec.requiresPinInput && !payload.pin) {
      setPinAction(spec);
      return;
    }
    await onRunAction(spec, payload);
  }

  async function submitPIN(payload: Record<string, unknown>) {
    if (!pinAction) return;
    if (await onRunAction(pinAction, payload)) setPinAction(null);
  }

  return (
    <div className="details gopayAccountDetails">
      <section className="grid gap-2 rounded-xl border bg-card p-3">
        <div className="sectionTitle">
          <h3 className="font-mono text-sm">{account.account ? accountSubject(account.account) || accountID : accountID}</h3>
        </div>
        <GoPayAccountBadges account={account} pinConfigured={profile.data?.pin_configured} />
      </section>
      <ActionSection title="账户动作" description="和 GPT 一样，先选中账号，再在账号详情里执行账户级操作。">
        <ActionButtonGroup actions={buttons(GOPAY_ACCOUNT_PRIMARY_ACTIONS, actionCatalog, account, Boolean(busy), run)} className="sectionActions" />
      </ActionSection>
      <ActionSection title="状态检查">
        <ActionButtonGroup actions={buttons(GOPAY_ACCOUNT_TOOL_ACTIONS, actionCatalog, account, Boolean(busy), run)} className="sectionActions" />
      </ActionSection>
      <ActionSection title="账号生命周期">
        <ActionButtonGroup actions={buttons(GOPAY_ACCOUNT_LIFECYCLE_ACTIONS, actionCatalog, account, Boolean(busy), run)} className="sectionActions" />
      </ActionSection>
      <ActionSection title="账户详情">
        <KVList items={[
          { label: 'GoPayAccount ID', value: accountID, mono: true },
          { label: '手机号', value: formatPhone(account) },
          { label: '国家码', value: account.country_code || '-' },
          { label: 'OTP 渠道', value: account.otp_channel || '-' },
          { label: '状态', value: account.account?.status?.label || account.account?.status?.value || '-' },
          { label: 'PIN', value: pinState(profile.data?.pin_configured, profile.isLoading) },
          { label: '余额', value: account.balance_currency ? `${account.balance_amount} ${account.balance_currency}` : '-' },
        ]} />
      </ActionSection>
      <GoPayAccountPINDialog open={!!pinAction} busy={busy} onSubmit={submitPIN} onOpenChange={(open) => { if (!open) setPinAction(null); }} />
    </div>
  );
}

function GoPayAccountBadges({ account, pinConfigured }: { account: GoPayAccountProjection; pinConfigured?: boolean }) {
  return (
    <div className="flex flex-wrap gap-2">
      <Badge variant="outline">{account.account?.status?.label || account.account?.status?.value || 'created'}</Badge>
      <Badge variant={pinConfigured ? 'secondary' : 'outline'}>{pinConfigured ? 'PIN 已配置' : 'PIN 未配置'}</Badge>
      {account.otp_channel && <Badge variant="outline">{account.otp_channel}</Badge>}
    </div>
  );
}

function buttons(
  specs: GoPayAccountActionSpec[],
  catalog: AccountActionCatalog | undefined,
  account: GoPayAccountProjection,
  busy: boolean,
  run: (spec: GoPayAccountActionSpec) => void | Promise<void>,
): ActionButtonDescriptor[] {
  return specs.map((spec) => accountActionButton({ catalog, account, busy, placement: 'gopay' }, { ...spec, onClick: () => run(spec) }));
}

function formatPhone(account: GoPayAccountProjection) {
  if (!account.phone) return '-';
  return `${account.country_code || '+62'} ${account.phone}`;
}

function pinState(configured?: boolean, loading?: boolean) {
  if (loading) return '读取中';
  return configured ? '已配置' : '未配置';
}
