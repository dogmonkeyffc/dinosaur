import { Button } from '@mui/material';
import { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { Select } from '../components/Select';
import { useLibraries } from '../libraries/LibraryProvider';
import { WorkEditor } from './WorkEditor';
import './works.css';

export function CreateWork({ libraryID }: { libraryID: number }) {
  const { user } = useAuth();
  const { libraries } = useLibraries();
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState(libraryID);
  const choices = libraries.filter((item) => item.type === 'tv');
  const library = choices.find((item) => item.id === (libraryID || selected)) || choices[0];
  if (user?.role !== 'admin' || !library) return null;
  return (
    <>
      {!libraryID && (
        <label>
          媒体库
          <Select value={library.id} onValueChange={(value) => setSelected(Number(value))}>
            {choices.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </Select>
        </label>
      )}
      <Button onClick={() => setOpen(true)}>新建作品分组</Button>
      {open && (
        <WorkEditor
          libraryID={library.id}
          close={() => setOpen(false)}
          saved={(work) => {
            setOpen(false);
            navigate(`/works/${work.id}`, { state: { from: location.pathname + location.search } });
          }}
        />
      )}
    </>
  );
}
