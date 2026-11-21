import { useState } from 'react';
import { useStore } from '../store/useStore';

export default function AuthScreen() {
  const login = useStore(s => s.login);
  const [name, setName] = useState('You');

  return (
    <div className="screen" style={{ justifyContent: 'center', padding: 28, maxWidth: 400, margin: '0 auto' }}>
      <div style={{ textAlign: 'center', marginBottom: 36 }}>
        <div style={{
          width: 96, height: 96, borderRadius: 28, margin: '0 auto 16px',
          background: 'linear-gradient(135deg,#2AABEE,#1a8cd8)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 48,
          boxShadow: '0 16px 40px rgba(42,171,238,0.4)',
        }}>🔒</div>
        <h1 style={{ fontSize: 28, fontWeight: 700 }}>FlexGram</h1>
        <p style={{ color: 'var(--text2)', marginTop: 6, fontSize: 15 }}>Приватный офлайн-мессенджер</p>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <input
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder="Имя"
          style={{
            background: 'var(--bg2)', border: 'none', borderRadius: 14,
            padding: '15px 16px', fontSize: 16, outline: 'none',
          }}
          onKeyDown={e => e.key === 'Enter' && login(name)}
        />
        <button className="btn-primary" onClick={() => login(name)}>Войти</button>
        <button className="btn-secondary" onClick={() => login('Guest', true)}>Демо-аккаунт (decoy)</button>
      </div>
      <p style={{ textAlign: 'center', color: 'var(--text3)', fontSize: 12.5, marginTop: 24, lineHeight: 1.5 }}>
        Полностью офлайн. Данные только на устройстве.<br />
        Нет серверов Telegram · нет реальной переписки.
      </p>
    </div>
  );
}
