import { KeyRound, LogOut, PhoneForwarded, Play, Search, ShieldCheck, WalletCards } from 'lucide-react';
import type { AccountCatalogActionBase } from '@byte-v-forge/common-ui';
import { accountCarrierID } from '@byte-v-forge/common-ui';
import { GoPayAccountWorkflowOperation } from '../proto/gopay_app';
import type { GoPayAccountProjection } from './gopay-api';

export const GOPAY_ACTIONS = {
  signup: 'GOPAY_ACCOUNT_SIGNUP',
  login: 'GOPAY_ACCOUNT_LOGIN',
  ensurePin: 'GOPAY_ACCOUNT_ENSURE_PIN_SETUP',
  checkBalance: 'GOPAY_ACCOUNT_CHECK_BALANCE',
  checkPin: 'GOPAY_ACCOUNT_CHECK_PIN',
  changePhone: 'GOPAY_ACCOUNT_CHANGE_PHONE',
  deactivate: 'GOPAY_ACCOUNT_DEACTIVATE',
} as const;

export type GoPayActionID = (typeof GOPAY_ACTIONS)[keyof typeof GOPAY_ACTIONS];

export type GoPayAccountActionSpec = AccountCatalogActionBase<GoPayAccountProjection, GoPayActionID> & {
  id: string;
  operation: GoPayAccountWorkflowOperation;
  requiresPinInput?: boolean;
};

const canRun = (account: GoPayAccountProjection) => !!accountCarrierID(account);

export const GOPAY_ACCOUNT_PRIMARY_ACTIONS: GoPayAccountActionSpec[] = [
  {
    id: 'gopay-signup',
    actionID: GOPAY_ACTIONS.signup,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_SIGNUP,
    fallbackLabel: '注册',
    icon: <Play size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '使用当前账号手机号发起 GoPay 注册编排',
  },
  {
    id: 'gopay-login',
    actionID: GOPAY_ACTIONS.login,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_LOGIN,
    fallbackLabel: '登录',
    icon: <KeyRound size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '使用当前账号手机号发起 GoPay 登录编排',
  },
];

export const GOPAY_ACCOUNT_TOOL_ACTIONS: GoPayAccountActionSpec[] = [
  {
    id: 'gopay-check-balance',
    actionID: GOPAY_ACTIONS.checkBalance,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_CHECK_BALANCE,
    fallbackLabel: '查余额',
    icon: <WalletCards size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '登录后检查当前 GoPay 余额',
  },
  {
    id: 'gopay-check-pin',
    actionID: GOPAY_ACTIONS.checkPin,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_CHECK_PIN,
    fallbackLabel: '查 PIN',
    icon: <Search size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '读取账号状态并检查 PIN 配置状态',
  },
  {
    id: 'gopay-ensure-pin',
    actionID: GOPAY_ACTIONS.ensurePin,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_ENSURE_PIN_SETUP,
    fallbackLabel: 'PIN 设置',
    icon: <ShieldCheck size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '为注册后的 GoPay 账号设置 PIN',
    requiresPinInput: true,
  },
];

export const GOPAY_ACCOUNT_LIFECYCLE_ACTIONS: GoPayAccountActionSpec[] = [
  {
    id: 'gopay-change-phone',
    actionID: GOPAY_ACTIONS.changePhone,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_CHANGE_PHONE,
    fallbackLabel: '改绑手机号',
    icon: <PhoneForwarded size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '登录后执行 GoPay 改绑手机号编排',
  },
  {
    id: 'gopay-deactivate',
    actionID: GOPAY_ACTIONS.deactivate,
    operation: GoPayAccountWorkflowOperation.GOPAY_ACCOUNT_WORKFLOW_OPERATION_DEACTIVATE,
    fallbackLabel: '注销',
    icon: <LogOut size={14} />,
    allowed: canRun,
    disabledReason: '缺少 GoPayAccount ID',
    hint: '对当前 GoPay 账号执行注销编排',
  },
];
