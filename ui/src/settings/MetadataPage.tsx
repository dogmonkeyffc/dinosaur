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
      setNotice(
        mode === 'test'
          ? '连接成功，凭据有效。测试不会保存配置。'
          : mode === 'clear'
            ? '凭据已移除。'
            : '配置已保存。',
      );
      if (mode !== 'test') resource.reload();
    });
  }
  return (
    <section id="metadata-section">
      <Heading title="元数据设置" />
      <div className="metadata-card">
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
              <Select value={type} onChange={(event) => setType(event.target.value)}>
                <option value="bearer">API Read Access Token（推荐）</option>
                <option value="api_key">API Key</option>
              </Select>
            </label>
            <label>
              凭据
              <input
                type="password"
                aria-label="凭据"
                value={credential}
                onChange={(event) => setCredential(event.target.value)}
                maxLength={4096}
                autoComplete="off"
                spellCheck={false}
                placeholder={
                  resource.data?.configured ? '已配置；留空保留，输入新凭据可替换' : '粘贴 TMDB 凭据'
                }
              />
            </label>
            <label>
              首选资料语言
              <Select value={language} onChange={(event) => setLanguage(event.target.value)}>
                <option value="zh-CN">简体中文</option>
                <option value="zh-TW">繁体中文（台湾）</option>
                <option value="zh-HK">繁体中文（香港）</option>
                <option value="en-US">英语</option>
              </Select>
            </label>
            <div className="metadata-actions">
              <button className="button primary">保存配置</button>
              <button className="button secondary" type="button" onClick={() => submit('test')}>
                测试连接
              </button>
              <button
                className="text-button"
                type="button"
                disabled={!resource.data?.configured}
                onClick={() => submit('clear')}
              >
                移除凭据
              </button>
            </div>
          </fieldset>
        </form>
        <p role="status">{notice}</p>
        <ErrorMessage error={resource.error || action.error || jobs.error} />
        {resource.error && (
          <button className="button secondary" onClick={resource.reload}>
            重新读取配置
          </button>
        )}
        <button
          className="button secondary"
          disabled={jobs.busy || jobs.running}
          onClick={() => jobs.start()}
        >
          刮削所有未识别影片
        </button>
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
          <strong>资料来源与致谢</strong>
          <p>This product uses the TMDB API but is not endorsed or certified by TMDB.</p>
        </div>
      </div>
    </section>
  );
}
