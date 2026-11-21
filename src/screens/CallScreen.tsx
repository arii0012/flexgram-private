import { useStore } from '../store/useStore';

export default function CallScreen() {
  const { callType, endCall, chats, currentChatId } = useStore();
  const chat = chats.find(c => c.id === currentChatId);

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 300,
      background: 'linear-gradient(180deg,#1a2a3a 0%,#0e1621 100%)',
      display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
      color: '#fff',
    }}>
      <div className="avatar" style={{ width: 100, height: 100, fontSize: 40, background: chat?.avatarColor || '#2AABEE', marginBottom: 20 }}>
        {chat?.name?.[0] || '?'}
      </div>
      <div style={{ fontSize: 24, fontWeight: 600 }}>{chat?.name || 'Звонок'}</div>
      <div style={{ color: 'var(--text2)', marginTop: 8, fontSize: 15 }}>
        {callType === 'video' ? 'Видеозвонок...' : 'Звонок...'}
      </div>
      <div style={{ marginTop: 8, fontSize: 13, opacity: 0.6 }}>ооо... (локальная имитация)</div>
      <div style={{ display: 'flex', gap: 32, marginTop: 60 }}>
        <button onClick={endCall} style={{
          width: 64, height: 64, borderRadius: '50%', background: '#e53935', border: 'none',
          color: '#fff', fontSize: 28, display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>📞</button>
        {callType === 'video' && (
          <button style={{
            width: 64, height: 64, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', border: 'none',
            color: '#fff', fontSize: 28, display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>📹</button>
        )}
      </div>
    </div>
  );
}
