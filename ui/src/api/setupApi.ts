import { baseApi } from './baseApi';
import type { CompileResponse, NormalizeResponse } from './types';

export const setupApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    normalizeSetup: builder.mutation<NormalizeResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setup/normalize',
        method: 'POST',
        body,
      }),
    }),
    compileSetup: builder.mutation<CompileResponse, { dsl: string }>({
      query: (body) => ({
        url: '/setup/compile',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Setup'],
    }),
  }),
});

export const { useNormalizeSetupMutation, useCompileSetupMutation } = setupApi;
