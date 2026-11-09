import { Button, type ButtonProps } from '@mui/material';
import { Link, type LinkProps } from 'react-router-dom';

type Props = Pick<LinkProps, 'to' | 'state' | 'children'> & Pick<ButtonProps, 'variant'>;

export function ButtonLink({ to, state, children, variant = 'outlined' }: Props) {
  return (
    <Button component={Link} to={to} state={state} variant={variant}>
      {children}
    </Button>
  );
}
