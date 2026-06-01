import { OneTimeOTPSubmit, accountCarrierID } from '@byte-v-forge/common-ui';
import { submitGoPayManualOTP, type GoPayAccountProjection } from './gopay-api';

export function GoPayManualOTPSubmit({
  account,
  disabled,
  onSubmitted,
  onToast,
  onError,
}: {
  account: GoPayAccountProjection;
  disabled?: boolean;
  onSubmitted?: () => void | Promise<void>;
  onToast: (kind: 'error' | 'ok', message: string) => void;
  onError: (message: unknown) => void;
}) {
  const accountID = accountCarrierID(account);

  return (
    <OneTimeOTPSubmit
      title="OTP 兜底提交"
      subtitle="GoPay 流程等待 OTP 时可从这里一次性提交；只恢复当前等待流程，不写入 OTP 缓存或历史。"
      disabled={disabled || !accountID}
      input={{ 'aria-label': 'GoPay OTP', className: 'w-40', maxLength: 12 }}
      submit={{
        key: `gopay-manual-otp:${accountID}`,
        label: '提交 OTP',
        pendingLabel: '提交中',
        clearOnSuccess: true,
        onRun: async (otp) => {
          const resp = await submitGoPayManualOTP(account, otp);
          onToast('ok', resp.resume_count ? `已恢复 ${resp.resume_count} 个 GoPay OTP 流程` : 'GoPay OTP 已提交');
          await onSubmitted?.();
        },
      }}
      onError={onError}
    />
  );
}
