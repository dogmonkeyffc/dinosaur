import { Button, type ButtonProps } from '@fluentui/react-components';
import { useHref, useLinkClickHandler, type LinkProps } from 'react-router-dom';

type Props = Pick<LinkProps, 'to' | 'state' | 'children'> & Pick<ButtonProps, 'appearance'>;

export function ButtonLink({ to, state, children, appearance }: Props) {
  const href = useHref(to);
  const follow = useLinkClickHandler<HTMLAnchorElement>(to, { state });
  return (
    <Button as="a" href={href} onClick={follow} appearance={appearance} className="button">
      {children}
    </Button>
  );
}
