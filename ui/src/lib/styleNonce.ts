import { setNonce } from 'get-nonce';

export const styleNonce = document.querySelector<HTMLMetaElement>(
  'meta[name="dinosaur-style-nonce"]',
)?.content;

if (styleNonce) setNonce(styleNonce);
