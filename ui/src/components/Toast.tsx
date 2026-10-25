import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

const Context = createContext<(message: string) => void>(() => {});
export const useToast = () => useContext(Context);
export function ToastProvider({ children }: { children: ReactNode }) {
  const [notice, setNotice] = useState('');
  useEffect(() => {
    if (!notice) return;
    const timer = window.setTimeout(() => setNotice(''), 5000);
    return () => clearTimeout(timer);
  }, [notice]);
  return (
    <Context.Provider value={setNotice}>
      {children}
      {notice && (
        <div className="toast" role="status">
          {notice}
        </div>
      )}
    </Context.Provider>
  );
}
