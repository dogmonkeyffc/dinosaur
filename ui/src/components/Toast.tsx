import { Snackbar } from '@mui/material';
import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';

const Context = createContext<(message: string) => void>(() => {});
export const useToast = () => useContext(Context);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [notice, setNotice] = useState({ message: '', key: 0 });
  const notify = useCallback((message: string) => {
    setNotice((previous) => ({ message, key: previous.key + 1 }));
  }, []);
  return (
    <Context.Provider value={notify}>
      {children}
      <Snackbar
        key={notice.key}
        open={Boolean(notice.message)}
        message={notice.message}
        autoHideDuration={5000}
        onClose={() => setNotice((previous) => ({ ...previous, message: '' }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      />
    </Context.Provider>
  );
}
