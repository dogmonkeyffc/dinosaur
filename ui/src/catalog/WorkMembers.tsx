import { Button } from '@mui/material';
import { FormInput } from '../components/FormInput';
import type { Series } from '../lib/types';

export function WorkMembers({
  members,
  setMembers,
  seasons,
  setSeason,
}: {
  members: Series[];
  seasons: Record<number, string>;
  setSeason: (id: number, value: string) => void;
  setMembers: (members: Series[]) => void;
}) {
  function move(index: number, offset: number) {
    const next = [...members];
    [next[index], next[index + offset]] = [next[index + offset], next[index]];
    setMembers(next);
  }
  return (
    <div>
      <h3>篇章</h3>
      <ol className="work-candidates">
        {members.map((item, index) => (
          <li key={item.id}>
            <span>
              {item.title}
              {item.episode_count === 0 ? ' · 文件暂不可用' : ''}
            </span>
            <label className="work-member-season">
              季号
              <FormInput
                inputMode="numeric"
                maxLength={3}
                value={seasons[item.id] ?? ''}
                aria-label={`${item.title} 季号`}
                onChange={(event) => setSeason(item.id, event.target.value)}
              />
            </label>
            <div>
              <Button
                aria-label={`上移 ${item.title}`}
                disabled={index === 0}
                onClick={() => move(index, -1)}
              >
                上移
              </Button>
              <Button
                aria-label={`下移 ${item.title}`}
                disabled={index === members.length - 1}
                onClick={() => move(index, 1)}
              >
                下移
              </Button>
              <Button
                aria-label={`移出 ${item.title}`}
                onClick={() => setMembers(members.filter((member) => member.id !== item.id))}
              >
                移出
              </Button>
            </div>
          </li>
        ))}
      </ol>
    </div>
  );
}
