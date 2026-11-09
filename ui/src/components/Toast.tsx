import { Toast, Toaster, ToastTitle, useToastController } from '@fluentui/react-components';
import { createContext, useCallback, useContext, useId, type ReactNode } from 'react';

const Context = createContext<(message: string) => void>(() => {});
export const useToast = () => useContext(Context);

export function ToastProvider({ children }: { children: ReactNode }) {
  const toasterId = useId();
  const { dispatchToast } = useToastController(toasterId);
  const notify = useCallback(
    (message: string) =>
      dispatchToast(
        <Toast>
          <ToastTitle>{message}</ToastTitle>
        </Toast>,
        {
          timeout: 5000,
          intent: 'info',
        },
      ),
    [dispatchToast, toasterId],
  );
  return (
    <Context.Provider value={notify}>
      {children}
      <Toaster toasterId={toasterId} position="bottom" limit={1} />
    </Context.Provider>
  );
}
