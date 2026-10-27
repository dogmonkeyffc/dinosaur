import type { ComponentProps } from 'react';
import '../styles/select.css';

// Keep native labeling, form submission, keyboard navigation and mobile pickers.
export function Select({ className = '', ...props }: ComponentProps<'select'>) {
  return <select {...props} className={`select-control ${className}`.trim()} />;
}
