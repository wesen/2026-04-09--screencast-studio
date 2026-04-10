import { fromJson, type JsonValue } from '@bufbuild/protobuf';
import type { AppDispatch } from '@/app/store';
import { ServerEventSchema } from '@/gen/proto/screencast/studio/v1/web_pb';
import { setPreviews, upsertPreview } from '@/features/previews/previewSlice';
import {
  addLog,
  setAudioMeter,
  setDiskStatus,
  setWsConnected,
  updateSessionState,
} from './sessionSlice';

const websocketUrl = () =>
  `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;

const RECONNECT_DELAY_MS = 2000;
const MAX_RECONNECT_DELAY_MS = 30000;

export class WsClient {
  private ws: WebSocket | null = null;
  private reconnectTimeout: number | null = null;
  private reconnectDelay = RECONNECT_DELAY_MS;
  private readonly dispatch: AppDispatch;
  private shouldReconnect = true;

  constructor(dispatch: AppDispatch) {
    this.dispatch = dispatch;
  }

  connect(): void {
    if (
      this.ws?.readyState === WebSocket.OPEN ||
      this.ws?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    this.shouldReconnect = true;
    this.ws = new WebSocket(websocketUrl());
    this.ws.onopen = () => {
      this.dispatch(setWsConnected(true));
      this.reconnectDelay = RECONNECT_DELAY_MS;
    };
    this.ws.onclose = () => {
      this.dispatch(setWsConnected(false));
      this.ws = null;
      this.scheduleReconnect();
    };
    this.ws.onerror = () => {
      this.dispatch(setWsConnected(false));
    };
    this.ws.onmessage = (messageEvent) => {
      try {
        const parsed = JSON.parse(messageEvent.data) as JsonValue;
        const event = fromJson(ServerEventSchema, parsed);

        switch (event.kind.case) {
          case 'sessionState':
            this.dispatch(updateSessionState(event.kind.value));
            break;
          case 'sessionLog':
          case 'previewLog':
            this.dispatch(addLog(event.kind.value));
            break;
          case 'previewList':
            this.dispatch(setPreviews(event.kind.value.previews));
            break;
          case 'previewState':
            this.dispatch(upsertPreview(event.kind.value));
            break;
          case 'audioMeter':
            this.dispatch(setAudioMeter(event.kind.value));
            break;
          case 'diskStatus':
            this.dispatch(setDiskStatus(event.kind.value));
            break;
          case undefined:
            break;
        }
      } catch {
        this.dispatch(setWsConnected(false));
      }
    };
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimeout !== null) {
      window.clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.dispatch(setWsConnected(false));
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect || this.reconnectTimeout !== null) {
      return;
    }

    this.reconnectTimeout = window.setTimeout(() => {
      this.reconnectTimeout = null;
      this.connect();
      this.reconnectDelay = Math.min(
        this.reconnectDelay * 2,
        MAX_RECONNECT_DELAY_MS
      );
    }, this.reconnectDelay);
  }
}
