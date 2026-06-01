import { ACCOUNT_PAGE_SIZE, accountCarrierID, api, fetchAccountList } from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import type {
  DeleteGopayAccountResponse,
  GetGopayAccountProfileResponse,
  GetGopayActionCatalogResponse,
  GetGoPaySettingsResponse,
  GoPayRegisterIndonesiaWASettings,
  GopayAccount,
  ListGopayAccountsResponse,
  SaveGoPaySettingsResponse,
  StartGoPayRegisterIndonesiaWAWorkflowResponse,
} from '../proto/gopay_app';

export type GoPayAccountProjection = GopayAccount;
export type GoPayOTPChannel = 'wa' | 'sms';

export type GoPayHealthResponse = {
  success?: boolean;
  ok?: boolean;
  service?: string;
  n8n_webhook_configured?: boolean;
  workflows?: Array<{ key: string; label: string; webhook_path: string }>;
};

export type CreateGoPayAccountRequest = {
  phone: string;
  otp_channel: GoPayOTPChannel;
  country_code: string;
};

export type CreateGoPayAccountResponse = {
  success?: boolean;
  error_message?: string;
  account?: GoPayAccountProjection;
};

export type GoPayPhoneCheckRequest = {
  phone: string;
  country_code: string;
};

export type GoPayPhoneCheckResponse = {
  success?: boolean;
  available?: boolean;
  phone?: string;
  country_code?: string;
  status?: string;
  error_message?: string;
  proxy_hash?: string;
  device_fingerprint?: string;
  generated_proxy_state?: boolean;
};

export type SubmitGoPayOTPResponse = {
  success?: boolean;
  manual_once?: boolean;
  gopay_account_id?: string;
  resume_count?: number;
  resumed_job_ids?: string[];
  error_message?: string;
};

export const goPayKeys = {
  health: ['gopay', 'health'] as const,
  accounts: ['gopay', 'accounts'] as const,
  actionCatalog: ['gopay', 'action-catalog'] as const,
  settings: ['gopay', 'settings'] as const,
  profile: (accountID: string) => ['gopay', 'profile', accountID] as const,
};

export function getGoPayHealth() {
  return api<GoPayHealthResponse>('/api/gopay/health');
}

export function getGoPayAccounts(cursor = '') {
  return fetchAccountList<GoPayAccountProjection, ListGopayAccountsResponse>({
    path: '/api/gopay/accounts',
    cursor,
    limit: ACCOUNT_PAGE_SIZE
  });
}

export async function createGoPayAccount(req: CreateGoPayAccountRequest) {
  const resp = await api<CreateGoPayAccountResponse>('/api/gopay/accounts', {
    method: 'POST',
    body: JSON.stringify(req)
  });
  if (resp.success === false || resp.error_message) throw new Error(resp.error_message || 'create GoPayAccount failed');
  if (!resp.account) throw new Error('GoPayAccount response is empty');
  return resp.account;
}

export async function deleteGoPayAccount(account: GoPayAccountProjection | string) {
  const accountID = typeof account === 'string' ? account : accountCarrierID(account);
  if (!accountID) throw new Error('gopay_account_id is required');
  const resp = await api<DeleteGopayAccountResponse>(`/api/gopay/accounts/${encodeURIComponent(accountID)}`, { method: 'DELETE' });
  if (!resp.success || resp.error_message) throw new Error(resp.error_message || 'delete GoPayAccount failed');
  return resp;
}

export async function getGoPayActionCatalog(): Promise<AccountActionCatalog | undefined> {
  const resp = await api<GetGopayActionCatalogResponse>('/api/gopay/action-catalog');
  if (!resp.success) throw new Error(resp.error_message || 'load GoPay action catalog failed');
  return resp.catalog;
}

export async function getGoPayAccountProfile(accountID: string) {
  const resp = await api<GetGopayAccountProfileResponse>(`/api/gopay/profile?gopay_account_id=${encodeURIComponent(accountID)}`);
  if (!resp.success) throw new Error(resp.error_message || 'load GoPay account profile failed');
  return resp;
}

export async function checkGoPayPhone(req: GoPayPhoneCheckRequest) {
  return api<GoPayPhoneCheckResponse>('/api/gopay/phone/check', {
    method: 'POST',
    body: JSON.stringify(req)
  });
}

export async function getGoPaySettings() {
  const resp = await api<GetGoPaySettingsResponse>('/api/gopay/settings');
  if (!resp.success || resp.error_message) throw new Error(resp.error_message || 'load GoPay settings failed');
  return resp.register_indonesia_wa;
}

export async function saveGoPaySettings(settings: GoPayRegisterIndonesiaWASettings) {
  const resp = await api<SaveGoPaySettingsResponse>('/api/gopay/settings', {
    method: 'POST',
    body: JSON.stringify({ register_indonesia_wa: settings })
  });
  if (!resp.success || resp.error_message) throw new Error(resp.error_message || 'save GoPay settings failed');
  return resp.register_indonesia_wa;
}

export async function startGoPayRegisterIndonesiaWA() {
  const resp = await api<StartGoPayRegisterIndonesiaWAWorkflowResponse>('/api/gopay/workflows/register-indonesia-wa', {
    method: 'POST',
    body: JSON.stringify({})
  });
  if (!resp.started || resp.error_message) throw new Error(resp.error_message || 'start Indonesia WA registration failed');
  return resp;
}

export async function submitGoPayManualOTP(account: GoPayAccountProjection, otp: string) {
  const accountID = accountCarrierID(account);
  const resp = await api<SubmitGoPayOTPResponse>('/api/gopay/otp/submit', {
    method: 'POST',
    body: JSON.stringify({
      gopay_account_id: accountID,
      channel: account.otp_channel || 'wa',
      target: accountID,
      otp,
      otp_source: 'manual_frontend',
      manual_once: true
    })
  });
  if (resp.success === false || resp.error_message) {
    throw new Error(resp.error_message || 'submit GoPay OTP failed');
  }
  return resp;
}
