import { useEffect, useMemo, useRef, useState } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Copy, Loader2, Sparkles, X } from 'lucide-react';
import clsx from 'clsx';

import { fmtRelative } from '../../lib/format';
import { inspireApi } from '../../lib/services';
import { toast } from '../../stores/toast';
import type { InspireItem } from '../../lib/types';

type FilterKind = 'all' | 'image' | 'video';

const FILTERS: { value: FilterKind; label: string }[] = [
  { value: 'all', label: '全部' },
  { value: 'image', label: '图片' },
  { value: 'video', label: '视频' },
];

export default function InspirePage() {
  const [kind, setKind] = useState<FilterKind>('all');
  const [active, setActive] = useState<InspireItem | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const feed = useInfiniteQuery({
    queryKey: ['inspire.feed', kind],
    queryFn: ({ pageParam }) =>
      inspireApi.feed({
        kind: kind === 'all' ? undefined : kind,
        cursor: pageParam || undefined,
        page_size: 30,
      }),
    initialPageParam: 0 as number,
    getNextPageParam: (last) => last.next_cursor || undefined,
  });

  const items = useMemo(
    () => (feed.data?.pages ?? []).flatMap((p) => p.list),
    [feed.data?.pages],
  );

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const io = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting && feed.hasNextPage && !feed.isFetchingNextPage) {
            void feed.fetchNextPage();
          }
        }
      },
      { rootMargin: '600px 0px' },
    );
    io.observe(el);
    return () => io.disconnect();
  }, [feed]);

  return (
    <div className="mx-auto min-h-screen w-full max-w-[1440px] px-6 pb-20 pt-10 sm:px-10 lg:px-16">
      <header className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="mb-1.5 flex items-center gap-2 text-[11px] uppercase tracking-[0.2em] text-neutral-500">
            <Sparkles size={12} />
            <span>Inspire</span>
          </div>
          <h1 className="text-[28px] leading-tight tracking-tight">
            来自社区的<span className="neon-text">灵感</span>
          </h1>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {FILTERS.map((f) => (
            <button
              key={f.value}
              type="button"
              onClick={() => setKind(f.value)}
              className={clsx('inspire-chip', kind === f.value && 'inspire-chip--active')}
            >
              {f.label}
            </button>
          ))}
        </div>
      </header>

      {feed.isLoading ? (
        <div className="flex items-center justify-center py-24 text-neutral-400">
          <Loader2 size={20} className="mr-2 animate-spin" />
          正在拉取灵感
        </div>
      ) : items.length === 0 ? (
        <div className="grid place-items-center rounded-[24px] border border-dashed border-white/[0.08] py-24 text-neutral-400">
          <Sparkles size={28} />
          <p className="mt-2 text-sm">暂无作品，去创作页生成第一张吧</p>
        </div>
      ) : (
        <div
          className="columns-1 gap-5 sm:columns-2 lg:columns-3 xl:columns-4"
          style={{ columnWidth: '300px' }}
        >
          {items.map((item, idx) => (
            <InspireCard
              key={item.result_id}
              item={item}
              order={idx}
              onOpen={() => setActive(item)}
            />
          ))}
        </div>
      )}

      <div ref={sentinelRef} className="h-12" />
      {feed.isFetchingNextPage && (
        <div className="flex items-center justify-center py-6 text-sm text-neutral-400">
          <Loader2 size={16} className="mr-2 animate-spin" />
          加载更多
        </div>
      )}

      {active && <InspireDrawer item={active} onClose={() => setActive(null)} />}
    </div>
  );
}

function InspireCard({
  item,
  order,
  onOpen,
}: {
  item: InspireItem;
  order: number;
  onOpen: () => void;
}) {
  const ratio = useMemo(() => {
    if (item.width && item.height) return `${item.width} / ${item.height}`;
    return item.kind === 'video' ? '16 / 9' : '1 / 1';
  }, [item.width, item.height, item.kind]);
  const cover = item.thumb_url || item.url;
  const isVideo = item.kind === 'video';
  const promptShort = compactPrompt(item.prompt);
  const riseClass =
    order < 5 ? `neon-rise neon-rise-${order + 1}` : 'neon-rise';

  return (
    <button
      type="button"
      onClick={onOpen}
      className={clsx('inspire-card mb-3 block w-full break-inside-avoid text-left', riseClass)}
      style={{ aspectRatio: ratio }}
    >
      {isVideo ? (
        item.thumb_url ? (
          <img src={cover} alt="" loading="lazy" className="h-full w-full object-cover" />
        ) : (
          <video
            src={item.url}
            muted
            playsInline
            preload="metadata"
            className="h-full w-full object-cover"
          />
        )
      ) : (
        <img src={cover} alt="" loading="lazy" className="h-full w-full object-cover" />
      )}
      <span className="absolute left-2 top-2 rounded-full bg-black/55 px-2 py-0.5 text-[11px] text-white">
        {isVideo ? '视频' : '图片'}
      </span>
      <div className="inspire-card__overlay">
        {promptShort && <p className="line-clamp-2 text-[13px] leading-snug">{promptShort}</p>}
        <p className="mt-1.5 text-[11px] text-white/70">@{item.author} · {fmtRelative(item.created_at)}</p>
      </div>
    </button>
  );
}

function InspireDrawer({ item, onClose }: { item: InspireItem; onClose: () => void }) {
  const navigate = useNavigate();
  const isVideo = item.kind === 'video';

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const usePrompt = () => {
    if (item.prompt) {
      sessionStorage.setItem('image2api:inspire-prompt', item.prompt);
    }
    navigate(isVideo ? '/create/video' : '/create/image');
  };

  const copyPrompt = async () => {
    if (!item.prompt) return;
    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(item.prompt);
        toast.success('提示词已复制');
        return;
      } catch {
        // fall through to legacy fallback
      }
    }
    try {
      const ta = document.createElement('textarea');
      ta.value = item.prompt;
      ta.setAttribute('readonly', '');
      ta.style.position = 'fixed';
      ta.style.top = '0';
      ta.style.left = '0';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.focus();
      ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      if (ok) toast.success('提示词已复制');
      else toast.error('复制失败');
    } catch {
      toast.error('复制失败');
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/72 backdrop-blur-sm"
      onMouseDown={onClose}
    >
      <div
        className="relative flex h-full w-full max-w-[960px] flex-col overflow-y-auto overscroll-contain border-l border-white/10 bg-[#07070F] lg:overflow-hidden"
        onMouseDown={(e) => e.stopPropagation()}
        onTouchStart={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-4 top-4 z-10 grid h-9 w-9 place-items-center rounded-full border border-white/10 bg-white/5 text-white transition hover:bg-white/10"
          title="关闭"
        >
          <X size={18} />
        </button>
        <div className="flex flex-col lg:grid lg:flex-1 lg:grid-cols-[1.4fr_1fr] lg:overflow-hidden">
          <div className="grid place-items-center bg-black/40 p-4">
            {isVideo ? (
              <video
                src={item.url}
                controls
                playsInline
                autoPlay
                className="max-h-[60vh] max-w-full rounded-[12px] bg-black lg:max-h-full"
              />
            ) : (
              <img
                src={item.url}
                alt=""
                className="max-h-[60vh] max-w-full rounded-[12px] object-contain lg:max-h-full"
              />
            )}
          </div>
          <div className="flex flex-col gap-4 border-t border-white/10 p-6 lg:overflow-y-auto lg:border-l lg:border-t-0">
            <div>
              <div className="mb-1 flex items-center gap-2 text-xs uppercase tracking-wide text-neutral-400">
                <Sparkles size={12} /> Prompt
              </div>
              <p className="whitespace-pre-wrap break-words text-[14px] leading-relaxed text-neutral-200">
                {item.prompt || '（无提示词）'}
              </p>
            </div>
            <div className="grid grid-cols-2 gap-2 text-[13px] text-neutral-400">
              <Meta label="作者" value={`@${item.author}`} />
              <Meta label="模型" value={item.model} />
              <Meta label="类型" value={isVideo ? '视频' : '图片'} />
              <Meta label="时间" value={fmtRelative(item.created_at)} />
            </div>
            <div className="mt-auto flex gap-2">
              <button type="button" onClick={usePrompt} className="btn-neon flex-1">
                <Sparkles size={16} /> 用这个提示词
              </button>
              <button
                type="button"
                onClick={copyPrompt}
                disabled={!item.prompt}
                className="inline-flex h-10 items-center justify-center gap-2 rounded-[12px] border border-white/15 bg-white/5 px-4 text-sm text-white transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-50"
                title="复制提示词"
              >
                <Copy size={16} /> 复制
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-[11px] uppercase tracking-wide text-neutral-500">{label}</div>
      <div className="mt-0.5 text-neutral-200">{value}</div>
    </div>
  );
}

function compactPrompt(prompt?: string) {
  return String(prompt || '').replace(/\s+/g, ' ').trim();
}
