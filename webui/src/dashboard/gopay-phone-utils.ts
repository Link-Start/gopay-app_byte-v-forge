import { parsePhoneNumberFromString } from 'libphonenumber-js';
import type { GoPayPhoneCheckResponse } from './gopay-api';

export type GoPayResolvedPhone = {
  e164: string;
  phone: string;
  country_code: string;
};

export function resolveGoPayPhone(value: string, countryCallingCode: string): GoPayResolvedPhone | null {
  const raw = value.trim();
  const digits = value.replace(/\D+/g, '');
  const callingCode = countryCallingCode.replace(/\D+/g, '');
  if (!digits || !callingCode) return null;
  const parsed = parsePhoneNumberFromString(raw.startsWith('+') ? raw : `+${digits.startsWith(callingCode) ? digits : `${callingCode}${digits}`}`);
  if (!parsed?.countryCallingCode || !parsed.nationalNumber || parsed.countryCallingCode !== callingCode) return null;
  return { e164: parsed.number, phone: parsed.nationalNumber, country_code: `+${parsed.countryCallingCode}` };
}

export function goPayPhoneStatusLabel(result?: GoPayPhoneCheckResponse | null) {
  if (!result) return '未执行';
  if (result.status === 'available') return '未注册';
  if (result.status === 'registered') return '已注册';
  return result.status || result.error_message || '失败';
}

export function displayGoPayValue(value: unknown) {
  if (value === undefined || value === null || value === '') return '-';
  return String(value);
}

export function redactGoPayPhoneCheckResult(result?: GoPayPhoneCheckResponse | null) {
  return result || {};
}
