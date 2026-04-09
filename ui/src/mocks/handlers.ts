import { http, HttpResponse, delay } from 'msw';
import {
  mockDiscoveryData,
  mockRecordingSession,
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

  http.post('/api/setup/normalize', async ({ request }) => {
    await delay(150);
    await request.json(); // Parse body
    return HttpResponse.json({
      session_id: `session-${Date.now()}`,
      warnings: [],
      config: {
        schema: 'recorder.config/v1',
        session_id: 'demo',
        destination_templates: { video: 'recordings/{session_id}/{name}.mov' },
        audio_mix_template: 'recordings/{session_id}/mix.aac',
        audio_output: { codec: 'aac', sample_rate_hz: 48000, channels: 2 },
        video_sources: [],
        audio_sources: [],
      },
    });
  }),

  http.post('/api/setup/compile', async ({ request }) => {
    await delay(200);
    await request.json(); // Parse body
    return HttpResponse.json({
      ...mockCompileResponse,
      session_id: `session-${Date.now()}`,
    });
  }),

  // Recording endpoints
  http.get('/api/recordings/current', async () => {
    await delay(50);
    return HttpResponse.json({ session: mockRecordingSession });
  }),

  http.post('/api/recordings/start', async ({ request }) => {
    await delay(300);
    await request.json(); // Parse body
    return HttpResponse.json({
      session: {
        ...mockRecordingSession,
        active: true,
        session_id: `session-${Date.now()}`,
        state: 'starting',
        reason: 'compiling and starting workers',
      },
    });
  }),

  http.post('/api/recordings/stop', async () => {
    await delay(200);
    return HttpResponse.json({
      session: {
        ...mockRecordingSession,
        active: false,
        session_id: 'current',
        state: 'stopped',
        reason: 'user requested stop',
      },
    });
  }),

  // Preview endpoints
  http.get('/api/previews', async () => {
    await delay(50);
    return HttpResponse.json({ previews: mockPreviews });
  }),

  http.post('/api/previews/ensure', async ({ request }) => {
    await delay(100);
    const body = await request.json() as { source_id: string };
    return HttpResponse.json({
      preview: {
        id: `preview-${body.source_id}`,
        source_id: body.source_id,
        name: body.source_id,
        source_type: 'display',
        state: 'starting',
        leases: 1,
        has_frame: false,
      },
    });
  }),

  http.post('/api/previews/release', async ({ request }) => {
    await delay(50);
    const body = await request.json() as { preview_id: string };
    return HttpResponse.json({
      preview: {
        id: body.preview_id,
        source_id: body.preview_id.replace(/^preview-/, ''),
        name: body.preview_id.replace(/^preview-/, ''),
        source_type: 'display',
        state: 'finished',
        reason: 'preview released',
        leases: 0,
        has_frame: false,
      },
    });
  }),
];
