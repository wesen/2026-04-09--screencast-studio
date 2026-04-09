import { baseApi } from './baseApi';
import { decodeProto, encodeProto } from './proto';
import type { CompileResponse, NormalizeResponse } from './types';
import {
  CompileResponseSchema,
  DslRequestSchema,
  NormalizeResponseSchema,
} from '@/gen/proto/screencast/studio/v1/web_pb';

export const setupApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    normalizeSetup: builder.mutation<NormalizeResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setup/normalize',
        method: 'POST',
        body: encodeProto(DslRequestSchema, body),
      }),
      transformResponse: (response) => decodeProto(NormalizeResponseSchema, response),
    }),
    compileSetup: builder.mutation<CompileResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setup/compile',
        method: 'POST',
        body: encodeProto(DslRequestSchema, body),
      }),
      transformResponse: (response) => decodeProto(CompileResponseSchema, response),
      invalidatesTags: ['Setup'],
    }),
  }),
});

export const { useNormalizeSetupMutation, useCompileSetupMutation } = setupApi;
