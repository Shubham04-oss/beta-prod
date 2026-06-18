import { useState, useEffect, useRef } from 'react';
import { getAuth } from 'firebase/auth';

export type AGUIIntent = {
  type: string;
  component?: string;
  data?: Record<string, unknown>;
};

export function useAgentStream() {
  const [messages, setMessages] = useState<AGUIIntent[]>([]);
  const [status, setStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let mounted = true;

    async function connect() {
      setStatus('connecting');
      try {
        const auth = getAuth();
        // We assume auth state is already initialized and the user is signed in
        const user = auth.currentUser;
        if (!user) {
          // If auth isn't ready immediately, you might want to handle this differently
          // in a production app (e.g., waiting for onAuthStateChanged)
          throw new Error('Not authenticated');
        }
        
        const token = await user.getIdToken();

        // 1. Fetch Ticket from BFF
        const res = await fetch('/api/tickets', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });

        if (!res.ok) {
          throw new Error('Failed to fetch ticket');
        }

        const { ticket_id } = await res.json();

        // 2. Connect WebSocket
        const wsBaseUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
        const wsUrl = `${wsBaseUrl}/api/v1/agent-stream?ticket=${ticket_id}`;
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          if (mounted) setStatus('connected');
        };

        ws.onmessage = (event) => {
          if (!mounted) return;
          try {
            const intent: AGUIIntent = JSON.parse(event.data);
            setMessages((prev) => [...prev, intent]);
          } catch (e) {
            console.error('Failed to parse WS message:', e);
          }
        };

        ws.onclose = () => {
          if (mounted) setStatus('disconnected');
        };

        ws.onerror = (e) => {
          console.error('WebSocket error:', e);
          if (mounted) setStatus('error');
        };

      } catch (error) {
        console.error('Agent stream connection error:', error);
        if (mounted) setStatus('error');
      }
    }

    connect();

    return () => {
      mounted = false;
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return { messages, status };
}
