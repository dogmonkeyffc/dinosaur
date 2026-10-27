import { Children, isValidElement, useId, type ComponentProps, type ReactNode } from 'react';
import * as Primitive from '@radix-ui/react-select';
import { styleNonce } from '../lib/styleNonce';
import '../styles/select.css';

type Props = Omit<ComponentProps<typeof Primitive.Root>, 'children' | 'value' | 'name'> &
  Pick<
    ComponentProps<typeof Primitive.Trigger>,
    'id' | 'className' | 'aria-label' | 'aria-describedby' | 'aria-invalid'
  > & {
    value: string | number;
    children: ReactNode;
    name?: string;
  };

export function Select({ children, value, onValueChange, className = '', name, id, ...props }: Props) {
  const empty = useId();
  const options = Children.toArray(children).filter(
    (child) => isValidElement<ComponentProps<'option'>>(child) && child.type === 'option',
  );
  return (
    <Primitive.Root
      {...props}
      value={String(value) || empty}
      onValueChange={(next) => onValueChange?.(next === empty ? '' : next)}
    >
      {name && <input type="hidden" name={name} value={value} disabled={props.disabled} />}
      <Primitive.Trigger
        id={id}
        className={`select-control ${className}`.trim()}
        onPointerDown={(event) => {
          // Fieldsets disable native clicks, but pointerdown can still reach a trigger.
          if (event.currentTarget.matches(':disabled')) event.preventDefault();
        }}
        aria-label={props['aria-label']}
        aria-describedby={props['aria-describedby']}
        aria-invalid={props['aria-invalid']}
      >
        <span className="select-value">
          <Primitive.Value placeholder="请选择" />
        </span>
        <Primitive.Icon className="select-chevron">
          <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
            <path d="m4 6 4 4 4-4" />
          </svg>
        </Primitive.Icon>
      </Primitive.Trigger>
      <Primitive.Portal>
        <Primitive.Content className="select-menu" position="popper" sideOffset={6} collisionPadding={12}>
          <Primitive.ScrollUpButton className="select-scroll">⌃</Primitive.ScrollUpButton>
          <Primitive.Viewport className="select-options" nonce={styleNonce}>
            {options.map((option) => {
              if (!isValidElement<ComponentProps<'option'>>(option)) return null;
              const key = String(option.props.value ?? '');
              return (
                <Primitive.Item
                  key={key}
                  value={key || empty}
                  disabled={option.props.disabled}
                  className="select-option"
                >
                  <Primitive.ItemText>{option.props.children}</Primitive.ItemText>
                  <Primitive.ItemIndicator className="select-check">
                    <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
                      <path d="m3.5 8 3 3 6-6" />
                    </svg>
                  </Primitive.ItemIndicator>
                </Primitive.Item>
              );
            })}
          </Primitive.Viewport>
          <Primitive.ScrollDownButton className="select-scroll">⌄</Primitive.ScrollDownButton>
        </Primitive.Content>
      </Primitive.Portal>
    </Primitive.Root>
  );
}
