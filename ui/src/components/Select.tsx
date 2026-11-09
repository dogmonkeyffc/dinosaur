import { Select as FluentSelect, type SelectProps } from '@fluentui/react-components';

type Props = Omit<SelectProps, 'value' | 'onChange'> & {
  value: string | number;
  onValueChange: (value: string) => void;
};

export function Select({ value, onValueChange, className = '', ...props }: Props) {
  return (
    <FluentSelect
      {...props}
      className={`select-control ${className}`.trim()}
      value={String(value)}
      onChange={(_, data) => onValueChange(data.value)}
    />
  );
}
