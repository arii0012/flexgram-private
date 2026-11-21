import { useEffect, useState } from 'react';
import { useStore } from './store/useStore';
import AuthScreen from './screens/AuthScreen';
import MainScreen from './screens/MainScreen';
import ChatScreen from './screens/ChatScreen';
import SettingsScreen from './screens/SettingsScreen';
import GiftsScreen from './screens/GiftsScreen';
import CallScreen from './screens/CallScreen';

export default function App() {
  const { isLoggedIn, currentChatId, settings, showCallScreen } = useStore();
  const [view, setView] = useState<'main' | 'settings' | 'gifts'>('main');

  useEffect(() => {
    const theme = settings.theme === 'system'
      ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
      : settings.theme;
    document.documentElement.setAttribute('data-theme', theme);
  }, [settings.theme]);

  if (!isLoggedIn) return <AuthScreen />;

  return (
    <div className="app">
      {/* Left column: chat list (always on desktop, or when no chat open on mobile) */}
      <div style={{
        width: currentChatId && window.innerWidth > 768 ? 360 : '100%',
        display: currentChatId && window.innerWidth <= 768 ? 'none' : 'flex',
        flexDirection: 'column',
        borderRight: '1px solid var(--sep)',
        height: '100%',
      }}>
        {view === 'main' && <MainScreen onOpenSettings={() => setView('settings')} onOpenGifts={() => setView('gifts')} />}
        {view === 'settings' && <SettingsScreen onBack={() => setView('main')} onOpenGifts={() => setView('gifts')} />}
        {view === 'gifts' && <GiftsScreen onBack={() => setView('settings')} />}
      </div>

      {/* Right column: active chat (desktop) or full screen chat (mobile) */}
      {currentChatId && (
        <div style={{ flex: 1, height: '100%', display: 'flex' }}>
          <ChatScreen />
        </div>
      )}

      {!currentChatId && window.innerWidth > 768 && (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text2)' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 64, marginBottom: 12 }}>💬</div>
            <div>Выберите чат</div>
            <div style={{ fontSize: 13, marginTop: 6 }}>или создайте новый</div>
          </div>
        </div>
      )}

      {showCallScreen && <CallScreen />}
    </div>
  );
}
