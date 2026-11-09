import { Button, Popover, PopoverSurface, PopoverTrigger } from '@fluentui/react-components';
import { Icon } from './common';
import '../styles/info-tip.css';

export function InfoTip({ label, children }: { label: string; children: string }) {
  return (
    <Popover positioning="below-start">
      <PopoverTrigger disableButtonEnhancement>
        <Button
          type="button"
          appearance="subtle"
          size="small"
          className="info-tip"
          aria-label={label}
          icon={<Icon name="info" />}
        />
      </PopoverTrigger>
      <PopoverSurface className="info-tip-content" aria-label={label}>
        {children}
      </PopoverSurface>
    </Popover>
  );
}
