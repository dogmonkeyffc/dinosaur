import { NativeSelect, OutlinedInput } from '@mui/material';
import type { ComponentProps } from 'react';

type Props = Omit<ComponentProps<'select'>, 'value' | 'onChange'> & {
  value: string | number;
  onValueChange: (value: string) => void;
};

export function Select({ value, onValueChange, children, id, name, className, ...inputProps }: Props) {
  return (
    <NativeSelect
      className={className}
      value={String(value)}
      onChange={(event) => onValueChange(event.target.value)}
      input={<OutlinedInput fullWidth size="small" />}
      variant="outlined"
      inputProps={{ ...inputProps, id, name }}
    >
      {children}
    </NativeSelect>
  );
}
