import { FormInput } from '../components/FormInput';
import { Button, Card } from '@mui/material';
import { Select } from '../components/Select';
import { useEffect, useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { ErrorMessage, Heading, Loading } from '../components/common';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import { jobDescription } from '../lib/format';
import { useJobs } from '../metadata/useJobs';

interface Settings {
  credential_type: string;
  language: string;
  configured: boolean;
}
export function MetadataPage() {
  const resource = useResource<Settings>('/api/admin/metadata/tmdb');
  const [type, setType] = useState('bearer');
  const [language, setLanguage] = useState('zh-CN');
  const [credential, setCredential] = useState('');
  const [notice, setNotice] = useState('');
  const action = useAction();
  const jobs = useJobs('movie');
  useEffect(() => {
    if (!resource.data) return;
    setType(resource.data.credential_type);
    setLanguage(resource.data.language);
    setCredential('');
  }, [resource.data]);
  function submit(mode: 'save' | 'test' | 'clear', event?: FormEvent) {
    event?.preventDefault();
    setNotice('');
    action.run(async () => {
      await api('/api/admin/metadata/tmdb' + (mode === 'test' ? '/test' : ''), {
        method: mode === 'test' ? 'POST' : mode === 'clear' ? 'DELETE' : 'PUT',
        body:
          mode === 'clear' ? undefined : { credential_type: type, language, credential: credential.trim() },
      });
      setNotice(mode === 'test' ? '连接成功' : mode === 'clear' ? '凭据已移除。' : '配置已保存。');
      if (mode !== 'test') resource.reload();
    });
  }
  return (
    <section id="metadata-section">
      <Heading title="元数据设置" />
      <Card className="metadata-card">
        <h2>TMDB</h2>
        {resource.loading && !resource.data ? (
          <Loading />
        ) : (
          <p role="status">{resource.data?.configured ? '已配置凭据' : '尚未配置凭据'}</p>
        )}
        <form onSubmit={(event) => submit('save', event)}>
          <fieldset disabled={!resource.data || resource.loading || action.busy}>
            <label>
              凭据类型
              <Select value={type} onValueChange={(value) => setType(value)}>
                <option value="bearer">API Read Access Token（推荐）</option>
                <option value="api_key">API Key</option>
              </Select>
            </label>
            <label htmlFor="tmdb-credential">
              凭据
              <FormInput
                className="form-control"
                id="tmdb-credential"
                type="password"
                value={credential}
                onChange={(event) => setCredential(event.target.value)}
                maxLength={4096}
                autoComplete="off"
                spellCheck={false}
                placeholder={resource.data?.configured ? '留空保留' : undefined}
              />
            </label>
            <label>
              首选资料语言
              <Select value={language} onValueChange={(value) => setLanguage(value)}>
                <option value="zh-CN">简体中文</option>
                <option value="zh-TW">繁体中文（台湾）</option>
                <option value="zh-HK">繁体中文（香港）</option>
                <option value="en-US">英语</option>
              </Select>
            </label>
            <div className="metadata-actions">
              <Button variant="contained" type="submit">
                保存配置
              </Button>
              <Button variant="outlined" type="button" onClick={() => submit('test')}>
                测试连接
              </Button>
              <Button
                variant="text"
                className="text-button"
                type="button"
                disabled={!resource.data?.configured}
                onClick={() => submit('clear')}
              >
                移除凭据
              </Button>
            </div>
          </fieldset>
        </form>
        <p role="status">{notice}</p>
        <ErrorMessage error={resource.error || action.error || jobs.error} />
        {resource.error && (
          <Button variant="outlined" onClick={resource.reload}>
            重新读取配置
          </Button>
        )}
        <Button variant="outlined" disabled={jobs.busy || jobs.running} onClick={() => jobs.start()}>
          刮削所有未识别影片
        </Button>
        {jobs.latest && (
          <p className="scrape-status" role="status">
            {jobDescription(jobs.latest)}
          </p>
        )}
        <p>
          <Link className="text-button" to="/settings/scraping">
            查看刮削管理
          </Link>
        </p>
        <div className="tmdb-attribution">
          <a href="https://www.themoviedb.org" target="_blank" rel="noopener noreferrer">
            <img src="/assets/vendor/tmdb-logo.svg" alt="TMDB" />
          </a>
          <p>This product uses the TMDB API but is not endorsed or certified by TMDB.</p>
        </div>
      </Card>
    </section>
  );
}
