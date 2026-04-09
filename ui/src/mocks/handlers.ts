import { http, HttpResponse, delay } from 'msw';
import { create, toJson } from '@bufbuild/protobuf';
import {
  mockDiscoveryData,
  mockRecordingSession,
  mockCompileResponse,
  mockPreviews,
} from './data';
import {
  CompileResponseSchema,
  DiscoveryResponseSchema,
  HealthResponseSchema,
  NormalizeResponseSchema,
  PreviewEnvelopeSchema,
  PreviewListResponseSchema,
  SessionEnvelopeSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export const handlers = [
  // Health check
  http.get('/api/healthz', () => {
    return HttpResponse.json(toJson(HealthResponseSchema, create(HealthResponseSchema, {
      ok: true,
      service: 'screencast-studio',
      previewLimit: 4,
    }), { alwaysEmitImplicit: true }));
  }),

  // Discovery endpoints
  http.get('/api/discovery', async () => {
    await delay(100);
    return HttpResponse.json(toJson(DiscoveryResponseSchema, mockDiscoveryData, {
      alwaysEmitImplicit: true,
    }));
  }),

  http.post('/api/setup/normalize', async ({ request }) => {
    await delay(150);
    await request.json(); // Parse body
    return HttpResponse.json(toJson(NormalizeResponseSchema, create(NormalizeResponseSchema, {
      sessionId: `session-${Date.now()}`,
      warnings: [],
      config: {
        schema: 'recorder.config/v1',
        sessionId: 'demo',
        destinationTemplates: { video: 'recordings/{session_id}/{name}.mov' },
        audioMixTemplate: 'recordings/{session_id}/mix.aac',
        audioOutput: { codec: 'aac', sampleRateHz: 48000, channels: 2 },
        videoSources: [],
        audioSources: [],
      },
    }), { alwaysEmitImplicit: true }));
  }),

  http.post('/api/setup/compile', async ({ request }) => {
    await delay(200);
    await request.json(); // Parse body
    return HttpResponse.json(toJson(CompileResponseSchema, create(CompileResponseSchema, {
      ...mockCompileResponse,
      sessionId: `session-${Date.now()}`,
    }), { alwaysEmitImplicit: true }));
  }),

  // Recording endpoints
  http.get('/api/recordings/current', async () => {
    await delay(50);
    return HttpResponse.json(toJson(SessionEnvelopeSchema, create(SessionEnvelopeSchema, {
      session: mockRecordingSession,
    }), { alwaysEmitImplicit: true }));
  }),

  http.post('/api/recordings/start', async ({ request }) => {
    await delay(300);
    await request.json(); // Parse body
    return HttpResponse.json(toJson(SessionEnvelopeSchema, create(SessionEnvelopeSchema, {
      session: {
        ...mockRecordingSession,
        active: true,
        sessionId: `session-${Date.now()}`,
        state: 'starting',
        reason: 'compiling and starting workers',
      },
    }), { alwaysEmitImplicit: true }));
  }),

  http.post('/api/recordings/stop', async () => {
    await delay(200);
    return HttpResponse.json(toJson(SessionEnvelopeSchema, create(SessionEnvelopeSchema, {
      session: {
        ...mockRecordingSession,
        active: false,
        sessionId: 'current',
        state: 'stopped',
        reason: 'user requested stop',
      },
    }), { alwaysEmitImplicit: true }));
  }),

  // Preview endpoints
  http.get('/api/previews', async () => {
    await delay(50);
    return HttpResponse.json(toJson(PreviewListResponseSchema, create(PreviewListResponseSchema, {
      previews: mockPreviews,
    }), { alwaysEmitImplicit: true }));
  }),

  http.post('/api/previews/ensure', async ({ request }) => {
    await delay(100);
    const body = await request.json() as { sourceId: string };
    return HttpResponse.json(toJson(PreviewEnvelopeSchema, create(PreviewEnvelopeSchema, {
      preview: {
        id: `preview-${body.sourceId}`,
        sourceId: body.sourceId,
        name: body.sourceId,
        sourceType: 'display',
        state: 'starting',
        leases: 1,
        hasFrame: false,
      },
    }), { alwaysEmitImplicit: true }));
  }),

  http.post('/api/previews/release', async ({ request }) => {
    await delay(50);
    const body = await request.json() as { previewId: string };
    return HttpResponse.json(toJson(PreviewEnvelopeSchema, create(PreviewEnvelopeSchema, {
      preview: {
        id: body.previewId,
        sourceId: body.previewId.replace(/^preview-/, ''),
        name: body.previewId.replace(/^preview-/, ''),
        sourceType: 'display',
        state: 'finished',
        reason: 'preview released',
        leases: 0,
        hasFrame: false,
      },
    }), { alwaysEmitImplicit: true }));
  }),
];
