import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User, Chat, Message, Gift, Story, AppSettings, WalletState, Theme } from '../types';

const AVATAR_COLORS = ['#e17076','#7bc862','#e5ca77','#6ec9cb','#65aadd','#a695e7','#ee7aae','#f4a261','#e67e22','#1abc9c'];

const DEMO_GIFTS: Gift[] = [
  { id: 'g1', emoji: '🧸', name: 'Мишка', price: 15, rarity: 'common' },
  { id: 'g2', emoji: '🌹', name: 'Роза', price: 25, rarity: 'common' },
  { id: 'g3', emoji: '💎', name: 'Алмаз', price: 100, rarity: 'rare' },
  { id: 'g4', emoji: '🎂', name: 'Торт', price: 50, rarity: 'common' },
  { id: 'g5', emoji: '🎁', name: 'Подарок', price: 30, rarity: 'common' },
  { id: 'g6', emoji: '🚀', name: 'Ракета', price: 75, rarity: 'rare' },
  { id: 'g7', emoji: '🦄', name: 'Единорог', price: 150, rarity: 'legendary' },
  { id: 'g8', emoji: '🌙', name: 'Луна', price: 40, rarity: 'common' },
  { id: 'g9', emoji: '⭐', name: 'Звезда', price: 10, rarity: 'common' },
  { id: 'g10', emoji: '👑', name: 'Корона', price: 250, rarity: 'legendary' },
];

function nowTime() {
  return new Date().toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' });
}

function createDemoChats(): Chat[] {
  const t = nowTime();
  return [
    {
      id: 'c1', name: 'Алекс', username: 'alex_dev', type: 'private', folder: 'personal',
      online: true, verified: true, premium: true, pinned: true, archived: false, muted: false,
      unread: 2, avatarColor: '#65aadd', lastTime: 'сейчас', lastMessage: 'Можно спокойно хранить заметки',
      messages: [
        { id: 'm1', chatId: 'c1', from: 'them', text: 'Привет! Как тебе идея приватного мессенджера?', type: 'text', time: '14:02', timestamp: Date.now()-3600000, status: 'read' },
        { id: 'm2', chatId: 'c1', from: 'me', text: 'Огонь 🔥 Полностью локальный — никто не видит', type: 'text', time: '14:03', timestamp: Date.now()-3500000, status: 'read' },
        { id: 'm3', chatId: 'c1', from: 'them', text: 'Именно. Никаких серверов, всё у тебя на устройстве', type: 'text', time: '14:04', timestamp: Date.now()-3400000, status: 'read' },
        { id: 'm4', chatId: 'c1', from: 'them', text: 'Можно спокойно хранить заметки и черновики', type: 'text', time: '14:05', timestamp: Date.now()-3300000, status: 'read' },
      ],
      autoReplies: [
        { keyword: 'привет', reply: 'Привет! Как дела?', delayMs: 1500 },
        { keyword: 'ок', reply: '👍', delayMs: 800 },
      ]
    },
    {
      id: 'c2', name: 'Мария', username: 'maria', type: 'private', folder: 'personal',
      online: false, verified: false, premium: false, pinned: false, archived: false, muted: false,
      unread: 0, avatarColor: '#ee7aae', lastTime: '12:41', lastMessage: 'Спасибо 🙏',
      messages: [
        { id: 'm1', chatId: 'c2', from: 'them', text: 'Скинь потом тот файл', type: 'text', time: '12:30', timestamp: Date.now()-7200000, status: 'read' },
        { id: 'm2', chatId: 'c2', from: 'me', text: 'Ок, вечером', type: 'text', time: '12:35', timestamp: Date.now()-7000000, status: 'read' },
        { id: 'm3', chatId: 'c2', from: 'them', text: 'Спасибо 🙏', type: 'text', time: '12:41', timestamp: Date.now()-6800000, status: 'read' },
      ]
    },
    {
      id: 'c3', name: 'Команда', username: 'team', type: 'group', folder: 'groups',
      online: false, verified: true, premium: false, pinned: false, archived: false, muted: false,
      unread: 5, avatarColor: '#7bc862', lastTime: 'вчера', lastMessage: 'Всё остальное — по плану',
      messages: [
        { id: 'm1', chatId: 'c3', from: 'them', text: 'Собрание переносим на 15:00', type: 'text', time: '18:20', timestamp: Date.now()-86400000, status: 'read' },
        { id: 'm2', chatId: 'c3', from: 'me', text: 'Принято', type: 'text', time: '18:22', timestamp: Date.now()-86300000, status: 'read' },
        { id: 'm3', chatId: 'c3', from: 'them', text: 'Не забудьте обновить статусы', type: 'text', time: '18:25', timestamp: Date.now()-86200000, status: 'read' },
      ]
    },
    {
      id: 'c4', name: 'Saved Messages', username: 'saved', type: 'saved', folder: 'personal',
      online: false, verified: false, premium: false, pinned: true, archived: false, muted: false,
      unread: 0, avatarColor: '#2AABEE', lastTime: 'пн', lastMessage: 'Пароль от Wi-Fi...',
      messages: [
        { id: 'm1', chatId: 'c4', from: 'me', text: 'Идея: полностью офлайн FlexGram', type: 'text', time: '10:12', timestamp: Date.now()-172800000, status: 'read' },
        { id: 'm2', chatId: 'c4', from: 'me', text: 'Список покупок:\n• молоко\n• хлеб\n• кофе', type: 'text', time: '11:40', timestamp: Date.now()-172000000, status: 'read' },
      ]
    },
    {
      id: 'c5', name: 'FlexBot', username: 'flexbot', type: 'bot', folder: 'bots',
      online: true, verified: true, premium: false, pinned: false, archived: false, muted: false,
      unread: 1, avatarColor: '#e17076', lastTime: '11:00', lastMessage: 'Напиши /help',
      messages: [
        { id: 'm1', chatId: 'c5', from: 'them', text: 'Привет! Я локальный бот. Напиши /help', type: 'text', time: '11:00', timestamp: Date.now()-100000, status: 'read' },
      ],
      autoReplies: [
        { keyword: '/help', reply: 'Команды:\n/help — помощь\n/time — время\n/stars — твои звёзды\n/joke — шутка\n/gift — случайный подарок', delayMs: 600 },
        { keyword: '/time', reply: `Сейчас ${t}`, delayMs: 400 },
        { keyword: '/joke', reply: 'Почему программисты путают Хэллоуин и Рождество? Oct 31 == Dec 25 😄', delayMs: 900 },
      ]
    },
    {
      id: 'c6', name: 'Секретный чат', username: 'secret1', type: 'secret', folder: 'personal',
      online: false, verified: false, premium: false, pinned: false, archived: false, muted: false,
      unread: 0, avatarColor: '#1abc9c', lastTime: 'пт', lastMessage: '🔒 Секретный чат',
      messages: [
        { id: 'm1', chatId: 'c6', from: 'system', text: 'Секретный чат создан. Сообщения с таймером самоуничтожения.', type: 'system', time: '10:00', timestamp: Date.now()-259200000, status: 'read' },
      ]
    },
  ];
}

interface AppState {
  // Auth / user
  user: User | null;
  isLoggedIn: boolean;
  isDecoyMode: boolean;
  // Data
  chats: Chat[];
  ownedGifts: Gift[];
  stories: Story[];
  settings: AppSettings;
  wallet: WalletState;
  giftCatalog: Gift[];
  // UI
  currentChatId: string | null;
  currentFolder: string;
  searchQuery: string;
  replyToId: string | null;
  directorMode: boolean; // Fake Chat Builder
  directorFrom: 'me' | 'them';
  showCallScreen: boolean;
  callType: 'audio' | 'video' | null;
  // Actions
  login: (name: string, isDecoy?: boolean) => void;
  logout: () => void;
  setTheme: (t: Theme) => void;
  setCurrentChat: (id: string | null) => void;
  setFolder: (f: string) => void;
  setSearch: (q: string) => void;
  sendMessage: (text: string, type?: Message['type'], extra?: Partial<Message>) => void;
  editMessage: (chatId: string, msgId: string, text: string) => void;
  deleteMessage: (chatId: string, msgId: string) => void;
  setReplyTo: (id: string | null) => void;
  togglePinChat: (chatId: string) => void;
  archiveChat: (chatId: string) => void;
  createChat: (name: string, type?: Chat['type']) => string;
  buyGift: (giftId: string) => boolean;
  addStars: (amount: number) => void;
  toggleDirectorMode: () => void;
  setDirectorFrom: (from: 'me' | 'them') => void;
  setFakeStatus: (chatId: string, status: { online?: boolean; typing?: boolean; views?: number }) => void;
  insertSystemMessage: (chatId: string, text: string) => void;
  triggerCall: (type: 'audio' | 'video') => void;
  endCall: () => void;
  resetDemo: () => void;
  clearHistory: () => void;
  exportData: () => string;
  importData: (json: string) => void;
  updateWallet: (patch: Partial<WalletState>) => void;
  togglePremium: () => void;
}

export const useStore = create<AppState>()(
  persist(
    (set, get) => ({
      user: null,
      isLoggedIn: false,
      isDecoyMode: false,
      chats: createDemoChats(),
      ownedGifts: [],
      stories: [
        { id: 's0', name: 'Моя история', own: true, seen: false, color: '#2AABEE' },
        { id: 's1', name: 'Алекс', own: false, seen: false, color: '#65aadd' },
        { id: 's2', name: 'Мария', own: false, seen: true, color: '#ee7aae' },
        { id: 's3', name: 'Кирилл', own: false, seen: false, color: '#a695e7' },
      ],
      settings: {
        theme: 'dark',
        accentColor: '#2AABEE',
        navStyle: 'tabs',
        language: 'ru',
        pinEnabled: false,
      },
      wallet: {
        ton: 12500,
        usdt: 340.5,
        notcoin: 1_250_000,
        history: [
          { id: 'w1', type: 'receive', amount: 100, currency: 'TON', time: 'вчера' },
          { id: 'w2', type: 'send', amount: 25, currency: 'USDT', time: '2 дня назад' },
        ],
      },
      giftCatalog: DEMO_GIFTS,
      currentChatId: null,
      currentFolder: 'all',
      searchQuery: '',
      replyToId: null,
      directorMode: false,
      directorFrom: 'me',
      showCallScreen: false,
      callType: null,

      login: (name, isDecoy = false) => {
        const color = AVATAR_COLORS[Math.floor(Math.random() * AVATAR_COLORS.length)];
        set({
          user: {
            id: 'u1',
            name: name || 'You',
            username: (name || 'you').toLowerCase().replace(/\s+/g, '_').slice(0, 16),
            bio: isDecoy ? 'Демонстрационный аккаунт' : 'Локальный приватный режим',
            avatarColor: color,
            premium: !isDecoy,
            stars: isDecoy ? 0 : 50,
          },
          isLoggedIn: true,
          isDecoyMode: !!isDecoy,
          chats: isDecoy
            ? createDemoChats().slice(0, 2).map(c => ({ ...c, unread: 0, messages: c.messages.slice(0, 2) }))
            : get().chats.length ? get().chats : createDemoChats(),
        });
      },

      logout: () => set({ user: null, isLoggedIn: false, isDecoyMode: false, currentChatId: null }),

      setTheme: (theme) => set(s => ({ settings: { ...s.settings, theme } })),

      setCurrentChat: (id) => {
        if (id) {
          set(s => ({
            currentChatId: id,
            chats: s.chats.map(c => c.id === id ? { ...c, unread: 0 } : c),
            replyToId: null,
          }));
        } else {
          set({ currentChatId: null, replyToId: null, directorMode: false });
        }
      },

      setFolder: (f) => set({ currentFolder: f }),
      setSearch: (q) => set({ searchQuery: q }),
      setReplyTo: (id) => set({ replyToId: id }),

      sendMessage: (text, type = 'text', extra = {}) => {
        const { currentChatId, chats, directorMode, directorFrom, replyToId, user } = get();
        if (!currentChatId) return;
        const from = directorMode ? directorFrom : 'me';
        const msg: Message = {
          id: 'm' + Date.now(),
          chatId: currentChatId,
          from,
          text,
          type,
          time: nowTime(),
          timestamp: Date.now(),
          status: from === 'me' ? 'sent' : 'read',
          replyTo: replyToId || undefined,
          ...extra,
        };
        set({
          chats: chats.map(c => {
            if (c.id !== currentChatId) return c;
            return {
              ...c,
              messages: [...c.messages, msg],
              lastMessage: type === 'voice' ? '🎤 Голосовое' : type === 'gift' ? '🎁 Подарок' : text,
              lastTime: msg.time,
            };
          }),
          replyToId: null,
        });

        // Auto-reply simulation
        if (!directorMode && from === 'me') {
          const chat = get().chats.find(c => c.id === currentChatId);
          if (!chat || chat.type === 'saved') return;

          // Bot commands
          if (chat.type === 'bot' && text.startsWith('/')) {
            const rule = chat.autoReplies?.find(r => text.toLowerCase().startsWith(r.keyword));
            const replyText = rule?.reply
              || (text === '/stars' ? `У тебя ${user?.stars ?? 0} ⭐` : 'Не знаю команду. /help');
            setTimeout(() => {
              get().sendMessage(replyText, 'text');
              // force from them
              set(s => ({
                chats: s.chats.map(c => {
                  if (c.id !== currentChatId) return c;
                  const msgs = [...c.messages];
                  const last = msgs[msgs.length - 1];
                  if (last) last.from = 'them';
                  return { ...c, messages: msgs };
                }),
              }));
            }, rule?.delayMs ?? 700);
            return;
          }

          // Generic auto reply
          if (Math.random() > 0.45) {
            const replies = ['Ок 👍', 'Понял', 'Хаха', 'Интересно...', 'Круто!', 'Спасибо', '👀', 'Хорошо', 'Давай'];
            const reply = chat.autoReplies?.find(r => text.toLowerCase().includes(r.keyword))?.reply
              || replies[Math.floor(Math.random() * replies.length)];
            const delay = 1200 + Math.random() * 2000;
            setTimeout(() => {
              set(s => ({
                chats: s.chats.map(c => {
                  if (c.id !== currentChatId) return c;
                  return {
                    ...c,
                    messages: [...c.messages, {
                      id: 'm' + Date.now(),
                      chatId: currentChatId,
                      from: 'them' as const,
                      text: reply,
                      type: 'text' as const,
                      time: nowTime(),
                      timestamp: Date.now(),
                      status: 'read' as const,
                    }],
                    lastMessage: reply,
                    lastTime: nowTime(),
                  };
                }),
              }));
            }, delay);
          }
        }
      },

      editMessage: (chatId, msgId, text) => {
        set(s => ({
          chats: s.chats.map(c => {
            if (c.id !== chatId) return c;
            return {
              ...c,
              messages: c.messages.map(m => m.id === msgId ? { ...m, text, edited: true } : m),
            };
          }),
        }));
      },

      deleteMessage: (chatId, msgId) => {
        set(s => ({
          chats: s.chats.map(c => {
            if (c.id !== chatId) return c;
            return { ...c, messages: c.messages.filter(m => m.id !== msgId) };
          }),
        }));
      },

      togglePinChat: (chatId) => {
        set(s => ({
          chats: s.chats.map(c => c.id === chatId ? { ...c, pinned: !c.pinned } : c),
        }));
      },

      archiveChat: (chatId) => {
        set(s => ({
          chats: s.chats.map(c => c.id === chatId ? { ...c, archived: true, pinned: false } : c),
          currentChatId: s.currentChatId === chatId ? null : s.currentChatId,
        }));
      },

      createChat: (name, type = 'private') => {
        const id = 'c' + Date.now();
        const chat: Chat = {
          id, name, username: name.toLowerCase().replace(/\s+/g, '_').slice(0, 12),
          type, folder: type === 'bot' ? 'bots' : type === 'group' ? 'groups' : type === 'channel' ? 'channels' : 'personal',
          online: Math.random() > 0.5, verified: false, premium: false,
          pinned: false, archived: false, muted: false, unread: 0,
          avatarColor: AVATAR_COLORS[Math.floor(Math.random() * AVATAR_COLORS.length)],
          lastTime: 'сейчас', lastMessage: 'Чат создан',
          messages: [{
            id: 'm1', chatId: id, from: 'them', text: 'Привет! Это локальный чат 👋',
            type: 'text', time: nowTime(), timestamp: Date.now(), status: 'read',
          }],
        };
        set(s => ({ chats: [chat, ...s.chats] }));
        return id;
      },

      buyGift: (giftId) => {
        const { giftCatalog, user, ownedGifts } = get();
        const g = giftCatalog.find(x => x.id === giftId);
        if (!g || !user || user.stars < g.price) return false;
        set({
          user: { ...user, stars: user.stars - g.price },
          ownedGifts: [...ownedGifts, { ...g, ownedAt: Date.now() }],
        });
        return true;
      },

      addStars: (amount) => {
        const { user } = get();
        if (user) set({ user: { ...user, stars: user.stars + amount } });
      },

      toggleDirectorMode: () => set(s => ({ directorMode: !s.directorMode })),
      setDirectorFrom: (from) => set({ directorFrom: from }),

      setFakeStatus: (chatId, status) => {
        set(s => ({
          chats: s.chats.map(c => c.id === chatId ? {
            ...c,
            online: status.online ?? c.online,
            fakeTyping: status.typing ?? c.fakeTyping,
            fakeViews: status.views ?? c.fakeViews,
          } : c),
        }));
      },

      insertSystemMessage: (chatId, text) => {
        set(s => ({
          chats: s.chats.map(c => {
            if (c.id !== chatId) return c;
            return {
              ...c,
              messages: [...c.messages, {
                id: 'sys' + Date.now(),
                chatId,
                from: 'system' as const,
                text,
                type: 'system' as const,
                time: nowTime(),
                timestamp: Date.now(),
                status: 'read' as const,
              }],
            };
          }),
        }));
      },

      triggerCall: (type) => set({ showCallScreen: true, callType: type }),
      endCall: () => set({ showCallScreen: false, callType: null }),

      resetDemo: () => set({
        chats: createDemoChats(),
        ownedGifts: [],
        stories: get().stories.map(s => ({ ...s, seen: false })),
      }),

      clearHistory: () => set(s => ({
        chats: s.chats.map(c => ({ ...c, messages: [], unread: 0, lastMessage: '' })),
      })),

      exportData: () => JSON.stringify({
        user: get().user,
        chats: get().chats,
        ownedGifts: get().ownedGifts,
        settings: get().settings,
        wallet: get().wallet,
      }, null, 2),

      importData: (json) => {
        try {
          const data = JSON.parse(json);
          set({
            user: data.user ?? get().user,
            chats: data.chats ?? get().chats,
            ownedGifts: data.ownedGifts ?? [],
            settings: data.settings ?? get().settings,
            wallet: data.wallet ?? get().wallet,
          });
        } catch { /* ignore */ }
      },

      updateWallet: (patch) => set(s => ({ wallet: { ...s.wallet, ...patch } })),

      togglePremium: () => {
        const { user } = get();
        if (user) set({ user: { ...user, premium: !user.premium } });
      },
    }),
    {
      name: 'flexgram-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (s) => ({
        user: s.user,
        isLoggedIn: s.isLoggedIn,
        chats: s.chats,
        ownedGifts: s.ownedGifts,
        stories: s.stories,
        settings: s.settings,
        wallet: s.wallet,
      }),
    }
  )
);
