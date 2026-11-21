export type Theme = 'dark' | 'light' | 'system';

export interface User {
  id: string;
  name: string;
  username: string;
  bio: string;
  phone?: string;
  avatarColor: string;
  avatarUrl?: string;
  premium: boolean;
  stars: number;
  statusEmoji?: string;
}

export interface Message {
  id: string;
  chatId: string;
  from: 'me' | 'them' | 'system';
  text?: string;
  type: 'text' | 'voice' | 'photo' | 'video' | 'file' | 'gift' | 'system' | 'video_note';
  time: string;
  timestamp: number;
  status: 'sending' | 'sent' | 'read';
  replyTo?: string;
  edited?: boolean;
  reactions?: Record<string, number>;
  duration?: string;
  giftEmoji?: string;
  giftName?: string;
  mediaUrl?: string;
  views?: number;
  selfDestructAt?: number;
}

export interface Chat {
  id: string;
  name: string;
  username: string;
  type: 'private' | 'group' | 'channel' | 'bot' | 'secret' | 'saved';
  folder: 'all' | 'personal' | 'groups' | 'channels' | 'bots' | 'archive';
  online: boolean;
  lastSeen?: string;
  verified: boolean;
  premium: boolean;
  pinned: boolean;
  archived: boolean;
  muted: boolean;
  unread: number;
  avatarColor: string;
  avatarUrl?: string;
  lastMessage?: string;
  lastTime: string;
  messages: Message[];
  // Fake / director mode
  fakeOnline?: boolean;
  fakeTyping?: boolean;
  fakeViews?: number;
  autoReplies?: { keyword: string; reply: string; delayMs: number }[];
}

export interface Gift {
  id: string;
  emoji: string;
  name: string;
  price: number;
  rarity: 'common' | 'rare' | 'legendary';
  ownedAt?: number;
}

export interface Story {
  id: string;
  name: string;
  own: boolean;
  seen: boolean;
  color: string;
  mediaUrl?: string;
}

export interface AppSettings {
  theme: Theme;
  accentColor: string;
  navStyle: 'tabs' | 'drawer';
  language: string;
  pinEnabled: boolean;
  pinCode?: string;
  decoyPin?: string;
  wallpaper?: string;
}

export interface WalletState {
  ton: number;
  usdt: number;
  notcoin: number;
  history: { id: string; type: string; amount: number; currency: string; time: string }[];
}
