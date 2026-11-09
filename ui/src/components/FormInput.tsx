import { InputAdornment, OutlinedInput } from '@mui/material';
import type { ComponentProps, ReactNode } from 'react';

type Props = Omit<ComponentProps<'input'>, 'size' | 'ref'> & {
  startAdornment?: ReactNode;
  endAdornment?: ReactNode;
};

export function FormInput({
  className,
  id,
  name,
  type,
  value,
  defaultValue,
  required,
  disabled,
  placeholder,
  autoComplete,
  readOnly,
  onChange,
  onFocus,
  startAdornment,
  endAdornment,
  ...htmlInput
}: Props) {
  return (
    <OutlinedInput
      fullWidth
      size="small"
      className={className}
      id={id}
      name={name}
      type={type}
      value={value}
      defaultValue={defaultValue}
      required={required}
      disabled={disabled}
      placeholder={placeholder}
      autoComplete={autoComplete}
      readOnly={readOnly}
      onChange={onChange}
      onFocus={onFocus}
      slotProps={{ input: htmlInput }}
      startAdornment={startAdornment && <InputAdornment position="start">{startAdornment}</InputAdornment>}
      endAdornment={endAdornment && <InputAdornment position="end">{endAdornment}</InputAdornment>}
    />
  );
}
