import { http, HttpResponse, delay } from 'msw';
import {
  mockDiscoveryData,
  mockRecordingState,
  mockCompileResponse,
  mockPreviews,
} from './data';

export const handlers = [
  // Health check
  http.get('/api/healthz', () => {
    return HttpResponse.json({ status: 'ok', timestamp: new Date().toISOString() });
  }),

  // Discovery endpoints
  http.get('/api/discovery', async () => {
    await delay(100);
    return HttpResponse.json(mockDiscoveryData);
  }),

  http.get('/api/discovery/:kind', async ({ params }) => {
    await delay(50);
    const kind = params.kind as string;
    const items = mockDiscoveryData.items.filter((item) => item.kind === kind);
    return HttpResponse.json(items);
  }),

  http.post('/api/discovery/refresh', async () => {
    await delay(200);
    return HttpResponse.json({ success: true });
  }),

  // Setups endpoints
  http.get('/api/setups/example', async () => {
    await delay(100);
    return HttpResponse.json({
      dsl: `schema: recorder.config/v1
session_id: demo
destination_templates:
  video: recordings/{session_id}/{name}.mov
video_sources:
  - id: desktop-1
    name: Full Desktop
    type: display
    target:
      display: display-1
settings:
  capture:
    fps: 24
  output:
    container: mov
    video_codec: h264
    quality: 75
audio_sources:
  - id: mic-1
    name: Built-in Mic
    device: default
`,
    });
  }),

  http.post('/api/setups/normalize', async ({ request }) => {
    await delay(150);
    await request.json(); // Parse body
    return HttpResponse.json({
      valid: true,
      config: {
        schema: 'recorder.config/v1',
        session_id: 'demo',
        destination_templates: { video: 'recordings/{session_id}/{name}.mov' },
        screen_capture_defaults: {
          capture: { fps: 24, size: '' },
          output: { container: 'mov', video_codec: 'h264', quality: 75 },
        },
        camera_capture_defaults: {
          capture: { fps: 24, size: '' },
          output: { container: 'mov', video_codec: 'h264', quality: 75 },
        },
        audio_defaults: {
          output: { codec: 'aac', sample_rate_hz: 48000, channels: 2 },
        },
        audio_mix: { destination_template: '' },
        video_sources: [],
        audio_sources: [],
      },
    });
  }),

  http.post('/api/setups/compile', async ({ request }) => {
    await delay(200);
    await request.json(); // Parse body
    return HttpResponse.json({
      ...mockCompileResponse,
      session_id: `session-${Date.now()}`,
    });
  }),

  http.post('/api/setups/validate', async () => {
    await delay(100);
    return HttpResponse.json({ valid: true, errors: [] });
  }),

  // Recording endpoints
  http.get('/api/recordings/current', async () => {
    await delay(50);
    return HttpResponse.json(mockRecordingState);
  }),

  http.post('/api/recordings/start', async ({ request }) => {
    await delay(300);
    await request.json(); // Parse body
    return HttpResponse.json({
      session_id: `session-${Date.now()}`,
      state: 'starting',
      reason: 'compiling and starting workers',
    });
  }),

  http.post('/api/recordings/stop', async () => {
    await delay(200);
    return HttpResponse.json({
      session_id: 'current',
      state: 'stopped',
      reason: 'user requested stop',
    });
  }),

  // Preview endpoints
  http.get('/api/previews', async () => {
    await delay(50);
    return HttpResponse.json(mockPreviews);
  }),

  http.post('/api/previews/ensure', async ({ request }) => {
    await delay(100);
    const body = await request.json() as { source_id: string };
    return HttpResponse.json({
      id: `preview-${body.source_id}`,
      source_id: body.source_id,
      state: 'starting',
      stream_url: `/api/previews/${body.source_id}/mjpeg`,
    });
  }),

  http.post('/api/previews/release', async () => {
    await delay(50);
    return HttpResponse.json({ success: true });
  }),
];
