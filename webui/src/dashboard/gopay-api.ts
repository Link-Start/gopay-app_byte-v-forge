import { ACCOUNT_PAGE_SIZE, api, fetchAccountList } from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import type {
  GetGopayAccountProfileResponse,
  GetGopayActionCatalogResponse,
  GopayAccount,
  ListGopayAccountsResponse,
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

export const goPayKeys = {
  health: ['gopay', 'health'] as const,
  accounts: ['gopay', 'accounts'] as const,
  actionCatalog: ['gopay', 'action-catalog'] as const,
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
