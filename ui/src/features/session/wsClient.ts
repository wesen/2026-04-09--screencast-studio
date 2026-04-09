import type { AppDispatch } from '@/app/store';
import { setWsConnected, addLog, setAudioLevel, updateSessionState } from './sessionSlice';
import type { WsEvent } from '@/api/types';

const WS_URL = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;

const RECONNECT_DELAY_MS = 2000;
const MAX_RECONNECT_DELAY_MS = 30000;

export class WsClient {
  private ws: WebSocket | null = null;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = RECONNECT_DELAY_MS;
  private dispatch: AppDispatch;
  private shouldReconnect = true;

  constructor(dispatch: AppDispatch) {
    this.dispatch = dispatch;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.shouldReconnect = true;
    this.ws = new WebSocket(WS_URL);

    this.ws.onopen = () => {
      console.log('[WS] Connected');
      this.dispatch(setWsConnected(true));
      this.reconnectDelay = RECONNECT_DELAY_MS;
    };

    this.ws.onclose = () => {
      console.log('[WS] Disconnected');
      this.dispatch(setWsConnected(false));
      this.scheduleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('[WS] Error:', error);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WsEvent;
        this.handleMessage(data);
      } catch (e) {
        console.error('[WS] Failed to parse message:', e);
      }
    };
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.dispatch(setWsConnected(false));
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) {
      return;
    }

    console.log(`[WS] Reconnecting in ${this.reconnectDelay}ms...`);
    this.reconnectTimeout = setTimeout(() => {
      this.connect();
      // Exponential backoff with max
      this.reconnectDelay = Math.min(
        this.reconnectDelay * 2,
        MAX_RECONNECT_DELAY_MS
      );
    }, this.reconnectDelay);
  }

  private handleMessage(event: WsEvent): void {
    switch (event.type) {
      case 'session.state':
        this.dispatch(updateSessionState({
          state: event.payload.state,
          reason: event.payload.reason,
        }));
        break;

      case 'session.log':
        this.dispatch(addLog({
          timestamp: Date.now(),
          level: event.payload.level,
          message: event.payload.line,
        }));
        break;

      case 'meter.audio':
        this.dispatch(setAudioLevel({
          deviceId: event.payload.device_id,
          level: event.payload.level,
        }));
        break;

      case 'session.output':
      case 'preview.state':
        // Handle in future phases
        console.log('[WS] Unhandled event:', event.type);
        break;

      default:
        console.log('[WS] Unknown event type:', event);
    }
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}

// Singleton instance
let wsClientInstance: WsClient | null = null;

export function getWsClient(dispatch: AppDispatch): WsClient {
  if (!wsClientInstance) {
    wsClientInstance = new WsClient(dispatch);
  }
  return wsClientInstance;
}
