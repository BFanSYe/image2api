import { useEffect, useRef, useState } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import {
  BookOpen,
  ChevronDown,
  Clock3,
  CreditCard,
  FileKey2,
  Gift,
  Image,
  LogIn,
  LogOut,
  Menu,
  MessageCircle,
  Settings,
  Sparkles,
  Video,
  X,
  type LucideIcon,
} from 'lucide-react';
import clsx from 'clsx';

import { useAuthStore } from '../stores/auth';
import { useLoginGateStore } from '../stores/loginGate';
import { toast } from '../stores/toast';

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  authed?: boolean;
}

const PRIMARY_NAV: NavItem[] = [
  { to: '/create/image', label: '图片', icon: Image },
  { to: '/create/text', label: '文字', icon: MessageCircle },
  { to: '/create/video', label: '视频', icon: Video },
  { to: '/inspire', label: '灵感', icon: Sparkles },
];

const SECONDARY_NAV: NavItem[] = [
  { to: '/history', label: '历史', icon: Clock3, authed: true },
  { to: '/docs', label: '文档', icon: BookOpen },
];

const ACCOUNT_MENU: NavItem[] = [
  { to: '/billing', label: '充值', icon: CreditCard, authed: true },
  { to: '/keys', label: '密钥', icon: FileKey2, authed: true },
  { to: '/invite', label: '邀请', icon: Gift, authed: true },
  { to: '/settings', label: '设置', icon: Settings, authed: true },
];

const APP_VERSION = 'v2.0.0';

export function AppLayout() {
  const token = useAuthStore((s) => s.token);
  const me = useAuthStore((s) => s.me);
  const logout = useAuthStore((s) => s.logout);
  const openGate = useLoginGateStore((s) => s.openGate);
  const navigate = useNavigate();
  const isAuthed = !!token;
  const [drawerOpen, setDrawerOpen] = useState(false);

  const onLogout = async () => {
    await logout();
    toast.info('已退出登录');
    navigate('/create/image', { replace: true });
  };

  const handleNav = (item: NavItem, e: React.MouseEvent) => {
    if (item.authed && !isAuthed) {
      e.preventDefault();
      openGate({ hint: `登录后即可使用「${item.label}」`, onLoggedIn: () => navigate(item.to) });
    }
  };

  return (
    <div className="min-h-full">
      <header className="sticky top-0 z-30 border-b border-white/[0.06] bg-[#07070F]/85 backdrop-blur-xl">
        <div className="mx-auto flex h-14 w-full max-w-[1440px] items-center gap-4 px-4 sm:px-6 lg:px-8">
          <button
            type="button"
            onClick={() => navigate('/create/image')}
            className="flex shrink-0 items-center gap-2 text-[15px] font-medium tracking-tight"
            title="首页"
          >
            <span className="grid h-7 w-7 place-items-center rounded-lg bg-gradient-to-br from-[#7C3AED] to-[#22D3EE] text-white shadow-[0_0_18px_rgba(124,58,237,.45)]">
              <Sparkles size={15} />
            </span>
            <span className="neon-text">image2api</span>
          </button>

          <nav className="hidden flex-1 items-center gap-0.5 lg:flex">
            {PRIMARY_NAV.map((item) => (
              <TopNavLink key={item.to} item={item} onClick={handleNav} />
            ))}
            <span className="mx-2 h-5 w-px bg-white/[0.10]" aria-hidden />
            {SECONDARY_NAV.map((item) => (
              <TopNavLink key={item.to} item={item} onClick={handleNav} />
            ))}
          </nav>

          <div className="ml-auto flex items-center gap-2 lg:ml-0">
            {isAuthed && me && typeof me.points === 'number' && (
              <button
                type="button"
                onClick={() => navigate('/billing')}
                className="hidden h-8 items-center gap-1.5 rounded-full border border-white/[0.10] bg-white/[0.04] px-3 text-[13px] text-neutral-200 transition hover:border-white/[0.20] hover:bg-white/[0.08] sm:inline-flex"
                title="点数余额"
              >
                <CreditCard size={13} className="text-[#C4B5FD]" />
                {(me.points / 100).toFixed(2)}
              </button>
            )}
            {isAuthed ? (
              <AccountMenu me={me} onLogout={onLogout} onNav={handleNav} navigate={navigate} />
            ) : (
              <button
                type="button"
                onClick={() => openGate({ hint: '登录后可保存作品和查看额度' })}
                className="inline-flex h-8 items-center gap-1.5 rounded-full border border-white/[0.10] bg-white/[0.04] px-3 text-[13px] text-neutral-200 transition hover:border-white/[0.20] hover:bg-white/[0.08]"
              >
                <LogIn size={14} />
                登录
              </button>
            )}
            <button
              type="button"
              onClick={() => setDrawerOpen(true)}
              className="grid h-8 w-8 place-items-center rounded-full border border-white/[0.10] bg-white/[0.04] text-neutral-300 transition hover:border-white/[0.20] hover:bg-white/[0.08] lg:hidden"
              title="菜单"
            >
              <Menu size={16} />
            </button>
          </div>
        </div>
        <div className="neon-bar opacity-60" />
      </header>

      <main className="min-h-[calc(100vh-3.5rem)]">
        <Outlet />
      </main>

      {drawerOpen && (
        <MobileDrawer
          isAuthed={isAuthed}
          onClose={() => setDrawerOpen(false)}
          onNav={handleNav}
          onLogout={onLogout}
          openGate={openGate}
          version={APP_VERSION}
        />
      )}
    </div>
  );
}

function TopNavLink({
  item,
  onClick,
}: {
  item: NavItem;
  onClick: (item: NavItem, e: React.MouseEvent) => void;
}) {
  const Icon = item.icon;
  return (
    <NavLink
      to={item.to}
      onClick={(e) => onClick(item, e)}
      className={({ isActive }) =>
        clsx(
          'relative inline-flex h-9 items-center gap-1.5 rounded-full px-3 text-[13px] transition',
          isActive
            ? 'bg-white/[0.06] text-white'
            : 'text-neutral-300 hover:bg-white/[0.04] hover:text-white',
        )
      }
    >
      {({ isActive }) => (
        <>
          <Icon size={14} className={isActive ? 'text-[#C4B5FD]' : ''} />
          <span>{item.label}</span>
          {isActive && (
            <span className="absolute -bottom-[7px] left-1/2 h-[2px] w-6 -translate-x-1/2 rounded-full bg-gradient-to-r from-[#7C3AED] to-[#22D3EE]" />
          )}
        </>
      )}
    </NavLink>
  );
}

function AccountMenu({
  me,
  onLogout,
  onNav,
  navigate,
}: {
  me: ReturnType<typeof useAuthStore.getState>['me'];
  onLogout: () => void;
  onNav: (item: NavItem, e: React.MouseEvent) => void;
  navigate: ReturnType<typeof useNavigate>;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    return () => document.removeEventListener('mousedown', onDoc);
  }, []);

  const initial = (me?.username || me?.email || 'U').slice(0, 1).toUpperCase();

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="inline-flex h-8 items-center gap-1.5 rounded-full border border-white/[0.10] bg-white/[0.04] py-0 pl-1 pr-2 text-[13px] text-neutral-200 transition hover:border-white/[0.20] hover:bg-white/[0.08]"
      >
        <span className="grid h-6 w-6 place-items-center rounded-full bg-gradient-to-br from-[#7C3AED] to-[#22D3EE] text-[11px] font-medium text-white">
          {initial}
        </span>
        <ChevronDown size={13} className={clsx('transition', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="absolute right-0 top-10 z-40 min-w-[220px] overflow-hidden rounded-2xl border border-white/[0.08] bg-[#0F0F1E]/95 p-1.5 shadow-[0_24px_60px_rgba(0,0,0,.55)] backdrop-blur-xl">
          <div className="px-3 py-2">
            <div className="truncate text-[13px] text-neutral-100">{me?.username || me?.email}</div>
            {typeof me?.points === 'number' && (
              <div className="mt-0.5 text-[11px] text-neutral-500">余额 {(me.points / 100).toFixed(2)} 点</div>
            )}
          </div>
          <div className="my-1 h-px bg-white/[0.06]" />
          {ACCOUNT_MENU.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.to}
                type="button"
                onMouseDown={(e) => e.preventDefault()}
                onClick={(e) => {
                  setOpen(false);
                  if (item.authed && !me) {
                    onNav(item, e);
                    return;
                  }
                  navigate(item.to);
                }}
                className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-[13px] text-neutral-200 transition hover:bg-white/[0.06] hover:text-white"
              >
                <Icon size={14} className="text-neutral-400" />
                {item.label}
              </button>
            );
          })}
          <div className="my-1 h-px bg-white/[0.06]" />
          <button
            type="button"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => {
              setOpen(false);
              onLogout();
            }}
            className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-[13px] text-neutral-300 transition hover:bg-white/[0.06] hover:text-white"
          >
            <LogOut size={14} className="text-neutral-400" />
            退出登录
          </button>
        </div>
      )}
    </div>
  );
}

function MobileDrawer({
  isAuthed,
  onClose,
  onNav,
  onLogout,
  openGate,
  version,
}: {
  isAuthed: boolean;
  onClose: () => void;
  onNav: (item: NavItem, e: React.MouseEvent) => void;
  onLogout: () => void;
  openGate: ReturnType<typeof useLoginGateStore.getState>['openGate'];
  version: string;
}) {
  return (
    <div
      className="fixed inset-0 z-40 flex items-stretch justify-end bg-black/72 backdrop-blur-sm lg:hidden"
      onMouseDown={onClose}
    >
      <div
        className="flex h-full w-[300px] flex-col border-l border-white/[0.08] bg-[#0F0F1E] p-4"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <span className="neon-text text-[15px] font-medium">image2api</span>
          <button
            type="button"
            onClick={onClose}
            className="grid h-8 w-8 place-items-center rounded-full border border-white/[0.10] text-neutral-300"
          >
            <X size={16} />
          </button>
        </div>
        <DrawerSection title="创作">
          {PRIMARY_NAV.map((item) => (
            <DrawerLink key={item.to} item={item} onClick={onNav} onClose={onClose} />
          ))}
        </DrawerSection>
        <DrawerSection title="内容">
          {SECONDARY_NAV.map((item) => (
            <DrawerLink key={item.to} item={item} onClick={onNav} onClose={onClose} />
          ))}
        </DrawerSection>
        <DrawerSection title="账户">
          {ACCOUNT_MENU.map((item) => (
            <DrawerLink key={item.to} item={item} onClick={onNav} onClose={onClose} />
          ))}
        </DrawerSection>
        <div className="mt-auto flex items-center justify-between text-[11px] text-neutral-500">
          <span>{version}</span>
          {isAuthed ? (
            <button type="button" onClick={onLogout} className="text-neutral-300 hover:text-white">
              退出登录
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                onClose();
                openGate({ hint: '登录后可保存作品和查看额度' });
              }}
              className="text-neutral-300 hover:text-white"
            >
              登录
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function DrawerSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-4">
      <div className="mb-1.5 px-2 text-[11px] uppercase tracking-wider text-neutral-500">{title}</div>
      <div className="flex flex-col gap-0.5">{children}</div>
    </div>
  );
}

function DrawerLink({
  item,
  onClick,
  onClose,
}: {
  item: NavItem;
  onClick: (item: NavItem, e: React.MouseEvent) => void;
  onClose: () => void;
}) {
  const Icon = item.icon;
  return (
    <NavLink
      to={item.to}
      onClick={(e) => {
        onClick(item, e);
        if (!e.defaultPrevented) onClose();
      }}
      className={({ isActive }) =>
        clsx(
          'flex items-center gap-2 rounded-lg px-2 py-2 text-[13px] transition',
          isActive
            ? 'bg-white/[0.08] text-white'
            : 'text-neutral-300 hover:bg-white/[0.06] hover:text-white',
        )
      }
    >
      <Icon size={15} className="text-neutral-400" />
      {item.label}
    </NavLink>
  );
}
