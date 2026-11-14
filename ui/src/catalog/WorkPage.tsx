import { Button, Paper } from '@mui/material';
import { useEffect, useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { ErrorMessage, Loading } from '../components/common';
import { useActiveLibrary } from '../layout/useActiveLibrary';
import { useResource } from '../lib/hooks';
import type { WorkDetail } from '../lib/types';
import { Artwork, DetailHeader } from './DetailParts';
import { MediaCard } from './MediaCard';
import { WorkEditor } from './WorkEditor';
import './works.css';

export function WorkPage() {
  const { id } = useParams();
  const location = useLocation();
  const { user } = useAuth();
  const result = useResource<WorkDetail>(`/api/works/${id}`);
  const [editing, setEditing] = useState(false);
  const work = result.data?.work;
  useActiveLibrary(work?.library_id);
  useEffect(() => {
    if (work) document.title = `${work.title} · Dinosaur`;
  }, [work]);
  return (
    <section>
      <DetailHeader series />
      <ErrorMessage error={result.error} />
      {result.error && <Button onClick={result.reload}>重新加载</Button>}
      {result.loading && !work && <Loading />}
      {work && result.data && (
        <Paper className="detail-layout">
          <Artwork url={work.poster_url} title={work.title} series />
          <article className="detail-copy">
            <h1>{work.title}</h1>
            <p>
              {work.member_count} 个篇章 · {work.season_count} 季 · {work.episode_count} 个文件 · 已看{' '}
              {work.watched_count}
            </p>
            {user?.role === 'admin' && <Button onClick={() => setEditing(true)}>编辑作品分组</Button>}
            <h2>篇章</h2>
            <div className="movie-grid">
              {result.data.members.map((member) => (
                <MediaCard
                  key={member.id}
                  item={{ ...member, work_id: undefined }}
                  from={location.pathname}
                />
              ))}
            </div>
          </article>
        </Paper>
      )}
      {editing && result.data && (
        <WorkEditor
          libraryID={result.data.work.library_id}
          detail={result.data}
          close={() => setEditing(false)}
          saved={() => {
            setEditing(false);
            result.reload();
          }}
        />
      )}
    </section>
  );
}
