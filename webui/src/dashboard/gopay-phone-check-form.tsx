import type { ReactNode } from 'react';
import { Search } from 'lucide-react';
import { Button, Card, CardContent, Controller, Input, Label, useForm, type Control } from '@byte-v-forge/common-ui';
import { resolveGoPayPhone, type GoPayResolvedPhone } from './gopay-phone-utils';

type FormValues = { phone: string; country_calling_code: string };

export function GoPayPhoneCheckForm({ disabled, resultSlot, onCheck, onError }: {
  disabled?: boolean;
  resultSlot?: ReactNode;
  onCheck: (target: GoPayResolvedPhone) => void | Promise<void>;
  onError: (message: string) => void;
}) {
  const form = useForm<FormValues>({ defaultValues: { phone: '', country_calling_code: '' } });
  const submit = form.handleSubmit((values) => {
    const target = resolveGoPayPhone(values.phone, values.country_calling_code);
    if (!target) {
      onError('请输入手机号和国家拨号码，例如手机号 81234567890、拨号码 62。');
      return;
    }
    void onCheck(target);
  });
  return (
    <Card className="w-full">
      <CardContent className="p-3">
        <div className="flex flex-wrap items-end gap-2">
          <div className="mb-1.5 mr-1 min-w-[5.5rem] text-sm font-medium">GoPay 号码检测</div>
          <form className="flex shrink-0 flex-wrap items-end gap-2" onSubmit={submit}>
            <CompactInput control={form.control} name="country_calling_code" label="拨号码" placeholder="+62" className="w-[86px]" />
            <CompactInput control={form.control} name="phone" label="手机号" placeholder="81234567890" className="w-[180px] sm:w-[220px]" />
            <Button className="size-8" type="submit" size="icon" aria-label="检测 GoPay 号码状态" title="检测 GoPay 号码状态" disabled={disabled}>
              <Search size={16} />
            </Button>
          </form>
          <div className="min-h-[58px] min-w-[300px] flex-1 rounded-lg border bg-muted/20 p-2">
            {resultSlot || <div className="flex h-full items-center text-xs text-muted-foreground">结果：是否已注册</div>}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function CompactInput({ control, name, label, placeholder, className }: {
  control: Control<FormValues>;
  name: keyof FormValues;
  label: string;
  placeholder: string;
  className: string;
}) {
  return (
    <div className={className}>
      <Label className="mb-1 text-[11px] text-muted-foreground">{label}</Label>
      <Controller control={control} name={name} render={({ field }) => <Input {...field} value={field.value || ''} type="tel" inputMode={name === 'country_calling_code' ? 'numeric' : 'tel'} placeholder={placeholder} className="h-8" />} />
    </div>
  );
}
